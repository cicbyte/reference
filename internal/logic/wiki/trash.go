package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DeletedFile struct {
	Path    string `json:"path"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Message string `json:"message"`
}

func ListDeletedFiles(wikiDir string, limit int) ([]DeletedFile, error) {
	args := []string{"log", "--diff-filter=D", "--name-only", "--pretty=format:%h|%ai|%s"}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-%d", limit))
	}
	output, err := gitExec(wikiDir, args...)
	if err != nil {
		return nil, fmt.Errorf("查询删除记录失败: %w", err)
	}

	var files []DeletedFile
	var current DeletedFile

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "|") {
			parts := strings.SplitN(line, "|", 3)
			if len(parts) >= 3 {
				current = DeletedFile{
					Commit:  parts[0],
					Date:    strings.TrimSpace(parts[1]),
					Message: parts[2],
				}
			}
		} else {
			current.Path = line
			files = append(files, current)
		}
	}

	return files, nil
}

func RestoreFile(wikiDir, filePath string) error {
	clean := filepath.Clean(filePath)
	if filepath.IsAbs(clean) || strings.Contains(clean, "..") {
		return fmt.Errorf("无效的文件路径: %s", filePath)
	}

	output, err := gitExec(wikiDir, "log", "--diff-filter=D", "--format=%h", "--", filePath)
	if err != nil || strings.TrimSpace(output) == "" {
		return fmt.Errorf("未找到删除记录: %s", filePath)
	}

	commit := strings.TrimSpace(strings.Split(output, "\n")[0])

	fullPath := filepath.Join(wikiDir, filePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	_, err = gitExec(wikiDir, "checkout", commit+"~1", "--", filePath)
	if err != nil {
		return fmt.Errorf("恢复失败: %w", err)
	}

	return nil
}
