# Tencent Cloud VectorDB GUI 客户端

<p align="center">
  <img src="frontend/public/wails.png" width="96" height="96" alt="VectorDB GUI Logo" />
</p>

<p align="center">
  基于 <b>Wails v3</b> + <b>Go</b> + <b>Vue 3</b> + <b>Tailwind CSS</b> 构建的现代化腾讯云向量数据库桌面客户端。
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Wails-v3.0.0--beta.22-df1a7a?logo=wails" alt="Wails v3" />
  <img src="https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TailwindCSS-3.x-38B2AC?logo=tailwind-css" alt="Tailwind CSS" />
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License" />
</p>

---

## 📖 项目简介

**Tencent Cloud VectorDB GUI** 是一款专为腾讯云向量数据库（Tencent Cloud VectorDB）设计的轻量级、跨平台桌面管理工具。旨在为算法工程师、开发者与数据库管理员提供直观的可视化界面，方便快速验证集群连通性、浏览数据库与集合（Collection）元数据、检查 Schema 与向量索引、以及分页检索和审视高维向量文档数据。

---

## ✨ 核心特性

- ⚡ **多连接配置与安全持久化**：
  - 支持添加、保存、切换多套腾讯云 VectorDB 实例配置（公网或内网地址）。
  - 连接配置存储于用户主目录（`~/.tcvectordb/connections.json`），设置严格的文件与目录访问权限（`0600`/`0700`），保障 API Key 安全。
  - 内置快速连通性测试（Ping）功能，即时反馈网络与鉴权状态。
- 🗂️ **层级树状导航与实时检索**：
  - 侧边栏以树形结构展示 `Database` 及所属 `Collection`。
  - 支持 Collection 名称关键字实时模糊过滤，秒级定位目标集合。
- 🔍 **集合元数据与 Schema 结构查阅**：
  - 概览卡片直观呈现分片数（Shard）、副本数（Replica）与描述信息。
  - **字段定义表格**：清晰标识字段名、数据类型、用途、主键（PK）标识。
  - **索引配置分析**：展示标量索引、向量索引类型（如 `HNSW`、`FLAT`）、距离度量方式（`COSINE` / `L2` / `IP`）及算法参数（`M`、`efConstruction` 等）。
- 📊 **文档浏览与高维向量渲染优化**：
  - **标量过滤查询**：支持在界面输入 Filter 表达式（例如 `id in ("doc-001")` 或 `age > 18`）执行精准过滤。
  - **灵活分页控制**：支持 10 / 20 / 50 / 100 条每页切换，前后翻页顺畅。
  - **高维向量防卡顿折叠**：自动识别高维数组（如 768 / 1536 维 Embedding），在表格中折叠展示为紧凑摘要标签（`[0.123, -0.456, ... (1536 维)]`），避免成千上万个 DOM 节点导致界面假死。
  - **JSON 详情弹窗**：点击行内操作一键展开整条文档的格式化 JSON，支持包含向量维度提示并可一键复制到剪贴板。
- 🌓 **浅色 / 深色双主题无缝切换**：
  - 针对高强度使用场景提供深度暗色极客主题与素雅高对比浅色主题。
  - 本地记忆主题偏好，标题栏一键切换。
- 🪟 **现代无边框桌面体验 (Frameless Window)**：
  - 隐藏系统自带沉重边框，内置现代化拖拽标题栏。
  - 支持拖拽移动、双击最大化/还原，以及自定义最小化、最大化与关闭按钮。
- 🛡️ **轻量纯原生实现，零依赖包袱**：
  - 后端直接采用 Go 标准库 `net/http` 封装腾讯云 VectorDB RESTful 接口与鉴权机制。
  - 完美避开官方 Go SDK 中的复杂历史依赖（如不同版本的 `genproto`、分词库、COS 存储等冲突），编译产物仅约 12MB，极速启动。

---

## 🛠️ 技术栈

