import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { cancelAll } from '@/api/request'
import { useUserStore } from '@/stores/user'
import { useKeepAliveStore } from '@/stores/keepAlive'
import { useTagsViewStore } from '@/stores/tagsView'
import type { MenuTreeVO } from '@/types/sys'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import { AUTH_ENABLED } from '@/config/runtime'

NProgress.configure({ showSpinner: false })

// 布局组件
const MainLayout = () => import('@/layouts/MainLayout.vue')

// 视图组件映射 (后端返回 component 字段对应的懒加载)
const viewModules = import.meta.glob('../views/**/*.vue')

function resolveComponent(component: string) {
  // component 可能是 "sys/user/index" 或 "dashboard/index"
  const path = `../views/${component}.vue`
  if (viewModules[path]) {
    return viewModules[path]
  }
  // 尝试加 /index
  const indexPath = `../views/${component}/index.vue`
  if (viewModules[indexPath]) {
    return viewModules[indexPath]
  }
  // 兜底 404
  return () => import('@/views/error/404.vue')
}

// 静态路由
export const constantRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/login.vue'),
    meta: { hidden: true, title: '登录' },
  },
  {
    path: '/sso/login',
    redirect: '/login',
  },
  {
    path: '/sso/register',
    name: 'SsoRegister',
    component: () => import('@/views/sso/register/index.vue'),
    meta: { hidden: true, title: '统一注册' },
  },
  {
    path: '/sso/forgot-password',
    name: 'SsoForgot',
    component: () => import('@/views/sso/forgot-password/index.vue'),
    meta: { hidden: true, title: '统一忘记密码' },
  },
  {
    path: '/sso/consent',
    name: 'SsoConsent',
    component: () => import('@/views/sso/consent/index.vue'),
    meta: { hidden: true, title: '统一授权确认' },
  },
  {
    path: '/sso/social-callback',
    name: 'SsoSocialCallback',
    component: () => import('@/views/sso/social-callback/index.vue'),
    meta: { hidden: true, title: '社交登录回调' },
  },
  {
    path: '/sso/social-bind',
    name: 'SsoSocialBind',
    component: () => import('@/views/sso/social-bind/index.vue'),
    meta: { hidden: true, title: '社交账号绑定' },
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { hidden: true, title: '404' },
  },
]

