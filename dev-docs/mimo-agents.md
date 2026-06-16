# MiMo Code Agents 文档

> 来源：https://mimo.xiaomi.com/zh/mimocode/agents

## 核心功能

Agents 是可针对特定任务和工作流配置的专用 AI 助手。允许创建带有自定义提示词、模型和工具访问权限的聚焦工具。

> **提示**：使用 plan agent 分析代码并审查建议，而不做任何代码改动。

可以在会话中切换 agents，或通过 `@` 提及调用。

## 类型

MiMo Code 有两类 agents：primary agents（主代理）和 subagents（子代理）。

### Primary Agents（主代理）

主代理是你直接交互的主要助手。可以用 **Tab** 键循环切换，或用配置的 `switch_agent` 快捷键。这些 agents 处理你的主对话。工具访问通过权限配置——例如 Build 启用所有工具，而 Plan 受限。

> **提示**：会话中可以用 **Tab** 键在主 agents 之间切换。

MiMo Code 自带两个内置主 agent：**Build** 和 **Plan**。

### Subagents（子代理）

子代理是主代理可以为特定任务调用的专用助手。也可以在消息中通过 **@ 提及**手动调用。

MiMo Code 自带两个内置子 agent：**General** 和 **Explore**。

## 内置 Agents

MiMo Code 自带两个内置主 agent 和两个内置子 agent。

### Build（主代理）

- **Mode**: `primary`
- Build 是**默认**主 agent，启用所有工具。这是开发工作的标准 agent，需要完整文件操作和系统命令访问权限。

### Plan（主代理）

- **Mode**: `primary`
- 专为规划和分析设计的受限 agent。使用权限系统提供更多控制并防止意外改动。默认所有以下设置为 `ask`：
  - `file edits`：所有写入、补丁和编辑
  - `bash`：所有 bash 命令
- 适用于让 LLM 分析代码、建议改动或创建计划，而不实际修改代码库。

### General（子代理）

- **Mode**: `subagent`
- 用于研究复杂问题和执行多步任务的通用 agent。拥有完整工具访问权限（除 todo），需要时可以做文件改动。用于并行运行多个工作单元。

### Explore（子代理）

- **Mode**: `subagent`
- 快速只读 agent，用于探索代码库。不能修改文件。需要快速按模式查找文件、搜索代码关键词或回答代码库相关问题时使用。

### Compaction / Title / Summary（系统主代理）

- **Mode**: `primary`
- 隐藏的系统 agent，分别用于压缩长上下文为摘要、生成简短会话标题、创建会话摘要。自动运行，UI 中不可选。

## 使用方式

1. **主 agents**：用 **Tab** 键在会话中循环切换，或用配置的 `switch_agent` 快捷键。
2. **子代理**可以：
   - **自动**被主代理根据描述为特定任务调用
   - 在消息中通过 **@ 提及**手动调用。例如：
     ```
     @general help me search for this function
     ```
3. **会话间导航**：子代理创建子会话时，用 `session_child_first`（默认 **+Down**）从父会话进入第一个子会话。
4. 进入子会话后：
   - `session_child_cycle`（默认 **Right**）循环到下一个子会话
   - `session_child_cycle_reverse`（默认 **Left**）循环到上一个子会话
   - `session_parent`（默认 **Up**）返回父会话

## 配置方式

可以自定义内置 agents 或创建自己的 agent。两种配置方式：

### JSON 配置

在 `mimocode.json` 配置文件中：

```json
{
  "$schema": "https://mimo.xiaomi.com//config.json",
  "agent": {
    "build": {
      "mode": "primary",
      "model": "mimo/mimo-v2.5-pro",
      "prompt": "{file:./prompts/build.txt}",
      "tools": {
        "write": true,
        "edit": true,
        "bash": true
      }
    },
    "plan": {
      "mode": "primary",
      "model": "mimo/mimo-v2.5-pro",
      "tools": {
        "write": false,
        "edit": false,
        "bash": false
      }
    },
    "code-reviewer": {
      "description": "Reviews code for best practices and potential issues",
      "mode": "subagent",
      "model": "mimo/mimo-v2.5-pro",
      "prompt": "You are a code reviewer. Focus on security, performance, and maintainability.",
      "tools": {
        "write": false,
        "edit": false
      }
    }
  }
}
```

### Markdown 配置

用 markdown 文件定义 agent，放置位置：

- **全局**：`~/.config/mimocode/agents/`
- **项目级**：`.mimocode/agents/`

`~/.config/mimocode/agents/review.md`：

```markdown
---
description: Reviews code for quality and best practices
mode: subagent
model: mimo/mimo-v2.5-pro
temperature: 0.1
tools:
  write: false
  edit: false
  bash: false
---
You are in code review mode. Focus on:
- Code quality and best practices
- Potential bugs and edge cases
- Performance implications
- Security considerations

Provide constructive feedback without making direct changes.
```

Markdown 文件名即 agent 名。例如 `review.md` 创建 `review` agent。

## 配置选项详解

### description（描述，必需）

提供 agent 做什么以及何时使用的简短描述。

### temperature（温度）

