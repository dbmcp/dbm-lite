import request from './request'

export interface Business {
  businessId: string
  projectId: string
  projectName: string
  code: string
  name: string
  description: string
  env: string
  status: string
  createdAt: string
}

export function listBusinesses(projectId: string, keyword: string, current: number, pageSize: number) {
  return request<{ list: Business[]; total: number }>({
    url: projectId ? `/projects/${projectId}/businesses` : '/businesses',
    method: 'GET',
    params: { keyword, page: current, pageSize }
  })
}

export function listAllBusinesses() {
  return request<any[]>({
    url: '/businesses/all',
    method: 'GET'
  })
}

export function createBusiness(data: any) {
  return request({
    url: '/businesses',
    method: 'POST',
    data
  })
}

export function updateBusiness(businessId: string, data: any) {
  return request({
    url: `/businesses/${businessId}`,
    method: 'PUT',
    data
  })
}

export function deleteBusiness(businessId: string) {
  return request({
    url: `/businesses/${businessId}`,
    method: 'DELETE'
  })
}

export function listProjects(keyword?: string) {
  return request<{ list: any[]; total: number }>({
    url: '/projects',
    method: 'GET',
    params: { keyword }
  })
}

export function createProject(data: any) {
  return request({
    url: '/projects',
    method: 'POST',
    data
  })
}

export function updateProject(projectId: string, data: any) {
  return request({
    url: `/projects/${projectId}`,
    method: 'PUT',
    data
  })
}

export function deleteProject(projectId: string) {
  return request({
    url: `/projects/${projectId}`,
    method: 'DELETE'
  })
}
