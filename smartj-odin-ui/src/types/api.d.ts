/**
 * AI Gateway API 类型定义
 */

// ==================== 通用 ====================

/** 分页响应 */
export interface PaginatedResponse<T> {
  records: T[]
  total: number
  current: number
  size: number
}

// ==================== 认证 ====================

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: AdminUser
}

export interface AdminUser {
  id: number
  username: string
  nickname?: string
  avatar?: string
  roles?: string[]
  lastLoginAt?: string
  createTime: string
}

// ==================== 渠道 ====================

export interface Channel {
  id: number
  name: string
  provider: string
  baseUrl: string
  apiKey: string
  models: string[]
  weight: number
  priority: number
  status: number // 0-禁用 1-启用
  timeout: number
  maxRetries: number
  rateLimit?: number
  remark?: string
  createTime: string
  updateTime: string
}

export interface ChannelTestResult {
  success: boolean
  latency: number
  message?: string
}

// ==================== API Key ====================

export interface APIKey {
  id: number
  name: string
  keyPrefix: string
  keyHash: string
  allowedModels?: string[]
  rateLimit?: number
  quotaLimit?: number
  quotaUsed: number
  status: number // 0-禁用 1-启用
  expireAt?: string
  lastUsedAt?: string
  remark?: string
  createTime: string
  updateTime: string
}

export interface KeyRotateResult {
  id: number
  keyPrefix: string
  plainKey: string
}

// ==================== 选择器规则 ====================

export interface SelectorRule {
  id: number
  name: string
  description?: string
  priority: number
  conditions: SelectorCondition[]
  targetChannels: number[]
  status: number // 0-禁用 1-启用
  createTime: string
  updateTime: string
}

export interface SelectorCondition {
  field: string
  operator: 'eq' | 'ne' | 'contains' | 'regex' | 'gt' | 'lt' | 'in'
  value: string | string[]
}

// ==================== 余额监控 ====================

export interface ProviderAccount {
  id: number
  name: string
  provider: string
  accountId: string
  apiKey: string
  balance?: number
  currency?: string
  status: number // 0-禁用 1-启用
  lastSyncAt?: string
  remark?: string
  createTime: string
  updateTime: string
}

export interface BalanceSnapshot {
  id: number
  accountId: number
  balance: number
  currency: string
  snapshotTime: string
}

export interface BalanceTrend {
  accountId: number
  accountName: string
  snapshots: BalanceSnapshot[]
}

// ==================== 用量统计 ====================

export interface UsageSummary {
  totalCalls: number
  totalTokens: number
  totalCost: number
  avgLatency: number
  successRate: number
  period: string
}

export interface UsageByUser {
  userId: string
  username: string
  totalCalls: number
  totalTokens: number
  totalCost: number
}

export interface UsageByModel {
  model: string
  provider: string
  totalCalls: number
  totalTokens: number
  totalCost: number
  avgLatency: number
}

export interface UsageByProvider {
  provider: string
  totalCalls: number
  totalTokens: number
  totalCost: number
  successRate: number
}

export interface UsageTrend {
  date: string
  calls: number
  tokens: number
  cost: number
}

export interface UsageRanking {
  rank: number
  name: string
  value: number
}

// ==================== 调用记录 ====================

export interface CallRecord {
  requestId: string
  userId: string
  username: string
  model: string
  provider: string
  channelName: string
  promptTokens: number
  completionTokens: number
  totalTokens: number
  cost: number
  latency: number
  status: number // 0-失败 1-成功
  errorMessage?: string
  requestTime: string
}

// ==================== 费用核验 ====================

export interface CostVerification {
  id: number
  provider: string
  accountId: string
  period: string
  gatewayCost: number
  providerCost: number
  difference: number
  differenceRate: number
  status: number // 0-待核验 1-核验中 2-已完成 3-异常
  remark?: string
  createTime: string
  updateTime: string
}

export interface CostVerifyReport {
  id: number
  provider: string
  period: string
  totalRecords: number
  matchedRecords: number
  unmatchedRecords: number
  gatewayTotal: number
  providerTotal: number
  difference: number
}

// ==================== 审计日志 ====================

export interface AuditLog {
  id: number
  requestId: string
  userId: string
  username: string
  action: string
  resource: string
  resourceId?: string
  method: string
  path: string
  ip: string
  userAgent?: string
  requestBody?: string
  responseBody?: string
  statusCode: number
  duration: number
  createTime: string
}

export interface AuditLogDetail extends AuditLog {
  requestHeaders?: Record<string, string>
  responseHeaders?: Record<string, string>
}
