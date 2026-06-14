<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DBA老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <el-container :class="['layout', 'layout-' + layoutMode]" style="height: 100vh;">
    <el-aside v-if="layoutMode === 'side'" :width="asideWidth" style="background:#001529;transition:width .3s;display:flex;flex-direction:column;">
      <div class="logo">
        <el-icon :size="24" color="#fff"><DataLine /></el-icon>
        <span v-if="!collapsed" style="margin-left:10px;color:#fff;font-weight:600;font-size:18px;">DBM-Lite</span>
      </div>

      <div style="flex:1;overflow-y:auto;">
        <el-menu
          :default-active="route.path"
          :collapse="collapsed"
          :unique-opened="true"
          background-color="#001529"
          text-color="#b4c0cc"
          :active-text-color="currentDomainColor"
          router
        >
          <template v-for="item in topMenu" :key="'top-' + item.index">
            <el-menu-item :index="item.index">
              <el-icon color="#ffffff"><component :is="item.icon" /></el-icon>
              <template #title><span style="color:#ffffff;">{{ item.title }}</span></template>
            </el-menu-item>
          </template>

          <el-menu-divider style="margin:6px 0;border-color:#1f2d3d;" />

          <template v-for="item in currentDomainMenu" :key="'d-' + item.index">
            <el-menu-item v-if="!item.visible || item.visible()" :index="item.index">
              <el-icon :color="currentDomainColor"><component :is="item.icon" /></el-icon>
              <template #title>
                <span :style="{ color: currentDomainColor }">{{ item.title }}</span>
              </template>
            </el-menu-item>
          </template>

          <el-menu-divider style="margin:6px 0;border-color:#1f2d3d;" />

          <template v-for="item in bottomMenu" :key="'btm-' + item.index">
            <el-menu-item v-if="!item.visible || item.visible()" :index="item.index">
              <el-icon color="#ffffff"><component :is="item.icon" /></el-icon>
              <template #title><span style="color:#ffffff;">{{ item.title }}</span></template>
            </el-menu-item>
          </template>
        </el-menu>
      </div>

      <div v-if="!collapsed" class="domain-tag" :style="{ background: currentDomainColor + '15', borderColor: currentDomainColor + '40' }">
        <span :style="{ color: currentDomainColor }">●</span>
        <span style="color:#b4c0cc;margin-left:6px;">{{ currentDomainName }}</span>
      </div>
    </el-aside>

    <el-container>
      <el-header style="background:#fff;border-bottom:1px solid #e4e7ed;display:flex;align-items:center;justify-content:space-between;padding:0 20px;">
        <div style="display:flex;align-items:center;gap:16px;">
          <el-button v-if="layoutMode === 'side'" link @click="collapsed = !collapsed">
            <el-icon :size="20"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
          </el-button>
          <span v-if="layoutMode === 'side'" style="font-size:16px;color:#303133;">{{ route.meta.title }}</span>

          <el-menu
            v-if="layoutMode === 'top'"
            :default-active="route.path"
            mode="horizontal"
            :menu-trigger="click"
            router
            style="border-bottom:0;flex:1;"
          >
            <template v-for="item in topMenu" :key="'top-' + item.index">
              <el-menu-item :index="item.index">
                <el-icon><component :is="item.icon" /></el-icon>
                <span style="margin-left:6px;">{{ item.title }}</span>
              </el-menu-item>
            </template>

            <el-sub-menu :index="'ops-group'" :popper-class="''">
              <template #title>
                <el-icon :color="opsColor"><EditPen /></el-icon>
                <span style="margin-left:6px;color:#303133;">DB操作管控域</span>
              </template>
              <el-menu-item v-for="item in opsDomainMenu.filter((m:any)=>!m.visible||m.visible())" :key="'top-ops-' + item.index" :index="item.index">
                <el-icon :color="opsColor"><component :is="item.icon" /></el-icon>
                <span style="margin-left:6px;">{{ item.title }}</span>
              </el-menu-item>
            </el-sub-menu>

            <el-sub-menu :index="'ops2-group'">
              <template #title>
                <el-icon :color="ops2Color"><Coin /></el-icon>
                <span style="margin-left:6px;color:#303133;">DB基础运维域</span>
              </template>
              <el-menu-item v-for="item in ops2DomainMenu.filter((m:any)=>!m.visible||m.visible())" :key="'top-ops2-' + item.index" :index="item.index">
                <el-icon :color="ops2Color"><component :is="item.icon" /></el-icon>
                <span style="margin-left:6px;">{{ item.title }}</span>
              </el-menu-item>
            </el-sub-menu>

            <el-menu-item v-for="item in bottomMenu.filter((m:any)=>!m.visible||m.visible())" :key="'top-btm-' + item.index" :index="item.index">
              <el-icon><component :is="item.icon" /></el-icon>
              <span style="margin-left:6px;">{{ item.title }}</span>
            </el-menu-item>
          </el-menu>
        </div>

        <div style="display:flex;align-items:center;gap:12px;">
          <el-tooltip content="切换布局" placement="bottom">
            <el-button link @click="toggleLayout">
              <el-icon :size="18"><Menu v-if="layoutMode === 'side'" /><Grid v-else /></el-icon>
            </el-button>
          </el-tooltip>

          <el-dropdown trigger="hover" @command="handleDomainCommand" placement="bottom" hide-on-click>
            <span :style="{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '8px', padding: '6px 12px', border: '1px solid', borderRadius: '6px', borderColor: currentDomainColor + '50', background: currentDomainColor + '10' }">
              <span :style="{ color: currentDomainColor }">●</span>
              <span :style="{ color: currentDomainColor, fontWeight: 500, fontSize: '13px' }">{{ currentDomainName }}</span>
              <el-icon :size="12" :color="currentDomainColor"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item :command="'ops'" :disabled="currentDomain === 'ops'">
                  <span :style="{ color: opsColor }">●</span>
                  <span style="margin-left:8px;">DB操作管控域</span>
                  <span v-if="currentDomain === 'ops'" style="margin-left:8px;color:#909399;font-size:12px;">当前</span>
                </el-dropdown-item>
                <el-dropdown-item :command="'ops2'" :disabled="currentDomain === 'ops2'">
                  <span :style="{ color: ops2Color }">●</span>
                  <span style="margin-left:8px;">DB基础运维域</span>
                  <span v-if="currentDomain === 'ops2'" style="margin-left:8px;color:#909399;font-size:12px;">当前</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown trigger="hover" @command="handleCommand" placement="bottom">
            <span style="cursor:pointer;display:flex;align-items:center;gap:8px;color:#606266;">
              <el-avatar :size="32" :style="{ background: currentDomainColor }">{{ userStore.displayName.charAt(0).toUpperCase() }}</el-avatar>
              <span>{{ userStore.displayName }}</span>
              <el-icon :size="12"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main v-if="layoutMode === 'top' && route.meta.title" style="padding:8px 20px 0;background:#f5f7fa;">
        <span style="font-size:16px;color:#303133;font-weight:500;">{{ route.meta.title }}</span>
      </el-main>

      <el-main style="padding:12px 20px;background:#f5f7fa;overflow:auto;">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>

      <div class="layout-footer">
        DBM-Lite v0.1.0 © DBA老王
      </div>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  DataLine, Fold, Expand, ArrowDown,
  EditPen, Coin, Folder, Monitor, MagicStick, User, Document, Setting, Menu, Grid, Lock
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const LAST_PAGE_KEY = 'dbm-lite-last-page'
const LAYOUT_KEY = 'dbm-lite-layout'
const DOMAIN_KEY = 'dbm-lite-domain'

