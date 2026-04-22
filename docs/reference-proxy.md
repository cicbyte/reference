# reference proxy

查看、设置或清除网络代理配置，用于加速 Git 克隆操作。

## 用法

```bash
reference proxy [command]
```

## 子命令

### `reference proxy set`

设置代理地址。

```bash
reference proxy set <url|port> [flags]
```

支持多种输入格式：

```bash
# 纯端口号（自动补全为 http://127.0.0.1:<port>）
reference proxy set 7890

# 完整 URL
reference proxy set http://127.0.0.1:7890
reference proxy set socks5://127.0.0.1:1080

# 使用预设
reference proxy set --preset clash
reference proxy set --preset v2ray
```

| 标志 | 说明 |
|:---|:---|
| `--preset` | 使用预设配置 |
| `--git` | 设置为 Git 专用代理（git_proxy，优先级高于 proxy） |

#### 预设列表

| 预设 | 地址 | 说明 |
|:---|:---|:---|
| `clash` | `http://127.0.0.1:7890` | Clash for Windows |
| `v2ray` | `http://127.0.0.1:1080` | V2Ray |
| `ss` | `socks5://127.0.0.1:1080` | Shadowsocks |
| `surge` | `http://127.0.0.1:6152` | Surge |
| `qv2ray` | `http://127.0.0.1:8080` | Qv2ray |
| `ssrdog` | `http://127.0.0.1:7897` | SSRDog |

### `reference proxy info`

查看当前代理配置。

```bash
reference proxy info
```

### `reference proxy clear`

清除代理配置。

```bash
reference proxy clear
```

## 代理优先级

```
git_proxy > proxy > 环境变量 HTTPS_PROXY > 无代理
```

- `proxy`：通用 HTTP/SOCKS5 代理
- `git_proxy`：Git 专用代理，优先级更高

## 配置存储

代理配置保存在 `~/.cicbyte/reference/config/config.yaml`：

```yaml
network:
  proxy: http://127.0.0.1:7890
  git_proxy: ""
```

## 相关命令

- `reference repo add` — 添加仓库时自动使用代理
- `reference repo update` — 更新仓库时自动使用代理
