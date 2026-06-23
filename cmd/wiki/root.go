package wiki

import (
	"fmt"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

var localWiki bool

func getWikiDir() string {
	if localWiki {
		return utils.ConfigInstance.GetLocalWikiDir()
	}
	return utils.ConfigInstance.GetWikiDir()
}

func GetWikiCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wiki",
		Short: "管理全局知识库的 Git 同步",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(logicwiki.GetWikiStatus(getWikiDir()))
		},
	}
	cmd.PersistentFlags().BoolVarP(&localWiki, "local", "l", false, "操作本地知识库（localwiki）而非公共知识库")
	cmd.AddCommand(getSyncCommand())
	cmd.AddCommand(getRemoteCommand())
	cmd.AddCommand(getCommitCommand())
	cmd.AddCommand(getTrashCommand())
	cmd.AddCommand(getRestoreCommand())
	return cmd
}
