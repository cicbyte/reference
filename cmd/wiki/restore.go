package wiki

import (
	"fmt"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

func getRestoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restore <path>",
		Short: "从 Git 历史恢复被删除的文件",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			wikiDir := utils.ConfigInstance.GetWikiDir()
			if err := logicwiki.RestoreFile(wikiDir, args[0]); err != nil {
				fmt.Printf("恢复失败: %v\n", err)
				return
			}
			fmt.Printf("已恢复: %s\n", args[0])
		},
	}
}
