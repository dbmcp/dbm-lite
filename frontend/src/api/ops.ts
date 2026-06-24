import request from './request'

/* ========== 导入导出 ========== */
export function createImportTask(params: {
  datasourceId: string
  mode: string
  scope: string
  tables: string[]
}) {
  return request({ url: '/import-export/tasks', method: 'POST', data: params })
}
export function listImportTasks(params?: { page?: number; pageSize?: number }) {
  return request({ url: '/import-export/tasks', method: 'GET', params })
}

/* ========== SQL 审核 ========== */
export function listAuditFlows(params?: { page?: number; pageSize?: number }) {
  return request({ url: '/sql-audit/flows', method: 'GET', params })
}
export function createAuditFlow(params: { datasourceId: string; changeType: string; env: string; sql: string }) {
  return request({ url: '/sql-audit/flows', method: 'POST', data: params })
}
export function listAuditRules() {
  return request({ url: '/sql-audit/rules', method: 'GET' })
}

/* ========== 敏感数据维护 ========== */
export function listSensitiveData(params?: { keyword?: string }) {
  return request({ url: '/sensitive-data', method: 'GET', params })
}
export function createSensitiveData(params: {
  datasourceId: string
  table: string
  column: string
  dataType: string
  level: string
  maskRule: string
}) {
  return request({ url: '/sensitive-data', method: 'POST', data: params })
}
export function deleteSensitiveData(id: string) {
  return request({ url: `/sensitive-data/${id}`, method: 'DELETE' })
}

/* ========== 生命周期管理 ========== */
export function listLifecycleNodes() {
  return request({ url: '/db-lifecycle/nodes', method: 'GET' })
}
export function listLifecycleDBs() {
  return request({ url: '/db-lifecycle/dbs', method: 'GET' })
}

/* ========== 健康巡检 ========== */
export function getHealthMetrics() {
  return request({ url: '/health/metrics', method: 'GET' })
}
export function getHealthInstances() {
  return request({ url: '/health/instances', method: 'GET' })
}
export function getHealthInspectResults() {
  return request({ url: '/health/inspect', method: 'GET' })
}
export function triggerHealthInspect(params: { env: string; items: string[]; target: string }) {
  return request({ url: '/health/inspect', method: 'POST', data: params })
}

/* ========== 数据迁移 ========== */
export function listMigrationTasks() {
  return request({ url: '/migration/tasks', method: 'GET' })
}
export function createMigrationTask(params: { source: string; target: string; mode: string; tables: string }) {
  return request({ url: '/migration/tasks', method: 'POST', data: params })
}
export function getSchemaDiff() {
  return request({ url: '/migration/schema-diff', method: 'GET' })
}
export function getDataDiff() {
  return request({ url: '/migration/data-diff', method: 'GET' })
}

/* ========== 平台配置 - 介质维护 ========== */
export function listMediums() {
  return request({ url: '/platform/mediums', method: 'GET' })
}
