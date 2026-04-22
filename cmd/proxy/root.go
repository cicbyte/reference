package proxy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cicbyte/reference/internal/models"
	"github.com/cicbyte/reference/internal/utils"
	"github.com/spf13/cobra"
)

var presets = map[string]string{
	"clash":    "http://127.0.0.1:7890",
	"v2ray":    "http://127.0.0.1:1080",
	"ss":       "socks5://127.0.0.1:1080",
	"surge":    "http://127.0.0.1:6152",
	"qv2ray":   "http://127.0.0.1:8080",
	"ssrdog":   "http://127.0.0.1:7897",
}

type ProxyProcessor struct {
	appConfig *models.AppConfig
}

func NewProxyProcessor() *ProxyProcessor {
	return &ProxyProcessor{
		appConfig: utils.ConfigInstance.LoadConfig(),
	}
}

func (p *ProxyProcessor) Get() string {
	proxy := p.appConfig.Network.Proxy
	gitProxy := p.appConfig.Network.GitProxy
	timeout := p.appConfig.Network.Timeout

	var sb strings.Builder
	if proxy != "" {
		sb.WriteString(fmt.Sprintf("  代理: %s\n", proxy))
	} else {
		sb.WriteString("  代理: (未设置)\n")
	}
	if gitProxy != "" {
		sb.WriteString(fmt.Sprintf("  Git 代理: %s\n", gitProxy))
	} else {
		sb.WriteString("  Git 代理: (未设置，回退到通用代理)\n")
	}
	if timeout > 0 {
		sb.WriteString(fmt.Sprintf("  超时: %ds\n", timeout))
	}
	return sb.String()
}

func normalizeProxyURL(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// 纯数字 → 端口号，默认 http://127.0.0.1
	if _, err := strconv.Atoi(input); err == nil {
		return "http://127.0.0.1:" + input
	}

	// 无 scheme → 默认 http
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") &&
		!strings.HasPrefix(input, "socks5://") && !strings.HasPrefix(input, "socks5h://") {
		input = "http://" + input
	}

	return input
}

func (p *ProxyProcessor) Set(proxy string) string {
	p.appConfig.Network.Proxy = proxy
	if err := utils.ConfigInstance.SaveConfig(p.appConfig); err != nil {
		return fmt.Sprintf("  设置代理失败: %v", err)
	}
	return fmt.Sprintf("  代理已设置为 %s", proxy)
}

func (p *ProxyProcessor) SetGit(gitProxy string) string {
	p.appConfig.Network.GitProxy = gitProxy
	if err := utils.ConfigInstance.SaveConfig(p.appConfig); err != nil {
		return fmt.Sprintf("  设置 Git 代理失败: %v", err)
	}
	return fmt.Sprintf("  Git 代理已设置为 %s", gitProxy)
}

func (p *ProxyProcessor) Clear() string {
	p.appConfig.Network.Proxy = ""
	p.appConfig.Network.GitProxy = ""
	if err := utils.ConfigInstance.SaveConfig(p.appConfig); err != nil {
		return fmt.Sprintf("  清除代理失败: %v", err)
	}
	return "  代理配置已清除"
}

func GetProxyCommand() *cobra.Command {
	var setGit bool
	var preset string

	rootCmd := &cobra.Command{
		Use:   "proxy",
		Short: "查看和管理代理配置",
		Long: `查看、设置或清除网络代理配置。

快捷设置：
  reference proxy set 7890           # 端口号，默认 http://127.0.0.1
  reference proxy set --preset clash # 预设：clash(7890) v2ray(1080) ss(1080/socks5) surge(6152) qv2ray(8080) ssrdog(7897)
  reference proxy set --git 1080     # 设置 Git 专用代理`,
		Run: func(cmd *cobra.Command, args []string) {
			processor := NewProxyProcessor()
			fmt.Print(processor.Get())
		},
	}

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "查看当前代理配置",
		Run: func(cmd *cobra.Command, args []string) {
			processor := NewProxyProcessor()
			fmt.Print(processor.Get())
		},
	}

	setCmd := &cobra.Command{
		Use:   "set <url|port>",
		Short: "设置代理",
		Long: `设置代理地址。支持完整 URL 或纯端口号。

纯端口号自动补全为 http://127.0.0.1:<port>
无 scheme 的地址自动补全 http://

预设（--preset）：
  clash   → http://127.0.0.1:7890
  v2ray   → http://127.0.0.1:1080
  ss      → socks5://127.0.0.1:1080
  surge   → http://127.0.0.1:6152
  qv2ray  → http://127.0.0.1:8080
  ssrdog  → http://127.0.0.1:7897`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			processor := NewProxyProcessor()

			if preset != "" {
				args = []string{presets[preset]}
			}

			if setGit {
				if len(args) == 0 {
					fmt.Println("  错误：请指定代理地址或端口号")
					return
				}
				url := normalizeProxyURL(args[0])
				fmt.Println(processor.SetGit(url))
			} else {
				if len(args) == 0 {
					fmt.Println("  错误：请指定代理地址、端口号或 --preset")
					return
				}
				url := normalizeProxyURL(args[0])
				fmt.Println(processor.Set(url))
			}
		},
	}

	setCmd.Flags().BoolVar(&setGit, "git", false, "设置为 Git 专用代理")
	setCmd.Flags().StringVar(&preset, "preset", "", "使用预设 (clash|v2ray|ss|surge|qv2ray)")

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "清除代理配置",
		Run: func(cmd *cobra.Command, args []string) {
			processor := NewProxyProcessor()
			fmt.Println(processor.Clear())
		},
	}

	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(clearCmd)
	return rootCmd
}
