package global

import (
	"context"
	"os"
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
	ProjectDir string           `json:"project_dir"`
	Exists     bool             `json:"exists"`
	RepoCount  int              `json:"repo_count"`
	Repos      []GlobalRepoItem `json:"repos,omitempty"`
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

		repoItems := make([]GlobalRepoItem, 0, len(repos))
		for _, r := range repos {
			refName := r.RefName
			if refName == "" {
				refName = r.LinkName
			}
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
			ProjectDir: dir,
			Exists:     dirExists,
			RepoCount:  len(repos),
			Repos:      repoItems,
		})
	}

	result.Total = len(result.Projects)
	return result, nil
}
