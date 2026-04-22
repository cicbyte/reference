package repo

import (
	"fmt"

	"github.com/cicbyte/reference/internal/common"
	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

func getUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [identifier]",
		Short: "更新远程仓库缓存",
		Long: `更新指定或全部引用的远程缓存仓库（git pull）。

更新所有远程引用:
  reference update

更新指定引用:
  reference update github.com-gin-gonic-gin

本地引用会被跳过。`,
		Args: cobra.MaximumNArgs(1),
		Run:  runUpdateCommand,
	}
	return cmd
}

func runUpdateCommand(cmd *cobra.Command, args []string) {
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

	config := &logicrepo.UpdateConfig{
		Identifier: identifier,
		ProjectDir: projectDir,
	}

	processor := logicrepo.NewUpdateProcessor(config, common.AppConfigModel, db)
	if err := processor.Execute(cmd.Context()); err != nil {
		fmt.Printf("更新失败: %v\n", err)
	}
}
