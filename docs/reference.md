# reference

无参数运行时的默认行为：初始化配置、修复链接、注入 Agent 配置。

## 用法

```bash
reference
```

## 首次运行

首次运行时会进入交互式引导，支持多选：

```
  欢迎使用 reference！

  请选择你的编程助手（可多选，用逗号分隔）：
    [1] Claude Code
    [2] Codex
    [3] OpenCode
    [4] ZCode
    [5] MiMo Code
    [6] 无（仅使用仓库引用管理功能）

  请输入选项:
```

可输入 `1,2` 同时选择多个助手。配置保存到 `.reference/reference.settings.json`，后续运行不再引导。

CI 集成可使用非交互式 `reference init --agent claude,codex`，详见 [reference init](./reference-init.md)。

## 执行内容

1. **修复软链接** — 检测 `.reference/repos/` 下的 Junction/Symlink，若被手动删除则静默重建
2. **生成 reference.map.jsonl** — 将仓库列表写入 `.reference/reference.map.jsonl`，供 AI Agent 读取
3. **创建 Wiki Junction** — 为每个仓库创建 Junction 链接到 `.reference/wiki/`（本地仓库链接到 localwiki）
4. **注入 AI 配置**（启用 AI 助手时）— 为每个已选助手注入子代理 + Skill

## 前提条件

无。reference 的核心功能（仓库引用管理）不依赖任何 AI 工具。

## 输出示例

```
  已配置: Claude Code, Codex
  已链接 3 个仓库知识，已更新 6 个 AI 配置文件。
```

```
  已配置: 无
  已链接 1 个仓库知识。
```

## 相关命令

- `reference doctor` — 诊断并修复引用健康状态（包含更多检查项）
- `reference global` — 全局引用管理（跨项目列表、GC、统计）
- `reference wiki` — 知识库 Git 管理（commit、sync、trash、restore）
- `reference wiki --local` — 本地知识库管理
