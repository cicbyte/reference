package repo

import (
	"path/filepath"
	"sort"
	"strings"
)

// AgentFile 描述一个需要注入的 agent 文件
type AgentFile struct {
	EmbedPath string // 嵌入资源路径，如 "prompts/agents/reference-explorer.md"
	DestPath  string // 相对于 BaseDir 的目标路径，如 "agents/reference-explorer.md"
}

// AgentConfig 描述一个 AI Agent 的注入配置
type AgentConfig struct {
	ID          string      // 唯一标识，如 "claude"
	DisplayName string      // 显示名称，如 "Claude Code"
	BaseDir     string      // 相对于项目根目录的配置目录，如 ".claude"
	Files       []AgentFile // 需要注入的文件列表
}

// AgentRegistry 所有已注册的 Agent 配置
var AgentRegistry = map[string]AgentConfig{
	"claude": {
		ID:          "claude",
		DisplayName: "Claude Code",
		BaseDir:     ".claude",
		Files: []AgentFile{
			{EmbedPath: "prompts/agents/reference-explorer.md", DestPath: filepath.Join("agents", "reference-explorer.md")},
			{EmbedPath: "prompts/agents/reference-analyzer.md", DestPath: filepath.Join("agents", "reference-analyzer.md")},
			{EmbedPath: "prompts/skills/reference/SKILL.md", DestPath: filepath.Join("skills", "reference", "SKILL.md")},
		},
	},
	"zcode": {
		ID:          "zcode",
		DisplayName: "ZCode",
		BaseDir:     ".zcode",
		Files: []AgentFile{
			{EmbedPath: "prompts/agents/reference-explorer.md", DestPath: filepath.Join("cli", "agents", "reference-explorer.md")},
			{EmbedPath: "prompts/agents/reference-analyzer.md", DestPath: filepath.Join("cli", "agents", "reference-analyzer.md")},
			{EmbedPath: "prompts/skills/reference/SKILL.md", DestPath: filepath.Join("skills", "reference", "SKILL.md")},
		},
	},
	"mimocode": {
		ID:          "mimocode",
		DisplayName: "MiMo Code",
		BaseDir:     ".mimocode",
		Files: []AgentFile{
			{EmbedPath: "prompts/agents/reference-explorer.md", DestPath: filepath.Join("agents", "reference-explorer.md")},
			{EmbedPath: "prompts/agents/reference-analyzer.md", DestPath: filepath.Join("agents", "reference-analyzer.md")},
			{EmbedPath: "prompts/skills/reference/SKILL.md", DestPath: filepath.Join("skills", "reference", "SKILL.md")},
		},
	},
	"opencode": {
		ID:          "opencode",
		DisplayName: "OpenCode",
		BaseDir:     ".opencode",
		Files: []AgentFile{
			{EmbedPath: "prompts/agents/reference-explorer.md", DestPath: filepath.Join("agents", "reference-explorer.md")},
			{EmbedPath: "prompts/agents/reference-analyzer.md", DestPath: filepath.Join("agents", "reference-analyzer.md")},
			{EmbedPath: "prompts/skills/reference/SKILL.md", DestPath: filepath.Join("skills", "reference", "SKILL.md")},
		},
	},
	"codex": {
		ID:          "codex",
		DisplayName: "Codex",
		BaseDir:     ".codex",
		Files: []AgentFile{
			{EmbedPath: "prompts/agents/reference-explorer.md", DestPath: filepath.Join("agents", "reference-explorer.md")},
			{EmbedPath: "prompts/agents/reference-analyzer.md", DestPath: filepath.Join("agents", "reference-analyzer.md")},
			{EmbedPath: "prompts/skills/reference/SKILL.md", DestPath: filepath.Join("skills", "reference", "SKILL.md")},
		},
	},
	"gemini": {
		ID:          "gemini",
		DisplayName: "Gemini CLI",
		BaseDir:     ".gemini",
		Files: []AgentFile{
			{EmbedPath: "prompts/agents/reference-explorer.md", DestPath: filepath.Join("agents", "reference-explorer.md")},
			{EmbedPath: "prompts/agents/reference-analyzer.md", DestPath: filepath.Join("agents", "reference-analyzer.md")},
			{EmbedPath: "prompts/skills/reference/SKILL.md", DestPath: filepath.Join("skills", "reference", "SKILL.md")},
		},
	},
	"qwen": {
		ID:          "qwen",
		DisplayName: "Qwen Code",
		BaseDir:     ".qwen",
		Files: []AgentFile{
			{EmbedPath: "prompts/agents/reference-explorer.md", DestPath: filepath.Join("agents", "reference-explorer.md")},
			{EmbedPath: "prompts/agents/reference-analyzer.md", DestPath: filepath.Join("agents", "reference-analyzer.md")},
			{EmbedPath: "prompts/skills/reference/SKILL.md", DestPath: filepath.Join("skills", "reference", "SKILL.md")},
		},
	},
	// kimi 仅注入 Skill：kimi-cli 的子代理是内置类型系统（coder/explore/plan，
	// YAML 类型定义），未提供项目级自定义子代理目录；`.kimi/skills/` 为已确认
	// 的项目级技能发现路径，SKILL 会引导主代理用内置 explore 类型完成探索。
	"kimi": {
		ID:          "kimi",
		DisplayName: "Kimi CLI",
		BaseDir:     ".kimi",
		Files: []AgentFile{
			{EmbedPath: "prompts/skills/reference/SKILL.md", DestPath: filepath.Join("skills", "reference", "SKILL.md")},
		},
	},
	// iflow 仅注入 agents：iFlow 通过 frontmatter 的 whenToUse 智能调度子代理，
	// 无需 Skill 文件；自定义 Agent 目录 `.iflow/agents/` 已确认。
	"iflow": {
		ID:          "iflow",
		DisplayName: "iFlow CLI",
		BaseDir:     ".iflow",
		Files: []AgentFile{
			{EmbedPath: "prompts/agents/reference-explorer.md", DestPath: filepath.Join("agents", "reference-explorer.md")},
			{EmbedPath: "prompts/agents/reference-analyzer.md", DestPath: filepath.Join("agents", "reference-analyzer.md")},
		},
	},
}

// GetAgentConfig 根据 ID 获取 agent 配置
func GetAgentConfig(id string) (AgentConfig, bool) {
	cfg, ok := AgentRegistry[id]
	return cfg, ok
}

// GetAgentDisplayName 获取 agent 显示名称，未注册则返回原始 ID
func GetAgentDisplayName(id string) string {
	if cfg, ok := AgentRegistry[id]; ok {
		return cfg.DisplayName
	}
	return id
}

// IsAgentRegistered 检查 agent 是否已注册
func IsAgentRegistered(id string) bool {
	_, ok := AgentRegistry[id]
	return ok
}

// ListAgentIDs 返回所有已注册的 agent ID 列表（按 ID 字母序，保证菜单与提示顺序稳定）
func ListAgentIDs() []string {
	ids := make([]string, 0, len(AgentRegistry))
	for id := range AgentRegistry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ValidateAgentIDs 验证 agent ID 列表，返回无效的 ID
func ValidateAgentIDs(ids []string) []string {
	var invalid []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" && !IsAgentRegistered(id) {
			invalid = append(invalid, id)
		}
	}
	return invalid
}
