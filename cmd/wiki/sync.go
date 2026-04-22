package wiki

import (
	"fmt"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

func getSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "同步知识库（pull + 自动提交 + push）",
		Run: func(cmd *cobra.Command, args []string) {
			wikiDir := utils.ConfigInstance.GetWikiDir()
			result, err := logicwiki.Sync(wikiDir)
			if err != nil {
				fmt.Printf("  同步失败: %v\n", err)
				return
			}
			if result.PullOK {
				fmt.Println("  拉取: 成功")
			} else if result.PullErr != "" {
				fmt.Printf("  拉取: 失败 (%s)\n", result.PullErr)
			}
			if result.Commit != nil {
				if result.Commit.HasChanges {
					fmt.Printf("  提交: %s (%s)\n", result.Commit.CommitHash, result.Commit.Message)
				} else {
					fmt.Println("  提交: 无变更")
				}
			}
			if result.PushOK {
				fmt.Println("  推送: 成功")
			} else if result.PushErr != "" {
				fmt.Printf("  推送: 失败 (%s)\n", result.PushErr)
			}
		},
	}
}
