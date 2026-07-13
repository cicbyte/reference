package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cicbyte/reference/internal/utils"
)

// =====================================================================
// Wiki browser — list and read knowledge files (reference.md + topic .md)
// from the global wiki/ and localwiki/ directories.
// =====================================================================

// WikiEntry describes one knowledge file inside wiki/ or localwiki/.
type WikiEntry struct {
	RepoName    string `json:"repoName"`    // last path segment (the repo)
	Platform    string `json:"platform"`    // github.com / gitlab.com / 本地
	Namespace   string `json:"namespace"`   // owner/org (empty for local)
	Source      string `json:"source"`      // "remote" | "local"
	RelPath     string `json:"relPath"`     // path relative to wiki root (slash-joined)
	FullPath    string `json:"fullPath"`    // absolute disk path
	FileName    string `json:"fileName"`    // reference.md / <topic>.md
	Commit      string `json:"commit"`      // from frontmatter
	Branch      string `json:"branch"`      // from frontmatter
	Description string `json:"description"` // from frontmatter
	Status      string `json:"status"`      // "ok" | "empty" | "no-fm" (fetched async)
	ExploredAt  string `json:"exploredAt"`  // from frontmatter
	ModifiedAt  string `json:"modifiedAt"`  // file mtime (RFC3339)
}

// resolveWikiRoot returns the on-disk directory for the given source.
func resolveWikiRoot(source string) (string, error) {
	switch source {
	case "remote", "":
		return utils.ConfigInstance.GetWikiDir(), nil
	case "local":
		return utils.ConfigInstance.GetLocalWikiDir(), nil
	default:
		return "", fmt.Errorf("未知 source: %s", source)
	}
}

// ListWikiEntries walks the wiki (source="remote") or localwiki (source="local")
// directory and returns every .md file with parsed frontmatter. source="all"
// merges both.
func (a *ReferenceApp) ListWikiEntries(source string) ([]WikiEntry, error) {
	sources := []string{"remote"}
	if source == "local" {
		sources = []string{"local"}
	} else if source == "all" {
		sources = []string{"remote", "local"}
	}

	entries := make([]WikiEntry, 0)
	for _, src := range sources {
		root, err := resolveWikiRoot(src)
		if err != nil {
			continue
		}
		walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			// skip .git
			if d.IsDir() {
				if d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			entry := parseWikiEntry(rel, src)
			entry.FullPath = path // absolute disk path from WalkDir
			// file mtime
			if info, ierr := d.Info(); ierr == nil {
				entry.ModifiedAt = info.ModTime().Format("2006-01-02")
			}
			entries = append(entries, entry)
			return nil
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}

	// sort: source (remote first) → platform → namespace → repoName → fileName
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Source != b.Source {
			return a.Source == "remote"
		}
		if a.Platform != b.Platform {
			return a.Platform < b.Platform
		}
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.RepoName != b.RepoName {
			return a.RepoName < b.RepoName
		}
		return a.FileName < b.FileName
	})
	return entries, nil
}

// ReadWikiEntry returns the full markdown content of a wiki file.
func (a *ReferenceApp) ReadWikiEntry(source string, relPath string) (string, error) {
	root, err := resolveWikiRoot(source)
	if err != nil {
		return "", err
	}
	full := filepath.Join(root, relPath)
	if !utils.IsPathWithin(full, root) {
		return "", fmt.Errorf("路径越界")
	}
	data, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("文件不存在")
		}
		return "", fmt.Errorf("读取失败: %w", err)
	}
	return string(data), nil
}

// CheckWikiStatus returns the health status of a wiki file:
//   "ok"     — has frontmatter + body content
//   "no-fm"  — missing frontmatter entirely
//   "empty"  — file is empty or whitespace-only
// Called async by the frontend so the initial list loads fast.
func (a *ReferenceApp) CheckWikiStatus(source string, relPath string) (string, error) {
	root, err := resolveWikiRoot(source)
	if err != nil {
		return "error", err
	}
	full := filepath.Join(root, relPath)
	data, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return "empty", nil
		}
		return "error", err
	}
	content := string(data)
	if len(strings.TrimSpace(content)) == 0 {
		return "empty", nil
	}
	if !strings.HasPrefix(content, "---") {
		return "no-fm", nil
	}
	// has frontmatter — check it closes properly
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "no-fm", nil
	}
	return "ok", nil
}

