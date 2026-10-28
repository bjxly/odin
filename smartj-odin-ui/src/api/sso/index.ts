/**
 * SSO 模块 API 函数
 */
import { get, post, put, del } from '../request'
import { SSO_PUBLIC_CONFIG_KEY } from '@/utils/constants'
import { SSO_ENABLED } from '@/config/sso'
import type {
  SsoPublicConfig,
  SsoSm2PublicKey,
  SsoIdpPublicInfo,
  SsoClientVO,
  SmsSendRequest,
  SmsVerifyRequest,
  SsoRegistrationVO,
  SsoAuditLogVO,
  SsoAccountLockVO,
  ForgotPasswordRequest,
  SsoLoginRequest,
  SsoTokenVO,
  SsoIdpConfigVO,
  SsoSessionVO,
  SsoAuthLoginRequest,
  SsoSmsLoginRequest,
  SsoAuthLoginResponse,
  SsoCaptchaVO,
  SocialBindRequest,
  SocialCreateAndBindRequest,
  SocialAccountInfo,
} from '@/types/sso'
import type { UserInfoVO, TenantOptionVO } from '@/types/auth'
import { encryptBySM2 } from '@/utils/crypto'

// ==================== 公开配置 ====================

/**
 * 获取 SSO 公开配置
 * <p>
 * 优先从 sessionStorage 读取缓存，不存在时从 API 获取并写入缓存。
 * 登出时会自动清除缓存。
 * <p>
 * 无 SSO 模式（VITE_SSO_ENABLED=false）：直接返回本地默认配置，不发起 HTTP 请求。
 */
export async function getSsoPublicConfig(): Promise<SsoPublicConfig> {
  // 无 SSO 模式：短路返回本地默认，避免请求不存在的 SSO 端点
  if (!SSO_ENABLED) {
    return {
      'sso.features.sms-login-enabled': 'false',
      'sso.features.qr-code-login-enabled': 'false',
      'sso.features.social-login-enabled': 'false',
      'sso.features.registration-enabled': 'false',
      'sso.features.forgot-password-enabled': 'false',
      'sso.features.remember-me-enabled': 'true',
      'sso.features.mfa-enabled': 'false',
      'ui.tenant.switcher.enabled': '0',
      'ui.login.theme': (import.meta.env.VITE_LOGIN_THEME as string) || 'split',
      'sys.name': (import.meta.env.VITE_APP_TITLE as string) || 'SmartJ',
    } as SsoPublicConfig
  }
  const cached = sessionStorage.getItem(SSO_PUBLIC_CONFIG_KEY)
  if (cached) {
    try {
      return JSON.parse(cached) as SsoPublicConfig
    } catch {
      sessionStorage.removeItem(SSO_PUBLIC_CONFIG_KEY)
    }
  }
  const data = await get<SsoPublicConfig>('/sso/public-config')
  try {
    sessionStorage.setItem(SSO_PUBLIC_CONFIG_KEY, JSON.stringify(data))
  } catch {
    // sessionStorage 不可用时静默降级
  }
  return data
}

/** 获取 SM2 公钥配置 */
export function getSsoSm2PublicKey() {
  return get<SsoSm2PublicKey>('/sso/auth/publicKey')
}

/** 获取公开 IdP 列表（登录页展示用） */
export function getSsoPublicIdps() {
  return get<SsoIdpPublicInfo[]>('/sso/idp/public')
}

// ==================== 短信验证码 ====================

export function sendSmsCode(data: SmsSendRequest) {
  return post<string>('/sso/sms/send', null, {
    params: data,
  })
}

export function verifySmsCode(data: SmsVerifyRequest) {
  return post<boolean>('/sso/sms/verify', null, { params: data })
}

// ==================== 用户注册 ====================

export function submitRegistration(data: SsoRegistrationVO) {
  return post<{ id: number; status: number }>('/sso/register', data)
}

// ==================== 注册审核 ====================

export interface RegistrationReviewVO {
  id: number
  username: string
  nickname: string
  mobile: string | null
  email: string | null
  status: number
  approvedBy: string | null
  approvedAt: string | null
  rejectReason: string | null
  createTime: string
}

/** 注册申请列表（分页） */
export function getRegistrationList(params: Record<string, any>) {
  return get<{ records: RegistrationReviewVO[]; totalRow: number }>('/sso/registration', { params })
}

/** 审批通过（指定目标租户） */
export function approveRegistration(id: number, tenantId: number) {
  return put<void>(`/sso/registration/${id}/approve`, null, { params: { tenantId } })
}

/** 获取活跃租户列表（审批弹窗用） */
export function getRegistrationTenants() {
  return get<Array<{ id: number; name: string; code: string }>>('/sso/registration/tenants')
}

/** 审批拒绝 */
export function rejectRegistration(id: number, reason: string) {
  return put<void>(`/sso/registration/${id}/reject`, null, { params: { reason } })
}

// ==================== 忘记密码 ====================

export function forgotPassword(data: ForgotPasswordRequest) {
  return post<void>('/sso/forgot-password', data)
}

// ==================== SSO 统一认证（新增） ====================

/** 获取图形验证码 */
export function getSsoCaptcha() {
  return get<SsoCaptchaVO>('/sso/auth/captcha')
}

/**
 * 账号密码登录 — 自动 SM2 加密密码。
 *
 * <p>SM2 参数从 {@link getSsoSm2PublicKey} 获取（由调用方在登录页挂载时预取），
 * 避免每次登录额外请求公钥端点。
 *
 * @param data 登录请求体
 * @param sm2   SM2 加密配置，传入 {@code { enabled: true, publicKey: "04..." }}
 *              以启用国密加密；不传或 {@code enabled=false} 则明文传输
 */
