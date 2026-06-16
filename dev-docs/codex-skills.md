# Codex Skills 文档

> 来源：https://www.codex-docs.com/configuration/skills/
> 官方文档：https://developers.openai.com/codex/skills

## 核心功能

使用智能体技能可以给 Codex 注入任务级能力。一个技能会把指令、参考资料和可选脚本打包在一起，让 Codex 可以更稳定地遵循某个工作流。Codex 的技能建立在**开放智能体技能标准**之上。

- **技能**：可复用工作流的创作格式
- **插件**：可安装的分发单元

先设计工作流用技能；要让其他开发者能安装使用，或希望和应用一起分发，再打包成插件。

Codex CLI、IDE 扩展和 Codex App 都支持技能。

## 上下文管理（按需展开）

Codex 使用"按需展开"管理上下文：
- **启动时**：只读取每个技能的 name、description 和文件路径
- **使用时**：Codex 决定使用某技能后，才加载完整 `SKILL.md` 指令

Codex 会在上下文里放一份初始技能列表。为避免挤占提示词其它部分，这份列表最多约占模型上下文窗口的 **2%**；上下文窗口未知时上限为 **8,000 字符**。

如果安装了很多技能，Codex 会先缩短技能描述；特别大的技能集合，部分技能可能不会出现在初始列表中，Codex 会显示警告。

> 这个预算只适用于初始技能列表。Codex 选中某技能后，仍会读取该技能完整的 `SKILL.md` 指令。

## 技能目录结构

一个技能就是一个目录，至少要有 `SKILL.md`，还可以按需添加脚本和参考资料：

```
my-skill/
├── SKILL.md              # 必需：指令与元数据
├── scripts/              # 可选：可执行脚本
├── references/           # 可选：文档参考
├── assets/               # 可选：模板与资源
└── agents/
    └── openai.yaml       # 可选：展示信息与依赖声明
```

`SKILL.md` 必须声明 `name` 和 `description`：

```markdown
---
name: skill-name
description: Explain exactly when this skill should and should not trigger.
---
Skill instructions for Codex to follow.
```

## 触发方式

Codex 有两种触发技能的方式：

1. **显式调用**：CLI / 应用中使用 `/skills` 选择，或在提示词中写 `$skill-name`
2. **隐式调用**：Codex 根据 description 判断是否适合使用

因为隐式匹配依赖 description，所以技能描述要简洁，并写清楚触发范围和边界。把关键用例和触发词放在前面，即使描述被缩短，Codex 仍能匹配。

## 技能保存位置

Codex 会从仓库、用户、管理员和系统四类位置读取技能。对于仓库级技能，Codex 从当前工作目录开始，一路向上扫描到仓库根目录中的 `.agents/skills`。

> 如果两个技能使用了同一个 name，Codex 不会合并它们，它们会分别出现在可选技能列表中。

| 技能范围 | 位置 | 推荐用途 |
| --- | --- | --- |
| REPO | `$CWD/.agents/skills` | 当前工作目录对应的小范围技能（微服务或模块专属） |
| REPO | `$CWD/../.agents/skills` | Git 仓库子目录启动时父目录共享技能自动生效 |
| REPO | `$REPO_ROOT/.agents/skills` | 仓库级共享技能，团队规范或工作流 |
| USER | `$HOME/.agents/skills` | 用户全局技能，适用于任意仓库 |
| ADMIN | `/etc/codex/skills` | 机器或容器级共享，运维脚本、SDK 自动化、管理员下发 |
| SYSTEM | OpenAI 随 Codex 内置 | 通用技能，如 `skill-creator` 和规划相关 |

Codex 支持通过符号链接组织技能目录；扫描时会跟随链接目标继续读取。

> 这些路径主要用于本地创作和本地发现。要把技能作为可复用产物分发到仓库之外，应使用插件。

## 创建技能

优先使用内置创建器：`$skill-creator`

创建器会询问：
1. 这个技能做什么
2. 在什么场景下触发
3. 是"纯指令"还是"带脚本"型（默认推荐先做纯指令技能）

也可以手动创建，建立目录并放入 `SKILL.md`。Codex 会自动检测技能变更；如果更新后没生效，重启 Codex 即可。

## 在本地安装精选技能

使用 `$skill-installer` 在本地 Codex 环境中安装内置之外的精选技能：

```
$skill-installer linear
```

也可以让安装器从其他仓库下载技能。Codex 会自动发现新安装的技能；如果没有立即出现，重启 Codex。

> 这个流程更适合本地试验和个人使用。对于自己希望复用和分发的技能，优先考虑插件。

## 启用/停用技能

通过 `~/.codex/config.toml` 中的 `[[skills.config]]` 条目，不删除技能即可停用：

```toml
[[skills.config]]
path = "/path/to/skill/SKILL.md"
enabled = false
```

修改后需重启 Codex。

## 界面元数据与依赖声明

在技能目录中加入 `agents/openai.yaml`，可在 Codex App 中配置界面元数据、调用策略和工具依赖：

```yaml
interface:
  display_name: "Optional user-facing name"
  short_description: "Optional user-facing description"
  icon_small: "./assets/small-logo.svg"
  icon_large: "./assets/large-logo.png"
  brand_color: "#3B82F6"
  default_prompt: "Optional surrounding prompt to use the skill with"

policy:
  allow_implicit_invocation: false

dependencies:
  tools:
    - type: "mcp"
      value: "openaiDeveloperDocs"
      description: "OpenAI Docs MCP server"
      transport: "streamable_http"
      url: "https://developers.openai.com/mcp"
```

> `allow_implicit_invocation` 默认为 `true`。设为 `false` 后，Codex 不会根据用户提示词自动隐式触发该技能，但显式使用 `$skill-name` 仍然有效。

## 最佳实践

- 让每个技能聚焦在一项明确工作上
- 除非确实需要确定性行为或外部工具，否则优先使用指令而不是脚本
- 步骤尽量写成祈使句，明确输入与输出
- 用真实提示词测试 `description`，确认它会在正确场景触发，也会在不该触发时保持沉默

## 与 Claude Code 技能的关系

OpenAI 采用了 Anthropic 的开放智能体技能格式，**已有的 Claude 技能可以开箱即用**。技能目录约定对应关系：

| 工具 | 项目级 skills 路径 |
|:---|:---|
| Claude Code | `.claude/skills/<name>/SKILL.md` |
| Codex | `.agents/skills/<name>/SKILL.md`（推荐，跨工具兼容） |

Codex 也读取 `.codex/skills/` 作为项目级路径，但 `.agents/skills/` 是开放标准，跨工具通用性更好。

## 参考资源

- [Agent Skills – OpenAI 官方](https://developers.openai.com/codex/skills)
- [VoltAgent/awesome-agent-skills](https://github.com/VoltAgent/awesome-agent-skills) — 1000+ 跨工具技能集合
- [AgentSkills.io](https://agentskills.io/home) — 开放技能格式规范
