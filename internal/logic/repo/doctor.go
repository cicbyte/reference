package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cicbyte/reference/internal/log"
	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/go-git/go-git/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DoctorConfig struct {
	ProjectDir string
}

type DoctorProcessor struct {
	config *DoctorConfig
	db     *gorm.DB
}

type checkResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok, fixed, warn
	Details string `json:"details"`
}

type DoctorResult struct {
	ProjectDir string      `json:"project_dir"`
	Checks     []CheckItem `json:"checks"`
	Summary    string      `json:"summary"`
}

type CheckItem struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Details string `json:"details"`
	Group  string `json:"group"` // core / agent
}

func NewDoctorProcessor(config *DoctorConfig, db *gorm.DB) *DoctorProcessor {
	return &DoctorProcessor{config: config, db: db}
}

func (p *DoctorProcessor) Execute(ctx context.Context) (*DoctorResult, error) {
	settings := models.LoadProjectSettings(p.config.ProjectDir)
	if !settings.Initialized {
		return nil, fmt.Errorf("尚未初始化。请先运行 reference 完成初始化配置。")
	}

	hasAgent := settings.Agent != ""

	coreChecks := []checkResult{
		p.checkSymlinks(),
		p.checkWikiJunctions(),
		p.checkReferenceMap(),
		p.checkDatabaseConsistency(),
		p.checkWikiGit(),
	}

	var agentChecks []checkResult
	if hasAgent {
		agentChecks = []checkResult{
			p.checkAgentFiles(),
			p.checkSkillFile(),
		}
	}

	result := &DoctorResult{
		ProjectDir: p.config.ProjectDir,
		Checks:     make([]CheckItem, 0, len(coreChecks)+len(agentChecks)),
	}

	var fixedCount, warnCount int
	for _, c := range coreChecks {
		if c.Status == "fixed" {
			fixedCount++
		} else if c.Status == "warn" {
			warnCount++
		}
		result.Checks = append(result.Checks, CheckItem{
			Name: c.Name, Status: c.Status, Details: c.Details, Group: "core",
		})
	}
	for _, c := range agentChecks {
		if c.Status == "fixed" {
			fixedCount++
		} else if c.Status == "warn" {
			warnCount++
		}
		result.Checks = append(result.Checks, CheckItem{
			Name: c.Name, Status: c.Status, Details: c.Details, Group: "agent",
		})
	}

	if fixedCount > 0 || warnCount > 0 {
		result.Summary = fmt.Sprintf("修复 %d 个问题", fixedCount)
		if warnCount > 0 {
			result.Summary += fmt.Sprintf("，%d 个警告", warnCount)
		}
	} else {
		result.Summary = "一切正常，无需修复"
	}

	return result, nil
}

func repoRefName(r *models.Repo) string {
	if r.RefName != "" {
		return r.RefName
	}
	return r.LinkName
}

func (p *DoctorProcessor) checkSymlinks() checkResult {
	indexer := NewRepoIndexer(p.db)
	repos, err := indexer.List(p.config.ProjectDir)
	if err != nil {
		return checkResult{Name: "软链接完整性", Status: "warn", Details: fmt.Sprintf("数据库查询失败: %v", err)}
	}

	reposDir := filepath.Join(p.config.ProjectDir, ".reference", "repos")
	total, ok := len(repos), 0
	var fixed []string

	for _, r := range repos {
		refName := repoRefName(&r)
		linkPath := filepath.Join(reposDir, refName)
		if _, err := os.Stat(linkPath); err != nil {
			var target string
			if r.RefType == "remote" {
				target = r.CachePath
			} else {
				target = r.LocalPath
			}
			if target != "" {
				if err := CreateLink(target, linkPath); err != nil {
					log.Warn("修复软链接失败", zap.String("repo", r.LinkName), zap.Error(err))
					continue
				}
				fixed = append(fixed, refName)
			}
		} else {
			ok++
		}
	}

	details := fmt.Sprintf("%d/%d 正常", ok, total)
	if len(fixed) > 0 {
		sort.Strings(fixed)
		details += fmt.Sprintf("，已修复: %s", strings.Join(fixed, ", "))
		return checkResult{Name: "软链接完整性", Status: "fixed", Details: details}
	}
	return checkResult{Name: "软链接完整性", Status: "ok", Details: details}
}

