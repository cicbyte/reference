package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ReferenceApp is the main application struct bound to Wails frontend.
type ReferenceApp struct {
	ctx    context.Context
	appMu  sync.RWMutex
}

func NewReferenceApp() *ReferenceApp {
	return &ReferenceApp{}
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
