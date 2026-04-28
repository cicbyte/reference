package models

type ConfigState struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

type AppConfig struct {
	ReposPath string `yaml:"repos_path,omitempty"` // 全局缓存目录，默认 ~/.cicbyte/reference/repos
	WikiPath  string `yaml:"wiki_path,omitempty"`  // 全局知识库目录，默认 ~/.cicbyte/reference/wiki

	Network struct {
		Proxy    string `yaml:"proxy"`     // HTTP/HTTPS 代理地址
		GitProxy string `yaml:"git_proxy"` // Git 专用代理，为空则回退使用 proxy
		Timeout  int    `yaml:"timeout"`   // 克隆/拉取超时时间（秒）
	} `yaml:"network"`

	Log struct {
		Level      string `yaml:"level"`
		MaxSize    int    `yaml:"maxSize"`
		MaxBackups int    `yaml:"maxBackups"`
		MaxAge     int    `yaml:"maxAge"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
}
