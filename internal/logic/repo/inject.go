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
	RefName  string          `json:"ref_name"`
	Type     string          `json:"type"`
	Platform string          `json:"platform,omitempty"`
	FullName string          `json:"full_name"`
	Desc     string          `json:"description,omitempty"`
	RepoPath string          `json:"repo_path"`
	WikiPath string          `json:"wiki_path"`
	Commit   string          `json:"commit,omitempty"`
	Topics   []mapTopicEntry `json:"topics,omitempty"`
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

	if err := EnsureGitignore(p.config.ProjectDir); err != nil {
		log.Warn("创建 .gitignore 失败", zap.Error(err))
	}

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
		wikiDir := resolveWikiDir(&r)
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

	settings := models.LoadProjectSettings(p.config.ProjectDir)
	var injectedFiles []string
	for _, agentID := range settings.Agents {
		cfg, ok := GetAgentConfig(agentID)
		if !ok {
			continue
		}
		baseDir := filepath.Join(p.config.ProjectDir, cfg.BaseDir)
		files := p.injectAgentFiles(agentID, baseDir, cfg.Files)
		injectedFiles = append(injectedFiles, files...)
	}

	total := len(injectedFiles) + len(wikiFiles)
	if total == 0 && repairCount == 0 {
		return "配置已是最新。", nil
	}

	var sb strings.Builder
	if len(wikiFiles) > 0 {
		sb.WriteString(fmt.Sprintf("已链接 %d 个仓库知识", len(wikiFiles)))
	}
	if len(injectedFiles) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("，")
		}
		sb.WriteString(fmt.Sprintf("已更新 %d 个 AI 配置文件", len(injectedFiles)))
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

func (p *InjectProcessor) injectAgentFiles(agentID, baseDir string, files []AgentFile) []string {
	var updated []string
	for _, f := range files {
		dst := filepath.Join(baseDir, f.DestPath)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			log.Warn("创建目录失败", zap.String("path", filepath.Dir(dst)), zap.Error(err))
			continue
		}
		data, err := readEmbedded(f.EmbedPath)
		if err != nil {
			log.Warn("读取嵌入资源失败", zap.String("file", f.EmbedPath), zap.Error(err))
			continue
		}
		// Transform frontmatter for platforms whose agent format differs from
		// the canonical Claude Code template.
		switch agentID {
		case "mimocode":
			data = transformMimocodeFrontmatter(data)
		case "opencode":
			data = transformOpencodeFrontmatter(data)
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			log.Warn("注入文件失败", zap.String("file", f.DestPath), zap.Error(err))
		} else {
			updated = append(updated, filepath.Base(f.DestPath))
		}
	}
	return updated
}

// transformMimocodeFrontmatter rewrites the agent frontmatter from the
// canonical "tools: Read, Grep, Glob, Bash, Write" (a comma-separated tool
// whitelist, as used by Claude Code / ZCode / OpenCode) into MiMo Code's
// boolean permission flags:
//
//   tools:
//     read: true
//     grep: true
//     glob: true
//     bash: true
//     write: true
//     edit: false
//
// Also adds `mode: subagent` so the agent runs in an isolated context (MiMo
// Code supports this natively; other platforms ignore the field).
func transformMimocodeFrontmatter(data []byte) []byte {
	content := string(data)
	// find the closing frontmatter delimiter
	end := strings.Index(content[3:], "\n---")
	if !strings.HasPrefix(content, "---") || end < 0 {
		return data // no valid frontmatter — return as-is
	}
	fmEnd := 3 + end + 4 // position after closing "---\n"
	fm := content[:fmEnd]
	body := content[fmEnd:]

	// Extract the tools line and convert
	toolsLine := ""
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "tools:") {
			toolsLine = strings.TrimSpace(line)
			break
		}
	}

	// Parse comma-separated tool names into a set
	toolSet := map[string]bool{}
	if toolsLine != "" {
		parts := strings.Split(strings.TrimPrefix(toolsLine, "tools:"), ",")
		for _, p := range parts {
			toolSet[strings.ToLower(strings.TrimSpace(p))] = true
		}
	}

	// Build MiMo Code boolean permission block.
	// The canonical template grants: Read, Grep, Glob, Bash, Write
	// Map each to MiMo's permission keys; default unlisted tools to false.
	mimoTools := "\ntools:\n" +
		fmt.Sprintf("  read: %v\n", toolSet["read"]) +
		fmt.Sprintf("  grep: %v\n", toolSet["grep"]) +
		fmt.Sprintf("  glob: %v\n", toolSet["glob"]) +
		fmt.Sprintf("  bash: %v\n", toolSet["bash"]) +
		fmt.Sprintf("  write: %v\n", toolSet["write"]) +
		fmt.Sprintf("  edit: %v\n", toolSet["edit"])

	// Replace the old "tools: ..." line with the new block
	var newFm []string
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "tools:") {
			newFm = append(newFm, strings.TrimSuffix(mimoTools, "\n"))
		} else {
			newFm = append(newFm, line)
		}
	}
	result := strings.Join(newFm, "\n")

	// Add mode: subagent before the closing --- (if not already present)
	if !strings.Contains(result, "mode:") {
		result = strings.TrimSuffix(result, "---") + "mode: subagent\n---"
	}

	return []byte(result + body)
}

