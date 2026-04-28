package utils

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/models"
	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	gormDB *gorm.DB
	dbOnce sync.Once
)

func GetGormDB() (*gorm.DB, error) {
	var err error
	dbOnce.Do(func() {
		dbPath := ConfigInstance.GetDbPath()
		gormDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: log.GetGormLogger(),
		})
		if err != nil {
			return
		}
		err = gormDB.AutoMigrate(&models.Repo{}, &models.ConfigState{})
	})
	if err != nil {
		return nil, err
	}
	return gormDB, nil
}

func CloseGormDB() error {
	if gormDB != nil {
		sqlDB, err := gormDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func getConfigState(key string) string {
	var state models.ConfigState
	if err := gormDB.Where("key = ?", key).First(&state).Error; err != nil {
		return ""
	}
	return state.Value
}

func setConfigState(key, value string) {
	gormDB.Save(&models.ConfigState{Key: key, Value: value})
}

func MigratePathsIfNeeded() {
	newReposDir := ConfigInstance.GetReposDir()
	oldReposDir := getConfigState("repos_path")

	if oldReposDir == "" {
		setConfigState("repos_path", newReposDir)
		return
	}

	if oldReposDir == newReposDir {
		return
	}

	oldDir := filepath.Clean(oldReposDir)
	newDir := filepath.Clean(newReposDir)

	var repos []models.Repo
	gormDB.Where("cache_path LIKE ?", oldDir+"%").Find(&repos)

	for _, r := range repos {
		if !strings.HasPrefix(r.CachePath, oldDir) {
			continue
		}
		r.CachePath = newDir + r.CachePath[len(oldDir):]
		gormDB.Save(&r)
	}

	setConfigState("repos_path", newReposDir)

	log.Info("检测到缓存路径变更，已迁移数据库记录",
		zap.String("old", oldDir),
		zap.String("new", newDir))
}
