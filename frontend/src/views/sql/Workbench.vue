<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DBA老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
  <div class="workbench-root" :class="{ 'dark-theme': darkTheme }">

    <!-- 顶部工具栏 -->
    <div class="top-toolbar">
      <div class="toolbar-left">
        <span class="workbench-title">
          <el-icon><Edit /></el-icon>
          SQL IDE
        </span>

        <el-select v-model="activeDsId" filterable placeholder="选择数据源" class="ds-select"
          @change="onDatasourceChange">
          <el-option v-for="d in datasources" :key="d.id" :label="d.name" :value="d.id">
            <span class="ds-label">{{ d.name }}</span>
            <span class="ds-type-tag">{{ d.dbType }}</span>
          </el-option>
        </el-select>

        <el-select v-model="activeDatabase" filterable placeholder="选择数据库"
          :disabled="!activeDsId" class="db-select" @change="onDatabaseChange">
          <el-option v-for="db in databases" :key="db" :label="db" :value="db" />
        </el-select>
      </div>

      <div class="toolbar-center">
        <el-button type="primary" :icon="VideoPlay" :loading="executing"
          :disabled="!activeDsId || !sqlContent" @click="doExecuteCurrent">
          执行 (Ctrl+Enter)
        </el-button>
        <el-button :icon="VideoPause" @click="doExecuteAll">执行全部</el-button>
        <el-button :icon="Cpu" :disabled="!activeDsId || !sqlContent || executing"
          @click="doExplain">Explain</el-button>
      </div>

      <div class="toolbar-right">
        <el-button :icon="MagicStick" @click="formatSql">格式化</el-button>
        <el-button :icon="Collection" @click="showSaveDialog = true">收藏</el-button>
        <el-button :icon="Clock" @click="historyVisible = true">历史</el-button>
        <el-button :icon="Setting" @click="settingsVisible = true">设置</el-button>
        <el-tooltip :content="darkTheme ? '浅色主题' : '深色主题'">
          <el-button :icon="darkTheme ? Sunny : Moon" circle @click="toggleTheme" />
        </el-tooltip>
      </div>
    </div>

    <!-- 主体：左右两栏 -->
    <div class="body-row">

      <!-- 左侧：对象树 -->
      <div class="left-pane" :style="{ width: leftWidth + 'px' }">
        <div class="pane-header">
          <span>对象</span>
          <div class="pane-actions">
            <el-input v-model="treeKeyword" placeholder="搜索" size="small" style="width: 140px" clearable />
            <el-button link :icon="RefreshRight" @click="refreshTree">刷新</el-button>
          </div>
        </div>

        <div v-if="treeLoading" class="tree-loading">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
          <span>加载中...</span>
        </div>

        <div v-else-if="connFailed" class="conn-error-panel">
          <el-icon :size="28" color="#f56c6c"><CircleClose /></el-icon>
          <div class="conn-error-title">连接失败</div>
          <div class="conn-error-msg">{{ connFailedMsg || '无法连接到数据源' }}</div>
          <el-button type="primary" :icon="RefreshRight" size="small" @click="refreshTree">重新连接</el-button>
        </div>

        <el-tree v-else ref="treeRef" class="object-tree" :data="treeData" node-key="key" default-expand-all
          :filter-node-method="filterNode" :expand-on-click-node="false" @node-contextmenu="onTreeContextMenu"
          @node-click="onTreeNodeClick" @node-double-click="onTreeNodeDblClick">
          <template #default="{ node, data }">
            <span class="tree-node">
              <el-icon class="tn-icon">
                <component :is="treeIconOf(data)" />
              </el-icon>
              <span class="tn-label" :title="data.label || data.name">{{ data.label || data.name }}</span>
              <span v-if="data.type === 'table' && data.rows != null" class="tn-hint">({{ data.rows }})</span>
            </span>
          </template>
        </el-tree>
      </div>

      <!-- 拖拽分割线 -->
      <div class="splitter-v" @mousedown="startDragV"></div>

      <!-- 右侧工作区 -->
      <div class="right-pane" :style="{ width: `calc(100% - ${leftWidth}px - 6px)` }">

        <!-- SQL Tab 栏 -->
        <div class="sql-tabs-bar">
          <el-tabs v-model="activeTabKey" type="card" closable @tab-remove="removeTab"
            @tab-click="onTabClick">
            <el-tab-pane v-for="t in sqlTabs" :key="t.key" :label="t.label" :name="t.key">
              <span @contextmenu.prevent="onTabContextMenu($event, t)">
                {{ t.label }}
                <span v-if="t.modified" class="mod-dot">●</span>
              </span>
            </el-tab-pane>
            <template #append>
              <el-button :icon="Plus" circle size="small" @click="addTab" />
            </template>
          </el-tabs>
        </div>

        <!-- 当前 Tab 内容：上编辑器 下结果区 -->
        <div class="tab-inner" v-if="activeTab">

          <!-- 编辑器区 -->
          <div class="editor-pane" :style="{ height: editorHeight + 'px' }">
            <div class="editor-toolbar">
              <span class="sql-type-tag">SQL</span>
              <span class="ds-meta">{{ currentDsName }} @ {{ activeDatabase || 'default' }}</span>
              <div class="editor-toolbar-right">
                <span class="hint-text">{{ executing ? '执行中...' : ('耗时: ' + (lastExecMs >= 0 ? lastExecMs + ' ms' : '-')) }}</span>
              </div>
            </div>
            <textarea v-model="activeTab.sqlContent" class="sql-textarea"
              @keydown.ctrl.enter.prevent="doExecuteCurrent"
              @keydown.ctrl.shift.enter.prevent="doExecuteAll"
              @keydown.ctrl.s.prevent="doTabSave"
              @keydown.meta.enter.prevent="doExecuteCurrent"
              @keydown.ctrl.alt.s.prevent="toggleTheme"
              @change="onSqlChange"
              spellcheck="false"></textarea>
          </div>

          <!-- 水平分割线 -->
          <div class="splitter-h" @mousedown="startDragH"></div>

          <!-- 结果区 -->
          <div class="result-pane" :style="{ height: `calc(100% - ${editorHeight}px - 6px)` }">
            <el-tabs v-model="activeTab.activeResultKey" type="border-card" size="small" closable @tab-remove="removeResultTab">
              <el-tab-pane v-for="r in activeTab.results" :key="r.key" :name="r.key">
                <template #label>
                  <span class="result-tab-label">
                    <el-icon :style="{ color: resultIconColor(r) }">
                      <component :is="resultIcon(r)" />
                    </el-icon>
                    {{ r.label }}
                  </span>
                </template>

                <div v-if="r.kind === 'error'" class="result-error">
                  <el-icon color="#f56c6c"><CircleClose /></el-icon>
                  <span>{{ r.message || '执行失败' }}</span>
                </div>

                <div v-else-if="r.kind === 'message'" class="result-message">
                  <el-icon color="#67c23a"><SuccessFilled /></el-icon>
                  <span>{{ r.message }}</span>
                  <span class="hint-text" style="margin-left:12px">影响行数：{{ r.affectedRows }}</span>
                </div>

                <el-table v-else-if="r.kind === 'table' || r.kind === 'explain'" :data="paginatedRows(r)"
                  border stripe size="small" height="100%" highlight-current-row @sort-change="onTableSortChange(r, $event)">
                  <el-table-column v-for="col in r.columns" :key="col" :prop="col" :label="col" sortable
                    show-overflow-tooltip min-width="140">
                    <template #default="{ row }">
                      <span v-if="row[col] === null || row[col] === undefined" class="null-cell">NULL</span>
                      <span v-else>{{ row[col] }}</span>
                    </template>
                  </el-table-column>
                </el-table>

                <div v-if="r.kind === 'table' && r.rows && r.rows.length" class="pager-row">
                  <span class="hint-text">共 {{ r.rows.length }} 行，执行耗时 {{ r.durationMs }} ms</span>
                  <el-pagination v-model:current-page="r.page" small layout="prev, pager, next, jumper"
                    :page-size="r.pageSize" :total="r.rows.length" />
                  <el-select v-model="r.pageSize" size="small" style="width:110px" @change="r.page = 1">
                    <el-option label="50 条/页" :value="50" />
                    <el-option label="100 条/页" :value="100" />
                    <el-option label="300 条/页" :value="300" />
                    <el-option label="500 条/页" :value="500" />
                    <el-option label="1000 条/页" :value="1000" />
                  </el-select>
                </div>
              </el-tab-pane>

              <el-tab-pane v-if="activeTab.results.length === 0" name="empty" label="执行结果">
                <div class="empty-result">
                  <el-icon :size="40" color="#909399"><InfoFilled /></el-icon>
                  <p>暂无执行结果。按 Ctrl+Enter 执行当前 SQL。</p>
                </div>
              </el-tab-pane>
            </el-tabs>
          </div>
        </div>

        <div v-else class="empty-tab">
          <el-icon :size="48" color="#909399"><Files /></el-icon>
          <p>点击右上角「+」新建一个 SQL 窗口</p>
        </div>
      </div>
    </div>

    <!-- 树右键菜单 -->
    <ul v-show="ctxMenu.visible" class="ctx-menu" :style="{ left: ctxMenu.x + 'px', top: ctxMenu.y + 'px' }">
      <template v-if="ctxMenu.node && ctxMenu.node.type === 'table'">
        <li @click="ctxMenuSelectTop100">查看前 100 行</li>
        <li @click="ctxMenuShowDdl">查看 DDL</li>
        <li @click="ctxMenuTableInfo">表信息</li>
      </template>
      <template v-else-if="ctxMenu.node && ctxMenu.node.type === 'view'">
        <li @click="ctxMenuSelectTop100">查看视图数据</li>
        <li @click="ctxMenuShowDdl">查看定义</li>
      </template>
      <template v-else-if="ctxMenu.node && ctxMenu.node.type === 'database'">
        <li @click="ctxMenuRefresh">刷新</li>
      </template>
      <li v-else @click="ctxMenuRefresh">刷新</li>
    </ul>

    <!-- 危险操作确认框 -->
    <el-dialog v-model="riskDialogVisible" title="高危操作确认" width="480px">
      <el-alert type="error" :closable="false" show-icon title="以下操作可能修改或删除数据，请谨慎确认。" />
      <pre class="risk-sql">{{ pendingRiskSql }}</pre>
      <div class="risk-confirm">
        <el-checkbox v-model="riskConfirmed">我已了解风险，确认执行</el-checkbox>
      </div>
      <template #footer>
        <el-button @click="riskDialogVisible = false">取消</el-button>
        <el-button type="danger" :disabled="!riskConfirmed" @click="confirmRiskAndExec">确认执行</el-button>
      </template>
    </el-dialog>

    <!-- 历史记录抽屉 -->
    <el-drawer v-model="historyVisible" title="SQL 执行历史" size="480px" direction="rtl">
      <div class="history-search">
        <el-input v-model="historyKeyword" placeholder="搜索 SQL" clearable />
      </div>
      <el-timeline v-if="historyList.length">
        <el-timeline-item v-for="h in historyList" :key="h.historyId" :timestamp="h.createdAt" placement="top">
          <el-card shadow="hover" size="small">
            <pre class="history-sql">{{ h.sqlText }}</pre>
            <div class="history-meta">
              <el-tag size="small" :type="h.status === 'success' ? 'success' : 'danger'">{{ h.status }}</el-tag>
              <span class="hint-text">{{ h.executeMs }} ms</span>
              <el-button size="small" link type="primary" @click="insertHistoryToEditor(h)">插入编辑器</el-button>
            </div>
          </el-card>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else description="暂无执行历史" />
    </el-drawer>

    <!-- 收藏保存对话框 -->
    <el-dialog v-model="showSaveDialog" title="收藏当前 SQL" width="480px">
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="saveForm.name" placeholder="请输入收藏名称" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="saveForm.tags" placeholder="多个标签用逗号分隔" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSaveDialog = false">取消</el-button>
        <el-button type="primary" @click="saveCurrentAsFavorite">保存</el-button>
      </template>
    </el-dialog>

    <!-- 设置抽屉 -->
    <el-drawer v-model="settingsVisible" title="SQL IDE 设置" size="420px" direction="rtl">
      <el-form label-width="160px">
        <el-form-item label="深色主题">
          <el-switch v-model="darkTheme" @change="persistSettings" />
        </el-form-item>
        <el-form-item label="Tab 内容持久化">
          <el-switch v-model="persistTab" @change="persistSettings" />
        </el-form-item>
        <el-form-item label="危险操作二次确认">
          <el-switch v-model="checkRisk" @change="persistSettings" />
        </el-form-item>
        <el-form-item label="默认每页行数">
          <el-select v-model="defaultPageSize" @change="persistSettings">
            <el-option label="100" :value="100" />
            <el-option label="300" :value="300" />
            <el-option label="500" :value="500" />
            <el-option label="1000" :value="1000" />
          </el-select>
        </el-form-item>
        <el-form-item label="编辑器默认高度占比">
          <el-slider v-model="editorHeightRatio" :min="30" :max="80" @change="persistEditorHeight" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="persistSettings">保存</el-button>
          <el-button @click="resetSettings">恢复默认</el-button>
        </el-form-item>
      </el-form>
    </el-drawer>

    <!-- 表信息抽屉 -->
    <el-drawer v-model="tableInfoVisible" :title="'表信息: ' + tableInfo?.table" size="520px" direction="rtl">
      <div v-if="tableInfoLoading" class="tree-loading">
        <el-icon class="is-loading" :size="20"><Loading /></el-icon>
        <span>加载表信息...</span>
      </div>
      <template v-else-if="tableInfo">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="数据库">{{ tableInfo.database }}</el-descriptions-item>
          <el-descriptions-item label="表名">{{ tableInfo.table }}</el-descriptions-item>
          <el-descriptions-item label="行数">{{ tableInfo.rows }}</el-descriptions-item>
          <el-descriptions-item label="大小 MB">{{ tableInfo.sizeMb ?? '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="tableInfo.engine" label="引擎">{{ tableInfo.engine }}</el-descriptions-item>
          <el-descriptions-item v-if="tableInfo.comment" label="注释">{{ tableInfo.comment }}</el-descriptions-item>
        </el-descriptions>

        <h4 class="sub-title">列信息</h4>
        <el-table :data="(tableInfo.columns || []) as any[]" border size="small">
          <el-table-column prop="name" label="列名" />
          <el-table-column prop="type" label="类型" />
          <el-table-column prop="key" label="索引" />
          <el-table-column prop="nullable" label="可空" />
          <el-table-column prop="default" label="默认值" />
          <el-table-column prop="comment" label="注释" />
        </el-table>

        <h4 class="sub-title">索引</h4>
        <el-table :data="(tableInfo.indexes || []) as any[]" border size="small" empty-text="暂无索引">
          <el-table-column prop="name" label="索引名" />
          <el-table-column prop="unique" label="是否唯一" />
          <el-table-column prop="seq" label="顺序" />
          <el-table-column prop="origin" label="源列" />
        </el-table>

        <h4 class="sub-title">DDL</h4>
        <pre class="ddl-box">{{ tableInfo.ddl || '-' }}</pre>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick, reactive } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Edit, VideoPlay, VideoPause, Cpu, MagicStick, Collection, Clock, Setting, Moon, Sunny,
  RefreshRight, Plus, Files, Loading, CircleClose, SuccessFilled, InfoFilled,
  DataAnalysis, Folder, DataBoard, Menu, Document
} from '@element-plus/icons-vue'
import type { Datasource } from '@/api/datasource'
import { listAllDatasources, listDatabases, getFullTree, testConnectionById } from '@/api/datasource'
import { executeSql, explainSql, getTableInfo, listSqlHistory } from '@/api/sql'