type DomainType = 'ops' | 'ops2'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const collapsed = ref(false)
const asideWidth = computed(() => (collapsed.value ? '64px' : '220px'))

const layoutMode = ref<'side' | 'top'>((localStorage.getItem(LAYOUT_KEY) as 'side' | 'top') || 'side')

const opsColor = '#409eff'
const ops2Color = '#2e6ba8'

const currentDomain = ref<DomainType>((localStorage.getItem(DOMAIN_KEY) as DomainType) || 'ops')

const domainConfig = {
  ops: { name: 'DB操作管控域', color: opsColor },
  ops2: { name: 'DB基础运维域', color: ops2Color }
} as const

const currentDomainName = computed(() => domainConfig[currentDomain.value].name)
const currentDomainColor = computed(() => domainConfig[currentDomain.value].color)

interface MenuItem {
  index: string
  title: string
  icon: any
  visible?: () => boolean
}

const topMenu: MenuItem[] = [
  { index: '/dashboard', title: '首页概览', icon: DataLine }
]

const opsDomainMenu: MenuItem[] = [
  { index: '/sql/workbench', title: 'SQL IDE', icon: EditPen },
  { index: '/datasources', title: '数据源管理', icon: Document },
  { index: '/db-users', title: '对象权限管理', icon: Lock }
]

