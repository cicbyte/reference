//go:build !windows

package utils

import "os/exec"

// HideWindow 在非 Windows 平台为 no-op——控制台窗口闪烁是 Windows GUI 子系统
// 特有的问题，Unix 平台无此概念。保留同名函数使调用方无需按平台分支。
func HideWindow(cmd *exec.Cmd) *exec.Cmd {
	return cmd
}
