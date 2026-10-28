/**
 * 多应用 SSO 单点登录前端配置与授权码流程辅助。
 *
 * <p>仅在构建时设置 <code>VITE_SSO_ENABLED=true</code> 时启用；默认关闭，保持单体应用（本地 /login）行为不变。
 * 启用后：未登录访问受保护路由 → 整页跳转到 SSO 认证中心的 <code>/oauth2/authorize</code>；
 * SSO 回跳到本应用 <code>/sso/callback</code> 页面 → 前端把授权码交给本应用后端（BFF）代换 Token。
 *
 * <p>设计要点：
 * <ul>
 *   <li>{@link SSO_CONFIG.serverUrl} 必须是浏览器可直达的 SSO 绝对地址（如 http://localhost:8000/sso），
 *       SSO 会话 Cookie 种在该域，从而实现跨应用“一次登录、处处可用”。</li>
 *   <li>redirect_uri 固定为 <code>window.location.origin + '/sso/callback'</code>，
 *       必须与 SSO 管理端为该 client 注册的回调地址完全一致。</li>
 *   <li>授权码换 Token 走后端 BFF（client_secret 不进浏览器），前端只负责跳转与回调页。</li>
 * </ul>
 */

/** 是否启用 SSO 单点登录（默认 false；仅当 VITE_SSO_ENABLED=true 时进入 SSO 模式）。 */
export const SSO_ENABLED: boolean = import.meta.env.VITE_SSO_ENABLED === 'true'

/** SSO 认证中心配置。 */
export const SSO_CONFIG = {
  /** SSO 服务端地址（浏览器直达，含 context-path，如 http://localhost:8000/sso）。 */
  serverUrl: (import.meta.env.VITE_SSO_SERVER_URL as string) || 'http://localhost:8000/sso',
  /** 本应用在 SSO 注册的 client_id。 */
  clientId: (import.meta.env.VITE_SSO_CLIENT_ID as string) || '',
}

/** sessionStorage key：一次性 state（防 CSRF）。 */
export const SSO_STATE_KEY = 'smartj-sso-state'
/** sessionStorage key：登录成功后要跳回的业务路径。 */
export const SSO_REDIRECT_KEY = 'smartj-sso-redirect'

/** SSO 授权码回调页路径（与路由注册保持一致）。 */
export const SSO_CALLBACK_PATH = '/sso/callback'

/** 本应用的 OAuth2 redirect_uri（须与 SSO 管理端注册值一致）。 */
export function getSsoRedirectUri(): string {
  return window.location.origin + SSO_CALLBACK_PATH
}

/** 生成一次性 state，优先使用 crypto.randomUUID，降级到随机串。 */
export function generateSsoState(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return Math.random().toString(36).slice(2) + Date.now().toString(36)
}

/**
 * 构建 SSO 授权码流程登录地址（浏览器整页跳转到此 URL）。
 *
 * @param state 防 CSRF 状态参数，回调时原样返回
 */
export function getSsoLoginUrl(state: string): string {
  const params = new URLSearchParams({
    response_type: 'code',
    client_id: SSO_CONFIG.clientId,
    redirect_uri: getSsoRedirectUri(),
    scope: 'openid',
    state,
  })
  return `${SSO_CONFIG.serverUrl}/oauth2/authorize?${params.toString()}`
}

/**
 * 跳转到 SSO 登录：生成并保存 state、记录登录后要跳回的业务路径，然后整页跳转。
 *
 * @param redirectPath 登录成功后希望回到的路径（如 /sys/user）；回调页与登录页自身不记录
 */
export function redirectToSsoLogin(redirectPath?: string): void {
  const state = generateSsoState()
  try {
    sessionStorage.setItem(SSO_STATE_KEY, state)
    if (redirectPath && redirectPath !== SSO_CALLBACK_PATH && !redirectPath.startsWith('/login')) {
      sessionStorage.setItem(SSO_REDIRECT_KEY, redirectPath)
    }
  } catch {
    // sessionStorage 不可用时降级：仍可完成登录，只是丢失 state 校验与回跳路径
  }
  window.location.href = getSsoLoginUrl(state)
}

/** 读取并清除已保存的 state（回调页用于校验）。 */
export function consumeSsoState(): string | null {
  try {
    const state = sessionStorage.getItem(SSO_STATE_KEY)
    sessionStorage.removeItem(SSO_STATE_KEY)
    return state
  } catch {
    return null
  }
}

/** 读取并清除登录后要跳回的业务路径，缺省回首页。 */
export function consumeSsoRedirect(): string {
  try {
    const redirect = sessionStorage.getItem(SSO_REDIRECT_KEY)
    sessionStorage.removeItem(SSO_REDIRECT_KEY)
    return redirect || '/dashboard'
  } catch {
    return '/dashboard'
  }
}

// 首屏构建产物自检：启用了 SSO 但缺少 client_id 时给出可发现性提示，避免线上“跳转后 SSO 报 invalid_client”。
if (SSO_ENABLED && !SSO_CONFIG.clientId && typeof console !== 'undefined') {
  // eslint-disable-next-line no-console
  console.warn('[smartj] VITE_SSO_ENABLED=true 但未配置 VITE_SSO_CLIENT_ID，SSO 授权将失败，请检查 .env')
}
