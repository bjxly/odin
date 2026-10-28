<template>
  <div class="login-page">
    <component
      :is="currentThemeComponent"
      :title="appTitle"
    />
    <!-- 语言切换 -->
    <div class="login-page__lang">
      <el-dropdown
        trigger="click"
        @command="changeLocale"
      >
        <span class="login-page__lang-btn">
          {{ currentLocale === 'zh-CN' ? '中文' : 'English' }}
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
            <el-dropdown-item command="en-US">English</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppTitle } from '@/composables/useAppTitle'
import { fetchPublicConfig } from '@/composables/usePublicConfig'
import { LOGIN_THEME_KEY } from '@/utils/constants'
import LoginThemeDark from './themes/LoginThemeDark.vue'
import LoginThemeLight from './themes/LoginThemeLight.vue'
import LoginThemeSplit from './themes/LoginThemeSplit.vue'

const { locale } = useI18n()
const currentLocale = computed(() => locale.value)

// 系统名称：从 sys_config sys.name 参数读取
const appTitle = useAppTitle()

// 登录页主题：从 sys_config ui.login.theme 参数读取
const publicConfig = ref<Record<string, string> | null>(null)
// 同步读取本地缓存的主题，保证首帧（含退出登录后）即使用正确主题，避免先闪 dark 再切换
const cachedTheme = localStorage.getItem(LOGIN_THEME_KEY)
const currentTheme = computed(() => publicConfig.value?.['ui.login.theme'] ?? cachedTheme ?? 'dark')
const currentThemeComponent = computed(() => {
  switch (currentTheme.value) {
    case 'light':
      return LoginThemeLight
    case 'split':
      return LoginThemeSplit
    case 'dark':
    default:
      return LoginThemeDark
  }
})

onMounted(() => {
  fetchPublicConfig().then((cfg) => {
    publicConfig.value = cfg
  })
})

function changeLocale(lang: string) {
  locale.value = lang
  localStorage.setItem('sm-locale', lang)
}
</script>

<style scoped>
.login-page {
  position: relative;
}

.login-page__lang {
  position: fixed;
  top: 24px;
  right: 24px;
  z-index: 20;
}

.login-page__lang-btn {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  user-select: none;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: rgba(0, 0, 0, 0.9);
  transition: background 0.2s ease;
}

.login-page__lang-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>
