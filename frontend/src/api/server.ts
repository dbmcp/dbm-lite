import request from '@/api/request'

export interface Server {
  serverId: string
  projectId: string
  businessId: string
  env: string
  name: string
  host: string
  port: number
  username: string
  authType: string
  status: string
  connStatus: string
  connLatencyMs: number
  os: string
  arch: string
  version: string
  cpuCores: number
  memoryGB: number
  diskGB: number
  remark: string
  tags: string
  timeout: number
  createdAt: string
  lastCheckTime: string
  createdBy: string
}

export function listServers(current: number, pageSize: number, keyword?: string, env?: string, status?: string) {
  return request<{ list: Server[]; total: number }>({
    url: '/servers',
    method: 'GET',
    params: { current, pageSize, keyword, env, status }
  })
}

export function listAllServers() {
  return request<Server[]>({
    url: '/servers/all',
    method: 'GET'
  })
}

export function statsServers() {
  return request<Record<string, any>>({
    url: '/servers/stats',
    method: 'GET'
  })
}

export function getServer(id: string) {
  return request<Server>({
    url: `/servers/${id}`,
    method: 'GET'
  })
}

export function createServer(data: any) {
  return request({
    url: '/servers',
    method: 'POST',
    data
  })
}

export function updateServer(id: string, data: any) {
  return request({
    url: `/servers/${id}`,
    method: 'PUT',
    data
  })
}

export function deleteServer(id: string) {
  return request({
    url: `/servers/${id}`,
    method: 'DELETE'
  })
}

export function toggleServer(id: string) {
  return request({
    url: `/servers/${id}/toggle`,
    method: 'POST'
  })
}

export function testServerById(id: string) {
  return request({
    url: `/servers/${id}/test`,
    method: 'POST'
  })
}

export function testServer(data: any) {
  return request({
    url: '/servers/test',
    method: 'POST',
    data
  })
}

export function execCommand(id: string, command: string) {
  return request<{ stdout: string; stderr: string }>({
    url: `/servers/${id}/exec`,
    method: 'POST',
    data: { command }
  })
}
