package global

import (
	"fmt"

	logicglobal "github.com/cicbyte/reference/internal/logic/global"
	"github.com/cicbyte/reference/internal/common"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

func getRemoveCommand() *cobra.Command {
	var projectDir, repoName string
	var all, purge, yes bool

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "全局移除项目引用",
		Long: `从任意目录移除指定项目的仓库引用，无需 cd 到目标项目。

用法:
  reference global remove --project <路径> <仓库名>   # 移除指定项目的某个引用
  reference global remove --project <路径> --all      # 移除指定项目的所有引用
  reference global remove --repo <仓库名>             # 从所有项目中移除该仓库`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			db, err := utils.GetGormDB()
			if err != nil {
				fmt.Printf("数据库连接失败: %v\n", err)
				return
			}

			if len(args) > 0 && repoName == "" {
				repoName = args[0]
			}

			processor := logicglobal.NewGlobalRemoveProcessor(&logicglobal.GlobalRemoveConfig{
				ProjectDir: projectDir,
				RepoName:   repoName,
				All:        all,
				Purge:      purge,
				Yes:        yes,
			}, common.AppConfigModel, db)

			_, err = processor.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("移除失败: %v\n", err)
			}
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "项目目录路径")
	cmd.Flags().StringVar(&repoName, "repo", "", "仓库名（全局匹配所有项目）")
	cmd.Flags().BoolVar(&all, "all", false, "移除该项目的所有引用")
	cmd.Flags().BoolVar(&purge, "purge", false, "同时删除缓存目录")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认提示")
	return cmd
}
