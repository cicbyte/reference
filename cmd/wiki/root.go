package wiki

import (
	"fmt"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

func GetWikiCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wiki",
		Short: "管理全局知识库的 Git 同步",
		Run: func(cmd *cobra.Command, args []string) {
			wikiDir := utils.ConfigInstance.GetWikiDir()
			fmt.Print(logicwiki.GetWikiStatus(wikiDir))
		},
	}
	cmd.AddCommand(getSyncCommand())
	cmd.AddCommand(getRemoteCommand())
	cmd.AddCommand(getCommitCommand())
	cmd.AddCommand(getTrashCommand())
	cmd.AddCommand(getRestoreCommand())
	return cmd
}
