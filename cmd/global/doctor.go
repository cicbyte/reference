package global

import (
	"fmt"

	logicglobal "github.com/cicbyte/reference/internal/logic/global"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func getDoctorCommand() *cobra.Command {
	var projectDir string
	var issuesOnly bool
	var concurrency int

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "批量跨项目诊断",
		Long: `对所有项目或指定项目执行诊断检查。

示例:
  reference global doctor                    # 诊断所有项目
  reference global doctor --project <path>   # 诊断指定项目
  reference global doctor --issues-only      # 只显示有问题的项目
  reference global doctor -f json            # JSON 输出`,
		Run: func(cmd *cobra.Command, args []string) {
			db, err := utils.GetGormDB()
			if err != nil {
				fmt.Printf("数据库连接失败: %v\n", err)
				return
			}

			config := &logicglobal.GlobalDoctorConfig{
				ProjectDir:  projectDir,
				IssuesOnly:  issuesOnly,
				Concurrency: concurrency,
			}

			processor := logicglobal.NewGlobalDoctorProcessor(config, db)
			result, err := processor.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("诊断失败: %v\n", err)
				return
			}

			f, _ := cmd.Flags().GetString("format")
			switch utils.ParseFormat(f) {
			case utils.FormatJSON:
				utils.OutputJSON(result)
			case utils.FormatJSONL:
				utils.OutputJSONL(result.Projects)
			default:
				renderGlobalDoctorTable(result)
			}
		},
	}

	cmd.Flags().StringVar(&projectDir, "project", "", "目标项目目录路径（默认诊断所有项目）")
	cmd.Flags().BoolVar(&issuesOnly, "issues-only", false, "只显示有问题的项目")
	cmd.Flags().IntVar(&concurrency, "concurrency", 8, "并发度")

	return cmd
}

func renderGlobalDoctorTable(result *logicglobal.GlobalDoctorResult) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter, WidthMax: 8},
		{Number: 2, Align: text.AlignLeft, WidthMax: 40},
		{Number: 3, Align: text.AlignCenter, WidthMax: 8},
		{Number: 4, Align: text.AlignLeft, WidthMax: 30},
	})
	t.AppendHeader(table.Row{"状态", "项目目录", "引用数", "Agent"})

	statusIcon := func(healthy bool) string {
		if healthy {
			return "OK"
		}
		return "!!"
	}

	for _, p := range result.Projects {
		agentStr := "-"
		if len(p.Agents) > 0 {
			agentStr = fmt.Sprintf("%v", p.Agents)
		}

		existStr := ""
		if !p.Exists {
			existStr = " (已删除)"
		}

		t.AppendRow(table.Row{
			statusIcon(p.Healthy),
			shortenPath(p.ProjectDir) + existStr,
			p.RepoCount,
			agentStr,
		})

		// 展开显示检查详情
		for _, c := range p.Checks {
			if c.Status != "ok" {
				t.AppendRow(table.Row{
					"",
					fmt.Sprintf("  └ %s", c.Name),
					c.Status,
					c.Details,
				})
			}
		}
	}

	t.AppendSeparator()
	t.AppendRow(table.Row{
		"",
		fmt.Sprintf("总计 %d 个项目", result.Summary.TotalProjects),
		"",
		fmt.Sprintf("健康: %d, 有问题: %d, 已删除: %d",
			result.Summary.Healthy,
			result.Summary.WithIssues,
			result.Summary.Deleted),
	})

	fmt.Println(t.Render())
}
