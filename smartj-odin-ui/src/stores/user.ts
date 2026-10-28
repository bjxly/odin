import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getUserInfo as ssoGetUserInfo, logout as ssoLogout, switchTenant as ssoSwitchTenant } from '@/api/sso'
import { localGetUserInfo, localLogout, localSwitchTenant } from '@/api/auth'
import type { UserInfoVO } from '@/types/auth'
import { useAuthStore } from '@/stores/auth'
import { usePermissionStore } from '@/stores/permission'
import { useTagsViewStore } from '@/stores/tagsView'
import { useKeepAliveStore } from '@/stores/keepAlive'
import { SSO_PUBLIC_CONFIG_KEY } from '@/utils/constants'
import { AUTH_ENABLED } from '@/config/runtime'
import { SSO_ENABLED } from '@/config/sso'
import i18n from '@/i18n'

/** 根据 SSO 开关选择认证 API 源 */
const fetchUserInfo = SSO_ENABLED ? ssoGetUserInfo : localGetUserInfo
const apiLogout = SSO_ENABLED ? ssoLogout : localLogout
const apiSwitchTenant = SSO_ENABLED ? ssoSwitchTenant : localSwitchTenant

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfoVO | null>(null)

  // 委托给 auth store
  const authStore = useAuthStore()
  // 委托给 permission store
  const permissionStore = usePermissionStore()
  // 委托给 tagsView / keepAlive store（租户切换时清理标签与页面缓存）
  const tagsViewStore = useTagsViewStore()
  const keepAliveStore = useKeepAliveStore()

  // 向后兼容的属性代理
  const token = computed(() => authStore.token)
  const refreshToken = computed(() => authStore.refreshToken)
  const isLoggedIn = computed(() => authStore.isLoggedIn)
  const menus = computed(() => permissionStore.menus)
  const permissions = computed(() => permissionStore.permissions)
  // hasRoutes 需要可写（Pinia 返回时 computed 会解包为值类型丢失 setter，故用 ref + watch 同步）
  const hasRoutes = ref(permissionStore.hasRoutes)
  watch(hasRoutes, (v) => {
    permissionStore.hasRoutes = v
  })
  watch(
    () => permissionStore.hasRoutes,
    (v) => {
      hasRoutes.value = v
    }
  )

  function setToken(accessToken: string, newRefreshToken?: string, rememberMe?: boolean) {
    authStore.setToken(accessToken, newRefreshToken, rememberMe)
  }

  function setUserInfo(info: UserInfoVO) {
    userInfo.value = info
  }

  async function getUserInfoAction() {
    // 免登录模式：不请求 /sso/auth/userinfo，直接返回合成对象，避免无谓接口失败
    if (!AUTH_ENABLED) {
      const synthesized: UserInfoVO = {
        id: '0',
        username: 'internal',
        nickname: i18n.global.t('layout.internalUser'),
        tenantId: '1',
        roles: [],
        permissions: ['*'],
      }
      userInfo.value = synthesized
      permissionStore.setPermissions(synthesized.permissions || [])
      return synthesized
    }
    const info = await fetchUserInfo()
    userInfo.value = info
    if (info.permissions) {
      permissionStore.setPermissions(info.permissions)
    }
    return info
  }

  async function getMenus() {
    return permissionStore.fetchMenus()
  }

  async function getPermissions() {
    return permissionStore.fetchPermissions()
  }

  /**
   * 切换租户
   * <p>
   * 成功后换新 token（JWT tid 为唯一租户来源）、同步用户信息与租户名称、
   * 清理权限/标签页/页面缓存，提示成功后整页刷新重建动态路由与菜单。
   * <p>
   * 异常分层：setToken 之前的接口失败直接上抛，由组件弹“切换失败”；
   * setToken 之后的任何状态清理异常都不会上抛（此时 storage 中 token/租户已是新值），
   * 统一通过整页刷新收敛，避免页面停留在旧菜单 + 新 token 的不一致状态。
   */
  async function switchTenant(tenantId: string) {
    // 接口本身失败（setToken 之前）直接上抛，组件侧提示“切换失败”
    const res = await apiSwitchTenant({ tenantId })
    // 换新租户的 token 与刷新令牌（此后 storage 中 token/租户已是新值）
    authStore.setToken(res.token, res.refreshToken)
    try {
      // 持久化新租户 ID（与 token 相同的 remember-me 存储策略）
      authStore.setTenantId(String(res.user.tenantId))
      // 同步更新用户信息（含租户名称）
      userInfo.value = {
        id: String(res.user.id),
        username: res.user.username,
        nickname: res.user.nickname,
        tenantId: String(res.user.tenantId),
        tenantName: res.user.tenantName,
        roles: res.user.roles,
        permissions: res.user.permissions,
      }
      // 清空权限状态（hasRoutes=false，路由守卫将重新拉取菜单/权限）
      permissionStore.clearPermission()
      // 清理 TagsView 标签与 KeepAlive 页面缓存（watch 会同步清空对应 sessionStorage 持久化数据）
      tagsViewStore.removeAll()
      keepAliveStore.clearCache()
      // 切换成功提示
      ElMessage.success(i18n.global.t('layout.tenantSwitchSuccess'))
    } catch {
      // token 已换新，状态清理异常不能上抛为“切换失败”，整页刷新是唯一可靠的收敛手段（finally 中执行）
    } finally {
      // 稍延迟后整页刷新，重建动态路由与菜单，并保证成功消息可见
      setTimeout(() => {
        window.location.reload()
      }, 600)
    }
    return res
  }

  async function logout() {
    // 免登录模式下无真实会话，直接重置本地状态（避免接口失败提示干扰内部用户体验）
    if (!AUTH_ENABLED) {
      resetState()
      try {
        sessionStorage.removeItem('smartj-tagsview')
        sessionStorage.removeItem('smartj-keepalive')
        sessionStorage.removeItem(SSO_PUBLIC_CONFIG_KEY)
      } catch {
        // ignore
      }
      return
    }
    try {
      await apiLogout()
    } catch {
      // ignore
    }
    resetState()
    authStore.logout()
    // 清除 TagsView、KeepAlive、SSO 公开配置等持久化数据
    try {
      sessionStorage.removeItem('smartj-tagsview')
      sessionStorage.removeItem('smartj-keepalive')
      sessionStorage.removeItem(SSO_PUBLIC_CONFIG_KEY)
    } catch {
      // ignore
    }
  }

  function resetState() {
    userInfo.value = null
    authStore.clearAuth()
    permissionStore.clearPermission()
  }

  return {
    token,
    refreshToken,
    userInfo,
    menus,
    permissions,
    hasRoutes,
    isLoggedIn,
    setToken,
    setUserInfo,
    getUserInfo: getUserInfoAction,
    getMenus,
    getPermissions,
    switchTenant,
    logout,
    resetState,
  }
})
