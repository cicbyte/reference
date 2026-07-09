package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
)

// =====================================================================
// Reference-repo code browser (read-only file tree + file content)
//
// Powers the /repos/browse/:name page. The repos live in the global cache
// (.reference/repos/<refName> → CachePath); we resolve refName → CachePath
// the same way RunSCC does, then walk that directory.
// =====================================================================

// BrowserFileNode is one entry in a directory listing (name + relative path).
type BrowserFileNode struct {
	Name  string `json:"name"`
	Path  string `json:"path"` // relative to repo root, slash-separated
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"` // files only
}

// BrowserFileResult is the content of a single file (with binary/guard flags).
type BrowserFileResult struct {
	Content  string `json:"content"`
	Lines    int    `json:"lines"`
	Binary   bool   `json:"binary"`
	NotFound bool   `json:"notFound"`
}

// maxBrowseFileSize caps how much file content we ship to the frontend.
const maxBrowseFileSize = 1 << 20 // 1 MB

// common noise names skipped at every level of the tree / search walk so the
// browser stays focused on the user's code.
var browserSkipNames = map[string]bool{
	".git": true, "node_modules": true, ".DS_Store": true,
}

// resolveRepoPath finds the on-disk cache path for a referenced repo by
// refName, scoped to projectDir. Returns "" if not found.
func resolveRepoPath(projectDir, refName string) (string, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return "", err
	}
	indexer := repo.NewRepoIndexer(db)
	repos, err := indexer.List(projectDir)
	if err != nil {
		return "", err
	}
	for _, r := range repos {
		if r.GetRefName() == refName {
			return r.CachePath, nil
		}
	}
	return "", fmt.Errorf("仓库 %s 未找到", refName)
}

// BrowseRepoList lists the immediate children of subPath ("") = repo root.
// Directories are listed before files; noise entries are pruned.
func (a *ReferenceApp) BrowseRepoList(refName string, subPath string) ([]BrowserFileNode, error) {
	projectDir, err := a.getCurrentProject()
	if err != nil {
		return nil, err
	}
	repoPath, err := resolveRepoPath(projectDir, refName)
	if err != nil {
		return nil, err
	}
	target := repoPath
	if subPath != "" {
		target = filepath.Join(repoPath, subPath)
	}
	// path-traversal guard: resolved target must stay under repo root
	if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(repoPath)) {
		return nil, fmt.Errorf("路径越界")
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		if os.IsNotExist(err) {
			return []BrowserFileNode{}, nil
		}
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}
	nodes := make([]BrowserFileNode, 0, len(entries))
	for _, e := range entries {
		nm := e.Name()
		if browserSkipNames[nm] {
			continue
		}
		info, ierr := e.Info()
		size := int64(0)
		if ierr == nil {
			size = info.Size()
		}
		rel := nm
		if subPath != "" {
			rel = subPath + "/" + nm
		}
		nodes = append(nodes, BrowserFileNode{
			Name:  nm,
			Path:  rel,
			IsDir: e.IsDir(),
			Size:  size,
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
	return nodes, nil
}

// BrowseRepoRead reads a file by relative path. Detects binary content,
// caps at 1 MB, and guards against path traversal.
func (a *ReferenceApp) BrowseRepoRead(refName string, relPath string) (BrowserFileResult, error) {
	projectDir, err := a.getCurrentProject()
	if err != nil {
		return BrowserFileResult{}, err
	}
	repoPath, err := resolveRepoPath(projectDir, refName)
	if err != nil {
		return BrowserFileResult{}, err
	}
	full := filepath.Join(repoPath, relPath)
	if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(repoPath)) {
		return BrowserFileResult{}, fmt.Errorf("路径越界")
	}
	data, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return BrowserFileResult{NotFound: true}, nil
		}
		return BrowserFileResult{}, fmt.Errorf("读取文件失败: %w", err)
	}
	// reject directories uniformly (ReadFile on a dir may behave oddly cross-platform)
	if info, serr := os.Stat(full); serr == nil && info.IsDir() {
		return BrowserFileResult{NotFound: true}, nil
	}
	// binary check: NUL byte in the first 8 KB
	scanLen := len(data)
	if scanLen > 8192 {
		scanLen = 8192
	}
	if bytes.IndexByte(data[:scanLen], 0) >= 0 {
		return BrowserFileResult{Binary: true}, nil
	}
	content := string(data)
	if len(content) > maxBrowseFileSize {
		content = content[:maxBrowseFileSize] + "\n\n…（文件超过 1MB，已截断显示）"
	}
	return BrowserFileResult{
		Content: content,
		Lines:   strings.Count(content, "\n") + 1,
	}, nil
}

// BrowseRepoSearch recursively walks the repo and returns file paths whose
// name (case-insensitive) contains the query. Caps at 500 results.
func (a *ReferenceApp) BrowseRepoSearch(refName string, query string) ([]BrowserFileNode, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return []BrowserFileNode{}, nil
	}
	projectDir, err := a.getCurrentProject()
	if err != nil {
		return nil, err
	}
	repoPath, err := resolveRepoPath(projectDir, refName)
	if err != nil {
		return nil, err
	}
	const maxResults = 500
	matches := make([]BrowserFileNode, 0, 64)
	walkErr := filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		base := d.Name()
		if d.IsDir() {
			if browserSkipNames[base] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.Contains(strings.ToLower(base), q) {
			return nil
		}
		rel, rerr := filepath.Rel(repoPath, path)
		if rerr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		info, _ := d.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		matches = append(matches, BrowserFileNode{
			Name: base, Path: rel, IsDir: false, Size: size,
		})
		if len(matches) >= maxResults {
			return errWalkDone
		}
		return nil
	})
	if walkErr != nil && walkErr != errWalkDone {
		return nil, fmt.Errorf("搜索失败: %w", walkErr)
	}
	sort.Slice(matches, func(i, j int) bool {
		return strings.ToLower(matches[i].Path) < strings.ToLower(matches[j].Path)
	})
	return matches, nil
}

// errWalkDone is a sentinel used by BrowseRepoSearch to stop walking once the
// result cap is reached.
var errWalkDone = fmt.Errorf("walk done")
