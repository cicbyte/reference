package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/cicbyte/reference/internal/log"
	"go.uber.org/zap"
)

func pidFilePath(wikiDir string) string {
	return filepath.Join(wikiDir, ".watch.pid")
}

func WritePID(wikiDir string) error {
	pid := os.Getpid()
	return os.WriteFile(pidFilePath(wikiDir), []byte(strconv.Itoa(pid)), 0644)
}

func ReadPID(wikiDir string) (int, error) {
	data, err := os.ReadFile(pidFilePath(wikiDir))
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("无效的 PID 文件: %w", err)
	}
	return pid, nil
}

func RemovePID(wikiDir string) {
	os.Remove(pidFilePath(wikiDir))
}

func IsDaemonRunning(wikiDir string) bool {
	pid, err := ReadPID(wikiDir)
	if err != nil {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		RemovePID(wikiDir)
		return false
	}
	// os.FindProcess 在 Windows 上不会检查进程是否存在，
	// 需要通过 OpenProcess API 验证。使用 Signal(syscall.Signal(0))
	// 在 Windows 上会调用 OpenProcess，不会杀死进程。
	if err := process.Signal(syscall.Signal(0)); err != nil {
		RemovePID(wikiDir)
		return false
	}
	return true
}

func StopDaemon(wikiDir string) error {
	pid, err := ReadPID(wikiDir)
	if err != nil {
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		RemovePID(wikiDir)
		return nil
	}

	err = process.Kill()
	RemovePID(wikiDir)
	if err != nil {
		return nil
	}

	log.Info("watcher 已停止", zap.Int("pid", pid))
	return nil
}
