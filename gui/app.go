package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/cicbyte/reference/internal/utils"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ReferenceApp is the main application struct bound to Wails frontend.
type ReferenceApp struct {
	ctx           context.Context
	appMu         sync.RWMutex
	currentProject string // explicitly selected project dir; empty → fall back to GetGitRoot
}

func NewReferenceApp() *ReferenceApp {
	return &ReferenceApp{}
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
