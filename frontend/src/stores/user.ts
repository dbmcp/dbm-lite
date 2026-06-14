import { defineStore } from 'pinia'
import { login as apiLogin, UserInfo, LoginParams } from '@/api/auth'

function safeParseJSON<T>(raw: string | null, fallback: T): T {
  if (!raw) return fallback
  if (raw === 'undefined' || raw === 'null') {
    localStorage.removeItem('dbm_user')
    return fallback
  }
  try {
    return JSON.parse(raw) as T
  } catch (e) {
    localStorage.removeItem('dbm_user')
    return fallback
  }
}

interface UserState {
  token: string
  userInfo: UserInfo | null
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: localStorage.getItem('dbm_token') || '',
    userInfo: safeParseJSON<UserInfo | null>(localStorage.getItem('dbm_user'), null)
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.userInfo?.role === 'admin',
    displayName: (state) => state.userInfo?.displayName || state.userInfo?.username || 'Guest'
  },
  actions: {
    async login(params: LoginParams) {
      const res = await apiLogin(params)
      this.token = res?.token || ''
      this.userInfo = res?.user || null
      if (this.token) {
        localStorage.setItem('dbm_token', this.token)
      } else {
        localStorage.removeItem('dbm_token')
      }
      if (this.userInfo) {
        localStorage.setItem('dbm_user', JSON.stringify(this.userInfo))
      } else {
        localStorage.removeItem('dbm_user')
      }
      return res
    },
    setUserInfo(user: UserInfo) {
      this.userInfo = user
      localStorage.setItem('dbm_user', JSON.stringify(user))
    },
    logout() {
      this.token = ''
      this.userInfo = null
      localStorage.removeItem('dbm_token')
      localStorage.removeItem('dbm_user')
    }
  }
})
