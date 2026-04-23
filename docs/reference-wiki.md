# reference wiki — 知识库管理

管理全局知识库的 Git 版本控制和文件恢复。

## 概述

知识库目录 `~/.cicbyte/reference/wiki/` 自动初始化为 Git 仓库，提供版本控制、自动提交和文件恢复能力。

- 首次启动 `reference` 时自动 `git init`
- 如果配置了远程仓库，启动时自动 `git pull`
- 子代理写入知识文件后可通过 watcher 自动提交
- 误删文件可随时从 Git 历史恢复

## 目录结构

```
wiki/
├── .git/                          # Git 仓库（自动初始化）
├── .gitignore                     # Git 忽略规则
├── github/                        # 远程仓库知识
│   ├── cicbyte/memos-cli/
│   │   ├── reference.md
│   │   └── <主题>.md
│   └── boyter/scc/
└── local/                         # 本地仓库知识
    └── my-project/
```

## 命令

### `reference wiki`

查看 wiki 状态（Git 初始化状态、远程仓库、工作区状态）。

```bash
reference wiki
```

### `reference wiki commit`

提交当前所有更改到 Git 仓库。

```bash
reference wiki commit
```

### `reference wiki sync`

完整同步：pull → 自动提交 → push。

```bash
reference wiki sync
```

### `reference wiki remote [url]`

查看或设置远程仓库。

```bash
reference wiki remote                    # 查看当前远程
reference wiki remote https://github.com/user/wiki.git  # 设置远程
```

### `reference wiki trash`

查看被删除的知识文件。

```bash
reference wiki trash            # 最近 20 条
reference wiki trash -n 50      # 最近 50 条
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
```

恢复后的文件会保留在 wiki 目录中，下次 commit 时自动提交。
