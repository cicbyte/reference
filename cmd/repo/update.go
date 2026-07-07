package repo

import (
	"fmt"
	"path/filepath"

	"github.com/cicbyte/reference/internal/common"
	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

var updateProjectDir string

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

	cmd.Flags().StringVar(&updateProjectDir, "project", "", "目标项目目录路径（默认当前目录）")

	return cmd
}

func runUpdateCommand(cmd *cobra.Command, args []string) {
	var projectDir string
	var err error

	if updateProjectDir != "" {
		projectDir, err = filepath.Abs(updateProjectDir)
		if err != nil {
			fmt.Printf("解析项目路径失败: %v\n", err)
			return
		}
	} else {
		projectDir, err = utils.GetGitRoot()
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			return
		}
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
