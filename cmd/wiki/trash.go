package wiki

import (
	"fmt"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func getTrashCommand() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "trash",
		Short: "查看被删除的知识文件",
		Run: func(cmd *cobra.Command, args []string) {
			wikiDir := utils.ConfigInstance.GetWikiDir()
			files, err := logicwiki.ListDeletedFiles(wikiDir, limit)
			if err != nil {
				fmt.Printf("查询失败: %v\n", err)
				return
			}
			if len(files) == 0 {
				fmt.Println("没有删除记录")
				return
			}

			f, _ := cmd.Flags().GetString("format")
			switch utils.ParseFormat(f) {
			case utils.FormatJSON:
				utils.OutputJSON(files)
			case utils.FormatJSONL:
				utils.OutputJSONL(files)
			default:
				renderTrashTable(files)
				fmt.Printf("\n共 %d 条记录，使用 reference wiki restore <path> 恢复\n", len(files))
			}
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "显示最近 N 条")
	return cmd
}

func renderTrashTable(files []logicwiki.DeletedFile) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter, WidthMax: 10},
		{Number: 2, Align: text.AlignRight, WidthMax: 12},
		{Number: 3, Align: text.AlignLeft},
	})
	t.AppendHeader(table.Row{"提交", "日期", "路径"})

	for _, f := range files {
		dateStr := f.Date
		if len(dateStr) > 10 {
			dateStr = dateStr[:10]
		}
		t.AppendRow(table.Row{f.Commit, dateStr, f.Path})
	}

	fmt.Println(t.Render())
}
