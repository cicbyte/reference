package wiki

import (
	"fmt"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/spf13/cobra"
)

func getRestoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restore <path>",
		Short: "从 Git 历史恢复被删除的文件",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := logicwiki.RestoreFile(getWikiDir(), args[0]); err != nil {
				fmt.Printf("恢复失败: %v\n", err)
				return
			}
			fmt.Printf("已恢复: %s\n", args[0])
		},
	}
}
