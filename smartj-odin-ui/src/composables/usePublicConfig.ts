import type { SsoPublicConfig } from '@/types/sso'
import { LOGIN_THEME_KEY } from '@/utils/constants'
import { AUTH_ENABLED } from '@/config/runtime'
import { SSO_ENABLED } from '@/config/sso'

/**
 * SSO 公开配置共享获取器
 * <p>
 * 应用生命周期内共享同一个 Promise，避免重复请求。
 * 缓存 30 秒后自动过期，下次调用会重新获取最新配置。
 * <p>
 * 使用原生 fetch 绕过 axios 拦截器的 cancelAll 机制，
 * 确保在路由切换时不会被取消。
 * <p>
 * 免登录模式（VITE_AUTH_ENABLED=false）：不请求 /sso/public-config，
 * 直接返回与后端 DEFAULTS 对齐的本地默认配置（系统名称/登录主题取自构建时环境变量），
 * 内部版产物对 SSO 服务（8000 端口）无任何依赖。
 */

const CACHE_TTL_MS = 30_000

/**
 * 免登录 / 无 SSO 模式本地默认配置。
 * key/value 与后端 SsoPublicConfigService.DEFAULTS 保持一致；
 * ssoOff=true 时关闭社交/短信/注册/忘记密码等 SSO 特性。
 */
function buildLocalPublicConfig(ssoOff = false): SsoPublicConfig {
  return {
    'sso.features.sms-login-enabled': ssoOff ? 'false' : 'true',
    'sso.features.qr-code-login-enabled': 'false',
    'sso.features.social-login-enabled': 'false',
    'sso.features.registration-enabled': ssoOff ? 'false' : 'true',
    'sso.features.registration-mode': 'approval',
    'sso.features.forgot-password-enabled': ssoOff ? 'false' : 'true',
    'sso.features.forgot-password-method': 'sms',
    'sso.features.remember-me-enabled': 'true',
    'sso.features.remember-me-ttl-seconds': '1209600',
    'sso.features.mfa-enabled': 'false',
    'sso.features.mfa-global-required': 'false',
    'sso.security.max-fail-count': '5',
    'sso.security.lock-duration-seconds': '1800',
    'sso.security.password-min-length': '8',
    'sso.security.password-complexity-required': 'true',
    'sso.security.sms-code-ttl-seconds': '300',
    'auth.captcha.expire-seconds': '120',
    'crypto.sm2.enabled': 'false',
    'ui.login.theme': (import.meta.env.VITE_LOGIN_THEME as string) || 'split',
    'ui.tenant.switcher.enabled': '0',
    'sys.name': (import.meta.env.VITE_APP_TITLE as string) || 'SmartJ',
  }
}

let sharedPromise: Promise<SsoPublicConfig | null> | null = null
let cacheTime = 0

/**
 * 获取 SSO 公开配置（共享单例请求，自动去重 + 30 秒 TTL 缓存）
 * <p>
 * 30 秒内多次调用共享同一个 Promise，不会重复发请求。
 * 超过 30 秒后缓存过期，下次调用会重新请求以获取最新配置。
 * <p>
 * 免登录模式下返回本地默认配置（常量，无 TTL 概念，直接复用同一个 Promise）。
 */
export function fetchPublicConfig(): Promise<SsoPublicConfig | null> {
  // 免登录模式或无 SSO 模式：短路返回本地默认配置，不发 HTTP 请求
  if (!AUTH_ENABLED || !SSO_ENABLED) {
    if (!sharedPromise) {
      const data = buildLocalPublicConfig(!SSO_ENABLED)
      // 同步登录页主题缓存，保持与正常模式一致的持久化行为
      try {
        localStorage.setItem(LOGIN_THEME_KEY, data['ui.login.theme'])
      } catch {
        /* localStorage 不可用时静默降级 */
      }
      sharedPromise = Promise.resolve(data)
    }
    return sharedPromise
  }

  const now = Date.now()
  if (!sharedPromise || now - cacheTime >= CACHE_TTL_MS) {
    cacheTime = now
    sharedPromise = fetch('/api/sso/public-config')
      .then((res) => res.json())
      .then((json) => {
        if (json.success && json.data) {
          const data = json.data as SsoPublicConfig
          // 持久化登录页主题：供下次（含退出登录后）登录页首帧同步渲染，避免主题闪烁
          const theme = data['ui.login.theme']
          if (theme) {
            try {
              localStorage.setItem(LOGIN_THEME_KEY, theme)
            } catch {
              /* localStorage 不可用时静默降级 */
            }
          }
          return data
        }
        return null
      })
      .catch(() => null)
  }
  return sharedPromise
}