// transformOpencodeFrontmatter rewrites the agent frontmatter from the
// canonical Claude Code format into OpenCode's native format.
//
// Key differences from the template:
//   - OpenCode has NO `name` field (the filename is the agent name).
//   - `tools: Read, Grep, ...` (comma list) → `permission:` block with
//     allow/deny tri-state values (OpenCode's `tools` field is deprecated).
//   - `mode: subagent` is required for context isolation.
//
// Example output:
//
//   ---
//   description: ...
//   mode: subagent
//   permission:
//     read: allow
//     grep: allow
//     glob: allow
//     bash: allow
//     write: allow
//     edit: deny
//   ---
func transformOpencodeFrontmatter(data []byte) []byte {
	content := string(data)
	end := strings.Index(content[3:], "\n---")
	if !strings.HasPrefix(content, "---") || end < 0 {
		return data
	}
	fmEnd := 3 + end + 4
	fm := content[:fmEnd]
	body := content[fmEnd:]

	// Parse the tools list into a set (same logic as MiMo transform)
	toolSet := map[string]bool{}
	var descLine string
	for _, line := range strings.Split(fm, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "tools:") {
			parts := strings.Split(strings.TrimPrefix(trimmed, "tools:"), ",")
			for _, p := range parts {
				toolSet[strings.ToLower(strings.TrimSpace(p))] = true
			}
		}
		if strings.HasPrefix(trimmed, "description:") {
			descLine = line
		}
	}

	// Build the OpenCode frontmatter: drop `name` and `tools`, keep `description`,
	// add `mode: subagent` + `permission:` block.
	var b strings.Builder
	b.WriteString("---\n")
	if descLine != "" {
		b.WriteString(descLine + "\n")
	}
	b.WriteString("mode: subagent\n")
	b.WriteString("permission:\n")
	// OpenCode permission keys mapped from our tool names.
	// Tools in the set get "allow", everything else gets "deny".
	for _, key := range []string{"read", "grep", "glob", "bash", "write", "edit"} {
		val := "deny"
		if toolSet[key] {
			val = "allow"
		}
		b.WriteString(fmt.Sprintf("  %s: %s\n", key, val))
	}
	b.WriteString("---\n")

	return []byte(b.String() + body)
}

func (p *InjectProcessor) injectWikiJunctions(wikiJunctionDir string, repos []repoData) []string {
	cleanStaleJunctions(wikiJunctionDir, repos)

	var linked []string
	for _, rd := range repos {
		wikiDir := rd.WikiDir
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

func resolveWikiDir(r *models.Repo) string {
	if r.RefType == models.RefTypeLocal {
		return filepath.Join(utils.ConfigInstance.GetLocalWikiDir(), r.WikiSubPath)
	}
	return filepath.Join(utils.ConfigInstance.GetWikiDir(), r.WikiSubPath)
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
