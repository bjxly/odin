/**
 * SSO Token 管理工具
 * <p>
 * 封装 Token 的存储、读取、刷新、清除逻辑。
 * 支持 localStorage（记住我）和 sessionStorage 双存储策略。
 */
import { REMEMBER_KEY } from '@/utils/constants'

/** SSO Token 存储键名 */
const SSO_TOKEN_KEY = 'SSO_ACCESS_TOKEN'
const SSO_REFRESH_KEY = 'SSO_REFRESH_TOKEN'

function getStorage(): Storage {
  return localStorage.getItem(REMEMBER_KEY) === '1' ? localStorage : sessionStorage
}

/** 保存 SSO Token */
export function saveSsoToken(accessToken: string, refreshToken?: string): void {
  const storage = getStorage()
  storage.setItem(SSO_TOKEN_KEY, accessToken)
  if (refreshToken) {
    storage.setItem(SSO_REFRESH_KEY, refreshToken)
  }
}

/** 获取 SSO Access Token */
export function getSsoAccessToken(): string | null {
  return localStorage.getItem(SSO_TOKEN_KEY) || sessionStorage.getItem(SSO_TOKEN_KEY)
}

/** 获取 SSO Refresh Token */
export function getSsoRefreshToken(): string | null {
  return localStorage.getItem(SSO_REFRESH_KEY) || sessionStorage.getItem(SSO_REFRESH_KEY)
}

/** 清除 SSO Token */
export function clearSsoToken(): void {
  localStorage.removeItem(SSO_TOKEN_KEY)
  localStorage.removeItem(SSO_REFRESH_KEY)
  sessionStorage.removeItem(SSO_TOKEN_KEY)
  sessionStorage.removeItem(SSO_REFRESH_KEY)
}

/** 检查 SSO Token 是否存在且未过期（本地粗略判断） */
export function hasValidSsoToken(): boolean {
  const token = getSsoAccessToken()
  if (!token) return false

  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    const exp = payload.exp * 1000
    return Date.now() < exp
  } catch {
    return false
  }
}

/** 从 SSO 回调 URL 中提取 code 参数 */
export function extractAuthCode(): string | null {
  const params = new URLSearchParams(window.location.search)
  return params.get('code')
}
