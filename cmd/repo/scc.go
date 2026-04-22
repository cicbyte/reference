package repo

import (
	"fmt"

	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

func getSccCommand() *cobra.Command {
	var topFiles bool
	var format string

	cmd := &cobra.Command{
		Use:   "scc [name]",
		Short: "查看仓库代码统计",
		Long:  `查看仓库代码统计信息，包括语言分布、代码行数和复杂度。不指定 name 时显示所有仓库的汇总。`,
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
					printRepoSCC(&r, projectDir, topFiles, format)
				}
				return
			}

			name := args[0]
			for _, r := range repos {
				if r.RefName == name || r.LinkName == name || r.RepoName == name || r.Namespace+"/"+r.RepoName == name {
					printRepoSCC(&r, projectDir, topFiles, format)
					return
				}
			}
			fmt.Printf("  未找到仓库: %s\n", name)
		},
	}

	cmd.Flags().BoolVarP(&topFiles, "top", "t", false, "显示 top 文件排名")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "输出格式: table, json, jsonl")
	return cmd
}

type sccResult struct {
	Repo   string                   `json:"repo"`
	Path   string                   `json:"path"`
	Langs  []logicrepo.SCCLanguageStat `json:"languages"`
	Files  []logicrepo.SCCFileStat     `json:"files,omitempty"`
}

func printRepoSCC(r *models.Repo, projectDir string, topFiles bool, format string) {
	sccPath := resolveSCCPath(r, projectDir)

	langStats, fileStats, err := logicrepo.RunSCC(sccPath)
	if err != nil {
		fmt.Printf("  [%s] 统计失败: %v\n", r.RefName, err)
		return
	}

	result := &sccResult{
		Repo:  r.RefName,
		Path:  sccPath,
		Langs: langStats,
	}
	if topFiles {
		result.Files = fileStats
	}

	switch utils.ParseFormat(format) {
	case utils.FormatJSON:
		utils.OutputJSON(result)
	case utils.FormatJSONL:
		utils.OutputJSONL(result.Langs)
		if topFiles {
			utils.OutputJSONL(result.Files)
		}
	default:
		fmt.Printf("  [%s] %s\n", r.RefName, sccPath)
		fmt.Print(logicrepo.FormatSummaryForCLI(langStats))
		if topFiles {
			fmt.Println()
			fmt.Print(logicrepo.FormatTopFilesForCLI(fileStats, 15))
		}
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
