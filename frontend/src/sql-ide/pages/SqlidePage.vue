<template>
  <div class="sqlide-root">
    <!-- 左侧对象树 -->
    <div class="object-tree-panel">
      <div class="ot-search-wrap">
        <svg class="ot-search-icon" viewBox="0 0 24 24" width="14" height="14">
          <path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z" fill="#78909c"/>
        </svg>
        <input v-model="searchKeyword" type="text" placeholder="搜索数据库/表/视图/函数..." class="ot-search-input" />
        <button v-if="searchKeyword" class="ot-search-clear" @click="searchKeyword = ''">×</button>
      </div>

      <div class="ot-toolbar">
        <button class="ot-btn" @click="refreshAll">
          <span class="ot-refresh-icon">⟳</span>
          <span>刷新</span>
        </button>
        <button class="ot-btn" @click="collapseAll">
          <svg viewBox="0 0 16 16" width="12" height="12"><path d="M3 6 h10 M3 10 h10" stroke="#606266" stroke-width="1.2" stroke-linecap="round"/></svg>
          <span>折叠</span>
        </button>
      </div>

      <div class="ot-body" ref="otBodyRef">
        <div v-if="(dsState.list || []).length === 0" class="ot-empty">
          <div class="ot-empty-icon">🗄️</div>
          <div class="ot-empty-text">暂无数据源</div>
        </div>
        <DatasourceNode
          v-for="ds in dsState.list"
          :key="'ds-' + ds.datasourceId"
          :datasource="ds"
          :search-keyword="searchKeyword"
          :is-connected="!!isConnected(ds.datasourceId)"
          :is-loading="!!isLoading(ds.datasourceId)"
          :is-expanded="!!isExpanded(ds.datasourceId)"
          :tree-root="treeFor(ds.datasourceId)"
          @toggle="handleToggleDs(ds.datasourceId)"
          @refresh="handleRefreshDs(ds.datasourceId)"
          @close-conn="handleCloseConn(ds.datasourceId)"
        />
      </div>

      <div class="ot-footer">
        <span>共 {{ (dsState.list || []).length }} 个数据源</span>
      </div>
    </div>

    <!-- 右侧编辑+结果面板 -->
    <div class="editor-result-panel">
      <!-- Tab Bar -->
      <div class="tab-bar">
        <div
          v-for="tab in tabList" :key="'tab-' + tab.id"
          class="tab-item"
          :class="{ active: activeTabId === tab.id, modified: (tab as any)._modified }"
          @click="switchTab(tab.id)"
          @contextmenu.prevent="openTabMenu($event, tab)"
        >
          <span class="tab-title">{{ tab.title || '查询' }}</span>
          <span class="tab-close" @click.stop="closeTab(tab.id)">×</span>
        </div>
        <div class="tab-bar-add" @click="addNewTab" title="新建查询">+ 新建查询</div>
      </div>

      <!-- Tab 内容 -->
        <div class="tab-content">
        <template v-for="tab in tabList" :key="'content-' + tab.id">
          <!-- 查询类型 Tab -->
          <div v-show="activeTabId === tab.id && tab.kind === 'query'" class="tab-pane">
            <QueryToolbar
              :datasource-id="(tab as any).datasourceId || ''"
              :database="(tab as any).database || ''"
              :datasources="dsState.list"
              :databases="databasesFor((tab as any).datasourceId)"
              :executing="!!(tab as any)._executing"
              :effective-db="(tab as any)._effectiveDb"
              @ds-change="(v: string) => onTabDsChange(tab, v)"
              @db-change="(v: string) => onTabDbChange(tab, v)"
              @run="onTabRun(tab)"
              @stop="onTabStop(tab)"
              @explain="onTabExplain(tab)"
              @refresh-db="onTabRefreshDb(tab)"
              @save="onTabSave(tab)"
              @favorite="onTabFavorite(tab)"
              @beautify="onTabBeautify(tab)"
            />
            <SqlEditor
              :ref="(el: any) => { if (el) editorRefs[tab.id] = el }"
              :value="(tab as any).sql || ''"
              :tab-id="tab.id"
              :suggestion-names="namesFor((tab as any).datasourceId, (tab as any).database)"
              :suggestion-columns="columnsFor((tab as any).datasourceId, (tab as any).database)"
              :readonly="!!(tab as any)._executing"
              @input="(v: string) => onEditorInput(tab, v)"
              @run-command="onTabRun(tab)"
              @beautify="onTabBeautify(tab)"
            />
            <ResultTabs
              :results="(tab as any)._results || []"
              :history="(tab as any)._history || []"
              :history-total="(tab as any)._historyTotal || 0"
              :messages="(tab as any)._messages || []"
              :explain="(tab as any)._explain || []"
              :ddl-text="(tab as any)._ddlText || ''"
              :favorites="favorites"
              :sql="tab.sql || ''"
              :datasource-id="(tab as any).datasourceId || ''"
              :database="(tab as any)._effectiveDb || (tab as any).database || ''"
              @close-result="(i: number) => onTabResultClose(tab, i)"
              @close-all="onTabResultClear(tab)"
              @replay="(sql: string) => onTabReplay(tab, sql)"
              @refresh="refreshHistory"
              @refresh-single="(rId: string, idx: number, sql: string) => onTabRefreshSingle(tab, rId, idx, sql)"
              @load-history-page="(page: number, pageSize: number) => loadHistory(page, pageSize)"
              @apply-favorite="(sql: string) => onTabInsertSql(tab, sql)"
              @remove-favorite="(id: string) => removeFavorite(id)"
            />
          </div>
          
          <!-- 表查看类型 Tab - 类似 Navicat 的打开表功能 -->
          <div v-show="activeTabId === tab.id && tab.kind === 'table'" class="tab-pane">
            <TableViewerTab
              :tab="tab"
              :datasources="dsState.list"
              @close="closeTab(tab.id)"
              @refresh="refreshTableTab(tab)"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- 节点右键菜单 -->
    <ContextMenu />

    <!-- Tab 右键菜单 -->
    <div v-if="tabMenu.show" class="tab-ctx-menu" :style="{ left: tabMenu.x + 'px', top: tabMenu.y + 'px' }" @mouseleave="tabMenu.show = false">
      <div class="tab-ctx-item" @click="tabMenuAction('close')">关闭当前标签</div>
      <div class="tab-ctx-item" @click="tabMenuAction('close-others')">关闭其他标签</div>
      <div class="tab-ctx-item" @click="tabMenuAction('close-all')">关闭所有标签</div>
      <div class="tab-ctx-sep"></div>
      <div class="tab-ctx-item" @click="tabMenuAction('duplicate')">复制标签</div>
      <div class="tab-ctx-item" @click="tabMenuAction('rename')">重命名</div>
    </div>

    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, provide, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import DatasourceNode from '../components/DatasourceNode.vue'
