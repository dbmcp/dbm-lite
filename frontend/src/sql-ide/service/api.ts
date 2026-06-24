import request from '@/api/request'
import type { DatasourceSummary } from '../types'

export interface SqlExecuteParams {
  datasourceId: string
  database: string
  sql: string
  ignoreRisk?: boolean
}

export interface ApiStatementResult {
  sql?: string
  isSelect?: boolean
  columns?: string[]
  columnNames?: string[]
  rows?: Record<string, any>[]
  data?: Record<string, any>[]
  affectedRows?: number
  durationMs?: number
  success?: boolean
  error?: any
  message?: string
}

export function listDatasources(): Promise<DatasourceSummary[]> {
  return request({
    url: '/datasources/all',
    method: 'GET'
  }).then((res: any) => {
    let list: any[] = []
    if (Array.isArray(res)) {
      list = res
    } else if (res && Array.isArray(res.list)) {
      list = res.list
    } else if (res && Array.isArray(res.data)) {
      list = res.data
    } else if (res && Array.isArray(res.rows)) {
      list = res.rows
    } else if (res && Array.isArray(res.items)) {
      list = res.items
    }
    if (!Array.isArray(list)) list = []
    list = list.filter((d: any) => d && typeof d === 'object')
    return list.map((d: any) => ({
      datasourceId: d.datasourceId || d.id || (d.id !== undefined ? String(d.id) : 'ds_' + Math.random().toString(36).slice(2, 8)),
      name: d.name || String(d.datasourceId || d.id || 'unknown'),
      dbType: d.dbType || d.type || 'mysql'
    }))
  })
}

export function executeSql(params: SqlExecuteParams): Promise<ApiStatementResult[]> {
  return request({
    url: '/dataquery/sql/execute',
    method: 'POST',
    data: params
  }).then((res: any) => {
    if (Array.isArray(res)) return res
    if (res && Array.isArray(res.results)) return res.results
    if (res && Array.isArray(res.data)) return res.data
    return res ? [res] : []
  })
}

export function explainSql(params: SqlExecuteParams): Promise<Record<string, any>[]> {
  return request({
    url: '/dataquery/sql/explain',
    method: 'POST',
    data: params
  }).then((res: any) => {
    if (Array.isArray(res)) return res
    if (res && Array.isArray(res.data)) return res.data
    if (res && Array.isArray(res.rows)) return res.rows
    return []
  })
}

export function getDatabases(datasourceId: string): Promise<string[]> {
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/databases',
    method: 'GET'
  }).then((res: any) => {
    if (Array.isArray(res)) return res as string[]
    if (res && Array.isArray(res.data)) return res.data as string[]
    return (res?.list || []) as string[]
  })
}

export function getFullTree(datasourceId: string): Promise<any[]> {
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/tree',
    method: 'GET'
  }).then((res: any) => {
    let result: any[] = []
    if (Array.isArray(res)) {
      result = res
    } else if (res && Array.isArray(res.data)) {
      result = res.data
    } else if (res && Array.isArray(res.list)) {
      result = res.list
    } else if (res && Array.isArray(res.nodes)) {
      result = res.nodes
    } else if (res && Array.isArray((res as any).tree)) {
      result = (res as any).tree
    }
    if (!Array.isArray(result)) result = []
    return result
  })
}

export function queryTableData(params: {
  datasourceId: string
  database: string
  table: string
  page?: number
  pageSize?: number
}): Promise<{ rows: Record<string, any>[]; columns: string[] }> {
  const { datasourceId, database, table, page = 1, pageSize = 200 } = params
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/data/query',
    method: 'GET',
    params: { database, table, page, pageSize }
  }).then((res: any) => {
    let rows: Record<string, any>[] = []
    let columns: string[] = []
    if (Array.isArray(res)) rows = res
    else if (res && Array.isArray(res.rows)) {
      rows = res.rows
      columns = res.columns || []
    } else if (res && Array.isArray(res.data)) rows = res.data
    else if (res && Array.isArray(res.list)) rows = res.list
    if (columns.length === 0 && rows.length > 0) columns = Object.keys(rows[0])
    return { rows, columns }
  })
}

export function getTableDDL(params: { datasourceId: string; database: string; table: string }): Promise<string> {
  return request({
    url: '/dataquery/datasources/' + params.datasourceId + '/ddl',
    method: 'GET',
    params: { database: params.database, table: params.table }
  }).then((res: any) => {
    if (typeof res === 'string') return res
    if (res && typeof res.ddl === 'string') return res.ddl
    if (res && typeof res.sql === 'string') return res.sql
    if (res && res.data && typeof res.data === 'string') return res.data
    return JSON.stringify(res)
  })
}

export function listHistory(params: {
  datasourceId: string
  page?: number
  pageSize?: number
  keyword?: string
}): Promise<any[]> {
  const { datasourceId, page = 1, pageSize = 100, keyword = '' } = params
  return request({
    url: '/dataquery/sqlHistory/list',
    method: 'GET',
    params: { datasourceId, page, pageSize, keyword }
  }).then((res: any) => {
    if (Array.isArray(res)) return res
    if (res && Array.isArray(res.list)) return res.list
    if (res && Array.isArray(res.data)) return res.data
    return []
  })
}

export function listSavedQueries(params: {
  datasourceId?: string
  page?: number
  pageSize?: number
  keyword?: string
}): Promise<any[]> {
  const { datasourceId = '', page = 1, pageSize = 100, keyword = '' } = params
  return request({
    url: '/dataquery/savedQueries',
    method: 'GET',
    params: { datasourceId, page, pageSize, keyword }
  }).then((res: any) => {
    if (Array.isArray(res)) return res
    if (res && Array.isArray(res.list)) return res.list
    if (res && Array.isArray(res.data)) return res.data
    return []
  })
}

export function saveQuery(params: {
  datasourceId: string
  database: string
  title: string
  description?: string
  sql: string
}): Promise<{ queryId: string }> {
  return request({
    url: '/dataquery/savedQueries',
    method: 'POST',
    data: params
  }) as Promise<{ queryId: string }>
}

export function deleteSavedQuery(queryId: string): Promise<void> {
  return request({
    url: '/dataquery/savedQueries/' + queryId,
    method: 'DELETE'
  })
}

export function listFavorites(params: {
  page?: number
  pageSize?: number
  keyword?: string
}): Promise<any[]> {
  const { page = 1, pageSize = 100, keyword = '' } = params
  return request({
    url: '/dataquery/favorites',
    method: 'GET',
    params: { page, pageSize, keyword }
  }).then((res: any) => {
    if (Array.isArray(res)) return res
    if (res && Array.isArray(res.list)) return res.list
    if (res && Array.isArray(res.data)) return res.data
    return []
  })
}

export function createFavorite(params: {
  title: string
  description?: string
  sql: string
}): Promise<void> {
  return request({
    url: '/dataquery/favorites',
    method: 'POST',
    data: params
  })
}

export function deleteFavorite(favoriteId: string): Promise<void> {
  return request({
    url: '/dataquery/favorites/' + favoriteId,
    method: 'DELETE'
  })
}
