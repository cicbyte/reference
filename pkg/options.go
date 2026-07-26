package reference

import (
	"io/fs"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Options Engine 配置选项
type Options struct {
	// AppDir 应用数据目录，默认 ~/.cicbyte/apps/reference
	AppDir string

	// DB 数据库连接，可选。为 nil 时内部自动创建 SQLite 连接
	DB *gorm.DB

	// PromptsFS 嵌入的提示词文件系统，可选。
	// 为 nil 时模板注入功能（Inject）不可用
	PromptsFS fs.FS

	// Logger 日志记录器，可选。为 nil 时使用 nop logger
	Logger *zap.Logger

	// ReposPath 覆盖默认的仓库缓存目录
	// 默认: AppDir/repos
	ReposPath string

	// WikiPath 覆盖默认的知识库目录
	// 默认: AppDir/wiki
	WikiPath string

	// LocalWikiPath 覆盖默认的本地知识库目录
	// 默认: AppDir/localwiki
	LocalWikiPath string

	// Proxy 代理设置，可选
	Proxy ProxyOptions
}

// ProxyOptions 代理配置
type ProxyOptions struct {
	// HTTP HTTP/HTTPS 代理地址
	HTTP string

	// Git Git 专用代理，为空则回退使用 HTTP
	Git string
}

// Validate 验证选项
func (o *Options) Validate() error {
	// 可以在这里添加验证逻辑
	return nil
}
