package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/reference/cmd/global"
	"github.com/cicbyte/reference/cmd/proxy"
	"github.com/cicbyte/reference/cmd/repo"
	"github.com/cicbyte/reference/cmd/version"
	cmdwiki "github.com/cicbyte/reference/cmd/wiki"
	"github.com/cicbyte/reference/internal/common"
	"github.com/cicbyte/reference/internal/log"
	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "reference",
	Short: "本地代码仓库引用管理器",
	Long: `reference - 面向 AI 辅助编程时代的本地代码仓库引用管理器。

通过统一的全局缓存和项目级链接机制，让开发者及 AI 助手能够
以零网络延迟、零上下文污染的方式查阅任意远程或本地 Git 仓库的代码实现。

无参数运行时自动注入 AI Agent 配置（agent 文件 + SKILL.md + wiki 链接）。`,
	RunE: runDefault,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&globalFormat, "format", "f", "table", "输出格式 (table|json|jsonl)")
}

var globalFormat string

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func runDefault(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	settings := models.LoadProjectSettings(projectDir)
	if !settings.Initialized {
		guideInit(projectDir, settings)
	}

	config := &logicrepo.InjectConfig{ProjectDir: projectDir}
	db, err := utils.GetGormDB()
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	processor := logicrepo.NewInjectProcessor(config, db)
	result, err := processor.Execute(ctx)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func guideInit(projectDir string, settings *models.ProjectSettings) {
	fmt.Println()
	fmt.Println("  欢迎使用 reference！")
	fmt.Println()
	fmt.Println("  请选择你的编程助手：")
	fmt.Println("    [1] Claude Code")
	fmt.Println("    [2] 无（仅使用仓库引用管理功能）")
	fmt.Println()
	fmt.Print("  请输入选项 (1/2): ")

	var input string
	fmt.Scanln(&input)

	switch strings.TrimSpace(input) {
	case "1":
		settings.Agent = "claude"
		settings.Initialized = true
	case "2":
		settings.Agent = ""
		settings.Initialized = true
	default:
		settings.Agent = ""
		settings.Initialized = true
		fmt.Println("  未识别选项，已设为无编程助手。可通过 .reference/reference.settings.json 修改。")
	}

	if err := models.SaveProjectSettings(projectDir, settings); err != nil {
		fmt.Printf("  保存配置失败: %v\n", err)
		return
	}

	agentName := "无"
	if settings.Agent == "claude" {
		agentName = "Claude Code"
	}
	fmt.Printf("  已配置: %s\n", agentName)
	fmt.Println()
}

func init() {
	if err := utils.InitAppDirs(); err != nil {
		fmt.Printf("初始化目录失败: %v\n", err)
		os.Exit(1)
	}
	common.AppConfigModel = utils.ConfigInstance.LoadConfig()
	utils.ConfigInstance.ApplyConfig(common.AppConfigModel)
	if err := log.Init(utils.ConfigInstance.GetLogPath()); err != nil {
		fmt.Printf("日志初始化失败: %v\n", err)
		os.Exit(1)
	}
	if _, err := utils.GetGormDB(); err != nil {
		log.Error("数据库连接失败",
			zap.String("operation", "db init"),
			zap.Error(err))
		os.Exit(1)
	}
	log.Info("数据库连接成功")

	wikiDir := utils.ConfigInstance.GetWikiDir()
	if err := logicwiki.EnsureGitInit(wikiDir); err != nil {
		log.Warn("wiki git 初始化失败", zap.Error(err))
	} else {
		logicwiki.EnsureAutoPull(wikiDir)
	}

	rootCmd.AddCommand(global.GetGlobalCommand())
	rootCmd.AddCommand(repo.GetRepoCommand())
	rootCmd.AddCommand(repo.GetDoctorCommand())
	rootCmd.AddCommand(proxy.GetProxyCommand())
	rootCmd.AddCommand(cmdwiki.GetWikiCommand())
	rootCmd.AddCommand(getVersionCommand())
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("reference %s\n", version.Version)
			fmt.Printf("  commit: %s\n", version.GitCommit)
			fmt.Printf("  built:  %s\n", version.BuildTime)
		},
	}
}
