package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
)

// =====================================================================
// Interactive project diagnosis & repair
//
// Replaces the "blind fix" DoctorProject flow with a per-repo diagnosis
// that surfaces the exact problem (missing cache / broken junction /
// moved local repo) and offers targeted fix actions.
// =====================================================================

// RepoDiagnosis is the health status of one referenced repo.
type RepoDiagnosis struct {
	RefName      string `json:"refName"`
	LinkName     string `json:"linkName"`
	Type         string `json:"type"`       // "remote" | "local"
	RemoteURL    string `json:"remoteUrl"`
	CachePath    string `json:"cachePath"`
	LocalPath    string `json:"localPath"`
	Branch       string `json:"branch"`
	TargetExists bool   `json:"targetExists"` // cache/local dir still present
	LinkExists   bool   `json:"linkExists"`   // .reference/repos/<refName> resolves
	WikiExists   bool   `json:"wikiExists"`   // .reference/wiki/<refName> resolves
	Status       string `json:"status"`       // ok / broken-link / missing-cache / missing-local / missing-wiki
	Suggestion   string `json:"suggestion"`
}

// statExists returns true if path exists (follows junctions/symlinks).
func statExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DiagnoseProject inspects every repo in a project and returns per-repo
// health status without making any changes.
func (a *ReferenceApp) DiagnoseProject(projectDir string) ([]RepoDiagnosis, error) {
	if projectDir == "" {
		return nil, fmt.Errorf("项目路径不能为空")
	}
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}
	indexer := repo.NewRepoIndexer(db)
	dbRepos, err := indexer.List(projectDir)
	if err != nil {
		return nil, err
	}

	reposDir := filepath.Join(projectDir, ".reference", "repos")
	wikiLinkDir := filepath.Join(projectDir, ".reference", "wiki")

	diagnoses := make([]RepoDiagnosis, len(dbRepos))
	for i, r := range dbRepos {
		refName := r.GetRefName()
		target := r.CachePath
		if target == "" {
			target = r.LocalPath
		}

		targetOK := statExists(target)
		linkOK := statExists(filepath.Join(reposDir, refName))
		wikiOK := statExists(filepath.Join(wikiLinkDir, refName))

		d := RepoDiagnosis{
			RefName:      refName,
			LinkName:     r.LinkName,
			Type:         string(r.RefType),
			RemoteURL:    r.RemoteURL,
			CachePath:    r.CachePath,
			LocalPath:    r.LocalPath,
			Branch:       r.Branch,
			TargetExists: targetOK,
			LinkExists:   linkOK,
			WikiExists:   wikiOK,
		}

		// determine status
		switch {
		case !targetOK && r.RefType == models.RefTypeRemote:
			d.Status = "missing-cache"
			d.Suggestion = "缓存目录不存在，可重新克隆"
		case !targetOK && r.RefType == models.RefTypeLocal:
			d.Status = "missing-local"
			d.Suggestion = "本地仓库路径不存在，请重新选择"
		case targetOK && !linkOK:
			d.Status = "broken-link"
			d.Suggestion = "Junction 断裂，可自动修复"
		case targetOK && linkOK && !wikiOK:
			d.Status = "missing-wiki"
			d.Suggestion = "Wiki 链接断裂，可自动修复"
		default:
			d.Status = "ok"
			d.Suggestion = ""
		}

		diagnoses[i] = d
	}
	return diagnoses, nil
}

