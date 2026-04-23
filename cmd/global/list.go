package global

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logicglobal "github.com/cicbyte/reference/internal/logic/global"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func getListCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有项目的引用关系",
		Long:  `显示所有已注册项目及其引用的仓库，按项目分组显示，标记项目目录是否存在。`,
		Run: func(cmd *cobra.Command, args []string) {
			db, err := utils.GetGormDB()
			if err != nil {
				fmt.Printf("数据库连接失败: %v\n", err)
				return
			}

			processor := logicglobal.NewGlobalListProcessor(db)
			result, err := processor.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("查询失败: %v\n", err)
				return
			}

			if result.Total == 0 {
				fmt.Println("暂无任何项目引用记录。")
				return
			}

			switch utils.ParseFormat(format) {
			case utils.FormatJSON:
				utils.OutputJSON(result)
			case utils.FormatJSONL:
				utils.OutputJSONL(result.Projects)
			default:
				renderGlobalListTable(result)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "输出格式: table, json, jsonl")
	return cmd
}

func renderGlobalListTable(result *logicglobal.GlobalListResult) {
	typeLabel := map[string]string{"remote": "远程", "local": "本地"}

	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter, WidthMax: 6},
		{Number: 3, Align: text.AlignCenter, WidthMax: 4},
		{Number: 4, Align: text.AlignCenter, WidthMax: 6},
	})
	t.AppendHeader(table.Row{"存在", "项目目录", "数量", "类型", "引用名"})

	for _, proj := range result.Projects {
		existsMark := "OK"
		if !proj.Exists {
			existsMark = "!!"
		}

		shortDir := shortenPath(proj.ProjectDir)

		if len(proj.Repos) == 0 {
			t.AppendRow(table.Row{existsMark, shortDir, 0, "", "(无引用)"})
			continue
		}

		for i, repo := range proj.Repos {
			label := typeLabel[repo.Type]
			if label == "" {
				label = repo.Type
			}
			if i == 0 {
				t.AppendRow(table.Row{existsMark, shortDir, proj.RepoCount, label, repo.RefName})
			} else {
				t.AppendRow(table.Row{"", "", "", label, repo.RefName})
			}
		}
	}

	fmt.Println(t.Render())
}

func shortenPath(p string) string {
	cleaned := filepath.Clean(p)
	sep := string(os.PathSeparator)
	parts := strings.Split(cleaned, sep)
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], sep)
	}
	return cleaned
}