export async function ssoPasswordLogin(data: SsoAuthLoginRequest, sm2?: { enabled: boolean; publicKey: string }) {
  if (sm2?.enabled && sm2.publicKey) {
    const encryptedPwd = encryptBySM2(data.password, sm2.publicKey)
    return post<SsoAuthLoginResponse>('/sso/auth/login', {
      ...data,
      password: encryptedPwd,
    })
  }
  return post<SsoAuthLoginResponse>('/sso/auth/login', data)
}

/** 短信验证码登录 */
export function ssoSmsLogin(data: SsoSmsLoginRequest) {
  return post<SsoAuthLoginResponse>('/sso/auth/sms-login', data)
}

// ==================== SSO 登录（扩展） ====================

export function ssoLogin(data: SsoLoginRequest) {
  return post<SsoTokenVO>('/sso/oauth2/token', {
    grant_type: 'sso_login',
    ...data,
  })
}

// ==================== SSO 客户端管理 ====================

export function getSsoClients(params?: { page?: number; size?: number; clientName?: string }) {
  return get<{ records: SsoClientVO[]; totalRow: number }>('/sso/clients', { params })
}

export function getSsoClient(id: number) {
  return get<SsoClientVO>(`/sso/clients/${id}`)
}

export function createSsoClient(data: SsoClientVO) {
  return post<number>('/sso/clients', data)
}

export function updateSsoClient(id: number, data: SsoClientVO) {
  return post<void>(`/sso/clients/${id}`, data)
}

export function deleteSsoClient(id: number) {
  return del<void>(`/sso/clients/${id}/delete`)
}

// ==================== 审计日志 ====================

export function getSsoAuditLogs(params: { page?: number; size?: number; userName?: string; eventType?: string; start?: string; end?: string }) {
  return get<{ records: SsoAuditLogVO[]; totalRow: number }>('/sso/audit', { params })
}

// ==================== 安全策略 ====================

export function getSsoAccountLocks(params?: { page?: number; size?: number; userName?: string; lockReason?: string }) {
  return get<{ records: SsoAccountLockVO[]; totalRow: number }>('/sso/security/locks', { params })
}

export function unlockSsoAccount(lockId: number, operatedBy: string) {
  return post<void>(`/sso/security/locks/${lockId}/unlock`, null, { params: { operatedBy } })
}

// ==================== IdP 配置管理 ====================

export function getSsoIdps(params?: { page?: number; size?: number; idpName?: string; idpType?: string }) {
  return get<{ records: SsoIdpConfigVO[]; totalRow: number }>('/sso/idp', { params })
}

export function getSsoIdp(id: number) {
  return get<SsoIdpConfigVO>(`/sso/idp/${id}`)
}

export function createSsoIdp(data: SsoIdpConfigVO) {
  return post<number>('/sso/idp', data)
}

export function updateSsoIdp(id: number, data: SsoIdpConfigVO) {
  return post<void>(`/sso/idp/${id}`, data)
}

export function deleteSsoIdp(id: number) {
  return del<void>(`/sso/idp/${id}/delete`)
}

// ==================== 会话管理 ====================

export function getSsoSessions(params?: { page?: number; size?: number; username?: string; clientId?: string }) {
  return get<{ records: SsoSessionVO[]; totalRow: number }>('/sso/session', { params })
}

export function kickSsoSession(sessionId: string) {
  return post<void>(`/sso/session/${sessionId}/kick`)
}

// ==================== 社交登录 ====================

/** 社交登录一次性授权码换 Token */
export function exchangeSocialCode(code: string) {
  return post<SsoAuthLoginResponse>('/sso/social/exchange', null, { params: { code } })
}

// ==================== 社交账号绑定 ====================

/** 绑定已有账号 */
export function bindSocialAccount(data: SocialBindRequest) {
  return post<SsoAuthLoginResponse>('/sso/social/bind', data)
}

/** 创建新用户并绑定 */
export function createAndBindSocial(data: SocialCreateAndBindRequest) {
  return post<SsoAuthLoginResponse>('/sso/social/create-and-bind', data)
}

/** 解绑社交账号 */
export function unbindSocialAccount(id: number) {
  return del<void>(`/sso/social/unbind/${id}`)
}

/** 获取用户的社交账号绑定列表 */
export function getSocialAccounts(userId: string | number) {
  return get<SocialAccountInfo[]>('/sso/social/accounts', { params: { userId } })
}

// ==================== Token 刷新 / 用户信息 / 租户切换 ====================

/** 刷新令牌 */
export function refreshToken(refreshTokenValue: string) {
  return post<SsoAuthLoginResponse>('/sso/auth/refresh', { refreshToken: refreshTokenValue })
}

/** 获取当前用户信息 */
export function getUserInfo() {
  return get<UserInfoVO>('/sso/auth/userinfo')
}

/** 获取当前用户可切换的租户列表（SSO 版） */
export function getUserTenants() {
  if (!SSO_ENABLED) return Promise.resolve([])
  return get<TenantOptionVO[]>('/sso/auth/tenants')
}

/** 切换租户 */
export function switchTenant(data: { tenantId: string }) {
  return post<SsoAuthLoginResponse>('/sso/auth/switch-tenant', data)
}

/** 登出 */
export function logout() {
  return post<void>('/sso/auth/logout')
}