// =================== 数据结构 ===================
interface SqlTab {
  key: string
  label: string
  datasourceId: string
  database: string
  sqlContent: string
  modified: boolean
  activeResultKey: string
  results: SqlResult[]
}

interface SqlResult {
  key: string
  kind: 'table' | 'explain' | 'message' | 'error'
  label: string
  columns: string[]
  rows: any[]
  message: string
  affectedRows: number
  durationMs: number
  page: number
  pageSize: number
  sql: string
}

// =================== 状态 ===================
const route = useRoute()

// 全局配置（localStorage）
const darkTheme = ref(false)
const persistTab = ref(true)
const checkRisk = ref(true)
const defaultPageSize = ref(50)
const editorHeightRatio = ref(55) // 编辑器占 55%
const favorites = ref<any[]>([])

// 数据源相关
const datasources = ref<Datasource[]>([])
const activeDsId = ref<string>('')
const databases = ref<string[]>([])
const activeDatabase = ref<string>('')

// 对象树
const treeRef = ref<any>(null)
const treeData = ref<any[]>([])
const treeKeyword = ref('')
const treeLoading = ref(false)
const connFailed = ref(false)
const connFailedMsg = ref('')

// SQL Tab 与编辑器
const sqlTabs = ref<SqlTab[]>([])
const activeTabKey = ref<string>('')
const activeTab = computed<SqlTab | null>(() => sqlTabs.value.find(t => t.key === activeTabKey.value) || null)
const sqlContent = computed({
  get: () => activeTab.value?.sqlContent || '',
  set: (v) => { if (activeTab.value) activeTab.value.sqlContent = v }
})
const editorHeight = ref(320)

