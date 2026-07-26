package main

// 壁纸后端（对齐 byte-stash commands/wallpaper.rs，Tauri→Wails/Go 改写）
//
// 壁纸【文件】由后端管（~/.cicbyte/reference/wallpapers/）；壁纸【设置】
// （开关/选中/模糊/遮罩/面板透明度/毛玻璃）纯前端 localStorage 持久化。
// 前端用 WallpaperDataURL 拿 base64 data URL 渲染（Wails 无 Tauri asset:
// 对等协议，决策 1：base64 data URL）。
//
// 安全：扩展名白名单 + 路径穿越防护（拒绝分隔符/..，二次 IsPathWithin 校验）
// + 30MB 上限（base64 后约 40MB data URL，IPC 与内存的硬上限）。

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cicbyte/reference/internal/utils"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// 允许的壁纸扩展名 → MIME（小写，含点）。
var wallpaperExts = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".avif": "image/avif",
}

// wallpaperMaxSize 单张壁纸大小硬上限（30MB）。
const wallpaperMaxSize = 30 * 1024 * 1024

// wallpapersDir 返回壁纸存储目录：~/.cicbyte/reference/wallpapers/
func wallpapersDir() string {
	return filepath.Join(utils.ConfigInstance.GetAppDir(), "wallpapers")
}

// wallpaperMIME 按扩展名查 MIME；不在白名单返回 ok=false。
func wallpaperMIME(name string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(name))
	mime, ok := wallpaperExts[ext]
	return mime, ok
}

// isSafeWallpaperName 拒绝路径穿越（含分隔符/..）并校验扩展名白名单。
func isSafeWallpaperName(name string) bool {
	if name == "" || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return false
	}
	if _, ok := wallpaperMIME(name); !ok {
		return false
	}
	return true
}

// WallpaperUpload 把选中图片复制到 wallpapers/，返回文件名。
// 扩展名白名单校验 + 30MB 上限 + 时间戳前缀防同名覆盖。
func (a *ReferenceApp) WallpaperUpload(src string) (string, error) {
	if src == "" {
		return "", fmt.Errorf("源文件路径不能为空")
	}
	if _, ok := wallpaperMIME(src); !ok {
		return "", fmt.Errorf("不支持的图片格式")
	}
	info, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("无法读取源文件: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("不能上传目录")
	}
	if info.Size() > wallpaperMaxSize {
		return "", fmt.Errorf("图片过大（超过 30MB），请压缩后重试")
	}

	dir := wallpapersDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建壁纸目录失败: %w", err)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("读取源文件失败: %w", err)
	}
	stem := filepath.Base(src)
	ts := time.Now().Unix()
	filename := fmt.Sprintf("%d-%s", ts, stem)
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
		return "", fmt.Errorf("写入壁纸失败: %w", err)
	}
	return filename, nil
}

// WallpaperList 列出全部壁纸文件名（仅图片扩展名，按修改时间倒序）。
func (a *ReferenceApp) WallpaperList() ([]string, error) {
	dir := wallpapersDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("读取壁纸目录失败: %w", err)
	}
	type wpEntry struct {
		name string
		mod  time.Time
	}
	items := make([]wpEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if _, ok := wallpaperMIME(name); !ok {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, wpEntry{name: name, mod: info.ModTime()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].mod.After(items[j].mod) })
	result := make([]string, 0, len(items))
	for _, it := range items {
		result = append(result, it.name)
	}
	return result, nil
}

// WallpaperDelete 删除指定壁纸（防路径穿越 + 白名单 + IsPathWithin 二次校验）。
func (a *ReferenceApp) WallpaperDelete(filename string) error {
	if !isSafeWallpaperName(filename) {
		return fmt.Errorf("非法文件名")
	}
	dir := wallpapersDir()
	path := filepath.Join(dir, filename)
	if !utils.IsPathWithin(path, dir) {
		return fmt.Errorf("非法文件名")
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除失败: %w", err)
	}
	return nil
}

// WallpaperDataURL 读取壁纸并返回 base64 data URL（前端 <img>/background-image 直接用）。
// 30MB 上限校验，超限报错引导压缩。
func (a *ReferenceApp) WallpaperDataURL(filename string) (string, error) {
	if !isSafeWallpaperName(filename) {
		return "", fmt.Errorf("非法文件名")
	}
	mime, _ := wallpaperMIME(filename)
	dir := wallpapersDir()
	path := filepath.Join(dir, filename)
	if !utils.IsPathWithin(path, dir) {
		return "", fmt.Errorf("非法文件名")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}
	if info.Size() > wallpaperMaxSize {
		return "", fmt.Errorf("图片过大（超过 30MB）")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取壁纸失败: %w", err)
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

// WallpaperDataDir 返回壁纸存储目录（设置页展示用）。
func (a *ReferenceApp) WallpaperDataDir() (string, error) {
	return wallpapersDir(), nil
}

// PickImageFile 打开原生图片文件选择器，返回选中路径（取消返回空串）。
func (a *ReferenceApp) PickImageFile() (string, error) {
	a.appMu.RLock()
	ctx := a.ctx
	a.appMu.RUnlock()
	if ctx == nil {
		return "", nil
	}
	selected, err := wailsruntime.OpenFileDialog(ctx, wailsruntime.OpenDialogOptions{
		Title: "选择壁纸图片",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "图片 (*.jpg;*.jpeg;*.png;*.webp;*.gif;*.bmp;*.avif)",
				Pattern:     "*.jpg;*.jpeg;*.png;*.webp;*.gif;*.bmp;*.avif",
			},
		},
	})
	if err != nil {
		// Windows 取消对话框报 "shellItem is nil"，视为正常空选择。
		if strings.Contains(err.Error(), "shellItem") {
			return "", nil
		}
		return "", err
	}
	return selected, nil
}
