package wiki

import (
	"fmt"

	"github.com/cicbyte/reference/internal/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"go.uber.org/zap"
)

func GetRemoteURL(wikiDir string) (string, error) {
	repo, err := git.PlainOpen(wikiDir)
	if err != nil {
		return "", fmt.Errorf("wiki 仓库无效: %w", err)
	}
	cfg, err := repo.Config()
	if err != nil {
		return "", err
	}
	if r, ok := cfg.Remotes["origin"]; ok && len(r.URLs) > 0 {
		return r.URLs[0], nil
	}
	return "", nil
}

func SetRemote(wikiDir, remoteURL string) error {
	repo, err := git.PlainOpen(wikiDir)
	if err != nil {
		return fmt.Errorf("wiki 仓库无效: %w", err)
	}
	_ = repo.DeleteRemote("origin")
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	})
	if err != nil {
		return fmt.Errorf("设置远程仓库失败: %w", err)
	}
	log.Info("wiki 远程仓库已设置", zap.String("url", remoteURL))
	return nil
}
