import request from './request'

// ====== 资源组 ======
export interface ResourceGroup {
  groupId: string
  name: string
  remark?: string
  webhook?: string
  createdAt?: string
  updatedAt?: string
  userCount?: number
  datasourceCount?: number
}

export function listGroups(current = 1, pageSize = 20, keyword = '') {
  return request<{ list: ResourceGroup[]; total: number }>({
    url: '/priv/group',
    method: 'GET',
    params: { current, pageSize, keyword }
  })
}

export function listAllGroups() {
  return request<ResourceGroup[]>({ url: '/priv/group/all', method: 'GET' })
}

export function createGroup(data: { name: string; remark?: string }) {
  return request<ResourceGroup>({ url: '/priv/group', method: 'POST', data })
}

export function updateGroup(id: string, updates: Record<string, any>) {
  return request({ url: '/priv/group/' + id, method: 'PUT', data: updates })
}

export function deleteGroup(id: string) {
  return request({ url: '/priv/group/' + id, method: 'DELETE' })
}

export function bindGroupUsers(id: string, userIds: string[]) {
  return request({ url: '/priv/group/' + id + '/users', method: 'POST', data: { userIds } })
}

export function bindGroupDatasources(id: string, datasourceIds: string[]) {
  return request({ url: '/priv/group/' + id + '/ds', method: 'POST', data: { datasourceIds } })
}

export function getGroupUsers(id: string) {
  return request<string[]>({ url: '/priv/group/' + id + '/users', method: 'GET' })
}

export function getGroupDatasources(id: string) {
  return request<string[]>({ url: '/priv/group/' + id + '/ds', method: 'GET' })
}

// ====== 权限清单与授权 ======
export interface PrivilegeItem {
  privId: string
  userId?: string
  userName?: string
  datasourceId: string
  databaseName?: string
  tableName?: string
  privType?: string
  operationType?: string
  columns?: string
  rowLimit?: number
  expireAt?: string
  createdAt?: string
  isExpired?: boolean
}

export function listPrivileges(current = 1, pageSize = 20, params: Record<string, any> = {}) {
  return request<{ list: PrivilegeItem[]; total: number }>({
    url: '/priv/privileges',
    method: 'GET',
    params: { current, pageSize, ...params }
  })
}

export function myPrivileges(datasourceId?: string) {
  return request<{ list: PrivilegeItem[]; total: number }>({
    url: '/priv/my',
    method: 'GET',
    params: { datasourceId }
  })
}

export interface GrantRequest {
  userId: string
  userName?: string
  datasourceId: string
  databaseName?: string
  tableName?: string
  privType?: string
  operationType?: string
  columns?: string[]
  rowLimit?: number
  validDays?: number
}

export function grantPrivilege(data: GrantRequest) {
  return request({ url: '/priv/grant', method: 'POST', data })
}

export function grantPrivilegeBatch(data: {
  userIds: string[]
  datasourceId: string
  databaseName?: string
  tableName?: string
  privType?: string
  operationType?: string
  columns?: string[]
  rowLimit?: number
  validDays?: number
}) {
  return request({ url: '/priv/grant-batch', method: 'POST', data })
}

export function revokePrivilege(id: string) {
  return request({ url: '/priv/priv/' + id, method: 'DELETE' })
}

export function revokePrivilegeBatch(ids: string[]) {
  return request({ url: '/priv/revoke-batch', method: 'POST', data: { ids } })
}

// ====== 敏感列配置 ======
export interface SensitiveColumn {
  id: number
  datasourceId: string
  databaseName: string
  tableName: string
  columnName: string
  rule: string
  createdAt: string
}

export function listSensitiveColumns(datasourceId?: string) {
  return request<SensitiveColumn[]>({
    url: '/priv/sensitive-columns',
    method: 'GET',
    params: { datasourceId }
  })
}

export function createSensitiveColumn(data: Omit<SensitiveColumn, 'id' | 'createdAt'>) {
  return request({ url: '/priv/sensitive-columns', method: 'POST', data })
}

export function deleteSensitiveColumn(id: number) {
  return request({ url: '/priv/sensitive-columns/' + id, method: 'DELETE' })
}

// ====== 权限审计日志 ======
export interface PrivAuditLog {
  logId: string
  operatorId?: string
  operator?: string
  operType: string
  applyId?: string
  targetId?: string
  before?: string
  after?: string
  detail?: string
  createdAt: string
}

export function listPrivAuditLogs(current = 1, pageSize = 20, params: Record<string, any> = {}) {
  return request<{ list: PrivAuditLog[]; total: number }>({
    url: '/priv/audit',
    method: 'GET',
    params: { current, pageSize, ...params }
  })
}


