package repo

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AddConfig struct {
	Target     string
	Local      bool
	Name       string
	Branch     string
	Depth      int
	NoUpdate   bool
	ProjectDir string
}

type AddResult struct {
	RefName string
	RefType models.RefType
}

type AddProcessor struct {
	config    *AddConfig
	appConfig *models.AppConfig
	db        *gorm.DB
}

func NewAddProcessor(config *AddConfig, appConfig *models.AppConfig, db *gorm.DB) *AddProcessor {
	return &AddProcessor{config: config, appConfig: appConfig, db: db}
}

func (p *AddProcessor) Execute(ctx context.Context) (*AddResult, error) {
	refDir := filepath.Join(p.config.ProjectDir, ".reference")
	reposDir := filepath.Join(refDir, "repos")
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return nil, fmt.Errorf("创建 .reference/repos 目录失败: %w", err)
	}

	var result *AddResult
	var err error
	if p.config.Local {
		result, err = p.addLocal(reposDir)
	} else {
		result, err = p.addRemote(reposDir)
	}
	if err != nil {
		return nil, err
	}

	injectProc := NewInjectProcessor(&InjectConfig{ProjectDir: p.config.ProjectDir}, p.db)
	if _, injectErr := injectProc.Execute(ctx); injectErr != nil {
		log.Warn("自动注入 Skill 失败", zap.Error(injectErr))
	}

	return result, nil
}

func (p *AddProcessor) addRemote(refDir string) (*AddResult, error) {
	reposBase := utils.ConfigInstance.GetReposDirFromConfig(p.appConfig)

	info, err := ParseGitURL(p.config.Target, reposBase)
	if err != nil {
		return nil, err
	}

	proxy := ResolveProxy(p.appConfig)
	err = CloneOrUpdate(CloneOptions{
		URL:      info.OriginalURL,
		Path:     info.CachePath,
		Branch:   p.config.Branch,
		Depth:    p.config.Depth,
		Proxy:    proxy,
		NoUpdate: p.config.NoUpdate,
	})
	if err != nil {
		return nil, err
	}

	linkName := p.config.Name
	if linkName == "" {
		linkName = info.LinkName
	}

	indexer := NewRepoIndexer(p.db)

	// 已存在则更新，不创建新链接
	if existing, err := indexer.Get(p.config.ProjectDir, linkName); err == nil {
		branch, commit, commitTime, _ := GetRepoMeta(info.CachePath)
		existing.Branch = branch
		existing.Commit = commit
		existing.CommitAt = commitTime
		existing.WikiSubPath = info.WikiSubPath
		if err := indexer.Add(existing); err != nil {
			return nil, fmt.Errorf("更新数据库索引失败: %w", err)
		}
		refName := existing.RefName
		if refName == "" {
			refName = existing.LinkName
		}
		return &AddResult{RefName: refName, RefType: models.RefTypeRemote}, nil
	}

	refName := resolveRemoteRefName(p.config.ProjectDir, indexer, info.RepoName, info.Namespace, info.Host, linkName)
	linkPath := filepath.Join(refDir, refName)

	if err := CreateLink(info.CachePath, linkPath); err != nil {
		return nil, fmt.Errorf("创建链接失败: %w", err)
	}

	if err := EnsureGitignore(p.config.ProjectDir); err != nil {
		log.Warn("更新 .gitignore 失败", zap.Error(err))
	}

	branch, commit, commitTime, _ := GetRepoMeta(info.CachePath)

	repo := &models.Repo{
		ProjectDir:  p.config.ProjectDir,
		LinkName:    linkName,
		RefName:     refName,
		RefType:     models.RefTypeRemote,
		RemoteURL:   info.OriginalURL,
		Host:        info.Host,
		Namespace:   info.Namespace,
		RepoName:    info.RepoName,
		CachePath:   info.CachePath,
		WikiSubPath: info.WikiSubPath,
		Branch:      branch,
		Commit:      commit,
		CommitAt:    commitTime,
	}
	if err := indexer.Add(repo); err != nil {
		RemoveLink(linkPath)
		return nil, fmt.Errorf("写入数据库索引失败: %w", err)
	}

	return &AddResult{RefName: refName, RefType: models.RefTypeRemote}, nil
}

