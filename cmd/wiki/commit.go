package wiki

import (
	"fmt"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

func getCommitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "提交知识库更改",
		Run: func(cmd *cobra.Command, args []string) {
			wikiDir := utils.ConfigInstance.GetWikiDir()
			result, err := logicwiki.StageAndCommit(wikiDir, "")
			if err != nil {
				fmt.Printf("  提交失败: %v\n", err)
				return
			}
			if !result.HasChanges {
				fmt.Println("  没有需要提交的更改。")
				return
			}
			fmt.Printf("  已提交: %s (%s)\n", result.CommitHash, result.Message)
		},
	}
}
