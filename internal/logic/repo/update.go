package repo

import (
	"context"
	"fmt"

	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UpdateConfig struct {
	Identifier string
	ProjectDir string
}

type UpdateProcessor struct {
	config    *UpdateConfig
	appConfig *models.AppConfig
	db        *gorm.DB
}

func NewUpdateProcessor(config *UpdateConfig, appConfig *models.AppConfig, db *gorm.DB) *UpdateProcessor {
	return &UpdateProcessor{config: config, appConfig: appConfig, db: db}
}

func (p *UpdateProcessor) Execute(ctx context.Context) error {
	indexer := NewRepoIndexer(p.db)

	if p.config.Identifier != "" {
		repo, err := indexer.Get(p.config.ProjectDir, p.config.Identifier)
		if err != nil {
			repo, err = indexer.GetByRefName(p.config.ProjectDir, p.config.Identifier)
		}
		if err != nil {
			return fmt.Errorf("未找到引用: %s", p.config.Identifier)
		}
		return p.updateOne(repo)
	}

	repos, err := indexer.List(p.config.ProjectDir)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println("当前项目暂无引用仓库。")
		return nil
	}

	updated, skipped := 0, 0
	for _, r := range repos {
		if r.RefType == models.RefTypeLocal {
			displayName := r.RefName
			if displayName == "" {
				displayName = r.LinkName
			}
			fmt.Printf("  [跳过] %s (本地引用需手动更新)\n", displayName)
			skipped++
			continue
		}
		if err := p.updateOne(&r); err != nil {
			log.Warn("更新失败", zap.String("repo", r.LinkName), zap.Error(err))
		} else {
			updated++
		}
	}

	fmt.Printf("\n更新完成: %d 个已更新, %d 个已跳过\n", updated, skipped)
	return nil
}

func (p *UpdateProcessor) updateOne(repo *models.Repo) error {
	displayName := repo.RefName
	if displayName == "" {
		displayName = repo.LinkName
	}

	if repo.RefType == models.RefTypeLocal {
		fmt.Printf("  [跳过] %s (本地引用需手动更新)\n", displayName)
		return nil
	}

	if repo.CachePath == "" || repo.RemoteURL == "" {
		log.Warn("缺少缓存路径或远程地址", zap.String("repo", repo.LinkName))
		return nil
	}

	proxy := ResolveProxy(p.appConfig)
	err := CloneOrUpdate(CloneOptions{
		URL:      repo.RemoteURL,
		Path:     repo.CachePath,
		Branch:   repo.Branch,
		Depth:    0,
		Proxy:    proxy,
		NoUpdate: false,
	})
	if err != nil {
		return err
	}

	branch, commit, commitTime, _ := GetRepoMeta(repo.CachePath)
	repo.Branch = branch
	repo.Commit = commit
	repo.CommitAt = commitTime

	indexer := NewRepoIndexer(p.db)
	if err := indexer.Add(repo); err != nil {
		log.Warn("更新数据库索引失败", zap.Error(err))
	}

	fmt.Printf("  [更新] %s -> %s\n", displayName, commit)
	return nil
}
