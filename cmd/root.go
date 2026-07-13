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
	projectDir, err := utils.GetGitRoot()
	if err != nil {
		return fmt.Errorf("错误: %v", err)
	}

	settings := models.LoadProjectSettings(projectDir)
	if !settings.Initialized {
		guideInit(projectDir)
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

func initProject(projectDir string, agents []string) error {
	settings := models.LoadProjectSettings(projectDir)

	// 合并 agents，避免重复
	existing := make(map[string]bool)
	for _, a := range settings.Agents {
		existing[a] = true
	}
	for _, a := range agents {
		if !existing[a] {
			settings.Agents = append(settings.Agents, a)
		}
	}

	settings.Initialized = true
	return models.SaveProjectSettings(projectDir, settings)
}

func guideInit(projectDir string) {
	agentIDs := logicrepo.ListAgentIDs()

	fmt.Println()
	fmt.Println("  欢迎使用 reference！")
	fmt.Println()
	fmt.Println("  请选择你的编程助手（可多选，用逗号分隔）：")
	for i, id := range agentIDs {
		fmt.Printf("    [%d] %s\n", i+1, logicrepo.GetAgentDisplayName(id))
	}
	fmt.Printf("    [%d] 无（仅使用仓库引用管理功能）\n", len(agentIDs)+1)
	fmt.Println()
	fmt.Print("  请输入选项: ")

	var input string
	fmt.Scanln(&input)

	input = strings.TrimSpace(input)
	var agents []string

	if input == fmt.Sprintf("%d", len(agentIDs)+1) || input == "" {
		// 无 agent
	} else {
		parts := strings.Split(input, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			var agentID string
			for i, id := range agentIDs {
				if p == fmt.Sprintf("%d", i+1) || p == id {
					agentID = id
					break
				}
			}
			if agentID != "" {
				agents = append(agents, agentID)
			}
		}
		if len(agents) == 0 {
			fmt.Println("  未识别选项，已设为无编程助手。可通过 .reference/reference.settings.json 修改。")
		}
	}

	if err := initProject(projectDir, agents); err != nil {
		fmt.Printf("  保存配置失败: %v\n", err)
		return
	}

	if len(agents) == 0 {
		fmt.Println("  已配置: 无")
	} else {
		names := make([]string, len(agents))
		for i, id := range agents {
			names[i] = logicrepo.GetAgentDisplayName(id)
		}
		fmt.Printf("  已配置: %s\n", strings.Join(names, ", "))
	}
	fmt.Println()
}

func init() {
	if err := utils.InitAppDirs(); err != nil {
		fmt.Printf("初始化目录失败: %v\n", err)
		os.Exit(1)
	}
	common.SetAppConfig(utils.ConfigInstance.LoadConfig())
	utils.ConfigInstance.ApplyConfig(common.AppConfigModel)
	if err := utils.InitDataDirs(); err != nil {
		fmt.Printf("初始化数据目录失败: %v\n", err)
		os.Exit(1)
	}
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
	utils.MigratePathsIfNeeded()

	wikiDir := utils.ConfigInstance.GetWikiDir()
	if err := logicwiki.EnsureGitInit(wikiDir); err != nil {
		log.Warn("wiki git 初始化失败", zap.Error(err))
	}

	localWikiDir := utils.ConfigInstance.GetLocalWikiDir()
	if err := logicwiki.EnsureGitInit(localWikiDir); err != nil {
		log.Warn("localwiki git 初始化失败", zap.Error(err))
	}

	rootCmd.AddCommand(getInitCommand())
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
