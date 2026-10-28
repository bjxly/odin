import { useUserStore } from '@/stores/user'
import { AUTH_ENABLED } from '@/config/runtime'
import type { Directive } from 'vue'

export const permission: Directive = {
  mounted(el: HTMLElement, binding) {
    // 免登录模式：所有按钮权限短路放行，不移除节点
    if (!AUTH_ENABLED) return

    const userStore = useUserStore()
    const required = binding.value as string | string[]
    const perms = userStore.permissions

    // 通配符约定：'*' 代表全部放行（免登录模式下 fetchPermissions 会写入）
    if (perms.includes('*')) return

    const hasPermission = Array.isArray(required) ? required.some((p) => perms.includes(p)) : perms.includes(required)

    if (!hasPermission) {
      el.parentNode?.removeChild(el)
    }
  },
}
