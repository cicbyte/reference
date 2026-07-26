# 全局引用管理后台 — reference CLI 改造需求

> 目标：让 reference CLI 支持「跨项目全局管理」语义，配合 git-mate 构建一个**全局引用管理后台**，类似管理控制台：在一个界面看到所有项目的初始化状态、健康问题、注入的 agent，并能跨项目执行管理操作（诊断、移除、更新、初始化），无需逐个 `cd` 切换。
>
> 本文档描述 reference CLI 侧需要补充的能力；git-mate 侧的对接在能力落地后只需调用新的/增强的命令。

## 背景与动机

### 现状

reference 当前以「**项目为中心**」运作 —— 所有写操作（`repo add/remove/update`、`init`、`doctor`、`proxy set`）都作用于 **当前工作目录** 所在的项目，没有 `-C <path>` / `--project <path>` 之类的全局定位 flag。唯一的例外是 `global remove`，它已经支持 `--project <path>`，是当前**唯一**的跨项目写操作。

读侧的 `global list` / `global stats` 能跨项目汇总，但**只返回最基本的元信息**（`project_dir / exists / repo_count / repos`），不包含：

- 项目的初始化状态（`initialized`）
- 项目注入的 agents 列表
- 项目的健康/链接状态（哪些项目软链接断了、Wiki junction 失效、缓存缺失）

健康检查只能逐项目跑 `reference doctor`（仅当前项目），想做一次全局体检只能 `cd` 进每个目录轮询 —— 62 个项目意味着 62 次 spawn，慢且不可用。

### 问题

在 git-mate 这类 GUI 工具里想做「全局引用管理后台」，会撞到三堵墙：

1. **看不到全局健康** — 没有命令一次性返回「哪些项目有问题」。后台管理的灵魂就是「一眼看到异常」，这个能力缺失
2. **看不到初始化状态** — `global list` 不暴露 agent 信息，「哪些项目初始化了、注入了什么」这个核心诉求无法满足
3. **跨项目写操作不完整** — 只有 `global remove --project`，`init / update / proxy set` 都没有跨项目能力

### 直读 SQLite 不是解法

DB 路径 `~/.cicbyte/apps/reference/db/app.db`（GORM 管理的单表 `repos`）虽然 schema 简单，但**写操作绕过 CLI 极其危险**：

- 删一个引用需要同步删除 `.reference/repos/<name>` 链接、`.reference/wiki/<name>` junction、`reference.map.jsonl`、缓存引用计数 —— 这些副作用全封装在 CLI 内
- GORM 自动迁移，schema 随版本演进，直读直写会随 reference 升级而失效
- `exists / initialized / agents / healthy` 这些字段根本不在 DB 里，需要文件系统 + JSON 综合判断

**结论：git-mate 永远走 exe，不直读 DB。reference 需要补齐跨项目语义。**

---

## 改造范围

### 总原则

- **读优先于写**：先补「全局可见性」（global doctor / global list 增强），后台管理就能立刻跑起来
- **复用现有逻辑**：所有新增命令复用已有的检查函数、项目遍历逻辑，避免重复实现
- **保持向后兼容**：现有命令的默认行为不变，新能力通过新命令或新 flag 解锁
- **JSON 优先**：所有新命令必须支持 `-f json`，GUI 消费结构化数据

### 优先级分级

| 优先级 | 能力 | 解锁的 GUI 场景 |
|--------|------|----------------|
| **P0** | `global doctor` — 批量跨项目诊断 | 后台管理核心表格（项目列表 + 健康状态）|
| **P0** | `global list` 增强 — 补 agent / initialized / health 字段 | 「哪些项目初始化了」「注入了什么 agent」|
| **P1** | `init / update / proxy set` 加 `--project` flag | 跨项目初始化、更新、改代理 |
| **P2** | `global doctor --fix` — 自动修复简单问题 | 一键修复断裂软链接等 |

---

## P0 改造详述

### 1. `global doctor` — 批量跨项目诊断（最关键）

#### 命令形态

```bash
# 诊断所有项目（默认）
reference global doctor

# 诊断指定项目
reference global doctor --project <path>

# JSON 输出（GUI 消费）
reference global doctor -f json

# 只返回有问题的项目（用于告警视图）
reference global doctor --issues-only

# 可选：并发度控制
reference global doctor --concurrency 8
```

#### 输出 schema（JSON）

