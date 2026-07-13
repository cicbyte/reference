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
		// WAL mode + busy_timeout lets concurrent readers coexist with a writer
		// (the GUI fires many IPC calls + the diagnostic goroutines), avoiding
		// "database is locked" under load. foreign_keys keeps referential integrity.
		dsn := "file:" + dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)"
		gormDB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: log.GetGormLogger(),
		})
		if err != nil {
			return
		}
		// Serialize writes through a single connection; WAL still allows
		// concurrent reads via the pool.
		if sqlDB, dbErr := gormDB.DB(); dbErr == nil {
			sqlDB.SetMaxOpenConns(1)
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

// escapeLike escapes SQL LIKE wildcards in a literal string so it matches
// verbatim (used for path-prefix queries where the path may contain _ or %).
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
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

	// Wrap in a transaction so a mid-migration failure leaves the DB consistent
	// with the "repos_path" config state — either all records move + marker
	// updates, or nothing does.
	err := gormDB.Transaction(func(tx *gorm.DB) error {
		var repos []models.Repo
		// Escape LIKE wildcards in the literal path; the IsPathWithin check
		// below is the authoritative filter (defends against prefix-lookalikes).
		if err := tx.Where("cache_path LIKE ? ESCAPE '\\'", escapeLike(oldDir)+"%").Find(&repos).Error; err != nil {
			return err
		}
		for _, r := range repos {
			if !IsPathWithin(r.CachePath, oldDir) {
				continue
			}
			r.CachePath = newDir + r.CachePath[len(oldDir):]
			if err := tx.Save(&r).Error; err != nil {
				return err
			}
		}
		return tx.Save(&models.ConfigState{Key: "repos_path", Value: newReposDir}).Error
	})
	if err != nil {
		log.Error("缓存路径迁移失败", zap.String("old", oldDir), zap.String("new", newDir), zap.Error(err))
		return
	}

	log.Info("检测到缓存路径变更，已迁移数据库记录",
		zap.String("old", oldDir),
		zap.String("new", newDir))
}