import QueryToolbar from '../components/QueryToolbar.vue'
import SqlEditor from '../components/SqlEditor.vue'
import ResultTabs from '../components/ResultTabs.vue'
import ContextMenu from '../components/ContextMenu.vue'
import TableViewerTab from '../components/TableViewerTab.vue'
import { useSqlIdeTabs } from '../hooks/useTabs'
import { useDatasource } from '../hooks/useDatasource'
import * as api from '../service/api'
import { cancelExecute } from '../../api/sql'
import { beautifySql, resolveEffectiveDatabase } from '../utils/sql'
import type { DatasourceSummary } from '../types'

const searchKeyword = ref('')
const tabs = useSqlIdeTabs()
const dsState = useDatasource()

// 从 tabs 对象中解构出所需属性
const { tabs: tabList, activeTabId, findActive } = tabs

// 扩展 tabs 状态 - 为每个 tab 增加执行相关状态
const tabRuntime = reactive<Record<string, any>>({})

// 存储原始行数据用于刷新后恢复状态
const originalRows = reactive<Record<string, any[]>>({})

// 存储编辑器组件引用
const editorRefs = reactive<Record<string, any>>({})

function getEditorComponent(tabId: string) {
  return editorRefs[tabId]
}

function ensureTabRuntime(tabId: string) {
  if (!tabRuntime[tabId]) {
    tabRuntime[tabId] = {
      _executing: false,
      _modified: false,
      _results: [],
      _history: [],
      _messages: [],
      _explain: [],
      _ddlText: ''
    }
  }
  return tabRuntime[tabId]
}

function getTabRuntime(tab: any) {
  if (!tab) return {}
  return ensureTabRuntime(tab.id)
}

// 使 tab 的 reactive 字段在模板中可用 - 通过代理访问
// 简化方案: 扩展原始 tab 对象
const tabBar = reactive({ show: false, x: 0, y: 0, tabId: '' })

// 在模板中通过 getter 方式访问 - 但需要保持响应性
// 使用 computed 方案: 直接扩展 tab 对象

// 由于 useTabs 返回的 tab 对象本身是 reactive，直接注入字段
function ensureTabFields(tab: any) {
  if (tab._executing === undefined) tab._executing = false
  if (tab._modified === undefined) tab._modified = false
  if (tab._results === undefined) tab._results = []
  
  // 如果是新建的tab，优先使用全局历史数据，其次从其他已有的query tab复制
  if (tab._history === undefined || tab._history.length === 0) {
    if (globalHistory.value.length > 0) {
      tab._history = [...globalHistory.value]
    } else {
      const existingTab = tabList.find((t: any) => t.kind === 'query' && t.id !== tab.id && t._history && t._history.length > 0)
      if (existingTab) {
        tab._history = [...existingTab._history]
      } else {
        tab._history = []
      }
    }
  }
  
  if (tab._messages === undefined) tab._messages = []
  if (tab._explain === undefined) tab._explain = []
  if (tab._ddlText === undefined) tab._ddlText = ''
  if (tab._effectiveDb === undefined) tab._effectiveDb = ''
  return tab
}

// === 基础工具 ===
function isConnected(datasourceId: string) {
  return dsState.trees && dsState.trees[datasourceId] && dsState.trees[datasourceId].loaded
}
function isLoading(datasourceId: string) {
  return dsState.trees && dsState.trees[datasourceId] && dsState.trees[datasourceId].loading
}
function isExpanded(datasourceId: string) {
  return dsState.expanded && dsState.expanded[datasourceId]
}
function treeFor(datasourceId: string): any[] {
  return (dsState.trees && dsState.trees[datasourceId]?.nodes) || []
}

function databasesFor(datasourceId: string): string[] {
  if (!datasourceId) return []
  return dsState.trees[datasourceId]?.databases || []
}

