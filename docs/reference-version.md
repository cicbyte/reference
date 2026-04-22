# reference version

显示当前版本信息，包括版本号、Git Commit 和构建时间。

## 用法

```bash
reference version
```

## 输出示例

```
reference 0.1.0
  commit: 6cef7885
  built:  2026-04-22T00:36:35
```

## 版本信息来源

版本号通过 Go ldflags 在编译时注入：

| 字段 | ldflags 变量 | 默认值 |
|:---|:---|:---|
| Version | `cmd/version.Version` | `dev` |
| GitCommit | `cmd/version.GitCommit` | `unknown` |
| BuildTime | `cmd/version.BuildTime` | `unknown` |

使用 `go build` 直接编译时显示默认值。通过 `scripts/build.py` 或 GitHub Actions 编译时会注入实际值。

## 相关命令

- `scripts/build.py` — 本地构建脚本，自动注入版本信息
- `scripts/release.py` — 版本发布脚本，管理版本号和 Git Tag
