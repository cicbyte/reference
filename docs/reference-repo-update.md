# reference repo update

更新远程仓库的全局缓存（执行 git pull）。

## 用法

```bash
reference repo update [identifier]
```

## 参数

| 参数 | 说明 |
|:---|:---|
| `identifier` | 可选，指定链接名称。不传则更新所有远程引用 |

## 示例

```bash
# 更新所有远程引用
reference repo update

# 更新指定引用
reference repo update github.com-gin-gonic-gin
```

## 行为说明

- 仅对远程仓库执行 `git pull`
- 本地仓库会被自动跳过（不需要更新）
- 更新完成后自动刷新数据库中的 commit、分支等信息
- 若缓存目录不存在（被手动删除），会重新克隆

## 相关命令

- `reference repo add --no-update` — 添加时跳过更新
- `reference repo list` — 查看当前 commit 信息
