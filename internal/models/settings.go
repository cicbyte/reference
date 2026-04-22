package models

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProjectSettings struct {
	Agent       string `json:"agent"`
	Initialized bool   `json:"initialized"`
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
