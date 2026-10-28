/**
 * 本地认证 API（无 SSO 模式使用）。
 *
 * <p>当 VITE_SSO_ENABLED=false 时，登录页及 Token 管理全部走主应用后端
 * （smartj-bootstrap，端口 9000，context-path /api），不依赖 SSO 认证中心。
 * 路径 baseURL=/api 由 request.ts 统一拼接，故此处只写 /auth/xxx。
 */
import { get, post } from '../request'
import type { CaptchaVO, UserInfoVO, TenantOptionVO } from '@/types/auth'
import type { SsoAuthLoginRequest, SsoAuthLoginResponse, SsoSm2PublicKey } from '@/types/sso'
import { encryptBySM2 } from '@/utils/crypto'

/** 获取图形验证码（主应用自带 CaptchaService） */
export function getLocalCaptcha() {
  return get<CaptchaVO>('/auth/captcha')
}

/** 获取 SM2 公钥 */
export function getLocalPublicKey() {
  return get<SsoSm2PublicKey>('/auth/publicKey')
}

/**
 * 账号密码登录 —— 自动 SM2 加密密码（与 SSO 版签名一致，返回类型兼容）。
 */
export async function localLogin(data: SsoAuthLoginRequest, sm2?: { enabled: boolean; publicKey: string }) {
  const payload: Record<string, unknown> = {
    username: data.username,
    password: data.password,
    captchaId: data.captchaId,
    captchaCode: data.captchaCode,
  }
  if (sm2?.enabled && sm2.publicKey) {
    payload.password = encryptBySM2(data.password, sm2.publicKey)
  }
  return post<SsoAuthLoginResponse>('/auth/login', payload)
}

/** 刷新 Token */
export function localRefreshToken(refreshTokenValue: string) {
  return post<SsoAuthLoginResponse>('/auth/refresh', { refreshToken: refreshTokenValue })
}

/** 获取当前登录用户信息 */
export function localGetUserInfo() {
  return get<UserInfoVO>('/auth/userinfo')
}

/** 获取当前用户可切换的租户列表 */
export function localGetUserTenants() {
  return get<TenantOptionVO[]>('/auth/tenants')
}

/** 切换租户 */
export function localSwitchTenant(data: { tenantId: string }) {
  return post<SsoAuthLoginResponse>('/auth/switch-tenant', data)
}

/** 登出 */
export function localLogout() {
  return post<void>('/auth/logout')
}
