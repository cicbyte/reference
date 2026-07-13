package common

import (
	"embed"
	"io/fs"
	"sync"

	"github.com/cicbyte/reference/internal/models"
)

var (
	// AppConfigModel is the in-memory bridge for the app config. Access it via
	// GetAppConfig / SetAppConfig to stay race-free across the startup goroutine
	// (writer) and IPC handlers (readers). Direct field access is kept for
	// legacy call sites but should migrate to the getters.
	AppConfigModel *models.AppConfig
	PromptsFS      embed.FS

	appConfigMu sync.RWMutex
)

// GetAppConfig returns a snapshot pointer to the current app config. Callers
// must not mutate the returned struct; use SetAppConfig to replace it.
func GetAppConfig() *models.AppConfig {
	appConfigMu.RLock()
	defer appConfigMu.RUnlock()
	return AppConfigModel
}

// SetAppConfig replaces the current app config atomically.
func SetAppConfig(cfg *models.AppConfig) {
	appConfigMu.Lock()
	AppConfigModel = cfg
	appConfigMu.Unlock()
}

func GetAssetFile(path string) ([]byte, error) {
	return PromptsFS.ReadFile(path)
}

func AssetExists(path string) bool {
	_, err := fs.Stat(PromptsFS, path)
	return err == nil
}
