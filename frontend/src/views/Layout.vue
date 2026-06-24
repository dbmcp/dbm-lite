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

          <div style="height:1px;background:#1f2d3d;margin:6px 0;"></div>

          <template v-for="item in currentDomainMenu" :key="'d-' + item.index">
            <el-menu-item v-if="!item.visible || item.visible()" :index="item.index">
              <el-icon :color="currentDomainColor"><component :is="item.icon" /></el-icon>
              <template #title>
                <span :style="{ color: currentDomainColor }">{{ item.title }}</span>
              </template>
            </el-menu-item>
          </template>

          <div style="height:1px;background:#1f2d3d;margin:6px 0;"></div>

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
        <div style="display:flex;align-items:center;gap:16px;flex:1;min-width:0;">
          <el-button v-if="layoutMode === 'side'" link @click="collapsed = !collapsed">
            <el-icon :size="20"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
          </el-button>

          <!-- 顶部布局：自定义菜单条（支持自适应折叠到下拉菜单） -->
          <div v-if="layoutMode === 'top'" ref="topBarRef" class="top-menu-bar">
            <!-- 测量用隐藏节点：与实际菜单项结构完全一致，用于获取真实宽度 -->
            <div class="measure-layer" aria-hidden="true">
              <span
                v-for="(item, idx) in composedTopMenu"
                :key="'m-' + item.index"
                :ref="(el: any) => setMeasureNode(idx, el)"
                class="top-menu-item"
              >
                <el-icon class="tmi-icon"><component :is="item.icon" /></el-icon>
                <span class="tmi-text">{{ item.title }}</span>
              </span>
            </div>

            <!-- 可见的菜单项 -->
            <template v-for="(item, idx) in visibleTopMenu" :key="'vis-' + item.index">
              <el-tooltip :content="item.title" placement="bottom" :disabled="!item.isTruncated">
                <a
                  class="top-menu-item"
                  :class="{ 'is-active': isActive(item.index) }"
                  @click="goTo(item.index)"
                >
                  <el-icon class="tmi-icon" :color="item.iconColor || '#303133'">
                    <component :is="item.icon" />
                  </el-icon>
                  <span class="tmi-text" :title="item.title">{{ item.displayTitle }}</span>
                </a>
              </el-tooltip>
            </template>

            <!-- 折叠下拉菜单 -->
            <el-dropdown
              v-if="collapsedTopMenu.length > 0"
              trigger="click"
              @command="(c: string) => goTo(c)"
              placement="bottom-end"
            >
              <a class="top-menu-item more-item">
                <el-icon :size="20" class="more-dots"><MoreFilled /></el-icon>
              </a>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item
                    v-for="item in collapsedTopMenu"
                    :key="'col-' + item.index"
                    :command="item.index"
                    :disabled="isActive(item.index)"
                  >
                    <span>{{ item.title }}</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <div style="display:flex;align-items:center;gap:12px;flex-shrink:0;">
          <el-tooltip content="切换布局" placement="bottom-end">
            <el-button link @click="toggleLayout">
              <el-icon :size="18"><Menu v-if="layoutMode === 'side'" /><Grid v-else /></el-icon>
            </el-button>
          </el-tooltip>

          <el-dropdown trigger="click" @command="handleDomainCommand" placement="bottom-end" hide-on-click>
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

          <el-dropdown trigger="click" @command="handleCommand" placement="bottom-end">
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

      <el-main style="padding:0;background:#ffffff;overflow:auto;">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  DataLine, Fold, Expand, ArrowDown,
  EditPen, Coin, Folder, Monitor, Document, Setting, Menu, Grid, Lock, DocumentChecked,
  Upload, Histogram, RefreshRight, Management, View, MoreFilled
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

const DEFAULT_LAYOUT: 'side' | 'top' = 'top'

function pickInitialLayout(): 'side' | 'top' {
  const saved = localStorage.getItem(LAYOUT_KEY) as 'side' | 'top'
  if (saved === 'side' || saved === 'top') return saved
  return DEFAULT_LAYOUT
}

const layoutMode = ref<'side' | 'top'>(pickInitialLayout())

const opsColor = '#409EFF'
const ops2Color = '#2E6BA8'

const domainDefaultPage: Record<DomainType, string> = {
  ops: '/sql/sqlide',
  ops2: '/db-lifecycle'
}

function pickInitialDomain(): DomainType {
  const saved = localStorage.getItem(DOMAIN_KEY) as DomainType
  if (saved === 'ops' || saved === 'ops2') return saved
  const last = localStorage.getItem(LAST_PAGE_KEY)
  if (last === '/sql/sqlide' || last === '/datasources' || last === '/db-users') return 'ops'
  if (last === '/db-lifecycle' || last === '/servers' || last === '/plugins') return 'ops2'
  return 'ops'
}

