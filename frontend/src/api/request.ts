import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const TOKEN_KEY = 'dbm_token'
const USER_KEY = 'dbm_user'

function getToken(): string {
  try {
    return localStorage.getItem(TOKEN_KEY) || ''
  } catch (e) {
    return ''
  }
}

function clearAuth() {
  try {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  } catch (e) {}
}

interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000
})

service.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers['Authorization'] = 'Bearer ' + token
    }
    config.headers['Content-Type'] = 'application/json'
    return config
  },
  (error) => Promise.reject(error)
)

service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const body = response.data
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) {
        return body.data
      }
      if (body.code === 40101 || body.code === 40102 || body.code === 40103) {
        clearAuth()
        ElMessage.error('登录已过期，请重新登录')
        router.push('/login')
        return Promise.reject(body)
      }
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(body)
    }
    return body as any
  },
  (error) => {
    if (error.response?.status === 401) {
      clearAuth()
      router.push('/login')
    } else {
      ElMessage.error(error.message || '网络请求失败')
    }
    return Promise.reject(error)
  }
)

export default service

export function request<T = any>(config: AxiosRequestConfig): Promise<T> {
  return service.request<any, T>(config)
}
