<template>
  <header
    class="fixed top-0 right-0 z-50 h-14 flex items-center px-4 shadow-md"
    :style="{ background: themeStore.navBackground, left: offsetLeft + 'px' }"
  >
    <!-- Logo -->
    <div
      v-if="showLogo"
      class="flex items-center mr-8"
    >
      <img
        src="/logo.png"
        alt="logo"
        class="h-7 w-auto"
      />
      <span class="text-white font-bold text-lg tracking-wide ml-2">{{ appTitle }}</span>
    </div>
    <!-- 默认插槽（横向模式放菜单，其他模式可为空或面包屑） -->
    <div class="flex-1">
      <slot />
    </div>
    <!-- 租户切换器（公开配置开关开启且用户存在多租户时渲染；免登录模式下隐藏） -->
    <TenantSwitcher v-if="authEnabled" />
    <!-- 布局配置按钮 -->
    <el-button
      :icon="Setting"
      circle
      size="small"
      style="color: #ffffff; border-color: rgba(255, 255, 255, 0.3); background: transparent"
      @click="$emit('openConfig')"
    />
    <!-- 语言切换 -->
    <el-dropdown
      class="ml-3"
      @command="changeLocale"
    >
      <el-button
        circle
        size="small"
        style="color: #ffffff; border-color: rgba(255, 255, 255, 0.3); background: transparent"
      >
        {{ currentLocale === 'zh-CN' ? '中' : 'En' }}
      </el-button>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
          <el-dropdown-item command="en-US">English</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
    <!-- 用户区 -->
    <div class="flex items-center ml-4">
      <el-dropdown @command="handleCommand">
        <div class="flex items-center cursor-pointer">
          <el-avatar
            :size="28"
            class="mr-2"
            :icon="UserFilled"
          />
          <span class="text-white text-sm">{{ userStore.userInfo?.nickname || userStore.userInfo?.username || $t('layout.admin') }}</span>
          <el-icon class="text-white ml-1"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item
              v-if="authEnabled"
              command="profile"
            >
              <el-icon><User /></el-icon>
              <span>{{ $t('layout.profile') }}</span>
            </el-dropdown-item>
            <el-dropdown-item
              v-if="authEnabled"
              command="logout"
            >
              <el-icon><SwitchButton /></el-icon>
              <span>{{ $t('layout.logout') }}</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Setting, UserFilled, ArrowDown, SwitchButton, User } from '@element-plus/icons-vue'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user'
import { useAppTitle } from '@/composables/useAppTitle'
import { AUTH_ENABLED } from '@/config/runtime'
import TenantSwitcher from './TenantSwitcher.vue'

const appTitle = useAppTitle()
// 免登录模式下隐藏租户切换、个人中心、退出登录入口
const authEnabled = AUTH_ENABLED

withDefaults(
  defineProps<{
    showLogo?: boolean
    offsetLeft?: number
  }>(),
  {
    showLogo: true,
    offsetLeft: 0,
  }
)

defineEmits<{
  (e: 'openConfig'): void
}>()

const themeStore = useThemeStore()
const userStore = useUserStore()

const { locale } = useI18n()
const currentLocale = computed(() => locale.value)
function changeLocale(lang: string) {
  locale.value = lang
  localStorage.setItem('sm-locale', lang)
}

const router = useRouter()

function handleCommand(command: string) {
  if (command === 'profile') {
    router.push('/profile')
  } else if (command === 'logout') {
    userStore.logout()
  }
}
</script>

<style scoped>
:deep(.el-menu--horizontal > .el-menu-item),
:deep(.el-menu--horizontal > .el-sub-menu :deep(.el-sub-menu__title)) {
  height: 56px;
  line-height: 56px;
  border-bottom: none;
  background-color: transparent !important;
}
:deep(.el-menu--horizontal > .el-menu-item.is-active),
:deep(.el-menu--horizontal > .el-sub-menu.is-active :deep(.el-sub-menu__title)) {
  background-color: rgba(255, 255, 255, 0.1) !important;
  border-bottom: none;
}
</style>
