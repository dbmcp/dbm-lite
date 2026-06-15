# dbm-lite 数据源管理模块 - 实现计划

## [x] Task 1: 创建数据源管理Handler层
- **Priority**: P0
- **Depends On**: None
- **Description**: 
  - 创建 `internal/handler/datasource_v2.go` 文件
  - 实现8个API接口：matrix、listDatasource、datasourceInfo、createDatasource、updateDatasource、deleteDatasource、testConnection、listRecentlyDatasource
  - 遵循现有Handler层编码规范
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6, AC-7, AC-8
- **Test Requirements**:
  - `programmatic` TR-1.1: 所有API接口返回正确的HTTP状态码
  - `programmatic` TR-1.2: 密码字段返回脱敏格式
  - `human-judgement` TR-1.3: 代码结构清晰，符合项目编码规范

## [x] Task 2: 创建数据源管理Service层
- **Priority**: P0
- **Depends On**: Task 1
- **Description**: 
  - 创建 `internal/service/datasource_v2.go` 文件
  - 实现业务逻辑：矩阵分组、分页查询、CRUD操作、测试连接、最近使用记录
  - 实现密码脱敏方法
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6, AC-7, AC-8
- **Test Requirements**:
  - `programmatic` TR-2.1: 矩阵分组正确，每组最多6条
  - `programmatic` TR-2.2: 分页逻辑正确，pageSize最大100
  - `programmatic` TR-2.3: 密码脱敏格式正确（首尾各2字符，中间星号）

## [x] Task 3: 更新数据模型添加必要字段
- **Priority**: P0
- **Depends On**: None
- **Description**: 
  - 更新 `internal/model/datasource.go` 添加 ownerId、orgId、type、datasourceType、lastUseTime字段
  - 添加 connectStatus 枚举值（success、failed、unknown、connecting）
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6, AC-7, AC-8
- **Test Requirements**:
  - `programmatic` TR-3.1: 新增字段正确映射到数据库表
  - `human-judgement` TR-3.2: 枚举值定义清晰

## [x] Task 4: 注册API路由
- **Priority**: P0
- **Depends On**: Task 1
- **Description**: 
  - 更新 `cmd/server/main.go` 注册新的数据源管理API路由
  - 添加 /api/datasource/matrix、/api/datasource/listDatasource、/api/datasource/testConnection 等路由
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6, AC-7, AC-8
- **Test Requirements**:
  - `programmatic` TR-4.1: 所有路由可正常访问
  - `programmatic` TR-4.2: 路由权限控制正确

## [x] Task 5: 更新前端API接口文件
- **Priority**: P1
- **Depends On**: Task 4
- **Description**: 
  - 更新 `frontend/src/api/datasource.ts` 添加新API接口
  - 定义接口参数和返回类型
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6, AC-7, AC-8
- **Test Requirements**:
  - `human-judgement` TR-5.1: API类型定义完整准确

## [x] Task 6: 创建数据源矩阵首页组件
- **Priority**: P1
- **Depends On**: Task 5
- **Description**: 
  - 创建 `frontend/src/views/datasource/Matrix.vue` 组件
  - 实现四列环境卡片布局，展示数据源卡片
  - 实现数据源卡片悬浮操作按钮
- **Acceptance Criteria Addressed**: AC-1
- **Test Requirements**:
  - `human-judgement` TR-6.1: 页面布局符合设计规范
  - `human-judgement` TR-6.2: 交互逻辑正确

## [x] Task 7: 更新数据源列表页
- **Priority**: P1
- **Depends On**: Task 5
- **Description**: 
  - 更新 `frontend/src/views/datasource/List.vue`
  - 添加筛选工具栏（名称搜索、类型筛选、环境筛选）
  - 更新表格列展示
- **Acceptance Criteria Addressed**: AC-2
- **Test Requirements**:
  - `human-judgement` TR-7.1: 筛选功能正常工作
  - `human-judgement` TR-7.2: 分页功能正常

## [x] Task 8: 创建数据源编辑弹窗组件
- **Priority**: P1
- **Depends On**: Task 5
- **Description**: 
  - 更新 `frontend/src/views/datasource/components/EditDialog.vue`
  - 实现动态表单：切换数据库类型时自动适配显示字段
  - 添加测试连接按钮
- **Acceptance Criteria Addressed**: AC-4, AC-5, AC-7
- **Test Requirements**:
  - `human-judgement` TR-8.1: 动态表单切换正确
  - `human-judgement` TR-8.2: 测试连接功能正常

## [x] Task 9: 编译验证
- **Priority**: P0
- **Depends On**: Task 1-4
- **Description**: 
  - 编译后端代码，确保无编译错误
  - 运行服务验证API接口正常
- **Acceptance Criteria Addressed**: 所有AC
- **Test Requirements**:
  - `programmatic` TR-9.1: 编译成功
  - `programmatic` TR-9.2: API接口返回正确格式