function escapeIdentifier(name: string): string {
  if (!name) return ''
  return '`' + String(name).replace(/`/g, '') + '`'
}

function namesFor(datasourceId: string, db: string): string[] {
  if (!datasourceId) return []
  const names: string[] = []
  const tree = treeFor(datasourceId)
  const walk = (nodes: any[]) => {
    if (!Array.isArray(nodes)) return
    for (const n of nodes) {
      if (!n) continue
      if ((n.type === 'table' || n.type === 'view') && (!db || n.database === db)) {
        if (n.name) names.push(n.name)
        if (n.table) names.push(n.table)
      }
      if (Array.isArray(n.children)) walk(n.children)
    }
  }
  walk(tree)
  return Array.from(new Set(names))
}

function columnsFor(datasourceId: string, db: string): string[] {
  if (!datasourceId) return []
  const cols: string[] = []
  const tree = treeFor(datasourceId)
  const walk = (nodes: any[]) => {
    if (!Array.isArray(nodes)) return
    for (const n of nodes) {
      if (!n) continue
      if ((n.type === 'table' || n.type === 'view') && (!db || n.database === db)) {
        if (Array.isArray(n.columns)) {
          for (const c of n.columns) {
            const colName = typeof c === 'string' ? c : (c.name || '')
            if (colName) cols.push(colName)
          }
        }
      }
      if (Array.isArray(n.children)) walk(n.children)
    }
  }
  walk(tree)
  return Array.from(new Set(cols))
}

// === 查询保存管理 ===
interface SavedQuery {
  id: string
  title: string
  sql: string
  database: string
  datasourceId: string
  createdAt: string
}

interface FavoriteItem {
  id: string
  title: string
  description: string
  sql: string
  createdAt: string
}

const savedQueries = ref<SavedQuery[]>([])
const favorites = ref<FavoriteItem[]>([])

async function loadSavedQueries(datasourceId: string) {
  try {
    const list = await api.listSavedQueries({ datasourceId, page: 1, pageSize: 100 })
    savedQueries.value = list.map((item: any) => ({
      id: item.id || item.queryId || item.savedQueryId,
      title: item.title,
      description: item.description || '',
      sql: item.sql,
      database: item.database || '',
      datasourceId: item.datasourceId || '',
      createdAt: item.createdAt || item.created_at || nowString()
    }))
  } catch (e) {
    console.error('加载收藏查询失败:', e)
    savedQueries.value = []
  }
}

async function saveQueryToStorage(title: string, sql: string, database: string, datasourceId: string) {
  const activeTab = tabs.activeTab as any
  const dsId = activeTab?.datasourceId || datasourceId
  
  try {
    await api.saveQuery({
      datasourceId: dsId,
      database: database || activeTab?.database || '',
      title,
      description: '',
      sql
    })
    await loadSavedQueries(dsId)
    ElMessage.success('查询已保存')
  } catch (e: any) {
    ElMessage.error('保存失败：' + (e?.message || String(e)))
  }
}

async function deleteSavedQuery(id: string, database: string, datasourceId: string) {
  try {
    await api.deleteSavedQuery(id)
    savedQueries.value = savedQueries.value.filter(q => q.id !== id)
    ElMessage.success('已删除')
  } catch (e: any) {
    ElMessage.error('删除失败：' + (e?.message || String(e)))
  }
}

async function loadFavorites() {
  try {
    const list = await api.listFavorites({ page: 1, pageSize: 100 })
    favorites.value = list.map((item: any) => ({
      id: item.favoriteId || item.id,
      title: item.title,
      description: item.description || '',
      sql: item.sqlText || item.sql,
      createdAt: item.createdAt || item.created_at || nowString()
    }))
  } catch (e) {
    console.error('加载收藏失败:', e)
    favorites.value = []
  }
}

async function addFavorite(title: string, description: string, sql: string) {
  try {
    await api.createFavorite({ title, description, sql })
    await loadFavorites()
    ElMessage.success('已添加收藏')
  } catch (e: any) {
    ElMessage.error('添加收藏失败：' + (e?.message || String(e)))
  }
}

async function removeFavorite(id: string) {
  try {
    await api.deleteFavorite(id)
    favorites.value = favorites.value.filter(f => f.id !== id)
    ElMessage.success('已取消收藏')
  } catch (e: any) {
    ElMessage.error('删除收藏失败：' + (e?.message || String(e)))
  }
}

// === Tab 操作 ===
function addNewTab() {
  const dsId = dsState.currentId || ((dsState.list || [])[0]?.datasourceId) || ''
  const tab = tabs.addQueryTab(dsId, '', '-- 在此编写 SQL\n\nSELECT 1;\n', '查询 ' + (tabs.tabs.filter((t: any) => t.kind === 'query').length + 1))
  ensureTabFields(tab)
  refreshHistory()
}

function switchTab(id: string) {
  tabs.switchTab(id)
}

function closeTab(id: string) {
  tabs.closeTab(id)
  delete tabRuntime[id]
}

function refreshTableTab(tab: any) {
  if (tab.kind === 'table') {
    tabs.loadTableData(tab).catch(() => undefined)
  }
}

function nowString(): string {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds())
}

function pushMessage(tab: any, level: string, text: string) {
  ensureTabFields(tab)
  tab._messages.unshift({ time: nowString(), level, text })
  if (tab._messages.length > 200) tab._messages.length = 200
}

// === Tab 右键菜单 ===
const tabMenu = reactive({ show: false, x: 0, y: 0, tabId: '' })
function openTabMenu(e: MouseEvent, tab: any) {
  tabMenu.show = true
  tabMenu.x = e.clientX
  tabMenu.y = e.clientY
  tabMenu.tabId = tab.id
  setTimeout(() => {
    const handler = (ev: MouseEvent) => {
      if (!(ev.target as HTMLElement).closest('.tab-ctx-menu')) tabMenu.show = false
      document.removeEventListener('mousedown', handler)
    }
    document.addEventListener('mousedown', handler)
  }, 0)
}