func (p *DoctorProcessor) checkAgentFiles() checkResult {
	agentsDir := filepath.Join(p.config.ProjectDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return checkResult{Name: "Agent 文件", Status: "warn", Details: fmt.Sprintf("创建目录失败: %v", err)}
	}

	agentNames := []string{"reference-explorer", "reference-analyzer"}
	var updated []string

	for _, name := range agentNames {
		embedPath := "prompts/agents/" + name + ".md"
		dst := filepath.Join(agentsDir, name+".md")
		newData, err := readEmbedded(embedPath)
		if err != nil {
			log.Warn("读取 agent 模板失败", zap.String("agent", name), zap.Error(err))
			continue
		}
		if oldData, err := os.ReadFile(dst); err == nil && string(oldData) == string(newData) {
			continue
		}
		if err := os.WriteFile(dst, newData, 0644); err != nil {
			log.Warn("写入 agent 文件失败", zap.String("agent", name), zap.Error(err))
			continue
		}
		updated = append(updated, name+".md")
	}

	if len(updated) == 0 {
		return checkResult{Name: "Agent 文件", Status: "ok", Details: "正常"}
	}
	return checkResult{Name: "Agent 文件", Status: "ok", Details: "已更新: " + strings.Join(updated, ", ")}
}

func (p *DoctorProcessor) checkSkillFile() checkResult {
	skillPath := filepath.Join(p.config.ProjectDir, ".claude", "skills", "reference", "SKILL.md")

	newData, err := readEmbedded("prompts/skills/reference/SKILL.md")
	if err != nil {
		return checkResult{Name: "SKILL.md", Status: "warn", Details: "读取模板失败: " + err.Error()}
	}

	if err := os.MkdirAll(filepath.Dir(skillPath), 0755); err != nil {
		return checkResult{Name: "SKILL.md", Status: "warn", Details: fmt.Sprintf("创建目录失败: %v", err)}
	}

	if oldData, err := os.ReadFile(skillPath); err == nil && string(oldData) == string(newData) {
		return checkResult{Name: "SKILL.md", Status: "ok", Details: "正常"}
	}
	if err := os.WriteFile(skillPath, newData, 0644); err != nil {
		return checkResult{Name: "SKILL.md", Status: "warn", Details: "写入失败: " + err.Error()}
	}
	return checkResult{Name: "SKILL.md", Status: "ok", Details: "已更新"}
}

func (p *DoctorProcessor) checkWikiJunctions() checkResult {
	indexer := NewRepoIndexer(p.db)
	repos, err := indexer.List(p.config.ProjectDir)
	if err != nil {
		return checkResult{Name: "Wiki Junction", Status: "warn", Details: fmt.Sprintf("数据库查询失败: %v", err)}
	}

	wikiLinkDir := filepath.Join(p.config.ProjectDir, ".reference", "wiki")
	wikiBase := filepath.Join(utils.ConfigInstance.GetAppDir(), "wiki")
	reposDir := filepath.Join(p.config.ProjectDir, ".reference", "repos")
	total, ok := len(repos), 0
	var fixed []string

	for _, r := range repos {
		refName := repoRefName(&r)
		linkDir := filepath.Join(wikiLinkDir, refName)
		wikiDir := filepath.Join(wikiBase, r.WikiSubPath)
		linkPath := filepath.Join(reposDir, refName)

		wikiFile := filepath.Join(wikiDir, "reference.md")
		if _, err := os.Stat(wikiFile); os.IsNotExist(err) {
			if genErr := generateWikiReference(wikiDir, linkPath, &r); genErr != nil {
				log.Warn("生成 wiki 失败", zap.String("repo", r.LinkName), zap.Error(genErr))
			}
		}

		if _, err := os.Stat(wikiDir); err == nil {
			if _, err := os.Lstat(linkDir); err != nil {
				if err := CreateLink(wikiDir, linkDir); err != nil {
					log.Warn("创建 wiki 链接失败", zap.String("repo", r.LinkName), zap.Error(err))
					continue
				}
				fixed = append(fixed, refName)
			} else {
				ok++
			}
		}
	}

	details := fmt.Sprintf("%d/%d 正常", ok, total)
	if len(fixed) > 0 {
		details += fmt.Sprintf("，已修复: %s", strings.Join(fixed, ", "))
		return checkResult{Name: "Wiki Junction", Status: "fixed", Details: details}
	}
	return checkResult{Name: "Wiki Junction", Status: "ok", Details: details}
}

