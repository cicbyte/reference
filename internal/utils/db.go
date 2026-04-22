package utils

import (
	"sync"

	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/models"
	"github.com/glebarez/sqlite"
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
		err = gormDB.AutoMigrate(&models.Repo{})
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
