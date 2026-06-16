package cmd

import (
	"context"
	"fmt"
	"os"

	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

var initAgent string

var validAgents = map[string]string{
	"":         "",
	"none":     "",
	"claude":   "claude",
	"codex":    "codex",
	"opencode": "opencode",
	"zcode":    "zcode",
	"mimocode": "mimocode",
}

func getInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化项目配置",
		Long: `非交互式初始化当前项目的 reference 配置。

适用于 CI/CD 集成、自动化脚本等场景。执行后会：
  1. 创建 .reference/ 目录结构
  2. 生成 reference.settings.json
  3. 更新 .gitignore
  4. 生成 reference.map.jsonl
  5. 注入 AI 配置文件（如果指定了 --agent claude/codex/opencode/zcode/mimocode）

首次交互式引导请直接运行无参数的 reference 命令。`,
		RunE: runInit,
	}
	cmd.Flags().StringVar(&initAgent, "agent", "",
		"编程助手类型: claude | codex | opencode | zcode | mimocode | none（默认 none）")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	normalized, ok := validAgents[initAgent]
	if !ok {
		return fmt.Errorf("不支持的助手类型: %s（可选: claude, codex, opencode, zcode, mimocode, none）", initAgent)
	}

	if err := initProject(projectDir, normalized); err != nil {
		return fmt.Errorf("初始化失败: %w", err)
	}

	config := &logicrepo.InjectConfig{ProjectDir: projectDir}
	db, err := utils.GetGormDB()
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	processor := logicrepo.NewInjectProcessor(config, db)
	result, err := processor.Execute(context.Background())
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}
