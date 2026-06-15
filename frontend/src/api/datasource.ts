import request from './request'

export interface Datasource {
  datasourceId: string
  projectId: string
  businessId: string
  serverId: string
  name: string
  dbType: string
  host: string
  port: number
  username: string
  defaultDatabase: string
  filePath: string
  openMode: string
  charset: string
  timezone: string
  sslMode: string
  sslCaFile: string
  readOnly: boolean
  colorLabel: string
  version: string
  tags: string
  env: string
  remark: string
  status: string
  createdBy: string
  createdAt: string
  updatedAt: string
  lastConnTestAt: string
  connStatus: string
  connLatencyMs: number
}

export interface DatasourceForm {
  projectId: string
  businessId: string
  serverId: string
  autoCreateServer: boolean
  name: string
  dbType: string
  host: string
  port: number
  username: string
  password: string
  defaultDatabase: string
  filePath: string
  openMode: string
  charset: string
  timezone: string
  sslMode: string
  sslCaFile: string
  readOnly: boolean
  colorLabel: string
  tags: string
  env: string
  remark: string
  status: string
}

export interface TestResult {
  success: boolean
  message: string
  latencyMs?: number
  version?: string
}

export function listDatasources(current: number, pageSize: number, keyword?: string, dbType?: string, status?: string, sortBy?: string) {
  return request({
    url: '/datasources',
    method: 'GET',
    params: { current, pageSize, keyword, dbType, status, sortBy }
  })
}

export function listAllDatasources() {
  return request<Datasource[]>({
    url: '/datasources/all',
    method: 'GET'
  })
}

export function getDatasource(id: string) {
  return request<Datasource>({
    url: '/datasources/' + id,
    method: 'GET'
  })
}

export function getDatasourceDetail(id: string) {
  return request({
    url: '/datasources/' + id + '/detail',
    method: 'GET'
  })
}

export function createDatasource(data: DatasourceForm) {
  return request<Datasource>({
    url: '/datasources',
    method: 'POST',
    data
  })
}

export function updateDatasource(id: string, data: DatasourceForm) {
  return request({
    url: '/datasources/' + id,
    method: 'PUT',
    data
  })
}

export function deleteDatasource(id: string) {
  return request({
    url: '/datasources/' + id,
    method: 'DELETE'
  })
}

export function copyDatasource(id: string) {
  return request<Datasource>({
    url: '/datasources/' + id + '/copy',
    method: 'POST'
  })
}

export function testConnection(data: DatasourceForm) {
  return request<TestResult>({
    url: '/datasources/testConnection',
    method: 'POST',
    data
  })
}

export function testConnectionById(id: string) {
  return request<TestResult>({
    url: '/datasources/' + id + '/test',
    method: 'POST'
  })
}

export function listDatabases(id: string) {
  return request<string[]>({
    url: '/datasources/' + id + '/databases',
    method: 'GET'
  })
}

export function listTables(id: string, database: string) {
  return request<string[]>({
    url: '/datasources/' + id + '/tables',
    method: 'GET',
    params: { database }
  })
}

export function getFullTree(id: string) {
  return request({
    url: '/datasources/' + id + '/tree',
    method: 'GET'
  })
}

// V2 API 接口

export interface DatasourceMatrixItem {
  id: string
  name: string
  dbType: string
  type: string
  env: string
  status: string
  connectStatus: string
  description: string
  tags: string[]
  createTime: string
  iconType: string
}

export interface EnvGroup {
  env: string
  envName: string
  datasourceList: DatasourceMatrixItem[]
}

export interface DatasourceListItem {
  id: string
  name: string
  datasourceType: string
  type: string
  dbType: string
  env: string
  host: string
  port: number | null
  username: string
  password: string
  databaseName: string
  description: string
  status: string
  connectStatus: string
  createTime: string
  updateTime: string
  tags: string[]
}

export interface DatasourceDetail {
  id: string
  name: string
  datasourceType: string
  type: string
  dbType: string
  env: string
  host: string
  port: number | null
  username: string
  password: string
  databaseName: string
  description: string
  status: string
  connectStatus: string
  createTime: string
  updateTime: string
  tags: string[]
  ownerId: string
  orgId: string
}

export interface ListResult {
  total: number
  current: number
  pageSize: number
  list: DatasourceListItem[]
}

export interface CreateDatasourceReq {
  name: string
  datasourceType?: string
  type?: string
  dbType: string
  env: string
  host: string
  port?: number
  username?: string
  password?: string
  databaseName?: string
  description?: string
  tags?: string[]
}

export interface UpdateDatasourceReq {
  name?: string
  env?: string
  host?: string
  port?: number
  username?: string
  password?: string
  databaseName?: string
  description?: string
  tags?: string[]
}

export interface TestConnectionReq {
  dbType: string
  host: string
  port?: number
  username?: string
  password?: string
  databaseName?: string
  filePath?: string
}

export interface TestConnectionResult {
  success: boolean
  message: string
  version: string
  cost: number
  status: string
}

export function getMatrix() {
  return request<EnvGroup[]>({
    url: '/datasource/matrix',
    method: 'GET'
  })
}

export function listDatasource(keyword?: string, dbType?: string, current = 1, pageSize = 10) {
  return request<ListResult>({
    url: '/datasource/listDatasource',
    method: 'GET',
    params: { keyword, type: dbType, current, pageSize }
  })
}

export function getDatasourceInfo(id: string) {
  return request<DatasourceDetail>({
    url: '/datasource/' + id + '/datasourceInfo',
    method: 'GET'
  })
}

export function createDatasourceV2(data: CreateDatasourceReq) {
  return request<DatasourceDetail>({
    url: '/datasource/createDatasource',
    method: 'POST',
    data
  })
}

export function updateDatasourceV2(id: string, data: UpdateDatasourceReq) {
  return request<DatasourceDetail>({
    url: '/datasource/' + id + '/updateDatasource',
    method: 'POST',
    data
  })
}

export function deleteDatasourceV2(id: string) {
  return request({
    url: '/datasource/' + id + '/deleteDatasource',
    method: 'POST'
  })
}

export function testConnectionV2(data: TestConnectionReq) {
  return request<TestConnectionResult>({
    url: '/datasource/testConnection',
    method: 'POST',
    data
  })
}

export function listRecentlyDatasource(limit = 8) {
  return request<DatasourceListItem[]>({
    url: '/datasource/listRecentlyDatasource',
    method: 'GET',
    params: { limit }
  })
}
