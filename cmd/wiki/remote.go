package wiki

import (
	"fmt"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

func getRemoteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remote [url]",
		Short: "查看或设置知识库远程仓库",
		Long: `查看当前远程仓库地址，或设置新的远程仓库地址。

  reference wiki remote                        # 查看当前远程地址
  reference wiki remote https://github.com/...  # 设置远程地址`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			wikiDir := utils.ConfigInstance.GetWikiDir()
			if len(args) == 0 {
				url, err := logicwiki.GetRemoteURL(wikiDir)
				if err != nil {
					fmt.Printf("  获取远程地址失败: %v\n", err)
					return
				}
				if url == "" {
					fmt.Println("  远程仓库: (未设置)")
				} else {
					fmt.Printf("  远程仓库: %s\n", url)
				}
				return
			}
			if err := logicwiki.SetRemote(wikiDir, args[0]); err != nil {
				fmt.Printf("  设置远程仓库失败: %v\n", err)
				return
			}
			fmt.Printf("  远程仓库已设置为 %s\n", args[0])
			fmt.Println("  提示: 运行 reference wiki sync 同步远程知识库")
		},
	}
}
