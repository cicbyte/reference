package repo

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cicbyte/reference/internal/log"
	"github.com/cicbyte/reference/internal/models"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

// ResolveProxy 解析最终使用的代理地址
// 优先级：git_proxy > proxy > 环境变量 HTTPS_PROXY/HTTP_PROXY > 空字符串
func ResolveProxy(appConfig *models.AppConfig) string {
	if appConfig.Network.GitProxy != "" {
		return appConfig.Network.GitProxy
	}
	if appConfig.Network.Proxy != "" {
		return appConfig.Network.Proxy
	}
	for _, key := range []string{"HTTPS_PROXY", "HTTP_PROXY", "https_proxy", "http_proxy"} {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

// SetupGitProxy 根据代理 URL 创建并注册 go-git transport
// 返回 cleanup 函数，调用后恢复默认 transport
func SetupGitProxy(proxyURL string) (func(), error) {
	if proxyURL == "" {
		return func() {}, nil
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("解析代理地址失败: %w", err)
	}

	var httpClient *http.Client

	switch parsed.Scheme {
	case "socks5", "socks5h":
		httpClient, err = newSOCKS5Client(parsed)
	case "http", "https":
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(parsed),
			},
		}
	default:
		return nil, fmt.Errorf("不支持的代理协议: %s（支持 http/https/socks5）", parsed.Scheme)
	}

	if err != nil {
		return nil, fmt.Errorf("创建代理客户端失败: %w", err)
	}

	// 注册自定义 transport（全局生效）
	customClient := githttp.NewClient(httpClient)
	client.InstallProtocol("https", customClient)
	client.InstallProtocol("http", customClient)

	log.Info("代理已启用", zap.String("proxy", proxyURL))

	return func() {
		client.InstallProtocol("https", githttp.DefaultClient)
		client.InstallProtocol("http", githttp.DefaultClient)
	}, nil
}

func newSOCKS5Client(parsed *url.URL) (*http.Client, error) {
	addr := parsed.Host
	if !strings.Contains(addr, ":") {
		addr = addr + ":1080"
	}

	var auth *proxy.Auth
	if parsed.User != nil {
		password, _ := parsed.User.Password()
		auth = &proxy.Auth{
			User:     parsed.User.Username(),
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("创建 SOCKS5 dialer 失败: %w", err)
	}

	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, fmt.Errorf("SOCKS5 dialer 不支持 ContextDialer")
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext: contextDialer.DialContext,
		},
	}, nil
}
