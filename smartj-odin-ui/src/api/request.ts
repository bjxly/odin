import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import { useUserStore } from '@/stores/user'
import { refreshToken as refreshTokenApi } from '@/api/sso'
import { localRefreshToken } from '@/api/auth'
import { REMEMBER_KEY, TOKEN_KEY, REFRESH_TOKEN_KEY } from '@/utils/constants'
import { AUTH_ENABLED } from '@/config/runtime'
import { SSO_ENABLED } from '@/config/sso'

/** 根据 SSO 开关选择刷新 Token 的 API 源 */
const doRefreshToken = SSO_ENABLED ? refreshTokenApi : localRefreshToken

// ==================== AbortController 请求取消（4.2） ====================
const controllers = new Set<AbortController>()

function createCancelableConfig(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig {
  const controller = new AbortController()
  controllers.add(controller)

  config.signal = controller.signal
  ;(config as any)._controller = controller
  return config
}

export function cancelAll() {
  controllers.forEach((c) => c.abort())
  controllers.clear()
}

// 获取当前存储引擎
function getStorage(): Storage {
  return localStorage.getItem(REMEMBER_KEY) === '1' ? localStorage : sessionStorage
}

const request: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 优先从 localStorage 读（记住我），然后 sessionStorage
    const token = localStorage.getItem(TOKEN_KEY) || sessionStorage.getItem(TOKEN_KEY)
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return createCancelableConfig(config)
  },
  (error) => {
    return Promise.reject(error)
  }
)

// ==================== 错误消息去重节流（4.3） ====================
let lastErrorMessage = ''
let lastErrorTime = 0
const ERROR_DEDUP_INTERVAL = 3000 // 3s 内不重复弹窗

function showError(message: string) {
  const now = Date.now()
  if (message === lastErrorMessage && now - lastErrorTime < ERROR_DEDUP_INTERVAL) {
    return
  }
  lastErrorMessage = message
  lastErrorTime = now
  ElMessage.error(message)
}

// ==================== 401 自动续期 ====================
let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

async function handleTokenRefresh(originalConfig: InternalAxiosRequestConfig): Promise<any> {
  const userStore = useUserStore()
  // 优先从 localStorage 读（记住我），然后 sessionStorage
  const storedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY) || sessionStorage.getItem(REFRESH_TOKEN_KEY)

  if (!storedRefreshToken) {
    userStore.resetState()
    router.push('/login')
    return Promise.reject(new Error('无刷新令牌'))
  }

  if (isRefreshing) {
    // 等待刷新完成
    return new Promise((resolve) => {
      pendingRequests.push((newToken: string) => {
        originalConfig.headers.Authorization = `Bearer ${newToken}`
        resolve(request(originalConfig))
      })
    })
  }

  isRefreshing = true
  try {
    const res = await doRefreshToken(storedRefreshToken)
    const storage = getStorage()
    userStore.setToken(res.token, res.refreshToken)
    // 同时更新当前 storage 中的 token（setToken 已处理，这里确保 request 拦截器能读到）
    storage.setItem(TOKEN_KEY, res.token)
    if (res.refreshToken) {
      storage.setItem(REFRESH_TOKEN_KEY, res.refreshToken)
    }

    // 重试所有等待的请求
    pendingRequests.forEach((cb) => cb(res.token))
    pendingRequests = []

    // 重试当前请求
    originalConfig.headers.Authorization = `Bearer ${res.token}`
    return request(originalConfig)
  } catch (refreshError) {
    // 刷新失败，清除状态并跳转登录
    pendingRequests = []
    userStore.resetState()
    showError('登录已过期，请重新登录')
    router.push('/login')
    return Promise.reject(refreshError)
  } finally {
    isRefreshing = false
  }
}

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response
    if (data.code !== 0) {
      showError(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return data.data
  },
  (error) => {
    if (axios.isCancel(error)) return Promise.reject(error)

    const { response, config } = error
    if (response) {
      const status = response.status
      const message = response.data?.message || `请求错误: ${status}`

      if (status === 401) {
        // 免登录模式：预置 JWT 一般无 refresh，不跳 /login，避免内部产品弹出登录页；
        // 仅提示友好错误，依靠 showError 的 3s 降噪不循环弹窗
        if (!AUTH_ENABLED) {
          showError('内部服务凭证已失效，请联系运维更新 VITE_INTERNAL_TOKEN 后重新构建')
          return Promise.reject(error)
        }
        // 登录请求本身 401 不尝试刷新
        const isLoginRequest = config?.url?.includes('/auth/login')
        if (isLoginRequest) {
          showError(message || '用户名或密码错误')
          return Promise.reject(error)
        }
        // 被其他浏览器顶掉：不刷新，直接清状态并跳登录页
        if (response.data?.kicked) {
          const userStore = useUserStore()
          userStore.resetState()
          showError(response.data?.message || '您的账号已在其他设备登录，请重新登录')
          router.push('/login')
          return Promise.reject(error)
        }
        // Access Token 正常过期：走 Refresh Token 续期
        return handleTokenRefresh(config as InternalAxiosRequestConfig)
      } else if (status === 403) {
        showError('权限不足')
      } else if (status === 404) {
        showError('请求资源不存在')
      } else if (status >= 500) {
        showError('服务器内部错误')
      } else {
        showError(message)
      }
    } else if (!axios.isCancel(error)) {
      showError('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export function get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return request.get(url, config)
}

export function post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return request.post(url, data, config)
}

export function put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return request.put(url, data, config)
}

export function del<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return request.delete(url, config)
}

export function patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return request.patch(url, data, config)
}

export default request
