/**
 * 运行时功能开关与预置凭证。
 *
 * <p>统一从 <code>import.meta.env</code> 读取构建时环境变量，向业务模块暴露语义化常量，
 * 避免各处散落 <code>import.meta.env</code> 直接读取；未来若加入运行时下发开关，只需改本文件。
 */

/**
 * 是否启用登录与权限控制。
 * <p>默认 <code>true</code>；仅当构建时显式设置 <code>VITE_AUTH_ENABLED=false</code> 时进入免登录模式。
 */
export const AUTH_ENABLED: boolean = import.meta.env.VITE_AUTH_ENABLED !== 'false'

/**
 * 免登录模式下由运维预先签发的长期 JWT。
 * <p>仅在 <code>AUTH_ENABLED=false</code> 时生效；缺失时路由守卫会展示配置缺失提示。
 */
export const INTERNAL_TOKEN: string = (import.meta.env.VITE_INTERNAL_TOKEN as string) || ''

/** 语义化访问器，避免调用方直接读取常量。 */
export function isAuthEnabled(): boolean {
  return AUTH_ENABLED
}

// 首屏构建产物自检：免登录模式但没配 token 时给出可发现性提示，避免线上出现"白屏无日志"
if (!AUTH_ENABLED && !INTERNAL_TOKEN && typeof console !== 'undefined') {
  // eslint-disable-next-line no-console
  console.warn('[smartj] VITE_AUTH_ENABLED=false 但未配置 VITE_INTERNAL_TOKEN，业务接口将全部 401，请检查 .env.internal')
}
