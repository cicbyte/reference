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

| 标志 | 简写 | 说明 |
|:---|:---|:---|
| `--top` | `-t` | 显示 Top 文件排名（按代码行数） |
| `--format` | `-f` | 输出格式：`table`（默认）、`json`、`jsonl` |

## 示例

```bash
# 所有仓库汇总
reference repo scc

# 指定仓库的详细统计
reference repo scc memos-cli

# Top 文件排名
reference repo scc memos-cli --top

# JSON 格式
reference repo scc memos-cli -f json

# JSONL 格式（每行一条语言记录）
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
```

**JSON 格式：**

```json
{
  "repo": "memos-cli",
  "path": "C:\\Users\\zhyj\\.cicbyte\\reference\\repos\\github.com\\cicbyte\\memos-cli",
  "languages": [
    { "language": "Go", "files": 71, "lines": 9797, "code": 7747, "comments": 631, "blanks": 1419, "complexity": 1457 }
  ],
  "files": []
}
```

## 排除目录

自动排除：`.git`、`vendor`、`node_modules`、`.svn`、`.hg`

## 相关命令

- `reference repo list` — 查看所有仓库
