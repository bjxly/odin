import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

/**
 * 响应式 Element Plus locale 切换
 * 根据 vue-i18n 的当前语言动态返回对应的 Element Plus locale，
 * 配合 el-config-provider 实现无需刷新页面的组件文案切换。
 */
export function useElementLocale() {
  const { locale } = useI18n()

  const elementLocale = computed(() => {
    return locale.value === 'en-US' ? en : zhCn
  })

  return { elementLocale }
}
