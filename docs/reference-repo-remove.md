# reference repo remove

移除当前项目中的某个仓库引用。

## 用法

```bash
reference repo remove <identifier> [flags]
```

## 参数

| 参数 | 说明 |
|:---|:---|
| `<identifier>` | 链接名称（如 `github.com-gin-gonic-gin`）或数据库 ID |

## 标志

| 标志 | 简写 | 说明 |
|:---|:---|:---|
| `--all` | | 移除当前项目全部引用 |
| `--purge` | `-p` | 同时删除全局缓存仓库（仅远程模式，需二次确认） |
| `--clean` | | 同时清除注入的 AI 配置和 `.reference/` 目录（需配合 `--all`） |
| `--yes` | `-y` | 跳过确认提示 |

## 示例

```bash
# 仅移除项目引用（软链接）
reference repo remove github.com-gin-gonic-gin

# 同时删除全局缓存
reference repo remove github.com-gin-gonic-gin --purge

# 移除当前项目全部引用
reference repo remove --all

# 移除全部引用并清除所有注入文件（AI 配置 + .reference/ 目录）
reference repo remove --all --clean
```

## 行为说明

**默认模式**（不加 `--purge`）：
- 删除 `.reference/<链接名>` 软链接
- 删除数据库索引记录
- 全局缓存 `~/.cicbyte/reference/repos/...` 保留，其他项目仍可使用

**清除模式**（`--purge`）：
- 在默认模式基础上，额外删除全局缓存目录
- 仅对远程仓库有效
- 对本地仓库的 `--purge` 会被忽略并提示（保护用户本地代码）
- 需要二次确认（交互式 y/n）

**完整清理**（`--all --clean`）：
- 移除当前项目所有引用（同 `--all`）
- 删除 `.claude/agents/reference-explorer.md`、`.claude/agents/reference-analyzer.md`
- 删除 `.claude/skills/reference/SKILL.md`
- 删除整个 `.reference/` 目录
- 效果等同于项目从未初始化过 reference，下次运行 `reference` 会重新引导

## 相关命令

- `reference repo list` — 查看链接名称
- `reference repo add` — 重新添加引用
