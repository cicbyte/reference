package utils

import (
	"path/filepath"
	"strings"
)

// IsPathWithin reports whether `child` is the same as or inside `parent`.
// It uses filepath.Rel to correctly reject prefix-lookalikes
// (e.g. parent="/a/repo", child="/a/repo-evil" → false), which a naive
// strings.HasPrefix check would wrongly accept.
//
// Both paths are cleaned first. Symlinks are NOT resolved — callers that
// need symlink-safe checks should filepath.EvalSymlinks first.
func IsPathWithin(child, parent string) bool {
	if child == "" || parent == "" {
		return false
	}
	c := filepath.Clean(child)
	p := filepath.Clean(parent)
	rel, err := filepath.Rel(p, c)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	// reject anything that escapes upward, and drive letters like "C:foo"
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	if filepath.IsAbs(rel) {
		return false
	}
	return true
}
