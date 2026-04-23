package common

import (
	"embed"
	"io/fs"

	"github.com/cicbyte/reference/internal/models"
)

var (
	AppConfigModel *models.AppConfig
	PromptsFS      embed.FS
)

func GetAssetFile(path string) ([]byte, error) {
	return PromptsFS.ReadFile(path)
}

func AssetExists(path string) bool {
	_, err := fs.Stat(PromptsFS, path)
	return err == nil
}
