package global

import "github.com/spf13/cobra"

func GetGlobalCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global",
		Short: "全局引用管理和维护",
		Long: `管理全局引用状态，包括跨项目视图、垃圾回收和统计信息。

支持的操作:
  - list: 查看所有项目和引用的全局视图
  - gc: 清理过期 DB 记录和孤立缓存
  - stats: 查看全局统计信息`,
	}
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getGCCommand())
	cmd.AddCommand(getStatsCommand())
	cmd.AddCommand(getRemoveCommand())
	return cmd
}
