package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func EnsureDir(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat directory: %v", err)
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func GetProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func GetExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// 初始化应用目录结构
func InitAppDirs() error {
	config := ConfigInstance

	dirs := []string{
		config.GetAppSeriesDir(),
		config.GetAppDir(),
		config.GetConfigDir(),
		config.GetDbDir(),
		config.GetLogDir(),
		config.GetReposDir(),
		config.GetWikiDir(),
	}

	for _, dir := range dirs {
		if err := EnsureDir(dir); err != nil {
			return fmt.Errorf("directory init failed: %v", err)
		}
	}

	return nil
}
