# MiMo Code Skills 文档

> 来源：https://mimo.xiaomi.com/zh/mimocode/skills

## 核心功能

Agent skills 让 MiMo Code 从你的仓库或家目录中发现可复用的指令。Skills 通过原生 `skill` 工具按需加载——agents 能看到可用 skills，并在需要时加载完整内容。

## 文件放置

在一个文件夹中放一个 `SKILL.md`。Skill 的名字来自 frontmatter（不是文件夹名），`SKILL.md` 文件会被递归发现，可以任意嵌套。

MiMo Code 搜索这些位置：

- **项目配置**：`.mimocode/skills/**/SKILL.md`（单数 `skill/` 也可）
- **全局配置**：`~/.config/mimocode/skills/**/SKILL.md`
- **项目兼容目录**：`.claude`、`.agents`、`.codex`、`.opencode` —— 每个都扫描 `skills/**/SKILL.md`
- **全局兼容目录**：家目录下同样的四个文件夹（如 `~/.claude/skills/**/SKILL.md`）

通过 `skills.paths` 和 `skills.urls` 配置项可添加更多来源。

## 发现机制

对于项目本地路径，MiMo Code 从当前工作目录向上遍历直到工作区根目录。沿途加载 `.mimocode/` 下所有匹配的 `skills/**/SKILL.md`（或 `skill/**/SKILL.md`），以及 `.claude`、`.agents`、`.codex`、`.opencode` 目录中的 `skills/**/SKILL.md`。

全局定义也会从 `~/.config/mimocode/` 和家目录下相同的兼容文件夹加载。

如果两个 skills 解析到同一个 `name`，**后加载的生效**，并记录警告。

## Frontmatter 字段

每个 `SKILL.md` 必须以 YAML frontmatter 开头。只读取以下字段：

- `name`（必需）
- `description`（必需）
- `hidden`（可选布尔值——为 `true` 时，skill 被加载但不进入可用 skills 列表）

其他 frontmatter 字段会被忽略。

## 命名规范

MiMo Code 不强制 `name` 格式或 `description` 长度，但 `name` 是 agent 引用 skill 的方式，建议简短且可预测。推荐小写字母数字加单个连字符，如 `git-release`。

`description` 要足够具体，让 agent 能正确选择 skill——它是 agent 在加载完整内容前唯一看到的信号。

## 示例

创建 `.mimocode/skills/git-release/SKILL.md`：

```markdown
---
name: git-release
description: Create consistent releases and changelogs
---

## What I do
- Draft release notes from merged PRs
- Propose a version bump
- Provide a copy-pasteable `gh release create` command

## When to use me
Use this when you are preparing a tagged release.
Ask clarifying questions if the target versioning scheme is unclear.
```

## 工具描述识别

MiMo Code 在 `skill` 工具描述中列出可用 skills。每个条目包含 skill 名和描述：

```
git-release
Create consistent releases and changelogs
```

Agent 通过调用工具加载 skill：

```
skill({ name: "git-release" })
```

## 权限配置

在 `mimocode.json` 中使用基于模式的权限控制 agents 能访问哪些 skills：

```json
{
  "permission": {
    "skill": {
      "*": "allow",
      "pr-review": "allow",
      "internal-*": "deny",
      "experimental-*": "ask"
    }
  }
}
```

| 权限 | 行为 |
| --- | --- |
| `allow` | 立即加载 skill |
| `deny` | 对 agent 隐藏，拒绝访问 |
| `ask` | 加载前提示用户确认 |

模式支持通配符：`internal-*` 匹配 `internal-docs`、`internal-tools` 等。

## 按 agent 覆盖

给特定 agent 不同于全局默认的权限。

**自定义 agent**（在 agent frontmatter）：

```yaml
---
permission:
  skill:
    "documents-*": "allow"
---
```

**内置 agent**（在 `mimocode.json`）：

```json
{
  "agent": {
    "plan": {
      "permission": {
        "skill": {
          "internal-*": "allow"
        }
      }
    }
  }
}
```

## 禁用 skill 工具

完全禁用某些 agent 的 skills：

**自定义 agent**：

```yaml
---
tools:
  skill: false
---
```

**内置 agent**：

```json
{
  "agent": {
    "plan": {
      "tools": {
        "skill": false
      }
    }
  }
}
```

禁用后，`<available-skills>` 部分被完全省略。

## 排查加载问题

如果 skill 没显示：

1. 确认 `SKILL.md` 全大写拼写
2. 检查 frontmatter 包含 `name` 和 `description`
3. 确保 skill 名字唯一——重复的 `name` 会被静默覆盖（后加载的生效）
4. 检查权限——`deny` 的 skills 对 agents 隐藏