// 执行状态
const executing = ref(false)
const lastExecMs = ref(-1)

// 危险操作
const riskDialogVisible = ref(false)
const pendingRiskSql = ref('')
const riskConfirmed = ref(false)
const pendingExecAll = ref(false)

// 历史 / 设置 / 收藏 / 表信息
const historyVisible = ref(false)
const historyList = ref<any[]>([])
const historyKeyword = ref('')
const settingsVisible = ref(false)
const showSaveDialog = ref(false)
const saveForm = reactive({ name: '', tags: '' })
const tableInfoVisible = ref(false)
const tableInfoLoading = ref(false)
const tableInfo = ref<any>(null)

// 右键菜单
const ctxMenu = reactive({ visible: false, x: 0, y: 0, node: null as any })

// 布局
const leftWidth = ref(280)

// =================== 工具函数 ===================
const currentDsName = computed(() => {
  const d = datasources.value.find(x => (x as any).id === activeDsId.value)
  return d ? d.name : '-'
})

function uid(): string {
  return 't_' + Math.random().toString(36).slice(2, 10)
}

// 检测是否为高危操作（DELETE/UPDATE 无 WHERE、DROP、TRUNCATE、ALTER）
function isRiskySql(sql: string): { risky: boolean; reasons: string[] } {
  const reasons: string[] = []
  const upper = sql.toUpperCase()
  if (/DELETE\s+FROM/i.test(upper) && !/WHERE/.test(upper)) reasons.push('DELETE 无 WHERE 子句')
  if (/UPDATE\s+\w+/i.test(upper) && !/WHERE/.test(upper)) reasons.push('UPDATE 无 WHERE 子句')
  if (/DROP\s+(TABLE|DATABASE|INDEX|VIEW)/i.test(upper)) reasons.push('包含 DROP 操作')
  if (/TRUNCATE/i.test(upper)) reasons.push('包含 TRUNCATE 操作')
  if (/ALTER\s+TABLE/i.test(upper)) reasons.push('包含 ALTER TABLE 操作')
  return { risky: reasons.length > 0, reasons }
}

function splitStatements(sql: string): string[] {
  const result: string[] = []
  let current = ''
  let inSingle = false, inDouble = false, inBacktick = false, inLine = false, inBlock = false
  for (let j = 0; j < sql.length; j++) {
    const ch = sql[j]
    const next = sql[j + 1]
    current += ch
    if (inLine) {
      if (ch === '\n') inLine = false
      continue
    }
    if (inBlock) {
      if (ch === '*' && next === '/') { inBlock = false; current += next; j++ }
      continue
    }
    if (inSingle) { if (ch === "'" && sql[j - 1] !== '\\') inSingle = false; continue }
    if (inDouble) { if (ch === '"' && sql[j - 1] !== '\\') inDouble = false; continue }
    if (inBacktick) { if (ch === '`') inBacktick = false; continue }
    if (ch === '-' && next === '-') { inLine = true; continue }
    if (ch === '/' && next === '*') { inBlock = true; continue }
    if (ch === '#') { inLine = true; continue }
    if (ch === "'") { inSingle = true; continue }
    if (ch === '"') { inDouble = true; continue }
    if (ch === '`') { inBacktick = true; continue }
    if (ch === ';') {
      // 完成一个语句
      const trimmed = current.trim()
      if (trimmed.length > 0 && trimmed !== ';') {
        result.push(trimmed.substring(0, trimmed.length - 1).trim() || trimmed)
      }
      current = ''
    }
  }
  // 最后一个语句（如果没有以分号结尾）
  const last = current.trim()
  if (last.length > 0 && last !== ';') result.push(last)
  return result.length > 0 ? result : []
}

