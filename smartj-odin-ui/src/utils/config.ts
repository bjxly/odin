import type { SsoPublicConfig } from '@/types/sso'

/**
 * 判断指定 sys_config 参数是否启用（值为 'true' 时返回 true）。
 */
export function isConfigEnabled(cfg: SsoPublicConfig | null | undefined, key: string): boolean {
  return cfg?.[key] === 'true'
}

/**
 * 获取指定 sys_config 参数字符串值，不存在时返回默认值。
 */
export function getConfigString(cfg: SsoPublicConfig | null | undefined, key: string, defaultValue = ''): string {
  return cfg?.[key] ?? defaultValue
}