const currentDomain = ref<DomainType>(pickInitialDomain())

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
  { index: '/sql/sqlide', title: 'SQL IDE', icon: EditPen },
  { index: '/datasources', title: '数据源', icon: Document },
  { index: '/priv/groups', title: '对象权限', icon: Lock },
  { index: '/import-export', title: '导入导出', icon: Upload },
  { index: '/sql-audit', title: 'SQL审核', icon: DocumentChecked },
  { index: '/sensitive-data', title: '敏感数据', icon: View }
]

const ops2DomainMenu: MenuItem[] = [
  { index: '/db-lifecycle', title: 'DB生命周期', icon: Coin },
  { index: '/servers', title: '服务器', icon: Monitor, visible: () => userStore.isAdmin },
  { index: '/health-check', title: '健康巡检', icon: Histogram },
  { index: '/data-migration', title: '数据迁移', icon: RefreshRight }
]

// 顺序：业务管理 -> 审批流程 -> 操作审计 -> 平台配置 -> 个人中心
// 个人中心从顶部菜单移除，仅通过右上角头像下拉访问
const bottomMenu: MenuItem[] = [
  { index: '/businesses', title: '业务管理', icon: Folder },
  { index: '/workflow/todo', title: '审批流程', icon: DocumentChecked },
  { index: '/audit', title: '操作审计', icon: Document, visible: () => userStore.isAdmin },
  { index: '/platform-config', title: '平台配置', icon: Management, visible: () => userStore.isAdmin },
  { index: '/profile', title: '个人中心', icon: Setting }
]

const currentDomainMenu = computed<MenuItem[]>(() => {
  return currentDomain.value === 'ops' ? opsDomainMenu : ops2DomainMenu
})

function pathToDomain(path: string): DomainType | null {
  if (!path) return null
  if (path.startsWith('/sql') || path === '/datasources' || path.startsWith('/datasource') ||
      path.startsWith('/priv') || path === '/db-users' || path === '/import-export' ||
      path === '/sql-audit' || path === '/sensitive-data') return 'ops'
  if (path === '/db-lifecycle' || path === '/servers' || path === '/health-check' ||
      path === '/data-migration') return 'ops2'
  return null
}

// ============ 顶部菜单自适应折叠逻辑 ============

interface TopMenuItem {
  index: string
  title: string
  icon: any
  iconColor?: string
  displayTitle: string
  isTruncated: boolean
}

const composedTopMenu = computed<TopMenuItem[]>(() => {
  const items: TopMenuItem[] = []
  topMenu.forEach((m) => items.push({
    index: m.index, title: m.title, icon: m.icon,
    displayTitle: m.title, isTruncated: false
  }))
  currentDomainMenu.value
    .filter((m) => !m.visible || m.visible())
    .forEach((m) => items.push({
      index: m.index, title: m.title, icon: m.icon,
      iconColor: currentDomainColor.value,
      displayTitle: m.title, isTruncated: false
    }))
  bottomMenu
    .filter((m) => !m.visible || m.visible())
    .forEach((m) => items.push({
      index: m.index, title: m.title, icon: m.icon,
      displayTitle: m.title, isTruncated: false
    }))
  return items
})

const topBarRef = ref<HTMLElement | null>(null)
const visibleCount = ref<number>(999)
const measuredWidths = ref<number[]>([])

// 单个菜单最小展示宽度：图标约 20px + 至少4个中文字(约64px) + 内边距，至少 110px 能稳定完整展示
const MORE_ITEM_WIDTH = 56 // "..." 项大致占位
const ITEM_SAFE_MIN = 110   // 单个菜单最小展示宽度（完整显示图标+4字标题）

const measureNodeArr: (HTMLElement | null)[] = []

function setMeasureNode(idx: number, el: any) {
  measureNodeArr[idx] = el as HTMLElement | null
}

function measureItemWidths() {
  const widths: number[] = []
  for (let i = 0; i < composedTopMenu.value.length; i++) {
    const el = measureNodeArr[i]
    if (el) {
      const w = el.getBoundingClientRect().width
      widths.push(Math.max(ITEM_SAFE_MIN, Math.ceil(w) + 4))
    } else {
      widths.push(ITEM_SAFE_MIN)
    }
  }
  measuredWidths.value = widths
  return widths
}

function computeLayout() {
  nextTick(() => {
    const bar = topBarRef.value
    if (!bar) return
    const widths = measureItemWidths()
    const maxWidth = bar.getBoundingClientRect().width

    if (widths.length === 0) return

    const total = widths.reduce((s, w) => s + w, 0)
    if (total <= maxWidth) {
      visibleCount.value = widths.length
      return
    }

    // 需要折叠：从前往后累加，剩余空间不足的整项进入下拉
    const budget = maxWidth - MORE_ITEM_WIDTH
    let used = 0
    let count = 0
    for (let i = 0; i < widths.length; i++) {
      if (used + widths[i] <= budget) {
        used += widths[i]
        count++
      } else {
        break
      }
    }
    if (count < 1) count = 1
    visibleCount.value = count
  })
}

