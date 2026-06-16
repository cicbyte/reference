# ZCode Skill 文档

> 来源：https://zcode.z.ai/cn/docs/skill

## 核心功能

Skill 是一组可复用的工作说明。它通常用一个 `SKILL.md` 描述触发场景、处理步骤和输出要求，让 Agent 在遇到特定任务时按同一套标准执行。

在 ZCode 中，当前建议优先围绕 ZCode Agent 沉淀技能：
- **ZCode Agent**：适合在 ZCode 内置 Agent 中使用，推荐放项目开发、文档协作、测试验证、飞书工作流等常用能力。

## 管理技能

进入 **设置 -> 技能** 后，可以按来源查看技能。列表里会展示技能名称、来源标签、描述和启用开关。

可以在这里完成几件事：
- 通过搜索框按名称筛选技能。
- 使用右侧开关启用或停用技能。
- 点击右上角 **新建技能**，ZCode 会引导你在聊天中用 Agent 生成一个新技能。

## 创建技能

一个技能本质上是一个目录加一个 `SKILL.md` 文件。目录名就是技能名，聊天中也会用这个名字引用。

ZCode Agent 的用户级技能目录：

```
~/.zcode/skills/<skill-name>/SKILL.md
```

一个最小可用的 `SKILL.md` 示例：

```markdown
---
name: code-review-checklist
description: Review code changes with a focused checklist for correctness, regressions, tests, and maintainability.
---
# Code Review Checklist

Use this skill when reviewing a pull request, merge request, or local diff.

Focus on correctness, regressions, missing tests, risky API changes, and maintainability.
```

创建或修改后，回到 **设置 -> 技能** 点击 **刷新**，确认技能出现在对应来源下，并保持开关开启。

## 从外部 Agent 导入技能

如果你已经在 Claude Code、Codex CLI、OpenClaw、Augment、Windsurf 等其他 AI 编程工具里维护了技能，不必在 ZCode 里重新创建。在 **设置 -> 技能** 页面右上角点击 **导入** 图标，ZCode 会自动扫描这些外部 Agent 的技能目录，列出可一键导入的技能。

在弹窗中可以：
- 按外部 Agent 分组查看检测到的技能，每个来源会显示对应的技能目录路径和可导入数量。
- 勾选需要的技能，或使用 **全选** 批量选择，顶部会实时显示已选数量。
- 选择 **导入方式**：
  - **软链**：创建指向外部技能目录的链接。ZCode 会跟随来源目录的后续变更，但该技能依赖来源路径持续可用。
  - **复制**：把技能复制成 ZCode 内的独立副本，与来源解耦，来源后续的修改不会再同步过来。
- 选择 **导入目标**：导入到 **全局**（用户级，所有工作区可用）或导入到当前 **项目**（仅当前工作区可用）。

确认后即可完成导入。导入的技能会出现在技能列表中，和自建技能一样可以启用、停用，并在聊天中通过 `$skill-name` 调用。

## 在聊天中使用

在输入框中输入 `$`，选择需要的技能。选中后，技能会以标签形式出现在输入框里，你可以继续补充具体需求。

比如：

```
$code-review-checklist 帮我 review 当前改动
```

```
$release-notes 根据这次提交写一版发版说明
```

ZCode 会把被引用的技能交给当前 Agent，让它按技能里的说明处理任务。

## 适合沉淀成 Skill 的内容

- 任务有固定流程，例如 code review、接口排查、发版说明、测试报告。
- 团队对输出格式有固定要求。
- 任务需要附带背景知识、检查清单、模板或示例。
- 同一个能力会在多个项目或多次对话里重复使用。

如果只是保存一句简单提示词，可以优先使用 Command；如果需要一套完整工作方法，就更适合做成 Skill。
