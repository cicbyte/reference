# Changelog

本文件由 git-cliff 自动生成（Tag Release 工作流维护）。

## [unreleased]

### 新增

- *(adapter)* 新增 Gemini/Qwen/Kimi/iFlow 四个 AI 助手适配
- *(cache)* 增批量获取缓存目录大小的后端API和前端调用逻辑
- *(gui)* 新增完整的动画系统
- *(wallpaper)* 添加壁纸系统
- *(gui)* 添加上传壁纸与面板毛玻璃背景功能
- *(global-list)* 重构全局项目列表页面
- 实现完整主题系统，重构明暗主题管理
- *(repo/inject)* 新增MiMo Code代理格式适配逻辑
- *(gui)* 实现了wiki文件的git状态展示功能
- *(gui)* 新增多语言国际化支持
- *(layout)* 新增页脚状态项色调支持并添加项目状态展示
- *(gui)* 新增通用目录选择
- *(gui)* 重构助手管理页面，新增助手基础信息展示
- *(settings)* 新增日志设置页，优化存储与关于页面
- *(gui/frontend)* 新增highlightjs-line-numbers.js依赖以实现代码行号功能
- *(RepoListView)* 增代码行号显示组件
- *(gui)* 实现仓库代码浏览功能，优化仓库列表匹配逻辑
- *(gui)* 新增显示设置页面
- *(gui)* 重构了缓存占用统计逻辑
- *(layout)* Add manual toggle for project sidebar rail
- *(gc)* 重写全局GC页面，支持选择性清理和可视化展示
- *(global-stats-view)* 重构并扩展全局统计页面功能
- *(global-stats)* 优化全局统计页面，异步计算目录大小并重构UI
- *(gui)* 增加移除操作
- *(gui)* 实现项目仓库诊断修复功能
- *(global-list)* 新增失效项目清理功能
- *(GlobalListView)* 添加项目修复、路径复制等右键菜单功能
- *(GlobalListView)* 新增按父目录分组展示功能
- *(global-list)* 重写全局项目列表页面，新增卡片视图与统计功能
- *(LocalWikiView&WikiView)* 为mermaid图表添加弹窗放大、拖拽和导出功能
- *(gui/frontend)* 添加mermaid依赖
- *(wiki-view)* 为本地和远程Wiki页面添加了源码/渲染模式切换功能
- *(wiki-view)* 为本地和在线wiki页面添加上下文菜单和文件管理器打开功能
- *(frontend)* 实现全局页脚状态栏，支持wiki和项目信息展示
- *(gui)* 实现了wiki条目状态检测功能
- *(gui)* 新增本地知识库
- *(gui)* 给GUI应用启动流程添加初始化逻辑
- *(gui)* 后端新增检查缓存仓库路径是否存在于磁盘的逻辑
- *(gui)* 仓库缓存分组展示
- *(gui)* 为CachedRepoItem模型新增type字段区分远程/本地缓存
- *(gui)* 新增基于缓存路径的文件浏览相关
- *(gui)* 新增仓库缓存管理页面
- *(gui)* 新增仓库代码浏览页面
- *(gui)* 新增项目右键菜单与相关工具方法
- 实现多项目管理核心功能
- *(gui)* 添加 SVG logo 并修复 Wails 版本兼容性
- *(gui)* 添加 Wails v2 桌面 GUI 支持
- *(global,cli)* 新增全局诊断功能，优化列表与命令参数
- 添加外部集成支持

### 修复

- *(repo)* 简化 wiki 子路径和链接名的命名规则
- *(agent)* 固定助手列表顺序为 ID 字母序
- *(global-stats)* 修复本地wiki统计的平台和仓库数据问题
- 完成多维度功能优化与安全加固
- *(GlobalStatsView)* 过滤掉本地仓库的平台统计
- 为重新克隆操作添加超时处理和panic恢复
- 修复重新克隆仓库功能的参数传递问题
- 修复项目诊断功能超时和异常处理问题
- *(frontend)* 修复代码高亮不生效的问题
- *(gui/repo_browser)* 修复仓库路径解析逻辑
- *(gui)* 实现自定义标题栏的窗口控制逻辑

### 重构

