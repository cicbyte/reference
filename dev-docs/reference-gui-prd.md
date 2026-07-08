# Reference GUI — 产品需求文档

## 文档信息

| 字段 | 内容 |
|------|------|
| **项目名称** | Reference GUI |
| **文档版本** | v1.0 |
| **状态** | 规划 |
| **最后更新** | 2026-07-07 |
| **参考架构** | git-mate (Wails v2 + Vue 3 + Ant Design Vue) |

### 版本历史

| 版本 | 日期 | 修改内容 |
|------|------|---------|
| v1.0 | 2026-07-07 | 初始版本 |

---

## 1. 产品概述

为 reference（本地代码仓库引用管理器）添加桌面 GUI 支持。采用与 git-mate 相同的技术栈（Wails v2 + Vue 3 + Ant Design Vue），复用其 GUI 脚手架、目录结构、通用 UI 组件和布局模式，实现全部 CLI 功能的图形化。GUI 既可作为独立桌面应用管理所有项目的引用仓库，也可作为 CLI 的可视化补充。

---

## 2. 背景与问题陈述

### 2.1 背景

reference 当前仅提供 CLI 交互方式。虽然 CLI 对开发者高效，但在以下场景存在不足：
- **可视化需求**：查看全局引用关系、缓存统计、健康诊断时，CLI 输出不够直观
- **批量操作**：批量管理多个项目的引用仓库时，CLI 操作繁琐
- **新用户引导**：纯 CLI 的交互式引导（`reference` 无参数运行）体验有限
- **知识浏览**：Wiki 知识库的内容浏览和管理在终端中体验不佳

### 2.2 问题陈述

用户需要一个可视化界面来：
1. 直观查看所有项目的引用仓库及其关系
2. 通过表单和按钮完成仓库添加/移除/更新操作
3. 可视化浏览代码统计（SCC）结果
4. 浏览和管理 Wiki 知识库内容
5. 查看诊断报告和修复建议

### 2.3 机会

git-mate 已经建立了成熟的 Wails v2 + Vue 3 GUI 脚手架，包括：
- 完整的构建配置（wails.json、vite.config）
- 通用布局组件（Sidebar、Navbar、MainContent）
- 主题系统（暗色/亮色切换）
- 通用 UI 组件（CodeViewer、DiffViewer、MarkdownView 等）
- 状态管理模板（Pinia stores）

reference GUI 可以直接复用这些基础设施，大幅降低开发成本。

---

## 3. 目标用户

### 3.1 主要用户

**使用 reference 的开发者**
- 日常使用 CLI 管理引用仓库
- 需要可视化查看全局引用状态
- 希望在 GUI 中快速完成批量操作

### 3.2 次要用户

**AI 辅助编程用户**
- 使用 Claude Code / Codex 等 AI 助手
- 通过 GUI 浏览 AI Agent 注入的配置和知识文件
- 查看 reference.map.jsonl 的可视化展示

### 3.3 用户故事

| 编号 | 用户故事 | 优先级 |
|------|---------|--------|
| US-001 | 作为开发者，我想要通过 GUI 添加远程仓库引用，以便不用记忆命令参数 | P0 |
| US-002 | 作为开发者，我想要查看所有项目的引用仓库列表，以便了解全局引用状态 | P0 |
| US-003 | 作为开发者，我想要可视化查看代码统计结果，以便快速了解仓库规模 | P0 |
| US-004 | 作为开发者，我想要通过 GUI 运行诊断并查看修复建议，以便快速定位问题 | P1 |
| US-005 | 作为开发者，我想要浏览 Wiki 知识库内容，以便查阅已有的仓库分析 | P1 |
| US-006 | 作为开发者，我想要管理代理设置，以便在受限网络环境下使用 | P1 |
| US-007 | 作为开发者，我想要查看全局统计信息，以便了解缓存占用和仓库数量 | P2 |
| US-008 | 作为开发者，我想要批量更新所有远程仓库缓存，以便一次性完成维护 | P2 |

---

## 4. 目标与成功指标

### 4.1 产品目标 (OKR)

