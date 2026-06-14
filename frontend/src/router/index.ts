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
    redirect: '/sql/workbench',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '首页概览', icon: 'DataLine' }
      },
      {
        path: 'sql/workbench',
        name: 'sql-workbench',
        component: () => import('@/views/sql/Workbench.vue'),
        meta: { title: 'SQL IDE', icon: 'EditPen', domain: 'ops' }
      },
      {
        path: 'datasources',
        name: 'datasources',
        component: () => import('@/views/datasource/List.vue'),
        meta: { title: '数据源管理', icon: 'Document', domain: 'ops' }
      },
      {
        path: 'db-users',
        name: 'db-users',
        component: () => import('@/views/dbuser/List.vue'),
        meta: { title: '对象权限管理', icon: 'Lock', domain: 'ops' }
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
        path: 'db-lifecycle',
        name: 'db-lifecycle',
        component: () => import('@/views/backup/List.vue'),
        meta: { title: 'DB生命周期管理', icon: 'Coin', domain: 'ops2', demoOnly: true, demoTitle: 'DB生命周期管理' }
      },
      {
        path: 'servers',
        name: 'servers',
        component: () => import('@/views/server/List.vue'),
        meta: { title: '服务器管理', icon: 'Monitor', domain: 'ops2', adminOnly: true }
      },
      {
        path: 'plugins',
        name: 'plugins',
        component: () => import('@/views/plugin/List.vue'),
        meta: { title: '插件管理', icon: 'MagicStick', domain: 'ops2', demoOnly: true, demoTitle: '插件管理', adminOnly: true }
      },
      {
        path: 'businesses',
        name: 'businesses',
        component: () => import('@/views/business/List.vue'),
        meta: { title: '业务管理', icon: 'Folder', domain: 'bottom' }
      },
      {
        path: 'accounts',
        name: 'accounts',
        component: () => import('@/views/account/List.vue'),
        meta: { title: '平台账号管理', icon: 'User', domain: 'bottom', adminOnly: true }
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
  { path: '/:pathMatch(.*)*', redirect: '/sql/workbench' }
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
    next('/sql/workbench')
    return
  }
  next()
})

export default router
