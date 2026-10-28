import { ref } from 'vue'
import { fetchPublicConfig } from './usePublicConfig'

/**
 * 密码复杂度要求开关（从 sys_config sso.security.password-complexity-required 读取）
 * <p>
 * 为 true 时要求密码同时包含大写字母、小写字母、数字和特殊字符；为 false 时不做复杂度校验。
 * 全局共享同一个 ref，跟随 fetchPublicConfig 的 30 秒 TTL 自动刷新。
 */
const passwordComplexityRequired = ref(true)
let initialized = false

/**
 * 获取密码复杂度要求开关（从 public-config 动态读取）
 * <p>
 * 首次调用时发起请求并初始化，后续直接返回缓存值。
 * 配合 fetchPublicConfig 的 30 秒 TTL 自动刷新。
 */
export function usePasswordComplexity() {
  if (!initialized) {
    initialized = true
    fetchPublicConfig().then((cfg) => {
      if (cfg) {
        passwordComplexityRequired.value = cfg['sso.security.password-complexity-required'] === 'true'
      }
    })
  }
  return passwordComplexityRequired
}