function tabMenuAction(action: string) {
  const tab = tabs.tabs.find((t: any) => t.id === tabMenu.tabId)
  tabMenu.show = false
  if (!tab) return
  if (action === 'close') closeTab(tab.id)
  else if (action === 'close-others') {
    const keepIds = new Set([tab.id])
    const toClose = tabs.tabs.filter((t: any) => !keepIds.has(t.id)).map((t: any) => t.id)
    toClose.forEach((id: string) => closeTab(id))
    tabs.switchTab(tab.id)
  }
  else if (action === 'close-all') {
    const ids = tabs.tabs.map((t: any) => t.id)
    ids.forEach((id: string) => closeTab(id))
  }
  else if (action === 'duplicate') {
    const newTab = tabs.addQueryTab(tab.datasourceId || '', tab.database || '', tab.sql || '', (tab.title || '查询') + ' 副本')
    ensureTabFields(newTab)
  }
  else if (action === 'rename') {
    ElMessageBox.prompt('输入新标题', '重命名标签', { defaultValue: tab.title || '' }).then(({ value }: any) => {
      if (value && value.trim()) tab.title = value.trim()
    }).catch(() => undefined)
  }
}

// === 编辑器输入 ===
function onEditorInput(tab: any, v: string) {
  ensureTabFields(tab)
  tab.sql = v
  tab._modified = true
}

// === 切换数据源/数据库 ===
function onTabDsChange(tab: any, v: string) {
  tab.datasourceId = v
  tab.database = ''
  tab._effectiveDb = ''
  if (v && !isConnected(v)) {
    dsState.loadTree(v).catch((e: Error) => {
      ElMessage.error('加载对象树失败: ' + e.message)
    })
  }
}

function onTabDbChange(tab: any, v: string) {
  tab.database = v
  tab._effectiveDb = v
}

function onTabRefreshDb(tab: any) {
  if (tab.datasourceId) {
    dsState.loadTree(tab.datasourceId, true).then(() => {
      ElMessage.success('刷新成功')
    }).catch((e: Error) => { ElMessage.error('刷新失败: ' + e.message) })
  }
}

function onTabInsertSql(tab: any, sql: string) {
  if (sql) {
    // 如果编辑器已有内容，在末尾追加新SQL
    const existingSql = tab.sql || ''
    // 在最后一条SQL后面添加新SQL，用两个换行分隔
    tab.sql = existingSql + (existingSql ? '\n\n' : '') + sql
    tab._modified = true
    ElMessage.success('SQL 已插入编辑器')
  }
}

function onTabSave(tab: any) {
  ElMessageBox.prompt('保存当前 SQL', '保存', {
    inputValue: tab.title || '',
    inputPlaceholder: '输入保存名称'
  }).then(({ value }: any) => {
    if (!value || !value.trim()) return
    const title = value.trim()
    tab.title = title
    saveQueryToStorage(title, tab.sql || '', tab.database || '', tab.datasourceId || '')
  }).catch(() => undefined)
}

function onTabFavorite(tab: any) {
  // 获取 textarea 元素
  const textarea = document.getElementById('sql-editor-' + tab.id) as HTMLTextAreaElement
  
  if (!textarea) {
    ElMessage.warning('未找到编辑器')
    return
  }
  
  // 获取选中的 SQL
  const selectedSql = textarea.value.substring(textarea.selectionStart, textarea.selectionEnd)
  
  if (!selectedSql || !selectedSql.trim()) {
    ElMessage.warning('请先选中要收藏的 SQL')
    return
  }
  
  // 保存选中状态
  const savedStart = textarea.selectionStart
  const savedEnd = textarea.selectionEnd
  
  ElMessageBox.confirm(
    `<div>
      <div style="margin-bottom: 8px;">请填写收藏信息</div>
      <input type="text" id="fav-title" placeholder="标题" style="width: 100%; padding: 6px; margin-bottom: 8px; border: 1px solid #d0d0d0; border-radius: 3px;" value="${tab.title || ''}">
      <textarea id="fav-desc" placeholder="描述" style="width: 100%; padding: 6px; height: 60px; border: 1px solid #d0d0d0; border-radius: 3px;"></textarea>
    </div>`,
    '添加收藏',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      dangerouslyUseHTMLString: true
    }
  ).then(() => {
    const title = (document.getElementById('fav-title') as HTMLInputElement)?.value || '未命名'
    const desc = (document.getElementById('fav-desc') as HTMLTextAreaElement)?.value || ''
    if (!title.trim()) {
      ElMessage.warning('请输入标题')
      return
    }
    addFavorite(title.trim(), desc, selectedSql)
  }).catch(() => undefined).finally(() => {
    // 外部操作结束，恢复选中状态
    setTimeout(() => {
      textarea.focus()
      textarea.setSelectionRange(savedStart, savedEnd)
    }, 100)
  })
}

function onTabBeautify(tab: any) {
  const formatted = beautifySql(tab.sql || '')
  tab.sql = formatted
  tab._modified = true
}

