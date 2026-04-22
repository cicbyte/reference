package models

import "time"

type RefType string

const (
	RefTypeRemote RefType = "remote"
	RefTypeLocal  RefType = "local"
)

type Repo struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProjectDir string  `gorm:"uniqueIndex:idx_project_link;not null" json:"project_dir"` // 所属项目目录（绝对路径）
	LinkName   string  `gorm:"uniqueIndex:idx_project_link;not null" json:"link_name"`
	RefType    RefType `gorm:"index;not null" json:"ref_type"` // remote / local

	// 远程仓库字段
	RemoteURL  string `json:"remote_url,omitempty"`
	Host       string `json:"host,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	RepoName   string `json:"repo_name,omitempty"`
	CachePath  string `json:"cache_path,omitempty"`

	// 本地仓库字段
	LocalPath string `json:"local_path,omitempty"`

	// Wiki 路径
	WikiSubPath string `json:"wiki_sub_path,omitempty"` // wiki 嵌套子路径，如 github/cicbyte/repo 或 local/project-name

	// 引用链接名（文件系统显示名，默认短名，重名时自动加前缀）
	RefName string `json:"ref_name,omitempty"`

	// 元数据
	Branch    string `json:"branch,omitempty"`
	Commit    string `json:"commit,omitempty"`
	CommitAt  *time.Time `json:"commit_at,omitempty"`
}
