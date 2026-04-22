# AI 编程 Agent 平台适配调研报告

## 概述

本报告调研主流 AI 编程 Agent 平台的配置体系（项目指令、MCP、子代理、Skill 等），评估 `reference` 项目对各平台的适配可行性，并给出适配优先级和实现方案。

## 调研范围

共调研 **16 个平台**，按类型分为三组：

| 类型         | 平台                                                                   |
| :--------- | :------------------------------------------------------------------- |
| CLI 工具     | Claude Code、OpenAI Codex CLI、Aider、Amp、OpenCode、Goose                |
| VS Code 扩展 | Cursor、Windsurf、Cline、Continue、Roo Code、Copilot、Kodu AI、Augment Code |
| IDE        | Trae                                                                 |

***

## 一、各平台配置体系详解

### 1. Claude Code

**指令文件**: `CLAUDE.md`（多层级：企业 → 项目 → 用户 → 本地）

**规则系统**: `.claude/rules/*.md`（支持 `paths:` glob 条件匹配）

**Skill 系统**: `.claude/skills/<name>/SKILL.md`（YAML frontmatter + Markdown，支持 `context: fork` 子代理隔离运行）

**子代理**: `.claude/agents/*.md`（YAML frontmatter 定义工具、模型、权限、MCP 服务器）

**MCP 配置**: `.mcp.json`（项目级）+ `~/.claude.json`（用户级）

**记忆系统**: `~/.claude/projects/<project>/memory/`（自动 + 手动）

**Hooks 系统**: `settings.json` 中定义，支持 20+ 事件类型

**关键特性**:

- `@path/to/file` 导入语法（递归，最大 5 层）
- Skill 支持 `context: fork` 在隔离子代理中运行
- 子代理支持 `isolation: worktree` 临时 git worktree
- `` !`command` `` 动态上下文注入
- 实时变更检测：会话中修改 skill/agent 立即生效

**适配难度**: ★☆☆☆☆（reference 当前原生支持）

***

### 2. Cursor

**指令文件**: `.cursorrules`（旧版）+ `.cursor/rules/*.mdc`（新版）

**规则系统**: `.cursor/rules/` 支持 `.md` 和 `.mdc` 格式

**规则触发**: YAML frontmatter 控制

- `alwaysApply: true` — 每次对话自动注入
- `globs: ["**/*.ts"]` — 文件匹配时激活
- 无 frontmatter — 用户手动 @-mention

**MCP 配置**: `.cursor/mcp.json`（项目级）+ `~/.cursor/mcp.json`（全局）

**AGENTS.md**: 根目录支持，零配置自动发现

**关键特性**:

- 约 40 个活跃 MCP 工具上限
- `.cursorignore` 排除文件
- 支持远程规则导入（`.cursor/rules/imported/`）
- MCP Apps 扩展（工具返回交互式 UI）
- `.mdc` 无正确 frontmatter 会静默忽略

**适配难度**: ★★★☆☆（需要转换格式）

***

### 3. Windsurf (Codeium)

**指令文件**: `.windsurfrules`（旧版）+ `.windsurf/rules/*.md`（新版）

**规则触发**: YAML frontmatter `trigger` 字段

- `always_on` — 始终激活
- `model_decision` — Agent 判断相关性
- `glob` — 文件匹配（`globs: **/*.test.ts`）
- `manual` — 手动 @-mention

**MCP 支持**: 原生支持（通过设置界面配置）

**AGENTS.md**: 根目录和子目录支持，零配置自动发现

**工作流**: `.windsurf/workflows/*.md`（可复用多步骤流程，`/workflow-name` 调用）

**记忆系统**: `~/.codeium/windsurf/memories/`（自动生成，工作区隔离）

**关键特性**:

- Rules 每文件 12000 字符限制
- Global Rules 单文件 6000 字符限制
- Skills 系统可捆绑附带文件（脚本、模板）
- 企业级 System Rules/Workflows

**适配难度**: ★★★☆☆（需要转换格式 + 拆分大规则）

***

### 4. Cline

