package reference

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/cicbyte/reference/internal/models"
	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Engine reference 库的核心引擎
type Engine struct {
	opts       Options
	appConfig  *models.AppConfig
	db         *gorm.DB
	logger     *zap.Logger
	promptsFS  interface{ ReadFile(string) ([]byte, error) }
	mu         sync.RWMutex
	closed     bool
}

// New 创建新的 Engine 实例
func New(opts Options) (*Engine, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	e := &Engine{
		opts: opts,
	}

	// 初始化应用目录
	if err := e.initAppDirs(); err != nil {
		return nil, fmt.Errorf("init app dirs: %w", err)
	}

	// 初始化日志
	if err := e.initLogger(); err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	// 初始化配置
	e.initConfig()

	// 初始化数据库
	if err := e.initDB(); err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	// 初始化数据目录
	if err := e.initDataDirs(); err != nil {
		return nil, fmt.Errorf("init data dirs: %w", err)
	}

	return e, nil
}

// initAppDirs 初始化应用目录
func (e *Engine) initAppDirs() error {
	if e.opts.AppDir == "" {
		homeDir, err := getHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}
		e.opts.AppDir = filepath.Join(homeDir, ".cicbyte", "reference")
	}

	dirs := []string{
		e.opts.AppDir,
		filepath.Join(e.opts.AppDir, "config"),
		filepath.Join(e.opts.AppDir, "db"),
		filepath.Join(e.opts.AppDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	return nil
}

// initLogger 初始化日志
func (e *Engine) initLogger() error {
	if e.opts.Logger != nil {
		e.logger = e.opts.Logger
		return nil
	}

	// 创建 nop logger
	e.logger = zap.NewNop()
	return nil
}

// initConfig 初始化配置
func (e *Engine) initConfig() {
	e.appConfig = &models.AppConfig{
		ReposPath: e.getReposPath(),
		WikiPath:  e.getWikiPath(),
	}

	if e.opts.Proxy.HTTP != "" {
		e.appConfig.Network.Proxy = e.opts.Proxy.HTTP
	}
	if e.opts.Proxy.Git != "" {
		e.appConfig.Network.GitProxy = e.opts.Proxy.Git
	}
}

// initDB 初始化数据库
func (e *Engine) initDB() error {
	if e.opts.DB != nil {
		e.db = e.opts.DB
		return nil
	}

	dbPath := filepath.Join(e.opts.AppDir, "db", "app.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newGormLogger(),
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&models.Repo{}, &models.ConfigState{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	e.db = db
	return nil
}

// initDataDirs 初始化数据目录
func (e *Engine) initDataDirs() error {
	dirs := []string{
		e.getReposPath(),
		e.getWikiPath(),
		e.getLocalWikiPath(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create data dir %s: %w", dir, err)
		}
	}

	return nil
}

// getReposPath 获取仓库缓存路径
func (e *Engine) getReposPath() string {
	if e.opts.ReposPath != "" {
		return e.opts.ReposPath
	}
	return filepath.Join(e.opts.AppDir, "repos")
}

// getWikiPath 获取知识库路径
func (e *Engine) getWikiPath() string {
	if e.opts.WikiPath != "" {
		return e.opts.WikiPath
	}
	return filepath.Join(e.opts.AppDir, "wiki")
}

// getLocalWikiPath 获取本地知识库路径
func (e *Engine) getLocalWikiPath() string {
	if e.opts.LocalWikiPath != "" {
		return e.opts.LocalWikiPath
	}
	return filepath.Join(e.opts.AppDir, "localwiki")
}

// GetDB 获取数据库连接
func (e *Engine) GetDB() *gorm.DB {
	return e.db
}

// GetAppConfig 获取应用配置
func (e *Engine) GetAppConfig() *models.AppConfig {
	return e.appConfig
}

// GetLogger 获取日志记录器
func (e *Engine) GetLogger() *zap.Logger {
	return e.logger
}

// SetPromptsFS 设置嵌入的提示词文件系统
func (e *Engine) SetPromptsFS(fs interface{ ReadFile(string) ([]byte, error) }) {
	e.promptsFS = fs
}

// Close 关闭 Engine，释放资源
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil
	}

	e.closed = true

	// 关闭数据库连接
	if e.db != nil {
		sqlDB, err := e.db.DB()
		if err == nil {
			return sqlDB.Close()
		}
	}

	return nil
}

// getHomeDir 获取用户主目录
func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		if home == "" {
			return "", fmt.Errorf("cannot determine home directory")
		}
		return home, nil
	}
	return usr.HomeDir, nil
}

// newGormLogger 创建 GORM 日志适配器
func newGormLogger() logger.Interface {
	return logger.Default.LogMode(logger.Warn)
}
