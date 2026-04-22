package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/reference/internal/log"
	"github.com/go-git/go-git/v5"
	"go.uber.org/zap"
)

func IsGitInitialized(wikiDir string) bool {
	_, err := git.PlainOpen(wikiDir)
	return err == nil
}

func EnsureGitInit(wikiDir string) error {
	if IsGitInitialized(wikiDir) {
		return nil
	}
	if err := os.MkdirAll(wikiDir, 0755); err != nil {
		return err
	}
	_, err := git.PlainInit(wikiDir, false)
	if err != nil {
		return err
	}
	// 写入 .gitignore
	gitignorePath := filepath.Join(wikiDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(".watch.pid\n"), 0644); err != nil {
			return fmt.Errorf("写入 .gitignore 失败: %w", err)
		}
	}
	log.Info("wiki 仓库已初始化", zap.String("path", wikiDir))
	return nil
}

func HasRemote(wikiDir string) bool {
	repo, err := git.PlainOpen(wikiDir)
	if err != nil {
		return false
	}
	_, err = repo.Remote("origin")
	return err == nil
}

func EnsureAutoPull(wikiDir string) {
	if !HasRemote(wikiDir) {
		return
	}
	branch := getDefaultBranch(wikiDir)
	output, err := gitExec(wikiDir, "pull", "origin", branch)
	if err != nil {
		if isAlreadyUpToDate(output) {
			return
		}
		log.Warn("wiki 自动拉取失败（可稍后运行 reference wiki sync 修复）",
			zap.String("path", wikiDir), zap.Error(err))
		return
	}
	log.Info("wiki 已自动同步", zap.String("path", wikiDir))
}

func getDefaultBranch(wikiDir string) string {
	output, err := gitExec(wikiDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "master"
	}
	branch := strings.TrimSpace(output)
	if branch == "" || branch == "HEAD" {
		return "master"
	}
	return branch
}