const ops2DomainMenu: MenuItem[] = [
  { index: '/db-lifecycle', title: 'DB生命周期管理', icon: Coin },
  { index: '/servers', title: '服务器管理', icon: Monitor, visible: () => userStore.isAdmin },
  { index: '/plugins', title: '插件管理', icon: MagicStick, visible: () => userStore.isAdmin }
]

const bottomMenu: MenuItem[] = [
  { index: '/businesses', title: '业务管理', icon: Folder },
  { index: '/accounts', title: '平台账号管理', icon: User, visible: () => userStore.isAdmin },
  { index: '/audit', title: '操作审计', icon: Document, visible: () => userStore.isAdmin },
  { index: '/profile', title: '个人中心', icon: Setting }
]

const currentDomainMenu = computed<MenuItem[]>(() => {
  return currentDomain.value === 'ops' ? opsDomainMenu : ops2DomainMenu
})

watch(layoutMode, (v) => localStorage.setItem(LAYOUT_KEY, v))
watch(currentDomain, (v) => {
  localStorage.setItem(DOMAIN_KEY, v)
  document.documentElement.style.setProperty('--primary-color', domainConfig[v].color)
}, { immediate: true })

watch(
  () => route.path,
  (newPath) => {
    if (newPath && newPath !== '/login' && newPath !== '/') {
      localStorage.setItem(LAST_PAGE_KEY, newPath)
    }
  }
)

function toggleLayout() {
  layoutMode.value = layoutMode.value === 'side' ? 'top' : 'side'
}

function handleDomainCommand(cmd: string) {
  if (cmd === 'ops') {
    currentDomain.value = 'ops'
    router.push('/sql/workbench')
  } else if (cmd === 'ops2') {
    currentDomain.value = 'ops2'
    router.push('/db-lifecycle')
  }
}

function handleCommand(cmd: string) {
  if (cmd === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', { type: 'warning' }).then(() => {
      if (route.path && route.path !== '/login') {
        localStorage.setItem(LAST_PAGE_KEY, route.path)
      }
      userStore.logout()
      ElMessage.success('已退出登录')
      router.push('/login')
    }).catch(() => {})
  } else if (cmd === 'profile') {
    router.push('/profile')
  }
}

onMounted(() => {
  document.documentElement.style.setProperty('--primary-color', currentDomainColor.value)
  nextTick(() => {
    const lastPage = localStorage.getItem(LAST_PAGE_KEY)
    const currentPath = route.path
    if (currentPath === '/' || currentPath === '' || currentPath === '/login') {
      if (lastPage && lastPage !== '/login' && lastPage !== '/') {
        router.replace(lastPage)
      } else {
        router.replace('/dashboard')
      }
    }
  })
})
</script>

<style scoped>
.layout {
}
.layout-top :deep(.el-header) {
  height: 56px !important;
}
.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #1f2d3d;
  flex-shrink: 0;
}
.fade-enter-active, .fade-leave-active {
  transition: opacity .2s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
.domain-tag {
  margin: 8px 12px;
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid;
  text-align: center;
  font-size: 12px;
  flex-shrink: 0;
}
.layout-footer {
  text-align: right;
  font-size: 12px;
  color: #909399;
  padding: 6px 20px;
  background: #fff;
  border-top: 1px solid #e4e7ed;
  flex-shrink: 0;
}
</style>
