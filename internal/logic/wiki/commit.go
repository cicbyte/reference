package wiki

import (
	"fmt"
	"os/exec"
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
	repo, err := git.PlainOpen(wikiDir)
	if err != nil {
		return nil, fmt.Errorf("wiki 仓库无效: %w", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("获取工作树失败: %w", err)
	}

	_, err = wt.Add(".")
	if err != nil {
		return nil, fmt.Errorf("暂存失败: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("检查状态失败: %w", err)
	}
	if status.IsClean() {
		return &CommitResult{HasChanges: false}, nil
	}

	if message == "" {
		message = fmt.Sprintf("wiki: auto-commit %s", time.Now().Format("2006-01-02 15:04:05"))
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