// =================== 初始化 ===================
onMounted(async () => {
  loadSettings()
  applyEditorHeight()

  await loadDatasources()
  const initDsId = (route.query.datasourceId as string) || (datasources.value[0] as any)?.id || ''
  if (initDsId) {
    activeDsId.value = initDsId
    await onDatasourceChange(initDsId, true)
  }

  // 根据持久化恢复 Tab，否则新建
  const saved = loadTabsFromStorage()
  if (saved && saved.length > 0) {
    sqlTabs.value = saved
    activeTabKey.value = sqlTabs.value[0].key
  } else {
    addTab()
  }

  document.addEventListener('click', hideCtxMenu)
  document.addEventListener('keydown', globalKey)
})

onUnmounted(() => {
  document.removeEventListener('click', hideCtxMenu)
  document.removeEventListener('keydown', globalKey)
})

// =================== 设置 / 持久化 ===================
function loadSettings() {
  try {
    const s = JSON.parse(localStorage.getItem('sql_ide_settings') || '{}')
    if (s.darkTheme !== undefined) darkTheme.value = !!s.darkTheme
    if (s.persistTab !== undefined) persistTab.value = !!s.persistTab
    if (s.checkRisk !== undefined) checkRisk.value = !!s.checkRisk
    if (s.defaultPageSize) defaultPageSize.value = s.defaultPageSize
    if (s.editorHeightRatio) editorHeightRatio.value = s.editorHeightRatio
    if (s.leftWidth) leftWidth.value = s.leftWidth
    if (s.favorites) favorites.value = s.favorites
    if (darkTheme.value) document.body.classList.add('dark-theme')
  } catch (e) { /* ignore */ }
}
function persistSettings() {
  localStorage.setItem('sql_ide_settings', JSON.stringify({
    darkTheme: darkTheme.value,
    persistTab: persistTab.value,
    checkRisk: checkRisk.value,
    defaultPageSize: defaultPageSize.value,
    editorHeightRatio: editorHeightRatio.value,
    leftWidth: leftWidth.value,
    favorites: favorites.value,
  }))
  if (darkTheme.value) document.body.classList.add('dark-theme')
  else document.body.classList.remove('dark-theme')
}
function persistEditorHeight() {
  persistSettings()
  applyEditorHeight()
}
function resetSettings() {
  darkTheme.value = false
  persistTab.value = true
  checkRisk.value = true
  defaultPageSize.value = 300
  editorHeightRatio.value = 55
  leftWidth.value = 280
  applyEditorHeight()
  persistSettings()
  ElMessage.success('已恢复默认设置')
}

function applyEditorHeight() {
  // 主体默认 ~ 620；动态更新
  nextTick(() => {
    const rightPane: any = document.querySelector('.right-pane')
    if (!rightPane) return
    const total = rightPane.clientHeight - 52 // 减去 sql-tabs-bar
    editorHeight.value = Math.max(160, Math.floor(total * editorHeightRatio.value / 100))
  })
}

function toggleTheme() {
  darkTheme.value = !darkTheme.value
  persistSettings()
}

// Tab 持久化
function loadTabsFromStorage(): SqlTab[] | null {
  if (!persistTab.value) return null
  try {
    const raw = localStorage.getItem('sql_ide_tabs_' + (activeDsId.value || 'default'))
    if (!raw) return null
    const list = JSON.parse(raw) as SqlTab[]
    return list.length ? list : null
  } catch (e) { return null }
}
function saveTabsToStorage() {
  if (!persistTab.value) return
  try {
    // 持久化只记录 SQL、label、datasourceId、database
    const slim = sqlTabs.value.map(t => ({
      key: t.key, label: t.label, datasourceId: t.datasourceId,
      database: t.database, sqlContent: t.sqlContent, modified: t.modified,
      activeResultKey: '', results: []
    }))
    localStorage.setItem('sql_ide_tabs_' + (activeDsId.value || 'default'), JSON.stringify(slim))
  } catch (e) { /* ignore */ }
}

watch([sqlTabs, activeTabKey], () => saveTabsToStorage(), { deep: true })
watch(() => route.query.datasourceId, (v: any) => {
  if (v && v !== activeDsId.value) {
    activeDsId.value = v
    onDatasourceChange(v, true)
  }
})

// =================== 数据源 / 数据库 / 对象树 ===================
async function loadDatasources() {
  try {
    const res: any = await listAllDatasources()
    const list = Array.isArray(res) ? res : (res?.data || res?.list || [])
    datasources.value = list
  } catch (e) {
    ElMessage.error('加载数据源失败')
  }
}

async function onDatasourceChange(dsId: string, skipIfSame?: boolean) {
  if (!dsId) return
  activeDatabase.value = ''
  databases.value = []
  treeData.value = []
  treeLoading.value = true
  connFailed.value = false
  try {
    const connRes: any = await testConnectionById(dsId)
    if (!connRes) throw new Error('连接失败')
    // 加载数据库
    const dbs: any = await listDatabases(dsId)
    const list = Array.isArray(dbs) ? dbs : (dbs?.data || dbs?.list || [])
    databases.value = list
    activeDatabase.value = list[0] || ''
    await loadTree(dsId)
  } catch (e: any) {
    connFailed.value = true
    connFailedMsg.value = e?.message || '连接失败，请检查数据源配置'
    treeData.value = []
  } finally {
    treeLoading.value = false
  }
}

async function onDatabaseChange() {
  // 切换数据库不需要重载整个树，但更新 Tab
  if (activeTab.value) {
    activeTab.value.database = activeDatabase.value || activeTab.value.database
  }
}

async function loadTree(dsId: string) {
  if (!dsId) return
  treeLoading.value = true
  try {
    const raw: any = await getFullTree(dsId)
    const list = Array.isArray(raw) ? raw : (raw?.data || raw?.list || [])
    treeData.value = normalizeTree(list)
  } catch (e: any) {
    connFailed.value = true
    connFailedMsg.value = e?.message || '加载对象树失败'
  } finally {
    treeLoading.value = false
  }
}

function refreshTree() {
  if (activeDsId.value) loadTree(activeDsId.value)
}

