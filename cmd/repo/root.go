package repo

import (
	"github.com/spf13/cobra"
)

func GetRepoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "管理代码仓库引用",
		Long: `管理项目级 Git 仓库引用，支持远程和本地仓库。

支持的操作:
  - 添加远程/本地仓库引用
  - 移除引用
  - 列出所有引用
  - 更新远程仓库缓存`,
	}
	cmd.AddCommand(getAddCommand())
	cmd.AddCommand(getRemoveCommand())
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getUpdateCommand())
	cmd.AddCommand(getSccCommand())
	return cmd
}
