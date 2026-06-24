import request from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  user: UserInfo
}

export interface UserInfo {
  userId: string
  username: string
  email: string
  displayName: string
  role: string
  status: string
  createdAt: string
}

export function login(params: LoginParams) {
  return request<LoginResult>({
    url: '/auth/login',
    method: 'POST',
    data: params
  })
}

export function getMe() {
  return request<UserInfo>({
    url: '/auth/me',
    method: 'GET'
  })
}

export function changePassword(oldPassword: string, newPassword: string) {
  return request({
    url: '/auth/password',
    method: 'PUT',
    data: { oldPassword, newPassword }
  })
}

export function listAccounts(current: number, pageSize: number, keyword?: string) {
  return request({
    url: '/users',
    method: 'GET',
    params: { current, pageSize, keyword }
  })
}

export function createAccount(data: any) {
  return request({
    url: '/users',
    method: 'POST',
    data
  })
}

export function updateAccount(id: string, data: any) {
  return request({
    url: '/users/' + id,
    method: 'PUT',
    data
  })
}

export function deleteAccount(id: string) {
  return request({
    url: '/users/' + id,
    method: 'DELETE'
  })
}

export function resetAccountPassword(id: string, password: string) {
  return request({
    url: '/users/' + id + '/reset-password',
    method: 'POST',
    data: { password }
  })
}
