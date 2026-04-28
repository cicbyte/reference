package repo

import (
	"fmt"

	"github.com/cicbyte/reference/internal/common"
	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

var removePurge bool
var removeClean bool
var removeYes bool
var removeAll bool

func getRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <identifier>",
		Short: "移除仓库引用",
		Long: `移除当前项目中的某个引用。

移除单个引用:
  reference remove memos-cli
  reference remove memos-cli --purge --yes

移除全部引用:
  reference remove --all
  reference remove --all --yes
  reference remove --all --clean    # 同时清除注入的 AI 配置和 .reference/ 目录

对本地引用 --purge 无效。`,
		Run: runRemoveCommand,
	}

	cmd.Flags().BoolVarP(&removePurge, "purge", "p", false, "同时删除全局缓存仓库（仅远程模式）")
	cmd.Flags().BoolVar(&removeClean, "clean", false, "同时清除注入的 AI 配置和 .reference/ 目录（需配合 --all）")
	cmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "跳过确认提示")
	cmd.Flags().BoolVar(&removeAll, "all", false, "移除当前项目全部引用")

	return cmd
}

func runRemoveCommand(cmd *cobra.Command, args []string) {
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

	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}

	config := &logicrepo.RemoveConfig{
		Identifier: identifier,
		Purge:      removePurge,
		Clean:      removeClean,
		Yes:        removeYes,
		All:        removeAll,
		ProjectDir: projectDir,
	}

	processor := logicrepo.NewRemoveProcessor(config, common.AppConfigModel, db)
	if err := processor.Execute(cmd.Context()); err != nil {
		fmt.Printf("移除失败: %v\n", err)
	}
}