// 需要认证但不依赖动态菜单的路由
export const staticAuthRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'MainLayout',
    component: MainLayout,
    redirect: '/dashboard',
    children: [
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '首页', icon: 'HomeFilled', affix: true },
      },
      {
        path: '/examples/svg-icon',
        name: 'ExampleSvgIcon',
        component: () => import('@/views/examples/example-svg-icon.vue'),
        meta: { title: 'SmSvgIcon 示例' },
      },
      {
        path: '/examples/dict-tag',
        name: 'ExampleDictTag',
        component: () => import('@/views/examples/example-dict-tag.vue'),
        meta: { title: 'SmDictTag 示例' },
      },
      {
        path: '/examples/search-form',
        name: 'ExampleSearchForm',
        component: () => import('@/views/examples/example-search-form.vue'),
        meta: { title: 'SmSearchForm 示例' },
      },
      {
        path: '/examples/crud-table',
        name: 'ExampleCrudTable',
        component: () => import('@/views/examples/example-crud-table.vue'),
        meta: { title: 'SmCrudTable 示例' },
      },
      {
        path: '/examples/crontab',
        name: 'ExampleCrontab',
        component: () => import('@/views/examples/example-crontab.vue'),
        meta: { title: 'SmCrontab 示例' },
      },
      {
        path: '/examples/tree-select',
        name: 'ExampleTreeSelect',
        component: () => import('@/views/examples/example-tree-select.vue'),
        meta: { title: 'SmTreeSelect 示例' },
      },
      {
        path: '/examples/file-upload',
        name: 'ExampleFileUpload',
        component: () => import('@/views/examples/example-file-upload.vue'),
        meta: { title: 'SmFileUpload 示例' },
      },
      {
        path: '/examples/avatar-crop',
        name: 'ExampleAvatarCrop',
        component: () => import('@/views/examples/example-avatar-crop.vue'),
        meta: { title: 'SmAvatarCrop 示例' },
      },
      {
        path: '/examples/color-picker',
        name: 'ExampleColorPicker',
        component: () => import('@/views/examples/example-color-picker.vue'),
        meta: { title: 'SmColorPicker 示例' },
      },
      {
        path: '/examples/json-editor',
        name: 'ExampleJsonEditor',
        component: () => import('@/views/examples/example-json-editor.vue'),
        meta: { title: 'SmJsonEditor 示例' },
      },
      {
        path: '/examples/video-player',
        name: 'ExampleVideoPlayer',
        component: () => import('@/views/examples/example-video-player.vue'),
        meta: { title: 'SmVideoPlayer 示例' },
      },
      {
        path: '/examples/rich-text',
        name: 'ExampleRichText',
        component: () => import('@/views/examples/example-rich-text.vue'),
        meta: { title: 'SmRichText 示例' },
      },
      {
        path: '/examples/code-editor',
        name: 'ExampleCodeEditor',
        component: () => import('@/views/examples/example-code-editor.vue'),
        meta: { title: 'SmCodeEditor 示例' },
      },
      {
        path: '/examples/icon-select',
        name: 'ExampleIconSelect',
        component: () => import('@/views/examples/example-icon-select.vue'),
        meta: { title: 'SmIconSelect 示例' },
      },
      {
        path: '/examples/image-preview',
        name: 'ExampleImagePreview',
        component: () => import('@/views/examples/example-image-preview.vue'),
        meta: { title: 'SmImagePreview 示例' },
      },
      {
        path: '/examples/context-menu',
        name: 'ExampleContextMenu',
        component: () => import('@/views/examples/example-context-menu.vue'),
        meta: { title: 'SmContextMenu 示例' },
      },
      {
        path: '/examples/steps-form',
        name: 'ExampleStepsForm',
        component: () => import('@/views/examples/example-steps-form.vue'),
        meta: { title: 'SmStepsForm 示例' },
      },
      {
        path: '/examples/drawer-form',
        name: 'ExampleDrawerForm',
        component: () => import('@/views/examples/example-drawer-form.vue'),
        meta: { title: 'SmDrawerForm 示例' },
      },
      {
        path: '/examples/count-to',
        name: 'ExampleCountTo',
        component: () => import('@/views/examples/example-count-to.vue'),
        meta: { title: 'SmCountTo 示例' },
      },
      {
        path: '/examples/scrollbar',
        name: 'ExampleScrollbar',
        component: () => import('@/views/examples/example-scrollbar.vue'),
        meta: { title: 'SmScrollbar 示例' },
      },
      {
        path: '/examples/empty',
        name: 'ExampleEmpty',
        component: () => import('@/views/examples/example-empty.vue'),
        meta: { title: 'SmEmpty 示例' },
      },
      {
        path: '/examples/watermark',
        name: 'ExampleWatermark',
        component: () => import('@/views/examples/example-watermark.vue'),
        meta: { title: 'SmWatermark 示例' },
      },
      {
        path: '/examples/rate',
        name: 'ExampleRate',
        component: () => import('@/views/examples/example-rate.vue'),
        meta: { title: 'SmRate 示例' },
      },
      {
        path: '/examples/calendar',
        name: 'ExampleCalendar',
        component: () => import('@/views/examples/example-calendar.vue'),
        meta: { title: 'SmCalendar 示例' },
      },
      {
        path: '/examples/qr-code',
        name: 'ExampleQRCode',
        component: () => import('@/views/examples/example-qr-code.vue'),
        meta: { title: 'SmQRCode 示例' },
      },
      {
        path: '/examples/input-otp',
        name: 'ExampleInputOtp',
        component: () => import('@/views/examples/example-input-otp.vue'),
        meta: { title: 'SmInputOtp 示例' },
      },
      {
        path: '/examples/map-china',
        name: 'ExampleMapChina',
        component: () => import('@/views/examples/map/map-china.vue'),
        meta: { title: '中国数据大屏' },
      },
      {
        path: '/examples/map-globe',
        name: 'ExampleMapGlobe',
        component: () => import('@/views/examples/map/map-globe.vue'),
        meta: { title: '全球数据大屏' },
      },
      {
        path: '/examples/wms',
        name: 'ExampleWms',
        component: () => import('@/views/examples/wms/index.vue'),
        meta: { title: '岩心仓储 3D 管理' },
      },
      {
        path: '/profile',
        name: 'Profile',
        component: () => import('@/views/sys/profile/index.vue'),
        meta: { title: '个人中心', icon: 'UserFilled', hidden: true },
      },
    ],
  },
]

