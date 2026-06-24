import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type {
  AnyTabState,
  QueryTabState,
  TableTabState,
  StatementResult,
  TreeNode,
  QueryHistoryItem,
  QueryMessage
} from '../types'
import { executeSql as apiExecute, explainSql as apiExplain, queryTableData, getTableDDL, listHistory } from '../service/api'
import { resolveEffectiveDatabase, nowString, escapeIdentifier, copyToClipboard } from '../utils/sql'

const tabs = reactive<AnyTabState[]>([])
const activeTabId = ref<string>('')
const executing = ref<boolean>(false)

let tabCounter = 0

function nextId(prefix: string): string {
  tabCounter += 1
  return prefix + '-' + Date.now().toString(36) + '-' + tabCounter
}

function emptyResult(): QueryTabState['result'] {
  return {
    activeName: 'result',
    activeResultIdx: '0',
    statementResults: [],
    explain: [],
    explainCols: [],
    messages: []
  }
}

function addQueryTab(datasourceId: string, database: string, preSql = '', title = '查询'): QueryTabState {
  const existingCount = tabs.filter((t) => t.kind === 'query').length + 1
  const tab: QueryTabState = {
    id: nextId('q'),
    kind: 'query',
    title: title || ('查询 ' + existingCount),
    datasourceId,
    database: database || '',
    sql: preSql || '',
    result: emptyResult(),
    history: []
  }
  tabs.push(tab)
  activeTabId.value = tab.id
  
  // 自动加载历史记录
  fetchHistoryForTab(tab)
  return tab
}

function addTableTab(datasourceId: string, database: string, table: string): TableTabState {
  const existing = tabs.find(
    (t) => t.kind === 'table' && t.datasourceId === datasourceId && (t as TableTabState).table === table && t.database === database
  ) as TableTabState | undefined
  if (existing) {
    activeTabId.value = existing.id
    loadTableData(existing).catch(() => undefined)
    return existing
  }
  const tab: TableTabState = {
    id: nextId('t'),
    kind: 'table',
    title: '表 · ' + table,
    datasourceId,
    database,
    table,
    columns: [],
    rows: [],
    tableLoading: false
  }
  tabs.push(tab)
  activeTabId.value = tab.id
  loadTableData(tab).catch(() => undefined)
  return tab
}

function closeTab(id: string) {
  const idx = tabs.findIndex((t) => t.id === id)
  if (idx < 0) return
  tabs.splice(idx, 1)
  if (activeTabId.value === id) {
    activeTabId.value = tabs.length > 0 ? tabs[Math.max(0, idx - 1)].id : ''
  }
}

function switchTab(id: string) {
  activeTabId.value = id
}

function findActive(): AnyTabState | undefined {
  return tabs.find((t) => t.id === activeTabId.value)
}

function updateTab(id: string, patch: Partial<QueryTabState & TableTabState>) {
  const tab = tabs.find((t) => t.id === id)
  if (!tab) return
  Object.assign(tab, patch)
}

async function runEditorSql(tab: QueryTabState, databases: string[]): Promise<StatementResult[]> {
  if (!tab.datasourceId) {
    ElMessage.warning('请先选择数据源')
    return []
  }
  const sql = (tab.sql || '').trim()
  if (!sql) {
    ElMessage.warning('请输入 SQL 语句')
    return []
  }
  executing.value = true
  tab.result.activeName = 'result'
  tab.result.statementResults = []
  const start = Date.now()
  pushMessage(tab, 'INFO', '开始执行 SQL（数据库优先级：显式引用 > 当前下拉数据库）')
  const effectiveDb = resolveEffectiveDatabase(sql, tab.database, databases)
  try {
    const results = await apiExecute({ datasourceId: tab.datasourceId, database: effectiveDb, sql, ignoreRisk: true })
    const mapped: StatementResult[] = results.map((r: any, i: number) => {
      const isSelect = !!r?.isSelect || (Array.isArray(r?.columns) && r.columns.length > 0) || false
      const columns: string[] = Array.isArray(r?.columns) ? r.columns : Array.isArray(r?.columnNames) ? r.columnNames : []
      const rows: Record<string, any>[] = Array.isArray(r?.rows) ? r.rows : Array.isArray(r?.data) ? r.data : []
      const sr: StatementResult = {
        sql: r?.sql || sql,
        isSelect,
        columns,
        rows,
        affectedRows: typeof r?.affectedRows === 'number' ? r.affectedRows : 0,
        durationMs: typeof r?.durationMs === 'number' ? r.durationMs : Math.max(0, Date.now() - start - i * 20),
        success: r?.success !== false && !r?.error,
        message: r?.message || (r?.success !== false ? '执行成功' : '执行失败'),
        effectiveDatabase: effectiveDb
      }
      return sr
    })
    tab.result.statementResults = mapped
    tab.result.activeResultIdx = '0'
    for (const sr of mapped) {
      pushMessage(
        tab,
        sr.success ? 'INFO' : 'ERROR',
        (sr.success ? 'OK' : 'FAILED') + ' · ' + sr.message + ' · 生效数据库：' + (sr.effectiveDatabase || '(默认)')
      )
      tab.history.unshift({ time: nowString(), sql: sr.sql, database: sr.effectiveDatabase, success: sr.success, durationMs: sr.durationMs })
    }
    return mapped
  } catch (e: any) {
    const msg = e?.message || String(e)
    const fallback: StatementResult = {
      sql,
      isSelect: false,
      columns: [],
      rows: [],
      affectedRows: 0,
      durationMs: Date.now() - start,
      success: false,
      message: msg,
      effectiveDatabase: effectiveDb
    }
    tab.result.statementResults = [fallback]
    pushMessage(tab, 'ERROR', msg)
    return [fallback]
  } finally {
    executing.value = false
  }
}

