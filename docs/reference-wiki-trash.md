# reference wiki trash

查看被删除的知识库文件，支持从 Git 历史恢复。

## 用法

```bash
reference wiki trash [flags]
```

## 标志

| 标志 | 简写 | 说明 |
|:---|:---|:---|
| `--limit` | `-n` | 显示最近 N 条（默认 20） |
| `--format` | `-f` | 输出格式：`table`（默认）、`json`、`jsonl` |

## 示例

```bash
# 默认表格输出
reference wiki trash

# 显示最近 50 条
reference wiki trash -n 50

# JSON 格式
reference wiki trash -f json

# JSONL 格式
reference wiki trash -f jsonl
```

## 输出示例

**表格格式（默认）：**

```
+---------+------------+----------------------------------------+
| 提交    | 日期       | 路径                                   |
+---------+------------+----------------------------------------+
| ab03992 | 2026-04-22 | github/boyter/scc/scc.md               |
| aaa6d6a | 2026-04-22 | github.com-spf13-cobra/reference.md    |
+---------+------------+----------------------------------------+

共 2 条记录，使用 reference wiki restore <path> 恢复
```

**JSON 格式：**

```json
[
  { "path": "github/boyter/scc/scc.md", "commit": "ab03992", "date": "2026-04-22 18:00:50 +0800", "message": "wiki: auto-update from reference inject" }
]
```

## 恢复文件

```bash
reference wiki restore <path>
```

## 相关命令

- `reference wiki restore` — 从 Git 历史恢复文件
- `reference wiki commit` — 提交知识库更改