func (p *DoctorProcessor) checkReferenceMap() checkResult {
	refDir := filepath.Join(p.config.ProjectDir, ".reference")
	mapPath := filepath.Join(refDir, "reference.map.jsonl")

	indexer := NewRepoIndexer(p.db)
	repos, err := indexer.List(p.config.ProjectDir)
	if err != nil || len(repos) == 0 {
		if _, err := os.Stat(mapPath); err == nil {
			os.Remove(mapPath)
			return checkResult{Name: "Reference Map", Status: "ok", Details: "已清理（无仓库记录）"}
		}
		return checkResult{Name: "Reference Map", Status: "ok", Details: "正常（无仓库记录）"}
	}

	if err := refreshReferenceMap(p.config.ProjectDir, refDir, indexer); err != nil {
		return checkResult{Name: "Reference Map", Status: "warn", Details: "重新生成失败: " + err.Error()}
	}
	return checkResult{Name: "Reference Map", Status: "fixed", Details: "已重新生成"}
}

func (p *DoctorProcessor) checkDatabaseConsistency() checkResult {
	indexer := NewRepoIndexer(p.db)
	repos, err := indexer.List(p.config.ProjectDir)
	if err != nil {
		return checkResult{Name: "数据库一致性", Status: "warn", Details: "查询失败"}
	}

	reposDir := filepath.Join(p.config.ProjectDir, ".reference", "repos")
	var orphaned []string

	for _, r := range repos {
		refName := repoRefName(&r)
		linkPath := filepath.Join(reposDir, refName)
		if _, err := os.Stat(linkPath); os.IsNotExist(err) {
			orphaned = append(orphaned, refName)
		}
	}

	if len(orphaned) > 0 {
		return checkResult{
			Name:    "数据库一致性",
			Status:  "warn",
			Details: fmt.Sprintf("发现 %d 条孤立记录: %s（链接已被手动删除）", len(orphaned), strings.Join(orphaned, ", ")),
		}
	}

	entries, err := os.ReadDir(reposDir)
	if err != nil {
		return checkResult{Name: "数据库一致性", Status: "ok", Details: "正常"}
	}

	repoSet := make(map[string]bool, len(repos))
	for _, r := range repos {
		repoSet[repoRefName(&r)] = true
	}

	var untracked []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := os.Lstat(filepath.Join(reposDir, e.Name()))
		if err != nil || (!info.IsDir() && info.Mode()&os.ModeSymlink == 0) {
			continue
		}
		if !repoSet[e.Name()] {
			untracked = append(untracked, e.Name())
		}
	}

	if len(untracked) > 0 {
		return checkResult{
			Name:    "数据库一致性",
			Status:  "warn",
			Details: fmt.Sprintf("发现 %d 个未跟踪的链接: %s", len(untracked), strings.Join(untracked, ", ")),
		}
	}

	return checkResult{Name: "数据库一致性", Status: "ok", Details: "正常"}
}

func (p *DoctorProcessor) checkWikiGit() checkResult {
	wikiDir := filepath.Join(utils.ConfigInstance.GetAppDir(), "wiki")

	if _, err := os.Stat(wikiDir); os.IsNotExist(err) {
		return checkResult{Name: "Wiki Git", Status: "warn", Details: "wiki 目录不存在"}
	}

	if !logicwiki.IsGitInitialized(wikiDir) {
		if err := logicwiki.EnsureGitInit(wikiDir); err != nil {
			return checkResult{Name: "Wiki Git", Status: "warn", Details: fmt.Sprintf("初始化失败: %v", err)}
		}
		return checkResult{Name: "Wiki Git", Status: "fixed", Details: "已初始化"}
	}

	var details []string
	if logicwiki.HasRemote(wikiDir) {
		if url, err := logicwiki.GetRemoteURL(wikiDir); err == nil && url != "" {
			details = append(details, "远程: "+url)
		}
	} else {
		details = append(details, "无远程仓库")
	}

	repo, err := git.PlainOpen(wikiDir)
	if err != nil {
		return checkResult{Name: "Wiki Git", Status: "warn", Details: fmt.Sprintf("仓库异常: %v", err)}
	}
	wt, err := repo.Worktree()
	if err != nil {
		return checkResult{Name: "Wiki Git", Status: "warn", Details: "无法获取工作树"}
	}
	status, err := wt.Status()
	if err != nil {
		return checkResult{Name: "Wiki Git", Status: "warn", Details: "无法获取状态"}
	}
	if status.IsClean() {
		details = append(details, "工作区干净")
	} else {
		details = append(details, "有未提交的更改")
	}

	return checkResult{Name: "Wiki Git", Status: "ok", Details: strings.Join(details, "\uFF0C")}
}