| 模块 | 技术选型 | 说明 |
| :--- | :--- | :--- |
| **桌面框架** | [Wails v3](https://v3.wails.io/) (`beta.22`) | Go + Webview 跨平台轻量桌面应用框架 |
| **后端语言** | Go 1.27+ | 标准库 REST 引擎、并发安全存储、Wails RPC 服务 |
| **前端框架** | Vue 3 (Composition API) | `<script setup>` 语法，模块化状态管理 |
| **样式与组件**| Tailwind CSS + Lucide Icons | 响应式双主题系统、精致轻量图标库 |
| **构建工具** | Vite + pnpm + Taskfile | 前端纳秒级构建与 Wails 任务流水线 |

---

## 📁 目录结构

```
vectordb-1/
├── main.go                       # Wails v3 应用入口与窗口配置（Frameless 模式）
├── Taskfile.yml                  # 自动化构建与开发任务
├── go.mod                        # Go 模块与依赖定义
├── backend/                      # Go 后端业务逻辑
│   ├── client/                   # 腾讯云 VectorDB HTTP REST 客户端封装
│   │   ├── client.go             # 鉴权、接口请求、错误捕获与重试
│   │   ├── types.go              # 对齐腾讯云 SDK 的 API 结构定义与容错反序列化
│   │   └── client_test.go        # HTTP Mock 单元测试
│   ├── service/                  # Wails 暴露给前端的 RPC 服务层
│   │   ├── connection.go         # ConnectionService（连接管理、Ping 测试）
│   │   ├── vectordb.go           # VectorDBService（库、集合、文档、缓存调度）
│   │   └── service_test.go       # 服务集成与并发测试
│   └── storage/                  # 本地持久化存储
│       ├── storage.go            # connections.json 文件读写与权限控制 (0600)
│       └── storage_test.go       # 存储层 TDD 单元测试
├── frontend/                     # Vue 3 前端工程
│   ├── src/
│   │   ├── components/           # UI 界面组件
│   │   │   ├── TitleBar.vue      # 自定义拖拽标题栏与窗口三键控制器
│   │   │   ├── Sidebar.vue       # 侧边栏（实例下拉、操作按钮、底栏）
│   │   │   ├── CollectionTree.vue# 集合树状折叠视图与模糊搜索
│   │   │   ├── MainView.vue      # 右侧工作区容器（面包屑、Tab 切换）
│   │   │   ├── DataExplorer.vue  # 文档数据动态表格、分页器与过滤栏
│   │   │   ├── SchemaViewer.vue  # 集合结构字段与索引卡片
│   │   │   ├── ConnectionModal.vue # 连接配置创建与编辑弹窗
│   │   │   └── JsonDetailModal.vue # 文档完整 JSON 与高维向量查看弹窗
│   │   ├── stores/               # 响应式状态管理 (connection, vectordb, theme)
│   │   ├── types/                # 前端 TypeScript 模型契约
│   │   ├── App.vue               # 根视图组件
│   │   └── main.ts               # Vue 入口
│   ├── tailwind.config.js        # Tailwind 样式配置（开启 class 深色模式）
│   └── vite.config.ts            # Vite 配置文件
└── bin/                          # 本地编译产物目录
```

---

## 🚀 快速上手

### 1. 前置环境准备
- **Go**：`go >= 1.22`
- **Node.js**：`node >= 18` 与 `pnpm`
- **Wails v3 CLI**：
  ```bash
  go install github.com/wailsapp/wails/v3/cmd/wails3@latest
  ```
- **系统依赖（Linux 用户）**：
  - GTK3 与 WebKitGTK 开发库（例如 openSUSE: `webkit2gtk-4_1-devel`；Ubuntu/Debian: `libgtk-3-dev libwebkit2gtk-4.1-dev`）。

### 2. 获取代码与安装前端依赖
```bash
git clone https://github.com/your-org/vectordb-gui.git
cd vectordb-gui/frontend
pnpm install
cd ..
```

### 3. 运行单元测试
运行后端完整测试套件（含竞态检测）：
```bash
go test ./backend/... -v -race
```

### 4. 本地开发模式（热重载）
```bash
wails3 task dev
```
此命令将同时启动 Vite 前端热重载服务器以及 Go Wails 调试应用，修改代码即时生效。

### 5. 生产环境打包构建
```bash
wails3 task build
```
构建完成后，原生二进制程序将输出在 `bin/` 目录下：
```bash
./bin/tcvectordb-gui
```

---

## 💡 使用指南

1. **创建实例连接**：
   - 首次启动时，点击左侧侧边栏顶部的 **`+`** 按钮打开连接弹窗。
   - 填入实例信息：
     - **实例地址 (URL)**：腾讯云控制台提供的 IP:Port 或公网负载均衡域名（例如 `http://10.0.0.1:80` 或 `http://lb-xxx.clb.ap-guangzhou.tencentclb.com:50000`）。
     - **账号 (Username)**：默认为 `root`。
     - **API Key**：腾讯云控制台获取的 API 密钥。
     - **超时时间**：默认 10 秒（建议网络延迟较高时调大至 20~30 秒）。
   - 点击 **测试连通性**，确认连接无误后点击 **保存连接**。
2. **选择库与集合**：
   - 展开左侧数据库树状节点，选择目标 Collection，右侧工作区将自动加载其元数据与文档。
3. **数据查询与过滤**：
   - 切换至 **数据浏览** Tab，可在顶部过滤框中输入条件过滤（如 `title = "VectorDB"`），点击 **查询**。
   - 点击高维向量单元格或右侧 **JSON** 按钮，即可在弹窗中完整查阅高维数值并一键复制。
4. **切换外观与窗口状态**：
   - 点击右上角或左下角太阳/月亮图标可自由切换明暗模式。
   - 双击顶部标题栏可直接最大化或还原窗口。

---

## 📄 开源许可

本项目基于 [MIT License](LICENSE) 开源。
