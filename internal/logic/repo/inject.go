package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/cicbyte/reference/internal/common"
	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type repoData struct {
	LinkName    string
	RefName     string
	WikiSubPath string
	WikiDir     string
	Type        string
	Platform    string
	FullName    string
	Description string
}

type mapRepoEntry struct {
	RefName   string             `json:"ref_name"`
	Type      string             `json:"type"`
	Platform  string             `json:"platform,omitempty"`
	FullName  string             `json:"full_name"`
	Desc      string             `json:"description,omitempty"`
	RepoPath  string             `json:"repo_path"`
	WikiPath  string             `json:"wiki_path"`
	Commit    string             `json:"commit,omitempty"`
	Topics    []mapTopicEntry    `json:"topics,omitempty"`
}

type mapTopicEntry struct {
	File        string `json:"file"`
	Description string `json:"description"`
	Commit      string `json:"commit"`
}

type InjectConfig struct {
	ProjectDir string
}

type InjectProcessor struct {
	config *InjectConfig
	db     *gorm.DB
}

func NewInjectProcessor(config *InjectConfig, db *gorm.DB) *InjectProcessor {
	return &InjectProcessor{config: config, db: db}
}

func (p *InjectProcessor) Execute(ctx context.Context) (string, error) {
	indexer := NewRepoIndexer(p.db)
	repos, err := indexer.List(p.config.ProjectDir)
	if err != nil {
		return "", err
	}

	refDir := filepath.Join(p.config.ProjectDir, ".reference")
	reposDir := filepath.Join(refDir, "repos")
	wikiJunctionDir := filepath.Join(refDir, "wiki")

	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return "", fmt.Errorf("创建 repos 目录失败: %w", err)
	}
	if err := os.MkdirAll(wikiJunctionDir, 0755); err != nil {
		return "", fmt.Errorf("创建 wiki 目录失败: %w", err)
	}

	repairCount := p.repairSymlinks(reposDir, repos)

	var repoDataList []repoData
	for _, r := range repos {
		refName := r.GetRefName()
		linkPath := filepath.Join(reposDir, refName)
		wikiBase := filepath.Join(utils.ConfigInstance.GetAppDir(), "wiki")
		wikiDir := filepath.Join(wikiBase, r.WikiSubPath)
		rd := repoData{
			LinkName:    r.LinkName,
			RefName:     refName,
			WikiSubPath: r.WikiSubPath,
			WikiDir:     wikiDir,
			Type:        string(r.RefType),
		}

		if r.RefType == models.RefTypeRemote {
			rd.Platform = r.Host
			rd.FullName = r.Namespace + "/" + r.RepoName
		} else {
			rd.Platform = "local"
			rd.FullName = filepath.Base(r.LocalPath)
		}

		wikiFile := filepath.Join(wikiDir, "reference.md")
		if _, err := os.Stat(wikiFile); os.IsNotExist(err) {
			if genErr := generateWikiReference(wikiDir, linkPath, &r); genErr != nil {
				log.Warn("生成 wiki 内容失败", zap.String("repo", r.LinkName), zap.Error(genErr))
			}
		}

		rd.Description = detectDescription(linkPath, &r)
		repoDataList = append(repoDataList, rd)
	}

	sort.Slice(repoDataList, func(i, j int) bool {
		return repoDataList[i].RefName < repoDataList[j].RefName
	})

	if err := generateReferenceMap(refDir, repoDataList); err != nil {
		log.Warn("生成 reference.map.jsonl 失败", zap.Error(err))
	}

	wikiFiles := p.injectWikiJunctions(wikiJunctionDir, repoDataList)

	var agentFiles, skillFiles []string
	settings := models.LoadProjectSettings(p.config.ProjectDir)
	if settings.Agent == "claude" {
		claudeDir := filepath.Join(p.config.ProjectDir, ".claude")
		agentFiles = p.injectAgents(claudeDir)
		skillFiles = p.injectSkill(claudeDir)
	}

	total := len(agentFiles) + len(skillFiles) + len(wikiFiles)
	if total == 0 && repairCount == 0 {
		return "配置已是最新。", nil
	}

	var sb strings.Builder
	if len(wikiFiles) > 0 {
		sb.WriteString(fmt.Sprintf("已链接 %d 个仓库知识", len(wikiFiles)))
	}
	if len(agentFiles)+len(skillFiles) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("，")
		}
		sb.WriteString(fmt.Sprintf("已更新 %d 个 AI 配置文件", len(agentFiles)+len(skillFiles)))
	}
	if repairCount > 0 {
		if sb.Len() > 0 {
			sb.WriteString("，")
		}
		sb.WriteString(fmt.Sprintf("已修复 %d 个引用链接", repairCount))
	}
	sb.WriteString("。")

	return sb.String(), nil
}

