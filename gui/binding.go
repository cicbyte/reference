package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cicbyte/reference/cmd/version"
	"github.com/cicbyte/reference/internal/common"
	"github.com/cicbyte/reference/internal/logic/global"
	"github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// --- Project methods ---
// Multi-project support: an explicit "current project" that scopes all
// project-level operations (repo list / add / scc / doctor / init). Switched
// from the left-hand project rail in the UI.

// ProjectItem is one row in the project rail (sourced from GlobalList, full fields).
type ProjectItem struct {
	Dir         string   `json:"dir"`
	Name        string   `json:"name"`
	Exists      bool     `json:"exists"`
	Initialized bool     `json:"initialized"`
	Agents      []string `json:"agents"`
	RepoCount   int      `json:"repoCount"`
	BrokenCount int      `json:"brokenCount"`
}

// ProjectInfo describes the currently active project for the rail/status bar.
type ProjectInfo struct {
	Dir    string `json:"dir"`
	Name   string `json:"name"`
	Exists bool   `json:"exists"`
}

// ListProjects returns all known projects from the DB (full fields, unlike the
// trimmed GlobalList). Powers the left-hand project rail.
func (a *ReferenceApp) ListProjects() ([]ProjectItem, error) {
	a.waitReady()
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}
	processor := global.NewGlobalListProcessor(db)
	result, err := processor.Execute(context.Background())
	if err != nil {
		return nil, err
	}
	items := make([]ProjectItem, len(result.Projects))
	for i, p := range result.Projects {
		agents := p.Agents
		if agents == nil {
			agents = []string{}
		}
		items[i] = ProjectItem{
			Dir:         p.ProjectDir,
			Name:        baseName(p.ProjectDir),
			Exists:      p.Exists,
			Initialized: p.Initialized,
			Agents:      agents,
			RepoCount:   p.RepoCount,
			BrokenCount: p.BrokenCount,
		}
	}
	return items, nil
}

// SwitchProject sets the active project dir. No existence check — allows
// switching to a project whose dir was removed (UI shows a warning instead).
func (a *ReferenceApp) SwitchProject(dir string) (ProjectInfo, error) {
	if dir == "" {
		return ProjectInfo{}, fmt.Errorf("项目路径不能为空")
	}
	a.setCurrentProject(dir)
	return ProjectInfo{Dir: dir, Name: baseName(dir), Exists: dirExists(dir)}, nil
}

// GetCurrentProject returns the active project, falling back to GetGitRoot if
// none was explicitly switched. Returns empty ProjectInfo (no error) if there
// is no project context at all — the UI shows its empty state.
func (a *ReferenceApp) GetCurrentProject() (ProjectInfo, error) {
	a.waitReady()
	dir, err := a.getCurrentProject()
	if err != nil {
		return ProjectInfo{}, nil // no error: empty state is valid
	}
	return ProjectInfo{Dir: dir, Name: baseName(dir), Exists: dirExists(dir)}, nil
}

// PickProjectFolder opens a native directory picker and returns the chosen
// path (empty string if cancelled). Used by the rail's "+ add project" button.
func (a *ReferenceApp) PickProjectFolder() (string, error) {
	a.appMu.RLock()
	ctx := a.ctx
	a.appMu.RUnlock()
	if ctx == nil {
		return "", nil
	}
	dir, err := wailsruntime.OpenDirectoryDialog(ctx, wailsruntime.OpenDialogOptions{
		Title: "选择项目目录",
	})
	if err != nil {
		// On Windows the native dialog reports cancellation as an error
		// ("shellItem is nil"); treat that as a normal empty selection.
		if strings.Contains(err.Error(), "shellItem") {
			return "", nil
		}
		return "", err
	}
	return dir, nil
}

// PickFolder opens a native directory picker with a caller-supplied dialog
// title (empty string if cancelled). Generic folder picker for settings forms.
func (a *ReferenceApp) PickFolder(title string) (string, error) {
	a.appMu.RLock()
	ctx := a.ctx
	a.appMu.RUnlock()
	if ctx == nil {
		return "", nil
	}
	if title == "" {
		title = "选择目录"
	}
	dir, err := wailsruntime.OpenDirectoryDialog(ctx, wailsruntime.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		if strings.Contains(err.Error(), "shellItem") {
			return "", nil
		}
		return "", err
	}
	return dir, nil
}

