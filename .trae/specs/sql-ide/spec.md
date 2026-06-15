# dbm-lite SQL IDE - Product Requirement Document

## Overview
- **Summary**: 基于dbm-lite现有架构增量开发SQL IDE模块，严格对标Navicat 17社区版，仅支持MySQL与TiDB两大数据库，复刻其界面布局、交互逻辑、操作体验。
- **Purpose**: 提供专业级数据库开发环境，支持SQL编辑、数据浏览、对象管理等核心功能，保持与Navicat社区版一致的用户体验。
- **Target Users**: 数据库开发人员、运维人员、数据分析师

## Goals
- [x] 复刻Navicat 17社区版整体界面布局与视觉风格
- [x] 实现完整的SQL编辑器（Monaco Editor内核）
- [x] 实现功能完备的数据编辑器（网格视图、表单视图）
- [x] 支持SQL执行与结果展示（含执行计划）
- [x] 实现数据库对象设计器（表设计器、视图设计）
- [x] 适配MySQL与TiDB差异化特性

## Non-Goals (Out of Scope)
- [ ] 不支持BI功能、数据建模、高级调试、跨库同步
- [ ] 不改动原有SQLite相关能力
- [ ] 不实现存储过程、触发器、函数设计器
- [ ] 不支持企业版专属数据分析模块

## Background & Context
- 基于现有dbm-lite项目增量开发，不改动原有架构
- 遵循项目统一编码规范、分层架构与接口标准
- 仅支持MySQL与TiDB，剔除SQLite相关的对象树差异处理

## Functional Requirements

### FR-1: 整体界面布局
- 全局菜单栏（文件、编辑、查看、工具、窗口、帮助）
- 快捷工具栏（新建查询、设计表、刷新、执行、执行计划、格式化SQL、保存、历史记录）
- 左侧导航对象树
- 右侧多标签工作区
- 底部状态栏

### FR-2: 左侧导航对象树
- 树状层级结构（连接→数据库→表/视图文件夹→列/索引）
- MySQL/TiDB差异化节点展示
- 全局搜索功能
- 对象交互（双击打开、拖拽分组、颜色标签）
- 右键菜单（按节点类型区分）
- 系统库显示开关

### FR-3: SQL编辑器
- Monaco Editor内核
- 语法高亮（MySQL/TiDB差异化）
- 智能代码补全
- 代码折叠、括号匹配、行高亮
- 查找替换、撤销重做、剪贴板历史
- SQL语句分割与执行
- 代码格式化
- 视图模式切换（普通/原始数据）
- 快捷键体系

### FR-4: 数据编辑器
- 网格视图与表单视图切换
- 列操作（调整宽度、拖拽排序、冻结、显示/隐藏）
- 单元格编辑（按字段类型差异化适配）
- 行操作（新增、删除、多选）
- 排序与筛选
- 数据复制导出
- 分页与导航
- 事务编辑模式

### FR-5: SQL执行与结果展示
- 执行按钮组（执行当前/全部、停止、Explain）
- 结果标签页管理
- 执行计划展示
- 错误定位与提示

### FR-6: 数据库对象设计器
- **表设计器**：字段页（字段名、数据类型、长度精度、是否可为空、默认值、主键、自增、无符号、字段注释）、索引页（索引名、类型、关联字段、升降序、注释）
- **MySQL外键标签页**（TiDB隐藏）
- **视图设计器**：可视化构建区与SQL编辑区双向同步，支持拖拽表、选择字段、配置关联、筛选、分组、排序
- **实时DDL预览**：语法高亮，支持手动微调，保存前完整性校验
- **表维护功能**：分析表、检查表、优化表、获取总行数（TiDB移除修复表功能）

### FR-7: SQL窗口Tab管理
- 每个Tab独立存储：数据源ID、当前数据库、SQL内容、光标位置、查询结果、修改标记
- Tab标题默认"新查询"，支持右键重命名，未保存内容标记星号
- 新建Tab、关闭Tab、切换Tab快捷键与交互保持统一
- 所有Tab内容、顺序、状态本地持久化，重启自动恢复
- 2秒防抖自动保存，失焦同步触发保存

### FR-8: 全局设置面板
- **编辑器设置**：字体大小、字体、Tab宽度、自动换行、括号配对、代码折叠、语法高亮、剪贴板历史条数、撤销栈深度
- **数据编辑器设置**：默认分页行数、可选分页值、NULL样式、日期格式、千位分隔符、列宽规则、行高预设、默认视图
- **执行安全设置**：高危语句拦截开关、执行超时时间、自动事务开关、执行后自动滚动至结果
- **外观设置**：深浅主题、对象树默认宽度、图标样式、按钮尺寸

