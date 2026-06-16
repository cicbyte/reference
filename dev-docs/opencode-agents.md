# OpenCode Agents 文档

> 来源：https://opencode.ai/docs/zh-cn/agents/

## 核心功能

代理是专门的 AI 助手，可以针对特定任务和工作流程进行配置。它们允许您创建具有自定义提示词、模型和工具访问权限的专用工具。

可以在会话期间切换代理，或使用 `@` 提及来调用它们。

## 类型

OpenCode 中有两种类型的代理：主代理和子代理。

### 主代理

主代理是您直接交互的主要助手。可以使用 **Tab** 键或配置的 `switch_agent` 快捷键来循环切换它们。这些代理处理主要对话。工具访问通过权限进行配置——例如，Build 启用了所有工具，而 Plan 则受到限制。

OpenCode 内置了两个主代理：**Build** 和 **Plan**。

### 子代理

子代理是主代理可以调用来执行特定任务的专业助手。也可以通过在消息中 **@ 提及**它们来手动调用。

OpenCode 内置了三个子代理：**General**、**Explore** 和 **Scout**。

## 内置代理

OpenCode 内置了两个主代理和三个子代理。

### Build（主代理）

- **模式**：`primary`
- Build 是启用了所有工具的**默认**主代理。这是用于需要完全访问文件操作和系统命令的开发工作的标准代理。

### Plan（主代理）

- **模式**：`primary`
- 一个专为规划和分析设计的受限代理。使用权限系统来提供更多控制权，并防止意外更改。默认情况下，以下所有项均设置为 `ask`：
  - `file edits`：所有写入、补丁和编辑
  - `bash`：所有 bash 命令
- 当希望 LLM 分析代码、建议更改或创建计划，而不对代码库进行任何实际修改时，此代理非常有用。

### General（子代理）

- **模式**：`subagent`
- 一个用于研究复杂问题和执行多步骤任务的通用代理。拥有完整的工具访问权限（todo 除外），因此可以在需要时修改文件。可用于并行运行多个工作单元。

### Explore（子代理）

- **模式**：`subagent`
- 一个用于探索代码库的快速只读代理。无法修改文件。当需要按模式快速查找文件、搜索代码中的关键字或回答有关代码库的问题时使用。

### Scout（子代理）

- **模式**：`subagent`
- 一个用于外部文档和依赖研究的只读代理。当需要将某个依赖仓库克隆到 OpenCode 的托管缓存中、检查库的源代码，或在不修改工作区的情况下将本地代码与 upstream 实现进行交叉对照时使用。

### Compaction / Title / Summary（系统主代理）

- **模式**：`primary`
- 隐藏的系统代理，分别用于将长上下文压缩为较小的摘要、生成简短的会话标题、创建会话摘要。它们会在需要时自动运行，且无法在 UI 中选择。

## 用法

1. **主代理**：在会话期间使用 **Tab** 键循环切换。也可以使用配置的 `switch_agent` 快捷键。
2. **子代理**可以通过以下方式调用：
   - 由主代理根据其描述**自动**调用以执行专门任务。
   - 通过在消息中 **@ 提及**子代理来手动调用。例如：
     ```
     @general help me search for this function
     ```
3. **会话间导航**：当子代理创建自己的子会话时，可以使用以下方式在父会话和所有子会话之间导航：
   - **<Leader>+Right**（或配置的 `session_child_cycle` 快捷键）向前循环：父会话 → 子会话1 → 子会话2 → … → 父会话
   - **<Leader>+Left**（或配置的 `session_child_cycle_reverse` 快捷键）向后循环：父会话 ← 子会话1 ← 子会话2 ← … ← 父会话

## 配置

可以自定义内置代理或通过配置创建自己的代理。两种方式：

### JSON 配置