**Objective**: 为 reference 提供与 CLI 功能对等的桌面 GUI，降低使用门槛

- KR1: 所有 CLI 命令均有对应的 GUI 操作入口
- KR2: GUI 启动时间 < 2 秒（不含 Wails 开发模式热加载）
- KR3: 复用 git-mate 80% 以上的 GUI 基础设施

### 4.2 成功指标

| 指标类型 | 指标名称 | 目标值 |
|---------|---------|--------|
| 功能覆盖 | CLI 命令 GUI 化率 | 100% |
| 性能 | 应用启动时间 | < 2s |
| 复用率 | git-mate GUI 组件复用比例 | >= 80% |
| 体积 | 安装包大小 | < 30MB |

---

## 5. 功能需求

### 5.1 功能列表

| 编号 | 功能 | 优先级 | 模块 | 简要描述 |
|------|------|--------|------|---------|
| F-001 | 项目概览 Dashboard | P0 | dashboard | 当前项目引用仓库总览、健康状态、快速操作入口 |
| F-002 | 仓库引用管理 | P0 | repo | add/remove/list/update 操作的 GUI 化 |
| F-003 | 代码统计 (SCC) | P0 | scc | 语言分布图表、Top 文件排名、复杂度分析 |
| F-004 | 诊断修复 (Doctor) | P1 | doctor | 健康检查列表、问题详情、一键修复 |
| F-005 | 全局管理 | P1 | global | 跨项目引用列表、GC 清理、全局统计 |
| F-006 | Wiki 知识库 | P1 | wiki | 知识文件浏览、sync/commit 操作、trash/restore |
| F-007 | 代理设置 | P1 | proxy | 代理 URL/端口设置、状态查看、清除 |
| F-008 | 项目初始化 | P1 | init | Agent 选择、项目配置向导 |
| F-009 | 设置管理 | P2 | settings | 全局配置（repos_path、wiki_path、网络、日志） |
| F-010 | 源码浏览器 | P2 | browser | 浏览引用仓库的文件树和代码内容 |

### 5.2 功能详细描述

#### F-001: 项目概览 Dashboard

**描述**: 打开应用后展示当前 Git 项目的引用仓库总览
**用户价值**: 一眼了解项目的引用状态，快速定位需要操作的仓库

**主流程**:
1. 应用启动，自动检测当前 Git 项目根目录
2. 加载 `.reference/reference.map.jsonl` 获取仓库列表
3. 展示卡片式仓库列表（名称、类型、平台、最后更新时间）
4. 展示健康状态摘要（通过 doctor 快速检查）
5. 提供快速操作入口（添加仓库、更新全部、运行诊断）

**验收标准**:
```gherkin
Scenario: 查看项目概览
  Given 用户在一个已初始化 reference 的 Git 项目中打开应用
  When 应用完成加载
  Then 显示所有引用仓库的卡片列表
  And 显示每个仓库的类型（远程/本地）、名称、最后更新时间
  And 显示项目健康状态摘要
```

#### F-002: 仓库引用管理

**描述**: 通过表单和操作按钮完成仓库的添加、移除、列表查看和更新

**F-002a: 添加仓库**

**主流程**:
1. 用户点击"添加仓库"按钮
2. 弹出表单：输入 Git URL 或本地路径、可选自定义名称、可选指定分支
3. 点击确认后调用 `repo add` 逻辑
4. 实时显示克隆进度
5. 完成后刷新仓库列表

**验收标准**:
```gherkin
Scenario: 添加远程仓库
  Given 用户在仓库管理页面
  When 用户输入 "https://github.com/owner/repo" 并确认
  Then 系统克隆仓库到全局缓存
  And 创建 Junction 链接到 .reference/repos/
  And 仓库出现在列表中

Scenario: 添加本地仓库
  Given 用户在仓库管理页面
  When 用户选择"本地仓库"并输入路径 "/path/to/repo" 并确认
  Then 系统创建本地引用
  And 仓库出现在列表中
```

**F-002b: 移除仓库**

