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
    url: '/dataquery/sql/execute',
    method: 'POST',
    data: params
  })
}

export function explainSql(params: SqlExecuteParams) {
  return request<Record<string, any>[]>({
    url: '/dataquery/sql/explain',
    method: 'POST',
    data: params
  })
}

export function cancelExecute(executionId: string) {
  return request<{ success: boolean; message: string }>({
    url: '/dataquery/sql/cancel',
    method: 'POST',
    data: { executionId }
  })
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

export function listSqlHistory(
  datasourceId: string,
  current = 1,
  pageSize = 20,
  keyword?: string
) {
  return request<{ list: SqlHistoryItem[]; total: number; current: number; pageSize: number }>({
    url: '/dataquery/sqlHistory/list',
    method: 'GET',
    params: { datasourceId, page: current, pageSize, keyword }
  })
}

export function getDatabases(id: string) {
  return request<string[]>({
    url: '/dataquery/datasources/' + id + '/databases',
    method: 'GET'
  })
}

export function listTables(id: string, database: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/tables',
    method: 'GET',
    params: { database }
  })
}

export function listColumns(id: string, database: string, table: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/columns',
    method: 'GET',
    params: { database, table }
  })
}

export function getFullTree(id: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/tree',
    method: 'GET'
  })
}

export function getTableInfo(params: { datasourceId: string; database: string; table: string }) {
  const { datasourceId, database, table } = params
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/table-info',
    method: 'GET',
    params: { database, table }
  })
}

export function getCapabilities() {
  return request({
    url: '/dataquery/capabilities',
    method: 'GET'
  })
}

export function getPrimaryKey(id: string, database: string, table: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/primary-key',
    method: 'GET',
    params: { database, table }
  })
}

export function getTableInfoFull(id: string, database: string, table: string) {
  return request({
    url: '/dataquery/datasources/' + id + '/table-info-full',
    method: 'GET',
    params: { database, table }
  })
}

export function queryTableData(params: {
  datasourceId: string
  database: string
  table: string
  page?: number
  pageSize?: number
}) {
  const { datasourceId, database, table, page = 1, pageSize = 300 } = params
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/data/query',
    method: 'GET',
    params: { database, table, page, pageSize }
  })
}

export function insertRow(params: {
  datasourceId: string
  database: string
  table: string
  row: Record<string, any>
}) {
  const { datasourceId, database, table, row } = params
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/data/insert',
    method: 'POST',
    data: { database, table, row }
  })
}

export function updateRow(params: {
  datasourceId: string
  database: string
  table: string
  row: Record<string, any>
  where: Record<string, any>
}) {
  const { datasourceId, database, table, row, where } = params
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/data/update',
    method: 'POST',
    data: { database, table, updates: row, where }
  })
}

export function deleteRow(params: {
  datasourceId: string
  database: string
  table: string
  where?: Record<string, any>
  rows?: Record<string, any>[]
}) {
  const { datasourceId, database, table, where, rows } = params
  const whereClause = where || (rows && rows.length > 0 ? rows[0] : {})
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/data/delete',
    method: 'POST',
    data: { database, table, where: whereClause }
  })
}

export function tableMaintenance(
  id: string,
  action: 'analyze' | 'check' | 'optimize' | 'repair' | 'count',
  database: string,
  table: string
) {
  return request({
    url: '/dataquery/datasources/' + id + '/maintenance/' + action,
    method: 'POST',
    data: { database, table }
  })
}

export function getTableDDL(params: { datasourceId: string; database: string; table: string }) {
  const { datasourceId, database, table } = params
  return request({
    url: '/dataquery/datasources/' + datasourceId + '/ddl',
    method: 'GET',
    params: { database, table }
  })
}

export interface SavedQueryItem {
  queryId: string
  userId: string
  username: string
  datasourceId: string
  database: string
  title: string
  description: string
  sql: string
  createdAt: string
  updatedAt: string
}

export interface SavedQueryResult {
  list: SavedQueryItem[]
  total: number
  current: number
  pageSize: number
}

export function listSavedQueries(
  datasourceId: string,
  current = 1,
  pageSize = 1000,
  keyword = ''
): Promise<SavedQueryResult> {
  return request({
    url: '/dataquery/savedQueries',
    method: 'GET',
    params: { datasourceId, page: current, pageSize, keyword }
  }) as Promise<SavedQueryResult>
}

export function getSavedQuery(id: string): Promise<SavedQueryItem> {
  return request({
    url: '/dataquery/savedQueries/' + id,
    method: 'GET'
  }) as Promise<SavedQueryItem>
}

export function saveQuery(data: {
  datasourceId: string
  database: string
  title: string
  description?: string
  sql: string
}): Promise<{ queryId: string }> {
  return request({
    url: '/dataquery/savedQueries',
    method: 'POST',
    data
  }) as Promise<{ queryId: string }>
}

export function updateSavedQuery(
  id: string,
  data: { title?: string; description?: string; sql?: string }
): Promise<{ queryId: string }> {
  return request({
    url: '/dataquery/savedQueries/' + id,
    method: 'PUT',
    data
  }) as Promise<{ queryId: string }>
}

export function deleteSavedQueryApi(id: string): Promise<{ queryId: string }> {
  return request({
    url: '/dataquery/savedQueries/' + id,
    method: 'DELETE'
  }) as Promise<{ queryId: string }>
}

export function getAuditStats() {
  return request({
    url: '/audit/stats',
    method: 'GET'
  })
}

export function getDashboardStats() {
  return request({
    url: '/dashboard/stats',
    method: 'GET'
  })
}

export function listAuditLogs(
  current: number,
  pageSize: number,
  action?: string,
  username?: string,
  keyword?: string,
  category?: string
) {
  return request({
    url: '/audit',
    method: 'GET',
    params: { current, pageSize, action, username, keyword, category }
  })
}