// 把后端返回的嵌套树结构转换为统一格式（带 key/children/label/name/type）
function normalizeTree(list: any[]): any[] {
  if (!list) return []
  const out: any[] = []
  list.forEach((n, i) => {
    const children: any[] = []
    // 可能有 group 节点（table/view/index/trigger）
    if (Array.isArray(n.children) && n.children.length > 0 && n.children[0].group) {
      n.children.forEach((g: any) => {
        const groupChildren = (g.children || []).map((c: any) => ({
          ...c,
          key: (n.name || n.database || i) + '.' + (g.group || g.name) + '.' + (c.name || c.table),
          label: c.name || c.table,
          type: c.type || g.group,
          children: []
        }))
        children.push({
          key: (n.name || n.database || i) + '.' + (g.group || g.name),
          label: groupLabel(g.group, groupChildren.length),
          type: 'group',
          group: g.group,
          children: groupChildren
        })
      })
    } else if (Array.isArray(n.children)) {
      // 普通嵌套
      n.children.forEach((c: any) => {
        children.push({
          ...c,
          key: (n.name || n.database || i) + '.' + (c.name || c.table),
          label: c.name || c.table,
          type: c.type || 'table',
          children: []
        })
      })
    }

    out.push({
      ...n,
      key: n.database || n.name || ('db_' + i),
      label: n.database || n.name || '数据库',
      type: 'database',
      children
    })
  })
  return out
}

function groupLabel(group: string, count: number): string {
  switch (group) {
    case 'table': return `数据表 (${count})`
    case 'view': return `视图 (${count})`
    case 'index': return `索引 (${count})`
    case 'trigger': return `触发器 (${count})`
    default: return `${group} (${count})`
  }
}

function treeIconOf(data: any): any {
  if (!data) return Document
  switch (data.type) {
    case 'database': return DataBoard
    case 'table': return Menu
    case 'view': return DataAnalysis
    case 'index': return Folder
    case 'trigger': return Folder
    case 'group': return Folder
    default: return Document
  }
}

function filterNode(value: string, data: any): boolean {
  if (!value) return true
  return (data.label || data.name || '').toLowerCase().includes(value.toLowerCase())
}

watch(treeKeyword, (v) => {
  if (treeRef.value) treeRef.value!.filter(v)
})

// 节点点击 / 双击
function onTreeNodeClick(data: any) {
  if (data.type === 'table' || data.type === 'view') {
    activeDatabase.value = data.database || activeDatabase.value
  }
}
function onTreeNodeDblClick(data: any) {
  if (data.type === 'table' || data.type === 'view') {
    const sql = `SELECT * FROM ${quoteIfNeeded(data.label || data.name || data.table)} LIMIT 100;`
    appendSqlToActiveTab(sql)
    doExecuteCurrent()
  }
}
function quoteIfNeeded(name: string): string {
  if (/^[A-Za-z0-9_]+$/.test(name)) return name
  return `"${name}"`
}

// 右键菜单
function onTreeContextMenu(e: MouseEvent, data: any) {
  ctxMenu.visible = true
  ctxMenu.x = e.clientX
  ctxMenu.y = e.clientY
  ctxMenu.node = data
  e.stopPropagation()
}
function hideCtxMenu() { ctxMenu.visible = false }

function ctxMenuSelectTop100() {
  hideCtxMenu()
  const node = ctxMenu.node
  if (!node) return
  if (node.type === 'table' || node.type === 'view') {
    const sql = `SELECT * FROM ${quoteIfNeeded(node.label || node.name)} LIMIT 100;`
    appendSqlToActiveTab(sql)
    doExecuteCurrent()
  }
}
async function ctxMenuShowDdl() {
  hideCtxMenu()
  const node = ctxMenu.node
  if (!node || !activeDsId.value) return
  try {
    const res: any = await getTableInfo(activeDsId.value, node.database || activeDatabase.value, node.label || node.name)
    const ddl = res?.ddl || res?.data?.ddl || ''
    if (!ddl) { ElMessage.warning('未获取到 DDL'); return }
    appendSqlToActiveTab(ddl)
  } catch (e: any) { ElMessage.error('获取 DDL 失败: ' + (e?.message || '')) }
}
async function ctxMenuTableInfo() {
  hideCtxMenu()
  const node = ctxMenu.node
  if (!node || !activeDsId.value) return
  tableInfoVisible.value = true
  tableInfoLoading.value = true
  tableInfo.value = null
  try {
    const res: any = await getTableInfo(activeDsId.value, node.database || activeDatabase.value, node.label || node.name)
    tableInfo.value = res?.data || res || null
  } catch (e: any) {
    ElMessage.error('获取表信息失败')
  } finally {
    tableInfoLoading.value = false
  }
}
function ctxMenuRefresh() { hideCtxMenu(); refreshTree() }

// =================== Tab 管理 ===================
function addTab() {
  const key = uid()
  const tab: SqlTab = {
    key,
    label: 'SQL ' + (sqlTabs.value.length + 1),
    datasourceId: activeDsId.value,
    database: activeDatabase.value,
    sqlContent: defaultPlaceholderSql(),
    modified: false,
    activeResultKey: '',
    results: []
  }
  sqlTabs.value.push(tab)
  activeTabKey.value = key
}

function removeTab(key: string) {
  const idx = sqlTabs.value.findIndex(t => t.key === key)
  if (idx < 0) return
  const t = sqlTabs.value[idx]
  if (t.modified && !confirm('当前 Tab 存在未保存的 SQL，确认关闭？')) return
  sqlTabs.value.splice(idx, 1)
  if (activeTabKey.value === key) {
    activeTabKey.value = sqlTabs.value[0]?.key || ''
  }
}

function removeResultTab(key: string | number) {
  if (!activeTab.value) return
  const idx = activeTab.value.results.findIndex(r => r.key === key)
  if (idx < 0) return
  activeTab.value.results.splice(idx, 1)
  if (activeTab.value.activeResultKey === key) {
    activeTab.value.activeResultKey = activeTab.value.results[0]?.key || 'empty'
  }
}

function onTabClick() { /* el-tabs 已经更新 activeTabKey */ }

function onTabContextMenu(e: MouseEvent, t: SqlTab) {
  // 简单处理：重命名
  const name = prompt('输入新名称', t.label)
  if (name && name.trim()) {
    t.label = name.trim()
  }
}

function defaultPlaceholderSql(): string {
  return '-- 在此编写 SQL，按 Ctrl+Enter 执行\n' +
    '-- 多条 SQL 以分号 ; 分隔\n\n'
}

function appendSqlToActiveTab(sql: string) {
  if (!activeTab.value) addTab()
  const t = activeTab.value!
  t.sqlContent = (t.sqlContent ? t.sqlContent + '\n\n' : '') + sql
  t.modified = true
}

function onSqlChange() {
  if (activeTab.value) activeTab.value.modified = true
}

function doTabSave() {
  if (activeTab.value) activeTab.value.modified = false
  ElMessage.success('已保存标记')
}