// FixRepoLink rebuilds the repos + wiki junctions for a single repo.
// Only works when the target directory still exists.
func (a *ReferenceApp) FixRepoLink(projectDir string, refName string) error {
	if projectDir == "" || refName == "" {
		return fmt.Errorf("参数不能为空")
	}
	db, err := utils.GetGormDB()
	if err != nil {
		return err
	}
	indexer := repo.NewRepoIndexer(db)
	dbRepos, err := indexer.List(projectDir)
	if err != nil {
		return err
	}

	var target *models.Repo
	for i := range dbRepos {
		if dbRepos[i].GetRefName() == refName {
			target = &dbRepos[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("仓库 %s 未找到", refName)
	}

	// determine target path
	targetPath := target.CachePath
	if targetPath == "" {
		targetPath = target.LocalPath
	}
	if targetPath == "" || !statExists(targetPath) {
		return fmt.Errorf("目标目录不存在，无法重建链接")
	}

	reposDir := filepath.Join(projectDir, ".reference", "repos")
	wikiLinkDir := filepath.Join(projectDir, ".reference", "wiki")
	linkPath := filepath.Join(reposDir, refName)
	wikiLink := filepath.Join(wikiLinkDir, refName)

	// rebuild repos junction
	if err := repo.CreateLink(targetPath, linkPath); err != nil {
		return fmt.Errorf("重建仓库链接失败: %w", err)
	}

	// rebuild wiki junction (if wiki dir exists)
	wikiTarget := resolveWikiDirForRepo(target)
	if wikiTarget != "" && statExists(wikiTarget) {
		repo.CreateLink(wikiTarget, wikiLink) // best-effort
	}

	return nil
}

// RecloneRepo re-clones a remote repo and then rebuilds its junctions.
func (a *ReferenceApp) RecloneRepo(repoName string) error {
	projectDir, err := a.getCurrentProject()
	if err != nil {
		return err
	}
	db, err := utils.GetGormDB()
	if err != nil {
		return err
	}
	indexer := repo.NewRepoIndexer(db)
	dbRepos, err := indexer.List(projectDir)
	if err != nil {
		return err
	}
	var target *models.Repo
	for i := range dbRepos {
		if dbRepos[i].GetRefName() == repoName {
			target = &dbRepos[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("仓库 %s 未找到", repoName)
	}
	if target.RefType != models.RefTypeRemote {
		return fmt.Errorf("本地仓库无法重新克隆，请使用「选择新路径」")
	}
	if target.RemoteURL == "" {
		return fmt.Errorf("仓库 %s 缺少远程 URL", repoName)
	}

	if err := os.MkdirAll(filepath.Dir(target.CachePath), 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}
	if err := repo.CloneOrUpdate(repo.CloneOptions{
		URL:    target.RemoteURL,
		Branch: target.Branch,
		Path:   target.CachePath,
	}); err != nil {
		return err
	}

	// rebuild junctions now that the cache exists
	return a.FixRepoLink(projectDir, target.GetRefName())
}

// RelocateLocalRepo updates a local repo's path to a new location and
// rebuilds its junctions.
func (a *ReferenceApp) RelocateLocalRepo(projectDir string, refName string, newPath string) error {
	if projectDir == "" || refName == "" || newPath == "" {
		return fmt.Errorf("参数不能为空")
	}
	// validate new path is a git repo
	if err := repo.ValidateLocalRepo(newPath); err != nil {
		return fmt.Errorf("不是有效的 Git 仓库: %w", err)
	}
	abs, err := filepath.Abs(newPath)
	if err != nil {
		return fmt.Errorf("路径解析失败: %w", err)
	}

	db, err := utils.GetGormDB()
	if err != nil {
		return err
	}
	indexer := repo.NewRepoIndexer(db)
	dbRepos, err := indexer.List(projectDir)
	if err != nil {
		return err
	}

	var target *models.Repo
	for i := range dbRepos {
		if dbRepos[i].GetRefName() == refName {
			target = &dbRepos[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("仓库 %s 未找到", refName)
	}
	if target.RefType != models.RefTypeLocal {
		return fmt.Errorf("只有本地仓库支持重新定位")
	}

	// update DB record (Add does upsert)
	target.LocalPath = abs
	if err := indexer.Add(target); err != nil {
		return fmt.Errorf("更新数据库失败: %w", err)
	}

	// rebuild junctions
	return a.FixRepoLink(projectDir, refName)
}

// resolveWikiDirForRepo returns the wiki directory for a repo (mirrors
// inject.go's resolveWikiDir but takes a value receiver).
func resolveWikiDirForRepo(r *models.Repo) string {
	if r.RefType == models.RefTypeLocal {
		return filepath.Join(utils.ConfigInstance.GetLocalWikiDir(), r.WikiSubPath)
	}
	return filepath.Join(utils.ConfigInstance.GetWikiDir(), r.WikiSubPath)
}
