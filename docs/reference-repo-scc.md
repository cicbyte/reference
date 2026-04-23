# reference repo scc

查看仓库代码统计信息，基于内置的 scc 库（无需安装外部工具）。

## 用法

```bash
reference repo scc [name] [flags]
```

## 参数

| 参数 | 说明 |
|:---|:---|
| `name` | 可选，指定引用名称。不传则显示所有仓库的汇总 |

## 标志

| 标志 | 简写 | 默认值 | 说明 |
|:---|:---|:---|:---|
| `--top` | `-n` | `15` | Top 文件排名数量 |
| `--format` | `-f` | `table` | 输出格式：`table`、`json`、`jsonl` |

## 示例

```bash
# 所有仓库汇总
reference repo scc

# 指定仓库的详细统计
reference repo scc memos-cli

# 自定义 Top 文件数量
reference repo scc memos-cli -n 5

# JSON 格式
reference repo scc memos-cli -f json

# JSONL 格式
reference repo scc memos-cli -f jsonl
```

## 输出示例

**表格格式（默认）：**

```
+-----------------------+--------+--------+--------+
| 语言                  | 文件数 | 代码行 | 复杂度 |
+-----------------------+--------+--------+--------+
| Go                    |     71 |   7747 |   1457 |
| Markdown              |     15 |   1118 |      0 |
+-----------------------+--------+--------+--------+

+---+-------------+----------+--------+--------+
| # | 文件        | 语言     | 代码行 | 复杂度 |
+---+-------------+----------+--------+--------+
| 1 | tools.go    | Go       |    418 |    109 |
+---+-------------+----------+--------+--------+
```

**JSON 格式：**

```json
{
  "repo": "memos-cli",
  "languages": [
    { "languages": "Go", "files": 71, "code": 7747, "complexity": 1457 }
  ],
  "topFiles": [
    { "filename": "tools.go", "languages": "Go", "location": "internal/ai/tools.go", "code": 418, "complexity": 109 }
  ]
}
```

**JSONL 格式：**

```jsonl
{"type":"language","languages":"Go","files":71,"code":7747,"complexity":1457}
{"type":"topFiles","filename":"tools.go","languages":"Go","location":"internal/ai/tools.go","code":418,"complexity":109}
```

## 排除目录

自动排除：`.git`、`vendor`、`node_modules`、`.svn`、`.hg`

## 相关命令

- `reference repo list` — 查看所有仓库
