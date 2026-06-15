# dbm-lite SQL IDE - Implementation Plan

## [ ] Task 1: 界面布局与全局组件
- **Priority**: P0
- **Depends On**: None
- **Description**: 
  - 实现全局菜单栏（文件、编辑、查看、工具、窗口、帮助）
  - 实现快捷工具栏（新建查询、设计表、刷新、执行、执行计划、格式化SQL、保存、历史记录）
  - 实现底部状态栏（数据库类型、执行状态、耗时、行数）
  - 实现主题切换功能（浅色/深色）
- **Acceptance Criteria Addressed**: AC-1, AC-8
- **Test Requirements**:
  - `programmatic` TR-1.1: 主题切换后界面样式正确更新，刷新页面保持设置
  - `programmatic` TR-1.2: 状态栏实时显示执行状态和耗时
  - `human-judgment` TR-1.3: 界面布局与Navicat 17社区版视觉一致性
- **Notes**: 参考Navicat 17社区版的菜单结构和工具栏按钮排列

## [ ] Task 2: 左侧导航对象树
- **Priority**: P0
- **Depends On**: Task 1
- **Description**: 
  - 实现树状层级结构（连接→数据库→表/视图→列/索引）
  - MySQL/TiDB差异化节点展示（TiDB隐藏存储过程、触发器）
  - 全局搜索功能（大小写不敏感模糊检索）
  - 对象拖拽分组、颜色标签功能
  - 右键菜单（按节点类型区分）
  - 系统库显示开关
- **Acceptance Criteria Addressed**: AC-1, AC-2
- **Test Requirements**:
  - `programmatic` TR-2.1: MySQL连接显示存储过程节点，TiDB连接不显示
  - `programmatic` TR-2.2: 搜索功能正确定位并展开匹配节点
  - `programmatic` TR-2.3: 系统库开关正确显示/隐藏系统库
  - `human-judgment` TR-2.4: 右键菜单选项与Navicat一致
- **Notes**: 系统库列表：MySQL隐藏information_schema、mysql、performance_schema、sys；TiDB额外隐藏metrics_schema

## [ ] Task 3: SQL编辑器核心功能
- **Priority**: P0
- **Depends On**: Task 1
- **Description**: 
  - 集成Monaco Editor作为编辑器内核
  - 实现语法高亮（MySQL/TiDB差异化）
  - 实现智能代码补全（表名、字段、关键字、函数）
  - 实现代码折叠、括号匹配、当前行高亮
  - 实现查找替换（支持正则、大小写匹配）
  - 实现撤销重做、剪贴板历史
- **Acceptance Criteria Addressed**: AC-1, AC-3
- **Test Requirements**:
  - `programmatic` TR-3.1: MySQL关键字与TiDB专属关键字正确高亮
  - `programmatic` TR-3.2: 代码补全正确显示当前库表名和字段
  - `programmatic` TR-3.3: 括号配对高亮正确识别跨行匹配
  - `human-judgment` TR-3.4: 编辑器视觉风格与Navicat一致
- **Notes**: 参考Navicat的语法高亮配色方案

## [ ] Task 4: SQL编辑器进阶功能
- **Priority**: P1
- **Depends On**: Task 3
- **Description**: 
  - SQL语句智能分割（分号识别、DELIMITER支持）
  - 代码格式化（MySQL/TiDB两套规则）
  - 视图模式切换（普通/原始数据）
  - 代码大纲功能
  - 完整快捷键体系
- **Acceptance Criteria Addressed**: AC-1, AC-3, AC-4
- **Test Requirements**:
  - `programmatic` TR-4.1: 多条SQL正确分割并独立执行
  - `programmatic` TR-4.2: 代码格式化保持原有业务逻辑不变
  - `programmatic` TR-4.3: 快捷键功能正常工作
  - `programmatic` TR-4.4: DELIMITER功能仅对MySQL生效
- **Notes**: TiDB环境自动隐藏DELIMITER功能入口

