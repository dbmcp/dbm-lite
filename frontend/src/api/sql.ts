import request from './request'

export interface SqlExecuteParams {
  datasourceId: string
  database: string
  sql: string
  ignoreRisk?: boolean
}

export interface SqlExecuteResult {
  success: boolean
  isSelect: boolean
  message: string
  columns: string[]
  rows: Record<string, any>[]
  affectedRows: number
  durationMs: number
  review?: any
}

export interface SqlHistoryItem {
  historyId: number
  datasourceId: string
  datasourceName: string
  userId: string
  username: string
  databaseName: string
  sqlText: string
  sqlType: string
  affectedRows: number
  rowCount: number
  executeMs: number
  status: string
  errorMsg: string
  createdAt: string
}

export function executeSql(params: SqlExecuteParams) {
  return request<SqlExecuteResult[]>({
    url: '/sql/execute',
    method: 'POST',
    data: params
  })
}

export function explainSql(params: SqlExecuteParams) {
  return request({
    url: '/sql/explain',
    method: 'POST',
    data: params
  })
}

export function getTableInfo(id: string, database: string, table: string) {
  return request({
    url: '/sql/datasources/' + id + '/table-info',
    method: 'GET',
    params: { database, table }
  })
}

export function listSqlHistory(current: number, pageSize: number, datasourceId?: string, keyword?: string) {
  return request({
    url: '/sql/history',
    method: 'GET',
    params: { current, pageSize, datasourceId, keyword }
  })
}

export function listAuditLogs(current: number, pageSize: number, action?: string, username?: string, keyword?: string, category?: string) {
  return request({
    url: '/audit/logs',
    method: 'GET',
    params: { current, pageSize, action, username, keyword, category }
  })
}

export function getAuditStats() {
  return request({
    url: '/audit/stats',
    method: 'GET'
  })
}

// ============ SQL IDE V2 API (规范路由前缀 /api/dataquery/) ============

export function executeSqlV2(params: SqlExecuteParams) {
  return request<SqlExecuteResult[]>({
    url: '/dataquery/sql/execute',
    method: 'POST',
    data: params
  })
}

export function explainSqlV2(params: SqlExecuteParams) {
  return request({
    url: '/dataquery/sql/explain',
    method: 'POST',
    data: params
  })
}

export function listSqlHistoryV2(current: number, pageSize: number, datasourceId?: string, keyword?: string) {
  return request({
    url: '/dataquery/sqlHistory/list',
    method: 'GET',
    params: { current, pageSize, datasourceId, keyword }
  })
}

export function getDatabasesV2(id: string) {
  return request<string[]>({
    url: '/dataquery/datasources/' + id + '/databases',
    method: 'GET'
  })
}

export function listTablesV2(id: string, database: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/tables',
    method: 'GET',
    params: { database }
  })
}

export function listColumnsV2(id: string, database: string, table: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/columns',
    method: 'GET',
    params: { database, table }
  })
}

export function getFullTreeV2(id: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/tree',
    method: 'GET'
  })
}

export function getTableInfoV2(id: string, database: string, table: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/table-info',
    method: 'GET',
    params: { database, table }
  })
}
