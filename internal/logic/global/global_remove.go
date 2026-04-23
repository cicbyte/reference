package global

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GlobalRemoveConfig struct {
	ProjectDir string
	RepoName   string
	All        bool
	Purge      bool
	Yes        bool
}

type RemoveDetail struct {
	ProjectDir string `json:"project_dir"`
	RepoName   string `json:"repo_name"`
}

type GlobalRemoveResult struct {
	RemovedCount int            `json:"removed_count"`
	Details      []RemoveDetail `json:"details,omitempty"`
}

type GlobalRemoveProcessor struct {
	config    *GlobalRemoveConfig
	appConfig *models.AppConfig
	db        *gorm.DB
}

func NewGlobalRemoveProcessor(config *GlobalRemoveConfig, appConfig *models.AppConfig, db *gorm.DB) *GlobalRemoveProcessor {
	return &GlobalRemoveProcessor{config: config, appConfig: appConfig, db: db}
}

func (p *GlobalRemoveProcessor) Execute(ctx context.Context) (*GlobalRemoveResult, error) {
	if p.config.ProjectDir != "" {
		return p.removeByProject(ctx)
	}
	if p.config.RepoName != "" {
		return p.removeByRepoGlobal()
	}
	return nil, fmt.Errorf("请指定 --project <路径> 或 --repo <仓库名>")
}

func (p *GlobalRemoveProcessor) removeByProject(ctx context.Context) (*GlobalRemoveResult, error) {
	projectDir := filepath.Clean(p.config.ProjectDir)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("项目目录不存在: %s", projectDir)
	}

	removeCfg := &repo.RemoveConfig{
		ProjectDir: projectDir,
		Purge:      p.config.Purge,
		Yes:        p.config.Yes,
		All:        p.config.All,
		Identifier: p.config.RepoName,
	}

	processor := repo.NewRemoveProcessor(removeCfg, p.appConfig, p.db)
	if err := processor.Execute(ctx); err != nil {
		return nil, err
	}

	return &GlobalRemoveResult{RemovedCount: -1}, nil
}

func (p *GlobalRemoveProcessor) removeByRepoGlobal() (*GlobalRemoveResult, error) {
	indexer := repo.NewRepoIndexer(p.db)
	allRepos, err := indexer.ListAll()
	if err != nil {
		return nil, err
	}

	var targets []struct {
		projectDir string
		repo       models.Repo
	}
	for projectDir, repos := range allRepos {
		for _, r := range repos {
			refName := r.RefName
			if refName == "" {
				refName = r.LinkName
			}
			if refName == p.config.RepoName || r.LinkName == p.config.RepoName {
				targets = append(targets, struct {
					projectDir string
					repo       models.Repo
				}{projectDir, r})
			}
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("未找到引用: %s", p.config.RepoName)
	}

	if !p.config.Yes {
		fmt.Printf("将在 %d 个项目中移除 '%s':\n", len(targets), p.config.RepoName)
		for _, t := range targets {
			fmt.Printf("  - %s\n", t.projectDir)
		}
		fmt.Print("确认? [y/N]: ")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" {
			fmt.Println("已取消")
			return nil, nil
		}
	}

	result := &GlobalRemoveResult{}
	reposBase := utils.ConfigInstance.GetReposDirFromConfig(p.appConfig)

	for _, t := range targets {
		refName := t.repo.RefName
		if refName == "" {
			refName = t.repo.LinkName
		}

		refDir := filepath.Join(t.projectDir, ".reference")
		reposLinkDir := filepath.Join(refDir, "repos")
		wikiLinkDir := filepath.Join(refDir, "wiki")

		linkPath := filepath.Join(reposLinkDir, refName)
		repo.RemoveLink(linkPath)

		junctionPath := filepath.Join(wikiLinkDir, refName)
		if _, err := os.Lstat(junctionPath); err == nil {
			repo.RemoveLink(junctionPath)
		}

		if p.config.Purge && t.repo.RefType == models.RefTypeRemote && t.repo.CachePath != "" {
			if strings.HasPrefix(t.repo.CachePath, reposBase) {
				if err := repo.PurgeCache(t.repo.CachePath); err != nil {
					log.Warn("删除缓存失败", zap.String("repo", refName), zap.Error(err))
				}
			}
		}

		if err := indexer.Remove(t.projectDir, t.repo.LinkName); err != nil {
			log.Warn("删除数据库记录失败", zap.String("repo", refName), zap.Error(err))
		}

		if err := repo.RefreshReferenceMap(t.projectDir, refDir, indexer); err != nil {
			log.Warn("更新 reference.map.jsonl 失败", zap.String("project", t.projectDir), zap.Error(err))
		}

		result.RemovedCount++
		result.Details = append(result.Details, RemoveDetail{
			ProjectDir: t.projectDir,
			RepoName:   refName,
		})
	}

	fmt.Printf("已从 %d 个项目中移除 '%s'\n", result.RemovedCount, p.config.RepoName)
	return result, nil
}
