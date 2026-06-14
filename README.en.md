<!--
@Project: DBM-Lite Lightweight Full-Scope Database Control Platform
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->

# DBM-Lite Lightweight Full-Scope Database Control Platform

> Full-stack open source with Go + Vue, a plugin-based DBA efficiency tool

---

<p align="center">
  <img src="https://img.shields.io/badge/version-v0.1.0-409EFF?style=for-the-badge" alt="Version">
  <img src="https://img.shields.io/badge/license-Apache--2.0%20%7C%20MulanPSL--2.0-2E6BA8?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Vue-3.4+-42B883?style=for-the-badge&logo=vuedotjs" alt="Vue">
  <img src="https://img.shields.io/badge/MySQL-%E2%9C%85-4479A1?style=for-the-badge&logo=mysql" alt="MySQL">
  <img src="https://img.shields.io/badge/SQLite-%E2%9C%85-003B57?style=for-the-badge&logo=sqlite" alt="SQLite">
</p>

## 📸 Product Preview

> Below is a placeholder of the product interface. Run `run.bat` to launch the actual application.

```
┌───────────────────────────────────────────────────────────────┐
│  DBM-Lite  ─ DB Operation Domain ──────────────────────────     │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  SQL Workbench  |  DB Permissions  |  Data Sources        │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  ┌──────────────┐   ┌────────────────────────────────┐  │  │
│  │  │ Datasource Tabs │   │  SELECT * FROM users LIMIT 100 │  │
│  │  │ [MySQL][TiDB]│   │                                │  │  │
│  │  │  [SQLite]    │   │  ┌──────────────────────────┐   │  │  │
│  │  └──────────────┘   │  │  id | name | email      │   │  │  │
│  │  ┌──Object Tree┐     │  │  ──────────────────────  │   │  │  │
│  │  │ mydb        │     │  │  1  | John | j@db.com   │   │  │  │
│  │  │  ├users     │     │  │  2  | Jane | je@db.com  │   │  │  │
│  │  │  └orders    │     │  └──────────────────────────┘   │  │  │
│  │  └──────────┘        │                                │  │  │
│  │                       └────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                         DBM-Lite v0.1.0 © DB老王 │
└───────────────────────────────────────────────────────────────┘
```

## 🔑 Default Account

- **Username:** `admin`
- **Password:** `admin123`

## ✨ Core Features

| Implemented (v0.1.0) | Planned |
| --- | --- |
| ✅ Multi datasource management (MySQL / TiDB / SQLite) | 🔲 Backup & recovery / Health check / Slow log analysis |
| ✅ Multi-tab SQL workbench (Object tree + Result set + High-risk SQL validation) | 🔲 Plugin-based ops script ecosystem |
| ✅ Audit log | 🔲 Cluster lifecycle management |
| ✅ Basic account & permission system | 🔲 Distributed database extension |
| ✅ Dual-domain switching (DB Operation Domain / DB Maintenance Domain) | 🔲 DB permissions (production-grade) |
| ✅ Page memory + Layout switching | 🔲 DB lifecycle management (production-grade) |

## 🚀 Quick Start

### Windows One-Click Start (Recommended)

```bash
# Clone the repository
git clone https://github.com/DB老王/dbm-lite
cd dbm-lite

# One-click start (auto-clean ports, build backend, launch frontend)
run.bat

# Browser auto-opens http://localhost:5173
# Default account: admin / admin123
```

### Docker Deployment

```bash
# Launch frontend + backend with one command
docker-compose up -d

# Open http://localhost:5173 in your browser
```

### Local Development

```bash
# === Backend ===
cd backend
go mod tidy
go build -o dbm-lite.exe ./cmd/server
./dbm-lite.exe         # or run directly on Windows
# Server port: 8080

# === Frontend (another terminal) ===
cd ../frontend
npm install
npm run dev
# Dev server port: 5173
```

## 🏗 Tech Stack

### Backend

