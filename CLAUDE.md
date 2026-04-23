# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

reference 是一个面向 AI 辅助编程时代的本地代码仓库引用管理器。通过全局缓存和项目级 Junction 链接，让开发者及 AI 助手零延迟查阅任意远程或本地 Git 仓库代码。核心价值：零延迟、知识可复用、离线可用、不污染 AI 上下文。

## 构建命令

```bash
go build -o reference.exe   # Windows（开发用，版本号为 dev）
go build -o reference        # Linux/macOS
go mod tidy                  # 整理依赖
python scripts/build.py --local   # 交叉编译 + UPX 压缩（注入版本号，输出到 dist/）
python scripts/build.py           # 三平台全量编译
```

## 命令行用法

```bash
reference                              # 无参数：首次运行交互式引导，后续自动注入 + 修复
reference version                      # 显示版本信息
reference repo add <url>               # 添加远程仓库（支持 owner/repo 简写）
reference repo add --local <path>      # 添加本地仓库
reference repo remove <name>           # 移除引用（按名称）
reference repo remove --all            # 移除当前项目全部引用
reference repo remove <name> --purge   # 同时删除缓存（需确认）
reference repo list [-f json|jsonl]    # 列出所有引用
reference repo update [name]           # 更新远程仓库
reference repo scc [name] [-n 15] [-f json|jsonl]  # 查看代码统计
reference doctor [-f json|jsonl]       # 诊断并修复引用健康状态
reference global list [-f json|jsonl] # 列出所有项目及其引用关系
reference global gc [--dry-run] [-y] [--cache]  # 清理过期 DB 记录（--cache 额外清理孤立缓存）
reference global stats [-f json|jsonl] # 显示全局统计信息
reference proxy set <url|port>        # 设置代理
reference proxy info                  # 查看代理
reference proxy clear                 # 清除代理
reference wiki                         # 查看 wiki 状态
reference wiki commit                  # 提交知识库更改
reference wiki sync                   # 同步知识库（pull + commit + push）
reference wiki remote [url]           # 查看/设置远程仓库
reference wiki trash                   # 查看被删除的文件
reference wiki restore <path>         # 从 Git 历史恢复文件
```

详细用法见 `docs/` 目录。

## 架构

### CMD/Logic 分层（cobra-app 规范）