// --- Project actions (context menu) ---
// These take an explicit projectDir so they can operate on the right-clicked
// project even if it isn't the currently active one.

// RemoveProject deletes all of a project's reference records from the DB and
// removes the junction links. When clean=true it also removes the .reference/
// directory and any injected AI config files.
func (a *ReferenceApp) RemoveProject(projectDir string, clean bool) error {
	if projectDir == "" {
		return fmt.Errorf("项目路径不能为空")
	}
	db, err := utils.GetGormDB()
	if err != nil {
		return err
	}
	config := &repo.RemoveConfig{
		ProjectDir: projectDir,
		All:        true,
		Clean:      clean,
		Yes:        true, // GUI confirms via the menu, skip CLI prompt
	}
	processor := repo.NewRemoveProcessor(config, common.AppConfigModel, db)
	if err := processor.Execute(context.Background()); err != nil {
		return err
	}
	// if we just removed the active project, clear it
	a.appMu.RLock()
	cur := a.currentProject
	a.appMu.RUnlock()
	if cur == projectDir {
		a.setCurrentProject("")
	}
	return nil
}

// DoctorProject runs diagnosis + auto-repair on a specific project (broken
// junctions are rebuilt, reference.map.jsonl regenerated). Returns the same
// DoctorResult shape as RunDoctor so the UI can reuse rendering.
func (a *ReferenceApp) DoctorProject(projectDir string) (result *DoctorResult, err error) {
	if projectDir == "" {
		return nil, fmt.Errorf("项目路径不能为空")
	}
	// recover from panics in doctor checks — log the stack trace so we can
	// diagnose the root cause, and return a clean error to the frontend.
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("DoctorProject panic: %v\n%s\n", r, debug.Stack())
			err = fmt.Errorf("诊断内部错误: %v", r)
			result = nil
		}
	}()
	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}
	// timeout so a stuck check (e.g. mklink on a dead network path) doesn't
	// hang the UI forever.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := &repo.DoctorConfig{ProjectDir: projectDir}
	processor := repo.NewDoctorProcessor(config, db)
	res, err := processor.Execute(ctx)
	if err != nil {
		return nil, err
	}
	checks := make([]DoctorCheck, len(res.Checks))
	for i, c := range res.Checks {
		checks[i] = DoctorCheck{
			Group:   c.Group,
			Name:    c.Name,
			Status:  c.Status,
			Details: c.Details,
		}
	}
	return &DoctorResult{Checks: checks, Summary: res.Summary}, nil
}

// OpenInExplorer reveals the project directory in the platform file manager
// (Explorer on Windows, Finder on macOS, xdg-open on Linux).
func (a *ReferenceApp) OpenInExplorer(dir string) error {
	if dir == "" {
		return fmt.Errorf("路径不能为空")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	return cmd.Start()
}

// CopyPath copies text to the system clipboard via the Wails runtime.
func (a *ReferenceApp) CopyPath(text string) error {
	a.appMu.RLock()
	ctx := a.ctx
	a.appMu.RUnlock()
	if ctx == nil {
		return fmt.Errorf("应用未就绪")
	}
	return wailsruntime.ClipboardSetText(ctx, text)
}

// --- Repo methods ---

type RepoItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Source      string `json:"source"`
	CachePath   string `json:"cache_path"`
	CommitAt    string `json:"commit_at"`
	Branch      string `json:"branch"`
	RemoteURL   string `json:"remoteUrl"`
	CacheExists bool   `json:"cacheExists"`
}

func (a *ReferenceApp) ListRepos() ([]RepoItem, error) {
	a.waitReady()
	projectDir, err := a.getCurrentProject()
	if err != nil {
		return nil, fmt.Errorf("无法获取 Git 根目录: %w", err)
	}

	db, err := utils.GetGormDB()
	if err != nil {
		return nil, err
	}

	// query DB directly to get full Repo fields (RemoteURL, CachePath, etc.)
	indexer := repo.NewRepoIndexer(db)
	dbRepos, err := indexer.List(projectDir)
	if err != nil {
		return nil, err
	}

	items := make([]RepoItem, len(dbRepos))
	for i, r := range dbRepos {
		path := r.CachePath
		if path == "" {
			path = r.LocalPath
		}
		_, statErr := os.Stat(path)
		items[i] = RepoItem{
			Type:        string(r.RefType),
			Name:        r.GetRefName(),
			Source:      r.RemoteURL,
			CachePath:   r.CachePath,
			CommitAt:    formatCommitAt(r.CommitAt),
			Branch:      r.Branch,
			RemoteURL:   r.RemoteURL,
			CacheExists: !os.IsNotExist(statErr),
		}
	}
	return items, nil
}