func (p *AddProcessor) addLocal(refDir string) (*AddResult, error) {
	absPath, err := filepath.Abs(p.config.Target)
	if err != nil {
		return nil, fmt.Errorf("解析路径失败: %w", err)
	}

	if err := ValidateLocalRepo(absPath); err != nil {
		return nil, err
	}

	linkName := p.config.Name
	if linkName == "" {
		linkName = "local-" + shortHash(absPath) + "-" + filepath.Base(absPath)
	}

	indexer := NewRepoIndexer(p.db)

	// 已存在则更新
	if existing, err := indexer.Get(p.config.ProjectDir, linkName); err == nil {
		branch, commit, commitTime, _ := GetRepoMeta(absPath)
		existing.Branch = branch
		existing.Commit = commit
		existing.CommitAt = commitTime
		if err := indexer.Add(existing); err != nil {
			return nil, fmt.Errorf("更新数据库索引失败: %w", err)
		}
		refName := existing.RefName
		if refName == "" {
			refName = existing.LinkName
		}
		return &AddResult{RefName: refName, RefType: models.RefTypeLocal}, nil
	}

	refName := resolveLocalRefName(p.config.ProjectDir, indexer, filepath.Base(absPath), linkName)
	linkPath := filepath.Join(refDir, refName)

	wikiSubPath := resolveLocalWikiSubPath(utils.ConfigInstance.GetWikiDir(), absPath)

	if err := CreateLink(absPath, linkPath); err != nil {
		return nil, fmt.Errorf("创建链接失败: %w", err)
	}

	if err := EnsureGitignore(p.config.ProjectDir); err != nil {
		log.Warn("更新 .gitignore 失败", zap.Error(err))
	}

	branch, commit, commitTime, _ := GetRepoMeta(absPath)

	repo := &models.Repo{
		ProjectDir:  p.config.ProjectDir,
		LinkName:    linkName,
		RefName:     refName,
		RefType:     models.RefTypeLocal,
		LocalPath:   absPath,
		WikiSubPath: wikiSubPath,
		Branch:      branch,
		Commit:      commit,
		CommitAt:    commitTime,
	}
	if err := indexer.Add(repo); err != nil {
		RemoveLink(linkPath)
		return nil, fmt.Errorf("写入数据库索引失败: %w", err)
	}

	return &AddResult{RefName: refName, RefType: models.RefTypeLocal}, nil
}

func shortHash(s string) string {
	h := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", h[:2])
}

// resolveRemoteRefName 远程仓库引用名：repoName → namespace-repoName → fullLinkName
func resolveRemoteRefName(projectDir string, indexer *RepoIndexer, repoName, namespace, host, fullLinkName string) string {
	candidates := []string{repoName, namespace + "-" + repoName, fullLinkName}
	for _, name := range candidates {
		if _, err := indexer.GetByRefName(projectDir, name); err != nil {
			return name
		}
	}
	return fullLinkName
}

// resolveLocalRefName 本地仓库引用名：dirname → fullLinkName
func resolveLocalRefName(projectDir string, indexer *RepoIndexer, dirName, fullLinkName string) string {
	if _, err := indexer.GetByRefName(projectDir, dirName); err != nil {
		return dirName
	}
	return fullLinkName
}

func resolveLocalWikiSubPath(wikiBase, absPath string) string {
	dirName := filepath.Base(absPath)
	subPath := "local/" + dirName
	wikiDir := filepath.Join(wikiBase, subPath)

	if _, err := os.Stat(wikiDir); os.IsNotExist(err) {
		return subPath
	}

	h := shortHash(absPath)
	return "local/" + h + "-" + dirName
}

func FormatAddResult(result *AddResult, duration time.Duration) string {
	typeStr := "本地"
	if result.RefType == models.RefTypeRemote {
		typeStr = "远程"
	}
	return fmt.Sprintf("[%s] 引用 '%s' 添加成功 (%s)", typeStr, result.RefName, duration.Round(time.Millisecond))
}
