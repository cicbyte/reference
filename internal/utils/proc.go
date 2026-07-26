package utils

import (
	"os/exec"
	"runtime"
	"syscall"
)

// HideWindow 让子进程不弹出控制台窗口。
//
// Windows GUI 子系统（如打包后的 Wails 应用）调用 cmd.exe / git.exe 等控制台
// 程序时，系统会为它们分配一个新控制台，表现为闪现黑色控制台窗口。设置
// SysProcAttr.HideWindow 后，子进程改为继承父进程的控制台（CLI 场景）或完全
// 不显示窗口（GUI 场景）。非 Windows 平台为 no-op。
//
// 用法：utils.HideWindow(cmd) 后再 cmd.Run()/Start()/CombinedOutput()。
// 返回 cmd 本身，便于链式：utils.HideWindow(exec.Command(...)).CombinedOutput()。
func HideWindow(cmd *exec.Cmd) *exec.Cmd {
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	return cmd
}