```jsonc
{
  "projects": [
    {
      "project_dir": "D:\\code\\cicbyte\\git-mate",
      "exists": true,                    // 项目目录是否存在
      "initialized": true,               // .reference/reference.settings.json 的 initialized 字段
      "agents": ["claude", "zcode"],     // 注入的 agent 列表（从 settings.json 读，兼容旧 agent 字段）
      "repo_count": 12,
      "checks": [
        {
          "name": "软链接完整性",
          "status": "error",             // ok | warn | error
          "details": "11/12 正常, 1 断裂: opencommit",
          "group": "core"
        },
        {
          "name": "Wiki Junction",
          "status": "ok",
          "details": "12/12 正常",
          "group": "core"
        }
        // ... 其余检查项同现有 doctor
      ],
      "healthy": false,                  // 所有 checks 都 ok 时为 true
      "issues_count": 1                  // status != ok 的检查项数量
    }
    // ... 其余项目
  ],
  "summary": {
    "total_projects": 62,
    "existing": 61,
    "deleted": 1,
    "healthy": 58,
    "with_issues": 3,
    "checks_total": 310,                 // 检查项总数
    "checks_failed": 4                   // status != ok 的检查项总数
  }
}
```

#### 实现要点

**复用现有代码**：

- 项目遍历：参考 `internal/logic/global/global_list.go` 的 `indexer.GetAllProjects()` 逻辑
- 单项目检查：抽离 `internal/logic/repo/doctor.go` 的 `runChecks` 为可独立调用的函数（当前可能是 cmd 层直接调用），签名类似 `func RunChecks(projectDir string, db *gorm.DB) ([]CheckResult, error)`
- Agent 读取：`models.LoadProjectSettings(projectDir)` 已有，迁移旧 `Agent` 字段到 `Agents` 数组（`MigrateAgent()` 已实现）

**性能优化**：

- 项目间无依赖，建议 worker pool 并发（默认 `runtime.NumCPU()`，可通过 `--concurrency` 控制）
- 文件系统操作（stat 链接、读 junction）是主要开销，并发能显著加速
- 给整个命令加超时保护（默认 5 分钟，超大项目集合时不卡死）

**边界情况**：

- 已删除项目目录（`exists=false`）：跳过文件系统检查，只标记 `issues=[{name:"项目目录", status:"error", details:"目录不存在"}]`，`healthy=false`
- 无 `.reference/` 的项目：`initialized=false`，`agents=[]`，正常跑检查（doctor 本身就处理这种）
- 权限不足读 settings.json：当 `agents=[]` 处理，不报错

#### 验收标准

- [ ] 不传 `--project` 时遍历所有项目，输出含每个项目的 `checks` 数组
- [ ] `--project <path>` 仅诊断指定项目，输出 `projects` 数组只有一项
- [ ] `--issues-only` 过滤掉 `healthy=true` 的项目
- [ ] `-f json` 输出严格符合上述 schema
- [ ] `summary` 字段统计正确（healthy + with_issues = existing）
- [ ] 已删除项目目录被正确标记且不 panic
- [ ] 62 个项目下，顺序模式 < 30s，并发 8 模式 < 8s

---

### 2. `global list` 增强

#### 新增字段

在现有 `global list -f json` 输出的每个 project 对象上，补三个字段：

```jsonc
{
  "project_dir": "D:\\code\\cicbyte\\git-mate",
  "exists": true,
  "repo_count": 12,
  "repos": [/* 现有结构不变 */],
  // ↓ 新增字段
  "initialized": true,
  "agents": ["claude", "zcode"],
  "broken_count": 0                      // 软链接/junction 断裂数（轻量检查，不等同 doctor）
}
```

#### 实现要点

- `initialized` / `agents`：复用 P0 #1 的 `LoadProjectSettings` 逻辑（如果做了 `global doctor`，这里直接复用其内部函数）
- `broken_count`：**轻量检查**，只统计 `.reference/repos/` 下 symlink/junction 失效的数量（`os.Lstat` + 验证目标存在），**不跑完整 doctor**。给列表一个快速的健康提示，详情看 `global doctor`

#### 验收标准

- [ ] `initialized` 准确反映 settings.json
- [ ] `agents` 兼容新旧 schema（`agents` 数组优先，回退 `agent` 单值）
- [ ] `broken_count` 与实际断裂数一致
- [ ] 现有字段保持不变（向后兼容）

---

## P1 改造详述

### 3. 写操作支持 `--project` flag

让以下命令都能脱离 cwd 跨项目执行，统一参考 `global remove --project` 的模式：

#### 受影响命令

