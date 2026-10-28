import { ref } from 'vue'
import { fetchPublicConfig } from './usePublicConfig'

/**
 * 密码最小长度（从 sys_config sso.security.password-min-length 读取）
 * <p>
 * 全局共享同一个 ref，跟随 fetchPublicConfig 的 30 秒 TTL 自动刷新。
 */
const passwordMinLength = ref(8)
let initialized = false

/**
 * 获取密码最小长度（从 public-config 动态读取）
 * <p>
 * 首次调用时发起请求并初始化，后续直接返回缓存值。
 * 配合 fetchPublicConfig 的 30 秒 TTL 自动刷新。
 */
export function usePasswordMinLength() {
  if (!initialized) {
    initialized = true
    fetchPublicConfig().then((cfg) => {
      if (cfg) {
        const val = parseInt(cfg['sso.security.password-min-length'] || '8', 10)
        if (val > 0) {
          passwordMinLength.value = val
        }
      }
    })
  }
  return passwordMinLength
}
