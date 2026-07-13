# Reference GUI

English | [简体中文](README.md)

> The Wails v2 + Vue 3 desktop client for Reference — visually manage Git repository references, knowledge base, and global diagnostics.

## Screenshots

### Dashboard

Project overview: referenced repos, broken links, AI assistant config, health status.

![Dashboard](images/仪表盘.png)

### Project Rail

Quick project switching from the left rail, with collapse and context menu (switch / fix / open / remove).

![Project Rail](images/项目轨道.png)

### Reference Management

Manage the current project's referenced repos, browse code, view statistics, diagnose and fix.

![Reference Management](images/引用管理.png)

### Project List

Global view of all projects, grouped by parent directory, with one-click cleanup of invalid entries.

![Project List](images/项目列表.png)

### Repository Cache

Browse global cached repos grouped by platform / owner, inspect disk usage, purge unused caches.

![Repository Cache](images/仓库缓存.png)

### Global Statistics

Multi-dimensional dashboard: project distribution, platform breakdown, AI assistant usage, top 5 cache by size.

![Global Statistics](images/全局统计.png)

### Remote Knowledge Base

Browse AI-explored knowledge files with Markdown rendering, Mermaid diagrams, and source toggle.

![Remote Knowledge Base](images/远程知识库.png)

### Local Knowledge Base

Manage knowledge for local repositories — stored independently, no shared sync.

![Local Knowledge Base](images/本地知识库.png)

### Settings

Theme, language, storage paths, network proxy, logging policy, AI assistant injection — all configurable.

![Settings](images/设置.png)

## Tech Stack

| Layer | Technology |
|:--|:--|
| Desktop framework | [Wails v2](https://wails.io) (Go backend + WebView frontend) |
| Frontend | Vue 3 + TypeScript + Vite |
| UI components | Ant Design Vue 4.x |
| State management | Pinia |
| Code highlighting | highlight.js + line-numbers plugin |
| Markdown | marked + DOMPurify (XSS protection) |
| Diagrams | Mermaid.js (strict security mode) |
| Internationalization | vue-i18n (Chinese / English) |
| Database | SQLite (glebarez/sqlite, pure Go) |

## Build

```bash
cd gui
wails build
```

Development mode:

```bash
cd gui
wails dev
```

Frontend dependencies:

```bash
cd gui/frontend
npm install
npm run build   # or npm run dev (paired with wails dev)
```

## Directory Structure

```
gui/
├── app.go              # Wails app entry, startup initialization
├── binding.go          # Go methods bound to frontend (IPC interface)
├── main.go             # Wails config, window, CSP security policy
├── diagnose.go         # Interactive diagnosis & repair
├── repo_browser.go     # Code browser (file tree / search / read)
├── wiki_browser.go     # Knowledge base browser
├── frontend/           # Vue 3 frontend
│   └── src/
│       ├── views/      # Page views (modular folders)
│       ├── components/  # Shared components
│       ├── composables/ # Reusable logic (useMermaid / useProjectActions)
│       ├── stores/      # Pinia state (project / theme / layout / locale)
│       ├── i18n/        # Internationalization locale packs
│       └── utils/       # Utility functions
└── images/             # README screenshots
```

## Features

- **Multi-project management** — project rail for quick switching, context menu for common actions
- **Reference management** — add / update / remove repo references, built-in code browser (syntax highlighting + line numbers + search)
- **Repository cache** — global view of cached repos, grouped by platform, async disk usage
- **Knowledge base** — browse AI-explored knowledge files, Markdown rendering + Mermaid zoom/export
- **Global statistics** — multi-dimensional dashboard, pure CSS bar charts
- **Garbage collection** — scan and clean stale DB records and orphaned cache directories
- **Diagnosis & repair** — interactive per-repo diagnosis with targeted fixes (rebuild link / re-clone / relocate)
- **Settings** — theme / language / storage / network / logging / AI assistants, fully visual configuration
