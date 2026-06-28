import request from './request'

export interface DatasourceInternalUser {
  id: string
  datasourceId: string
  username: string
  host: string
  isBuiltIn: boolean
  status: string
  remark: string
  createdAt: string
  createdBy: string
  updatedAt: string
}

export interface DatasourceInternalRole {
  id: string
  datasourceId: string
  name: string
  description: string
  createdAt: string
  updatedAt: string
}

export interface DatasourcePermissionRule {
  id: string
  datasourceId: string
  principalType: string
  principalId: string
  privilegeType: string
  objectLevel: string
  databaseName: string
  tableName: string
  columns: string
  enabled: boolean
  createdAt: string
  privilegeCategory: string
  systemPrivileges: string
}

export interface SystemPrivilege {
  name: string
  description: string
  category: string
}

export interface DatasourceAuthAudit {
  id: string
  operator: string
  operatorId: string
  operType: string
  operObject: string
  datasourceId: string
  clientIP: string
  result: string
  detail: string
  operTime: string
}

export function listInternalUsers(params: {
  datasourceId?: string
  keyword?: string
  status?: string
  page?: number
  pageSize?: number
}) {
  return request<{ list: DatasourceInternalUser[]; total: number; current: number; pageSize: number }>({
    url: '/datasource-internal-auth/users',
    method: 'get',
    params
  })
}

export function getInternalUser(id: string) {
  return request<DatasourceInternalUser>({
    url: `/datasource-internal-auth/users/${id}`,
    method: 'get'
  })
}

export function createInternalUser(data: {
  datasourceId: string
  username: string
  host?: string
  password?: string
  remark?: string
}) {
  return request<DatasourceInternalUser>({
    url: '/datasource-internal-auth/users',
    method: 'post',
    data
  })
}

export function updateInternalUser(id: string, data: {
  host?: string
  remark?: string
}) {
  return request({
    url: `/datasource-internal-auth/users/${id}`,
    method: 'put',
    data
  })
}

export function deleteInternalUser(id: string) {
  return request({
    url: `/datasource-internal-auth/users/${id}`,
    method: 'delete'
  })
}

export function resetUserPassword(id: string, data: { password?: string }) {
  return request({
    url: `/datasource-internal-auth/users/${id}/reset-password`,
    method: 'post',
    data
  })
}

export function toggleUserStatus(id: string, enable: boolean) {
  return request({
    url: `/datasource-internal-auth/users/${id}/status`,
    method: 'post',
    params: { enable }
  })
}

export function syncDBUsers(datasourceId: string) {
  return request({
    url: '/datasource-internal-auth/users/sync',
    method: 'post',
    params: { datasourceId }
  })
}

export function getUserRoles(userId: string) {
  return request<DatasourceInternalRole[]>({
    url: `/datasource-internal-auth/users/${userId}/roles`,
    method: 'get'
  })
}

export function assignUserRoles(userId: string, data: { roleIds: string[] }) {
  return request({
    url: `/datasource-internal-auth/users/${userId}/roles`,
    method: 'post',
    data
  })
}

export function listInternalRoles(params: {
  datasourceId?: string
  keyword?: string
  page?: number
  pageSize?: number
}) {
  return request<{ list: DatasourceInternalRole[]; total: number; current: number; pageSize: number }>({
    url: '/datasource-internal-auth/roles',
    method: 'get',
    params
  })
}

export function getInternalRole(id: string) {
  return request<DatasourceInternalRole>({
    url: `/datasource-internal-auth/roles/${id}`,
    method: 'get'
  })
}

export function createInternalRole(data: {
  datasourceId: string
  name: string
  description?: string
}) {
  return request<DatasourceInternalRole>({
    url: '/datasource-internal-auth/roles',
    method: 'post',
    data
  })
}

export function updateInternalRole(id: string, data: {
  name?: string
  description?: string
}) {
  return request({
    url: `/datasource-internal-auth/roles/${id}`,
    method: 'put',
    data
  })
}

export function deleteInternalRole(id: string) {
  return request({
    url: `/datasource-internal-auth/roles/${id}`,
    method: 'delete'
  })
}

export function getRoleUserCount(id: string) {
  return request<{ count: number }>({
    url: `/datasource-internal-auth/roles/${id}/user-count`,
    method: 'get'
  })
}

export function listPermissionRules(params: {
  datasourceId?: string
  principalType?: string
  principalId?: string
  privilegeType?: string
  objectLevel?: string
  privilegeCategory?: string
  page?: number
  pageSize?: number
}) {
  return request<{ list: DatasourcePermissionRule[]; total: number; current: number; pageSize: number }>({
    url: '/datasource-internal-auth/permissions',
    method: 'get',
    params
  })
}

export function grantPermission(data: {
  datasourceId: string
  principalType: string
  principalId: string
  privilegeType: string
  objectLevel: string
  databaseName: string
  tableName?: string
  columns?: string[]
}) {
  return request({
    url: '/datasource-internal-auth/permissions',
    method: 'post',
    data
  })
}

