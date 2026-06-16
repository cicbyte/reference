# reference init

非交互式初始化项目配置，适用于 CI/CD 集成、自动化脚本等场景。

## 用法

```bash
reference init [--agent <type>]
```

## 标志

| 标志 | 说明 |
|:---|:---|
| `--agent <type>` | 编程助手类型，可选值见下表，默认 `none` |

## 支持的助手类型

| 值 | 助手 | agents 目录 | skills 目录 | agent 格式 |
|:---|:---|:---|:---|:---|
| `claude` | Claude Code | `.claude/agents/` | `.claude/skills/` | Markdown |
| `codex` | Codex | `.codex/agents/` | `.codex/skills/` | TOML |
| `opencode` | OpenCode | `.opencode/agents/` | `.opencode/skills/` | Markdown |
| `zcode` | ZCode | `.zcode/cli/agents/` | `.zcode/skills/` | Markdown |
| `mimocode` | MiMo Code | `.mimocode/agents/` | `.mimocode/skills/` | Markdown |
| `none` | 无 | — | — | — |

## 执行内容

1. **创建 `.reference/` 目录结构**
2. **生成 `reference.settings.json`** — 记录所选助手类型和初始化状态
3. **更新 `.gitignore`** — 确保 `.reference/` 被忽略
4. **生成 `reference.map.jsonl`** — 仓库导航数据
5. **注入 AI 配置文件**（指定 `--agent` 时）— 复制子代理 + Skill 到对应目录

## 示例

```bash
# CI/CD 中初始化为 Claude Code 项目
reference init --agent claude

# 初始化为 Codex 项目
reference init --agent codex

# 仅使用仓库管理功能（不注入 AI 配置）
reference init
reference init --agent none

# 切换助手类型（覆盖现有配置）
reference init --agent opencode
```

## 与无参数运行的差异

| | `reference`（无参数） | `reference init` |
|:---|:---|:---|
| 交互方式 | 首次运行交互式引导 | 完全非交互 |
| 适用场景 | 人工使用 | CI/CD、自动化脚本 |
| 注入流程 | 相同 | 相同 |

## 相关命令

- `reference` — 交互式引导入口
- `reference doctor` — 诊断并修复注入的 AI 配置
