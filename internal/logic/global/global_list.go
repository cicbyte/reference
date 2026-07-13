package global

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/models"
	"gorm.io/gorm"
)

type GlobalListConfig struct{}

type GlobalListProcessor struct {
	db *gorm.DB
}

type GlobalListResult struct {
	Projects []GlobalProjectItem `json:"projects"`
	Total    int                 `json:"total_projects"`
}

type GlobalProjectItem struct {
	ProjectDir  string           `json:"project_dir"`
	Exists      bool             `json:"exists"`
	Initialized bool             `json:"initialized"`
	Agents      []string         `json:"agents"`
	RepoCount   int              `json:"repo_count"`
	BrokenCount int              `json:"broken_count"`
	Repos       []GlobalRepoItem `json:"repos,omitempty"`
}

type GlobalRepoItem struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	RefName string `json:"ref_name"`
}

func NewGlobalListProcessor(db *gorm.DB) *GlobalListProcessor {
	return &GlobalListProcessor{db: db}
}

func (p *GlobalListProcessor) Execute(ctx context.Context) (*GlobalListResult, error) {
	indexer := repo.NewRepoIndexer(p.db)
	allRepos, err := indexer.ListAll()
	if err != nil {
		return nil, err
	}

	result := &GlobalListResult{
		Projects: make([]GlobalProjectItem, 0, len(allRepos)),
	}

	projectDirs := make([]string, 0, len(allRepos))
	for dir := range allRepos {
		projectDirs = append(projectDirs, dir)
	}
	sort.Strings(projectDirs)

	for _, dir := range projectDirs {
		repos := allRepos[dir]
		_, statErr := os.Stat(dir)
		dirExists := !os.IsNotExist(statErr)

		// 加载项目设置
		settings := models.LoadProjectSettings(dir)
		initialized := settings.Initialized
		agents := settings.Agents
		if agents == nil {
			agents = []string{}
		}

		// 轻量检查 broken_count
		brokenCount := 0
		if dirExists {
			reposDir := filepath.Join(dir, ".reference", "repos")
			for _, r := range repos {
				refName := r.GetRefName()
				linkPath := filepath.Join(reposDir, refName)
				if _, err := os.Stat(linkPath); err != nil {
					brokenCount++
				}
			}
		}

		repoItems := make([]GlobalRepoItem, 0, len(repos))
		for _, r := range repos {
			refName := r.GetRefName()
			typeStr := "local"
			if r.RefType == models.RefTypeRemote {
				typeStr = "remote"
			}
			repoItems = append(repoItems, GlobalRepoItem{
				Name:    r.LinkName,
				Type:    typeStr,
				RefName: refName,
			})
		}

		result.Projects = append(result.Projects, GlobalProjectItem{
			ProjectDir:  dir,
			Exists:      dirExists,
			Initialized: initialized,
			Agents:      agents,
			RepoCount:   len(repos),
			BrokenCount: brokenCount,
			Repos:       repoItems,
		})
	}

	result.Total = len(result.Projects)
	return result, nil
}