func (p *InjectProcessor) repairSymlinks(reposDir string, repos []models.Repo) int {
	fixed := 0
	for _, r := range repos {
		refName := r.GetRefName()
		linkPath := filepath.Join(reposDir, refName)
		if _, err := os.Stat(linkPath); err != nil {
			target := r.CachePath
			if r.RefType == models.RefTypeLocal {
				target = r.LocalPath
			}
			if target != "" {
				if err := CreateLink(target, linkPath); err != nil {
					log.Warn("修复软链接失败", zap.String("repo", r.LinkName), zap.Error(err))
					continue
				}
				fixed++
			}
		}
	}
	return fixed
}

func (p *InjectProcessor) injectAgents(claudeDir string) []string {
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		log.Warn("创建 agents 目录失败", zap.Error(err))
		return nil
	}

	var updated []string
	for _, agentName := range []string{"reference-explorer", "reference-analyzer"} {
		dst := filepath.Join(agentsDir, agentName+".md")
		if err := extractEmbedded("prompts/agents/"+agentName+".md", dst); err != nil {
			log.Warn("复制 agent 文件失败", zap.String("agent", agentName), zap.Error(err))
		} else {
			updated = append(updated, agentName+".md")
		}
	}
	return updated
}

func (p *InjectProcessor) injectSkill(claudeDir string) []string {
	skillsDir := filepath.Join(claudeDir, "skills", "reference")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		log.Warn("创建 skills 目录失败", zap.Error(err))
		return nil
	}

	data, err := readEmbedded("prompts/skills/reference/SKILL.md")
	if err != nil {
		log.Warn("读取 SKILL.md 模板失败", zap.Error(err))
		return nil
	}

	skillPath := filepath.Join(skillsDir, "SKILL.md")
	if err := renderSkill(data, skillPath); err != nil {
		log.Warn("生成 SKILL.md 失败", zap.Error(err))
		return nil
	}
	return []string{"SKILL.md"}
}

func (p *InjectProcessor) injectWikiJunctions(wikiJunctionDir string, repos []repoData) []string {
	cleanStaleJunctions(wikiJunctionDir, repos)

	wikiBase := filepath.Join(utils.ConfigInstance.GetAppDir(), "wiki")

	var linked []string
	for _, rd := range repos {
		wikiDir := filepath.Join(wikiBase, rd.WikiSubPath)
		linkDir := filepath.Join(wikiJunctionDir, rd.RefName)

		if _, err := os.Lstat(linkDir); err == nil {
			RemoveLink(linkDir)
		}

		if _, err := os.Stat(wikiDir); err == nil {
			if err := CreateLink(wikiDir, linkDir); err != nil {
				log.Warn("创建 wiki 链接失败", zap.String("repo", rd.RefName), zap.Error(err))
			} else {
				linked = append(linked, rd.RefName)
			}
		}
	}
	return linked
}

func scanWikiTopics(wikiDir string) ([]mapTopicEntry, string) {
	var topics []mapTopicEntry
	var refCommit string

	entries, err := os.ReadDir(wikiDir)
	if err != nil {
		return nil, ""
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := e.Name()
		data, err := os.ReadFile(filepath.Join(wikiDir, name))
		if err != nil {
			continue
		}
		text := string(data)
		desc, commit := parseFrontmatter(text)
		if commit != "" && refCommit == "" {
			refCommit = commit
		}

		topicName := strings.TrimSuffix(name, ".md")
		if topicName == "reference" || topicName == "scc" {
			continue
		}
		if desc == "" {
			desc = topicName
		}
		topics = append(topics, mapTopicEntry{
			File:        topicName + ".md",
			Description: desc,
			Commit:      commit,
		})
	}
	return topics, refCommit
}

func parseFrontmatter(text string) (description, commit string) {
	if !strings.HasPrefix(text, "---") {
		return "", ""
	}
	end := strings.Index(text[3:], "\n---")
	if end < 0 {
		return "", ""
	}
	fm := text[3 : 3+end]
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description:") {
			description = strings.TrimPrefix(line, "description:")
			description = strings.TrimSpace(description)
		}
		if strings.HasPrefix(line, "commit:") {
			commit = strings.TrimPrefix(line, "commit:")
			commit = strings.TrimSpace(commit)
		}
	}
	return description, commit
}

