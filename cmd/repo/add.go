package repo

import (
	"fmt"
	"time"

	"github.com/cicbyte/reference/internal/common"
	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

var (
	addLocal    bool
	addName     string
	addBranch   string
	addDepth    int
	addNoUpdate bool
)

func getAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <target>",
		Short: "添加仓库引用",
		Long: `添加一个远程 Git 仓库或本地 Git 仓库到当前项目的引用中。

远程模式（默认）:
  reference add https://github.com/gin-gonic/gin
  reference add spf13/cobra --depth 1

本地模式:
  reference add --local ~/projects/my-lib --name my-utils`,
		Args: cobra.ExactArgs(1),
		Run:  runAddCommand,
	}

	cmd.Flags().BoolVarP(&addLocal, "local", "l", false, "标记为本地 Git 仓库路径")
	cmd.Flags().StringVarP(&addName, "name", "n", "", "自定义链接名称")
	cmd.Flags().StringVarP(&addBranch, "branch", "b", "", "指定克隆的分支或标签（仅远程模式）")
	cmd.Flags().IntVarP(&addDepth, "depth", "d", 1, "浅克隆深度（仅远程模式）")
	cmd.Flags().BoolVar(&addNoUpdate, "no-update", false, "若缓存已存在，跳过 git pull（仅远程模式）")

	return cmd
}

func runAddCommand(cmd *cobra.Command, args []string) {
	if err := validateAddParams(args); err != nil {
		fmt.Printf("参数验证失败: %v\n", err)
		cmd.Help()
		return
	}

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

	target := args[0]
	if !addLocal {
		target = logicrepo.NormalizeGitURL(target)
	}

	config := &logicrepo.AddConfig{
		Target:     target,
		Local:      addLocal,
		Name:       addName,
		Branch:     addBranch,
		Depth:      addDepth,
		NoUpdate:   addNoUpdate,
		ProjectDir: projectDir,
	}

	processor := logicrepo.NewAddProcessor(config, common.AppConfigModel, db)
	start := time.Now()
	result, err := processor.Execute(cmd.Context())
	if err != nil {
		fmt.Printf("添加失败: %v\n", err)
		return
	}

	fmt.Println(logicrepo.FormatAddResult(result, time.Since(start)))
}

func validateAddParams(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("必须指定目标（Git URL 或本地路径）")
	}
	return nil
}
