package utils



import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/cicbyte/reference/internal/models"
	"go.yaml.in/yaml/v3"
)

var ConfigInstance = Config{}

type Config struct {
	HomeDir      string
	AppSeriesDir string
	AppDir       string
	ConfigDir    string
	ConfigPath   string
	DbDir        string
	DbPath       string
	LogDir       string
	LogPath      string
	ReposDir     string
	WikiDir      string
}

func (c *Config) GetHomeDir() string {
	if c.HomeDir != "" {
		return c.HomeDir
	}
	usr, err := user.Current()
	if err != nil {
		c.HomeDir = os.Getenv("HOME")
		if c.HomeDir == "" {
			c.HomeDir = os.Getenv("USERPROFILE")
		}
		return c.HomeDir
	}
	c.HomeDir = usr.HomeDir
	return c.HomeDir
}

func (c *Config) GetAppSeriesDir() string {
	if c.AppSeriesDir != "" {
		return c.AppSeriesDir
	}
	c.AppSeriesDir = filepath.Join(c.GetHomeDir(), ".cicbyte")
	return c.AppSeriesDir
}

func (c *Config) GetAppDir() string {
	if c.AppDir != "" {
		return c.AppDir
	}
	c.AppDir = filepath.Join(c.GetAppSeriesDir(), "reference")
	return c.AppDir
}

func (c *Config) GetConfigDir() string {
	if c.ConfigDir != "" {
		return c.ConfigDir
	}
	c.ConfigDir = filepath.Join(c.GetAppDir(), "config")
	return c.ConfigDir
}
func (c *Config) GetConfigPath() string {
	if c.ConfigPath != "" {
		return c.ConfigPath
	}
	c.ConfigPath = filepath.Join(c.GetConfigDir(), "config.yaml")
	return c.ConfigPath
}

func (c *Config) GetDbDir() string {
	if c.DbDir != "" {
		return c.DbDir
	}
	dbDir := filepath.Join(c.GetAppDir(), "db")
	c.DbDir = dbDir
	return c.DbDir
}

func (c *Config) GetDbPath() string {
	if c.DbPath != "" {
		return c.DbPath
	}
	c.DbPath = filepath.Join(c.GetDbDir(), "app.db")
	return c.DbPath
}

func (c *Config) GetLogDir() string {
	if c.LogDir == "" {
		c.LogDir = filepath.Join(c.GetAppDir(), "logs")
	}
	return c.LogDir
}

func (c *Config) GetLogPath() string {
	if c.LogPath == "" {
		now := time.Now().Format("20060102")
		c.LogPath = filepath.Join(c.GetLogDir(), fmt.Sprintf("reference_log_%s.log", now))
	}
	return c.LogPath
}

func (c *Config) GetReposDir() string {
	if c.ReposDir != "" {
		return c.ReposDir
	}
	c.ReposDir = filepath.Join(c.GetAppDir(), "repos")
	return c.ReposDir
}

func (c *Config) GetWikiDir() string {
	if c.WikiDir != "" {
		return c.WikiDir
	}
	c.WikiDir = filepath.Join(c.GetAppDir(), "wiki")
	return c.WikiDir
}

func (c *Config) ApplyConfig(appConfig *models.AppConfig) {
	if appConfig.ReposPath != "" {
		c.ReposDir = appConfig.ReposPath
	}
	if appConfig.WikiPath != "" {
		c.WikiDir = appConfig.WikiPath
	}
}

func (c *Config) GetReposDirFromConfig(appConfig *models.AppConfig) string {
	if appConfig.ReposPath != "" {
		return appConfig.ReposPath
	}
	return c.GetReposDir()
}

func (c *Config) LoadConfig() *models.AppConfig {
	config_path := c.GetConfigPath()

	// 检查配置文件是否存在
	if _, err := os.Stat(config_path); os.IsNotExist(err) {
		defaultConfig := GetDefaultConfig()
		// 将默认配置写入文件
		header := "# reference 配置文件\n# repos_path: 全局仓库缓存目录，默认 ~/.cicbyte/reference/repos\n# wiki_path: 全局知识库目录，默认 ~/.cicbyte/reference/wiki\n\n"
		data, err := yaml.Marshal(defaultConfig)
		if err == nil {
			_ = os.WriteFile(config_path, []byte(header+string(data)), 0644)
		}
		return defaultConfig
	}

	// 读取配置文件内容
	data, err := os.ReadFile(config_path)
	if err != nil {
		return GetDefaultConfig()
	}

	// 解析YAML配置
	var config models.AppConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return GetDefaultConfig()
	}

	return &config
}

func (c *Config) SaveConfig(config *models.AppConfig) error {
	config_path := c.GetConfigPath()
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(config_path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

// 默认配置
func GetDefaultConfig() *models.AppConfig {
	config := &models.AppConfig{}

	// 日志默认配置
	config.Log.Level = "info"
	config.Log.MaxSize = 10
	config.Log.MaxBackups = 30
	config.Log.MaxAge = 30
	config.Log.Compress = true

	// 网络默认配置
	config.Network.Timeout = 300

	return config
}
