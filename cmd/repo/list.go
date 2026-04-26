package repo

import (
	"fmt"

	logicrepo "github.com/cicbyte/reference/internal/logic/repo"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func getListCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有引用仓库",
		Long:  `列出当前项目所有引用的仓库，显示类型、链接名、目标路径、commit、分支等信息。`,
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

			config := &logicrepo.ListConfig{ProjectDir: projectDir}
			processor := logicrepo.NewListProcessor(config, db)

			result, err := processor.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("列出失败: %v\n", err)
				return
			}

			if len(result.Repos) == 0 {
				fmt.Println("当前项目暂无引用仓库。\n运行 reference repo add <url> 添加引用。")
				return
			}

			f, _ := cmd.Flags().GetString("format")
			switch utils.ParseFormat(f) {
			case utils.FormatJSON:
				utils.OutputJSON(result)
			case utils.FormatJSONL:
				utils.OutputJSONL(result.Repos)
			default:
				renderListTable(result)
			}
		},
	}
	return cmd
}

func renderListTable(result *logicrepo.ListResult) {
	typeLabel := map[string]string{"remote": "远程", "local": "本地"}

	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter},
		{Number: 4, Align: text.AlignRight},
	})
	t.AppendHeader(table.Row{"类型", "名称", "来源", "源路径", "更新时间", "分支"})

	for _, r := range result.Repos {
		label := typeLabel[r.Type]
		if label == "" {
			label = r.Type
		}
		t.AppendRow(table.Row{label, r.Name, r.Source, r.CachePath, r.CommitAt, r.Branch})
	}

	fmt.Println(t.Render())
}