// =================== 格式化 / 历史 / 收藏 ===================
function formatSql() {
  if (!activeTab.value || !activeTab.value.sqlContent) return
  const sql = activeTab.value.sqlContent.trim()
  // 在主要关键字前换行（忽略字符串和注释内的内容）
  const primaryKeys = [
    'SELECT', 'FROM', 'WHERE', 'GROUP BY', 'ORDER BY', 'HAVING', 'LIMIT',
    'INSERT INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE FROM',
    'CREATE TABLE', 'DROP TABLE', 'ALTER TABLE',
    'LEFT JOIN', 'RIGHT JOIN', 'INNER JOIN', 'JOIN', 'ON', 'UNION', 'UNION ALL', 'WITH',
  ]
  // 从右到左替换较长的关键字（例如 'GROUP BY' 在 'BY' 之前）
  primaryKeys.sort((a, b) => b.length - a.length)

  // 简单的状态机拆分器（忽略字符串/注释内部）
  let result = ''
  let i = 0
  let inSingle = false
  let inDouble = false
  let inBacktick = false
  let inLineComment = false
  let inBlockComment = false
  const upper = sql.toUpperCase()

  while (i < sql.length) {
    if (inLineComment) {
      if (sql[i] === '\n') inLineComment = false
      result += sql[i]
      i++
      continue
    }
    if (inBlockComment) {
      if (sql[i] === '*' && sql[i + 1] === '/') {
        inBlockComment = false
        result += '*/'
        i += 2
        continue
      }
      result += sql[i]
      i++
      continue
    }
    // 注释
    if (!inSingle && !inDouble && !inBacktick) {
      if (sql[i] === '-' && sql[i + 1] === '-') { inLineComment = true; result += sql[i]; i++; continue }
      if (sql[i] === '/' && sql[i + 1] === '*') { inBlockComment = true; result += sql[i]; i++; continue }
      if (sql[i] === '#') { inLineComment = true; result += sql[i]; i++; continue }
    }
    // 字符串
    if (sql[i] === "'" && !inDouble && !inBacktick && !inLineComment && !inBlockComment) { inSingle = !inSingle; result += sql[i]; i++; continue }
    if (sql[i] === '"' && !inSingle && !inBacktick && !inLineComment && !inBlockComment) { inDouble = !inDouble; result += sql[i]; i++; continue }
    if (sql[i] === '`' && !inSingle && !inDouble && !inLineComment && !inBlockComment) { inBacktick = !inBacktick; result += sql[i]; i++; continue }

    // 关键字匹配（仅在非字符串/注释时）
    let matched = false
    if (!inSingle && !inDouble && !inBacktick) {
      for (const kw of primaryKeys) {
        // 前面必须是空格/括号/行首/逗号
        if (i + kw.length > sql.length) continue
        const slice = upper.substring(i, i + kw.length)
        if (slice !== kw) continue
        const before = i === 0 ? ' ' : sql[i - 1]
        if (!/[\s\(\)\,\;]/.test(before)) continue
        // 避免在开头或换行后再添加不必要的换行
        const trimmed = result.replace(/\s+$/, '')
        if (trimmed.length === 0) {
          result = kw + ' '
        } else {
          result = trimmed + '\n' + kw + ' '
        }
        i += kw.length
        // 跳过其后紧接的空格
        while (i < sql.length && sql[i] === ' ') i++
        matched = true
        break
      }
    }
    if (!matched) {
      result += sql[i]
      i++
    }
  }
  activeTab.value.sqlContent = result
  activeTab.value.modified = true
}

async function loadHistory() {
  if (!activeDsId.value) return
  try {
    const res: any = await listSqlHistory(1, 50, activeDsId.value, historyKeyword.value)
    historyList.value = (res?.data?.list || res?.list || []).slice(0, 50)
  } catch (e) { /* ignore */ }
}
watch(historyVisible, (v) => { if (v) loadHistory() })

function insertHistoryToEditor(h: any) {
  appendSqlToActiveTab(h.sqlText || h.sql || '')
  historyVisible.value = false
}

function saveCurrentAsFavorite() {
  const name = saveForm.name.trim()
  if (!name) { ElMessage.warning('请输入名称'); return }
  favorites.value.push({
    id: uid(),
    name,
    tags: saveForm.tags,
    sql: activeTab.value?.sqlContent || '',
    datasourceId: activeDsId.value,
    createdAt: new Date().toISOString()
  })
  saveForm.name = ''
  saveForm.tags = ''
  showSaveDialog.value = false
  persistSettings()
  ElMessage.success('已添加到收藏')
}

// =================== 执行流程 ===================
function doExecuteCurrent() {
  if (!activeDsId.value) { ElMessage.warning('请先选择数据源'); return }
  const sql = getStatementAtCursor() || (activeTab.value?.sqlContent || '')
  executeWithRiskCheck(sql, false)
}
function doExecuteAll() {
  if (!activeDsId.value) { ElMessage.warning('请先选择数据源'); return }
  const sql = activeTab.value?.sqlContent || ''
  executeWithRiskCheck(sql, true)
}
async function doExplain() {
  if (!activeDsId.value || !activeTab.value) return
  const sql = getStatementAtCursor() || activeTab.value.sqlContent
  if (!sql.trim()) return
  executing.value = true
  try {
    const res: any = await explainSql({ datasourceId: activeDsId.value, database: activeDatabase.value, sql })
    const data = Array.isArray(res) ? res : (res?.data || [])
    const columns = data.length ? Object.keys(data[0]) : []
    const rows: any[] = data.map((r: any) => {
      const row: any = {}
      for (const c of columns) row[c] = r[c]
      return row
    })
    addResult({
      key: uid(), kind: 'explain', label: 'Explain #' + (activeTab.value!.results.length + 1),
      columns, rows, message: '', affectedRows: 0, durationMs: 0,
      page: 1, pageSize: defaultPageSize.value, sql
    })
  } catch (e: any) {
    ElMessage.error('Explain 失败: ' + (e?.message || ''))
  } finally {
    executing.value = false
  }
}

