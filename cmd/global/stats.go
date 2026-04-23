package global

import (
	"fmt"

	"github.com/cicbyte/reference/internal/common"
	logicglobal "github.com/cicbyte/reference/internal/logic/global"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func getStatsCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "显示全局统计信息",
		Long:  `显示全局引用系统的统计信息，包括项目数量、缓存大小、Wiki 大小和数据库大小。`,
		Run: func(cmd *cobra.Command, args []string) {
			db, err := utils.GetGormDB()
			if err != nil {
				fmt.Printf("数据库连接失败: %v\n", err)
				return
			}

			processor := logicglobal.NewGlobalStatsProcessor(common.AppConfigModel, db)
			result, err := processor.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("查询失败: %v\n", err)
				return
			}

			switch utils.ParseFormat(format) {
			case utils.FormatJSON:
				utils.OutputJSON(result)
			case utils.FormatJSONL:
				utils.OutputJSONL([]any{result})
			default:
				renderStatsTable(result)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "输出格式: table, json, jsonl")
	return cmd
}

func renderStatsTable(result *logicglobal.GlobalStatsResult) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignLeft, WidthMax: 20},
		{Number: 2, Align: text.AlignLeft},
	})
	t.AppendHeader(table.Row{"指标", "值"})

	t.AppendRow(table.Row{"项目总数", result.Projects.Total})
	t.AppendRow(table.Row{"  已存在", result.Projects.Existing})
	t.AppendRow(table.Row{"  已删除", result.Projects.Deleted})
	t.AppendSeparator()
	t.AppendRow(table.Row{"缓存仓库数", result.Repos.TotalCached})
	t.AppendRow(table.Row{"缓存总大小", humanize.Bytes(uint64(result.CacheSize))})
	t.AppendRow(table.Row{"Wiki 总大小", humanize.Bytes(uint64(result.WikiSize))})
	t.AppendRow(table.Row{"数据库大小", humanize.Bytes(uint64(result.DBSize))})

	fmt.Println(t.Render())
}
