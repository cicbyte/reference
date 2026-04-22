package wiki

import (
	"fmt"
	"strings"

	"github.com/cicbyte/reference/internal/log"
	"github.com/go-git/go-git/v5"
	"go.uber.org/zap"
)

type SyncResult struct {
	PullOK  bool
	PushOK  bool
	Commit  *CommitResult
	PullErr string
	PushErr string
}

func Sync(wikiDir string) (*SyncResult, error) {
	result := &SyncResult{}

	if !HasRemote(wikiDir) {
		return nil, fmt.Errorf("未配置远程仓库，请先运行 reference wiki remote <url>")
	}

	// pull 使用 git CLI（继承系统凭证）
	branch := getDefaultBranch(wikiDir)
	output, err := gitExec(wikiDir, "pull", "origin", branch)
	if err != nil {
		if isAlreadyUpToDate(output) {
			result.PullOK = true
		} else {
			result.PullErr = output
			log.Warn("wiki sync: 拉取失败", zap.String("output", output))
			return result, nil
		}
	} else {
		result.PullOK = true
		log.Info("wiki sync: 拉取成功")
	}

	commitResult, err := StageAndCommit(wikiDir, "")
	if err != nil {
		log.Warn("wiki sync: 自动提交失败", zap.Error(err))
	} else {
		result.Commit = commitResult
	}

	if err := Push(wikiDir); err != nil {
		result.PushErr = err.Error()
	} else {
		result.PushOK = true
	}

	return result, nil
}

func GetWikiStatus(wikiDir string) string {
	var lines []string

	if !IsGitInitialized(wikiDir) {
		return "  wiki 仓库未初始化"
	}

	lines = append(lines, "  wiki 仓库: 已初始化")

	if url, err := GetRemoteURL(wikiDir); err == nil && url != "" {
		lines = append(lines, fmt.Sprintf("  远程仓库: %s", url))
	} else {
		lines = append(lines, "  远程仓库: (未设置)")
	}

	repo, err := git.PlainOpen(wikiDir)
	if err == nil {
		wt, err := repo.Worktree()
		if err == nil {
			status, err := wt.Status()
			if err == nil && !status.IsClean() {
				lines = append(lines, "  工作区: 有未提交的更改")
			} else if err == nil {
				lines = append(lines, "  工作区: 干净")
			}
		}
	}

	return strings.Join(lines, "\n") + "\n"
}