// 光标所在语句（考虑字符串/注释内的分号）
function getStatementAtCursor(): string {
  if (!activeTab.value) return ''
  const sql = activeTab.value.sqlContent
  const ta = document.querySelector('.sql-textarea') as HTMLTextAreaElement | null
  const cursor = ta?.selectionStart ?? sql.length

  // 向前找到最近的顶级分号/开头
  let start = 0
  let inSingle = false, inDouble = false, inBacktick = false, inLine = false, inBlock = false
  for (let j = 0; j < cursor; j++) {
    const ch = sql[j]
    const next = sql[j + 1]
    if (inLine) { if (ch === '\n') inLine = false }
    else if (inBlock) { if (ch === '*' && next === '/') { inBlock = false; j++ } }
    else if (inSingle) { if (ch === "'" && sql[j - 1] !== '\\') inSingle = false }
    else if (inDouble) { if (ch === '"' && sql[j - 1] !== '\\') inDouble = false }
    else if (inBacktick) { if (ch === '`') inBacktick = false }
    else {
      if (ch === '-' && next === '-') { inLine = true; continue }
      if (ch === '/' && next === '*') { inBlock = true; continue }
      if (ch === '#') { inLine = true; continue }
      if (ch === "'") { inSingle = true; continue }
      if (ch === '"') { inDouble = true; continue }
      if (ch === '`') { inBacktick = true; continue }
      if (ch === ';') { start = j + 1 }
    }
  }
  // 向后找到最近的顶级分号/结尾
  let end = sql.length
  inSingle = inDouble = inBacktick = inLine = inBlock = false
  for (let j = cursor; j < sql.length; j++) {
    const ch = sql[j]
    const next = sql[j + 1]
    if (inLine) { if (ch === '\n') inLine = false }
    else if (inBlock) { if (ch === '*' && next === '/') { inBlock = false; j++ } }
    else if (inSingle) { if (ch === "'" && sql[j - 1] !== '\\') inSingle = false }
    else if (inDouble) { if (ch === '"' && sql[j - 1] !== '\\') inDouble = false }
    else if (inBacktick) { if (ch === '`') inBacktick = false }
    else {
      if (ch === '-' && next === '-') { inLine = true; continue }
      if (ch === '/' && next === '*') { inBlock = true; continue }
      if (ch === '#') { inLine = true; continue }
      if (ch === "'") { inSingle = true; continue }
      if (ch === '"') { inDouble = true; continue }
      if (ch === '`') { inBacktick = true; continue }
      if (ch === ';') { end = j + 1; break }
    }
  }
  return sql.substring(start, end).trim()
}

function executeWithRiskCheck(sql: string, all: boolean) {
  if (!sql.trim()) return
  const statements = all ? splitStatements(sql) : [sql]
  const joined = statements.join(';\n')

  // 高危检查
  if (checkRisk.value) {
    const risk = isRiskySql(joined)
    if (risk.risky) {
      pendingRiskSql.value = joined
      pendingExecAll.value = all
      riskConfirmed.value = false
      riskDialogVisible.value = true
      return
    }
  }
  realExecute(statements)
}

function confirmRiskAndExec() {
  if (!riskConfirmed.value) return
  riskDialogVisible.value = false
  const statements = splitStatements(pendingRiskSql.value)
  realExecute(statements)
}

async function realExecute(statements: string[]) {
  if (!activeDsId.value || !activeTab.value) return
  executing.value = true
  const tab = activeTab.value
  const t0 = Date.now()
  try {
    for (let i = 0; i < statements.length; i++) {
      const stmt = statements[i]
      const res: any = await executeSql({ datasourceId: activeDsId.value, database: activeDatabase.value, sql: stmt })
      const data: any = res?.data || res
      const results = Array.isArray(data) ? data : (Array.isArray(data?.results) ? data.results : [data])

      for (const one of results) {
        const durationMs = one?.durationMs ?? (Date.now() - t0)
        lastExecMs.value = durationMs

        if (one?.columns && one?.columns.length > 0) {
          const columns: string[] = one.columns
          const rows: any[] = (one.rows || []).map((r: any, _idx: number) => {
            const obj: any = { _idx }
            for (const c of columns) obj[c] = r[c]
            return obj
          })
          addResult({
            key: uid(),
            kind: 'table',
            label: '结果 #' + (tab.results.length + 1),
            columns, rows,
            message: '',
            affectedRows: one.affectedRows || 0,
            durationMs,
            page: 1, pageSize: defaultPageSize.value,
            sql: stmt
          })
        } else if (one?.success || one?.affectedRows !== undefined) {
          addResult({
            key: uid(), kind: 'message', label: '执行 #' + (tab.results.length + 1),
            columns: [], rows: [], message: one.message || '执行成功',
            affectedRows: one.affectedRows || 0, durationMs,
            page: 1, pageSize: defaultPageSize.value, sql: stmt
          })
        } else {
          addResult({
            key: uid(), kind: 'error', label: '失败 #' + (tab.results.length + 1),
            columns: [], rows: [], message: one?.message || '执行失败',
            affectedRows: 0, durationMs,
            page: 1, pageSize: defaultPageSize.value, sql: stmt
          })
        }
      }
    }
    if (tab) tab.modified = false
  } catch (e: any) {
    const msg = e?.message || '执行失败'
    if (activeTab.value) {
      addResult({
        key: uid(), kind: 'error', label: '失败 #' + (activeTab.value.results.length + 1),
        columns: [], rows: [], message: msg,
        affectedRows: 0, durationMs: Date.now() - t0,
        page: 1, pageSize: defaultPageSize.value, sql: statements.join(';\n')
      })
    }
    ElMessage.error(msg)
  } finally {
    executing.value = false
  }
}

function addResult(r: SqlResult) {
  if (!activeTab.value) return
  activeTab.value.results.push(r)
  activeTab.value.activeResultKey = r.key
}

function resultIcon(r: SqlResult) {
  if (r.kind === 'error') return CircleClose
  if (r.kind === 'message') return SuccessFilled
  if (r.kind === 'explain') return Cpu
  return InfoFilled
}
function resultIconColor(r: SqlResult) {
  if (r.kind === 'error') return '#f56c6c'
  if (r.kind === 'message') return '#67c23a'
  if (r.kind === 'explain') return '#e6a23c'
  return '#409eff'
}

function paginatedRows(r: SqlResult): any[] {
  if (!r.rows || r.rows.length === 0) return []
  const start = (r.page - 1) * r.pageSize
  return r.rows.slice(start, start + r.pageSize)
}

function onTableSortChange(r: SqlResult, { prop, order }: any) {
  if (!prop || !order) return
  r.rows.sort((a: any, b: any) => {
    const av = a[prop], bv = b[prop]
    if (av == null && bv == null) return 0
    if (av == null) return order === 'ascending' ? -1 : 1
    if (bv == null) return order === 'ascending' ? 1 : -1
    if (typeof av === 'number' && typeof bv === 'number') {
      return order === 'ascending' ? av - bv : bv - av
    }
    return order === 'ascending' ? String(av).localeCompare(String(bv)) : String(bv).localeCompare(String(av))
  })
}

// =================== 全局快捷键 ===================
function globalKey(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 't' && !e.shiftKey) {
    e.preventDefault()
    addTab()
  }
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'w' && !e.shiftKey) {
    if (activeTabKey.value && document.activeElement?.tagName !== 'TEXTAREA') {
      e.preventDefault()
      removeTab(activeTabKey.value)
    }
  }
}

// =================== 拖拽调整布局 ===================
let dragV: any = null
function startDragV(e: MouseEvent) {
  dragV = { x: e.clientX, start: leftWidth.value }
  document.addEventListener('mousemove', onDragV)
  document.addEventListener('mouseup', endDragV)
}
function onDragV(e: MouseEvent) {
  if (!dragV) return
  const delta = e.clientX - dragV.x
  leftWidth.value = Math.max(180, Math.min(600, dragV.start + delta))
  persistSettings()
}
function endDragV() {
  dragV = null
  document.removeEventListener('mousemove', onDragV)
  document.removeEventListener('mouseup', endDragV)
  applyEditorHeight()
}

