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

// resolveRepoPath finds the on-disk path for a referenced repo by refName,
// scoped to projectDir. Remote repos store their path in CachePath; local
// repos store it in LocalPath (CachePath is empty for them), so we fall back.
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
			if r.CachePath != "" {
				return r.CachePath, nil
			}
			if r.LocalPath != "" {
				return r.LocalPath, nil
			}
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

// =====================================================================
// Shared path-based browsing helpers (used by both refName and cachePath
// variants — the actual file operations are identical once you have a root).
// =====================================================================

func browsePathList(rootPath, subPath string) ([]BrowserFileNode, error) {
	target := rootPath
	if subPath != "" {
		target = filepath.Join(rootPath, subPath)
	}
	if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(rootPath)) {
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
			Name: nm, Path: rel, IsDir: e.IsDir(), Size: size,
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

func browsePathRead(rootPath, relPath string) (BrowserFileResult, error) {
	full := filepath.Join(rootPath, relPath)
	if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(rootPath)) {
		return BrowserFileResult{}, fmt.Errorf("路径越界")
	}
	data, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return BrowserFileResult{NotFound: true}, nil
		}
		return BrowserFileResult{}, fmt.Errorf("读取文件失败: %w", err)
	}
	if info, serr := os.Stat(full); serr == nil && info.IsDir() {
		return BrowserFileResult{NotFound: true}, nil
	}
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

func browsePathSearch(rootPath, query string) ([]BrowserFileNode, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return []BrowserFileNode{}, nil
	}
	const maxResults = 500
	matches := make([]BrowserFileNode, 0, 64)
	walkErr := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
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
		rel, rerr := filepath.Rel(rootPath, path)
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

// =====================================================================
// Path-based browsing (no project context — for the global cache view)
//
// Same semantics as BrowseRepoList/Read/Search but keyed by an absolute
// cachePath instead of refName+project. Lets the cache-repos page browse
// a cached repo without switching projects.
// =====================================================================

func (a *ReferenceApp) BrowseCacheByPathList(cachePath string, subPath string) ([]BrowserFileNode, error) {
	return browsePathList(cachePath, subPath)
}

func (a *ReferenceApp) BrowseCacheByPathRead(cachePath string, relPath string) (BrowserFileResult, error) {
	return browsePathRead(cachePath, relPath)
}

func (a *ReferenceApp) BrowseCacheByPathSearch(cachePath string, query string) ([]BrowserFileNode, error) {
	return browsePathSearch(cachePath, query)
}

// errWalkDone is a sentinel used by BrowseRepoSearch to stop walking once the
// result cap is reached.
var errWalkDone = fmt.Errorf("walk done")

// =====================================================================
// Global cache management (deduplicated view of all cached remote repos)
// =====================================================================

// CachedRepoItem describes one unique cached repo (deduplicated by path).
type CachedRepoItem struct {
	Name      string   `json:"name"`      // filepath.Base(path)
	CachePath string   `json:"cachePath"` // on-disk path (CachePath or LocalPath)
	Type      string   `json:"type"`      // "remote" | "local"
	Size      int64    `json:"size"`      // bytes (fetched async)
	RefCount  int      `json:"refCount"`  // how many projects reference it
	Projects  []string `json:"projects"`  // project dirs referencing it
	Branch    string   `json:"branch"`
	Commit    string   `json:"commit"`
}

// ListCachedRepos returns all unique repos (remote caches + local paths),
// deduplicated by on-disk path, enriched with reference count and git metadata.
// Size is loaded async by GetCacheSize to keep this call instant.
func (a *ReferenceApp) ListCachedRepos() ([]CachedRepoItem, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}
	indexer := repo.NewRepoIndexer(db)
	allRepos, err := indexer.ListAll()
	if err != nil {
		return nil, err
	}

	// group by on-disk path (remote: CachePath, local: LocalPath)
	type cacheInfo struct {
		projects map[string]bool
		rtype    string
	}
	groups := map[string]*cacheInfo{}
	order := []string{}
	for projectDir, repos := range allRepos {
		for _, r := range repos {
			path := ""
			if r.RefType == "remote" && r.CachePath != "" {
				path = r.CachePath
			} else if r.RefType == "local" && r.LocalPath != "" {
				path = r.LocalPath
			}
			if path == "" {
				continue
			}
			if _, ok := groups[path]; !ok {
				groups[path] = &cacheInfo{projects: map[string]bool{}, rtype: string(r.RefType)}
				order = append(order, path)
			}
			groups[path].projects[projectDir] = true
		}
	}

	items := make([]CachedRepoItem, 0, len(order))
	for _, path := range order {
		info := groups[path]
		projects := make([]string, 0, len(info.projects))
		for dir := range info.projects {
			projects = append(projects, dir)
		}
		sort.Strings(projects)

		item := CachedRepoItem{
			Name:      filepath.Base(path),
			CachePath: path,
			Type:      info.rtype,
			Size:      0, // fetched async by GetCacheSize
			RefCount:  len(projects),
			Projects:  projects,
		}
		if branch, commit, _, err := repo.GetRepoMeta(path); err == nil {
			item.Branch = branch
			item.Commit = commit
		}
		items = append(items, item)
	}

	// sort by name (size is loaded async — frontend can re-sort if desired)
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items, nil
}

// GetCacheSize returns the total disk size (bytes) of a cache directory.
// Called async by the frontend so the initial list loads instantly.
func (a *ReferenceApp) GetCacheSize(cachePath string) (int64, error) {
	if cachePath == "" {
		return 0, nil
	}
	return walkDirSize(cachePath), nil
}

// PurgeCachedRepo deletes a cache directory from disk. The path must live
// inside the repos cache dir (safety guard against arbitrary deletion).
// DB records are NOT touched — the repo can be re-cloned on next `add`.
func (a *ReferenceApp) PurgeCachedRepo(cachePath string) error {
	if cachePath == "" {
		return fmt.Errorf("路径不能为空")
	}
	reposDir := utils.ConfigInstance.GetReposDir()
	cleanPath := filepath.Clean(cachePath)
	cleanBase := filepath.Clean(reposDir)
	if !strings.HasPrefix(cleanPath, cleanBase) {
		return fmt.Errorf("路径不在缓存目录内，拒绝删除")
	}
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return fmt.Errorf("缓存目录不存在")
	}
	return repo.PurgeCache(cleanPath)
}

// walkDirSize returns the total byte size of all files under path (recursive).
func walkDirSize(path string) int64 {
	var size int64
	_ = filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if info, ierr := d.Info(); ierr == nil {
				size += info.Size()
			}
		}
		return nil
	})
	return size
}