// === SQL 执行 ===
function onTabRun(tab: any) {
  if (!tab.datasourceId) {
    ElMessage.warning('请先选择数据源')
    return
  }
  
  // 获取 textarea 元素
  const textarea = document.getElementById('sql-editor-' + tab.id) as HTMLTextAreaElement
  
  // 获取编辑器中的选中SQL
  let selectedSql = ''
  if (textarea) {
    selectedSql = textarea.value.substring(textarea.selectionStart, textarea.selectionEnd).trim()
  }
  
  // 如果有选中SQL，执行选中的SQL；否则执行全部SQL
  const sql = selectedSql || (tab.sql || '').trim()
  
  if (!sql) {
    ElMessage.warning('SQL 为空')
    return
  }
  
  ensureTabFields(tab)
  const dbs = databasesFor(tab.datasourceId)
  const effectiveDb = resolveEffectiveDatabase(sql, tab.database, dbs)
  tab._effectiveDb = effectiveDb
  tab._executing = true
  tab._results = []
  tab._explain = []
  pushMessage(tab, 'INFO', '开始执行 SQL...')

  api.executeSql({ datasourceId: tab.datasourceId, database: effectiveDb, sql, ignoreRisk: true }).then((res: any) => {
    if (res && res.executionId) {
      tab._executionId = res.executionId
    }
    const results = Array.isArray(res) ? res : (res && Array.isArray(res.results) ? res.results : (res ? [res] : []))
    
    const sqlStatements = sql.split(';').map(s => s.trim()).filter(s => s.length > 0)
    
    const mapped = results.map((r: any, i: number) => {
      const r2: any = Object.assign({}, r)
      if (!r2.id) r2.id = tab.id + '_r_' + Date.now() + '_' + i
      if (!r2.title && r2.statementText) {
        const s = String(r2.statementText).trim().substring(0, 30)
        r2.title = s
      }
      if (!r2.title) r2.title = '结果 ' + (i + 1)
      
      // 优先使用后端返回的单条SQL语句，其次使用statementText，最后使用对应索引的SQL语句
      if (!r2.sql || r2.sql === '') {
        if (r.SQL && r.SQL !== '') {
          r2.sql = r.SQL
        } else if (r.sql && r.sql !== '') {
          r2.sql = r.sql
        } else if (r2.statementText && r2.statementText !== '') {
          r2.sql = r2.statementText
        } else if (i < sqlStatements.length) {
          r2.sql = sqlStatements[i] + ';'
        } else {
          r2.sql = sqlStatements[Math.min(i, sqlStatements.length - 1)] + ';'
        }
      }
      
      // 确保 success 字段存在
      if (r2.success === undefined) {
        r2.success = !(r2.error || r2.message?.toLowerCase().includes('error'))
      }
      
      // 确保 rows 和 columns 字段正确映射 - 处理多种API返回格式
      if (!Array.isArray(r2.rows)) {
        r2.rows = Array.isArray(r2.data) ? r2.data : 
                   Array.isArray(r.records) ? r.records : 
                   Array.isArray(r.result) ? r.result : []
      }
      if (!Array.isArray(r2.columns)) {
        r2.columns = Array.isArray(r2.columnNames) ? r2.columnNames : 
                     Array.isArray(r.columnNames) ? r.columnNames : 
                     Array.isArray(r.columns) ? r.columns : []
      }
      // 如果列信息为空但有数据行，从第一行推断列名
      if (r2.columns.length === 0 && r2.rows.length > 0 && typeof r2.rows[0] === 'object' && r2.rows[0] !== null) {
        r2.columns = Object.keys(r2.rows[0])
      }
      
      // 确保 affectedRows 字段存在
      if (r2.affectedRows === undefined) {
        r2.affectedRows = r2.rows?.length || 0
      }
      
      return r2
    })
    tab._results = mapped

    // 加入历史
    const totalAffected = mapped.reduce((acc: number, r: any) => acc + (Number(r.affectedRows) || 0), 0)
    const hasError = mapped.some((r: any) => r.success === false)
    tab._history.unshift({
      time: nowString(),
      datasourceId: tab.datasourceId,
      datasourceName: ((dsState.list || []).find((d: any) => d.datasourceId === tab.datasourceId) as any)?.name || '',
      database: effectiveDb,
      sql: sql,
      success: !hasError,
      affectedRows: totalAffected,
      durationMs: 0
    })
    if (tab._history.length > 500) tab._history.length = 500

    for (const r of mapped) {
      pushMessage(tab, r.success === false ? 'ERROR' : 'INFO', (r.success === false ? '失败: ' : 'OK: ') + (r.message || '执行成功'))
    }
  }).catch((e: any) => {
    console.error('SQL execution error:', e)
    tab._results = [{
      id: tab.id + '_r_' + Date.now(),
      title: '错误',
      success: false,
      message: e?.message || String(e) || '执行失败',
      columns: [],
      rows: [],
      affectedRows: 0
    }]
    pushMessage(tab, 'ERROR', e?.message || String(e) || '执行失败')
  }).finally(() => {
    tab._executing = false
    setTimeout(() => {
      refreshHistory()
    }, 300)
  })
}

function onTabStop(tab: any) {
  if (!tab._executionId) {
    ElMessage.warning('未找到执行任务')
    return
  }
  
  cancelExecute(tab._executionId).then((res: any) => {
    if (res && res.success) {
      ElMessage.success(res.message)
    } else {
      ElMessage.warning(res?.message || '停止请求失败')
    }
  }).catch((e: any) => {
    ElMessage.error('停止请求失败: ' + (e?.message || String(e)))
  })
}

function onTabExplain(tab: any) {
  if (!tab.datasourceId) { ElMessage.warning('请先选择数据源'); return }
  
  const textarea = document.querySelector(`[data-tab-id="${tab.id}"] textarea`) as HTMLTextAreaElement
  let sql = ''
  
  if (textarea) {
    const selectedSql = textarea.value.substring(textarea.selectionStart, textarea.selectionEnd).trim()
    if (selectedSql) {
      sql = selectedSql
    } else {
      ElMessage.warning('请先选中要解释的SQL语句');
      return
    }
  } else {
    sql = (tab.sql || '').trim()
    if (!sql) {
      ElMessage.warning('SQL 为空');
      return
    }
  }
  
  ensureTabFields(tab)
  const effectiveDb = resolveEffectiveDatabase(sql, tab.database, databasesFor(tab.datasourceId))
  tab._executing = true
  tab._explain = []
  
  api.explainSql({ datasourceId: tab.datasourceId, database: effectiveDb, sql }).then((res: any) => {
    const results = Array.isArray(res) ? res : (res && Array.isArray(res.results) ? res.results : (res ? [res] : []))
    
    const explainResults = results.map((r: any, i: number) => {
      const r2: any = Object.assign({}, r)
      if (!r2.id) r2.id = tab.id + '_exp_' + Date.now() + '_' + i
      if (!r2.title) r2.title = '执行计划 ' + (i + 1)
      
      r2.rows = Array.isArray(r2.rows) ? r2.rows : (Array.isArray(r2.data) ? r2.data : [])
      r2.success = r.success !== false
      r2.durationMs = r.durationMs || 0
      r2.sql = r.sql || ''
      
      return r2
    })
    
    tab._explain = explainResults
  }).catch((e: any) => {
    ElMessage.error('解释失败: ' + (e?.message || String(e)))
    tab._explain = []
  }).finally(() => { tab._executing = false })
}

