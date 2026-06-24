import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/Login.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '首页概览', icon: 'DataLine' }
      },
      // ===== DB 操作管控域 =====
      {
        path: 'sql/sqlide',
        name: 'sql-sqlide',
        component: () => import('@/sql-ide/pages/SqlidePage.vue'),
        meta: { title: 'SQL IDE', icon: 'EditPen', domain: 'ops' }
      },
      {
        path: 'sql/table/:datasourceId/:database/:table',
        name: 'sql-table-viewer',
        component: () => import('@/sql-ide/pages/TableViewerPage.vue'),
        meta: { title: '表数据', icon: 'Table', domain: 'ops', hideMenu: true }
      },
      {
        path: 'datasources',
        name: 'datasources',
        component: () => import('@/views/datasource/List.vue'),
        meta: { title: '数据源', icon: 'Document', domain: 'ops' }
      },
      {
        path: 'datasource',
        redirect: '/datasources'
      },
      {
        path: 'datasources/:id',
        name: 'datasource-detail',
        component: () => import('@/views/datasource/Detail.vue'),
        meta: { title: '数据源详情', icon: 'Document', domain: 'ops', hideMenu: true }
      },
      {
        path: 'priv/groups',
        name: 'priv-groups',
        component: () => import('@/views/priv/ObjectPrivilege.vue'),
        meta: { title: '对象权限', icon: 'Lock', domain: 'ops' }
      },
      {
        path: 'priv/grant',
        name: 'priv-grant',
        component: () => import('@/views/priv/ObjectPrivilege.vue'),
        meta: { title: '权限分配', icon: 'Key', domain: 'ops', adminOnly: true, hideMenu: true }
      },
      {
        path: 'priv/list',
        name: 'priv-list',
        component: () => import('@/views/priv/ObjectPrivilege.vue'),
        meta: { title: '权限清单', icon: 'Tickets', domain: 'ops', hideMenu: true }
      },
      {
        path: 'priv/audit',
        name: 'priv-audit',
        component: () => import('@/views/priv/Audit.vue'),
        meta: { title: '权限审计', icon: 'DocumentChecked', domain: 'ops', adminOnly: true, hideMenu: true }
      },
      {
        path: 'priv/sensitive',
        name: 'priv-sensitive',
        component: () => import('@/views/priv/SensitiveColumns.vue'),
        meta: { title: '敏感列配置', icon: 'Lock', domain: 'ops', adminOnly: true, hideMenu: true }
      },
      {
        path: 'db-users',
        redirect: '/priv/groups'
      },
      {
        path: 'import-export',
        name: 'import-export',
        component: () => import('@/views/ops/ImportExport.vue'),
        meta: { title: '导入导出', icon: 'Upload', domain: 'ops' }
      },
      {
        path: 'sql-audit',
        name: 'sql-audit',
        component: () => import('@/views/ops/SqlAudit.vue'),
        meta: { title: 'SQL审核', icon: 'List', domain: 'ops' }
      },
      {
        path: 'sensitive-data',
        name: 'sensitive-data',
        component: () => import('@/views/ops/SensitiveData.vue'),
        meta: { title: '敏感数据', icon: 'Lock', domain: 'ops', adminOnly: true }
      },
      // ===== DB 基础运维域 =====
      {
        path: 'db-lifecycle',
        name: 'db-lifecycle',
        component: () => import('@/views/ops/DbLifecycle.vue'),
        meta: { title: 'DB生命周期', icon: 'Coin', domain: 'ops2' }
      },
      {
        path: 'servers',
        name: 'servers',
        component: () => import('@/views/server/List.vue'),
        meta: { title: '服务器', icon: 'Monitor', domain: 'ops2', adminOnly: true }
      },
      {
        path: 'health-check',
        name: 'health-check',
        component: () => import('@/views/ops/HealthCheck.vue'),
        meta: { title: '健康巡检', icon: 'Histogram', domain: 'ops2' }
      },
      {
        path: 'data-migration',
        name: 'data-migration',
        component: () => import('@/views/ops/DataMigration.vue'),
        meta: { title: '数据迁移', icon: 'RefreshRight', domain: 'ops2' }
      },
      // ===== 底部固定菜单 =====
      {
        path: 'businesses',
        name: 'businesses',
        component: () => import('@/views/business/List.vue'),
        meta: { title: '业务管理', icon: 'Folder', domain: 'bottom' }
      },
      {
        path: 'workflow/:tab?',
        name: 'workflow',
        component: () => import('@/views/workflow/WorkflowCenter.vue'),
        meta: { title: '审批流程', icon: 'DocumentChecked', domain: 'bottom' }
      },
      {
        path: 'platform-config',
        name: 'platform-config',
        component: () => import('@/views/platform/PlatformConfig.vue'),
        meta: { title: '平台配置', icon: 'Management', domain: 'bottom', adminOnly: true }
      },
      {
        path: 'platform-config/medium',
        name: 'platform-medium',
        component: () => import('@/views/platform/Medium.vue'),
        meta: { title: '介质维护', icon: 'Tools', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/plugins',
        name: 'platform-plugin',
        component: () => import('@/views/plugin/List.vue'),
        meta: { title: '插件管理', icon: 'MagicStick', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/account',
        name: 'platform-account',
        component: () => import('@/views/account/List.vue'),
        meta: { title: '账号管理', icon: 'User', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/roles',
        name: 'platform-roles',
        component: () => import('@/views/platform/RoleManager.vue'),
        meta: { title: '角色管理', icon: 'UserFilled', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/posts',
        name: 'platform-posts',
        component: () => import('@/views/platform/PostManager.vue'),
        meta: { title: '岗位管理', icon: 'Film', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/departments',
        name: 'platform-departments',
        component: () => import('@/views/platform/DepartmentManager.vue'),
        meta: { title: '部门管理', icon: 'OfficeBuilding', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/menus',
        name: 'platform-menus',
        component: () => import('@/views/platform/MenuManager.vue'),
        meta: { title: '菜单管理', icon: 'Menu', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/dictionaries',
        name: 'platform-dictionaries',
        component: () => import('@/views/platform/DictionaryManager.vue'),
        meta: { title: '字典管理', icon: 'Collection', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/parameters',
        name: 'platform-parameters',
        component: () => import('@/views/platform/ParameterManager.vue'),
        meta: { title: '参数配置', icon: 'Setting', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'platform-config/system-settings',
        name: 'platform-system-settings',
        component: () => import('@/views/platform/SystemSettings.vue'),
        meta: { title: '系统配置', icon: 'Setting', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'process-manager',
        name: 'process-manager',
        component: () => import('@/views/workflow/ProcessManager.vue'),
        meta: { title: '流程管理', icon: 'DocumentChecked', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'classify-manager',
        name: 'classify-manager',
        component: () => import('@/views/workflow/ClassifyManager.vue'),
        meta: { title: '流程分类管理', icon: 'Collection', domain: 'bottom', adminOnly: true, hideMenu: true }
      },
      {
        path: 'plugins',
        redirect: '/platform-config/plugins'
      },
      {
        path: 'accounts',
        redirect: '/platform-config/account'
      },
      {
        path: 'audit',
        name: 'audit',
        component: () => import('@/views/audit/List.vue'),
        meta: { title: '操作审计', icon: 'Document', domain: 'bottom', adminOnly: true }
      },
      {
        path: 'profile',
        name: 'profile',
        component: () => import('@/views/Profile.vue'),
        meta: { title: '个人中心', icon: 'Setting', domain: 'bottom' }
      }
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  if (to.meta.public) {
    next()
    return
  }
  if (!userStore.isLoggedIn) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }
  if (to.meta.adminOnly && !userStore.isAdmin) {
    next('/dashboard')
    return
  }
  next()
})

export default router
