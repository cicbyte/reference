package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RemoveConfig struct {
	Identifier string
	Purge      bool
	Yes        bool
	All        bool
	ProjectDir string
}

type RemoveProcessor struct {
	config    *RemoveConfig
	appConfig *models.AppConfig
	db        *gorm.DB
}

func NewRemoveProcessor(config *RemoveConfig, appConfig *models.AppConfig, db *gorm.DB) *RemoveProcessor {
	return &RemoveProcessor{config: config, appConfig: appConfig, db: db}
}

func (p *RemoveProcessor) Execute(ctx context.Context) error {
	if p.config.All {
		return p.removeAll()
	}
	return p.removeOne()
}

func (p *RemoveProcessor) removeAll() error {
	refDir := filepath.Join(p.config.ProjectDir, ".reference")
	reposLinkDir := filepath.Join(refDir, "repos")
	wikiLinkDir := filepath.Join(refDir, "wiki")

	indexer := NewRepoIndexer(p.db)
	repos, err := indexer.List(p.config.ProjectDir)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		cleaned := p.cleanOrphanedJunctions(reposLinkDir, wikiLinkDir)
		os.Remove(filepath.Join(refDir, "reference.map.jsonl"))
		if cleaned > 0 {
			fmt.Printf("数据库无记录，已清理 %d 个残留链接\n", cleaned)
		} else {
			fmt.Println("当前项目暂无引用仓库。")
		}
		return nil
	}

	if !p.config.Yes {
		fmt.Printf("确认移除全部 %d 个引用? [y/N]: ", len(repos))
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" {
			fmt.Println("已取消")
			return nil
		}
	}

	reposBase := utils.ConfigInstance.GetReposDirFromConfig(p.appConfig)
	removed := 0

	for _, r := range repos {
		refName := r.RefName
		if refName == "" {
			refName = r.LinkName
		}

		linkPath := filepath.Join(reposLinkDir, refName)
		RemoveLink(linkPath)

		junctionPath := filepath.Join(wikiLinkDir, refName)
		if _, err := os.Lstat(junctionPath); err == nil {
			removeLink(junctionPath)
		}

		if p.config.Purge && r.RefType == models.RefTypeRemote && r.CachePath != "" {
			if strings.HasPrefix(r.CachePath, reposBase) {
				if err := PurgeCache(r.CachePath); err != nil {
					log.Warn("删除缓存失败", zap.String("repo", refName), zap.Error(err))
				}
			}
		}

		if err := indexer.Remove(p.config.ProjectDir, r.LinkName); err != nil {
			log.Warn("删除数据库记录失败", zap.String("repo", refName), zap.Error(err))
		}
		removed++
	}

	removed += p.cleanOrphanedJunctions(reposLinkDir, wikiLinkDir)
	os.Remove(filepath.Join(refDir, "reference.map.jsonl"))

	fmt.Printf("已移除 %d 个引用\n", removed)
	return nil
}

func (p *RemoveProcessor) cleanOrphanedJunctions(reposDir, wikiDir string) int {
	cleaned := 0
	for _, dir := range []string{reposDir, wikiDir} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			fullPath := filepath.Join(dir, e.Name())
			info, err := os.Lstat(fullPath)
			if err != nil || (!info.IsDir() && info.Mode()&os.ModeSymlink == 0) {
				continue
			}
			RemoveLink(fullPath)
			cleaned++
		}
	}
	return cleaned
}

func (p *RemoveProcessor) removeOne() error {
	refDir := filepath.Join(p.config.ProjectDir, ".reference")
	reposLinkDir := filepath.Join(refDir, "repos")
	wikiLinkDir := filepath.Join(refDir, "wiki")

	indexer := NewRepoIndexer(p.db)
	repo, err := indexer.Get(p.config.ProjectDir, p.config.Identifier)
	if err != nil {
		repo, err = indexer.GetByRefName(p.config.ProjectDir, p.config.Identifier)
	}
	if err != nil {
		linkPath := filepath.Join(reposLinkDir, p.config.Identifier)
		if _, statErr := os.Lstat(linkPath); statErr != nil {
			return fmt.Errorf("未找到引用: %s", p.config.Identifier)
		}
		repo = &models.Repo{RefName: p.config.Identifier}
	}

	refName := repo.RefName
	if refName == "" {
		refName = repo.LinkName
	}

	linkPath := filepath.Join(reposLinkDir, refName)
	if err := RemoveLink(linkPath); err != nil {
		return fmt.Errorf("删除链接失败: %w", err)
	}

	if repo.RefType == models.RefTypeLocal {
		fmt.Println("本地引用源目录不会被删除")
	} else if p.config.Purge && repo.CachePath != "" {
		reposBase := utils.ConfigInstance.GetReposDirFromConfig(p.appConfig)
		if strings.HasPrefix(repo.CachePath, reposBase) {
			confirmed := p.config.Yes
			if !confirmed {
				fmt.Printf("确认删除缓存目录? %s [y/N]: ", repo.CachePath)
				var input string
				fmt.Scanln(&input)
				confirmed = strings.ToLower(input) == "y"
				if !confirmed {
					fmt.Println("已取消")
				}
			}
			if confirmed {
				if err := PurgeCache(repo.CachePath); err != nil {
					return fmt.Errorf("删除缓存失败: %w", err)
				}
				log.Info("缓存已清除", zap.String("path", repo.CachePath))
			}
		}
	}

	if err := indexer.Remove(p.config.ProjectDir, repo.LinkName); err != nil {
		log.Warn("删除数据库记录失败", zap.Error(err))
	}

	wikiJunctionPath := filepath.Join(wikiLinkDir, refName)
	if _, err := os.Lstat(wikiJunctionPath); err == nil {
		removeLink(wikiJunctionPath)
	}

	// 更新 reference.map.jsonl
	if err := RefreshReferenceMap(p.config.ProjectDir, refDir, indexer); err != nil {
		log.Warn("更新 reference.map.jsonl 失败", zap.Error(err))
	}

	fmt.Printf("引用 '%s' 已移除\n", refName)
	return nil
}

func RefreshReferenceMap(projectDir, refDir string, indexer *RepoIndexer) error {
	repos, err := indexer.List(projectDir)
	if err != nil || len(repos) == 0 {
		return os.Remove(filepath.Join(refDir, "reference.map.jsonl"))
	}
	var rdList []repoData
	for _, r := range repos {
		refName := r.RefName
		if refName == "" {
			refName = r.LinkName
		}
		rd := repoData{
			LinkName: r.LinkName,
			RefName:  refName,
			Type:     string(r.RefType),
			WikiDir: filepath.Join(utils.ConfigInstance.GetAppDir(), "wiki", r.WikiSubPath),
		}
		if r.RefType == models.RefTypeRemote {
			rd.Platform = r.Host
			rd.FullName = r.Namespace + "/" + r.RepoName
		} else {
			rd.Platform = "local"
			rd.FullName = filepath.Base(r.LocalPath)
		}
		reposDir := filepath.Join(refDir, "repos", refName)
		rd.Description = detectDescription(reposDir, &r)
		rdList = append(rdList, rd)
	}
	return generateReferenceMap(refDir, rdList)
}
