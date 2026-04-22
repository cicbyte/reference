# reference repo remove

移除当前项目中的某个仓库引用。

## 用法

```bash
reference repo remove <identifier> [flags]
```

## 参数

| 参数 | 说明 |
|:---|:---|
| `<identifier>` | 链接名称（如 `github.com-gin-gonic-gin`）或数据库 ID |

## 标志

| 标志 | 简写 | 说明 |
|:---|:---|:---|
| `--purge` | `-p` | 同时删除全局缓存仓库（仅远程模式，需二次确认） |

## 示例

```bash
# 仅移除项目引用（软链接）
reference repo remove github.com-gin-gonic-gin

# 同时删除全局缓存
reference repo remove github.com-gin-gonic-gin --purge
```

## 行为说明

**默认模式**（不加 `--purge`）：
- 删除 `.reference/<链接名>` 软链接
- 删除数据库索引记录
- 全局缓存 `~/.cicbyte/reference/repos/...` 保留，其他项目仍可使用

**清除模式**（`--purge`）：
- 在默认模式基础上，额外删除全局缓存目录
- 仅对远程仓库有效
- 对本地仓库的 `--purge` 会被忽略并提示（保护用户本地代码）
- 需要二次确认（交互式 y/n）

## 相关命令

- `reference repo list` — 查看链接名称
- `reference repo add` — 重新添加引用
