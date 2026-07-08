package main

import (
	"context"
	"fmt"

	"github.com/cicbyte/reference/internal/common"
	"github.com/cicbyte/reference/internal/logic/global"
	"github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// --- Repo methods ---

type RepoItem struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	CachePath string `json:"cache_path"`
	CommitAt  string `json:"commit_at"`
	Branch    string `json:"branch"`
}

func (a *ReferenceApp) ListRepos() ([]RepoItem, error) {
	projectDir, err := utils.GetGitRoot()
	if err != nil {
		return nil, fmt.Errorf("无法获取 Git 根目录: %w", err)
	}

	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}

	config := &repo.ListConfig{ProjectDir: projectDir}
	processor := repo.NewListProcessor(config, db)
	result, err := processor.Execute(context.Background())
	if err != nil {
		return nil, err
	}

	items := make([]RepoItem, len(result.Repos))
	for i, r := range result.Repos {
		items[i] = RepoItem{
			Type:      r.Type,
			Name:      r.Name,
			Source:    r.Source,
			CachePath: r.CachePath,
			CommitAt:  r.CommitAt,
			Branch:    r.Branch,
		}
	}
	return items, nil
}

func (a *ReferenceApp) AddRepo(target string, isLocal bool, name string, branch string) error {
	projectDir, err := utils.GetGitRoot()
	if err != nil {
		return err
	}

	db, err := utils.GetGormDB()
	if err != nil {
		return err
	}

	config := &repo.AddConfig{
		Target:     target,
		Local:      isLocal,
		Name:       name,
		Branch:     branch,
		ProjectDir: projectDir,
	}
	processor := repo.NewAddProcessor(config, common.AppConfigModel, db)
	_, err = processor.Execute(context.Background())
	return err
}

func (a *ReferenceApp) RemoveRepo(identifier string, purge bool) error {
	projectDir, err := utils.GetGitRoot()
	if err != nil {
		return err
	}

	db, err := utils.GetGormDB()
	if err != nil {
		return err
	}

	config := &repo.RemoveConfig{
		Identifier: identifier,
		Purge:      purge,
		Yes:        true,
		ProjectDir: projectDir,
	}
	processor := repo.NewRemoveProcessor(config, common.AppConfigModel, db)
	return processor.Execute(context.Background())
}

func (a *ReferenceApp) UpdateRepo(identifier string) error {
	projectDir, err := utils.GetGitRoot()
	if err != nil {
		return err
	}

	db, err := utils.GetGormDB()
	if err != nil {
		return err
	}

	config := &repo.UpdateConfig{
		Identifier: identifier,
		ProjectDir: projectDir,
	}
	processor := repo.NewUpdateProcessor(config, common.AppConfigModel, db)
	return processor.Execute(context.Background())
}

// --- SCC methods ---

type SCCResult struct {
	Repo      string        `json:"repo"`
	Languages []SCCLangStat `json:"languages"`
	TopFiles  []SCCFileStat `json:"topFiles"`
}

type SCCLangStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Code  int    `json:"code"`
}

type SCCFileStat struct {
	Type       string `json:"type"`
	File       string `json:"file"`
	Language   string `json:"language"`
	Code       int    `json:"code"`
	Complexity int    `json:"complexity"`
}

func (a *ReferenceApp) RunSCC(repoName string) (*SCCResult, error) {
	projectDir, err := utils.GetGitRoot()
	if err != nil {
		return nil, err
	}

	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}

	indexer := repo.NewRepoIndexer(db)
	repos, _ := indexer.List(projectDir)
	var repoPath string
	for _, r := range repos {
		if r.RefName == repoName {
			repoPath = r.CachePath
			break
		}
	}
	if repoPath == "" {
		return nil, fmt.Errorf("仓库 %s 未找到", repoName)
	}

	langStats, fileStats, err := repo.RunSCC(repoPath)
	if err != nil {
		return nil, err
	}

	langs := make([]SCCLangStat, len(langStats))
	for i, l := range langStats {
		langs[i] = SCCLangStat{
			Name:  l.Language,
			Count: int(l.Files),
			Code:  int(l.Code),
		}
	}

	files := make([]SCCFileStat, len(fileStats))
	for i, f := range fileStats {
		files[i] = SCCFileStat{
			Type:       f.Type,
			File:       f.Filename,
			Language:   f.Language,
			Code:       int(f.Code),
			Complexity: int(f.Complexity),
		}
	}

	return &SCCResult{Repo: repoName, Languages: langs, TopFiles: files}, nil
}

// --- Doctor methods ---

type DoctorResult struct {
	Checks  []DoctorCheck `json:"checks"`
	Summary string        `json:"summary"`
}

type DoctorCheck struct {
	Group   string `json:"group"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details"`
}

func (a *ReferenceApp) RunDoctor() (*DoctorResult, error) {
	projectDir, err := utils.GetGitRoot()
	if err != nil {
		return nil, err
	}

	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}

	config := &repo.DoctorConfig{ProjectDir: projectDir}
	processor := repo.NewDoctorProcessor(config, db)
	result, err := processor.Execute(context.Background())
	if err != nil {
		return nil, err
	}

	checks := make([]DoctorCheck, len(result.Checks))
	for i, c := range result.Checks {
		checks[i] = DoctorCheck{
			Group:   c.Group,
			Name:    c.Name,
			Status:  c.Status,
			Details: c.Details,
		}
	}

	return &DoctorResult{Checks: checks, Summary: result.Summary}, nil
}

