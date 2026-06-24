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

interface UnifiedResponse<T = any> {
  success?: boolean
  code?: number
  message?: string
  data?: T
  total?: number
  current?: number
  pageSize?: number
}

const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: parseInt(import.meta.env.VITE_API_TIMEOUT || '60000')
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
  (response: AxiosResponse<UnifiedResponse>) => {
    const body = response.data
    if (!body || typeof body !== 'object') {
      return body as any
    }

    // V2格式：success/data/message，列表接口还附带 total/current/size
    if ('success' in body) {
      if (body.success) {
        // 列表类接口：把 total/current/size 挂到返回的数组对象上，供前端分页使用
        if (Array.isArray(body.data) && (typeof body.total === 'number' || typeof body.current === 'number')) {
          const arr: any = body.data
          if (typeof body.total === 'number') (arr as any).total = body.total
          if (typeof body.current === 'number') (arr as any).current = body.current
          if (typeof body.size === 'number') (arr as any).pageSize = body.size
          return arr
        }
        return body.data as any
      }
      // 检测token相关错误：token无效或已过期
      const msg = body.message || ''
      if (msg.includes('token') || msg.includes('Token') || msg.includes('TOKEN')) {
        clearAuth()
        ElMessage.error('登录已过期，请重新登录')
        router.push('/login')
        return Promise.reject(body)
      }
      // 其他业务错误：仅以标准错误消息提示，绝不阻塞编辑器
      ElMessage({ type: 'warning', message: msg || '请求失败', duration: 3000, showClose: true })
      return Promise.reject(body)
    }

    // V1格式：code/message/data
    if ('code' in body) {
      if (body.code === 0) {
        return body.data as any
      }
      if (body.code === 40101 || body.code === 40102 || body.code === 40103) {
        clearAuth()
        ElMessage.error('登录已过期，请重新登录')
        router.push('/login')
        return Promise.reject(body)
      }
      ElMessage({ type: 'warning', message: body.message || '请求失败', duration: 3000, showClose: true })
      return Promise.reject(body)
    }

    return body as any
  },
  (error) => {
    if (error.response?.status === 401) {
      clearAuth()
      router.push('/login')
    } else {
      // 网络或 5xx 错误：以短暂 warning 形式提示，避免阻塞编辑器
      const msg = error?.response?.data?.message || error?.response?.data?.error || error.message || '网络请求失败'
      ElMessage({ type: 'warning', message: String(msg), duration: 3000, showClose: true })
    }
    return Promise.reject(error)
  }
)

export default service

export function request<T = any>(config: AxiosRequestConfig): Promise<T> {
  return service.request<any, T>(config)
}
