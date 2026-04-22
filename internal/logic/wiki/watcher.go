package wiki

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/cicbyte/reference/internal/log"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type Watcher struct {
	wikiDir    string
	fsWatcher  *fsnotify.Watcher
	debounce   time.Duration
	autoPush   bool
	done       chan struct{}
}

func NewWatcher(wikiDir string, autoPush bool) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		wikiDir:   wikiDir,
		fsWatcher: fsWatcher,
		debounce:  3 * time.Second,
		autoPush:  autoPush,
		done:      make(chan struct{}),
	}

	if err := w.addWatch(wikiDir); err != nil {
		fsWatcher.Close()
		return nil, err
	}

	return w, nil
}

func (w *Watcher) addWatch(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" {
				return filepath.SkipDir
			}
			if err := w.fsWatcher.Add(path); err != nil {
				log.Debug("watcher 添加目录失败", zap.String("path", path), zap.Error(err))
			}
		}
		return nil
	})
}

func (w *Watcher) Run() {
	log.Info("wiki watcher 已启动",
		zap.String("dir", w.wikiDir),
		zap.Bool("autoPush", w.autoPush))

	var timer *time.Timer
	var pending atomic.Bool

	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			name := filepath.Base(event.Name)
			if name == ".git" {
				continue
			}

			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						if filepath.Base(event.Name) != ".git" {
							w.fsWatcher.Add(event.Name)
						}
					}
				}

				if pending.CompareAndSwap(false, true) {
					timer = time.AfterFunc(w.debounce, func() {
						w.commit()
						pending.Store(false)
					})
				} else {
					timer.Reset(w.debounce)
				}
			}

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			log.Warn("watcher 错误", zap.Error(err))

		case <-w.done:
			if timer != nil {
				timer.Stop()
				if pending.Swap(false) {
					w.commit()
				}
			}
			return
		}
	}
}

func (w *Watcher) commit() {
	result, err := StageAndCommit(w.wikiDir, "wiki: auto-commit from watcher")
	if err != nil {
		log.Debug("watcher 自动提交跳过", zap.Error(err))
		return
	}
	if result != nil && result.HasChanges {
		log.Info("watcher 自动提交", zap.String("hash", result.CommitHash))
		if w.autoPush {
			if pushErr := Push(w.wikiDir); pushErr != nil {
				log.Warn("watcher 自动推送失败", zap.Error(pushErr))
			}
		}
	}
}

func (w *Watcher) Stop() {
	close(w.done)
	w.fsWatcher.Close()
	log.Info("wiki watcher 已停止")
}
