package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

var initAgent string
var initProjectDir string

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
  5. 注入 AI 配置文件（如果指定了 --agent）

首次交互式引导请直接运行无参数的 reference 命令。`,
		RunE: runInit,
	}
	cmd.Flags().StringVar(&initAgent, "agent", "",
		"编程助手类型，多个用逗号分隔: claude | codex | opencode | zcode | mimocode | none（默认 none）")
	cmd.Flags().StringVar(&initProjectDir, "project", "",
		"目标项目目录路径（默认当前目录）")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	var projectDir string
	var err error

	if initProjectDir != "" {
		projectDir, err = filepath.Abs(initProjectDir)
		if err != nil {
			return fmt.Errorf("解析项目路径失败: %w", err)
		}
	} else {
		projectDir, err = utils.GetGitRoot()
		if err != nil {
			return fmt.Errorf("错误: %v", err)
		}
	}

	var agents []string
	agentValue := strings.TrimSpace(initAgent)
	if agentValue != "" && agentValue != "none" {
		for _, part := range strings.Split(agentValue, ",") {
			id := strings.TrimSpace(part)
			if id != "" {
				agents = append(agents, id)
			}
		}
		if invalid := logicrepo.ValidateAgentIDs(agents); len(invalid) > 0 {
			allIDs := strings.Join(logicrepo.ListAgentIDs(), ", ")
			return fmt.Errorf("不支持的助手类型: %s（可选: %s）", strings.Join(invalid, ", "), allIDs)
		}
	}

	if err := initProject(projectDir, agents); err != nil {
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
