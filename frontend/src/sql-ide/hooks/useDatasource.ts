import { ref, reactive, onMounted } from 'vue'
import type { DatasourceSummary, TreeNode } from '../types'
import { listDatasources, getFullTree } from '../service/api'
import { normalizeNode } from '../utils/tree'

export interface DatasourceTreeEntry {
  loaded: boolean
  loading: boolean
  databases: string[]
  nodes: TreeNode[]
}

const state: {
  list: DatasourceSummary[]
  currentId: string
  trees: Record<string, DatasourceTreeEntry>
  expanded: Record<string, boolean>
  connected: Record<string, boolean>
} = reactive({
  list: [],
  currentId: '',
  trees: {},
  expanded: {},
  connected: {}
})

function ensureTree(datasourceId: string): DatasourceTreeEntry {
  if (!state.trees[datasourceId]) {
    state.trees[datasourceId] = reactive({
      loaded: false,
      loading: false,
      databases: [],
      nodes: []
    })
  }
  return state.trees[datasourceId]
}

function normalizeDbType(dbType: string): string {
  const t = (dbType || '').toLowerCase()
  if (t === 'mysql' || t === 'mariadb') return 'MySQL'
  if (t === 'tidb') return 'TiDB'
  if (t === 'sqlite' || t === 'sqlite3') return 'SQLite'
  if (t === 'postgres' || t === 'postgresql') return 'PostgreSQL'
  if (t === 'oracle') return 'Oracle'
  if (t === 'sqlserver' || t === 'mssql') return 'SQL Server'
  return dbType || 'Unknown'
}

function extractDbNames(nodes: any[]): string[] {
  const set = new Set<string>()
  const walk = (list: any[]) => {
    if (!Array.isArray(list)) return
    for (const n of list) {
      if (n && typeof n === 'object') {
        if (n.type === 'database') {
          if (n.name) set.add(n.name)
        } else if (n.database) {
          set.add(n.database)
        }
        if (Array.isArray(n.children)) walk(n.children)
      }
    }
  }
  walk(nodes)
  return Array.from(set)
}

export function useDatasource(): any {
  async function loadDatasources(): Promise<DatasourceSummary[]> {
    return await loadAll()
  }

  async function loadAll(): Promise<DatasourceSummary[]> {
    try {
      const list = await listDatasources()
      state.list = Array.isArray(list) ? list : []
      if (!state.currentId && state.list.length > 0) {
        const first = state.list[0] as any
        state.currentId = first.datasourceId || first.id || ''
      }
      return state.list
    } catch (e: any) {
      console.error('loadAll error:', e)
      state.list = []
      return []
    }
  }

  async function openConnection(datasourceId: string): Promise<TreeNode[]> {
    if (!datasourceId) return []
    state.connected[datasourceId] = true
    state.expanded[datasourceId] = true
    return await loadTree(datasourceId, true)
  }

  async function refreshConnection(datasourceId: string): Promise<TreeNode[]> {
    if (!datasourceId) return []
    return await loadTree(datasourceId, true)
  }

  function closeConnection(datasourceId: string) {
    state.connected[datasourceId] = false
    state.expanded[datasourceId] = false
    if (state.trees[datasourceId]) {
      state.trees[datasourceId].loaded = false
      state.trees[datasourceId].nodes = []
      state.trees[datasourceId].databases = []
    }
  }

  function toggleExpand(datasourceId: string) {
    state.expanded[datasourceId] = !state.expanded[datasourceId]
  }

  function isConnected(datasourceId: string): boolean {
    return !!state.connected[datasourceId]
  }

  function isLoading(datasourceId: string): boolean {
    const entry = state.trees[datasourceId]
    return !!(entry && entry.loading)
  }

  function isExpanded(datasourceId: string): boolean {
    return !!state.expanded[datasourceId]
  }

  async function loadTree(datasourceId: string, force = false): Promise<TreeNode[]> {
    if (!datasourceId) return []
    const entry = ensureTree(datasourceId)
    if (entry.loading) return entry.nodes
    if (entry.loaded && !force) return entry.nodes
    entry.loading = true
    try {
      const treeData = await getFullTree(datasourceId)
      let rawNodes: any[] = []
      if (Array.isArray(treeData)) {
        rawNodes = treeData
      } else if (treeData && typeof treeData === 'object') {
        const td: any = treeData
        rawNodes = Array.isArray(td.nodes) ? td.nodes : (Array.isArray(td.tree) ? td.tree : (Array.isArray(td.data) ? td.data : []))
      }
      entry.nodes = rawNodes.map(n => normalizeNode(n, '')).filter((n): n is TreeNode => n !== undefined)
      entry.databases = extractDbNames(entry.nodes)
      entry.loaded = true
      state.expanded[datasourceId] = true
      state.currentId = datasourceId
      return entry.nodes
    } catch (e: any) {
      console.error('loadTree error:', e)
      entry.nodes = []
      entry.databases = []
      return []
    } finally {
      entry.loading = false
    }
  }

  function treeFor(datasourceId: string): TreeNode[] {
    const entry = state.trees[datasourceId]
    return (entry && entry.nodes) ? entry.nodes : []
  }

  function setCurrentId(id: string) {
    state.currentId = id
  }

  const result = reactive({
    state,
    get list() { return state.list },
    set list(v: DatasourceSummary[]) { state.list = v },
    get currentId() { return state.currentId },
    set currentId(v: string) { state.currentId = v },
    get trees() { return state.trees },
    set trees(v: Record<string, DatasourceTreeEntry>) { state.trees = v },
    get expanded() { return state.expanded },
    set expanded(v: Record<string, boolean>) { state.expanded = v },
    loadAll,
    loadDatasources,
    openConnection,
    refreshConnection,
    closeConnection,
    toggleExpand,
    isConnected,
    isLoading,
    isExpanded,
    loadTree,
    treeFor,
    normalizeDbType,
    setCurrentId
  })
  return result
}

export { useDatasource as useSqlIdeDatasource }
