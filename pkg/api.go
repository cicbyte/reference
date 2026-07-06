package reference

import (
	"context"
	"fmt"
	"time"

	"github.com/cicbyte/reference/internal/logic/repo"
	"go.uber.org/zap"
)

// AddRepo 添加仓库引用
func (e *Engine) AddRepo(ctx context.Context, opts AddRepoOptions) (*AddRepoResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEngineClosed
	}

	config := &repo.AddConfig{
		Target:     opts.Target,
		Local:      opts.Local,
		Name:       opts.Name,
		Branch:     opts.Branch,
		Update:     opts.Update,
		ProjectDir: opts.ProjectDir,
	}

	processor := repo.NewAddProcessor(config, e.appConfig, e.db)
	start := time.Now()
	result, err := processor.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("add repo: %w", err)
	}

	e.logger.Info("repo added",
		zap.String("ref_name", result.RefName),
		zap.String("ref_type", string(result.RefType)),
		zap.Duration("duration", time.Since(start)),
	)

	return &AddRepoResult{
		RefName: result.RefName,
		RefType: RefType(result.RefType),
	}, nil
}

// RemoveRepo 移除仓库引用
func (e *Engine) RemoveRepo(ctx context.Context, opts RemoveRepoOptions) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return ErrEngineClosed
	}

	config := &repo.RemoveConfig{
		Identifier: opts.Identifier,
		Purge:      opts.Purge,
		Clean:      opts.Clean,
		Yes:        opts.Yes,
		All:        opts.All,
		ProjectDir: opts.ProjectDir,
	}

	processor := repo.NewRemoveProcessor(config, e.appConfig, e.db)
	if err := processor.Execute(ctx); err != nil {
		return fmt.Errorf("remove repo: %w", err)
	}

	e.logger.Info("repo removed", zap.String("identifier", opts.Identifier))
	return nil
}

// ListRepos 列出项目中的所有引用仓库
func (e *Engine) ListRepos(ctx context.Context, projectDir string) (*ListReposResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEngineClosed
	}

	config := &repo.ListConfig{ProjectDir: projectDir}
	processor := repo.NewListProcessor(config, e.db)

	result, err := processor.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repos: %w", err)
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

	return &ListReposResult{Repos: items}, nil
}

// UpdateRepos 更新远程仓库缓存
func (e *Engine) UpdateRepos(ctx context.Context, opts UpdateReposOptions) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return ErrEngineClosed
	}

	config := &repo.UpdateConfig{
		Identifier: opts.Identifier,
		ProjectDir: opts.ProjectDir,
	}

	processor := repo.NewUpdateProcessor(config, e.appConfig, e.db)
	if err := processor.Execute(ctx); err != nil {
		return fmt.Errorf("update repos: %w", err)
	}

	e.logger.Info("repos updated", zap.String("identifier", opts.Identifier))
	return nil
}

// Doctor 诊断并修复引用健康状态
func (e *Engine) Doctor(ctx context.Context, projectDir string) (*DoctorResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEngineClosed
	}

	config := &repo.DoctorConfig{ProjectDir: projectDir}
	processor := repo.NewDoctorProcessor(config, e.db)

	result, err := processor.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("doctor: %w", err)
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

	return &DoctorResult{
		Checks:  checks,
		Summary: result.Summary,
	}, nil
}

// Inject 生成 reference.map.jsonl 并注入 AI Agent 配置
func (e *Engine) Inject(ctx context.Context, projectDir string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return "", ErrEngineClosed
	}

	if e.promptsFS == nil {
		return "", ErrPromptsFSNotSet
	}

	config := &repo.InjectConfig{ProjectDir: projectDir}
	processor := repo.NewInjectProcessor(config, e.db)

	result, err := processor.Execute(ctx)
	if err != nil {
		return "", fmt.Errorf("inject: %w", err)
	}

	e.logger.Info("injected", zap.String("project_dir", projectDir))
	return result, nil
}

// SCC 运行代码统计
func (e *Engine) SCC(repoPath string) (*SCCResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEngineClosed
	}

	langStats, fileStats, err := repo.RunSCC(repoPath)
	if err != nil {
		return nil, fmt.Errorf("scc: %w", err)
	}

	langs := make([]SCCLangStat, len(langStats))
	for i, l := range langStats {
		langs[i] = SCCLangStat{
			Type:       l.Type,
			Name:       l.Language,
			Count:      int(l.Files),
			Code:       int(l.Code),
			Complex:    int(l.Complexity),
		}
	}

	files := make([]SCCFileStat, len(fileStats))
	for i, f := range fileStats {
		files[i] = SCCFileStat{
			Type:     f.Type,
			File:     f.Filename,
			Language: f.Language,
			Code:     int(f.Code),
			Complex:  int(f.Complexity),
		}
	}

	return &SCCResult{
		Repo:     repoPath,
		Path:     repoPath,
		Langs:    langs,
		TopFiles: files,
	}, nil
}

// GetRepoIndexer 获取仓库索引器（高级用法）
func (e *Engine) GetRepoIndexer() *repo.RepoIndexer {
	return repo.NewRepoIndexer(e.db)
}