**主流程**:
1. 用户在仓库列表中选择目标仓库
2. 点击"移除"按钮
3. 确认对话框：可选"同时删除缓存"(purge)
4. 执行移除操作
5. 刷新列表

**F-002c: 更新仓库**

**主流程**:
1. 用户点击单个仓库的"更新"按钮，或点击"更新全部"
2. 显示拉取进度
3. 完成后显示更新结果（新增 commit 数、最后 commit 时间）

#### F-003: 代码统计 (SCC)

**描述**: 可视化展示引用仓库的代码统计信息

**主流程**:
1. 用户在仓库卡片上点击"代码统计"
2. 调用 SCC 分析
3. 展示结果：
   - 语言分布饼图/环形图
   - Top 文件排名表格（可排序）
   - 代码行数、复杂度等关键指标

**验收标准**:
```gherkin
Scenario: 查看代码统计
  Given 用户选择了一个引用仓库
  When 用户点击"代码统计"
  Then 显示语言分布图表
  And 显示 Top 15 文件排名表格
  And 显示总代码行数、注释行数、空白行数
```

#### F-004: 诊断修复 (Doctor)

**描述**: 运行健康检查并展示诊断结果

**检查项**:
- 软链接/Junction 完整性
- Wiki Junction 状态
- Reference Map 一致性
- 数据库记录与文件系统一致性
- Wiki Git 仓库状态
- Agent 配置完整性

**主流程**:
1. 用户点击"运行诊断"
2. 逐项执行检查，实时显示进度
3. 展示检查结果列表（通过/警告/失败）
4. 失败项提供修复建议或一键修复按钮

#### F-005: 全局管理

**描述**: 跨项目视角的引用管理

**子功能**:
- **全局列表**: 展示所有项目及其引用仓库关系（树形结构）
- **GC 清理**: 展示可清理的过期记录，确认后执行清理
- **全局统计**: 总项目数、总引用数、缓存占用空间、数据库大小

#### F-006: Wiki 知识库

**描述**: 浏览和管理 Wiki 知识库内容

**子功能**:
- **文件浏览**: 树形展示 Wiki 目录结构
- **内容查看**: Markdown 渲染展示知识文件
- **同步操作**: pull + commit + push 一键同步
- **提交更改**: 暂存并提交本地修改
- **远程管理**: 查看/设置远程仓库 URL
- **回收站**: 查看被删除文件、从 Git 历史恢复

#### F-007: 代理设置

**描述**: 管理网络代理配置

**主流程**:
1. 展示当前代理状态（已设置/未设置）
2. 输入代理 URL 或端口
3. 保存设置
4. 可一键清除代理

#### F-008: 项目初始化

**描述**: 为新项目配置 reference

**主流程**:
1. 检测当前项目是否已初始化
2. 未初始化则展示 Agent 选择界面（复选框列表）
3. 选择后执行初始化（保存 settings、注入配置）
4. 显示初始化结果

---

## 6. 范围排除

- **V1 不包含**: 内嵌终端（reference 无终端交互需求）
- **V1 不包含**: AI 集成（reference 本身不含 AI 功能）
- **V1 不包含**: 实时文件监控（reference 无实时变更监控需求）
- **延后**: 源码浏览器（F-010，复杂度高，P2 优先级）
- **明确不做**: Git 操作（reference 只管理引用，不操作用户仓库）

---

## 7. 非功能性需求

| 类别 | 需求 | 量化标准 | 优先级 |
|------|------|---------|--------|
| 性能 | 应用启动时间 | < 2s | P0 |
| 性能 | 仓库列表加载 | < 500ms（100 个仓库） | P0 |
| 性能 | SCC 分析响应 | < 5s（大型仓库） | P1 |
| 可靠性 | 数据库操作原子性 | 操作失败不产生脏数据 | P0 |
| 安全性 | 本地数据存储 | 不上传任何用户数据 | P0 |
| 可用性 | 新手上手时间 | < 3 分钟完成首次仓库添加 | P1 |
| 可移植性 | 平台支持 | Windows 10+、macOS 12+、Linux (Ubuntu 20.04+) | P0 |
| 可维护性 | 代码复用 | 与 git-mate 共享 >= 80% GUI 基础设施 | P1 |

