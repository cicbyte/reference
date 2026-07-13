package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/cicbyte/reference/internal/common"
	"github.com/cicbyte/reference/internal/log"
	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

// ReferenceApp is the main application struct bound to Wails frontend.
type ReferenceApp struct {
	ctx           context.Context
	appMu         sync.RWMutex
	currentProject string // explicitly selected project dir; empty → fall back to GetGitRoot

	// ready is closed once the startup init chain (dirs / config / db) has run,
	// so binding methods can block until the backend is actually usable instead
	// of racing the startup goroutine.
	ready     chan struct{}
	readyOnce sync.Once
}

func NewReferenceApp() *ReferenceApp {
	return &ReferenceApp{ready: make(chan struct{})}
}

// waitReady blocks until the startup init chain completes. Binding methods that
// touch the DB / config call this first to avoid racing the async startup.
func (a *ReferenceApp) waitReady() {
	select {
	case <-a.ready:
	case <-a.ctx.Done():
	}
}

// getCurrentProject resolves the active project directory.
// Priority: explicitly switched project > process CWD's git root.
// Returns an error if neither is available (no project selected and not in a
// git repo). Callers should surface this to the user as "请先选择一个项目".
func (a *ReferenceApp) getCurrentProject() (string, error) {
	a.appMu.RLock()
	cur := a.currentProject
	a.appMu.RUnlock()
	if cur != "" {
		return cur, nil
	}
	// fall back to the CLI-style discovery (process CWD → git root)
	if dir, err := utils.GetGitRoot(); err == nil {
		return dir, nil
	}
	return "", fmt.Errorf("请先选择一个项目")
}

// setCurrentProject sets the active project dir (thread-safe).
func (a *ReferenceApp) setCurrentProject(dir string) {
	a.appMu.Lock()
	a.currentProject = dir
	a.appMu.Unlock()
}

// dirExists reports whether the given directory still exists on disk.
func dirExists(dir string) bool {
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// baseName returns the last path component of dir (for display), or dir itself
// if it is empty/root.
func baseName(dir string) string {
	if dir == "" {
		return ""
	}
	return filepath.Base(dir)
}

func (a *ReferenceApp) startup(ctx context.Context) {
	a.appMu.Lock()
	a.ctx = ctx
	a.appMu.Unlock()

	// App init — mirrors cmd/root.go init() so the GUI process loads the same
	// config (custom repos_path / wiki_path), creates dirs, opens the DB, etc.
	// Runs in a goroutine so the window appears fast; binding methods block on
	// waitReady() until this completes. Errors are logged, not silently dropped.
	go func() {
		defer a.readyOnce.Do(func() { close(a.ready) })

		if err := utils.InitAppDirs(); err != nil {
			log.Error("初始化目录失败", zap.Error(err))
			return
		}
		common.SetAppConfig(utils.ConfigInstance.LoadConfig())
		utils.ConfigInstance.ApplyConfig(common.AppConfigModel)
		if err := utils.InitDataDirs(); err != nil {
			log.Error("初始化数据目录失败", zap.Error(err))
			return
		}
		if err := log.Init(utils.ConfigInstance.GetLogPath()); err != nil {
			fmt.Fprintf(os.Stderr, "日志初始化失败: %v\n", err)
		}
		if _, err := utils.GetGormDB(); err != nil {
			log.Error("数据库连接失败", zap.String("operation", "db init"), zap.Error(err))
			return
		}
		log.Info("数据库连接成功")
		utils.MigratePathsIfNeeded()

		wikiDir := utils.ConfigInstance.GetWikiDir()
		if err := logicwiki.EnsureGitInit(wikiDir); err != nil {
			log.Warn("wiki git 初始化失败", zap.Error(err))
		}
		localWikiDir := utils.ConfigInstance.GetLocalWikiDir()
		if err := logicwiki.EnsureGitInit(localWikiDir); err != nil {
			log.Warn("localwiki git 初始化失败", zap.Error(err))
		}
	}()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		runtime.Quit(ctx)
	}()
}

func (a *ReferenceApp) shutdown(ctx context.Context) {
	// cleanup if needed
}
