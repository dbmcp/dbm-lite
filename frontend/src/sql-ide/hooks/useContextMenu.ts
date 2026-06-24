import { reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import type { TreeNode } from '../types'
import { useSqlIdeTabs } from './useTabs'
import { useDatasource } from './useDatasource'

export interface DatasourceCtxInfo {
  datasourceId: string
  name: string
  dbType?: string
  connected: boolean
}

interface ContextMenuState {
  show: boolean
  x: number
  y: number
  node: TreeNode | null
  datasource: DatasourceCtxInfo | null
}

const ctxMenu = reactive<ContextMenuState>({ show: false, x: 0, y: 0, node: null, datasource: null })

function closeMenuOnClickOutside(e: MouseEvent) {
  if (ctxMenu.show) {
    const target = e.target as HTMLElement
    if (!target.closest('.ctx-menu')) {
      closeMenu()
    }
  }
}

function escapeIdentifier(name: string): string {
  if (!name) return ''
  const backtick = String.fromCharCode(96)
  return backtick + String(name).split(backtick).join('') + backtick
}

async function copyToClipboard(text: string) {
  try {
    if (navigator && navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text)
    } else {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    }
    ElMessage.success('已复制: ' + text.substring(0, 50))
  } catch (e: any) {
    ElMessage.success(text.substring(0, 50))
  }
}

function setActiveTabDatasource(datasourceId: string, database?: string) {
  const tabs = useSqlIdeTabs()
  const active = tabs.findActive() as any
  if (active) {
    if (datasourceId) active.datasourceId = datasourceId
    if (database !== undefined && database !== '') active.database = database
  }
}

function setActiveTabSql(sql: string) {
  const tabs = useSqlIdeTabs()
  const active = tabs.findActive() as any
  if (active) {
    active.sql = sql
  }
}

export function openMenu(e: MouseEvent, node: TreeNode) {
  if (e && e.preventDefault) e.preventDefault()
  if (e && e.stopPropagation) e.stopPropagation()
  ctxMenu.node = node
  ctxMenu.datasource = null
  ctxMenu.x = e.clientX
  ctxMenu.y = e.clientY
  ctxMenu.show = true
  document.addEventListener('click', closeMenuOnClickOutside)
}

export function openDatasourceMenu(e: MouseEvent, ds: DatasourceCtxInfo) {
  if (e && e.preventDefault) e.preventDefault()
  if (e && e.stopPropagation) e.stopPropagation()
  ctxMenu.node = null
  ctxMenu.datasource = ds
  ctxMenu.x = e.clientX
  ctxMenu.y = e.clientY
  ctxMenu.show = true
  document.addEventListener('click', closeMenuOnClickOutside)
}

export function closeMenu() {
  ctxMenu.show = false
  ctxMenu.node = null
  ctxMenu.datasource = null
  document.removeEventListener('click', closeMenuOnClickOutside)
}