function onTabResultClose(tab: any, i: number) {
  ensureTabFields(tab)
  if (tab._results && tab._results[i] !== undefined) tab._results.splice(i, 1)
}

function onTabResultClear(tab: any) {
  ensureTabFields(tab)
  tab._results = []
  tab._explain = []
  tab._ddlText = ''
}

function onTabReplay(tab: any, sql: string) {
  if (!tab) return
  tab.sql = sql
  tab._modified = true
  onTabRun(tab)
}

function onTabRefreshSingle(tab: any, rId: string, resultIdx: number, sql: string) {
  if (!tab || !tab.datasourceId) return
  
  ensureTabFields(tab)
  const dbs = databasesFor(tab.datasourceId)
  const effectiveDb = resolveEffectiveDatabase(sql, tab.database, dbs)
  
  tab._executing = true
  
  api.executeSql({ datasourceId: tab.datasourceId, database: effectiveDb, sql, ignoreRisk: true }).then((res: any) => {
    const results = Array.isArray(res) ? res : (res && Array.isArray(res.results) ? res.results : (res ? [res] : []))
    
    if (results.length > 0) {
      const r2: any = Object.assign({}, results[0])
      // 保持原来的结果ID，确保originalRows的key一致
      const existingResult = tab._results[resultIdx]
      if (existingResult && existingResult.id) {
        r2.id = existingResult.id
      } else if (!r2.id) {
        r2.id = tab.id + '_r_' + Date.now()
      }
      if (!r2.title && r2.statementText) {
        const s = String(r2.statementText).trim().substring(0, 30)
        r2.title = s
      }
      if (!r2.title) r2.title = '结果 ' + (resultIdx + 1)
      r2.sql = sql
      
      if (r2.success === undefined) {
        r2.success = !(r2.error || r2.message?.toLowerCase().includes('error'))
      }
      
      if (!Array.isArray(r2.rows)) {
        r2.rows = Array.isArray(r2.data) ? r2.data : 
                   Array.isArray(r2.records) ? r2.records : 
                   Array.isArray(r2.result) ? r2.result : []
      }
      if (!Array.isArray(r2.columns)) {
        r2.columns = Array.isArray(r2.columnNames) ? r2.columnNames : 
                     Array.isArray(r2.columns) ? r2.columns : []
      }
      if (r2.columns.length === 0 && r2.rows.length > 0 && typeof r2.rows[0] === 'object' && r2.rows[0] !== null) {
        r2.columns = Object.keys(r2.rows[0])
      }
      
      if (r2.affectedRows === undefined) {
        r2.affectedRows = r2.rows?.length || 0
      }
      
      tab._results[resultIdx] = r2
      
      originalRows[rId] = r2.rows.map((row: any) => ({ ...row }))
    }
    
    pushMessage(tab, 'INFO', '结果已刷新')
  }).catch((e: any) => {
    console.error('刷新失败:', e)
    ElMessage.error('刷新失败: ' + (e?.message || String(e)))
  }).finally(() => {
    tab._executing = false
  })
}

// === DDL 查看 ===
function viewDDL(datasourceId: string, database: string, table: string) {
  // 确保有一个活动 tab
  let active: any = tabs.findActive()
  if (!active) {
    active = tabs.addQueryTab(datasourceId, database, '', 'DDL ' + table)
    ensureTabFields(active)
  }
  ensureTabFields(active)
  tabs.switchTab(active.id)
  api.getTableDDL({ datasourceId, database, table }).then((ddl: string) => {
    if (typeof ddl === 'string') {
      active._ddlText = ddl
    } else if (ddl && (ddl as any).ddl) {
      active._ddlText = (ddl as any).ddl
    } else {
      active._ddlText = JSON.stringify(ddl, null, 2)
    }
  }).catch((e: any) => {
    ElMessage.error('获取 DDL 失败: ' + (e?.message || String(e)))
  })
}

// === 数据源/节点联动 ===
function handleToggleDs(datasourceId: string) {
  if (!isConnected(datasourceId)) {
    dsState.loadTree(datasourceId).then(() => {
      dsState.currentId = datasourceId
    }).catch((e: Error) => {
      ElMessage.error('加载对象树失败: ' + e.message)
    })
  } else {
    if (dsState.expanded) dsState.expanded[datasourceId] = !dsState.expanded[datasourceId]
  }
}

function handleRefreshDs(datasourceId: string) {
  dsState.loadTree(datasourceId, true).then(() => {
    ElMessage.success('刷新成功')
  }).catch((e: Error) => { ElMessage.error('刷新失败: ' + e.message) })
}

function handleCloseConn(datasourceId: string) {
  if (dsState.trees && dsState.trees[datasourceId]) {
    dsState.trees[datasourceId].loaded = false
  }
  if (dsState.expanded) dsState.expanded[datasourceId] = false
  ElMessage.success('已关闭连接')
}

function refreshAll() {
  const promises: Promise<any>[] = []
  for (const ds of (dsState.list || [])) {
    if (isConnected((ds as any).datasourceId)) {
      promises.push(dsState.loadTree((ds as any).datasourceId, true))
    }
  }
  Promise.allSettled(promises).then(() => {
    ElMessage.success('刷新完成')
  })
}