// --- Global methods ---

func (a *ReferenceApp) GlobalList() ([]map[string]interface{}, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}

	processor := global.NewGlobalListProcessor(db)
	result, err := processor.Execute(context.Background())
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, len(result.Projects))
	for i, p := range result.Projects {
		items[i] = map[string]interface{}{
			"project_dir": p.ProjectDir,
			"repo_count":  p.RepoCount,
		}
	}
	return items, nil
}

func (a *ReferenceApp) GlobalStats() (map[string]interface{}, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}

	processor := global.NewGlobalStatsProcessor(common.AppConfigModel, db)
	result, err := processor.Execute(context.Background())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_projects": result.Projects.Total,
		"total_repos":    result.Repos.TotalCached,
		"cache_size":     result.CacheSize,
		"db_size":        result.DBSize,
	}, nil
}

func (a *ReferenceApp) GlobalGC(dryRun bool) (map[string]interface{}, error) {
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}

	config := &global.GlobalGCConfig{DryRun: dryRun}
	processor := global.NewGlobalGCProcessor(config, common.AppConfigModel, db)
	result, err := processor.Execute(context.Background())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"stale_records":  result.StaleDBRecords,
		"orphaned_caches": result.OrphanedCaches,
		"db_removed":     result.DBRecordsRemoved,
		"cache_removed":  result.CacheDirsRemoved,
		"dry_run":        result.DryRun,
	}, nil
}

// --- Wiki methods ---

func (a *ReferenceApp) WikiStatus() (map[string]interface{}, error) {
	wikiDir := utils.ConfigInstance.GetWikiDir()
	initialized := wiki.IsGitInitialized(wikiDir)
	remoteURL, _ := wiki.GetRemoteURL(wikiDir)

	status := "未初始化"
	if initialized {
		status = "已初始化"
	}

	return map[string]interface{}{
		"status": status,
		"path":   wikiDir,
		"remote": remoteURL,
	}, nil
}

func (a *ReferenceApp) WikiSync() error {
	wikiDir := utils.ConfigInstance.GetWikiDir()
	_, err := wiki.Sync(wikiDir)
	return err
}

// --- Proxy methods ---

func (a *ReferenceApp) GetProxyInfo() (map[string]interface{}, error) {
	cfg := common.AppConfigModel
	proxy := ""
	if cfg != nil {
		proxy = cfg.Network.Proxy
	}
	return map[string]interface{}{"proxy": proxy}, nil
}

func (a *ReferenceApp) SetProxy(proxyURL string) error {
	cfg := utils.ConfigInstance.LoadConfig()
	cfg.Network.Proxy = proxyURL
	return utils.ConfigInstance.SaveConfig(cfg)
}

func (a *ReferenceApp) ClearProxy() error {
	cfg := utils.ConfigInstance.LoadConfig()
	cfg.Network.Proxy = ""
	cfg.Network.GitProxy = ""
	return utils.ConfigInstance.SaveConfig(cfg)
}

// --- Init methods ---

type AgentInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

func (a *ReferenceApp) ListAgents() ([]AgentInfo, error) {
	ids := repo.ListAgentIDs()
	agents := make([]AgentInfo, len(ids))
	for i, id := range ids {
		agents[i] = AgentInfo{
			ID:          id,
			DisplayName: repo.GetAgentDisplayName(id),
		}
	}
	return agents, nil
}

func (a *ReferenceApp) InitProject(agents []string) error {
	projectDir, err := utils.GetGitRoot()
	if err != nil {
		return err
	}

	settings := models.LoadProjectSettings(projectDir)
	existing := make(map[string]bool)
	for _, a := range settings.Agents {
		existing[a] = true
	}
	for _, a := range agents {
		if !existing[a] {
			settings.Agents = append(settings.Agents, a)
		}
	}
	settings.Initialized = true
	return models.SaveProjectSettings(projectDir, settings)
}

// --- Window methods (Wails frameless window) ---
// Called from the Navbar's custom title-bar buttons. The frameless window has
// no native controls, so we drive the window via the Wails runtime using the
// context captured in startup(). The nil-ctx guard is defensive: Wails only
// dispatches bound calls after OnStartup, but window control should never
// panic even if invoked in a race with shutdown.

func (a *ReferenceApp) WindowMinimize() {
	a.appMu.RLock()
	ctx := a.ctx
	a.appMu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.WindowMinimise(ctx)
}

func (a *ReferenceApp) WindowMaximize() {
	// Toggle so the button also restores the window when already maximised
	// (matches native Windows title-bar behaviour).
	a.appMu.RLock()
	ctx := a.ctx
	a.appMu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.WindowToggleMaximise(ctx)
}

func (a *ReferenceApp) WindowClose() {
	a.appMu.RLock()
	ctx := a.ctx
	a.appMu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.Quit(ctx)
}