- **Go 1.22+**：High-performance backend service
- **Gin**：Lightweight web framework
- **GORM**：ORM framework for database operations
- **SQLite**：Embedded database (default storage), switchable to MySQL / TiDB
- **JWT**：Stateless authentication
- **AES**：Sensitive data (database password) encryption
- **gin-contrib/cors**：Production-grade CORS middleware

### Frontend

- **Vue 3 Composition API**：Reactive frontend framework
- **Element Plus**：UI component library (dual-domain theme)
- **Vue Router**：Page routing management
- **Pinia**：Frontend state management
- **Monaco Editor**：SQL code editor (syntax highlighting, auto-complete)
- **Vite**：Blazing fast build tool

### Deployment

- **Docker**：Frontend-backend separation containerization
- **Docker Compose**：Container orchestration, one-click launch
- **Windows batch script**：One-click start (`run.bat`)

## 🏛 Architecture

```
              ┌───────────────────────────────┐
              │ Web Browser (localhost:5173) │
              └──────────┬────────────────────┘
                         │ HTTP/HTTPS
              ┌──────────┴────────────────────┐
              │ Frontend (Vue3 + Element Plus)│
              │ · SQL Workbench · Datasources │
              │ · Ops Domain · Audit Log      │
              └──────────┬────────────────────┘
                         │ REST API
              ┌──────────┴────────────────────┐
              │  Backend (Gin + JWT)          │
              │ · /api/auth · /api/sql        │
              │ · /api/datasources ...        │
              └──────────┬────────────────────┘
                         │
              ┌──────────┴────────────────────┐
              │   Embedded SQLite Storage     │
              │ (users · datasources · audit) │
              └───────────────────────────────┘
                         │
               ┌────────┼─────────┐
               ▼        ▼         ▼
          MySQL       TiDB      SQLite    ← Target datasources
```

## 📂 Project Structure

```
dbm-lite/
├── backend/                    # Backend service
│   ├── cmd/server/             # Entry point
│   ├── config/                 # Configuration
│   ├── internal/
│   │   ├── database/           # Database initialization
│   │   ├── dbtype/             # Database type enum
│   │   ├── handler/            # HTTP route handlers
│   │   ├── middleware/         # Middleware (auth, CORS)
│   │   ├── model/              # Data models
│   │   └── service/            # Business services
│   ├── pkg/
│   │   ├── crypto/             # AES / password utilities
│   │   ├── dbpool/             # Connection pool
│   │   ├── sqllint/            # High-risk SQL lint
│   │   └── plugin/             # Plugin protocol framework (reserved)
│   ├── plugins/                # Plugin directory (reserved)
│   ├── go.mod
│   └── .env.example
├── frontend/                   # Frontend application
│   ├── src/
│   │   ├── api/                # API request wrappers
│   │   ├── router/             # Routes & permissions
│   │   ├── stores/             # Pinia state management
│   │   ├── styles/             # Global styles
│   │   └── views/              # Page components
│   ├── package.json
│   └── vite.config.ts
├── docs/                       # Design documents
├── docker-compose.yml
├── run.bat                    # Windows one-click start
├── LICENSE                    # Apache-2.0
├── LICENSE-MulanPSL2          # Mulan PSL v2
└── README.md
```

## 👤 Author

**DB老王**

- **WeChat ID:** `db00db00db00` (for technical communication only)

## 🗺 Roadmap

| Version | Status | Capabilities |
| --- | --- | --- |
| **v0.1.0** | ✅ Released | Core: datasource management, SQL workbench, dual-domain, audit log |
| **v0.2.0** | 🔲 Planned | Ops plugin ecosystem |
| **v0.3.0** | 🔲 Planned |  |

## 📄 License

> This software is licensed under **Apache License 2.0 OR MulanPSL-2.0** dual license — users may freely choose either one.

- Apache License 2.0 — see [LICENSE](LICENSE)
- 木兰宽松许可证第2版 — see [LICENSE-MulanPSL2](LICENSE-MulanPSL2)

© 2026 DBlaowang