### FR-9: 错误处理与全局交互
- Toast轻量提示（3秒自动消失，成功绿色、警告黄色、错误红色）
- 高危操作模态确认弹窗（文案样式复刻Navicat）
- 加载状态（spinner图标、骨架屏、进度文本、终止按钮）
- 连接失败弹窗（错误信息+编辑连接入口）
- SQL语法错误定位行号
- 网络状态检测（中断锁定功能、恢复后自动重试）

### FR-10: MySQL与TiDB差异化适配
- **对象树差异**：TiDB隐藏存储过程、触发器、外键节点，屏蔽metrics_schema系统库
- **语法差异**：TiDB禁用DELIMITER、存储过程相关语法，编辑器移除对应补全与入口
- **索引类型差异**：TiDB仅保留常规索引、唯一索引（隐藏全文索引、空间索引）
- **表维护差异**：TiDB移除REPAIR TABLE功能
- **端口差异**：MySQL默认3306端口、TiDB默认4000端口
- **执行计划与错误提示**：分别沿用对应数据库原生内容

### FR-11: 分层架构与通用代码规则
- **四层调用链**：Handler→Service/Store→Adapter→model，禁止反向依赖
- **Handler层**：仅处理HTTP请求解析、参数校验、响应包装
- **Store层**：统一管理内存数据、SQL窗口、执行记录、数据库元数据
- **Adapter层**：工厂模式实现DBAdapter接口，MySQL与TiDB复用基础逻辑
- **DBAdapter接口方法**：测试连接、获取库列表、获取表列表、获取列信息、执行查询、执行写语句、执行计划、获取表DDL、表维护
- **连接池配置**：按dsId缓存，最大连接10、空闲连接5、连接生命周期1小时
- **API响应格式**：统一包含success、data、message、total、current、pageSize字段

## Non-Functional Requirements
- **NFR-1**: 性能优化，大数据量表使用虚拟滚动
- **NFR-2**: 支持浅色/深色主题切换
- **NFR-3**: 配置项本地持久化（主题、窗口状态等）
- **NFR-4**: 高危SQL语句前置拦截与二次确认

## Constraints
- **Technical**: Go后端 + Vue前端技术栈，遵循现有分层架构
- **Business**: 仅支持MySQL与TiDB，不扩展其他数据库
- **Dependencies**: Monaco Editor、Gin框架、GORM

## Assumptions
- [x] 用户熟悉Navicat社区版操作习惯
- [x] 后端API接口已就绪，前端只需调用
- [x] 项目已引入Monaco Editor依赖

## Acceptance Criteria

### AC-1: 界面布局复刻
- **Given**: 用户启动dbm-lite应用
- **When**: 进入SQL IDE模块
- **Then**: 界面布局与Navicat 17社区版一致，包含菜单栏、工具栏、对象树、多标签工作区、状态栏
- **Verification**: `human-judgment`

### AC-2: 对象树差异化展示
- **Given**: 已配置MySQL和TiDB连接
- **When**: 展开对象树
- **Then**: MySQL显示存储过程、触发器文件夹，TiDB隐藏这些节点
- **Verification**: `programmatic`

### AC-3: SQL语法高亮差异化
- **Given**: 编辑SQL文件
- **When**: 输入TiDB专属关键字
- **Then**: 正确高亮显示，MySQL不兼容语法自动屏蔽
- **Verification**: `programmatic`

### AC-4: 高危SQL拦截
- **Given**: 编辑无WHERE条件的DELETE/UPDATE语句
- **When**: 执行该语句
- **Then**: 弹出红色风险警告，必须确认后方可执行
- **Verification**: `programmatic`

### AC-5: 数据编辑器字段类型适配
- **Given**: 打开数据编辑器
- **When**: 编辑不同类型字段（DATE、ENUM、JSON等）
- **Then**: 显示对应编辑器（日历、下拉框、JSON格式化器）
- **Verification**: `programmatic`

### AC-6: 事务编辑模式
- **Given**: 在数据编辑器中修改数据
- **When**: 点击提交事务按钮
- **Then**: 预览DML语句并批量提交，失败自动回滚
- **Verification**: `programmatic`

### AC-7: 执行计划展示
- **Given**: 执行EXPLAIN语句
- **When**: 查看执行计划结果
- **Then**: 显示全表扫描标红、Using filesort标黄
- **Verification**: `programmatic`

### AC-8: 主题切换
- **Given**: 在设置中切换主题
- **When**: 选择深色主题
- **Then**: 界面全局切换为深色样式，设置持久化
- **Verification**: `programmatic`

## Open Questions
- [ ] Monaco Editor与现有前端技术栈的集成方案
- [ ] 对象树懒加载策略与性能优化
- [ ] 事务模式下的本地缓存实现