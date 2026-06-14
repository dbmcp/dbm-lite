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
  createdAt: string
}

export function listServers(current: number, pageSize: number, keyword?: string, env?: string) {
  return request<{ list: Server[]; total: number }>({
    url: '/servers',
    method: 'GET',
    params: { page: current, pageSize, keyword, env }
  })
}

export function listAllServers() {
  return request<Server[]>({
    url: '/servers/all',
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

export function testServerById(id: string) {
  return request({
    url: `/servers/${id}/test`,
    method: 'POST'
  })
}

export function testServer(data: any) {
  return request({
    url: `/servers/_test_/test`,
    method: 'POST',
    data
  })
}
