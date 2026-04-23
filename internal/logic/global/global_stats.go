package global

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"gorm.io/gorm"
)

type GlobalStatsConfig struct{}

type GlobalStatsProcessor struct {
	appConfig *models.AppConfig
	db        *gorm.DB
}

type GlobalStatsResult struct {
	Projects struct {
		Total    int `json:"total"`
		Existing int `json:"existing"`
		Deleted  int `json:"deleted"`
	} `json:"projects"`
	Repos struct {
		TotalCached int `json:"total_cached"`
	} `json:"repos"`
	CacheSize int64 `json:"cache_size_bytes"`
	WikiSize  int64 `json:"wiki_size_bytes"`
	DBSize    int64 `json:"db_size_bytes"`
}

func NewGlobalStatsProcessor(appConfig *models.AppConfig, db *gorm.DB) *GlobalStatsProcessor {
	return &GlobalStatsProcessor{appConfig: appConfig, db: db}
}

func (p *GlobalStatsProcessor) Execute(ctx context.Context) (*GlobalStatsResult, error) {
	result := &GlobalStatsResult{}
	indexer := repo.NewRepoIndexer(p.db)

	projectDirs, err := indexer.ListAllProjectDirs()
	if err != nil {
		return nil, err
	}
	result.Projects.Total = len(projectDirs)
	for _, dir := range projectDirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			result.Projects.Existing++
		}
	}
	result.Projects.Deleted = result.Projects.Total - result.Projects.Existing

	cachePaths, err := indexer.ListAllCachePaths()
	if err != nil {
		return nil, err
	}
	result.Repos.TotalCached = len(cachePaths)

	reposDir := utils.ConfigInstance.GetReposDirFromConfig(p.appConfig)
	if size, err := dirSize(reposDir); err == nil {
		result.CacheSize = size
	}

	wikiDir := utils.ConfigInstance.GetWikiDir()
	if size, err := dirSize(wikiDir); err == nil {
		result.WikiSize = size
	}

	dbPath := utils.ConfigInstance.GetDbPath()
	if info, err := os.Stat(dbPath); err == nil {
		result.DBSize = info.Size()
	}

	return result, nil
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return nil
	})
	return size, err
}
