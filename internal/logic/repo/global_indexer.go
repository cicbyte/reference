package repo

import "github.com/cicbyte/reference/internal/models"

func (idx *RepoIndexer) ListAllProjectDirs() ([]string, error) {
	var dirs []string
	err := idx.db.Model(&models.Repo{}).
		Distinct("project_dir").
		Order("project_dir").
		Pluck("project_dir", &dirs).Error
	return dirs, err
}

func (idx *RepoIndexer) DeleteByProjectDir(projectDir string) (int64, error) {
	result := idx.db.Where("project_dir = ?", projectDir).Delete(&models.Repo{})
	return result.RowsAffected, result.Error
}

func (idx *RepoIndexer) ListAllCachePaths() ([]string, error) {
	var paths []string
	err := idx.db.Model(&models.Repo{}).
		Where("ref_type = ? AND cache_path != ''", models.RefTypeRemote).
		Distinct("cache_path").
		Pluck("cache_path", &paths).Error
	return paths, err
}

func (idx *RepoIndexer) ListAllWikiSubPaths() ([]string, error) {
	var paths []string
	err := idx.db.Model(&models.Repo{}).
		Where("wiki_sub_path != ''").
		Distinct("wiki_sub_path").
		Pluck("wiki_sub_path", &paths).Error
	return paths, err
}

func (idx *RepoIndexer) ListAll() (map[string][]models.Repo, error) {
	var repos []models.Repo
	err := idx.db.Order("project_dir, link_name").Find(&repos).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string][]models.Repo)
	for _, r := range repos {
		result[r.ProjectDir] = append(result[r.ProjectDir], r)
	}
	return result, nil
}

func (idx *RepoIndexer) Count() (int64, error) {
	var count int64
	err := idx.db.Model(&models.Repo{}).Count(&count).Error
	return count, err
}
