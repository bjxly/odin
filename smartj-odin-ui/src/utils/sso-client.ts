/**
 * SSO 客户端辅助工具
 * <p>
 * 提供登录跳转、登出跳转、注册页跳转等 URL 构建方法。
 *
 * 用法：
 * ```ts
 * import { createSsoClient } from '@/utils/sso-client'
 * const sso = createSsoClient({ serverUrl: 'http://localhost:9000/sso', clientId: 'smartj-admin' })
 * window.location.href = sso.getLoginUrl()
 * ```
 */

export interface SsoClientConfig {
  /** SSO 服务端地址 */
  serverUrl: string
  /** 当前应用 client_id */
  clientId: string
  /** 回调地址 */
  redirectUri: string
  /** 登出回调地址 */
  logoutUri?: string
}

export function createSsoClient(config: SsoClientConfig) {
  const { serverUrl, clientId, redirectUri, logoutUri } = config

  function urlEncode(value: string): string {
    return encodeURIComponent(value)
  }

  return {
    /** 获取 SSO 登录页 URL */
    getLoginUrl(state?: string): string {
      let url = `${serverUrl}/oauth2/authorize?response_type=code&client_id=${clientId}&redirect_uri=${urlEncode(redirectUri)}&scope=openid`
      if (state) url += `&state=${urlEncode(state)}`
      return url
    },

    /** 获取 SSO 登出 URL */
    getLogoutUrl(): string {
      let url = `${serverUrl}/logout`
      if (logoutUri) url += `?post_logout_redirect_uri=${urlEncode(logoutUri)}`
      return url
    },

    /** 获取 SSO 注册页 URL */
    getRegisterUrl(): string {
      return `${serverUrl}/register`
    },

    /** 获取 SSO 忘记密码页 URL */
    getForgotPasswordUrl(): string {
      return `${serverUrl}/forgot-password`
    },

    /** 跳转到 SSO 登录页 */
    redirectToLogin(state?: string): void {
      window.location.href = this.getLoginUrl(state)
    },

    /** 跳转到 SSO 登出 */
    redirectToLogout(): void {
      window.location.href = this.getLogoutUrl()
    },

    /** 跳转到 SSO 注册页 */
    redirectToRegister(): void {
      window.location.href = this.getRegisterUrl()
    },

    /** 跳转到 SSO 忘记密码页 */
    redirectToForgotPassword(): void {
      window.location.href = this.getForgotPasswordUrl()
    },
  }
}
