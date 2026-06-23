# Changelog

本文件记录 reference 项目的版本变更历史。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## [Unreleased]

### Added
- `repo add` 新增 `--update` / `-u` 标志，缓存已存在时强制 git pull 更新
- `wiki` 新增 `--local` / `-l` 持久化标志，支持操作独立的本地知识库（localwiki）
- 本地仓库知识文件自动存储到 `~/.cicbyte/reference/localwiki/`，与公共 wiki 隔离
- `GetLocalWikiDir()` 配置方法，`InitDataDirs` 自动创建 localwiki 目录
- 多 Agent 支持：`ProjectSettings.Agents` 数组字段替代原 `Agent` 单值，支持同时注入多个 AI Agent 配置
- Agent 注册表机制（`AgentRegistry`），inject/doctor/remove 通过查表实现，新增 agent 只需注册即可
- `reference init --agent claude,cursor` 支持逗号分隔多值
- 交互式引导支持多选菜单

### Changed
- `repo add` 默认行为变更：缓存已存在时不再自动更新，需显式传 `--update`
- `ProjectSettings.Agent` 字段已废弃，自动迁移为 `Agents` 数组（向后兼容旧格式）

## [0.0.5] - 2026-04-22

### Fixed
- 修复删除链接时可能出现的错误

### Docs
- 添加 VS Code 扩展相关文档
- 添加禁止绝对路径写入规则

## [0.0.4] - 2026-04-18

### Added
- `init` 命令支持非交互式模式（`--agent claude|none`），适配 CI 集成

## [0.0.3] - 2026-04-15

### Added
- `repo remove --all --clean` 标志，清除 AI 配置和 `.reference/` 目录
- 支持自定义存储路径配置（`repos_path`、`wiki_path`）
- 配置状态迁移功能，启动时自动检测路径变更并更新数据库

### Changed
- `--format` 标志统一为全局选项，适用于所有输出类子命令

## [0.0.2] - 2026-04-10

### Added
- `global remove` 全局移除项目引用命令
- `repo scc` 内置代码统计功能，无外部依赖
- `reference.map.json` 升级为 JSONL 格式，增强 AI Agent 导航能力
- `global list` / `global gc` / `global stats` 全局引用管理命令
- 英文版 README

### Changed
- 使用短主机名统一处理路径和链接名称
- 移除静态 `scc.md` 文件，改用实时命令获取代码统计
- 移除自动监听提交功能，简化 wiki 提交逻辑
- 重构嵌入文件系统结构，将 PromptsFS 移至 common 包

### Docs
- 添加全局引用管理文档和相关命令说明
- 更新 README 文档添加功能描述和截图

## [0.0.1] - 2026-03-28

### Added
- 项目初始化，核心功能上线
- 远程仓库引用（clone + symlink/junction）
- 本地仓库引用
- Wiki 知识库（Git 版本控制、sync、trash、restore）
- AI Agent 配置自动注入（Claude Code）
- 数据库索引（SQLite + GORM）
- 代理解析
- 跨平台支持（Windows Junction / Unix Symlink）

[Unreleased]: https://github.com/cicbyte/reference/compare/v0.0.5...HEAD
[0.0.5]: https://github.com/cicbyte/reference/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/cicbyte/reference/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/cicbyte/reference/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/cicbyte/reference/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/cicbyte/reference/releases/tag/v0.0.1