---

## 8. 用户体验设计

### 8.1 布局结构

复用 git-mate 的 AppLayout 布局模式：

```
+------------------+----------------------------------------+
|                  |  Navbar (项目名、面包屑、操作按钮)       |
|   Sidebar        +----------------------------------------+
|   (导航菜单)     |                                        |
|                  |  MainContent                           |
|                  |  (当前页面内容)                         |
|                  |                                        |
+------------------+----------------------------------------+
```

### 8.2 侧边栏菜单结构

```
[Logo] reference

Dashboard          (F-001: 项目概览)
Repository
  ├─ List           (F-002: 仓库列表)
  ├─ Add            (F-002: 添加仓库)
  └─ Update         (F-002: 更新仓库)
Statistics          (F-003: 代码统计)
Doctor              (F-004: 诊断修复)
─────────────────
Global
  ├─ Projects       (F-005: 全局项目列表)
  ├─ GC             (F-005: 垃圾回收)
  └─ Stats          (F-005: 全局统计)
Wiki
  ├─ Browse         (F-006: 知识浏览)
  ├─ Sync           (F-006: 同步)
  └─ Trash          (F-006: 回收站)
─────────────────
Settings
  ├─ Proxy          (F-007: 代理设置)
  ├─ Init           (F-008: 项目初始化)
  └─ Config         (F-009: 全局配置)
```

### 8.3 关键页面设计

**Dashboard 页面**
- 顶部：项目名 + 健康状态指示灯
- 中部：仓库卡片网格（类型图标、名称、来源、最后更新）
- 底部：快速操作栏（添加、更新全部、诊断）

**仓库列表页面**
- 表格视图：名称、类型、来源、分支、最后更新、操作按钮
- 支持搜索和筛选（按类型、按平台）
- 行操作：查看、更新、移除、代码统计

**代码统计页面**
- 左侧：语言分布环形图
- 右侧：关键指标卡片（总行数、文件数、语言数）
- 下方：Top 文件排名表格（可排序、可分页）

**诊断页面**
- 检查项列表，每项显示状态图标（通过/警告/失败）
- 点击展开详情和修复建议
- 顶部"全部修复"按钮

### 8.4 主题系统

复用 git-mate 的主题方案：
- 主色调：`#BD93F9`（Dracula Purple）
- 暗色模式：深蓝背景 (`#0f172a`)，浅色文字 (`#f1f5f9`)
- 亮色模式：白色背景，深色文字 (`#0f172a`)
- 通过 CSS 变量实现主题切换

### 8.5 交互规范

- 所有异步操作显示 Loading 状态
- 操作成功/失败显示 Toast 通知
- 危险操作（移除、purge、GC）需二次确认
- 长时间操作（克隆、更新全部）显示进度条
- 表单验证即时反馈

---

## 9. 技术考量

### 9.1 系统架构

```
┌─────────────────────────────────────────────┐
│                Wails v2 Runtime              │
├─────────────────────────────────────────────┤
│  gui/                    │  gui/frontend/    │
│  ReferenceApp (Go)       │  Vue 3 + Pinia    │
│  ├─ app.go (核心)        │  ├─ views/        │
│  ├─ repo.go (仓库管理)   │  ├─ components/   │
│  ├─ scc.go (代码统计)    │  ├─ stores/       │
│  ├─ doctor.go (诊断)     │  ├─ composables/  │
│  ├─ global.go (全局管理) │  └─ router/       │
│  ├─ wiki.go (知识库)     │                   │
│  ├─ proxy.go (代理)      │                   │
│  └─ settings.go (设置)   │                   │
├─────────────────────────────────────────────┤
│           internal/logic/ (共享业务逻辑)      │
│  ├─ repo/ (仓库管理逻辑)                     │
│  ├─ wiki/ (Wiki 逻辑)                        │
│  └─ global/ (全局管理逻辑)                    │
├─────────────────────────────────────────────┤
│           internal/models/ (数据模型)         │
│           internal/utils/ (工具函数)          │
└─────────────────────────────────────────────┘
```

