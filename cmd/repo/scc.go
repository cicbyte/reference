package repo

import (
	"fmt"
	"sort"

	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

const defaultTopN = 15

func getSccCommand() *cobra.Command {
	var topN int
	var format string

	cmd := &cobra.Command{
		Use:   "scc [name]",
		Short: "查看仓库代码统计",
		Long:  `查看仓库代码统计信息，包括语言分布、代码行数、复杂度和 Top 文件排名。不指定 name 时显示所有仓库的汇总。`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, err := utils.GetProjectRoot()
			if err != nil {
				fmt.Printf("未找到项目根目录: %v\n", err)
				return
			}

			db, err := utils.GetGormDB()
			if err != nil {
				fmt.Printf("数据库连接失败: %v\n", err)
				return
			}

			indexer := logicrepo.NewRepoIndexer(db)
			repos, err := indexer.List(projectDir)
			if err != nil {
				fmt.Printf("查询仓库列表失败: %v\n", err)
				return
			}

			if len(repos) == 0 {
				fmt.Println("  当前项目没有引用仓库")
				return
			}

			if len(args) == 0 {
				for _, r := range repos {
					printRepoSCC(&r, projectDir, topN, format)
				}
				return
			}

			name := args[0]
			for _, r := range repos {
				if r.RefName == name || r.LinkName == name || r.RepoName == name || r.Namespace+"/"+r.RepoName == name {
					printRepoSCC(&r, projectDir, topN, format)
					return
				}
			}
			fmt.Printf("  未找到仓库: %s\n", name)
		},
	}

	cmd.Flags().IntVarP(&topN, "top", "n", defaultTopN, "Top 文件排名数量")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "输出格式: table, json, jsonl")
	return cmd
}

type sccResult struct {
	Repo     string                       `json:"repo"`
	Path     string                       `json:"path"`
	Langs    []logicrepo.SCCLanguageStat  `json:"languages"`
	TopFiles []logicrepo.SCCFileStat      `json:"topFiles"`
}

func printRepoSCC(r *models.Repo, projectDir string, topN int, format string) {
	sccPath := resolveSCCPath(r, projectDir)

	langStats, fileStats, err := logicrepo.RunSCC(sccPath)
	if err != nil {
		fmt.Printf("  [%s] 统计失败: %v\n", r.RefName, err)
		return
	}

	sort.Slice(fileStats, func(i, j int) bool {
		return fileStats[i].Code > fileStats[j].Code
	})
	if len(fileStats) > topN {
		fileStats = fileStats[:topN]
	}

	result := &sccResult{
		Repo:     r.RefName,
		Path:     sccPath,
		Langs:    langStats,
		TopFiles: fileStats,
	}

	switch utils.ParseFormat(format) {
	case utils.FormatJSON:
		utils.OutputJSON(result)
	case utils.FormatJSONL:
		for i := range langStats {
			langStats[i].Type = "language"
		}
		for i := range fileStats {
			fileStats[i].Type = "topFiles"
		}
		utils.OutputJSONL(langStats)
		utils.OutputJSONL(fileStats)
	default:
		fmt.Printf("  [%s] %s\n", r.RefName, sccPath)
		fmt.Print(logicrepo.FormatSummaryForCLI(langStats))
		fmt.Println()
		fmt.Print(logicrepo.FormatTopFilesForCLI(fileStats, topN))
		fmt.Println()
	}
}

func resolveSCCPath(r *models.Repo, projectDir string) string {
	if r.RefType == models.RefTypeRemote && r.CachePath != "" {
		return r.CachePath
	}
	if r.RefType == models.RefTypeLocal && r.LocalPath != "" {
		return r.LocalPath
	}
	return projectDir + "/.reference/" + r.RefName
}