**指令文件**: `.clinerules/*.md`（兼容 `.cursorrules`、`.windsurfrules`、`AGENTS.md`）

**规则触发**: YAML frontmatter `paths:` glob 条件匹配

**MCP 配置**: `~/.cline/data/settings/cline_mcp_settings.json`

**MCP 传输**: STDIO + SSE

**记忆系统**: `memory-bank/`（推荐结构化文档系统）+ Checkpoints

**关键特性**:

- 兼容三种其他平台的规则文件
- `.clineignore` 大幅减少上下文消耗
- Plan & Act 模式分离
- MCP Marketplace
- CLI 支持：`cline mcp add`、`/newrule`、`/smol`

**适配难度**: ★★☆☆☆（天然兼容部分格式）

***

### 5. OpenAI Codex CLI

**指令文件**: `AGENTS.md`（多层级：全局 `~/.codex/` → 项目根目录 → 当前目录，逐级拼接）

**Skill 系统**: `.agents/skills/<name>/SKILL.md`（YAML frontmatter + `openai.yaml` 依赖声明）

**子代理**: 完整多代理协作（`spawn_agent`/`send_input`/`resume_agent`/`wait_agent`/`close_agent`）

**MCP 配置**: `~/.codex/config.toml` 的 `[mcp_servers]` 节（支持 stdio + streamable HTTP）

**沙箱**: 三级沙箱（read-only / workspace-write / full-access），Linux Seatbelt + Windows 原生

**配置格式**: TOML（`~/.codex/config.toml`）

**记忆系统**: Memories（实验性，自动提取 + 全局整合）

**关键特性**:

- `AGENTS.override.md` 可覆盖同目录的 `AGENTS.md`
- 指令大小限制 32 KiB（可配置 `project_doc_max_bytes`）
- MCP 工具审批策略：auto / prompt / approve（全局 + 单工具粒度）
- Guardian 代理实验性审查审批请求
- Plugin 系统（打包技能 + App 集成 + MCP 配置的分发单元）
- Hooks 生命周期钩子
- Profiles 配置分组继承
- App Server 供 IDE 扩展连接
- Python + TypeScript SDK

**适配难度**: ★★☆☆☆（AGENTS.md + Skill 系统与 Claude Code 高度相似）

***

### 6. Aider

**指令文件**: `CONVENTIONS.md`（通过 `--read` 或 `.aider.conf.yml` 的 `read:` 字段加载）

**上下文管理**: Repo Map（自动生成的仓库地图，token 预算控制）

**配置文件**: `.aider.conf.yml`（项目级 + 全局）

**MCP 支持**: **不支持**

**关键特性**:

- Architect 双模型模式（architect 方案 + editor 执行）
- 无子代理、无 Skill、无规则系统
- 完全基于 CLI + git 集成
- 支持 15+ 模型提供商

**适配难度**: ★★★★★（几乎无法适配，无 MCP/Skill/规则系统）

***

### 7. GitHub Copilot

**指令文件**: `.github/copilot-instructions.md` + `.github/instructions/*.instructions.md`

**规则触发**: `.instructions.md` 的 YAML frontmatter `applyTo:` glob 模式

**MCP 配置**: `.vscode/mcp.json`（项目级）+ VS Code `settings.json`（全局）

**AGENTS.md**: 支持（实验性）

**CLAUDE.md**: 兼容读取

**记忆系统**: 本地 Memory Tool（User/Repository/Session 三层）+ GitHub Copilot Memory（托管）

**关键特性**:

- 支持 `#tool:web/fetch` 在指令文件中引用 Agent 工具
- Chat Customizations Editor 统一管理
- Toolsets 工具分组
- 跨 VS Code/JetBrains/Eclipse/Xcode/CLI

**适配难度**: ★★★☆☆（需要转换指令格式 + MCP 配置路径）

***

### 8. Continue

**指令文件**: `.continue/rules/*.md`（YAML frontmatter）

**规则触发**: `globs:` + `regex:` 条件匹配

**MCP 配置**: `.continue/mcpServers/*.yaml`（推荐）+ `config.yaml` 内联

