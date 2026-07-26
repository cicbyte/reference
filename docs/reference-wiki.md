# reference wiki — 知识库管理

管理全局知识库的 Git 版本控制和文件恢复。

## 概述

知识库分为两个独立的 Git 仓库：

- **公共知识库** `~/.cicbyte/apps/reference/wiki/` — 远程仓库的知识文件，可推送到公共 Git 仓库
- **本地知识库** `~/.cicbyte/apps/reference/localwiki/` — 本地仓库的知识文件，独立管理（如推送到 Gitea）

两者均自动初始化为 Git 仓库，提供版本控制、自动提交和文件恢复能力。

- 首次启动 `reference` 时自动 `git init`
- 如果配置了远程仓库，启动时自动 `git pull`
- 误删文件可随时从 Git 历史恢复

## 全局标志

| 标志 | 简写 | 说明 |
|:---|:---|:---|
| `--local` | `-l` | 操作本地知识库（localwiki）而非公共知识库 |

## 目录结构

```
~/.cicbyte/apps/reference/
├── wiki/                              # 公共知识库（远程仓库）
│   ├── .git/
│   ├── github/
│   │   ├── cicbyte/memos-cli/
│   │   │   ├── reference.md
│   │   │   └── <主题>.md
│   │   └── boyter/scc/
│   └── gitlab.com/group/project/
└── localwiki/                         # 本地知识库（本地仓库）
    ├── .git/
    ├── my-project/
    │   ├── reference.md
    │   └── <主题>.md
    └── another-lib/
```

项目中通过 Junction 链接统一访问：

```
.reference/
├── repos/<refName>    → 全局缓存
└── wiki/<refName>     → wiki/ 或 localwiki/ 中对应目录
```

## 命令

### `reference wiki`

查看 wiki 状态（Git 初始化状态、远程仓库、工作区状态）。

```bash
reference wiki              # 查看公共知识库状态
reference wiki --local      # 查看本地知识库状态
```

### `reference wiki commit`

提交当前所有更改到 Git 仓库。

```bash
reference wiki commit              # 提交公共知识库
reference wiki --local commit      # 提交本地知识库
```

### `reference wiki sync`

完整同步：pull → 自动提交 → push。

```bash
reference wiki sync                # 同步公共知识库
reference wiki --local sync        # 同步本地知识库
```

### `reference wiki remote [url]`

查看或设置远程仓库。

```bash
reference wiki remote                                    # 查看公共知识库远程
reference wiki remote https://github.com/user/wiki.git   # 设置公共知识库远程

reference wiki --local remote                            # 查看本地知识库远程
reference wiki --local remote https://gitea.example.com/user/localwiki.git  # 设置本地知识库远程
```

### `reference wiki trash`

查看被删除的知识文件。

```bash
reference wiki trash                # 公共知识库最近 20 条
reference wiki --local trash        # 本地知识库最近 20 条
reference wiki trash -n 50          # 最近 50 条
```

输出示例：
```
  [a894455 2026-04-22] github/boyter/scc/scc.md
  [a894455 2026-04-22] test_watcher.md
```

### `reference wiki restore <path>`

从 Git 历史恢复被删除的文件。

```bash
reference wiki restore github/boyter/scc/scc.md
reference wiki --local restore my-project/api.md
```

恢复后的文件会保留在 wiki 目录中，下次 commit 时自动提交。

## 典型用法

### 公共知识库（默认）

适用于远程仓库，可推送到 GitHub 等公共平台共享：

```bash
reference wiki remote https://github.com/user/reference-wiki.git
reference wiki sync
```

### 本地知识库

适用于本地/私有仓库，推送到 Gitea 等私有平台：

```bash
reference wiki --local remote https://gitea.example.com/user/localwiki.git
reference wiki --local sync
```
