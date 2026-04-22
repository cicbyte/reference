//go:build !windows

package wiki

import (
	"fmt"
	"os"
	"syscall"
)

func Daemonize(wikiDir string) error {
	if IsDaemonRunning(wikiDir) {
		return fmt.Errorf("watcher 已在运行中")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行路径失败: %w", err)
	}

	cmd := os.Args[1:]
	var filtered []string
	for _, a := range cmd {
		if a != "--daemon" && a != "-d" {
			filtered = append(filtered, a)
		}
	}
	filtered = append(filtered, "--foreground")

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("打开 DevNull 失败: %w", err)
	}

	attr := &os.ProcAttr{
		Dir: wikiDir,
		Env: os.Environ(),
		Sys: &syscall.SysProcAttr{Setsid: true},
		Files: []*os.File{devNull, devNull, devNull},
	}

	process, err := os.StartProcess(execPath, append([]string{execPath}, filtered...), attr)
	if err != nil {
		devNull.Close()
		return fmt.Errorf("启动守护进程失败: %w", err)
	}

	devNull.Close()
	process.Release()
	return nil
}