let dragH: any = null
function startDragH(e: MouseEvent) {
  dragH = { y: e.clientY, start: editorHeight.value }
  document.addEventListener('mousemove', onDragH)
  document.addEventListener('mouseup', endDragH)
}
function onDragH(e: MouseEvent) {
  if (!dragH) return
  const delta = e.clientY - dragH.y
  const rightPane: any = document.querySelector('.right-pane')
  const total = rightPane ? rightPane.clientHeight - 52 : 620
  const newH = Math.max(120, Math.min(total - 160, dragH.start + delta))
  editorHeight.value = newH
  editorHeightRatio.value = Math.round(newH * 100 / total)
}
function endDragH() {
  dragH = null
  document.removeEventListener('mousemove', onDragH)
  document.removeEventListener('mouseup', endDragH)
  persistSettings()
}

// window resize
function onWindowResize() {
  applyEditorHeight()
}
if (typeof window !== 'undefined') {
  window.addEventListener('resize', onWindowResize)
}
</script>

<style scoped>
.workbench-root {
  height: calc(100vh - 48px);
  background: #f5f7fa;
  display: flex;
  flex-direction: column;
  color: #303133;
  font-size: 13px;
}
.workbench-root.dark-theme {
  background: #1e1e2a;
  color: #dcdfe6;
}

/* 顶部工具栏 */
.top-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: #ffffff;
  border-bottom: 1px solid #e4e7ed;
  flex-wrap: wrap;
}
.dark-theme .top-toolbar { background: #2b2b3d; border-color: #3a3a4e; }

.workbench-title {
  font-weight: 600;
  font-size: 15px;
  color: #303133;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-right: 8px;
}
.dark-theme .workbench-title { color: #e6a23c; }

.toolbar-left, .toolbar-center, .toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.toolbar-center { margin: 0 auto; }

.ds-select { width: 200px; }
.db-select { width: 180px; }
.ds-type-tag {
  color: #909399;
  font-size: 12px;
  margin-left: 8px;
}

/* 主体 */
.body-row {
  flex: 1;
  display: flex;
  overflow: hidden;
  padding: 0;
}

.left-pane, .right-pane {
  background: #ffffff;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.dark-theme .left-pane, .dark-theme .right-pane { background: #2b2b3d; }

.pane-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
}
.dark-theme .pane-header { border-color: #3a3a4e; }

.pane-actions { display: flex; gap: 6px; align-items: center; }

.object-tree {
  flex: 1;
  overflow: auto;
  padding: 4px 0 8px 0;
  background: transparent;
}
.tree-node {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}
.tn-icon { color: #409eff; }
.tn-hint { color: #909399; font-size: 11px; margin-left: 6px; }

.tree-loading, .conn-error-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #909399;
  padding: 32px;
}
.conn-error-title { color: #f56c6c; font-weight: 600; }
.conn-error-msg { font-size: 12px; max-width: 220px; text-align: center; }

/* 拖拽分割线 */
.splitter-v {
  width: 6px;
  background: #ebeef5;
  cursor: col-resize;
  flex-shrink: 0;
}
.dark-theme .splitter-v { background: #3a3a4e; }

.splitter-h {
  height: 6px;
  background: #ebeef5;
  cursor: row-resize;
}
.dark-theme .splitter-h { background: #3a3a4e; }

/* 右侧 */
.sql-tabs-bar {
  padding: 6px 12px 0 12px;
  background: #ffffff;
  border-bottom: 1px solid #e4e7ed;
}
.dark-theme .sql-tabs-bar { background: #2b2b3d; border-color: #3a3a4e; }

.tab-inner {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.editor-pane {
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  margin: 6px;
  border-radius: 4px;
  overflow: hidden;
}
.dark-theme .editor-pane { background: #1e1e2a; border-color: #3a3a4e; }

.editor-toolbar {
  display: flex;
  align-items: center;
  padding: 4px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  gap: 10px;
  font-size: 12px;
}
.dark-theme .editor-toolbar { background: #252534; border-color: #3a3a4e; }

.editor-toolbar-right { margin-left: auto; }
.sql-type-tag {
  display: inline-block;
  padding: 2px 8px;
  background: #409eff;
  color: #fff;
  font-size: 11px;
  border-radius: 3px;
  font-weight: 600;
}
.hint-text { color: #909399; font-size: 12px; }

.sql-textarea {
  flex: 1;
  padding: 12px 14px;
  font-family: Consolas, Monaco, 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.7;
  border: none;
  outline: none;
  resize: none;
  background: #ffffff;
  color: #303133;
  white-space: pre;
  overflow: auto;
  tab-size: 4;
}
.dark-theme .sql-textarea { background: #1e1e2a; color: #dcdfe6; }

/* 结果区 */
.result-pane {
  margin: 0 6px 6px 6px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #ffffff;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.dark-theme .result-pane { background: #2b2b3d; border-color: #3a3a4e; }

.result-tab-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.pager-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 8px 12px;
  gap: 12px;
  border-top: 1px solid #ebeef5;
}
.dark-theme .pager-row { border-color: #3a3a4e; }

.empty-result, .empty-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
  padding: 40px;
  gap: 12px;
}

.result-error, .result-message {
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.null-cell {
  color: #c0c4cc;
  font-style: italic;
  font-size: 12px;
}

.mod-dot {
  color: #e6a23c;
  margin-left: 4px;
  font-size: 10px;
}

/* 右键菜单 */
.ctx-menu {
  position: fixed;
  z-index: 9999;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 4px 0;
  list-style: none;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  min-width: 180px;
}
.dark-theme .ctx-menu { background: #2b2b3d; border-color: #3a3a4e; }
.ctx-menu li {
  padding: 8px 16px;
  cursor: pointer;
  font-size: 13px;
}
.ctx-menu li:hover { background: #ecf5ff; color: #409eff; }
.dark-theme .ctx-menu li:hover { background: #36364a; color: #e6a23c; }

/* 历史 / 表信息 */
.history-search { margin-bottom: 12px; }
.history-sql {
  background: #f5f7fa;
  padding: 8px 10px;
  border-radius: 3px;
  font-family: Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  margin-bottom: 6px;
  max-height: 120px;
  overflow: auto;
}
.dark-theme .history-sql { background: #1e1e2a; color: #dcdfe6; }
.history-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.sub-title {
  margin: 16px 0 8px 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  border-left: 3px solid #409eff;
  padding-left: 8px;
}
.dark-theme .sub-title { color: #e6a23c; border-left-color: #e6a23c; }
.ddl-box {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  font-family: Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  max-height: 260px;
  overflow: auto;
}
.dark-theme .ddl-box { background: #1e1e2a; color: #dcdfe6; }

.risk-sql {
  background: #fef0f0;
  padding: 12px;
  border-radius: 4px;
  font-family: Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  margin: 12px 0;
  max-height: 240px;
  overflow: auto;
}
.risk-confirm { margin-top: 8px; }
</style>