func generateReferenceMap(refDir string, repos []repoData) error {
	var buf bytes.Buffer
	for _, r := range repos {
		entry := mapRepoEntry{
			RefName:  r.RefName,
			Type:     r.Type,
			Platform: r.Platform,
			FullName: r.FullName,
			Desc:     r.Description,
			RepoPath: filepath.Join(".reference", "repos", r.RefName),
			WikiPath: filepath.Join(".reference", "wiki", r.RefName),
		}
		entry.Topics, entry.Commit = scanWikiTopics(r.WikiDir)
		line, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	return os.WriteFile(filepath.Join(refDir, "reference.map.jsonl"), buf.Bytes(), 0644)
}

func generateWikiReference(wikiDir, repoPath string, r *models.Repo) error {
	if err := os.MkdirAll(wikiDir, 0755); err != nil {
		return err
	}

	shortCommit := r.Commit
	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}

	repoID := repoIdentifier(r)
	today := time.Now().Format("2006-01-02")

	refFile := filepath.Join(wikiDir, "reference.md")
	if _, err := os.Stat(refFile); os.IsNotExist(err) {
		description := detectDescription(repoPath, r)
		language := detectLanguage(repoPath)

		refFrontmatter := fmt.Sprintf("---\nrepo: %s\ncommit: %s\nbranch: %s\ndescription: 仓库架构总览\nexplored_at: %s\n---\n\n",
			repoID, shortCommit, r.Branch, today)

		var sb strings.Builder
		refName := r.GetRefName()
		sb.WriteString(fmt.Sprintf("# %s\n\n", refName))
		if r.RefType == models.RefTypeRemote {
			sb.WriteString(fmt.Sprintf("- **仓库**: %s/%s\n", r.Host, r.Namespace+"/"+r.RepoName))
		} else {
			sb.WriteString(fmt.Sprintf("- **路径**: %s\n", r.LocalPath))
		}
		sb.WriteString(fmt.Sprintf("- **描述**: %s\n", description))
		sb.WriteString(fmt.Sprintf("- **语言**: %s\n", language))
		if r.CommitAt != nil {
			sb.WriteString(fmt.Sprintf("- **更新**: %s\n", r.CommitAt.Format("2006-01-02")))
		}
		return os.WriteFile(refFile, []byte(refFrontmatter+sb.String()), 0644)
	}

	return nil
}

func repoIdentifier(r *models.Repo) string {
	if r.RefType == models.RefTypeRemote {
		return fmt.Sprintf("%s/%s/%s", r.Host, r.Namespace, r.RepoName)
	}
	return fmt.Sprintf("local/%s", filepath.Base(r.LocalPath))
}

func detectLanguage(repoPath string) string {
	indicators := map[string]string{
		"go.mod":           "Go",
		"package.json":     "JavaScript/TypeScript",
		"pom.xml":          "Java",
		"build.gradle":     "Java/Kotlin",
		"Cargo.toml":       "Rust",
		"pyproject.toml":   "Python",
		"requirements.txt": "Python",
		"setup.py":         "Python",
		"Gemfile":          "Ruby",
		"composer.json":    "PHP",
		"CMakeLists.txt":   "C/C++",
		"Makefile":         "C/C++",
	}

	for file, lang := range indicators {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			return lang
		}
	}
	return "Unknown"
}

func renderSkill(templateData []byte, outputPath string) error {
	tmpl, err := template.New("skill").Parse(string(templateData))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{}{}); err != nil {
		return err
	}
	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}


func readEmbedded(embedPath string) ([]byte, error) {
	data, err := common.PromptsFS.ReadFile(embedPath)
	if err == nil {
		return data, nil
	}
	fallback := filepath.Join(utils.GetExeDir(), "..", "..", embedPath)
	return os.ReadFile(fallback)
}

func extractEmbedded(embedPath, dstPath string) error {
	data, err := readEmbedded(embedPath)
	if err != nil {
		return err
	}
	return os.WriteFile(dstPath, data, 0644)
}

func detectDescription(repoPath string, r *models.Repo) string {
	readmePath := filepath.Join(repoPath, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "!") && !strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "<") {
			if len(line) > 100 {
				return line[:100] + "..."
			}
			return line
		}
	}
	return ""
}

func cleanStaleJunctions(dir string, activeRepos []repoData) {
	activeSet := make(map[string]bool, len(activeRepos))
	for _, r := range activeRepos {
		activeSet[r.RefName] = true
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		fullPath := filepath.Join(dir, e.Name())
		info, err := os.Lstat(fullPath)
		if err != nil || (!info.IsDir() && info.Mode()&os.ModeSymlink == 0) {
			continue
		}
		if !activeSet[e.Name()] {
			RemoveLink(fullPath)
			log.Info("清理残留 junction", zap.String("name", e.Name()))
		}
	}
}