### 9.2 GUI 绑定层设计

复用 git-mate 的模式：`ReferenceApp` 结构体按功能拆分为多个文件：

| 文件 | 职责 | 对应 CLI 命令 |
|------|------|--------------|
| `gui/app.go` | 核心结构体、startup/shutdown 生命周期 | — |
| `gui/repo.go` | 仓库 CRUD 操作 | `repo add/remove/list/update` |
| `gui/scc.go` | 代码统计 | `repo scc` |
| `gui/doctor.go` | 诊断检查 | `doctor` |
| `gui/global.go` | 全局管理 | `global list/gc/stats/remove/doctor` |
| `gui/wiki.go` | Wiki 操作 | `wiki commit/sync/remote/trash/restore` |
| `gui/proxy.go` | 代理管理 | `proxy set/info/clear` |
| `gui/init.go` | 项目初始化 | `init` |
| `gui/settings.go` | 设置管理 | 配置读写 |
| `gui/browser.go` | 浏览器/文件夹操作 | — |

### 9.3 前端目录结构

```
gui/frontend/
├── src/
│   ├── views/              # 页面视图
│   │   ├── DashboardView.vue
│   │   ├── RepoListView.vue
│   │   ├── RepoAddView.vue
│   │   ├── SccView.vue
│   │   ├── DoctorView.vue
│   │   ├── GlobalListView.vue
│   │   ├── GlobalStatsView.vue
│   │   ├── WikiBrowseView.vue
│   │   ├── WikiSyncView.vue
│   │   ├── WikiTrashView.vue
│   │   ├── ProxyView.vue
│   │   ├── InitView.vue
│   │   └── SettingsView.vue
│   ├── components/         # 组件
│   │   ├── layout/         # 布局组件（从 git-mate 复用）
│   │   │   ├── AppLayout.vue
│   │   │   ├── Sidebar.vue
│   │   │   ├── Navbar.vue
│   │   │   └── MainContent.vue
│   │   ├── common/         # 通用组件（从 git-mate 复用）
│   │   │   ├── CodeViewer.vue
│   │   │   ├── MarkdownView.vue
│   │   │   └── ...
│   │   ├── repo/           # 仓库相关组件
│   │   ├── scc/            # 代码统计组件
│   │   ├── doctor/         # 诊断组件
│   │   ├── global/         # 全局管理组件
│   │   ├── wiki/           # Wiki 组件
│   │   └── settings/       # 设置组件
│   ├── stores/             # Pinia 状态管理
│   │   ├── repo.ts         # 仓库状态
│   │   ├── theme.ts        # 主题状态（复用）
│   │   ├── layout.ts       # 布局状态（复用）
│   │   └── global.ts       # 全局状态
│   ├── composables/        # 组合式函数
│   ├── router/             # 路由配置
│   ├── assets/             # 静态资源
│   │   └── styles/         # 样式文件
│   ├── App.vue
│   └── main.ts
├── wails.json
├── package.json
└── vite.config.ts
```

### 9.4 复用清单

**从 git-mate 直接复用**：
- `gui/main.go` — Wails 启动模板（修改 Title、尺寸、绑定对象）
- `gui/frontend/src/components/layout/` — 完整布局组件
- `gui/frontend/src/components/common/` — 通用组件（CodeViewer、MarkdownView 等）
- `gui/frontend/src/stores/theme.ts` — 主题切换逻辑
- `gui/frontend/src/stores/layout.ts` — 布局响应式逻辑
- `gui/frontend/src/assets/styles/variables.css` — CSS 变量定义
- `gui/frontend/src/App.vue` — 根组件（主题配置、全局样式）
- `gui/frontend/src/main.ts` — 应用入口
- `gui/frontend/vite.config.ts` — Vite 构建配置
- `gui/frontend/package.json` — 依赖模板（按需裁剪）

