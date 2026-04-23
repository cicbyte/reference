package wiki

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cicbyte/reference/internal/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
)

type CommitResult struct {
	HasChanges bool
	CommitHash string
	Message    string
}

func StageAndCommit(wikiDir, message string) (*CommitResult, error) {
	if HasRemote(wikiDir) {
		pullBeforeCommit(wikiDir)
	}

	repo, err := git.PlainOpen(wikiDir)
	if err != nil {
		return nil, fmt.Errorf("wiki 仓库无效: %w", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("获取工作树失败: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("检查状态失败: %w", err)
	}
	if status.IsClean() {
		return &CommitResult{HasChanges: false}, nil
	}

	if message == "" {
		message = buildAutoMessage(status, wikiDir)
		if message == "" {
			return &CommitResult{HasChanges: false}, nil
		}
	}

	_, err = wt.Add(".")
	if err != nil {
		return nil, fmt.Errorf("暂存失败: %w", err)
	}

	name, email := getGitUser(wikiDir)
	hash, err := wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("提交失败: %w", err)
	}

	log.Info("wiki 已提交", zap.String("hash", hash.String()[:7]), zap.String("message", message))
	return &CommitResult{HasChanges: true, CommitHash: hash.String()[:7], Message: message}, nil
}

func Push(wikiDir string) error {
	if !HasRemote(wikiDir) {
		return fmt.Errorf("未配置远程仓库，请先运行 reference wiki remote <url>")
	}
	branch := getDefaultBranch(wikiDir)
	_, err := gitExec(wikiDir, "push", "origin", branch)
	if err != nil {
		return fmt.Errorf("推送失败: %w", err)
	}
	log.Info("wiki 已推送到远程", zap.String("path", wikiDir))
	return nil
}

// gitExec 在 wikiDir 目录下执行 git 命令，继承系统凭证管理
func gitExec(wikiDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = wikiDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func getGitUser(wikiDir string) (name, email string) {
	if n, err := gitExec(wikiDir, "config", "user.name"); err == nil {
		name = strings.TrimSpace(n)
	}
	if e, err := gitExec(wikiDir, "config", "user.email"); err == nil {
		email = strings.TrimSpace(e)
	}
	if name == "" {
		name = "reference"
	}
	if email == "" {
		email = "reference@local"
	}
	return name, email
}

func isAlreadyUpToDate(output string) bool {
	return strings.Contains(output, "Already up to date") ||
		strings.Contains(output, "Everything up-to-date")
}

func pullBeforeCommit(wikiDir string) {
	branch := getDefaultBranch(wikiDir)
	output, err := gitExec(wikiDir, "pull", "origin", branch)
	if err != nil {
		if !isAlreadyUpToDate(output) {
			log.Warn("wiki commit: 拉取失败", zap.String("output", output))
		}
		return
	}
	log.Info("wiki commit: 已同步远程更新")
}

func buildAutoMessage(status git.Status, baseDir string) string {
	var added, modified, deleted []string
	for path, fs := range status {
		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			rel = path
		}
		clean := filepath.ToSlash(rel)
		switch fs.Worktree {
		case git.Untracked, git.Added:
			added = append(added, clean)
		case git.Modified:
			modified = append(modified, clean)
		case git.Deleted:
			deleted = append(deleted, clean)
		}
	}
	sort.Strings(added)
	sort.Strings(modified)
	sort.Strings(deleted)

	if len(added) == 0 && len(modified) == 0 && len(deleted) == 0 {
		return ""
	}

	var parts []string
	if len(added) > 0 {
		parts = append(parts, fmt.Sprintf("add %s", joinPaths(added)))
	}
	if len(modified) > 0 {
		parts = append(parts, fmt.Sprintf("update %s", joinPaths(modified)))
	}
	if len(deleted) > 0 {
		parts = append(parts, fmt.Sprintf("delete %s", joinPaths(deleted)))
	}
	return "wiki: " + strings.Join(parts, ", ")

}

func joinPaths(paths []string) string {
	if len(paths) <= 3 {
		return strings.Join(paths, ", ")
	}
	return strings.Join(paths[:3], ", ") + fmt.Sprintf(" and %d more", len(paths)-3)
}