控制 LLM 响应的随机性和创造性。值越低响应越聚焦和确定，越高越有创意和多变。典型范围 0.0-1.0：

- **0.0-0.2**：非常聚焦和确定，适合代码分析和规划
- **0.3-0.5**：平衡响应带一些创意，适合通用开发任务
- **0.6-1.0**：更有创意和多变，适合头脑风暴和探索

未指定时使用模型特定默认值；多数模型为 0，Qwen 模型为 0.55。

### steps（最大步数，新）

控制 agent 在被迫仅用文本响应前可执行的最大代理迭代次数。允许用户控制成本，限制代理动作。

未设置时，agent 持续迭代直到模型选择停止或用户中断会话。

> **注意**：旧字段 `maxSteps` 已弃用，改用 `steps`。

达到限制时，agent 收到特殊系统提示，指示其用工作总结和推荐剩余任务响应。

### disable（禁用）

设为 `true` 禁用 agent。

### prompt（提示词文件）

为 agent 指定自定义系统提示词文件：

```json
{
  "agent": {
    "review": {
      "prompt": "{file:./prompts/code-review.txt}"
    }
  }
}
```

路径相对于配置文件位置。全局配置和项目特定配置都适用。

### model（模型）

覆盖 agent 的模型。格式 `provider/model-id`，例如 `mimo/mimo-v2.5-pro`。

未指定时，主 agent 使用全局配置的模型，子代理使用调用它的主 agent 的模型。

### tools（已弃用）

> `tools` 已**弃用**。新配置请使用 agent 的 `permission` 字段获取更细粒度控制。

允许通过 `true`/`false` 控制 agent 中可用的工具。`true` 等同于 `{"*": "allow"}` 权限，`false` 等同于 `{"*": "deny"}`。

支持通配符控制多个工具。agent 特定配置覆盖全局配置。

### permission（权限）

配置权限管理 agent 可采取的动作。当前 `edit`、`bash`、`webfetch` 工具的权限可配置为：

- `"ask"` — 运行工具前提示确认
- `"allow"` — 允许所有操作无需批准
- `"deny"` — 禁用工具

可按 agent 覆盖，也可设置特定 bash 命令权限（支持 glob 模式）：

```json
{
  "agent": {
    "build": {
      "permission": {
        "bash": {
          "*": "ask",
          "git status *": "allow"
        }
      }
    }
  }
}
```

最后匹配的规则优先，所以把 `*` 通配符放前面，具体规则放后面。

### mode（模式）

控制 agent 模式。可设为 `primary`、`subagent` 或 `all`。未指定默认 `all`。

### hidden（隐藏）

`hidden: true` 从 `@` 自动补全菜单隐藏子代理。适用于只应被其他 agent 通过 Task 工具程序性调用的内部子代理。

> 仅适用于 `mode: subagent` 的 agents。

### task permissions（任务权限）

通过 `permission.task` 控制 agent 可通过 Task 工具调用哪些子代理。使用 glob 模式灵活匹配。规则按顺序评估，**最后匹配的规则生效**。

> 用户始终可通过 `@` 自动补全菜单直接调用任何子代理，即使 agent 的任务权限会拒绝。

### color（颜色）

自定义 agent 在 UI 中的视觉外观。使用有效十六进制颜色（如 `#FF5733`）或主题色：`primary`、`secondary`、`accent`、`success`、`warning`、`error`、`info`。

### top_p

`top_p` 控制响应多样性，是 temperature 的替代方案。范围 0.0-1.0，值越低越聚焦，越高越多样。

### Additional（透传选项）

其他在 agent 配置中指定的选项会**直接透传**给 provider 作为模型选项。允许使用 provider 特定功能和参数（如 OpenAI reasoning 模型的 `reasoningEffort`）。

> 运行 `mimo models` 查看可用模型列表。

## 创建 agent

使用命令：`mimo agent create`

交互式命令会：

1. 询问 agent 保存位置（全局或项目特定）
2. 询问 agent 应做什么的描述
3. 生成合适的系统提示词和标识符
4. 让你选择 agent 可访问的工具
5. 最后创建带 agent 配置的 markdown 文件

## 常见用例

- **Build agent**：启用所有工具的完整开发工作
- **Plan agent**：分析和规划，不做改动
- **Review agent**：只读访问加文档工具的代码审查
- **Debug agent**：聚焦调查，启用 bash 和读取工具
- **Docs agent**：文档编写，文件操作但无系统命令

## 示例 Agents

### 文档 agent

`~/.config/mimocode/agents/docs-writer.md`：

```markdown
---
description: Writes and maintains project documentation
mode: subagent
tools:
  bash: false
---
You are a technical writer. Create clear, comprehensive documentation.

Focus on:
- Clear explanations
- Proper structure
- Code examples
- User-friendly language
```

### 安全审计 agent

`~/.config/mimocode/agents/security-auditor.md`：

```markdown
---
description: Performs security audits and identifies vulnerabilities
mode: subagent
tools:
  write: false
  edit: false
---
You are a security expert. Focus on identifying potential security issues.

Look for:
- Input validation vulnerabilities
- Authentication and authorization flaws
- Data exposure risks
- Dependency vulnerabilities
- Configuration security issues
```
