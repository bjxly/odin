/** localStorage / sessionStorage 存储键常量 */
export const REMEMBER_KEY = 'smartj_remember_me'
export const TOKEN_KEY = 'smartj_access_token'
export const REFRESH_TOKEN_KEY = 'smartj_refresh_token'
export const SSO_PUBLIC_CONFIG_KEY = 'smartj_sso_public_config'
/** 当前租户 ID 存储键（跟随 remember-me 策略存 localStorage / sessionStorage，与 token 同策略） */
export const TENANT_ID_KEY = 'smartj_tenant_id'
/** 登录页主题缓存（localStorage，跨登出保留，用于登录页首帧同步渲染避免主题闪烁） */
export const LOGIN_THEME_KEY = 'smartj_login_theme'
