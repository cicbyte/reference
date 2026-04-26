package repo

import (
	"fmt"

	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func GetDoctorCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "诊断并修复引用健康状态",
		Long: `检查并修复项目引用的各种问题：
  - 软链接是否完整
  - Agent 文件是否最新
  - SKILL.md 是否存在
  - Wiki Junction 是否正确
  - 数据库与文件系统是否一致`,
		Run: func(cmd *cobra.Command, args []string) {
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

			processor := logicrepo.NewDoctorProcessor(&logicrepo.DoctorConfig{ProjectDir: projectDir}, db)
			result, err := processor.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("  %s\n", err)
				return
			}

			f, _ := cmd.Flags().GetString("format")
			switch utils.ParseFormat(f) {
			case utils.FormatJSON:
				utils.OutputJSON(result)
			case utils.FormatJSONL:
				utils.OutputJSONL(result.Checks)
			default:
				renderDoctorTable(result)
			}
		},
	}
	return cmd
}

func renderDoctorTable(result *logicrepo.DoctorResult) {
	statusIcon := func(s string) string {
		switch s {
		case "ok", "fixed":
			return "OK"
		case "warn":
			return "!!"
		default:
			return "??"
		}
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter, WidthMax: 8},
		{Number: 2, Align: text.AlignLeft, WidthMax: 20},
		{Number: 3, Align: text.AlignLeft},
	})
	t.AppendHeader(table.Row{"状态", "检查项", "详情"})

	var lastGroup string
	for _, c := range result.Checks {
		if c.Group == "agent" && lastGroup == "core" {
			t.AppendSeparator()
		}
		lastGroup = c.Group
		t.AppendRow(table.Row{statusIcon(c.Status), c.Name, c.Details})
	}

	t.AppendSeparator()
	t.AppendRow(table.Row{"", "结果", result.Summary})

	fmt.Println(t.Render())
}
