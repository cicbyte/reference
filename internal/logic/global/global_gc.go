package global

import (
	"context"
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

type GlobalGCConfig struct {
	DryRun bool
	Cache  bool
}

type GlobalGCProcessor struct {
	config    *GlobalGCConfig
	appConfig *models.AppConfig
	db        *gorm.DB
}

type GlobalGCResult struct {
	StaleDBRecords   []StaleDBRecord `json:"stale_db_records"`
	OrphanedCaches   []OrphanedCache `json:"orphaned_caches,omitempty"`
	DBRecordsRemoved int             `json:"db_records_removed"`
	CacheDirsRemoved int             `json:"cache_dirs_removed,omitempty"`
	DryRun           bool            `json:"dry_run"`
}

type StaleDBRecord struct {
	ProjectDir string `json:"project_dir"`
	RepoCount  int    `json:"repo_count"`
}

type OrphanedCache struct {
	Path string `json:"path"`
}

func NewGlobalGCProcessor(config *GlobalGCConfig, appConfig *models.AppConfig, db *gorm.DB) *GlobalGCProcessor {
	return &GlobalGCProcessor{config: config, appConfig: appConfig, db: db}
}

func (p *GlobalGCProcessor) Execute(ctx context.Context) (*GlobalGCResult, error) {
	result := &GlobalGCResult{DryRun: p.config.DryRun}

	if err := p.cleanStaleDBRecords(result); err != nil {
		return nil, err
	}
	if p.config.Cache {
		if err := p.cleanOrphanedCaches(result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (p *GlobalGCProcessor) cleanStaleDBRecords(result *GlobalGCResult) error {
	indexer := repo.NewRepoIndexer(p.db)
	allRepos, err := indexer.ListAll()
	if err != nil {
		return err
	}

	reposBase := utils.ConfigInstance.GetReposDirFromConfig(p.appConfig)

	for projectDir, repos := range allRepos {
		if strings.HasPrefix(projectDir, reposBase) {
			continue
		}
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			result.StaleDBRecords = append(result.StaleDBRecords, StaleDBRecord{
				ProjectDir: projectDir,
				RepoCount:  len(repos),
			})
			if !p.config.DryRun {
				removed, delErr := indexer.DeleteByProjectDir(projectDir)
				if delErr != nil {
					log.Warn("删除过期 DB 记录失败", zap.String("project", projectDir), zap.Error(delErr))
				} else {
					result.DBRecordsRemoved += int(removed)
				}
			}
		}
	}
	return nil
}

func (p *GlobalGCProcessor) cleanOrphanedCaches(result *GlobalGCResult) error {
	reposBase := utils.ConfigInstance.GetReposDirFromConfig(p.appConfig)

	indexer := repo.NewRepoIndexer(p.db)
	referencedPaths, err := indexer.ListAllCachePaths()
	if err != nil {
		return err
	}

	isReferenced := func(path string) bool {
		clean := filepath.Clean(path)
		for _, rp := range referencedPaths {
			if filepath.Clean(rp) == clean {
				return true
			}
			if strings.HasPrefix(filepath.Clean(rp), clean+string(os.PathSeparator)) {
				return true
			}
		}
		return false
	}

	entries, err := os.ReadDir(reposBase)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fullPath := filepath.Join(reposBase, entry.Name())
		if !isReferenced(fullPath) {
			result.OrphanedCaches = append(result.OrphanedCaches, OrphanedCache{Path: fullPath})
			if !p.config.DryRun {
				if err := repo.PurgeCache(fullPath); err != nil {
					log.Warn("删除孤立缓存失败", zap.String("path", fullPath), zap.Error(err))
				} else {
					result.CacheDirsRemoved++
				}
			}
		}
	}
	return nil
}
