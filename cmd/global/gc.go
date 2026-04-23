package global

import (
	"fmt"
	"strings"

	"github.com/cicbyte/reference/internal/common"
	logicglobal "github.com/cicbyte/reference/internal/logic/global"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

func getGCCommand() *cobra.Command {
	var dryRun, yes, cache bool

	cmd := &cobra.Command{
		Use:   "gc",
		Short: "清理过期记录和孤立缓存",
		Long: `执行全局垃圾回收:
  1. 删除已不存在项目目录的 DB 记录（过期记录）
  2. 加 --cache 标志额外删除无任何项目引用的缓存目录（孤立缓存）

知识库（wiki）由 Git 管理，不会被 GC 删除。

使用 --dry-run 预览将被清理的内容，不实际执行。`,
		Run: func(cmd *cobra.Command, args []string) {
			db, err := utils.GetGormDB()
			if err != nil {
				fmt.Printf("数据库连接失败: %v\n", err)
				return
			}

			previewConfig := &logicglobal.GlobalGCConfig{DryRun: true, Cache: cache}
			previewProc := logicglobal.NewGlobalGCProcessor(previewConfig, common.AppConfigModel, db)
			preview, err := previewProc.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("扫描失败: %v\n", err)
				return
			}

			totalItems := len(preview.StaleDBRecords) + len(preview.OrphanedCaches)
			if totalItems == 0 {
				fmt.Println("  一切正常，无需清理。")
				return
			}

			fmt.Println("  [预览] 以下内容将被清理:")
			renderGCTable(preview)

			if dryRun {
				fmt.Printf("\n  共发现 %d 项可清理。移除 --dry-run 以执行清理。\n", totalItems)
				return
			}

			if !yes {
				fmt.Printf("\n  确认执行清理? [y/N]: ")
				var input string
				fmt.Scanln(&input)
				if strings.ToLower(input) != "y" {
					fmt.Println("  已取消")
					return
				}
			}

			execConfig := &logicglobal.GlobalGCConfig{DryRun: false, Cache: cache}
			execProc := logicglobal.NewGlobalGCProcessor(execConfig, common.AppConfigModel, db)
			result, err := execProc.Execute(cmd.Context())
			if err != nil {
				fmt.Printf("清理失败: %v\n", err)
				return
			}

			parts := []string{}
			if result.DBRecordsRemoved > 0 {
				parts = append(parts, fmt.Sprintf("%d 条 DB 记录", result.DBRecordsRemoved))
			}
			if result.CacheDirsRemoved > 0 {
				parts = append(parts, fmt.Sprintf("%d 个缓存目录", result.CacheDirsRemoved))
			}
			if len(parts) > 0 {
				fmt.Printf("\n  已清理: %s\n", strings.Join(parts, ", "))
			}
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "预览模式，不实际删除")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认提示")
	cmd.Flags().BoolVar(&cache, "cache", false, "同时清理无引用的缓存目录（默认仅清理 DB 记录）")
	return cmd
}

func renderGCTable(result *logicglobal.GlobalGCResult) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter, WidthMax: 8},
		{Number: 2, Align: text.AlignLeft},
	})
	t.AppendHeader(table.Row{"类型", "路径", "详情"})

	for _, r := range result.StaleDBRecords {
		t.AppendRow(table.Row{"DB记录", r.ProjectDir, fmt.Sprintf("%d 条引用记录", r.RepoCount)})
	}
	for _, c := range result.OrphanedCaches {
		t.AppendRow(table.Row{"孤立缓存", c.Path, ""})
	}

	fmt.Println(t.Render())
}
