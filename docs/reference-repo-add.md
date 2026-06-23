# reference repo add

添加一个远程 Git 仓库或本地 Git 仓库到当前项目的引用中。

## 用法

```bash
reference repo add <target> [flags]
```

## 参数

| 参数 | 说明 |
|:---|:---|
| `<target>` | 远程仓库 URL 或本地仓库路径 |

## 标志

| 标志 | 简写 | 说明 |
|:---|:---|:---|
| `--local` | `-l` | 标记为本地 Git 仓库路径 |
| `--name` | `-n` | 自定义链接名称（默认从 URL/路径自动生成） |
| `--branch` | `-b` | 指定克隆的分支或标签（仅远程模式） |
| `--update` | `-u` | 若缓存已存在，强制 git pull 更新（仅远程模式） |

## 远程仓库

支持多种 URL 格式，主流平台自动识别：

```bash
# 完整 URL
reference repo add https://github.com/gin-gonic/gin
reference repo add https://gitlab.com/group/project.git

# owner/repo 简写（自动补全为 GitHub）
reference repo add spf13/cobra
reference repo add golang/go

# 指定分支
reference repo add golang/go --branch master

# 自定义链接名称
reference repo add spf13/cobra --name cobra

# 强制更新已有缓存
reference repo add spf13/cobra --update
```

### 远程模式执行流程

1. 解析 URL，识别平台和命名空间
2. **完整克隆**（非浅克隆，保证 commit 历史完整）到全局缓存 `~/.cicbyte/reference/repos/<platform>/<namespace>/<repo>`
3. 在 `.reference/` 下创建 Junction/Symlink 链接
4. 确保 `.reference/` 在 `.gitignore` 中
5. 写入数据库索引
6. 自动调用 inject 生成 AI Agent 配置和知识文件

> **关于浅克隆**：早期版本使用 `--depth 1` 浅克隆以节省空间，但实践中发现浅克隆会导致更新时 commit 不匹配等问题，现已移除该机制，统一使用完整克隆。对历史遗留的浅克隆缓存，`repo update` 会自动执行 `git fetch --unshallow` 转为完整仓库。

## 本地仓库

```bash
# 添加本地仓库
reference repo add --local ~/projects/my-lib

# 自定义链接名称
reference repo add --local ~/projects/my-lib --name my-utils
```

### 本地模式特点

- 不执行克隆，直接使用本地路径
- 链接名称默认取目录名
- 不会在 `repo update` 时被更新
- 知识文件存储在独立的 `localwiki/` 目录（与公共 wiki 隔离），可通过 `reference wiki --local` 管理

## 链接命名规则

默认从 URL 或路径自动生成，格式为 `<host>-<namespace>-<repo>`：

| 输入 | 链接名称 |
|:---|:---|
| `https://github.com/gin-gonic/gin` | `github.com-gin-gonic-gin` |
| `spf13/cobra` | `github.com-spf13-cobra` |
| `~/projects/my-lib` | `my-lib` |

## 相关命令

- `reference repo list` — 列出所有引用
- `reference repo remove` — 移除引用
- `reference repo update` — 更新远程缓存