| 命令 | flag | 行为 |
|------|------|------|
| `init` | `--project <path>` | 在指定项目下初始化，不动 cwd |
| `repo update` | `--project <path>` | 更新指定项目的远程引用缓存 |
| `proxy set` | `--project <path>` | 改指定项目的代理配置（如果代理是项目级）/ 全局代理（如果代理是全局级，见下） |
| `doctor` | `--project <path>` | 诊断指定项目（可选，`global doctor` 已覆盖）|

#### 实现要点

cobra 层加 local flag：

```go
cmd.Flags().String("project", "", "目标项目目录路径（默认当前目录）")
```

logic 层把 flag 注入到 `config.ProjectDir`：

```go
if p, _ := cmd.Flags().GetString("project"); p != "" {
    config.ProjectDir = abs(p)  // 解析为绝对路径
}
```

后续逻辑**零改动** —— 现有所有 logic 都已经通过 `config.ProjectDir` 定位项目，只是默认是 cwd。

#### Proxy 的特殊性

需要先确认：`reference proxy` 设置的代理是**全局生效**（影响所有项目的克隆）还是**项目级**？

- 如果是全局（当前实现推测是全局，因为只有一个 `~/.cicbyte/apps/reference/config/`）：`proxy set` 不需要 `--project` flag，本身就是全局操作
- 如果是项目级：加 `--project` flag

**建议**：保持 proxy 全局语义不变，**不加** `--project` flag，避免语义混淆。

#### 验收标准

- [ ] `init --project <path> --agent claude` 在目标项目生成 `.reference/`
- [ ] `repo update --project <path> <name>` 更新指定项目的指定引用
- [ ] `--project` 不传时行为与现在完全一致（向后兼容）
- [ ] `--project` 接受相对路径（自动转绝对）和绝对路径

---

## P2 改造详述（可选）

### 4. `global doctor --fix` — 自动修复

#### 适合自动修复的问题

| 问题 | 修复动作 | 安全性 |
|------|---------|--------|
| 断裂的软链接/junction | 重新创建（依据 DB 中的 `cache_path` / `local_path`）| 安全 |
| 缺失的 Agent 配置文件 | 重新注入（依据 settings.json 的 `agents`）| 安全 |
| 缺失的 SKILL.md | 从模板重新生成 | 安全 |
| `reference.map.jsonl` 与 DB 不一致 | 以 DB 为准重建 map | 较安全 |

#### 不自动修复的问题

- 远程缓存缺失（需要重新克隆，可能 GB 级，需用户确认）
- DB 与文件系统严重不一致（需人工判断哪边对）

#### 命令形态

```bash
# 预览会修复什么
reference global doctor --fix --dry-run

# 实际修复
reference global doctor --fix

# 只修复指定项目
reference global doctor --fix --project <path>

# 只修复指定类别
reference global doctor --fix --only links,agents
```

输出在 `global doctor` schema 基础上，每个 check 加 `fixed: true/false` 字段。

---

## 数据模型参考

### DB schema（仅用于理解，**不直读**）

`~/.cicbyte/apps/reference/db/app.db` 单表 `repos`：

```sql
CREATE TABLE `repos` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `created_at` datetime,
  `updated_at` datetime,
  `project_dir` text NOT NULL,           -- 项目绝对路径
  `link_name` text NOT NULL,             -- 引用名
  `ref_type` text NOT NULL,              -- remote | local
  `remote_url` text,                     -- 远程 URL
  `host` text,                           -- github.com 等
  `namespace` text,                      -- owner
  `repo_name` text,                      -- 仓库名
  `cache_path` text,                     -- 全局缓存路径
  `local_path` text,                     -- 本地仓库路径（local 类型）
  `branch` text,
  `commit` text,
  `commit_at` datetime,
  `wiki_sub_path` text,
  `ref_name` text
);
```

### 项目配置文件（每个项目独立）

`<project>/.reference/reference.settings.json`：

```jsonc
// 新版 schema
{
  "agents": ["claude", "zcode"],
  "initialized": true
}

// 旧版 schema（自动迁移）
{
  "agent": "claude",
  "initialized": true
}
```

### 项目引用映射

`<project>/.reference/reference.map.jsonl`（每行一个引用，含 topics 索引）—— 现有结构不变。

---

## git-mate 对接方案（exe 落地后）

### 新增后端绑定

| 绑定 | 调用 | 用途 |
|------|------|------|
| `ReferenceGlobalDoctor()` | `global doctor -f json` | 一次性获取所有项目健康状态 |
| `ReferenceProjectInit(projectDir, agent)` | `init --project <path> --agent X` | 跨项目初始化 |
| `ReferenceProjectUpdate(projectDir, name)` | `repo update --project <path> <name>` | 跨项目更新 |