function collapseAll() {
  for (const ds of (dsState.list || [])) {
    const id = (ds as any).datasourceId
    if (dsState.expanded && dsState.expanded[id]) {
      dsState.expanded[id] = false
    }
  }
}



// === 节点回调（提供给 DatasourceNode 和子节点使用）===
provide('sqlTreeCallbacks', {
  onNodeDblClick: (node: any) => {
    const dsId = node.datasourceId || dsState.currentId
    if (dsId) dsState.currentId = dsId
    
    if (node.type === 'query-save') {
      return
    }
    
    if (node.type === 'saved-query') {
      const dsId = node.datasourceId || dsState.currentId
      const db = node.database || ''
      const sql = node.sql || ''
      const title = node.name || '查询'
      
      const existingTab = tabList.find((t: any) => 
        t.datasourceId === dsId && 
        t.sql === sql && 
        t.title === title
      )
      
      if (existingTab) {
        tabs.switchTab(existingTab.id)
      } else {
        const newTab = tabs.addQueryTab(dsId, db, sql, title)
        if (newTab) {
          ensureTabFields(newTab)
          newTab._effectiveDb = db
          tabs.switchTab(newTab.id)
        }
      }
      refreshHistory()
      return
    }
    
    let active: any = tabs.findActive()
    if (!active) {
      active = tabs.addQueryTab(dsId, node.database || '', '', '查询')
      ensureTabFields(active)
    }
    ensureTabFields(active)
    if (node.type === 'table' || node.type === 'view') {
      active.datasourceId = dsId
      active.database = node.database || ''
      active._effectiveDb = node.database || ''
      active.sql = 'SELECT * FROM ' + (node.database ? escapeIdentifier(node.database) + '.' : '') + escapeIdentifier(node.table || node.name) + ' LIMIT 200;'
      active._modified = true
      tabs.switchTab(active.id)
      onTabRun(active)
    } else if (node.type === 'function' || node.type === 'procedure') {
      active.datasourceId = dsId
      active.database = node.database || ''
      active.sql = 'SELECT ' + escapeIdentifier(node.name) + '();'
      active._modified = true
    } else if (node.type === 'trigger' || node.type === 'event') {
      viewDDL(dsId, node.database || '', node.name || '')
    } else {
      tabs.switchTab(active.id)
    }
  },
  onNodeCtxMenu: (e: MouseEvent, node: any) => {
    if (e && e.preventDefault) e.preventDefault()
    if (e && e.stopPropagation) e.stopPropagation()
    const dsId = node.datasourceId || dsState.currentId
    if (dsId) dsState.currentId = dsId
    const ctxMenu = (window as any).__sqlCtxMenu
    if (ctxMenu) {
      ctxMenu.openMenu(e, node)
    }
  },
  onNodeViewDDL: (node: any) => {
    const dsId = node.datasourceId || dsState.currentId
    viewDDL(dsId, node.database || '', node.table || node.name || '')
  }
})

// 挂载 ContextMenu 到 window，供节点右键菜单调用
import { useSqlIdeContextMenu } from '../hooks/useContextMenu'
const cm = useSqlIdeContextMenu()
;(window as any).__sqlCtxMenu = {
  openMenu: (e: MouseEvent, node: any) => {
    cm.ctxMenu.node = node
    cm.ctxMenu.datasource = null
    cm.ctxMenu.x = e.clientX
    cm.ctxMenu.y = e.clientY
    cm.ctxMenu.show = true
  },
  openDatasourceMenu: (e: MouseEvent, info: any) => {
    cm.ctxMenu.node = null
    cm.ctxMenu.datasource = info
    cm.ctxMenu.x = e.clientX
    cm.ctxMenu.y = e.clientY
    cm.ctxMenu.show = true
  }
}



// 全局历史数据（所有tab共享）
const globalHistory = ref<any[]>([])

const historyPageSize = ref(50)
const historyCurrentPage = ref(1)
const historyTotal = ref(0)

async function loadHistory(page = 1, pageSize = 50) {
  try {
    const [countRes, listRes] = await Promise.all([
      api.countHistory(),
      api.listHistory({ page, pageSize })
    ])
    
    historyTotal.value = countRes.count || (listRes.total || 0)
    historyCurrentPage.value = page
    historyPageSize.value = pageSize
    
    const listData = listRes.list || listRes.data || listRes.rows || (Array.isArray(listRes) ? listRes : [])
    
    const newHistory = listData.map((item: any) => ({
      id: item.historyId || item.id || item.sqlHistoryId,
      sql: item.sqlText || item.sql || '',
      database: item.databaseName || item.database || '',
      datasourceId: item.datasourceId || '',
      datasourceName: item.datasourceName || '',
      createdAt: item.createdAt || item.created_at || nowString(),
      durationMs: item.durationMs || 0,
      affectedRows: item.rowsAffected || item.affectedRows || 0,
      status: item.status || '',
      success: item.status === 'success' || item.success === true
    }))
    
    globalHistory.value = newHistory
    
    tabList.forEach((tab: any) => {
      if (tab.kind === 'query') {
        tab._history = [...globalHistory.value]
        tab._historyTotal = historyTotal.value
        tab._historyCurrentPage = historyCurrentPage.value
        tab._historyPageSize = historyPageSize.value
      }
    })
  } catch (e) {
    console.error('加载历史记录失败:', e)
    globalHistory.value = []
    historyTotal.value = 0
  }
}

function refreshHistory() {
  loadHistory(1, historyPageSize.value)
}

function changeHistoryPage(page: number) {
  loadHistory(page, historyPageSize.value)
}

watch(() => dsState.currentId, () => {
  refreshHistory()
})

