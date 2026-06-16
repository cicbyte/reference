# Codex Subagents 文档

> 来源：https://www.codex-docs.com/configuration/subagents/
> 官方文档：https://developers.openai.com/codex

## 核心功能

Codex 可以通过并行启动多个专用智能体，并在最后汇总各自的结果，来运行子智能体工作流。适合并行度高的复杂任务，例如大规模代码库探索，或按步骤推进的多阶段功能实现。

借助子智能体工作流，可以按任务类型定义自己的自定义智能体，并为它们指定不同的模型配置和开发者指令。

## 可用性

- 当前版本的 Codex **默认启用**子智能体工作流
- 子智能体活动会显示在 Codex App 和 CLI 中；IDE 扩展的可视化支持后续补上
- Codex 只会在**明确要求**时生成子智能体
- 每个子智能体都会独立进行模型调用和工具调用，相比单智能体运行会消耗更多 token

## 典型工作流

Codex 负责所有编排工作，包括：
- 生成新的子智能体
- 给不同智能体路由后续指令
- 等待结果
- 关闭子线程

多个智能体并行运行时，Codex 会等待所有要求的结果都返回后，再给出一份汇总答复。

示例提示词：

```
I would like to review the following points on the current PR (this branch vs main).
Spawn one agent per point, wait for all of them, and summarize the result for each point.
1. Security issue
2. Code quality
3. Bugs
4. Race
5. Test flakiness
6. Maintainability of the code
```

## 管理子智能体

- **CLI 中使用 `/agent`**：在活跃智能体线程之间切换，查看正在运行的线程
- 也可以直接要求 Codex 引导某个正在运行的子智能体、停止它，或关闭已完成的智能体线程

## 审批与沙箱控制

- 子智能体会**继承当前会话的沙箱策略**
- 交互式 CLI 中，即使停留在主线程，审批请求也可能来自暂时不在前台的智能体线程
- 审批浮层显示来源线程标签；按 `o` 打开对应线程
- 非交互流程中，任何需要新批准的动作都会失败，错误抛回父工作流
- Codex 生成子智能体时，会重新应用父 `turn` 的实时运行时覆盖项（包括 `/approvals` 改的设置、`--yolo` 等）
- 可以**为单个自定义智能体单独覆盖沙箱配置**（例如强制只读模式）

## 内建智能体

Codex 自带三种内建智能体：

| 智能体 | 用途 |
|:---|:---|
| `default` | 通用回退智能体 |
| `worker` | 偏执行，适合实现和修复类任务 |
| `explorer` | 偏只读，适合代码库探索类任务 |

## 自定义智能体

在以下两个目录之一中放置独立的 **TOML 文件**（注意：是 `.toml`，不是 `.md`）：

- **个人级**：`~/.codex/agents/`
- **项目级**：`.codex/agents/`

每个文件定义一个自定义智能体。Codex 把它作为"额外配置层"应用到生成的会话中，所以自定义智能体能覆盖的键，基本与普通 `config.toml` 可覆盖项一致。

> Codex 识别自定义智能体依赖的是 `name` 字段，不是文件名。通常让文件名和 `name` 一致最省事，但最终以 `name` 为准。
> 如果自定义智能体的 `name` 与内建智能体重名（如 `explorer`），自定义智能体优先生效。

### 必需字段

| 字段 | 类型 | 必需 | 说明 |
|:---|:---|:---|:---|
| `name` | string | 是 | Codex 生成或引用该智能体时使用的名字 |
| `description` | string | 是 | 给 Codex 的人类可读说明，提示何时应选择该智能体 |
| `developer_instructions` | string | 是 | 定义该智能体行为的核心指令 |

### 可选字段

`nickname_candidates`、`model`、`model_reasoning_effort`、`sandbox_mode`、`mcp_servers`、`skills.config` 都是可选的；省略时继承父会话。

### 全局设置

全局子智能体设置写在主配置的 `[agents]` 下：

| 字段 | 类型 | 必需 | 说明 |
|:---|:---|:---|:---|
| `agents.max_threads` | number | 否 | 允许同时打开的智能体线程上限（默认 **6**） |
| `agents.max_depth` | number | 否 | 智能体生成嵌套深度，根会话深度从 0 开始（默认 **1**） |
| `agents.job_max_runtime_seconds` | number | 否 | `spawn_agents_on_csv` 中每个 `worker` 的默认超时（未设置时回退到 **1800** 秒） |

> 提高 `max_depth` 会放大 token 消耗、延迟和本地资源占用，建议保持默认值。

### 显示昵称

`nickname_candidates` 让 Codex 在界面中给同一类智能体显示更可读的昵称。这在同时运行很多个同类自定义智能体时特别有用，避免界面出现大量重复名称。

- 昵称只影响展示层
- Codex 内部识别和生成智能体时，依然使用 `name`
- 必须是非空、去重后的列表
- 昵称只允许使用 `ASCII` 字母、数字、空格、连字符和下划线

