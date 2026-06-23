# reference init

非交互式初始化项目配置，适用于 CI/CD 集成、自动化脚本等场景。

## 用法

```bash
reference init [--agent <types>]
```

## 标志

| 标志 | 说明 |
|:---|:---|
| `--agent <types>` | 编程助手类型，多个用逗号分隔，默认 `none` |

## 支持的助手类型

| 值 | 助手 | 配置目录 |
|:---|:---|:---|
| `claude` | Claude Code | `.claude/` |
| `codex` | Codex | `.codex/` |
| `opencode` | OpenCode | `.opencode/` |
| `zcode` | ZCode | `.zcode/` |
| `mimocode` | MiMo Code | `.mimocode/` |
| `none` | 无 | — |

每个助手目录下会注入：
- `agents/reference-explorer.md` — 知识探索子代理
- `agents/reference-analyzer.md` — 深度分析子代理
- `skills/reference/SKILL.md` — Skill 定义

> ZCode 的 agents 子目录为 `cli/agents/`，其余均为 `agents/`。

## 执行内容

1. **创建 `.reference/` 目录结构**
2. **生成 `reference.settings.json`** — 记录所选助手列表和初始化状态
3. **更新 `.gitignore`** — 确保 `.reference/` 被忽略
4. **生成 `reference.map.jsonl`** — 仓库导航数据
5. **注入 AI 配置文件**（指定 `--agent` 时）— 复制子代理 + Skill 到每个助手目录

## 示例

```bash
# 单个助手
reference init --agent claude
reference init --agent codex

# 多个助手同时配置
reference init --agent claude,codex
reference init --agent claude,opencode,zcode

# 仅使用仓库管理功能（不注入 AI 配置）
reference init
reference init --agent none

# 切换助手类型（覆盖现有配置）
reference init --agent opencode
```

## 与无参数运行的差异

| | `reference`（无参数） | `reference init` |
|:---|:---|:---|
| 交互方式 | 首次运行交互式引导（支持多选） | 完全非交互 |
| 适用场景 | 人工使用 | CI/CD、自动化脚本 |
| 注入流程 | 相同 | 相同 |

## 相关命令

- `reference` — 交互式引导入口
- `reference doctor` — 诊断并修复注入的 AI 配置
