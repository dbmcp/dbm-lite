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