在 `opencode.json` 配置文件中配置代理：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "agent": {
    "build": {
      "mode": "primary",
      "model": "anthropic/claude-sonnet-4-20250514",
      "prompt": "{file:./prompts/build.txt}",
      "tools": {
        "write": true,
        "edit": true,
        "bash": true
      }
    },
    "plan": {
      "mode": "primary",
      "model": "anthropic/claude-haiku-4-20250514",
      "tools": {
        "write": false,
        "edit": false,
        "bash": false
      }
    },
    "code-reviewer": {
      "description": "Reviews code for best practices and potential issues",
      "mode": "subagent",
      "model": "anthropic/claude-sonnet-4-20250514",
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

还可以使用 Markdown 文件定义代理。放置位置：

- **全局**：`~/.config/opencode/agents/`
- **项目级**：`.opencode/agents/`

```markdown
---
description: Reviews code for quality and best practices
mode: subagent
model: anthropic/claude-sonnet-4-20250514
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

Markdown 文件名即为代理名称。例如，`review.md` 会创建一个名为 `review` 的代理。

## 配置选项详解

### description（描述，必需）

提供代理的功能及使用场景的简要描述。

### temperature（温度）

控制 LLM 响应的随机性和创造力。较低的值使响应更加集中和确定，较高的值增加创造力和多样性。值范围通常 0.0-1.0：

- **0.0-0.2**：非常集中和确定性的响应，适合代码分析和规划
- **0.3-0.5**：平衡的响应，兼顾一定创造力，适合一般开发任务
- **0.6-1.0**：更有创造力和多样性的响应，适合头脑风暴和探索

未指定温度时，OpenCode 使用模型特定的默认值；大多数模型通常为 0，Qwen 模型为 0.55。

### steps（最大步数）

控制代理在被强制以纯文本响应之前可以执行的最大代理迭代次数。允许希望控制成本的用户对代理操作设置限制。

未设置时，代理将持续迭代，直到模型选择停止或用户中断会话。

达到限制时，代理会收到一个特殊的系统提示词，指示其回复工作摘要和建议的剩余任务。

### disable（禁用）

设置为 `true` 以禁用代理。

### prompt（提示词文件）

为代理指定自定义系统提示词文件：

```json
{
  "agent": {
    "review": {
      "prompt": "{file:./prompts/code-review.txt}"
    }
  }
}
```

路径相对于配置文件所在位置。同时适用于全局 OpenCode 配置和项目级配置。

### model（模型）

为代理覆盖模型。模型 ID 使用 `provider/model-id` 格式。例如使用 OpenCode Zen，可以用 `opencode/gpt-5.1-codex` 表示 GPT 5.1 Codex。

### tools（工具）

控制代理中可用的工具。可以通过将特定工具设置为 `true` 或 `false` 来启用或禁用它们。

支持通配符同时控制多个工具（如 `mymcp_*: false` 禁用 MCP 服务器中所有工具）。代理特定配置覆盖全局配置。

### permission（权限）

配置权限管理代理可以执行的操作。目前 `edit`、`bash` 和 `webfetch` 工具的权限可配置为：

- `"ask"` — 运行工具前提示审批
- `"allow"` — 允许所有操作，无需审批
- `"deny"` — 禁用该工具

可按代理覆盖，也可以为特定 bash 命令设置权限（支持 glob 模式）。最后匹配的规则优先，所以把 `*` 通配符放前面，具体规则放后面。

### mode（模式）

控制代理的模式。可设为 `primary`、`subagent` 或 `all`。未指定默认为 `all`。

### hidden（隐藏）

`hidden: true` 将子代理从 `@` 自动补全菜单中隐藏。适用于只应由其他代理通过 Task 工具以编程方式调用的内部子代理。仅影响自动补全菜单中的用户可见性。

### task permissions（任务权限）

通过 `permission.task` 控制代理可以通过 Task 工具调用哪些子代理。使用 glob 模式灵活匹配。设置为 `deny` 时，子代理将从 Task 工具描述中完全移除。

### color（颜色）

自定义代理在 UI 中的视觉外观。使用有效十六进制颜色（如 `#FF5733`）或主题颜色：`primary`、`secondary`、`accent`、`success`、`warning`、`error`、`info`。

### top_p

控制响应多样性，是温度的替代方案。值范围 0.0-1.0，较低值更集中，较高值更多样化。

### 其他选项（透传）

在代理配置中指定的其他选项都会作为模型选项**直接传递**给提供商。允许使用提供商特定的功能和参数（如 OpenAI 推理模型的 `reasoningEffort`）。

## 创建代理

使用交互式命令创建新代理：

1. 询问代理的保存位置——全局或项目级
2. 描述代理应该做什么
3. 生成合适的系统提示词和标识符
4. 选择代理可以访问哪些工具
5. 创建包含代理配置的 Markdown 文件

## 常见使用场景

- **Build 代理**：启用所有工具的完整开发工作
- **Plan 代理**：分析和规划，不进行任何更改
- **Review 代理**：具有只读访问权限和文档工具的代码审查
- **Debug 代理**：专注于问题排查，启用 bash 和读取工具
- **Docs 代理**：文档编写，具有文件操作但不使用系统命令

## 示例代理

### 文档代理

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

### 安全审计代理

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
