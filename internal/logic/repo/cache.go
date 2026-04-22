package repo

import (
	"fmt"
	"os"
	"time"

	"github.com/cicbyte/reference/internal/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"go.uber.org/zap"
)

type CloneOptions struct {
	URL      string
	Path     string
	Branch   string
	Depth    int
	Proxy    string
	NoUpdate bool
}

func CloneOrUpdate(opts CloneOptions) error {
	cleanup, err := SetupGitProxy(opts.Proxy)
	if err != nil {
		return fmt.Errorf("设置代理失败: %w", err)
	}
	defer cleanup()

	if _, err := os.Stat(opts.Path); os.IsNotExist(err) {
		return cloneRepo(opts)
	}
	if !opts.NoUpdate {
		return pullRepo(opts)
	}
	log.Info("缓存已存在，跳过更新", zap.String("path", opts.Path))
	return nil
}

func cloneRepo(opts CloneOptions) error {
	if err := os.MkdirAll(opts.Path, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:          opts.URL,
		Depth:        opts.Depth,
		SingleBranch: true,
	}

	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
	}

	log.Info("正在克隆仓库", zap.String("url", opts.URL), zap.String("path", opts.Path))
	_, err := git.PlainClone(opts.Path, false, cloneOpts)
	if err != nil {
		os.RemoveAll(opts.Path)
		return fmt.Errorf("克隆失败: %w", err)
	}
	log.Info("克隆完成", zap.String("path", opts.Path))
	return nil
}

func pullRepo(opts CloneOptions) error {
	repo, err := git.PlainOpen(opts.Path)
	if err != nil {
		log.Warn("缓存目录无效，重新克隆", zap.String("path", opts.Path))
		os.RemoveAll(opts.Path)
		return cloneRepo(opts)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作树失败: %w", err)
	}

	pullOpts := &git.PullOptions{
		SingleBranch: true,
		RemoteName:   "origin",
	}

	log.Info("正在更新仓库", zap.String("path", opts.Path))
	err = wt.Pull(pullOpts)
	if err == git.NoErrAlreadyUpToDate {
		log.Info("仓库已是最新", zap.String("path", opts.Path))
		return nil
	}
	if err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	log.Info("更新完成", zap.String("path", opts.Path))
	return nil
}

func GetRepoMeta(repoPath string) (branch, commit string, commitTime *time.Time, err error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", "", nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", "", nil, err
	}
	commit = ref.Hash().String()[:7]
	branch = ref.Name().Short()

	obj, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return branch, commit, nil, nil
	}
	t := obj.Author.When
	commitTime = &t

	return branch, commit, commitTime, nil
}

func ValidateLocalRepo(path string) error {
	_, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("不是有效的 Git 仓库: %s", path)
	}
	return nil
}

func PurgeCache(path string) error {
	return os.RemoveAll(path)
}

func GetRemoteURL(repoPath string) (string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", err
	}
	cfg, err := repo.Config()
	if err != nil {
		return "", err
	}
	if r, ok := cfg.Remotes["origin"]; ok && len(r.URLs) > 0 {
		return r.URLs[0], nil
	}
	return "", fmt.Errorf("未找到 origin 远程地址")
}

func EnsureRemote(repoPath, remoteURL string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}
	_, err = repo.Remote("origin")
	if err == git.ErrRemoteNotFound {
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{remoteURL},
		})
	}
	return err
}