## [ ] Task 5: 数据编辑器网格视图
- **Priority**: P0
- **Depends On**: Task 1
- **Description**: 
  - 网格视图展示数据
  - 列操作（调整宽度、拖拽排序、冻结、显示/隐藏）
  - 单元格编辑（按字段类型差异化适配）
  - 行操作（新增、删除、多选）
  - 排序与筛选功能
- **Acceptance Criteria Addressed**: AC-5
- **Test Requirements**:
  - `programmatic` TR-5.1: 不同字段类型显示对应编辑器（DATE显示日历、ENUM显示下拉）
  - `programmatic` TR-5.2: 列冻结功能正常工作
  - `programmatic` TR-5.3: NULL值显示为灰色斜体
  - `human-judgment` TR-5.4: 单元格编辑体验与Navicat一致
- **Notes**: 大数据量表使用虚拟滚动优化性能

## [ ] Task 6: 数据编辑器表单视图与事务模式
- **Priority**: P1
- **Depends On**: Task 5
- **Description**: 
  - 表单视图（单条记录全屏展示）
  - 事务编辑模式（开始/提交/回滚）
  - 数据复制导出（CSV、TXT、SQL脚本）
  - 分页与导航逻辑
- **Acceptance Criteria Addressed**: AC-5, AC-6
- **Test Requirements**:
  - `programmatic` TR-6.1: 网格视图与表单视图切换正常
  - `programmatic` TR-6.2: 事务提交失败自动回滚
  - `programmatic` TR-6.3: 分页导航正确显示数据范围
  - `human-judgment` TR-6.4: 事务预览弹窗与Navicat一致
- **Notes**: 默认每页300行，支持100、500、1000行切换

## [ ] Task 7: SQL执行与结果展示
- **Priority**: P0
- **Depends On**: Task 3
- **Description**: 
  - 执行按钮组（执行当前/全部、停止、Explain）
  - 结果标签页管理（多SQL多结果集）
  - 执行计划展示（EXPLAIN、EXPLAIN ANALYZE）
  - 错误定位与提示
- **Acceptance Criteria Addressed**: AC-7
- **Test Requirements**:
  - `programmatic` TR-7.1: 单条/全部SQL正确执行并展示结果
  - `programmatic` TR-7.2: EXPLAIN结果全表扫描标红、Using filesort标黄
  - `programmatic` TR-7.3: 执行错误正确定位行号
  - `human-judgment` TR-7.4: 执行状态与Navicat一致
- **Notes**: EXPLAIN ANALYZE开关默认关闭

## [ ] Task 8: 数据库对象设计器
- **Priority**: P1
- **Depends On**: Task 1
- **Description**: 
  - 表设计器字段页（字段名、数据类型、长度精度、是否可为空、默认值、主键、自增、无符号、字段注释）
  - 表设计器索引页（索引名、类型、关联字段、升降序、注释，TiDB隐藏全文索引、空间索引）
  - MySQL外键标签页（TiDB隐藏）
  - 视图设计器（可视化构建区与SQL编辑区双向同步）
  - 实时DDL预览（语法高亮、手动微调、完整性校验）
  - 表维护功能（分析表、检查表、优化表、获取总行数，TiDB移除修复表）
- **Acceptance Criteria Addressed**: AC-1, FR-6
- **Test Requirements**:
  - `programmatic` TR-8.1: TiDB连接不显示外键标签页
  - `programmatic` TR-8.2: 字段类型下拉过滤TiDB不兼容选项
  - `programmatic` TR-8.3: 实时DDL预览正确生成
  - `programmatic` TR-8.4: TiDB隐藏全文索引、空间索引选项
  - `programmatic` TR-8.5: 视图设计器可视化与SQL双向同步
  - `human-judgment` TR-8.6: 设计器布局与Navicat一致
- **Notes**: 剔除存储过程、函数设计等企业功能

