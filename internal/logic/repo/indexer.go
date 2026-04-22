package repo

import (
	"github.com/cicbyte/reference/internal/models"
	"gorm.io/gorm"
)

type RepoIndexer struct {
	db *gorm.DB
}

func NewRepoIndexer(db *gorm.DB) *RepoIndexer {
	return &RepoIndexer{db: db}
}

func (idx *RepoIndexer) Add(repo *models.Repo) error {
	var existing models.Repo
	result := idx.db.Where("project_dir = ? AND link_name = ?", repo.ProjectDir, repo.LinkName).First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		return idx.db.Create(repo).Error
	}
	if result.Error != nil {
		return result.Error
	}
	repo.ID = existing.ID
	return idx.db.Save(repo).Error
}

func (idx *RepoIndexer) Remove(projectDir, linkName string) error {
	return idx.db.Where("project_dir = ? AND link_name = ?", projectDir, linkName).Delete(&models.Repo{}).Error
}

func (idx *RepoIndexer) Get(projectDir, linkName string) (*models.Repo, error) {
	var repo models.Repo
	err := idx.db.Where("project_dir = ? AND link_name = ?", projectDir, linkName).First(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (idx *RepoIndexer) GetByRefName(projectDir, refName string) (*models.Repo, error) {
	var repo models.Repo
	err := idx.db.Where("project_dir = ? AND ref_name = ?", projectDir, refName).First(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (idx *RepoIndexer) List(projectDir string) ([]models.Repo, error) {
	var repos []models.Repo
	err := idx.db.Where("project_dir = ?", projectDir).Find(&repos).Error
	return repos, err
}

func (idx *RepoIndexer) ListByType(projectDir string, refType models.RefType) ([]models.Repo, error) {
	var repos []models.Repo
	err := idx.db.Where("project_dir = ? AND ref_type = ?", projectDir, refType).Find(&repos).Error
	return repos, err
}