export async function handleClick(action: string) {
  const node: any = ctxMenu.node
  const dsInfo: any = ctxMenu.datasource
  const tabs = useSqlIdeTabs()
  const dsState = useDatasource()
  
  closeMenu()
  
  await new Promise(resolve => setTimeout(resolve, 0))

  if (dsInfo) {
    if (action === 'ds-connect') {
      if (!dsInfo.connected) {
        await dsState.loadTree(dsInfo.datasourceId)
      } else {
        await dsState.loadTree(dsInfo.datasourceId, true)
      }
      return
    }
    if (action === 'ds-refresh') {
      await dsState.loadTree(dsInfo.datasourceId, true)
      return
    }
    if (action === 'ds-close') {
      if (dsState.trees && dsState.trees[dsInfo.datasourceId]) {
        dsState.trees[dsInfo.datasourceId].loaded = false
      }
      ElMessage.success('连接已关闭')
      return
    }
    if (action === 'ds-query') {
      const tab = tabs.addQueryTab(dsInfo.datasourceId, '', '-- 在此编写 SQL\nSELECT 1;\n', '查询')
      setActiveTabDatasource(dsInfo.datasourceId)
      return
    }
    if (action === 'ds-copy-name') {
      copyToClipboard(dsInfo.name || '')
      return
    }
    return
  }

  if (!node) return
  const db = node.database || ''
  const table = node.table || node.name
  const datasourceId = node.datasourceId || (dsState as any).currentId || ''

  if (action === 'set-db') {
    setActiveTabDatasource(datasourceId, db)
    ElMessage.success('已切换数据库：' + db)
    return
  }

  if (action === 'db-query' || action === 'query') {
    const tab = tabs.addQueryTab(datasourceId, db, '-- 在此编写 SQL\nSELECT 1;\n', '查询')
    setActiveTabDatasource(datasourceId, db)
    return
  }

  if (action === 'open-table') {
    tabs.addTableTab(datasourceId, db, table)
    return
  }

  if (action === 'select-top') {
    const limit = 200
    const sql = 'SELECT * FROM ' + (db ? escapeIdentifier(db) + '.' : '') + escapeIdentifier(table) + ' LIMIT ' + limit + ';'
    const tab = tabs.addQueryTab(datasourceId, db, sql, '查询 ' + table)
    setActiveTabDatasource(datasourceId, db)
    tabs.runEditorSql(tab as any, [db]).catch(() => undefined)
    return
  }

  if (action === 'select-top-1000') {
    const sql = 'SELECT * FROM ' + (db ? escapeIdentifier(db) + '.' : '') + escapeIdentifier(table) + ' LIMIT 1000;'
    const tab = tabs.addQueryTab(datasourceId, db, sql, '查询 ' + table)
    setActiveTabDatasource(datasourceId, db)
    tabs.runEditorSql(tab as any, [db]).catch(() => undefined)
    return
  }

  if (action === 'select-all') {
    const sql = 'SELECT * FROM ' + (db ? escapeIdentifier(db) + '.' : '') + escapeIdentifier(table) + ';'
    const tab = tabs.addQueryTab(datasourceId, db, sql, '查询 ' + table)
    setActiveTabDatasource(datasourceId, db)
    return
  }

  if (action === 'insert-template') {
    const sql = 'INSERT INTO ' + (db ? escapeIdentifier(db) + '.' : '') + escapeIdentifier(table) + ' (/* columns */) VALUES (/* values */);'
    const tab = tabs.addQueryTab(datasourceId, db, sql, '插入 ' + table)
    setActiveTabDatasource(datasourceId, db)
    return
  }

  if (action === 'update-template') {
    const sql = 'UPDATE ' + (db ? escapeIdentifier(db) + '.' : '') + escapeIdentifier(table) + ' SET /* column = value */ WHERE /* condition */;'
    const tab = tabs.addQueryTab(datasourceId, db, sql, '更新 ' + table)
    setActiveTabDatasource(datasourceId, db)
    return
  }

  if (action === 'delete-template') {
    const sql = 'DELETE FROM ' + (db ? escapeIdentifier(db) + '.' : '') + escapeIdentifier(table) + ' WHERE /* condition */;'
    const tab = tabs.addQueryTab(datasourceId, db, sql, '删除 ' + table)
    setActiveTabDatasource(datasourceId, db)
    return
  }

  if (action === 'count') {
    const sql = 'SELECT COUNT(*) AS cnt FROM ' + (db ? escapeIdentifier(db) + '.' : '') + escapeIdentifier(table) + ';'
    const tab = tabs.addQueryTab(datasourceId, db, sql, '统计 ' + table)
    setActiveTabDatasource(datasourceId, db)
    tabs.runEditorSql(tab as any, [db]).catch(() => undefined)
    return
  }

  if (action === 'ddl') {
    let active: any = tabs.findActive()
    if (!active) {
      active = tabs.addQueryTab(datasourceId, db, '', 'DDL ' + table)
    } else {
      active.database = db
      active.datasourceId = datasourceId
    }
    if (active._ddlText === undefined) active._ddlText = ''
    tabs.switchTab(active.id)
    tabs.fetchDDL(datasourceId, db, table).then((text: string) => {
      if (text) {
        active._ddlText = text
      }
    })
    setActiveTabDatasource(datasourceId, db)
    return
  }

  if (action === 'copy-name') {
    copyToClipboard(node.name || '')
    return
  }
  if (action === 'copy-qualified') {
    copyToClipboard((db ? escapeIdentifier(db) + '.' : '') + escapeIdentifier(table))
    return
  }
  if (action === 'insert-col') {
    const active = tabs.findActive() as any
    if (active && active.kind === 'query') {
      active.sql = (active.sql || '') + (active.sql ? ' ' : '') + escapeIdentifier(node.name || '')
      ElMessage.success('已插入列名到编辑器')
    } else {
      copyToClipboard(node.name || '')
    }
    return
  }
  if (action === 'refresh' || action === 'db-refresh') {
    if (datasourceId) {
      dsState.loadTree(datasourceId, true).then(() => {
        ElMessage.success('刷新成功')
      }).catch(() => undefined)
    }
    return
  }

  if (action === 'open-saved-queries') {
    const savedQueriesModal = (window as any).__sqlSavedQueriesModal
    if (savedQueriesModal && savedQueriesModal.open) {
      savedQueriesModal.open(node)
    }
    return
  }
}

export function useSqlIdeContextMenu() {
  return {
    ctxMenu,
    openMenu,
    openDatasourceMenu,
    closeMenu,
    handleClick
  }
}
