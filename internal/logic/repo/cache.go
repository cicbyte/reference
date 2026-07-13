package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cicbyte/reference/internal/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"go.uber.org/zap"
)

type CloneOptions struct {
	URL    string
	Path   string
	Branch string
	Proxy  string
	Update bool
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
	if opts.Update {
		return pullRepo(opts)
	}
	log.Info("缓存已存在，跳过更新（使用 --update 强制更新）", zap.String("path", opts.Path))
	return nil
}

func cloneRepo(opts CloneOptions) error {
	if err := os.MkdirAll(opts.Path, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL: opts.URL,
	}

	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
		cloneOpts.SingleBranch = true
	}

	log.Info("正在克隆仓库", zap.String("url", opts.URL), zap.String("path", opts.Path))
	_, err := git.PlainClone(opts.Path, false, cloneOpts)
	if err == nil {
		log.Info("克隆完成", zap.String("path", opts.Path))
		return nil
	}

	os.RemoveAll(opts.Path)
	log.Warn("go-git 克隆失败，尝试 git clone", zap.Error(err))
	return cloneViaGitCmd(opts)
}

func cloneViaGitCmd(opts CloneOptions) error {
	args := []string{"clone"}
	if opts.Branch != "" {
		args = append(args, "--single-branch", "--branch", opts.Branch)
	}
	args = append(args, opts.URL, opts.Path)

	cmd := exec.Command("git", args...)
	// Apply the configured proxy to the git subprocess too — SetupGitProxy only
	// covers go-git's transport; when we fall back to system git it would
	// otherwise connect directly and hang on restricted networks.
	applyProxyToCmd(cmd, opts.Proxy)
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(opts.Path)
		return fmt.Errorf("克隆失败: %s\n%s", err, string(out))
	}
	log.Info("git clone 完成", zap.String("path", opts.Path))
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
		RemoteName: "origin",
	}

	log.Info("正在更新仓库", zap.String("path", opts.Path))
	err = wt.Pull(pullOpts)
	if err == git.NoErrAlreadyUpToDate {
		log.Info("仓库已是最新", zap.String("path", opts.Path))
		return nil
	}
	if err == nil {
		log.Info("更新完成", zap.String("path", opts.Path))
		return nil
	}

	log.Warn("go-git 更新失败，尝试 git pull", zap.Error(err))
	return pullViaGitCmd(opts)
}

func pullViaGitCmd(opts CloneOptions) error {
	branch := "HEAD"
	if opts.Branch != "" {
		branch = opts.Branch
	}

	if _, err := os.Stat(filepath.Join(opts.Path, ".git", "shallow")); err == nil {
		log.Info("检测到浅克隆，执行 unshallow", zap.String("path", opts.Path))
		unshallow := exec.Command("git", "-C", opts.Path, "fetch", "--unshallow", "origin")
		applyProxyToCmd(unshallow, opts.Proxy)
		if out, err := unshallow.CombinedOutput(); err != nil {
			return fmt.Errorf("unshallow 失败: %s\n%s", err, string(out))
		}
	}

	fetch := exec.Command("git", "-C", opts.Path, "fetch", "origin")
	applyProxyToCmd(fetch, opts.Proxy)
	if out, err := fetch.CombinedOutput(); err != nil {
		return fmt.Errorf("fetch 失败: %s\n%s", err, string(out))
	}

	reset := exec.Command("git", "-C", opts.Path, "reset", "--hard", "origin/"+branch)
	if out, err := reset.CombinedOutput(); err != nil {
		return fmt.Errorf("reset 失败: %s\n%s", err, string(out))
	}

	log.Info("git fetch + reset 完成", zap.String("path", opts.Path))
	return nil
}

// applyProxyToCmd injects the configured proxy URL into a git subprocess via
// environment variables (works for both http/https and git+ssh-with-proxy-helper
// setups). No-op when proxyURL is empty.
func applyProxyToCmd(cmd *exec.Cmd, proxyURL string) {
	if proxyURL == "" {
		return
	}
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env,
		"HTTP_PROXY="+proxyURL,
		"HTTPS_PROXY="+proxyURL,
		"http_proxy="+proxyURL,
		"https_proxy="+proxyURL,
	)
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
