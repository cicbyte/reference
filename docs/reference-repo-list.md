# reference repo list

列出当前项目所有引用的仓库。

## 用法

```bash
reference repo list [flags]
```

## 标志

| 标志 | 简写 | 说明 |
|:---|:---|:---|
| `--format` | `-f` | 输出格式：`table`（默认）、`json`、`jsonl` |

## 示例

```bash
# 默认表格输出
reference repo list

# JSON 格式
reference repo list -f json

# JSONL 格式（每行一条记录）
reference repo list -f jsonl
```

## 输出示例

**表格格式（默认）：**

```
+------+-----------+--------------------------------------+---------------------------------------------------------------------+------------+--------+
| 类型 | 名称      | 来源                                 | 源路径                                                              | 更新时间   | 分支   |
+------+-----------+--------------------------------------+---------------------------------------------------------------------+------------+--------+
| 远程 | memos-cli | https://github.com/cicbyte/memos-cli | C:\Users\zhyj\.cicbyte\reference\repos\github.com\cicbyte\memos-cli | 2026-04-22 | master |
+------+-----------+--------------------------------------+---------------------------------------------------------------------+------------+--------+
```

**JSON 格式：**

```json
{
  "repos": [
    {
      "type": "remote",
      "name": "memos-cli",
      "source": "https://github.com/cicbyte/memos-cli",
      "cache_path": "C:\\Users\\zhyj\\.cicbyte\\reference\\repos\\github.com\\cicbyte\\memos-cli",
      "commit_at": "2026-04-22",
      "branch": "master"
    }
  ]
}
```

## 输出字段

| 字段 | 说明 |
|:---|:---|
| type | 引用类型（remote / local） |
| name | 引用名 |
| source | 远程 URL 或本地路径 |
| cache_path | Junction 实际指向的本地缓存目录 |
| commit_at | 最新 commit 日期 |
| branch | 当前分支 |

## 相关命令

- `reference repo add` — 添加引用
- `reference repo remove` — 移除引用