async function explainForTab(tab: QueryTabState, databases: string[]) {
  if (!tab.datasourceId) {
    ElMessage.warning('请先选择数据源')
    return
  }
  const sql = (tab.sql || '').trim()
  if (!sql) {
    ElMessage.warning('请输入 SQL 语句')
    return
  }
  executing.value = true
  tab.result.activeName = 'explain'
  const effectiveDb = resolveEffectiveDatabase(sql, tab.database, databases)
  try {
    const rows = await apiExplain({ datasourceId: tab.datasourceId, database: effectiveDb, sql, ignoreRisk: true })
    tab.result.explain = rows
    tab.result.explainCols = rows.length > 0 ? Object.keys(rows[0]) : []
  } catch (e: any) {
    ElMessage.error(e?.message || String(e))
  } finally {
    executing.value = false
  }
}

async function loadTableData(tab: TableTabState) {
  if (!tab.datasourceId || !tab.table) return
  tab.tableLoading = true
  try {
    const res = await queryTableData({ datasourceId: tab.datasourceId, database: tab.database, table: tab.table })
    tab.rows = res.rows || []
    tab.columns = res.columns || (tab.rows.length > 0 ? Object.keys(tab.rows[0]) : [])
  } catch (e: any) {
    ElMessage.error('表数据加载失败：' + (e?.message || String(e)))
  } finally {
    tab.tableLoading = false
  }
}

async function fetchDDL(datasourceId: string, database: string, table: string): Promise<string> {
  try {
    return await getTableDDL({ datasourceId, database, table })
  } catch (e: any) {
    ElMessage.error('获取 DDL 失败：' + (e?.message || String(e)))
    return ''
  }
}

async function fetchHistoryForTab(tab: QueryTabState, keyword = '') {
  try {
    const list = await listHistory({ datasourceId: tab.datasourceId || '', page: 1, pageSize: 100, keyword })
    tab.history = list.map((item: any) => ({
      time: item.time || item.createdAt || item.created_at || nowString(),
      sql: item.sql || item.sqlText || item.sql_text || '',
      database: item.database || item.databaseName || item.database_name || '',
      success: String(item.status || item.success || '').toLowerCase() !== 'failed' && item.success !== false,
      durationMs: typeof item.durationMs === 'number' ? item.durationMs : item.duration
    }))
  } catch (e: any) {
    ElMessage.warning('历史记录加载失败：' + (e?.message || String(e)))
  }
}

function pushMessage(tab: QueryTabState, level: QueryMessage['level'], text: string) {
  tab.result.messages.push({ time: nowString(), level, text })
}

function openNodeInTab(node: TreeNode, datasourceId: string): QueryTabState | TableTabState | undefined {
  const db = node.database || ''
  if (node.type === 'table' || node.type === 'view') {
    return addQueryTab(datasourceId, db, 'SELECT * FROM ' + qualifiedName(db, node.table || node.name) + ' LIMIT 200;', '查询 ' + node.name)
  }
  if (node.type === 'function' || node.type === 'procedure') {
    return addQueryTab(datasourceId, db, 'SELECT ' + qualifiedName(db, node.name) + '();', '查询 ' + node.name)
  }
  return addQueryTab(datasourceId, db, '', '查询')
}

function qualifiedName(db: string, name: string): string {
  const prefix = db ? escapeIdentifier(db) + '.' : ''
  return prefix + escapeIdentifier(name)
}

async function copyName(name: string) {
  try {
    await copyToClipboard(name)
    ElMessage.success('已复制：' + name)
  } catch {
    ElMessage.success(name)
  }
}

export function useSqlIdeTabs() {
  return {
    tabs,
    activeTabId,
    executing,
    addQueryTab,
    addTableTab,
    closeTab,
    switchTab,
    findActive,
    updateTab,
    runEditorSql,
    explainForTab,
    loadTableData,
    fetchDDL,
    fetchHistoryForTab,
    openNodeInTab,
    qualifiedName,
    copyName
  }
}
