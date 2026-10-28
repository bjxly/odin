export interface LoginRequest {
  username: string
  password: string
  captchaId: string
  captchaCode: string
  rememberMe?: boolean
}

export interface LoginResponse {
  token: string
  refreshToken?: string
  expiresIn?: number
  user: UserInfoVO
}

export interface UserInfoVO {
  id: string
  username: string
  nickname?: string
  avatar?: string
  roles?: string[]
  permissions?: string[]
  tenantId: string
  tenantName?: string
}

export interface CaptchaVO {
  captchaId: string
  image: string
}

export interface TenantOptionVO {
  id: string
  name: string
  code: string
}
