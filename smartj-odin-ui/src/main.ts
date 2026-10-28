import { createApp } from 'vue'
import App from './App.vue'

// 路由
import router from './router'

// 状态管理
import pinia from './stores'

// Element Plus
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

// 国际化
import i18n from './i18n'

// 权限指令
import { permission } from './directives/permission'
// 图片懒加载指令
import { lazy } from './directives/lazy'

// 共享的 SSO 公开配置获取器
import { fetchPublicConfig } from './composables/usePublicConfig'

// 样式体系（按优先级从低到高加载）
import './styles/index.css' // 1. Tailwind + 全局基础
import './styles/variables.css' // 2. CSS 变量（浅色 + 深色）
import './styles/element-override.css' // 3. Element Plus 覆盖
import './styles/dark.css' // 4. 深色模式组件样式
import './styles/layout.css' // 5. 页面布局
import './styles/sm-common.css' // 6. 业务通用样式
import './styles/responsive.css' // 7. 响应式
import './styles/transition.css' // 8. 过渡动画
import './styles/nprogress.css' // 9. NProgress 进度条

const app = createApp(App)

// 注册 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(i18n)

// 注册自定义指令
app.directive('permission', permission)
app.directive('lazy', lazy)

// 动态更新页面标题和 SEO meta（从 sys_config ui.login.theme 参数读取，环境变量作为 fallback）
const appName = (import.meta.env.VITE_APP_TITLE as string) || 'SmartJ'
updateAppMeta(appName)
// 从参数配置获取实际名称并更新（共享请求，全局只发一次）
fetchPublicConfig()
  .then((cfg) => {
    const name = cfg?.['sys.name']
    if (name) {
      updateAppMeta(name)
    }
  })
  .catch(() => {
    /* 获取失败时保留默认值 */
  })

function updateAppMeta(name: string) {
  document.title = name
  const metaMap: Record<string, string> = {
    'og:title': name,
    'og:site_name': name,
    'twitter:title': name,
  }
  for (const [prop, value] of Object.entries(metaMap)) {
    const selector = prop.startsWith('og:') ? `meta[property="${prop}"]` : `meta[name="${prop}"]`
    const el = document.querySelector(selector)
    if (el) el.setAttribute('content', value)
  }
  // 更新 description 中的 SmartJ 替换为动态名称
  const descSelectors = ['meta[name="description"]', 'meta[property="og:description"]', 'meta[name="twitter:description"]']
  for (const sel of descSelectors) {
    const el = document.querySelector(sel)
    if (el) {
      const original = el.getAttribute('content') || ''
      el.setAttribute('content', original.replace(/SmartJ/g, name))
    }
  }
  // 更新 JSON-LD 结构化数据中的 name
  const ldScript = document.querySelector('script[type="application/ld+json"]')
  if (ldScript) {
    try {
      const ld = JSON.parse(ldScript.textContent || '{}')
      ld.name = name
      if (ld.description) ld.description = ld.description.replace(/SmartJ/g, name)
      ldScript.textContent = JSON.stringify(ld)
    } catch {
      /* ignore */
    }
  }
}

app.mount('#app')
