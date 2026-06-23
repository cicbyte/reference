package models

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProjectSettings struct {
	Agents      []string `json:"agents,omitempty"`
	Agent       string   `json:"agent,omitempty"` // 已废弃，仅兼容旧格式读取
	Initialized bool     `json:"initialized"`
}

// MigrateAgent 兼容旧格式：将 agent 字段迁移到 agents 数组
func (s *ProjectSettings) MigrateAgent() {
	if s.Agent != "" && len(s.Agents) == 0 {
		s.Agents = []string{s.Agent}
	}
	s.Agent = ""
}

// HasAgent 检查是否配置了指定 agent
func (s *ProjectSettings) HasAgent(id string) bool {
	for _, a := range s.Agents {
		if a == id {
			return true
		}
	}
	return false
}

func settingsPath(projectDir string) string {
	return filepath.Join(projectDir, ".reference", "reference.settings.json")
}

func LoadProjectSettings(projectDir string) *ProjectSettings {
	data, err := os.ReadFile(settingsPath(projectDir))
	if err != nil {
		return &ProjectSettings{}
	}
	var s ProjectSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return &ProjectSettings{}
	}
	s.MigrateAgent()
	return &s
}

func SaveProjectSettings(projectDir string, s *ProjectSettings) error {
	if err := os.MkdirAll(filepath.Dir(settingsPath(projectDir)), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath(projectDir), data, 0644)
}
