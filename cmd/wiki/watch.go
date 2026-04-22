package wiki

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	logicwiki "github.com/cicbyte/reference/internal/logic/wiki"
	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func getWatchCommand() *cobra.Command {
	var daemon, stop, autoPush, foreground bool

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "监听 wiki 目录变化并自动提交",
		Long: `监听 wiki 目录的文件变化，自动 stage + commit 到 Git 仓库。

运行模式：
  reference wiki watch              前台运行（阻塞终端）
  reference wiki watch --daemon     后台守护进程
  reference wiki watch --stop       停止后台守护进程`,
		Run: func(cmd *cobra.Command, args []string) {
			wikiDir := utils.ConfigInstance.GetWikiDir()

			if stop {
				if err := logicwiki.StopDaemon(wikiDir); err != nil {
					fmt.Printf("停止失败: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("watcher 已停止")
				return
			}

			if daemon {
				if err := logicwiki.Daemonize(wikiDir); err != nil {
					fmt.Printf("启动失败: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("watcher 已在后台启动")
				return
			}

			if !foreground && logicwiki.IsDaemonRunning(wikiDir) {
				fmt.Println("watcher 已在运行中，如需重启请先执行 reference wiki watch --stop")
				return
			}

			runForeground(wikiDir, autoPush)
		},
	}

	cmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "以守护进程方式后台运行")
	cmd.Flags().BoolVarP(&stop, "stop", "s", false, "停止后台守护进程")
	cmd.Flags().BoolVar(&autoPush, "push", false, "自动推送（commit 后自动 push）")
	cmd.Flags().BoolVar(&foreground, "foreground", false, "内部使用：以前台模式运行")
	_ = cmd.Flags().MarkHidden("foreground")
	return cmd
}

func runForeground(wikiDir string, autoPush bool) {
	watcher, err := logicwiki.NewWatcher(wikiDir, autoPush)
	if err != nil {
		log.Error("创建 watcher 失败", zap.Error(err))
		os.Exit(1)
	}

	if err := logicwiki.WritePID(wikiDir); err != nil {
		log.Warn("写入 PID 文件失败", zap.Error(err))
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go watcher.Run()

	<-sigCh
	watcher.Stop()
	logicwiki.RemovePID(wikiDir)
}