**MCP 传输**: stdio + sse + streamable-http

**关键特性**:

- `config.ts` 可编程化配置（自定义 slash commands）
- Hub 规则引用（`uses: username/my-hub-rule`）
- 兼容 Claude/Cursor/Cline 的 JSON MCP 配置（`.continue/mcpServers/mcp.json`）
- Context Providers 已废弃，改用 MCP

**适配难度**: ★★★☆☆（需要转换规则格式 + MCP 配置路径）

***

### 9. Roo Code

**指令文件**: `.roo/rules/*.md` + `.roomodes`（自定义模式配置）

**规则触发**: 无条件全部加载（无 glob 匹配）

**MCP 配置**: `.roo/mcp.json`（项目级）+ VS Code `mcp_settings.json`（全局）

**AGENTS.md**: 支持

**模式系统**: 内置 Code/Ask/Debug/Architect/Orchestrator + 自定义模式（YAML）

**关键特性**:

- 模式专用规则目录（`.roo/rules-code/`、`.roo/rules-{modeSlug}/`）
- 工具组权限控制（`groups: [read, edit, command, mcp]`）
- MCP 服务器自动创建
- 模式导入/导出（YAML 格式，含规则）

**适配难度**: ★★★☆☆（需要转换规则 + MCP 路径）

***

### 10. Amp (Sourcegraph)

**指令文件**: `AGENTS.md`（Amp 是该标准的发起者）

**Skill 系统**: `.agents/skills/`（打包指令 + MCP 服务器）

**MCP 配置**: Skill 内 `mcp.json` 或 `amp mcp add` CLI

**子代理**: 内置 Subagents + Oracle（第二意见）+ Librarian（代码搜索）

**关键特性**:

- 3 种模式：smart（Opus 4.6）、rush（快速）、deep（GPT-5.4）
- Toolboxes：简单脚本扩展（无需 MCP 服务器）
- Checks：自定义代码审查规则
- Execute Mode：`amp -x "任务"` 非交互式执行
- `@file` 引用 + `globs:` 条件匹配

**适配难度**: ★★☆☆☆（AGENTS.md 兼容，Skill 系统类似）

***

### 11. Kodu AI

**指令文件**: `.kodu`（纯文本，直接注入系统提示词）

**MCP 支持**: 继承 VS Code Agent Mode MCP

**关键特性**:

- 无规则系统、无子代理、无 Skill
- 极简配置，仅一个文件
- 基于 Claude 3.7 Sonnet

**适配难度**: ★★★★☆（只能注入指令文件，无 Skill/MCP/子代理）

***

### 12. Augment Code

**指令文件**: `AUGMENT.md`（项目根目录）

**MCP 支持**: 全面支持（Easy MCP 一键集成 + Settings Panel + JSON 导入）

**关键特性**:

- 支持 VS Code/JetBrains/Vim/Neovim/Slack
- `${workspaceFolder}` 变量扩展
- Next Edit 逐步引导

**适配难度**: ★★★☆☆（只需生成指令文件 + MCP 配置）

***

### 13. Trae (ByteDance)

**指令文件**: `.trae/rules`（IDE）+ `trae_config.yaml`（CLI）

**MCP 支持**: 全面支持（CLI 在 `trae_config.yaml` 中配置）

**Agent 模式**: SOLO Mode（AI 主导）+ Builder Mode（项目生成）

**关键特性**:

- Agent 市场（创建和分享自定义 Agent）
- 开源 CLI（`trae-agent`，MIT 协议）
- 多模型支持（Claude/GPT/DeepSeek/Gemini/Doubao）
- Docker 执行模式

**适配难度**: ★★★☆☆（需要转换规则格式 + MCP 配置路径）

***

### 14. Goose (Block)

**指令文件**: Recipes（YAML 格式工作流）

**MCP 支持**: 最早期的 MCP 采用者，70+ 文档化扩展

**子代理**: 支持并行子代理

**关键特性**:

