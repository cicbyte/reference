package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cicbyte/reference/internal/log"
	"go.uber.org/zap"
)

func CreateLink(target, linkPath string) error {
	if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
		return fmt.Errorf("创建链接目录失败: %w", err)
	}

	if _, err := os.Lstat(linkPath); err == nil {
		removeLink(linkPath)
	}

	if runtime.GOOS == "windows" {
		target, _ = filepath.Abs(target)
		linkPath, _ = filepath.Abs(linkPath)
		wTarget := toWindowsPath(target)
		wLink := toWindowsPath(linkPath)

		cmdArgs := []string{"/c", "mklink", "/J", wLink, wTarget}
		log.Debug("创建 Junction", zap.Strings("cmd", cmdArgs))
		out, err := exec.Command("cmd", cmdArgs...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("创建 Junction 失败: %s\n%s", err, string(out))
		}
	} else {
		if err := os.Symlink(target, linkPath); err != nil {
			return fmt.Errorf("创建符号链接失败: %w", err)
		}
	}

	return nil
}

func RemoveLink(linkPath string) error {
	linkPath, _ = filepath.Abs(linkPath)
	return removeLink(linkPath)
}

func removeLink(linkPath string) error {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("cmd", "/c", "rmdir", toWindowsPath(linkPath)).CombinedOutput()
		if err != nil {
			return fmt.Errorf("删除 Junction 失败: %s\n%s", err, string(out))
		}
		return nil
	}
	return os.Remove(linkPath)
}

func ReadLink(linkPath string) (string, error) {
	if runtime.GOOS == "windows" {
		linkPath, _ = filepath.Abs(linkPath)
		out, err := exec.Command("cmd", "/c", "fsutil", "reparsepoint", "query", toWindowsPath(linkPath)).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("读取链接目标失败: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}
	return os.Readlink(linkPath)
}

func toWindowsPath(p string) string {
	return strings.ReplaceAll(p, "/", "\\")
}
