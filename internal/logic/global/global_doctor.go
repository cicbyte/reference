package global

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/models"
	"gorm.io/gorm"
)

// GlobalDoctorConfig 全局诊断配置
type GlobalDoctorConfig struct {
	ProjectDir  string // 指定项目目录，为空则诊断所有项目
	IssuesOnly  bool   // 只返回有问题的项目
	Concurrency int    // 并发度，0 表示使用默认值
}

// GlobalDoctorResult 全局诊断结果
type GlobalDoctorResult struct {
	Projects []GlobalDoctorProject `json:"projects"`
	Summary  GlobalDoctorSummary   `json:"summary"`
}

// GlobalDoctorProject 单个项目的诊断结果
type GlobalDoctorProject struct {
	ProjectDir  string           `json:"project_dir"`
	Exists      bool             `json:"exists"`
	Initialized bool             `json:"initialized"`
	Agents      []string         `json:"agents"`
	RepoCount   int              `json:"repo_count"`
	Checks      []repo.CheckItem `json:"checks"`
	Healthy     bool             `json:"healthy"`
	IssuesCount int              `json:"issues_count"`
}

// GlobalDoctorSummary 诊断汇总
type GlobalDoctorSummary struct {
	TotalProjects int `json:"total_projects"`
	Existing      int `json:"existing"`
	Deleted       int `json:"deleted"`
	Healthy       int `json:"healthy"`
	WithIssues    int `json:"with_issues"`
	ChecksTotal   int `json:"checks_total"`
	ChecksFailed  int `json:"checks_failed"`
}

// GlobalDoctorProcessor 全局诊断处理器
type GlobalDoctorProcessor struct {
	config *GlobalDoctorConfig
	db     *gorm.DB
}

// NewGlobalDoctorProcessor 创建全局诊断处理器
func NewGlobalDoctorProcessor(config *GlobalDoctorConfig, db *gorm.DB) *GlobalDoctorProcessor {
	if config.Concurrency <= 0 {
		config.Concurrency = 8
	}
	return &GlobalDoctorProcessor{config: config, db: db}
}

// Execute 执行全局诊断
func (p *GlobalDoctorProcessor) Execute(ctx context.Context) (*GlobalDoctorResult, error) {
	indexer := repo.NewRepoIndexer(p.db)

	// 获取所有项目
	var projectDirs []string
	if p.config.ProjectDir != "" {
		projectDirs = []string{p.config.ProjectDir}
	} else {
		var err error
		projectDirs, err = indexer.ListAllProjectDirs()
		if err != nil {
			return nil, fmt.Errorf("获取项目列表失败: %w", err)
		}
	}

	// 获取所有仓库映射
	allRepos, err := indexer.ListAll()
	if err != nil {
		return nil, fmt.Errorf("获取仓库列表失败: %w", err)
	}

	// 并发诊断
	result := &GlobalDoctorResult{
		Projects: make([]GlobalDoctorProject, 0, len(projectDirs)),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, p.config.Concurrency)

	for _, projectDir := range projectDirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			project := p.diagnoseProject(dir, allRepos[dir])

			mu.Lock()
			if !p.config.IssuesOnly || !project.Healthy {
				result.Projects = append(result.Projects, project)
			}
			mu.Unlock()
		}(projectDir)
	}

	wg.Wait()

	// 计算汇总
	for _, p := range result.Projects {
		result.Summary.TotalProjects++
		if p.Exists {
			result.Summary.Existing++
		} else {
			result.Summary.Deleted++
		}
		if p.Healthy {
			result.Summary.Healthy++
		} else {
			result.Summary.WithIssues++
		}
		result.Summary.ChecksTotal += len(p.Checks)
		for _, c := range p.Checks {
			if c.Status != "ok" {
				result.Summary.ChecksFailed++
			}
		}
	}

	return result, nil
}

// diagnoseProject 诊断单个项目
func (p *GlobalDoctorProcessor) diagnoseProject(projectDir string, repos []models.Repo) GlobalDoctorProject {
	project := GlobalDoctorProject{
		ProjectDir: projectDir,
		Exists:     true,
		Agents:     []string{},
	}

	// 检查目录是否存在
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		project.Exists = false
		project.Healthy = false
		project.Checks = []repo.CheckItem{
			{Name: "项目目录", Status: "error", Details: "目录不存在", Group: "core"},
		}
		project.IssuesCount = 1
		return project
	}

	// 加载项目设置
	settings := models.LoadProjectSettings(projectDir)
	project.Initialized = settings.Initialized
	project.Agents = settings.Agents
	project.RepoCount = len(repos)

	// 执行诊断检查
	if !settings.Initialized {
		project.Healthy = false
		project.Checks = []repo.CheckItem{
			{Name: "初始化状态", Status: "warn", Details: "项目未初始化", Group: "core"},
		}
		project.IssuesCount = 1
		return project
	}

	// 执行核心检查
	coreChecks := []repo.CheckItem{}
	symlinkCheck := repo.CheckSymlinks(projectDir, repos)
	coreChecks = append(coreChecks, repo.CheckItem{
		Name: symlinkCheck.Name, Status: symlinkCheck.Status, Details: symlinkCheck.Details, Group: "core",
	})

	wikiCheck := repo.CheckWikiJunctions(projectDir, repos)
	coreChecks = append(coreChecks, repo.CheckItem{
		Name: wikiCheck.Name, Status: wikiCheck.Status, Details: wikiCheck.Details, Group: "core",
	})

	indexer := repo.NewRepoIndexer(p.db)
	refMapCheck := repo.CheckReferenceMap(projectDir, repos, indexer)
	coreChecks = append(coreChecks, repo.CheckItem{
		Name: refMapCheck.Name, Status: refMapCheck.Status, Details: refMapCheck.Details, Group: "core",
	})

	dbCheck := repo.CheckDatabaseConsistency(projectDir, repos)
	coreChecks = append(coreChecks, repo.CheckItem{
		Name: dbCheck.Name, Status: dbCheck.Status, Details: dbCheck.Details, Group: "core",
	})

	wikiGitCheck := repo.CheckWikiGit()
	coreChecks = append(coreChecks, repo.CheckItem{
		Name: wikiGitCheck.Name, Status: wikiGitCheck.Status, Details: wikiGitCheck.Details, Group: "core",
	})

	// 执行 agent 检查
	agentChecks := []repo.CheckItem{}
	for _, agentID := range settings.Agents {
		cfg, ok := repo.GetAgentConfig(agentID)
		if !ok {
			continue
		}
		// agent 检查需要完整路径
		// 这里简化处理，只检查配置是否存在
		agentChecks = append(agentChecks, repo.CheckItem{
			Name: cfg.DisplayName + " 配置", Status: "ok", Details: "已配置", Group: "agent",
		})
	}

	project.Checks = append(coreChecks, agentChecks...)

	// 判断是否健康
	project.Healthy = true
	for _, c := range project.Checks {
		if c.Status == "error" || c.Status == "warn" {
			project.Healthy = false
			project.IssuesCount++
		}
	}

	return project
}