**需要新建**：
- `gui/app.go` 及各功能文件 — 绑定 reference 的 Logic 层
- `gui/frontend/src/views/` — 所有页面视图
- `gui/frontend/src/components/repo|scc|doctor|global|wiki|settings/` — 功能组件
- `gui/frontend/src/stores/repo.ts` — 仓库状态管理
- `gui/frontend/src/router/index.ts` — 路由配置

### 9.5 数据流

```
Frontend (Vue)
    ↓ 调用 Wails 绑定方法
ReferenceApp (Go)
    ↓ 创建 Processor
internal/logic/repo|wiki|global/ (Logic 层)
    ↓ 操作数据库/文件系统
SQLite + 文件系统 + Git
    ↓ 返回结果
ReferenceApp → Frontend (更新状态)
```

### 9.6 技术约束

- Wails v2 要求所有绑定方法在同一个结构体上
- 前端构建产物通过 `//go:embed` 嵌入 Go 二进制
- Windows Junction 操作需要 PowerShell（Wails 环境下可用）
- 数据库路径由 `internal/utils` 管理，GUI 层通过 Logic 层间接访问

---

## 10. 发布策略

### 10.1 MVP 定义

**Phase 1 — 核心框架 + 仓库管理**:
- Wails 应用脚手架（从 git-mate 复制并适配）
- Dashboard 页面（仓库卡片列表）
- 仓库添加/移除/列表操作
- 基础设置页面

### 10.2 发布阶段

| 阶段 | 内容 | 目标 |
|------|------|------|
| Phase 1 | 脚手架 + Dashboard + Repo CRUD | 可用的仓库管理 GUI |
| Phase 2 | SCC + Doctor + 代理设置 | 完整的项目管理功能 |
| Phase 3 | Global 管理 + Wiki 浏览 | 全局视角管理 |
| Phase 4 | 项目初始化 + 设置管理 + 源码浏览器 | 功能完全对等 CLI |

### 10.3 上线标准

- [ ] 所有 P0 功能验收通过
- [ ] Windows/macOS/Linux 三平台构建成功
- [ ] 应用启动时间 < 2s
- [ ] 与 CLI 共享的 Logic 层测试通过

---

## 11. 风险评估

| 编号 | 风险 | 概率 | 影响 | 缓解策略 |
|------|------|------|------|---------|
| R-001 | Wails v2 跨平台兼容性问题 | 中 | 高 | 优先保证 Windows 平台稳定，其他平台逐步适配 |
| R-002 | git-mate GUI 组件复用时引入不需要的依赖 | 低 | 中 | 按需裁剪 package.json，移除 AI/GitHub 等无关依赖 |
| R-003 | Junction 操作在 Wails 环境下权限问题 | 低 | 高 | 复用 git-mate 已验证的 PowerShell Junction 方案 |
| R-004 | 大型仓库 SCC 分析阻塞 UI | 中 | 中 | 异步执行 + 进度事件流，参考 git-mate 的 review.go 模式 |

---

## 12. 开放问题

| 编号 | 问题 | 状态 |
|------|------|------|
| Q-001 | 是否需要支持多项目同时管理（切换不同 Git 项目）？ | 待定 |
| Q-002 | GUI 版本号是否与 CLI 版本号保持同步？ | 待定 |
| Q-003 | 是否需要将 GUI 打包为独立安装包（NSIS/DMG/AppImage）？ | 待定 |
| Q-004 | Wiki 浏览是否需要支持 Markdown 编辑功能？ | 待定 |

---

## 13. 附录

### 参考资料

- git-mate GUI 架构：`.reference/wiki/git-mate/reference.md`
- git-mate 源码：`.reference/repos/git-mate/gui/`
- reference CLI 命令文档：`docs/`
- Wails v2 官方文档：https://wails.io/docs/
- Ant Design Vue：https://antdv.com/

### 术语表

| 术语 | 说明 |
|------|------|
| Junction | Windows NTFS 目录联接，类似 Unix Symlink，不需要管理员权限 |
| Reference Map | `reference.map.jsonl` 文件，AI Agent 的仓库导航数据 |
| SCC | Source Code Counter，代码统计工具 |
| Logic 层 | `internal/logic/` 中的纯业务逻辑，不依赖 cobra |
