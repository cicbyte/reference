package repo

import (
	"context"
	"path/filepath"

	"github.com/cicbyte/reference/internal/models"
	"gorm.io/gorm"
)

type ListConfig struct {
	ProjectDir string
}

type ListProcessor struct {
	config *ListConfig
	db     *gorm.DB
}

type ListResult struct {
	Repos []ListRepoItem `json:"repos"`
}

type ListRepoItem struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	CachePath string `json:"cache_path"`
	CommitAt  string `json:"commit_at"`
	Branch    string `json:"branch"`
}

func NewListProcessor(config *ListConfig, db *gorm.DB) *ListProcessor {
	return &ListProcessor{config: config, db: db}
}

func (p *ListProcessor) Execute(ctx context.Context) (*ListResult, error) {
	indexer := NewRepoIndexer(p.db)
	repos, err := indexer.List(p.config.ProjectDir)
	if err != nil {
		return nil, err
	}

	result := &ListResult{Repos: make([]ListRepoItem, 0, len(repos))}
	refDir := filepath.Join(p.config.ProjectDir, ".reference", "repos")

	for _, r := range repos {
		refName := r.RefName
		if refName == "" {
			refName = r.LinkName
		}
		linkPath := filepath.Join(refDir, refName)
		target := resolveTarget(linkPath, &r)
		cachePath := r.CachePath
		if cachePath == "" && r.LocalPath != "" {
			cachePath = r.LocalPath
		}

		typeStr := "remote"
		if r.RefType == models.RefTypeLocal {
			typeStr = "local"
		}

		commitTime := ""
		if r.CommitAt != nil {
			commitTime = r.CommitAt.Format("2006-01-02")
		}

		result.Repos = append(result.Repos, ListRepoItem{
			Type:      typeStr,
			Name:      refName,
			Source:    target,
			CachePath: cachePath,
			CommitAt:  commitTime,
			Branch:    r.Branch,
		})
	}

	return result, nil
}

func resolveTarget(linkPath string, r *models.Repo) string {
	if r.RefType == models.RefTypeRemote && r.RemoteURL != "" {
		return r.RemoteURL
	}
	if r.RefType == models.RefTypeLocal && r.LocalPath != "" {
		return r.LocalPath
	}
	if target, err := ReadLink(linkPath); err == nil && target != "" {
		return target
	}
	return "(无法解析目标)"
}