- MCP Apps 规范（扩展可渲染交互式 UI）
- Agent Client Protocol (ACP)：可作为服务器或客户端
- 开源（Apache 2.0，Linux Foundation）
- 安全特性：提示注入检测、沙箱模式

**适配难度**: ★★★☆☆（需要转换为 Recipes YAML 格式）

***

### 15. Sourcegraph Cody

**指令文件**: `.cody/instructions` 或 `.cody/instructions.md`

**MCP 支持**: 支持（Sourcegraph MCP Server）

**兼容**: `.github/copilot-instructions.md`

**适配难度**: ★★★☆☆（只需生成指令文件）

***

### 16. OpenCode (opencode-ai)

**指令文件**: `opencode.md`（同时兼容 `CLAUDE.md`、`.cursorrules`、`.cursor/rules/`、`.github/copilot-instructions.md`）

**MCP 配置**: `.opencode.json` 的 `mcpServers` 节（支持 stdio + SSE）

**配置格式**: JSON（`.opencode.json`）

**关键特性**:

- **自动发现多种指令文件**（contextPaths 默认包含 CLAUDE.md、.cursorrules 等 10+ 种格式）
- 4 个内置 Agent 角色（coder / summarizer / task / title），`task` 角色用于子任务
- Auto Compact（95% 上下文窗口时自动触发摘要）
- LSP 集成（多语言 diagnostics）
- 自定义命令（Markdown 文件，支持 `$PARAM` 参数）
- 多模型支持（OpenAI/Anthropic/Gemini/Groq/Azure/Bedrock 等 10+ 提供商）
- **项目已归档**，迁移至 [charmbracelet/crush](https://github.com/charmbracelet/crush)

**适配难度**: ★☆☆☆☆（天然兼容 CLAUDE.md 和 .cursorrules 等多种格式）

***

## 二、跨平台兼容标准

### AGENTS.md 标准

由 Amp 发起的供应商中立标准，越来越多平台支持：

| 平台               | 支持情况                                         |
| :--------------- | :------------------------------------------- |
| Amp              | 发起者，完整支持                                     |
| Claude Code      | 通过 `@AGENTS.md` 导入                           |
| Cursor           | 支持                                           |
| Windsurf         | 支持                                           |
| Cline            | 支持                                           |
| Roo Code         | 支持                                           |
| GitHub Copilot   | 支持（实验性）                                      |
| Sourcegraph Cody | 支持                                           |
| Claude Code      | 通过 CLAUDE.md 导入                              |
| OpenAI Codex CLI | 原生支持（AGENTS.md 是核心指令机制）                      |
| OpenCode         | 不原生支持 AGENTS.md，但兼容 CLAUDE.md 和 .cursorrules |

### MCP 配置格式对比

| 平台          | 配置路径                          | 格式   |
| :---------- | :---------------------------- | :--- |
| Claude Code | `.mcp.json`                   | JSON |
| Cursor      | `.cursor/mcp.json`            | JSON |
| Cline       | `cline_mcp_settings.json`     | JSON |
| Continue    | `.continue/mcpServers/*.yaml` | YAML |
| Roo Code    | `.roo/mcp.json`               | JSON |
| Copilot     | `.vscode/mcp.json`            | JSON |
| Augment     | GUI + JSON 导入                 | JSON |
| Trae        | `trae_config.yaml`            | YAML |
| Goose       | CLI 配置                        | JSON |
| Codex CLI   | `~/.codex/config.toml`        | TOML |
| OpenCode    | `.opencode.json`              | JSON |

**共性**: 几乎所有平台都使用 `mcpServers` 作为顶层键，`command`/`args`/`env` 作为 STDIO 传输的字段名。

### 规则/指令文件格式对比

| 平台          | 格式                                       | 条件匹配                     | Frontmatter                                    |
| :---------- | :--------------------------------------- | :----------------------- | :--------------------------------------------- |
| Claude Code | `.claude/rules/*.md`                     | `paths:` glob            | `paths:`                                       |
| Cursor      | `.cursor/rules/*.mdc`                    | `globs:` + `alwaysApply` | `globs`, `alwaysApply`, `description`          |
| Windsurf    | `.windsurf/rules/*.md`                   | `trigger` + `globs`      | `trigger`, `globs`, `description`              |
| Cline       | `.clinerules/*.md`                       | `paths:` glob            | `paths`                                        |
| Continue    | `.continue/rules/*.md`                   | `globs:` + `regex`       | `globs`, `regex`, `alwaysApply`, `description` |
| Copilot     | `.github/instructions/*.instructions.md` | `applyTo:` glob          | `applyTo`, `name`, `description`               |
| Roo Code    | `.roo/rules/*.md`                        | 无                        | 无                                              |
| Codex CLI   | `AGENTS.md`                              | 无（逐级拼接）                  | 无（纯 Markdown）                                  |
| OpenCode    | `opencode.md`（兼容多种）                      | 无（扁平拼接）                  | 无                                              |

**共性**: 大多数平台使用 Markdown + YAML frontmatter，条件匹配通过 glob 模式实现，但字段名不统一。

***

## 三、reference 适配方案

### 核心适配原则

reference 的价值在于让 AI Agent 零延迟查阅本地仓库源码。适配的核心需求：

1. **注入项目指令**: 告诉 Agent "本地有哪些仓库可以参考"
2. **提供 Skill/Agent**: 让 Agent 能探索源码并沉淀知识
3. **配置 MCP 服务器**: 可选，用于直接查询仓库信息

### 适配优先级矩阵

| 优先级          | 平台               | 理由                                                   |
| :----------- | :--------------- | :--------------------------------------------------- |
| **P0 — 已适配** | Claude Code      | 原生支持，Skill + Agent + MCP 完整实现                        |
| **P1 — 高价值** | OpenAI Codex CLI | AGENTS.md + Skill + 子代理 + MCP 完整体系，与 Claude Code 最相似 |
| **P1 — 高价值** | Cursor           | 市占率高，有规则 + MCP，但无子代理/Skill                           |
| **P1 — 高价值** | Roo Code         | 开源，有自定义模式 + MCP，社区活跃                                 |
| **P2 — 中价值** | Windsurf         | 有规则 + MCP + Workflows                                |
| **P2 — 中价值** | Cline            | 兼容多种规则格式 + MCP                                       |
| **P2 — 中价值** | Continue         | 有规则 + MCP + config.ts 可编程化                           |
| **P2 — 中价值** | GitHub Copilot   | 用户基数最大，有指令 + MCP                                     |
| **P2 — 低成本** | OpenCode         | 天然兼容 CLAUDE.md/.cursorrules 等多种格式（已归档→Crush）         |
| **P3 — 低价值** | Amp              | 有 Skill + MCP，但 CLI 工具用户较少                           |
| **P3 — 低价值** | Trae             | 免费但生态较新                                              |
| **P3 — 低价值** | Augment Code     | 有 MCP 但用户较少                                          |
| **P3 — 低价值** | Goose            | 开源但偏通用 Agent                                         |
| **P3 — 低价值** | Sourcegraph Cody | 有 MCP 但用户较少                                          |
| **不适配**      | Aider            | 无 MCP/Skill/规则系统                                     |
| **不适配**      | Kodu AI          | 仅支持 `.kodu` 纯文本指令                                    |

### 各平台适配方案

#### P0: Claude Code（已适配）

```
.claude/
├── skills/reference/SKILL.md           # Skill 模板
├── agents/reference-explorer.md        # 探索子代理
├── agents/reference-analyzer.md        # 分析子代理
└── skills/reference/<仓库名>/           # Junction → 全局 wiki
.mcp.json                               # MCP 服务器配置（可选）
```

当前已完整实现，无需修改。

#### P1: OpenAI Codex CLI

```
AGENTS.md                               # 项目指令（Codex 原生格式）
.agents/skills/reference/SKILL.md       # Skill（与 Claude Code 格式高度相似）
.agents/skills/reference/agents/
└── openai.yaml                         # Skill 元数据和 MCP 依赖
~/.codex/config.toml                    # MCP 服务器配置（可选）
```

**转换规则**:

- `SKILL.md` 格式几乎相同（YAML frontmatter `name` + `description`）
- 子代理 → Codex 的多代理协作（`spawn_agent` 工具）
- MCP 配置需从 JSON 转为 TOML 格式
- `openai.yaml` 声明 Skill 依赖的 MCP 服务器

**优势**: Codex CLI 的 Skill + 子代理体系与 Claude Code 最相似，适配成本最低。

#### P1: Cursor

```
.cursor/rules/
├── reference.mdc                       # 项目指令（alwaysApply: true）
├── reference-explorer.mdc              # 探索指令（手动 @-mention）
└── reference-analyzer.mdc              # 分析指令（手动 @-mention）
.cursor/rules/<仓庛>/                   # Junction → 全局 wiki（按需读取）
.cursor/mcp.json                        # MCP 服务器配置（可选）
```

**转换规则**:

- `SKILL.md` → `.mdc` 文件（`description` + `alwaysApply` frontmatter）
- 子代理 prompt → 规则文件（Cursor 无子代理，降级为手动触发的规则）
- `paths:` → `globs:` 字段名映射

**限制**: Cursor 无子代理机制，探索任务会在主对话中执行，会污染上下文。可通过 AGENTS.md 注入基础指令，让用户手动触发探索。

#### P1: Roo Code

```
.roo/rules/
├── 01-reference.md                     # 基础指令（无条件加载）
├── 02-reference-explorer.md            # 探索指令
└── 03-reference-analyzer.md            # 分析指令
.roo/rules-code/reference-*.md          # Code 模式专用规则
.roomodes                               # 自定义模式（可选）
.roo/mcp.json                           # MCP 服务器配置
```

**转换规则**:

- Skill prompt → `.roo/rules/*.md`（Roo 无条件加载所有规则）
- 可通过 `.roomodes` 自定义模式，创建 `reference-explorer` 和 `reference-analyzer` 模式
- 模式中限制工具组为 `[read, mcp]`，避免误编辑源码

**优势**: Roo 的自定义模式系统最接近 Claude Code 的子代理概念。

#### P2: Windsurf

```
.windsurf/rules/
├── reference.md                        # 基础指令（trigger: always_on）
├── reference-explorer.md               # 探索指令（trigger: manual）
└── reference-analyzer.md               # 分析指令（trigger: manual）
.windsurf/workflows/
├── explore-repo.md                     # 探索工作流
└── analyze-repo.md                     # 分析工作流
```

**转换规则**:

- Skill → Workflow（Windsurf 的 Workflow 最接近 Claude Code 的 Skill）
- `context: fork` → 无对应机制，降级为手动触发

#### P2: Cline

```
.clinerules/
├── 01-reference.md                     # 基础指令
├── 02-reference-explorer.md            # 探索指令
└── 03-reference-analyzer.md            # 分析指令
```

**转换规则**:

- Skill → 规则文件（Cline 无 Skill，但有 Plan & Act 模式）
- 天然兼容 `.cursorrules` 和 `.windsurfrules`

#### P2: Continue

```
.continue/rules/
├── reference.md                        # 基础指令
├── reference-explorer.md               # 探索指令
└── reference-analyzer.md               # 分析指令
.continue/mcpServers/mcp.json           # MCP 配置（兼容 JSON 格式）
```

**转换规则**:

- Skill → 规则文件
- MCP 配置可直接复用 Claude Code 的 `.mcp.json`

#### P2: GitHub Copilot

```
.github/copilot-instructions.md         # 基础指令
.github/instructions/
├── reference.instructions.md           # 仓库参考指令
└── code-review.instructions.md         # 代码审查指令（可选）
.vscode/mcp.json                        # MCP 服务器配置
```

**转换规则**:

- Skill → `.instructions.md` 文件（`applyTo:` glob 条件匹配）
- 子代理 → 无对应机制
- CLAUDE.md 自动兼容读取

#### P3: Amp

```
AGENTS.md                               # 基础指令（Amp 原生支持）
.agents/skills/reference/               # Skill 目录（类似 Claude Code）
```

**转换规则**: Amp 的 Skill 系统与 Claude Code 高度相似，转换成本最低。

#### P2: OpenCode（低成本）

```
# OpenCode 自动发现以下任一文件即可：
CLAUDE.md                                # 已有，OpenCode 原生兼容
opencode.md                              # OpenCode 专用指令文件
.opencode.json                           # MCP 服务器配置
```

**转换规则**:

- OpenCode 的 `contextPaths` 默认已包含 `CLAUDE.md` 和 `.cursorrules`
- 如果项目已为 Claude Code 适配了 CLAUDE.md，OpenCode **零额外配置即可使用**
- MCP 配置从 JSON 格式直接复用

#### P3: Trae / Augment / Goose / Sourcegraph Cody

```
# Trae
.trae/rules/reference.md               # 项目规则

# Augment
AUGMENT.md                              # 项目指令

# Goose
recipes/reference.yaml                  # Recipe 工作流
```

均只支持基础指令注入，无 Skill/子代理。

***

## 四、实现建议

### Phase 1: 多平台指令注入（低成本，高覆盖）

为每个目标平台生成对应格式的**项目指令文件**，内容基于当前的 `SKILL.md` 模板：

```go
// internal/logic/inject/inject.go
// 新增平台参数
func InjectForPlatform(projectDir string, platform string) error
```

支持的平台标识：`claude`、`cursor`、`windsurf`、`cline`、`continue`、`roo`、`copilot`、`amp`、`trae`、`augment`

每个平台生成：

1. **项目指令文件**（告知 Agent 本地有哪些仓库可参考）
2. **MCP 配置文件**（可选，统一 `mcpServers` 格式）
3. **知识目录 Junction**（全局 wiki 链接）

### Phase 2: Skill/规则转换（中等成本）

将 `SKILL.md` 模板转换为各平台规则格式：

| 源格式              | 目标格式                       | 转换点                           |
| :--------------- | :------------------------- | :---------------------------- |
| YAML frontmatter | Cursor `.mdc`              | `description` + `alwaysApply` |
| YAML frontmatter | Windsurf `.md`             | `trigger:` 字段                 |
| YAML frontmatter | Continue `.md`             | `globs:` + `alwaysApply`      |
| 无 frontmatter    | Roo `.md`                  | 直接写入                          |
| YAML frontmatter | Copilot `.instructions.md` | `applyTo:` 字段                 |

### Phase 3: 子代理降级方案（高成本）

对于不支持子代理的平台，核心挑战是**上下文隔离**。可能的方案：

1. **MCP 服务器方式**: 将探索逻辑封装为 MCP 服务器，Agent 通过工具调用获取结果，不直接读取源码
2. **Pre-computed 方式**: `reference analyze` 预先完成深度分析，生成轻量知识文件供任何平台使用
3. **CLI 脚本方式**: 提供 `reference explore <仓库> <主题>` 命令，用户手动运行后将结果粘贴给 Agent

***

## 五、总结

| 维度        | 结论                                                                        |
| :-------- | :------------------------------------------------------------------------ |
| **最容易适配** | Claude Code（已适配）、OpenCode（天然兼容 CLAUDE.md）、Amp（Skill 系统相似）、Cline（兼容多格式）    |
| **最值得适配** | OpenAI Codex CLI（体系最相似）、Cursor（市占率）、Roo Code（自定义模式）、Windsurf（Workflows）   |
| **最难适配**  | Aider（无 MCP/Skill/规则）、Kodu AI（极简）                                         |
| **最大挑战**  | 子代理隔离 — 仅 Claude Code、Codex CLI 和 Amp 支持真正的上下文隔离                          |
| **通用方案**  | AGENTS.md + CLAUDE.md + MCP 配置是跨平台兼容性最好的组合                                |
| **推荐策略**  | Phase 1 指令注入覆盖 16 个平台 → Phase 2 规则转换覆盖 10 个平台 → Phase 3 MCP 服务器实现跨平台上下文隔离 |