export function batchGrantPermission(data: {
  datasourceId: string
  principalType: string
  principalId: string
  privilegeType: string
  rules: Array<{
    objectLevel: string
    databaseName: string
    tableName?: string
    columns?: string[]
  }>
}) {
  return request({
    url: '/datasource-internal-auth/permissions/batch',
    method: 'post',
    data
  })
}

export function revokePermission(id: string) {
  return request({
    url: `/datasource-internal-auth/permissions/${id}`,
    method: 'delete'
  })
}

export function batchRevokePermission(data: { ids: string[] }) {
  return request({
    url: '/datasource-internal-auth/permissions/batch-revoke',
    method: 'post',
    data
  })
}

export function getUserPermissions(userId: string) {
  return request<DatasourcePermissionRule[]>({
    url: `/datasource-internal-auth/users/${userId}/permissions`,
    method: 'get'
  })
}

export function getRolePermissions(roleId: string) {
  return request<DatasourcePermissionRule[]>({
    url: `/datasource-internal-auth/roles/${roleId}/permissions`,
    method: 'get'
  })
}

export function getUserGrants(userId: string) {
  return request<{ grants: string }>({
    url: `/datasource-internal-auth/users/${userId}/grants`,
    method: 'get'
  })
}

export function getUserEffectiveGrants(userId: string) {
  return request<{ grants: any[] }>({
    url: `/datasource-internal-auth/users/${userId}/effective-grants`,
    method: 'get'
  })
}

export function getRoleGrants(roleId: string) {
  return request<{ grants: string }>({
    url: `/datasource-internal-auth/roles/${roleId}/grants`,
    method: 'get'
  })
}

export function checkSQLPermission(data: {
  datasourceId: string
  userId: string
  sqlText: string
}) {
  return request<{ allowed: boolean }>({
    url: '/datasource-internal-auth/permissions/check-sql',
    method: 'post',
    data
  })
}

export function listAuditLogs(params: {
  datasourceId?: string
  operator?: string
  operType?: string
  result?: string
  page?: number
  pageSize?: number
}) {
  return request<{ list: DatasourceAuthAudit[]; total: number; current: number; pageSize: number }>({
    url: '/datasource-internal-auth/audit',
    method: 'get',
    params
  })
}

export function getSystemPrivileges(datasourceId: string) {
  return request<SystemPrivilege[]>({
    url: '/datasource-internal-auth/system-privileges',
    method: 'get',
    params: { datasourceId }
  })
}

export function grantSystemPrivileges(data: {
  datasourceId: string
  principalType: string
  principalId: string
  systemPrivileges: string[]
}) {
  return request({
    url: '/datasource-internal-auth/system-privileges',
    method: 'post',
    data
  })
}

export function revokeSystemPrivileges(id: string) {
  return request({
    url: `/datasource-internal-auth/system-privileges/${id}`,
    method: 'delete'
  })
}

export function getUserSystemPrivileges(userId: string) {
  return request<string[]>({
    url: `/datasource-internal-auth/users/${userId}/system-privileges`,
    method: 'get'
  })
}

export function grantObjectPermission(data: {
  datasourceId: string
  principalType: string
  principalId: string
  objectType: string
  databaseName: string
  objectName: string
  columns?: string[]
  privileges: string[]
}) {
  return request({
    url: '/datasource-internal-auth/object-permissions',
    method: 'post',
    data
  })
}

export function revokeObjectPermission(id: string) {
  return request({
    url: `/datasource-internal-auth/object-permissions/${id}`,
    method: 'delete'
  })
}

export function getObjectPrivileges(datasourceId: string, objectType: string) {
  return request<string[]>({
    url: '/datasource-internal-auth/object-privileges',
    method: 'get',
    params: { datasourceId, objectType }
  })
}

export function listDatabases(datasourceId: string) {
  return request<string[]>({
    url: '/datasource-internal-auth/databases',
    method: 'get',
    params: { datasourceId }
  })
}

export function listObjects(datasourceId: string, dbName: string, objectType: string) {
  return request<string[]>({
    url: '/datasource-internal-auth/objects',
    method: 'get',
    params: { datasourceId, dbName, objectType }
  })
}

export function listColumns(datasourceId: string, dbName: string, tableName: string) {
  return request<string[]>({
    url: '/datasource-internal-auth/columns',
    method: 'get',
    params: { datasourceId, dbName, tableName }
  })
}

export function saveUserPermissions(data: {
  datasourceId: string
  userId: string
  objectPermissions: Array<{
    objectType: string
    privilegeType: string
    databaseName: string
    tableName: string
    columns: string[]
  }>
  systemPermissions: string[]
}) {
  return request({
    url: '/datasource-internal-auth/user/permission/save',
    method: 'post',
    data
  })
}

export function getUserPermissionDetail(params: {
  datasourceId: string
  userId: string
}) {
  return request<{
    user: DatasourceInternalUser
    objectPermissions: Array<{
      id: string
      objectType: string
      privilegeType: string
      databaseName: string
      tableName: string
      columns: string
      createdAt: string
    }>
    systemPermissions: string[]
  }>({
    url: '/datasource-internal-auth/user/permission/detail',
    method: 'get',
    params
  })
}