- **cmd/root.go** — 唯一根命令，注册 global/repo/proxy/wiki/version 模块，无参数时首次运行交互式引导、后续自动注入
- **cmd/version/** — 版本变量包（供 ldflags 注入 Version/GitCommit/BuildTime）
- **cmd/global/** — global 模块 CMD 层（list、gc、stats，跨项目全局管理）
- **cmd/repo/** — repo 模块 CMD 层（参数绑定、验证、调用 Logic）
- **cmd/proxy/** — proxy 模块 CMD 层
- **cmd/wiki/** — wiki 模块 CMD 层（commit、sync、remote、watch、trash、restore）
- **internal/logic/repo/** — repo 模块 Logic 层（纯业务逻辑，不依赖 cobra），含 global_* 全局查询扩展
- **internal/logic/global/** — global 模块 Logic 层（全局管理：跨项目列表、GC、统计）

两层通过 `*Config` 结构体交互，Logic 层通过工厂方法创建 Processor，入口为 `Execute(ctx)`。

### 核心模块（internal/logic/repo/）

| 文件 | 职责 |
|:---|:---|
| url.go | Git URL 解析，支持 https/ssh/简写格式，主流平台自动识别 |
| cache.go | go-git/v5 克隆/拉取/元数据提取/本地仓库校验 |
| linker.go | 跨平台链接（Unix Symlink / Windows PowerShell Junction） |
| indexer.go | SQLite 仓库索引 CRUD（GORM） |
| gitignore.go | 确保 .reference/ 在 .gitignore 中 |
| proxy.go | 代理解析（git_proxy > proxy > 无代理） |
| inject.go | 生成 reference.map.jsonl + 创建 Wiki Junction + 静默修复软链接 + 注入 Agent 文件（仅 Claude Code 用户） |
| scc.go | 内置 scc 库封装（代码统计 + Top 文件排名），无外部依赖 |
| doctor.go | 诊断检查：软链接、Wiki Junction、Reference Map、数据库一致性、Wiki Git、Agent 配置 |

### Wiki 模块（internal/logic/wiki/）

| 文件 | 职责 |
|:---|:---|
| init.go | EnsureGitInit、IsGitInitialized、HasRemote、EnsureAutoPull |
| commit.go | StageAndCommit、Push（本地 go-git，远程 exec.Command） |
| remote.go | SetRemote、GetRemoteURL |
| sync.go | Sync（pull + auto-commit + push） |
| watcher.go | fsnotify 文件监听 + 防抖 + 自动提交 |
| daemon.go | 后台守护进程管理（PID 文件、启动、停止） |
| trash.go | 查看被删除文件、从 Git 历史恢复 |

### 数据模型（internal/models/）

- **AppConfig** — 配置模型（ReposPath、WikiPath、Network、Log）
- **Repo** — 仓库索引模型（ProjectDir+LinkName 唯一索引；RefName 文件系统短名；WikiSubPath wiki 嵌套路径）
- **ProjectSettings** — 项目级设置（Agent 选择、Initialized 标志），存储于 `.reference/reference.settings.json`

### 项目目录（.reference/）

所有引用数据统一在 `.reference/` 下管理：
- `repos/<refName>` — 仓库源码 Junction（→ 全局缓存）
- `wiki/<refName>` — 知识库 Junction（→ 全局 wiki）
- `reference.map.jsonl` — AI 导航数据（仓库列表，供 Agent 读取）
- `reference.settings.json` — 项目配置（agent 选择、initialized 标志）

`.claude/` 下仅存放 AI 配置（仅 Claude Code 用户）：
- `agents/reference-explorer.md`、`agents/reference-analyzer.md` — 子代理提示词
- `skills/reference/SKILL.md` — Skill 定义

### 用户数据目录

`~/.cicbyte/reference/` 下：
- `config/config.yaml` — 配置文件
- `db/app.db` — SQLite 数据库（纯 Go，无 CGO）
- `repos/` — 全局缓存（可通过 config.yaml 中 repos_path 覆盖）
- `wiki/` — 全局知识库（Git 版本控制，嵌套目录：`<platform>/<namespace>/<repo>/`）
- `logs/` — 日志文件

### AI Agent 配置体系

- `prompts/skills/reference/SKILL.md` — Skill 模板（通过读取 `reference.map.jsonl` 发现仓库）
- `prompts/agents/reference-explorer.md` — 知识探索子代理提示词
- `prompts/agents/reference-analyzer.md` — 深度分析子代理提示词
- `reference` 命令（无参数）和 `doctor` 命令都会将上述文件注入到 `.claude/`（仅 Claude Code 用户）
- 知识文件写入 `~/.cicbyte/reference/wiki/<platform>/<namespace>/<repo>/`，通过 Junction 链接到 `.reference/wiki/<refName>/`

### 应用初始化流程（cmd/root.go init()）

严格按顺序：`InitAppDirs` → `LoadConfig` → `ApplyConfig` → `InitLog` → `GetGormDB`（含 AutoMigrate） → `EnsureGitInit` → `EnsureAutoPull`，任何步骤失败 `os.Exit(1)`。

无参数运行 `reference` 时，若 `reference.settings.json` 不存在或 `initialized == false`，进入交互式引导（选择编程助手），保存后继续注入流程。

## 版本发布

```bash
python scripts/gen_release_notes.py   # 从上次 tag 到 HEAD 生成 msg.txt
python scripts/release.py             # 升级版本号 + 提交 + 打 tag → 触发 GitHub Actions
# 或手动指定版本：python scripts/release.py 0.2.0
```

- VERSION 文件为唯一版本号来源，CI 和本地构建均从该文件读取
- GitHub Actions 监听 `v*` tag 推送，交叉编译 Windows/Linux/macOS 并创建 Release
- 版本信息通过 `-ldflags -X` 注入 `cmd/version` 包

## 开发注意事项

- Go 1.24+（go-git/v5 要求），当前使用 1.25.2
- 新增命令在 `cmd/repo/` 下创建文件，在 `root.go` 中 `AddCommand`
- Windows Junction 通过 PowerShell `New-Item -ItemType Junction` 创建，不需要管理员权限
- 全局状态通过 `common.AppConfigModel` 桥接
- 功能列表见 `features.json`，进度见 `claude-progress.txt`
- 文档目录 `docs/` 包含每个命令的详细用法
