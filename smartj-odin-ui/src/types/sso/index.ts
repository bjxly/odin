/**
 * SSO 模块 TypeScript 类型定义
 */

/** 认证方式 */
export type AuthMethod = 'PASSWORD' | 'SMS_CODE' | 'QR_CODE'

/** MFA 类型 */
export type MfaType = 'SMS' | 'TOTP' | 'EMAIL'

/** 短信验证码用途 */
export type SmsScope = 'PRIMARY_LOGIN' | 'MFA_VERIFY' | 'BIND' | 'RESET_PWD' | 'REGISTER'

/** SSO 协议 */
export type SsoProtocol = 'OAUTH2' | 'OIDC' | 'SAML2' | 'CAS' | 'JWT'

/** IdP 类型 */
export type IdpType = 'GITHUB' | 'WECHAT' | 'WECOM' | 'DINGTALK' | 'FEISHU' | 'OAUTH2' | 'OIDC' | 'SAML2' | 'CAS'

/** 注册状态 */
export type RegistrationStatus = 0 | 1 | 2 | 3 // 0=待验证 1=待审批 2=已通过 3=已拒绝

// ==================== 公开配置 ====================

/** 已启用的 IdP 简要信息（公开字段，不含密钥） */
export interface SsoIdpPublicInfo {
  id: number
  idpName: string
  idpType: IdpType
  logoUrl?: string
}

/** SSO 公开配置，key/value 与后端 sys_config 表完全一致 */
export type SsoPublicConfig = Record<string, string>

/** SM2 公钥配置（由 /sso/auth/publicKey 返回） */
export interface SsoSm2PublicKey {
  sm2PublicKey: string
  enabled: string
}

// ==================== SSO 客户端 ====================

export interface SsoClientVO {
  id?: number
  clientId: string
  clientSecret?: string
  clientName: string
  protocol: SsoProtocol
  grantTypes: string
  redirectUris: string
  logoutUri?: string
  scopes?: string
  tokenTtl?: number
  refreshTtl: number
  mfaRequired?: number
  idpId?: number
  status?: number
}

// ==================== 短信验证码 ====================

export interface SmsSendRequest {
  phone: string
  scope: SmsScope
  captchaId?: string
  captchaCode?: string
}

export interface SmsVerifyRequest {
  phone: string
  code: string
  scope: SmsScope
}

// ==================== 用户注册 ====================

export interface SsoRegistrationVO {
  id?: number
  username: string
  password: string
  nickname: string
  mobile?: string
  email?: string
  status?: RegistrationStatus
}

// ==================== 审计日志 ====================

export interface SsoAuditLogVO {
  id: number
  eventType: string
  eventCategory: string
  eventLevel: string
  eventTime: string
  userId?: number
  username?: string
  operatorIp?: string
  action: string
  result: string
  resultMessage?: string
  clientId?: string
  protocol?: string
}

// ==================== 账号锁定 ====================

export interface SsoAccountLockVO {
  id: number
  userId: number
  username: string
  lockReason: string
  failedCount: number
  lockTime: string
  unlockTime?: string
  unlockedBy?: string
}

// ==================== 忘记密码 ====================

export interface ForgotPasswordRequest {
  phone?: string
  email?: string
  smsCode: string
  newPassword: string
}

// ==================== 登录扩展 ====================

export interface SsoLoginRequest {
  username?: string
  password?: string
  phone?: string
  smsCode?: string
  authMethod: AuthMethod
  rememberMe?: boolean
  clientId?: string
}

export interface SsoTokenVO {
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresIn: number
  scope: string
}

/** SSO 统一登录请求（与后端 SsoLoginRequest 对应） */
export interface SsoAuthLoginRequest {
  username: string
  password: string
  captchaId?: string
  captchaCode?: string
  rememberMe?: boolean
  authMethod?: string
}

/** SSO 短信登录请求 */
export interface SsoSmsLoginRequest {
  phone: string
  smsCode: string
  rememberMe?: boolean
}

/** SSO 登录响应（与后端 SsoLoginResponse 对应） */
export interface SsoAuthLoginResponse {
  token: string
  refreshToken: string
  expiresIn: number
  user: {
    id: string
    username: string
    /** 用户昵称 */
    nickname?: string
    tenantId: string
    /** 当前租户名称（切换租户等场景返回） */
    tenantName?: string
    permissions: string[]
    /** 当前租户内绑定的启用角色编码 */
    roles: string[]
  }
  /** 密码是否已过期（超过密码过期天数未修改） */
  passwordExpired?: boolean
  /** 密码距上次修改的天数（0 表示无法计算） */
  passwordDaysSinceChange?: number
}

/** SSO 验证码 */
export interface SsoCaptchaVO {
  captchaId: string
  image: string
}

// ==================== IdP 配置 ====================

export interface SsoIdpConfigVO {
  id?: number
  idpName: string
  idpType: IdpType
  protocol: string
  clientId?: string
  clientSecret?: string
  authEndpoint?: string
  tokenEndpoint?: string
  userinfoEndpoint?: string
  logoutEndpoint?: string
  metadataUrl?: string
  logoUrl?: string
  /** 首次登录策略: AUTO_CREATE/PROMPT_BIND */
  bindPolicy: string
  status?: number
  createTime?: string
  remark?: string
}

// ==================== 会话管理 ====================

export interface SsoSessionVO {
  id: string
  userId: number
  username: string
  clientId: string
  tokenId?: string
  loginIp?: string
  userAgent?: string
  loginTime: string
  lastAccess?: string
  expiresAt: string
  status: number
}

// ==================== 社交账号绑定 ====================

/** 绑定已有账号请求 */
export interface SocialBindRequest {
  tempToken: string
  username: string
  password: string
}

/** 创建新用户并绑定请求 */
export interface SocialCreateAndBindRequest {
  tempToken: string
}

/** 社交账号绑定信息 */
export interface SocialAccountInfo {
  id: number
  idpType: string
  openId: string
  nickname: string
  avatarUrl?: string
  createTime: string
}