// formatCommitAt formats a *time.Time as a date string (empty if nil).
func formatCommitAt(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func (a *ReferenceApp) AddRepo(target string, isLocal bool, name string, branch string) error {
	projectDir, err := a.getCurrentProject()
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
	projectDir, err := a.getCurrentProject()
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
	projectDir, err := a.getCurrentProject()
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

// RecloneRepo re-clones a remote repo whose cache was deleted. Finds the repo
// by refName in the current project, then re-clones from RemoteURL to CachePath.
func (a *ReferenceApp) RunSCC(repoName string) (*SCCResult, error) {
	projectDir, err := a.getCurrentProject()
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
	projectDir, err := a.getCurrentProject()
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
	a.waitReady()
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
		"total_projects":  result.Projects.Total,
		"existing_projects": result.Projects.Existing,
		"deleted_projects":  result.Projects.Deleted,
		"total_repos":     result.Repos.TotalCached,
		"cache_size":      result.CacheSize,
		"wiki_size":       result.WikiSize,
		"db_size":         result.DBSize,
		"repos_dir":       utils.ConfigInstance.GetReposDir(),
		"wiki_dir":        utils.ConfigInstance.GetWikiDir(),
	}, nil
}

// GetDirSizeAsync returns the total size (bytes) of a directory. Called
// async by the frontend so the stats page loads instantly — dirSize on
// large repos/wiki trees can take several seconds.
func (a *ReferenceApp) GetDirSizeAsync(path string) (int64, error) {
	if path == "" {
		return 0, nil
	}
	return dirSizeAsync(path)
}

func dirSizeAsync(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
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
	return size, err
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
	DisplayName string `json:"displayName"`
	BaseDir     string `json:"baseDir"`
	FileCount   int    `json:"fileCount"`
}

func (a *ReferenceApp) ListAgents() ([]AgentInfo, error) {
	ids := repo.ListAgentIDs()
	agents := make([]AgentInfo, len(ids))
	for i, id := range ids {
		cfg, ok := repo.GetAgentConfig(id)
		if !ok {
			continue
		}
		agents[i] = AgentInfo{
			ID:          cfg.ID,
			DisplayName: cfg.DisplayName,
			BaseDir:     cfg.BaseDir,
			FileCount:   len(cfg.Files),
		}
	}
	return agents, nil
}

func (a *ReferenceApp) InitProject(agents []string) error {
	projectDir, err := a.getCurrentProject()
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

// --- AppConfig methods ---
// Unified config read/write so the settings page can manage storage paths,
// network and log settings in one place. These supersede the narrow
// GetProxyInfo/SetProxy/ClearProxy trio (kept for compatibility).

// AppConfigDTO is the shape the settings page reads. Paths marked "actual"
// are resolved from the Config singleton (post-ApplyConfig) and are read-only.
type AppConfigDTO struct {
	// editable
	ReposPath string `json:"reposPath"`
	WikiPath  string `json:"wikiPath"`
	Network   struct {
		Proxy    string `json:"proxy"`
		GitProxy string `json:"gitProxy"`
		Timeout  int    `json:"timeout"`
	} `json:"network"`
	Log struct {
		Level      string `json:"level"`
		MaxSize    int    `json:"maxSize"`
		MaxBackups int    `json:"maxBackups"`
		MaxAge     int    `json:"maxAge"`
		Compress   bool   `json:"compress"`
	} `json:"log"`
	// read-only resolved paths (informational)
	Paths struct {
		Config string `json:"config"`
		Db     string `json:"db"`
		LogDir string `json:"logDir"`
		Repos  string `json:"repos"` // actual repos dir (after ApplyConfig)
		Wiki   string `json:"wiki"`  // actual wiki dir
	} `json:"paths"`
}

func (a *ReferenceApp) GetAppConfig() (*AppConfigDTO, error) {
	a.waitReady()
	cfg := utils.ConfigInstance.LoadConfig() // fresh from disk
	dto := &AppConfigDTO{
		ReposPath: cfg.ReposPath,
		WikiPath:  cfg.WikiPath,
	}
	dto.Network.Proxy = cfg.Network.Proxy
	dto.Network.GitProxy = cfg.Network.GitProxy
	dto.Network.Timeout = cfg.Network.Timeout
	dto.Log.Level = cfg.Log.Level
	dto.Log.MaxSize = cfg.Log.MaxSize
	dto.Log.MaxBackups = cfg.Log.MaxBackups
	dto.Log.MaxAge = cfg.Log.MaxAge
	dto.Log.Compress = cfg.Log.Compress
	dto.Paths.Config = utils.ConfigInstance.GetConfigPath()
	dto.Paths.Db = utils.ConfigInstance.GetDbPath()
	dto.Paths.LogDir = utils.ConfigInstance.GetLogDir()
	dto.Paths.Repos = utils.ConfigInstance.GetReposDir()
	dto.Paths.Wiki = utils.ConfigInstance.GetWikiDir()
	return dto, nil
}

// SaveAppConfig applies a partial patch to config.yaml. Only the editable
// fields are honoured; unknown keys are ignored. After writing it re-syncs the
// in-memory singleton (common.AppConfigModel) and re-applies paths, fixing the
// stale-memory bug the old SetProxy had.
func (a *ReferenceApp) SaveAppConfig(patch map[string]interface{}) error {
	cfg := utils.ConfigInstance.LoadConfig() // merge onto disk-latest

	// top-level paths
	if v, ok := patch["reposPath"].(string); ok {
		cfg.ReposPath = v
	}
	if v, ok := patch["wikiPath"].(string); ok {
		cfg.WikiPath = v
	}
	// network group
	if net, ok := patch["network"].(map[string]interface{}); ok {
		if v, ok := net["proxy"].(string); ok {
			cfg.Network.Proxy = v
		}
		if v, ok := net["gitProxy"].(string); ok {
			cfg.Network.GitProxy = v
		}
		if v, ok := toInt(net["timeout"]); ok {
			cfg.Network.Timeout = v
		}
	}
	// log group (currently read-only in UI but supported for completeness)
	if lg, ok := patch["log"].(map[string]interface{}); ok {
		if v, ok := lg["level"].(string); ok {
			cfg.Log.Level = v
		}
		if v, ok := toInt(lg["maxSize"]); ok {
			cfg.Log.MaxSize = v
		}
		if v, ok := toInt(lg["maxBackups"]); ok {
			cfg.Log.MaxBackups = v
		}
		if v, ok := toInt(lg["maxAge"]); ok {
			cfg.Log.MaxAge = v
		}
		if v, ok := lg["compress"].(bool); ok {
			cfg.Log.Compress = v
		}
	}

	if err := utils.ConfigInstance.SaveConfig(cfg); err != nil {
		return err
	}
	// keep the in-memory bridge consistent so subsequent reads see new values
	common.SetAppConfig(cfg)
	utils.ConfigInstance.ApplyConfig(cfg)
	// If repos_path changed, migrate DB cache_path records to the new location
	// so GC / remove --purge / global list keep working without a restart.
	utils.MigratePathsIfNeeded()
	return nil
}

// GetVersionInfo exposes the ldflags-injected version triplet for the About tab.
func (a *ReferenceApp) GetVersionInfo() map[string]interface{} {
	return map[string]interface{}{
		"version":   version.Version,
		"commit":    version.GitCommit,
		"buildTime": version.BuildTime,
	}
}

// toInt coerces JSON numbers (float64) to int.
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	default:
		return 0, false
	}
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
	wailsruntime.WindowMinimise(ctx)
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
	wailsruntime.WindowToggleMaximise(ctx)
}

func (a *ReferenceApp) WindowClose() {
	a.appMu.RLock()
	ctx := a.ctx
	a.appMu.RUnlock()
	if ctx == nil {
		return
	}
	wailsruntime.Quit(ctx)
}
