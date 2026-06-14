<<<<<<< HEAD
<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DBA老王
@License: Apache-2.0 OR MulanPSL-2.0
-->

# DBM-Lite 轻量级全域数据库管控平台

> Go + Vue 全栈开源，面向 研发、测试、运维、DBA 的插件化数据库效率工具

***

<p align="center">
  <img src="https://img.shields.io/badge/version-v0.1.0-409EFF?style=for-the-badge" alt="Version">
  <img src="https://img.shields.io/badge/license-Apache--2.0%20%7C%20MulanPSL--2.0-2E6BA8?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Vue-3.4+-42B883?style=for-the-badge&logo=vuedotjs" alt="Vue">
  <img src="https://img.shields.io/badge/MySQL-%E2%9C%85-4479A1?style=for-the-badge&logo=mysql" alt="MySQL">
  <img src="https://img.shields.io/badge/SQLite-%E2%9C%85-003B57?style=for-the-badge&logo=sqlite" alt="SQLite">
</p>

## 📸 产品预览

> 以下为产品界面预览占位，实际使用请运行 run.bat 启动。

```
┌───────────────────────────────────────────────────────────────┐
│  DBM-Lite  ─ DB操作管控域 ─────────────────────────────────     │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  SQL IDE  |  数据库权限管理  |  数据源管理             │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  ┌──────────────┐   ┌────────────────────────────────┐  │  │
│  │  │ 数据源Tab栏  │   │  SELECT * FROM users LIMIT 100  │  │  │
│  │  │ [MySQL] [TiDB│   │                                │  │  │
│  │  │  [SQLite]    │   │  ┌──────────────────────────┐   │  │  │
│  │  └──────────────┘   │  │  id | name | email      │   │  │  │
│  │  ┌──对象树──┐        │  │  ──────────────────────  │   │  │  │
│  │  │ mydb     │        │  │  1  | 张三 | z@db.com   │   │  │  │
│  │  │  ├users  │        │  │  2  | 李四 | l@db.com   │   │  │  │
│  │  │  └orders │        │  └──────────────────────────┘   │  │  │
│  │  └──────────┘        │                                │  │  │
│  │                       └────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                         DBM-Lite v0.1.0 © DBA老王 │
└───────────────────────────────────────────────────────────────┘
```

## 🔑 默认账号

- **用户名：** `admin`
- **密码：** `admin123`

## ✨ 核心特性

| 已实现 (v0.1.0)                    | 规划中                    |
| ------------------------------- | ---------------------- |
| ✅ 多数据源管理（MySQL / TiDB / SQLite） | 🔲 备份恢复 / 健康巡检 / 慢日志分析 |
| ✅ 多标签 SQL 工作台（对象树 + 结果集 + 高危校验） | 🔲 插件化运维脚本生态           |
| ✅ 操作审计日志                        | 🔲 集群生命周期管理            |
| ✅ 基础账号权限体系                      | 🔲 分布式数据库扩展            |
| ✅ 双域切换（DB 操作管控域 / DB 基础运维域）     | 🔲 数据库权限管理（生产级）        |
| ✅ 页面记忆 + 布局切换                   | 🔲 DB 生命周期管理（生产级）      |

## 🚀 快速开始

### Windows 一键启动（推荐）

```bash
# 克隆仓库
git clone https://github.com/DBA老王/dbm-lite
cd dbm-lite

# 一键启动（自动清理端口、编译后端、启动前端）
run.bat

# 浏览器自动打开 http://localhost:5173
# 默认账号: admin / admin123
```

### Docker 部署

```bash
# 一键启动前后端
docker-compose up -d

# 浏览器打开 http://localhost:5173
```

### 本地开发调试

```bash
# === 后端 ===
cd backend
go mod tidy
go build -o dbm-lite.exe ./cmd/server
./dbm-lite.exe     # Windows 下直接运行
# 服务端口: 8080

# === 前端（另一个终端）===
cd ../frontend
npm install
npm run dev
# 开发端口: 5173
```

## 🏗 技术栈

### 后端