示例：

```toml
name = "reviewer"
description = "PR reviewer focused on correctness, security, and missing tests."
developer_instructions = """
Review code like an owner.
Prioritize correctness, security, behavior regressions, and missing test coverage.
"""
nickname_candidates = ["Atlas", "Delta", "Echo"]
```

## 自定义智能体示例

### 示例 1：PR 评审

把 PR 评审拆分给三个各自聚焦的自定义智能体：

- `pr_explorer` — 梳理代码库并收集证据
- `reviewer` — 检查正确性、安全性和测试风险
- `docs_researcher` — 通过专用 MCP server 核对框架或 API 文档

`.codex/config.toml`：

```toml
[agents]
max_threads = 6
max_depth = 1
```

`.codex/agents/pr-explorer.toml`：

```toml
name = "pr_explorer"
description = "Read-only codebase explorer for gathering evidence before changes are proposed."
model = "gpt-5.3-codex-spark"
model_reasoning_effort = "medium"
sandbox_mode = "read-only"
developer_instructions = """
Stay in exploration mode.
Trace the real execution path, cite files and symbols, and avoid proposing fixes unless the parent agent asks for them.
Prefer fast search and targeted file reads over broad scans.
"""
```

`.codex/agents/reviewer.toml`：

```toml
name = "reviewer"
description = "PR reviewer focused on correctness, security, and missing tests."
model = "gpt-5.4"
model_reasoning_effort = "high"
sandbox_mode = "read-only"
developer_instructions = """
Review code like an owner.
Prioritize correctness, security, behavior regressions, and missing test coverage.
Lead with concrete findings, include reproduction steps when possible, and avoid style-only comments unless they hide a real bug.
"""
```

`.codex/agents/docs-researcher.toml`：

```toml
name = "docs_researcher"
description = "Documentation specialist that uses the docs MCP server to verify APIs and framework behavior."
model = "gpt-5.4-mini"
model_reasoning_effort = "medium"
sandbox_mode = "read-only"
developer_instructions = """
Use the docs MCP server to confirm APIs, options, and version-specific behavior.
Return concise answers with links or exact references when available.
Do not make code changes.
"""

[mcp_servers.openaiDeveloperDocs]
url = "https://developers.openai.com/mcp"
```

## CSV 批量任务（实验性）

> 仍属实验性能力，后续可能变化。

`spawn_agents_on_csv` 适合"每一行对应一个相似工作项"的批量任务：Codex 读取 CSV，按行生成 `worker` 子智能体，等待所有任务完成，然后把结果重新导出成 CSV。

适合场景：
- 按行审查文件、包或服务
- 检查一组事故、PR 或迁移目标
- 为大量相似输入生成结构化摘要

核心参数：
- `csv_path`：源 CSV 路径
- `instruction`：`worker` 提示词模板，可使用 `{column_name}` 占位
- `id_column`：用某一列作为稳定的 `item id`
- `output_schema`：每个 `worker` 返回固定结构的 JSON 对象
- `output_csv_path`、`max_concurrency`、`max_runtime_seconds`：控制整个批量任务

> 每个 `worker` 都必须**恰好调用一次** `report_agent_job_result`。未上报结果的 `worker` 会在导出 CSV 里被标记为错误。

通过 `codex exec` 运行时，批量任务执行期间 Codex 会在 `stderr` 输出单行进度信息。导出的 CSV 除了原始列数据，还会附加 `job_id`、`item_id`、`status`、`last_error`、`result_json` 等元数据。

## 与其他工具的差异

| 工具 | 自定义智能体格式 | 项目级目录 | 个人级目录 |
|:---|:---|:---|:---|
| Claude Code | Markdown (`.md`) | `.claude/agents/` | `~/.claude/agents/` |
| ZCode | Markdown (`.md`) | `.zcode/cli/agents/` | `~/.zcode/cli/agents/` |
| MiMo Code | Markdown (`.md`) | `.mimocode/agents/` | `~/.config/mimocode/agents/` |
| OpenCode | Markdown (`.md`) | `.opencode/agents/` | `~/.config/opencode/agents/` |
| **Codex** | **TOML (`.toml`)** | `.codex/agents/` | `~/.codex/agents/` |

**Codex 独特之处**：
- 使用 TOML 而非 Markdown 定义智能体
- 必须字段是 `name` + `description` + `developer_instructions`
- 支持嵌套子智能体（`max_depth` 控制）
- 内建 CSV 批量工作流

## 参考资源

- [Agent Skills – OpenAI 官方](https://developers.openai.com/codex/skills)
- [Custom instructions with AGENTS.md](https://developers.openai.com/codex/guides/agents-md)
- [VoltAgent/awesome-agent-skills](https://github.com/VoltAgent/awesome-agent-skills)
