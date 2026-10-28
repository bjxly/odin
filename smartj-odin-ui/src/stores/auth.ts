import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import router from '@/router'
import { REMEMBER_KEY, TOKEN_KEY, REFRESH_TOKEN_KEY, TENANT_ID_KEY } from '@/utils/constants'
import { AUTH_ENABLED, INTERNAL_TOKEN } from '@/config/runtime'

function getStorage(): Storage {
  return localStorage.getItem(REMEMBER_KEY) === '1' ? localStorage : sessionStorage
}

/**
 * 免登录模式下从构建产物注入预置 JWT。
 * <p>优先取 storage（尊重上次注入结果），回退到 <code>INTERNAL_TOKEN</code>；
 * 首次注入时同步写入 localStorage，后续刷新不再重复写盘。
 */
function bootstrapToken(): string {
  // 免登录模式：以构建产物中的 INTERNAL_TOKEN 为唯一可信来源，忽略上次会话可能残留的 storage 旧 token，
  // 避免不同部署之间互干扰；写入后下面正常读取链路自动兼容。
  if (!AUTH_ENABLED && INTERNAL_TOKEN) {
    try {
      localStorage.setItem(TOKEN_KEY, INTERNAL_TOKEN)
      localStorage.setItem(REMEMBER_KEY, '1')
    } catch {
      /* localStorage 不可用时降级为内存仅本次会话使用 */
    }
    return INTERNAL_TOKEN
  }
  return getStorage().getItem(TOKEN_KEY) || localStorage.getItem(TOKEN_KEY) || ''
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(bootstrapToken())
  const refreshTokenVal = ref<string>(getStorage().getItem(REFRESH_TOKEN_KEY) || localStorage.getItem(REFRESH_TOKEN_KEY) || '')

  const isLoggedIn = computed(() => !!token.value)

  function setToken(accessToken: string, newRefreshToken?: string, rememberMe?: boolean) {
    if (rememberMe !== undefined) {
      localStorage.setItem(REMEMBER_KEY, rememberMe ? '1' : '0')
    }
    const storage = getStorage()
    token.value = accessToken
    storage.setItem(TOKEN_KEY, accessToken)
    if (newRefreshToken) {
      refreshTokenVal.value = newRefreshToken
      storage.setItem(REFRESH_TOKEN_KEY, newRefreshToken)
    }
  }

  /**
   * 持久化当前租户 ID
   * <p>
   * 与 token 相同的 remember-me 存储策略：勾选记住我存 localStorage，否则存 sessionStorage。
   */
  function setTenantId(tenantId: string) {
    getStorage().setItem(TENANT_ID_KEY, tenantId)
  }

  function clearAuth() {
    token.value = ''
    refreshTokenVal.value = ''
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(REMEMBER_KEY)
    localStorage.removeItem(TENANT_ID_KEY)
    sessionStorage.removeItem(TOKEN_KEY)
    sessionStorage.removeItem(REFRESH_TOKEN_KEY)
    sessionStorage.removeItem(TENANT_ID_KEY)
  }

  function logout() {
    clearAuth()
    router.push('/login')
  }

  return {
    token,
    refreshToken: refreshTokenVal,
    isLoggedIn,
    setToken,
    setTenantId,
    clearAuth,
    logout,
  }
})