- *(utils,gui,logic)* 新增隐藏子进程窗口工具并统一使用
- 调整应用数据目录结构，将资源移至apps子目录
- *(gui)* 统一侧边栏折叠状态管理，优化代码浏览器自动加载逻辑
- *(settings)* 重构设置页面，拆分显示分组为子分类
- *(gui/frontend)* 重构失效项目清理功能，优化交互流程
- *(theme)* 新增项目轨道栏区域并重构主题色系统
- *(gui)* 统一修正仓库相关文案为引用管理相关
- *(gui)* 将前端界面英文菜单改为中文
- *(gui/frontend)* 配置路径别名并替换所有相对导入
- 重构前端代码，抽离公共逻辑并调整路由结构
- *(gui)* 统一路径显示格式，替换反斜杠为正斜杠
- *(gui)* 重构布局组件，调整页脚与侧边栏项目展示
- *(layout)* 优化页脚项目信息显示逻辑
- *(frontend)* 调整仪表盘页诊断功能交互方式
- *(gui)* 替换Ant Design机器人图标为自定义AI火花图标
- *(layout)* 重构项目侧边栏的显示控制逻辑
- *(diagnose-modal)* 优化诊断弹窗的计数和按钮逻辑
- *(gui)* 移除了stub相关的状态标识逻辑
- *(wiki views)* 移除多余右键菜单项
- *(gui)* 优化页脚显示和文件路径处理
- *(gui/wiki_browser)* 重构wiki条目解析逻辑，优化路径检测
- *(WikiView)* 重构知识库页面UI和交互逻辑
- *(frontend)* 重构知识库页面，合并浏览和同步功能
- *(gui/repo_browser)* 提前从缓存读取分支和提交信息
- *(gui)* 优化仓库列表和代码统计页面UI与交互
- *(gui)* 将代码统计从独立页面改为弹窗组件
- *(gui)* 将添加仓库功能从独立页面改为弹窗组件
- *(gui)* 重写仪表盘页面，新增统计卡片与健康状态展示
- *(layout)* 重构项目侧边栏切换逻辑，添加平滑过渡动画
- *(gui)* 重构设置页面，合并代理和初始化功能到统一设置面板
- *(repo/cache)* 调整仓库克隆和拉取的单分支配置逻辑
- *(cmd/root)* 移除自动拉取wiki的默认执行逻辑
- 统一使用Git根目录作为项目基准路径

### 文档

- *(agent-platform)* 实现 OpenCode 平台的代理配置适配
- 新增了Wails+Vue3桌面客户端的完整文档

### 格式

- *(global-list)* 重构页面头部布局，优化统计展示样式
- 修复多个页面的间距溢出问题
- *(frontend/settings)* 修复设置页布局偏移问题
- *(settings)* 替换emoji图标为Ant Design矢量图标，优化样式
- *(ui)* 调整mermaid弹窗的尺寸和显示样式
- *(Navbar.vue)* 调整主题切换按钮图标尺寸并交换图标显示逻辑
- *(layout)* 重构项目侧边栏的底部布局
- *(settings)* 调整设置页面表单宽度为全屏适配
- *(DiagnoseModal)* 替换原生按钮为Ant Design按钮并调整布局
- *(GlobalListView)* 优化缺失项目的UI展示与交互
- *(ui)* 移除wiki页面内容区的最大宽度限制
- 移除所有页面的固定最大宽度，改为全屏宽度
- *(gui)* 更新项目所有图标资源与logo设计

### 持续集成

- 重构为 git-cliff 一键发版流水线

### 杂务

- *(gui)* 添加gui运行截图
- 回退wails相关依赖版本并添加构建资源文件
- 更新gitignore并修复项目设置初始化逻辑

### Build

- *(gui)* 替换logo资源并新增logo生成脚本
## [0.0.8] - 2026-06-23

### 新增

- 新增本地知识库支持与多Agent能力升级

### 文档

- 更新相关文档
## [0.0.7] - 2026-06-16

### 新增

- 添加codex助手支持，重构agent文件加载逻辑
- 新增对zcode、mimocode、opencode三种AI助手的支持

### 重构

- 移除浅克隆相关逻辑，适配多AI助手平台
- 调整编程助手选项顺序并修正映射关系
## [0.0.6] - 2026-05-27

### 新增

- *(repo)* 优化了仓库克隆的更新逻辑，改为fetch + reset流程
- *(repo)* 增加通过系统命令执行 git clone 和 git pull 的兼容性逻辑

### 重构

- *(repo)* 替换NoUpdate参数为Update统一更新逻辑

### 文档

- *(prompts)* 统一并强化路径安全规则约束
## [0.0.5] - 2026-05-11

### 修复

- 修复删除链接时可能出现的错误

### 文档

- Add prohibition of absolute path writing rule
- 添加 VS Code 扩展相关文档
## [0.0.4] - 2026-05-08

### 新增

- *(init)* 添加非交互式初始化命令支持
## [0.0.3] - 2026-04-28

### 新增

- *(配置)* 支持自定义存储路径配置
- 添加配置状态迁移功能并改进清理逻辑
- *(repo)* 添加 --clean 标志以清除 AI 配置和 .reference/ 目录

### 重构

- *(全局标志)* 将--format标志统一为全局选项并优化相关代码
## [0.0.2] - 2026-04-23

### 新增

- *(global)* 添加全局移除项目引用命令
- *(repo-scc)* 改进代码统计功能并优化输出格式
- 将 reference.map.json 升级为 JSONL 格式并增强功能
- *(global)* 添加全局引用管理功能

### 重构

- *(repo)* 提取 GetRefName 方法并重构相关调用
- *(embed)* 重构嵌入文件系统结构，将PromptsFS移至common包
- *(repo)* 使用短主机名统一处理路径和链接名称
- *(wiki)* 移除自动监听提交功能并优化提交逻辑
- *(reference)* 移除静态scc.md文件改用实时命令获取代码统计

### 文档

- *(skills/reference)* 添加子代理串行执行的约束说明
- 添加英文版README并更新中文README链接
- 添加全局引用管理文档和相关命令说明
- 更新README文档添加功能描述
- 更新README并调整图片资源
- 更新README文档并添加参考图
- 更新README添加仓库示例截图
## [0.0.1] - 2026-04-22

### 新增

- 初始化项目基础结构和核心功能

### 文档

- 移除过时的命令行扩展说明和需求文档
