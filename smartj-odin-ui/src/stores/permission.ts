import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getMyMenus, getMyPermissions } from '@/api/sys/menu'
import type { MenuTreeVO } from '@/types/sys'
import { AUTH_ENABLED } from '@/config/runtime'
import { STATIC_MENUS } from '@/router/staticMenus'

/** 默认首页菜单（始终显示在菜单栏最前，不受后端菜单数据影响） */
const DASHBOARD_MENU: MenuTreeVO = {
  path: '/dashboard',
  name: 'Dashboard',
  component: 'dashboard/index',
  meta: {
    title: '首页',
    icon: 'HomeFilled',
  },
  children: [],
}

/** 免登录模式下向 <code>permissions</code> 写入的通配符，约定代表“全部放行”。 */
const PERMISSION_WILDCARD = '*'

export const usePermissionStore = defineStore('permission', () => {
  const menus = ref<MenuTreeVO[]>([])
  const permissions = ref<string[]>([])
  const hasRoutes = ref(false)

  async function fetchMenus() {
    // 免登录模式：使用前端静态菜单，与 sys_menu 完全解耦
    if (!AUTH_ENABLED) {
      menus.value = [DASHBOARD_MENU, ...STATIC_MENUS]
      return menus.value
    }
    const data = await getMyMenus()
    // 确保首页始终在菜单列表最前面（如果后端未返回）
    const hasDashboard = data.some((m) => m.path === '/dashboard')
    menus.value = hasDashboard ? data : [DASHBOARD_MENU, ...data]
    return menus.value
  }

  async function fetchPermissions() {
    // 免登录模式：写入通配符，与 v-permission 指令、路由守卫约定一致
    if (!AUTH_ENABLED) {
      permissions.value = [PERMISSION_WILDCARD]
      return permissions.value
    }
    const data = await getMyPermissions()
    permissions.value = Array.isArray(data) ? data : []
    return permissions.value
  }

  function setPermissions(permList: string[]) {
    permissions.value = permList
  }

  function clearPermission() {
    menus.value = []
    permissions.value = []
    hasRoutes.value = false
  }

  return {
    menus,
    permissions,
    hasRoutes,
    fetchMenus,
    fetchPermissions,
    setPermissions,
    clearPermission,
  }
})
