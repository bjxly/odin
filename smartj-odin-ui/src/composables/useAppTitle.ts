import { ref } from 'vue'
import { fetchPublicConfig } from './usePublicConfig'

// 初始值使用构建时环境变量兜底，避免接口返回前（或免登录短路前）顶栏/侧边栏标题短暂为空
const appTitle = ref((import.meta.env.VITE_APP_TITLE as string) || 'SmartJ')
let lastFetchTime = 0

/**
 * 获取系统名称（从 sys_config 的 sys.name 参数读取）。
 * 全局共享同一份数据，跟随 fetchPublicConfig 的 30 秒 TTL 自动刷新。
 */
export function useAppTitle() {
  const now = Date.now()
  // 跟随共享配置的 TTL，过期时重新获取最新值
  if (now - lastFetchTime >= 30_000 || !appTitle.value) {
    lastFetchTime = now
    fetchPublicConfig()
      .then((cfg) => {
        const name = cfg?.['sys.name']
        if (name) {
          appTitle.value = name
        }
      })
      .catch(() => {
        /* 获取失败时保留默认值 */
      })
  }
  return appTitle
}