## [ ] Task 9: 高危SQL安全管控
- **Priority**: P1
- **Depends On**: Task 3
- **Description**: 
  - 语法预检测与错误提示
  - 高危语句前置拦截（无WHERE的DELETE/UPDATE/DROP/TRUNCATE）
  - 拦截开关配置
- **Acceptance Criteria Addressed**: AC-4
- **Test Requirements**:
  - `programmatic` TR-9.1: 无WHERE条件的DELETE语句触发警告弹窗
  - `programmatic` TR-9.2: 拦截开关可在设置中关闭
  - `human-judgment` TR-9.3: 警告弹窗样式与Navicat一致
- **Notes**: 红色风险警告弹窗必须勾选确认方可执行

## [ ] Task 10: 标签页管理
- **Priority**: P1
- **Depends On**: Task 1
- **Description**: 
  - 多标签页管理（SQL窗口、数据窗口、对象窗口）
  - 标签页拖拽排序
  - 右键菜单（重命名、复制、关闭当前/其他/全部标签）
  - 未保存内容关闭确认弹窗
- **Acceptance Criteria Addressed**: AC-1, FR-7
- **Test Requirements**:
  - `programmatic` TR-10.1: 标签页拖拽调整顺序正常
  - `programmatic` TR-10.2: 关闭未保存标签弹出确认弹窗
  - `human-judgment` TR-10.3: 右键菜单选项与Navicat一致
- **Notes**: 确认弹窗样式、提示文案复刻Navicat

## [ ] Task 11: SQL窗口Tab管理
- **Priority**: P1
- **Depends On**: Task 10
- **Description**: 
  - 每个Tab独立存储：数据源ID、当前数据库、SQL内容、光标位置、查询结果、修改标记
  - Tab标题默认"新查询"，支持右键重命名，未保存内容标记星号
  - 2秒防抖自动保存，失焦同步触发保存
  - 所有Tab内容、顺序、状态本地持久化，重启自动恢复
- **Acceptance Criteria Addressed**: FR-7
- **Test Requirements**:
  - `programmatic` TR-11.1: Tab内容正确持久化，重启后恢复
  - `programmatic` TR-11.2: 2秒防抖自动保存正常工作
  - `programmatic` TR-11.3: 未保存内容正确标记星号
- **Notes**: 持久化开关可手动关闭

## [ ] Task 12: 全局设置面板
- **Priority**: P2
- **Depends On**: Task 1
- **Description**: 
  - 编辑器设置（字体大小、字体、Tab宽度、自动换行、括号配对、代码折叠、语法高亮、剪贴板历史条数、撤销栈深度）
  - 数据编辑器设置（默认分页行数、可选分页值、NULL样式、日期格式、千位分隔符、列宽规则、行高预设、默认视图）
  - 执行安全设置（高危语句拦截开关、执行超时时间、自动事务开关、执行后自动滚动至结果）
  - 外观设置（深浅主题、对象树默认宽度、图标样式、按钮尺寸）
- **Acceptance Criteria Addressed**: FR-8
- **Test Requirements**:
  - `programmatic` TR-12.1: 所有设置正确保存并持久化
  - `programmatic` TR-12.2: 设置修改后立即生效
  - `human-judgment` TR-12.3: 设置面板布局与Navicat一致
- **Notes**: 所有设置本地持久化，支持一键恢复默认

## [ ] Task 13: 错误处理与全局交互
- **Priority**: P1
- **Depends On**: Task 1
- **Description**: 
  - Toast轻量提示（3秒自动消失，成功绿色、警告黄色、错误红色）
  - 高危操作模态确认弹窗
  - 加载状态（spinner图标、骨架屏、进度文本、终止按钮）
  - 连接失败弹窗（错误信息+编辑连接入口）
  - SQL语法错误定位行号
  - 网络状态检测（中断锁定功能、恢复后自动重试）