// DeleteWikiEntry deletes a wiki file from disk. Guards against path traversal.
func (a *ReferenceApp) DeleteWikiEntry(source string, relPath string) error {
	root, err := resolveWikiRoot(source)
	if err != nil {
		return err
	}
	full := filepath.Join(root, relPath)
	if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(root)) {
		return fmt.Errorf("路径越界")
	}
	if _, err := os.Stat(full); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在")
	}
	return os.Remove(full)
}
// inside the wiki tree. Everything else at the top level (local/, forks/, etc.)
// is treated as a local repo.
var knownRemotePlatforms = map[string]bool{
	"github": true, "github.com": true,
	"gitlab.com": true, "bitbucket.org": true, "sourcehut": true,
	"gitee.com": true, "codeberg.org": true,
}

// parseWikiEntry builds a WikiEntry from a relative path + frontmatter.
// Handles both remote and local path patterns, auto-detecting "local" even
// inside the wiki/ directory (old versions stored local repos under wiki/local/).
//
//	remote:  <platform>/<namespace>/<repo>/<file.md>   (4+ parts)
//	local:   local/<repo>/<file.md>                    (3 parts, "local" prefix)
//	         <repo>/<file.md>                           (2 parts, legacy)
func parseWikiEntry(rel, source string) WikiEntry {
	parts := strings.Split(rel, "/")
	e := WikiEntry{
		Source:   source,
		RelPath:  rel,
		FileName: parts[len(parts)-1],
	}

	// detect local repos: explicit source, or path starts with "local/",
	// or top-level dir is not a known remote platform.
	isLocal := source == "local"
	if !isLocal && len(parts) >= 2 {
		topDir := strings.ToLower(parts[0])
		if topDir == "local" || !knownRemotePlatforms[topDir] {
			isLocal = true
		}
	}

	if isLocal {
		e.Source = "local"
		e.Platform = "本地"
		// local/<repo>/<file> or <repo>/<file>
		if len(parts) >= 3 && strings.ToLower(parts[0]) == "local" {
			e.RepoName = parts[1]
		} else if len(parts) >= 2 {
			e.RepoName = parts[0]
		}
	} else {
		// remote: <platform>/<namespace>/<repo>/<file>
		// also handle 3-part shallow paths: <platform>/<repo>/<file>
		if len(parts) >= 4 {
			e.Platform = parts[0]
			e.Namespace = parts[1]
			e.RepoName = parts[2]
		} else if len(parts) == 3 {
			e.Platform = parts[0]
			e.RepoName = parts[1]
		}
	}

	// best-effort frontmatter parse
	data, err := os.ReadFile(filepath.Join(resolveWikiRootSafe(source), rel))
	if err != nil {
		return e
	}
	fm := parseFrontmatter(string(data))
	e.Commit = fm["commit"]
	e.Branch = fm["branch"]
	e.Description = fm["description"]
	e.ExploredAt = fm["explored_at"]
	if e.Description == "" {
		e.Description = fm["repo"]
	}
	return e
}

// resolveWikiRootSafe is resolveWikiRoot without error return (for parseWikiEntry).
func resolveWikiRootSafe(source string) string {
	root, _ := resolveWikiRoot(source)
	return root
}

// parseFrontmatter extracts top-level key: value pairs from the YAML
// frontmatter block (between leading --- markers). Lightweight — no nested
// structures, no external yaml dependency.
func parseFrontmatter(content string) map[string]string {
	result := map[string]string{}
	// frontmatter must start at the very beginning
	if !strings.HasPrefix(content, "---") {
		return result
	}
	// find the closing ---
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return result
	}
	block := rest[:end]
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		// strip quotes
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') {
			val = val[1 : len(val)-1]
		}
		result[key] = val
	}
	return result
}