// 将后端菜单树转为 vue-router 格式（方案 A：提取所有叶子路由，打平注册到 MainLayout 下）
export function generateRoutes(menus: MenuTreeVO[]): RouteRecordRaw[] {
  const routes: RouteRecordRaw[] = []

  function extractLeafRoutes(items: MenuTreeVO[]) {
    for (const menu of items) {
      if (!menu.path) continue

      // 判断 children 中是否有真正的路由子节点（有 path 的才算）
      const routeChildren = menu.children?.filter((c) => c.path) || []

      if (menu.component) {
        // 有 component 的节点：注册为页面路由
        // （即使有 children，children 可能都是按钮权限，无 path）
        routes.push({
          path: menu.path,
          name: menu.name || menu.path.replace(/\//g, '-') || `route-${Date.now()}`,
          component: resolveComponent(menu.component) as any,
          meta: {
            title: menu.meta?.title || '',
            icon: menu.meta?.icon,
            keepAlive: menu.meta?.keepAlive,
            hidden: menu.meta?.hidden,
            permission: menu.meta?.permission,
            affix: menu.meta?.affix,
          },
        })
      } else if (routeChildren.length > 0) {
        // 纯目录类型（无 component，有可路由的子菜单）：递归处理
        extractLeafRoutes(routeChildren)
      }
      // 既没有 component 也没有可路由子节点的菜单自动跳过
    }
  }

  extractLeafRoutes(menus)
  return routes
}

const router = createRouter({
  history: createWebHistory(),
  routes: [...constantRoutes, ...staticAuthRoutes],
})

// 白名单（无需登录即可访问的路由）
const whiteList = ['/login', '/404', '/sso/login', '/sso/register', '/sso/forgot-password', '/sso/consent', '/sso/social-callback', '/sso/social-bind']

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  NProgress.start()
  cancelAll()

  // 免登录模式：强制不访问 /login，登录相关路由一律回首页；token 分支中建立静态动态路由。
  if (!AUTH_ENABLED && (to.path === '/login' || to.path === '/sso/login')) {
    next('/')
    NProgress.done()
    return
  }

  // 从 store 获取 token
  const userStore = useUserStore()
  const token = userStore.token

  if (token) {
    if (to.path === '/login') {
      next('/')
      NProgress.done()
    } else {
      if (!userStore.hasRoutes) {
        try {
          // 获取用户信息和菜单
          await userStore.getUserInfo()
          const menus = await userStore.getMenus()
          await userStore.getPermissions()

          // 收集静态子路由路径，防止动态路由覆盖
          const staticPaths = new Set(staticAuthRoutes.flatMap((r) => (r.children || []).map((c) => c.path.replace(/^\//, ''))))

          // 将后端菜单转为路由并注册到 MainLayout 下
          const dynamicRoutes = generateRoutes(menus)
          dynamicRoutes.forEach((route) => {
            // 子路由需要使用相对路径（去掉开头的 /）
            if (route.path.startsWith('/')) {
              route.path = route.path.substring(1)
            }
            // 跳过与静态路由重复的路径，保留静态路由的完整 meta（如 affix）
            if (staticPaths.has(route.path)) return
            router.addRoute('MainLayout', route)
          })

          // 添加兜底 404 路由（必须放最后）
          router.addRoute({
            path: '/:pathMatch(.*)*',
            redirect: '/404',
          })

          userStore.hasRoutes = true

          // 过滤已持久化但路径已不存在的标签（菜单变更后防止死链）
          const tagsViewStore = useTagsViewStore()
          const registeredPaths = new Set(
            router
              .getRoutes()
              .filter((r) => r.name)
              .map((r) => r.path)
          )
          tagsViewStore.restoreValidTags(registeredPaths)

          next({ ...to, replace: true })
        } catch {
          // 免登录模式：不跳 /login，保持已初始化的静态路由，仅提示后回首页
          if (!AUTH_ENABLED) {
            next('/dashboard')
            NProgress.done()
            return
          }
          userStore.resetState()
          next('/login')
          NProgress.done()
        }
      } else {
        next()
      }
    }
  } else {
    // 免登录：白名单精确路径 或 带 noAuth 标记的路由（如 /share/agent/:token）
    if (whiteList.includes(to.path) || to.meta?.noAuth) {
      next()
    } else if (!AUTH_ENABLED) {
      // 免登录模式下 token 缺失（例如 INTERNAL_TOKEN 忘配）：不再弹登录页，
      // 直接放行至目标路由，后续接口失败由 request 拦截器弹一次友好提示
      next()
    } else {
      next(`/login?redirect=${to.fullPath}`)
      NProgress.done()
    }
  }
})

router.afterEach((to) => {
  NProgress.done()
  // KeepAlive 缓存管理
  const keepAliveStore = useKeepAliveStore()
  if (to.name && to.meta?.keepAlive) {
    keepAliveStore.addCache(to.name as string)
  }
})

export default router