### 前端页面：全局管理后台

替换现有「全局概览」页的「全局视图」section 为完整的**项目表格**：

```
┌──────────────────────────────────────────────────────────────┐
│ 全局管理                                          [刷新] [批量GC]│
├──────────────────────────────────────────────────────────────┤
│ 筛选: [全部 ▼] [健康 ▼] [Agent ▼]      搜索: [___________]    │
├──────────────────────────────────────────────────────────────┤
│ 项目                              │ 引用 │ Agent       │ 健康 │
├───────────────────────────────────┼──────┼─────────────┼──────┤
│ D:\code\cicbyte\git-mate          │  12  │ claude,zcode│  ✓   │
│ D:\code\cmsg                      │  3   │ claude      │  ⚠ 1 │ ← 展开
│ D:\code\glm\glm_bill (已删除)      │  3   │ —           │  ✗   │
│ ...                                                              │
├──────────────────────────────────────────────────────────────┤
│ 总计 62 个项目,58 健康,3 有问题,1 已删除                       │
└──────────────────────────────────────────────────────────────┘
```

**交互**：
- 点击项目行展开 → 显示该项目的所有检查项（复用 `DoctorSection` 渲染）
- 右键项目 → 「重新初始化」「移除所有引用」「更新所有引用」（跨项目操作）
- 顶部筛选：健康状态（全部/健康/有问题/已删除）、Agent 类型
- 批量选择 + 批量 GC

### 工作量

| 模块 | 工作量 |
|------|--------|
| 后端 3 个绑定 + 类型 | 1 小时 |
| 前端后台表格页面 | 3-4 小时 |
| 筛选 + 展开详情 + 右键菜单 | 2 小时 |
| **合计** | **半天** |

---

## 实施计划

### 阶段一：P0（解锁后台最小可用）

1. **`global doctor` 命令实现**
   - 抽离 `doctor.go` 的检查逻辑为可复用函数
   - 新建 `cmd/global/doctor.go` + `internal/logic/global/global_doctor.go`
   - 实现 worker pool 并发
   - 完善边界处理（已删除项目、权限错误）

2. **`global list` 字段增强**
   - 复用 `LoadProjectSettings` 读 agent
   - 加 `broken_count` 轻量检查

3. **git-mate 接入**
   - 新增 `ReferenceGlobalDoctor` 绑定
   - 重构「全局概览」页为项目表格 + 健康状态

### 阶段二：P1（体验完善）

4. **写操作加 `--project` flag**
   - `init` / `repo update` 统一加 flag
   - proxy 保持全局语义不变

5. **git-mate 接入跨项目操作**
   - 表格右键菜单接入 `init --project` / `update --project`

### 阶段三：P2（锦上添花，可选）

6. **`global doctor --fix` 自动修复**
7. git-mate 接入「一键修复」按钮

### 时间预估

| 阶段 | reference 侧 | git-mate 侧 | 合计 |
|------|-------------|-------------|------|
| 阶段一（P0） | 半天 | 半天 | 1 天 |
| 阶段二（P1） | 2-3 小时 | 2 小时 | 半天 |
| 阶段三（P2） | 半天 | 2 小时 | 半天 |
| **总计** | **~1.5 天** | **~1 天** | **~2.5 天** |

---

## 不做的事（明确排除）

- ❌ **git-mate 直读 SQLite** — 写操作绕过 CLI 极其危险，schema 随版本演进会失效
- ❌ **git-mate 循环 spawn `doctor`** — 62 次进程启动慢且不可用，必须 CLI 提供批量能力
- ❌ **在 git-mate 里重写 reference 的业务逻辑** — 链接/缓存/junction/map 的同步清理封装在 CLI 内，不应在 GUI 层复制
- ❌ **proxy 加 `--project` flag** — 代理是全局配置，加项目级语义会混淆
- ❌ **破坏现有命令的向后兼容** — 所有新能力通过新命令或新 flag 解锁

---

## 验收（端到端）

改造完成后，以下场景在 git-mate 后台页面可用：

- [ ] 进入页面 5 秒内看到所有 62 个项目的健康状态
- [ ] 能筛选出「所有有问题的项目」并一眼看到问题详情
- [ ] 能看到「每个项目注入了哪些 agent」
- [ ] 不离开 git-mate 就能给任意项目重新初始化
- [ ] 不离开 git-mate 就能更新任意项目的引用
- [ ] 已删除项目目录被正确标记，可一键清理（GC）
- [ ] 一次 `global doctor` 命令的耗时在可接受范围内（顺序 < 30s，并发 < 8s）
