package reference

import "time"

// RefType 引用类型
type RefType string

const (
	RefTypeRemote RefType = "remote"
	RefTypeLocal  RefType = "local"
)

// RepoInfo 仓库信息
type RepoInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProjectDir string  `json:"project_dir"`
	LinkName   string  `json:"link_name"`
	RefType    RefType `json:"ref_type"`

	// 远程仓库字段
	RemoteURL string `json:"remote_url,omitempty"`
	Host      string `json:"host,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	RepoName  string `json:"repo_name,omitempty"`
	CachePath string `json:"cache_path,omitempty"`

	// 本地仓库字段
	LocalPath string `json:"local_path,omitempty"`

	// Wiki 路径
	WikiSubPath string `json:"wiki_sub_path,omitempty"`

	// 引用链接名
	RefName string `json:"ref_name,omitempty"`

	// 元数据
	Branch   string     `json:"branch,omitempty"`
	Commit   string     `json:"commit,omitempty"`
	CommitAt *time.Time `json:"commit_at,omitempty"`
}

// AddRepoOptions 添加仓库选项
type AddRepoOptions struct {
	Target     string // Git URL 或本地路径
	Local      bool   // 是否为本地仓库
	Name       string // 自定义链接名称
	Branch     string // 指定分支或标签
	Update     bool   // 强制更新已有缓存
	ProjectDir string // 项目目录
}

// AddRepoResult 添加仓库结果
type AddRepoResult struct {
	RefName string
	RefType RefType
}

// RemoveRepoOptions 移除仓库选项
type RemoveRepoOptions struct {
	Identifier string // 仓库标识符
	Purge      bool   // 同时删除全局缓存
	Clean      bool   // 同时清除 AI 配置和 .reference/ 目录
	Yes        bool   // 跳过确认提示
	All        bool   // 移除全部引用
	ProjectDir string // 项目目录
}

// ListReposResult 列出仓库结果
type ListReposResult struct {
	Repos []RepoItem `json:"repos"`
}

// RepoItem 仓库列表项
type RepoItem struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	CachePath string `json:"cache_path"`
	CommitAt  string `json:"commit_at"`
	Branch    string `json:"branch"`
}

// UpdateReposOptions 更新仓库选项
type UpdateReposOptions struct {
	Identifier string // 仓库标识符，为空则更新全部
	ProjectDir string // 项目目录
}

// DoctorResult 诊断结果
type DoctorResult struct {
	Checks  []DoctorCheck `json:"checks"`
	Summary string        `json:"summary"`
}

// DoctorCheck 诊断检查项
type DoctorCheck struct {
	Group   string `json:"group"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details"`
}

// SCCResult 代码统计结果
type SCCResult struct {
	Repo     string          `json:"repo"`
	Path     string          `json:"path"`
	Langs    []SCCLangStat   `json:"languages"`
	TopFiles []SCCFileStat   `json:"topFiles"`
}

// SCCLangStat 语言统计
type SCCLangStat struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Count    int    `json:"count"`
	Code     int    `json:"code"`
	Complex  int    `json:"complexity"`
}

// SCCFileStat 文件统计
type SCCFileStat struct {
	Type     string `json:"type"`
	File     string `json:"file"`
	Language string `json:"language"`
	Code     int    `json:"code"`
	Complex  int    `json:"complexity"`
}