// === 生命周期 ===
onMounted(async () => {
  try {
    await dsState.loadAll()
    
    // 加载收藏查询、收藏和历史记录（按当前用户）
    await Promise.all([
      loadSavedQueries(''),
      loadFavorites(),
      loadHistory()
    ])
    
    if (tabList.length === 0 && (dsState.list || []).length > 0) {
      // 优先选择标记为默认的数据源
      const defaultDs = ((dsState.list || []).find((ds: any) => ds.isDefault || ds.default === true)) as any
      const targetDs = defaultDs || ((dsState.list || [])[0] as any)
      const datasourceId = targetDs.datasourceId || targetDs.id || ''
      
      let defaultDb = ''
      // 如果有默认数据源且已连接，尝试获取其第一个数据库
      if (datasourceId && dsState.trees[datasourceId]?.loaded && dsState.trees[datasourceId].databases.length > 0) {
        defaultDb = dsState.trees[datasourceId].databases[0]
      } else if (datasourceId) {
        // 尝试加载数据源的对象树以获取数据库列表
        try {
          await dsState.loadTree(datasourceId)
          if (dsState.trees[datasourceId]?.databases.length > 0) {
            defaultDb = dsState.trees[datasourceId].databases[0]
          }
        } catch (e) {
          console.warn('Failed to load default datasource tree:', e)
        }
      }
      
      const tab = tabs.addQueryTab(datasourceId, defaultDb, '-- 在此编写 SQL\n\nSELECT 1;\n', '查询 1')
      ensureTabFields(tab)
      tab._effectiveDb = defaultDb
    }
  } catch (e: any) {
    console.error('加载数据源失败:', e)
    ElMessage.error('加载数据源失败: ' + (e?.message || String(e)))
  }
})
</script>

<style scoped>
.sqlide-root {
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
  background: #ffffff;
}

.object-tree-panel {
  flex: 0 0 280px;
  width: 280px;
  height: 100%;
  min-height: 0;
  border-right: 1px solid #d0d0d0;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ot-search-wrap {
  position: relative;
  padding: 8px 10px;
  border-bottom: 1px solid #e0e0e0;
  background: #fafafa;
}

.ot-search-icon {
  position: absolute;
  left: 20px;
  top: 50%;
  transform: translateY(-50%);
}

.ot-search-input {
  width: 100%;
  height: 28px;
  padding: 0 30px 0 32px;
  border: 1px solid #d0d0d0;
  border-radius: 3px;
  font-size: 12.5px;
  outline: none;
  background: #ffffff;
  box-sizing: border-box;
}

.ot-search-input:focus { border-color: #1976d2; }

.ot-search-clear {
  position: absolute;
  right: 18px;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  border: none;
  background: transparent;
  color: #909399;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.ot-search-clear:hover { background: #eceff1; color: #37474f; }

.ot-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-bottom: 1px solid #e8e8e8;
  background: #ffffff;
}

.ot-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border: 1px solid transparent;
  background: transparent;
  color: #606266;
  font-size: 12px;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.1s;
}

.ot-btn:hover {
  background: #e3f2fd;
  color: #1976d2;
  border-color: #bbdefb;
}

.ot-body {
  flex: 1 1 auto;
  overflow: auto;
  min-height: 0;
  padding: 4px 0;
}

.ot-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px 10px;
  color: #909399;
  gap: 8px;
}

.ot-empty-icon { font-size: 28px; opacity: 0.5; }
.ot-empty-text { font-size: 13px; }

.ot-footer {
  border-top: 1px solid #e0e0e0;
  padding: 6px 12px;
  font-size: 11px;
  color: #78909c;
  background: #fafafa;
  text-align: right;
  flex-shrink: 0;
}

.editor-result-panel {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #ffffff;
}

.tab-bar {
  display: flex;
  align-items: center;
  flex: 0 0 auto;
  border-bottom: 1px solid #d0d0d0;
  background: #f0f0f0;
  padding: 0 4px;
  height: 32px;
  overflow-x: auto;
  overflow-y: hidden;
}

.tab-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  margin: 2px 2px 0 2px;
  border: 1px solid transparent;
  background: transparent;
  cursor: pointer;
  font-size: 12.5px;
  color: #606266;
  border-radius: 3px 3px 0 0;
  user-select: none;
  white-space: nowrap;
  height: 28px;
}

.tab-item:hover { background: #e3e3e3; }

.tab-item.active {
  background: #ffffff;
  color: #1976d2;
  border: 1px solid #d0d0d0;
  border-bottom-color: #ffffff;
  font-weight: 500;
}

.tab-item.modified .tab-title::after {
  content: ' •';
  color: #ef6c00;
  font-weight: 700;
}

.tab-close {
  font-size: 14px;
  color: #909399;
  cursor: pointer;
  padding: 0 2px;
  line-height: 1;
  border-radius: 2px;
}

.tab-close:hover { color: #c62828; background: #ffebee; }

.tab-bar-add {
  padding: 4px 12px;
  cursor: pointer;
  font-size: 12.5px;
  color: #1976d2;
  user-select: none;
  margin-left: 4px;
  border-radius: 3px;
}

.tab-bar-add:hover { background: #e3f2fd; }

.tab-content {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.tab-pane {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.tab-ctx-menu {
  position: fixed;
  z-index: 9999;
  background: #ffffff;
  border: 1px solid #d0d0d0;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  padding: 4px 0;
  min-width: 150px;
}

.tab-ctx-item {
  padding: 6px 14px;
  cursor: pointer;
  font-size: 12.5px;
  color: #303133;
  transition: background 0.1s;
}

.tab-ctx-item:hover { background: #1976d2; color: #ffffff; }

.tab-ctx-sep {
  height: 1px;
  background: #e0e0e0;
  margin: 4px 0;
}

</style>
