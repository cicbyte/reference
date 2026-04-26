# reference doctor

诊断并修复项目引用的各种问题，确保引用环境健康。

## 用法

```bash
reference doctor [flags]
```

## 标志

> `--format` / `-f` 为全局标志，详见 `reference --help`。

## 示例

```bash
# 默认表格输出
reference doctor

# JSON 格式
reference doctor -f json

# JSONL 格式（每行一条检查记录）
reference doctor -f jsonl
```

## 前提条件

必须先运行 `reference` 完成初始化配置。若未初始化，会提示：

```
  尚未初始化。请先运行 reference 完成初始化配置。
```

## 检查项

检查结果分为两个区块：

### 引用数据（所有用户）

| 检查项 | 说明 | 修复方式 |
|:---|:---|:---|
| 软链接完整性 | `.reference/repos/` 下的 Junction 是否存在 | 重建缺失的链接 |
| Wiki Junction | `.reference/wiki/` 下的 Junction 链接是否正确 | 重建缺失的链接 |
| Reference Map | `.reference/reference.map.jsonl` 是否存在且最新 | 自动重新生成 |
| 数据库一致性 | 数据库记录与实际链接是否匹配 | 标记孤立记录和未跟踪链接 |
| Wiki Git | 全局知识库 Git 状态是否正常 | 自动初始化 |

### AI Agent 配置（仅 Claude Code 用户）

| 检查项 | 说明 | 修复方式 |
|:---|:---|:---|
| Agent 文件 | `.claude/agents/` 下的 agent 配置是否最新 | 从源模板强制覆盖 |
| SKILL.md | `.claude/skills/reference/SKILL.md` 是否存在 | 从模板重新渲染 |

## 输出示例

### 表格格式（默认）

```
+------+---------------+---------------------------------------------------------------------+
| 状态 | 检查项        | 详情                                                                |
+------+---------------+---------------------------------------------------------------------+
|  OK  | 软链接完整性  | 1/1 正常                                                            |
|  OK  | Wiki Junction | 1/1 正常                                                            |
|  OK  | Reference Map | 正常                                                                |
|  OK  | 数据库一致性  | 正常                                                                |
|  OK  | Wiki Git      | 远程: xxx，工作区干净                                               |
+------+---------------+---------------------------------------------------------------------+
|  OK  | Agent 文件    | 正常                                                                |
|  OK  | SKILL.md      | 正常                                                                |
+------+---------------+---------------------------------------------------------------------+
|      | 结果          | 一切正常，无需修复                                                  |
+------+---------------+---------------------------------------------------------------------+
```

### JSON 格式

```json
{
  "project_dir": "D:\\code\\cicbyte\\reference",
  "checks": [
    { "name": "软链接完整性", "status": "ok", "details": "1/1 正常", "group": "core" },
    { "name": "Agent 文件", "status": "ok", "details": "正常", "group": "agent" }
  ],
  "summary": "一切正常，无需修复"
}
```

## 状态值

| 值 | 含义 |
|:---|:---|
| `ok` | 正常 |
| `fixed` | 已修复 |
| `warn` | 警告（需要手动处理） |

## 与 `reference`（默认行为）的区别

| | `reference` | `reference doctor` |
|:---|:---|:---|
| 修复软链接 | 静默修复 | 检查并报告 |
| Agent 文件 | 强制覆盖 | 强制覆盖 |
| SKILL.md | 重新渲染 | 重新渲染 |
| Wiki Junction | 重建 | 检查并修复 |
| 数据库一致性 | 不检查 | 检查并报告 |
| Reference Map | 生成 | 检查并修复 |

## 相关命令

- `reference` — 默认行为（快速注入）