let ro: ResizeObserver | null = null

function bindObserver() {
  if (layoutMode.value !== 'top') return
  nextTick(() => {
    if (ro) { ro.disconnect(); ro = null }
    computeLayout()
    if (topBarRef.value && typeof (window as any).ResizeObserver !== 'undefined') {
      ro = new ResizeObserver(() => computeLayout())
      ro.observe(topBarRef.value)
    }
  })
}

watch(() => composedTopMenu.value.length, () => bindObserver())

watch(layoutMode, (v) => {
  localStorage.setItem(LAYOUT_KEY, v)
  if (v === 'top') {
    bindObserver()
  } else {
    if (ro) { ro.disconnect(); ro = null }
  }
})

watch(currentDomain, (v) => {
  localStorage.setItem(DOMAIN_KEY, v)
  document.documentElement.style.setProperty('--primary-color', domainConfig[v].color)
  nextTick(() => bindObserver())
}, { immediate: true })

watch(
  () => route.path,
  (newPath) => {
    if (!newPath || newPath === '/login' || newPath === '/') return
    localStorage.setItem(LAST_PAGE_KEY, newPath)
    const d = pathToDomain(newPath)
    if (d && d !== currentDomain.value) {
      currentDomain.value = d
    }
  }
)

const visibleTopMenu = computed<TopMenuItem[]>(() => {
  const items = composedTopMenu.value
  const cap = Math.min(visibleCount.value, items.length)
  return items.slice(0, cap).filter((m) => m.index !== '/profile')
})

const collapsedTopMenu = computed<TopMenuItem[]>(() => {
  const items = composedTopMenu.value
  const cap = Math.min(visibleCount.value, items.length)
  const tail = items.slice(cap)
  const hasProfile = tail.some((m) => m.index === '/profile')
  if (!hasProfile) {
    const profile = items.find((m) => m.index === '/profile')
    if (profile) tail.push(profile)
  }
  return tail
})

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(path + '/')
}

function goTo(path: string) {
  if (route.path === path) return
  router.push(path)
}

function toggleLayout() {
  layoutMode.value = layoutMode.value === 'side' ? 'top' : 'side'
}

function handleDomainCommand(cmd: string) {
  if (cmd === 'ops') {
    currentDomain.value = 'ops'
    router.push(domainDefaultPage.ops)
  } else if (cmd === 'ops2') {
    currentDomain.value = 'ops2'
    router.push(domainDefaultPage.ops2)
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
    if (layoutMode.value === 'top') bindObserver()

    const lastPage = localStorage.getItem(LAST_PAGE_KEY)
    const currentPath = route.path
    if (currentPath === '/' || currentPath === '' || currentPath === '/login') {
      if (lastPage && lastPage !== '/login' && lastPage !== '/') {
        const d = pathToDomain(lastPage)
        if (d) currentDomain.value = d
        router.replace(lastPage)
      } else {
        router.replace('/dashboard')
      }
    }
  })
})

onBeforeUnmount(() => {
  if (ro) { ro.disconnect(); ro = null }
})
</script>

<style scoped>
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

/* ============ 顶部自定义菜单样式 ============ */
.top-menu-bar {
  position: relative;
  display: flex;
  align-items: center;
  min-width: 0;
  flex: 1;
  gap: 4px;
  overflow: hidden;
}
.measure-layer {
  position: absolute;
  left: -9999px;
  top: 0;
  visibility: hidden;
  pointer-events: none;
  display: flex;
  align-items: center;
  gap: 4px;
}
.top-menu-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 40px;
  padding: 0 14px;
  color: #303133;
  font-size: 14px;
  border-radius: 4px;
  text-decoration: none;
  cursor: pointer;
  white-space: nowrap;
  transition: background-color .2s, color .2s;
  flex-shrink: 0;
}
.top-menu-item:hover {
  background: #f0f2f5;
  color: var(--primary-color, #409EFF);
}
.top-menu-item.is-active {
  color: var(--primary-color, #409EFF);
  background: rgba(64, 158, 255, 0.1);
  font-weight: 500;
}
.tmi-icon {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
}
.tmi-text {
  display: inline-block;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.more-item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-left: 4px;
  padding: 0 10px;
  color: #1f2937;
  line-height: 40px;
  border-radius: 4px;
}
.more-item:hover {
  color: var(--primary-color, #409EFF);
  background: #f0f2f5;
}
.more-item .more-dots {
  display: inline-flex;
  align-items: center;
  font-weight: 700;
}
</style>
