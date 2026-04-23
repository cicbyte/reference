# reference

无参数运行时的默认行为：初始化配置、修复链接、注入 Agent 配置。

## 用法

```bash
reference
```

## 首次运行

首次运行时会进入交互式引导：

```
  欢迎使用 reference！

  请选择你的编程助手：
    [1] Claude Code
    [2] 无（仅使用仓库引用管理功能）

  请输入选项 (1/2):
```

选择后配置保存到 `.reference/reference.settings.json`，后续运行不再引导。

## 执行内容

1. **修复软链接** — 检测 `.reference/repos/` 下的 Junction/Symlink，若被手动删除则静默重建
2. **生成 reference.map.json** — 将仓库列表写入 `.reference/reference.map.json`，供 AI Agent 读取
3. **创建 Wiki Junction** — 为每个仓库创建 Junction 链接到 `.reference/wiki/`
4. **注入 Agent 配置**（仅 Claude Code 用户）— 复制 Agent 文件和 SKILL.md 到 `.claude/`

## 前提条件

无。reference 的核心功能（仓库引用管理）不依赖任何 AI 工具。

## 输出示例

```
  已配置: Claude Code
  已链接 2 个仓库知识，已修复 1 个引用链接。
```

```
  已配置: 无
  已链接 1 个仓库知识。
```

## 相关命令

- `reference doctor` — 诊断并修复引用健康状态（包含更多检查项）
- `reference global` — 全局引用管理（跨项目列表、GC、统计）
- `reference wiki` — 知识库 Git 管理（commit、sync、watch、trash、restore）
- `reference repo inject` — 已合并到默认行为中，无需单独调用