- **Go 1.22+**：高性能后端服务
- **Gin**：轻量级 Web 框架
- **GORM**：ORM 框架，驱动数据库操作
- **SQLite**：嵌入式数据库（默认存储），支持切换到 MySQL / TiDB
- **JWT**：无状态鉴权认证
- **AES**：敏感信息（数据库密码）加密存储
- **gin-contrib/cors**：生产级跨域中间件

### 前端

- **Vue 3 Composition API**：响应式前端框架
- **Element Plus**：UI 组件库（双域主题色）
- **Vue Router**：页面路由管理
- **Pinia**：前端状态管理
- **Monaco Editor**：SQL 代码编辑器（语法高亮、自动补全）
- **Vite**：极速构建工具

### 部署与运行

- **Docker**：前后端分离容器化部署
- **Docker Compose**：容器编排，一键启动
- **Windows 批处理脚本**：一键启动（`run.bat`）

## 🏛 整体架构

```
              ┌──────────────────────────┐
              │   Web 浏览器 (localhost:5173) │
              └──────────┬───────────────┘
                         │ HTTP/HTTPS
              ┌──────────┴───────────────┐
              │  前端 (Vue3 + Element Plus) │
              │  · SQL工作台 · 数据源管理    │
              │  · DB运维域 · 审计日志      │
              └──────────┬───────────────┘
                         │ REST API
              ┌──────────┴───────────────┐
              │   后端 (Gin + JWT)        │
              │  · /api/auth · /api/sql   │
              │  · /api/datasources ...   │
              └──────────┬───────────────┘
                         │
              ┌──────────┴───────────────┐
              │   嵌入式 SQLite 存储       │
              │  (users · datasources · audit)│
              └──────────────────────────┘
                         │
               ┌────────┼─────────┐
               ▼        ▼         ▼
          MySQL       TiDB      SQLite     ← 目标数据源
```

## 📂 项目结构

```
dbm-lite/
├── backend/                    # 后端服务
│   ├── cmd/server/             # 入口
│   ├── config/                 # 配置
│   ├── internal/
│   │   ├── database/           # 数据库初始化
│   │   ├── dbtype/             # 数据库类型枚举
│   │   ├── handler/            # HTTP 路由处理器
│   │   ├── middleware/         # 中间件（鉴权、CORS）
│   │   ├── model/              # 数据模型
│   │   └── service/            # 业务服务
│   ├── pkg/
│   │   ├── crypto/             # AES/密码工具
│   │   ├── dbpool/             # 连接池
│   │   ├── sqllint/            # SQL 高危校验
│   │   └── plugin/             # 插件协议框架（预留）
│   ├── plugins/                # 插件目录（预留）
│   ├── go.mod
│   └── .env.example
├── frontend/                   # 前端应用
│   ├── src/
│   │   ├── api/                # API 请求封装
│   │   ├── router/             # 路由与权限
│   │   ├── stores/             # Pinia 状态管理
│   │   ├── styles/             # 全局样式
│   │   └── views/              # 页面组件
│   ├── package.json
│   └── vite.config.ts
├── docs/                       # 设计文档
│   ├── 01-竞品架构与功能拆解.md
│   └── 02-dbm-lite-全生命周期设计文档.md
├── docker-compose.yml
├── run.bat                    # Windows 一键启动
├── LICENSE                    # Apache-2.0
├── LICENSE-MulanPSL2          # 木兰宽松许可证 v2
└── README.md
```

## 👤 作者与交流

- **作者：** DBA老王
- **交流微信号：** `db00db00db00`（仅用于技术交流）

## 🗺 版本路线图

| 版本         | 状态     | 核心能力                        |
| ---------- | ------ | --------------------------- |
| **v0.1.0** | ✅ 已发布  | 核心能力：数据源管理、SQL工作台、双域切换、审计日志 |
| **v0.2.0** | 🔲 规划中 | 运维插件体系                      |
| **v0.3.0** | 🔲 规划中 | <br />                      |

## 📄 开源协议

> 本软件采用 **Apache License 2.0 OR MulanPSL-2.0** 双许可模式，使用者可自由选择其中一种协议使用。

- [Apache License 2.0](LICENSE)
- [木兰宽松许可证第2版](LICENSE-MulanPSL2)

© 2026 DBA老王