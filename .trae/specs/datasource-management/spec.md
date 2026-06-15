# dbm-lite 数据源管理模块 - 产品需求文档

## Overview
- **Summary**: 基于NineData官方MySQL与TiDB数据源配置文档，结合dbm-lite现有技术栈与架构规范，实现完整的数据源管理模块前后端全量代码，支持MySQL、SQLite、TiDB三种数据库类型。
- **Purpose**: 提供统一的数据源管理入口，支持环境分组矩阵展示、列表管理、测试连接、最近使用等核心功能，适配轻量工具定位。
- **Target Users**: 数据库管理员、开发人员、运维人员

## Goals
- [x] 实现8个数据源管理API接口
- [x] 支持MySQL、SQLite、TiDB三种数据库类型
- [x] 实现数据源矩阵首页（按环境分组展示）
- [x] 实现数据源列表页（分页、筛选、搜索）
- [x] 实现新建/编辑表单弹窗（动态表单适配不同数据库类型）
- [x] 实现测试连接功能（支持未保存配置预览测试）
- [x] 实现最近使用数据源功能
- [x] 密码脱敏处理与安全管控

## Non-Goals (Out of Scope)
- 多租户、组织权限管理
- 网关接入、SSH隧道
- SSL高级加密、多节点容灾读写分离
- 企业级复杂能力

## Background & Context
- 后端采用Handler→Service→Store→Adapter→Model四层架构
- 前端使用ProComponents组件库
- 支持Gin框架，使用gorm进行数据库操作
- 现有项目已支持MySQL、SQLite、TiDB数据库连接

## Functional Requirements
- **FR-1**: 数据源矩阵展示 - 按开发、测试、预发、生产四个环境分组展示数据源卡片
- **FR-2**: 数据源列表管理 - 支持分页、名称模糊搜索、类型筛选
- **FR-3**: 数据源详情查看 - 返回完整配置信息，密码脱敏
- **FR-4**: 数据源创建 - 支持MySQL、SQLite、TiDB三种类型，自动填充默认端口
- **FR-5**: 数据源更新 - 支持部分字段更新，密码可选择更新
- **FR-6**: 数据源删除 - 支持删除及关联资源清理
- **FR-7**: 测试连接 - 支持已保存数据源和未保存配置的连接测试
- **FR-8**: 最近使用 - 记录并返回最近使用的数据源列表

## Non-Functional Requirements
- **NFR-1**: 密码字段脱敏处理，仅展示首尾各2个字符
- **NFR-2**: 测试连接超时时间5秒
- **NFR-3**: 错误信息长度限制1024字符
- **NFR-4**: 分页最大100条限制

## Constraints
- **Technical**: Go 1.21, Gin 1.9.1, GORM 1.30.0
- **Dependencies**: go-sql-driver/mysql, go-sqlite3
- **Database**: SQLite本地文件存储

## Assumptions
- 单用户场景，ownerId固定默认值
- 轻量版本不支持多组织，orgId置空
- 连接池按数据源ID独立缓存

## Acceptance Criteria

### AC-1: 数据源矩阵获取
- **Given**: 系统存在多个数据源，分布在不同环境
- **When**: 调用GET /api/datasource/matrix
- **Then**: 返回包含四个环境分组的数组，每个分组最多6条数据源
- **Verification**: `programmatic`

### AC-2: 数据源列表分页
- **Given**: 系统存在多个数据源
- **When**: 调用GET /api/datasource/listDatasource?keyword=xxx&type=mysql&current=1&pageSize=10
- **Then**: 返回分页数据，包含total、当前页列表，密码字段脱敏
- **Verification**: `programmatic`

### AC-3: 数据源详情获取
- **Given**: 存在ID为ds123的数据源
- **When**: 调用GET /api/datasource/ds123/datasourceInfo
- **Then**: 返回完整数据源配置，密码脱敏展示
- **Verification**: `programmatic`

### AC-4: 创建数据源
- **Given**: 提交合法的数据源配置
- **When**: 调用POST /api/datasource/createDatasource
- **Then**: 创建成功并返回完整数据源信息，后台异步测试连接
- **Verification**: `programmatic`

### AC-5: 更新数据源
- **Given**: 存在ID为ds123的数据源
- **When**: 调用POST /api/datasource/ds123/updateDatasource
- **Then**: 更新指定字段，updateTime更新为当前时间
- **Verification**: `programmatic`

### AC-6: 删除数据源
- **Given**: 存在ID为ds123的数据源
- **When**: 调用POST /api/datasource/ds123/deleteDatasource
- **Then**: 删除成功，清理关联连接池
- **Verification**: `programmatic`

### AC-7: 测试连接
- **Given**: 提交有效的数据库连接配置
- **When**: 调用POST /api/datasource/testConnection
- **Then**: 返回连接状态、数据库版本、连接耗时
- **Verification**: `programmatic`

### AC-8: 最近使用数据源
- **Given**: 用户使用过多个数据源
- **When**: 调用GET /api/datasource/listRecentlyDatasource?limit=8
- **Then**: 返回最近使用的前N条数据源
- **Verification**: `programmatic`

## Open Questions
- [ ] 暂无