- **Acceptance Criteria Addressed**: FR-9
- **Test Requirements**:
  - `programmatic` TR-13.1: Toast提示正确显示并自动消失
  - `programmatic` TR-13.2: 网络中断时锁定执行功能
  - `programmatic` TR-13.3: SQL语法错误正确定位行号
  - `human-judgment` TR-13.4: 弹窗样式与Navicat一致
- **Notes**: 文案样式复刻Navicat

## [ ] Task 14: MySQL与TiDB差异化适配
- **Priority**: P1
- **Depends On**: Task 2, Task 3, Task 8
- **Description**: 
  - 对象树差异：TiDB隐藏存储过程、触发器、外键节点，屏蔽metrics_schema系统库
  - 语法差异：TiDB禁用DELIMITER、存储过程相关语法
  - 索引类型差异：TiDB仅保留常规索引、唯一索引
  - 表维护差异：TiDB移除REPAIR TABLE功能
  - 端口差异：MySQL默认3306端口、TiDB默认4000端口
- **Acceptance Criteria Addressed**: FR-10
- **Test Requirements**:
  - `programmatic` TR-14.1: TiDB连接不显示存储过程、触发器节点
  - `programmatic` TR-14.2: DELIMITER功能仅对MySQL生效
  - `programmatic` TR-14.3: TiDB隐藏全文索引、空间索引选项
  - `programmatic` TR-14.4: MySQL默认端口3306，TiDB默认端口4000
- **Notes**: 执行计划与错误提示分别沿用对应数据库原生内容

## [x] Task 15: 分层架构与Adapter层实现
- **Priority**: P0
- **Depends On**: None
- **Description**: 
  - 四层调用链：Handler→Service/Store→Adapter→model
  - Handler层：HTTP请求解析、参数校验、响应包装
  - Store层：内存数据、SQL窗口、执行记录、数据库元数据管理
  - Adapter层：工厂模式实现DBAdapter接口
  - DBAdapter接口九大数据方法实现
  - 连接池配置（按dsId缓存，最大连接10、空闲连接5、连接生命周期1小时）
- **Acceptance Criteria Addressed**: FR-11
- **Test Requirements**:
  - `programmatic` TR-15.1: DBAdapter接口方法正常工作
  - `programmatic` TR-15.2: 连接池按dsId正确缓存
  - `programmatic` TR-15.3: API响应格式统一（success、data、message、total、current、pageSize）
- **Notes**: MySQL与TiDB复用基础逻辑，差异化语法单独编写

---

## 分阶段实现规划

### 第一阶段：MVP
- [x] 基础界面布局（Task 1）- 已有后端API支持
- [x] 对象树基础加载（Task 2基础功能）- 支持数据库/表/视图/列/索引/存储过程/触发器
- [x] SQL编辑器基础语法高亮与执行（Task 3基础功能）- 已有execute/explain接口
- [x] 简单数据网格（Task 5基础功能）- 已有GetTables/GetColumns接口
- [x] 基础表设计（Task 8基础功能）- 新增GetTableDDL/GetTableInfo/GetColumns接口
- [x] MySQL与TiDB基础适配（Task 14基础功能）- 新增GetDatabaseTypes接口返回差异配置
- [x] 基础API对接与错误处理（Task 13基础功能）- 统一响应格式与错误处理

### 第二阶段：增强
- [ ] SQL编辑器全功能（Task 3完整功能、Task 4）
- [ ] 数据编辑器网格/表单视图、编辑、筛选、排序、事务、复制导出（Task 5完整功能、Task 6）
- [ ] Tab全量管理（Task 10、Task 11）
- [ ] 对象树完整右键菜单（Task 2完整功能）
- [ ] Explain执行计划（Task 7）
- [ ] 表维护（Task 8表维护功能）
- [ ] 深浅主题（Task 1外观设置）

### 第三阶段：完善
- [ ] MySQL与TiDB全量差异化适配（Task 14完整功能）
- [ ] 全局设置面板（Task 12）
- [ ] 网络异常处理（Task 13完整功能）
- [ ] 大数据虚拟滚动（Task 5优化）
- [ ] 全功能回归测试