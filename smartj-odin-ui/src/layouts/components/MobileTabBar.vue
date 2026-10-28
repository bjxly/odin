<template>
  <div class="mobile-tabbar">
    <!-- 主 Tab 区域 -->
    <div
      v-for="item in visibleTabs"
      :key="item.path"
      class="mobile-tabbar__item"
      :class="{ 'is-active': isActive(item) }"
      @click="handleTabClick(item)"
    >
      <el-icon
        class="mobile-tabbar__icon"
        :size="20"
      >
        <component :is="item.meta?.icon" />
      </el-icon>
      <span class="mobile-tabbar__label">{{ $t(item.meta?.title ?? '') }}</span>
    </div>

    <!-- 更多按钮 -->
    <div
      class="mobile-tabbar__item"
      :class="{ 'is-active': drawerVisible }"
      @click="drawerVisible = true"
    >
      <el-icon
        class="mobile-tabbar__icon"
        :size="20"
      >
        <More />
      </el-icon>
      <span class="mobile-tabbar__label">{{ $t('common.more') }}</span>
    </div>

    <!-- 更多菜单 Drawer -->
    <el-drawer
      v-model="drawerVisible"
      direction="btt"
      :show-close="false"
      :size="'auto'"
      class="mobile-tabbar__drawer"
    >
      <template #header>
        <span class="mobile-tabbar__drawer-title">{{ $t('common.more') }}</span>
      </template>
      <div class="mobile-tabbar__drawer-body">
        <!-- 完整菜单列表 -->
        <div
          v-for="item in flatMenuItems"
          :key="item.path"
          class="mobile-tabbar__drawer-item"
          :class="{ 'is-active': isActive(item) }"
          @click="handleDrawerMenuClick(item)"
        >
          <el-icon
            class="mobile-tabbar__drawer-item-icon"
            :size="18"
          >
            <component :is="item.meta?.icon" />
          </el-icon>
          <span class="mobile-tabbar__drawer-item-label">{{ $t(item.meta?.title ?? '') }}</span>
        </div>

        <!-- 分隔线 -->
        <div class="mobile-tabbar__drawer-divider" />

        <!-- 语言切换 -->
        <div
          class="mobile-tabbar__drawer-item"
          @click="toggleLocale"
        >
          <el-icon
            class="mobile-tabbar__drawer-item-icon"
            :size="18"
          >
            <Odometer />
          </el-icon>
          <span class="mobile-tabbar__drawer-item-label">
            {{ currentLocale === 'zh-CN' ? '中文 / English' : 'English / 中文' }}
          </span>
        </div>

        <!-- 深色模式切换 -->
        <div
          class="mobile-tabbar__drawer-item"
          @click="themeStore.toggleDark()"
        >
          <el-icon
            class="mobile-tabbar__drawer-item-icon"
            :size="18"
          >
            <component :is="themeStore.isDark ? Sunny : Moon" />
          </el-icon>
          <span class="mobile-tabbar__drawer-item-label">
            {{ themeStore.isDark ? $t('common.lightMode', '浅色模式') : $t('common.darkMode', '深色模式') }}
          </span>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { More, Sunny, Moon, Odometer } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { MenuTreeVO } from '@/types/sys'
import { useThemeStore } from '@/stores/theme'

const VISIBLE_TAB_COUNT = 4

const router = useRouter()
const route = useRoute()
const { locale } = useI18n()
const themeStore = useThemeStore()
const userStore = useUserStore()

const drawerVisible = ref(false)

/** 过滤按钮权限项和隐藏项 */
function getVisibleChildren(item: MenuTreeVO): MenuTreeVO[] {
  return (item.children || []).filter((child) => child.path && child.meta?.hidden !== true)
}

/** 过滤顶级菜单 */
const filteredMenus = computed(() => {
  return (userStore.menus || []).filter((item) => item.path && item.meta?.hidden !== true)
})

/** 展平菜单树为叶子节点 */
const flatMenuItems = computed(() => {
  const result: MenuTreeVO[] = []
  function flatten(items: MenuTreeVO[]) {
    for (const item of items) {
      const children = getVisibleChildren(item)
      if (children.length && !item.component) {
        flatten(children)
      } else {
        result.push(item)
      }
    }
  }
  flatten(filteredMenus.value)
  return result
})

const visibleTabs = computed(() => {
  return filteredMenus.value.slice(0, VISIBLE_TAB_COUNT)
})

const currentLocale = computed(() => locale.value)

function isActive(item: MenuTreeVO): boolean {
  const path = getMenuPath(item)
  return path ? route.path.startsWith(path) : false
}

function getMenuPath(item: MenuTreeVO): string {
  const children = getVisibleChildren(item)
  if (children.length && !item.component) return children[0].path
  return item.path
}

function handleTabClick(item: MenuTreeVO) {
  const path = getMenuPath(item)
  if (isActive(item) && route.path === path) return
  if (path) router.push(path)
}

function handleDrawerMenuClick(item: MenuTreeVO) {
  drawerVisible.value = false
  const path = getMenuPath(item)
  if (isActive(item) && route.path === path) return
  if (path) router.push(path)
}

function toggleLocale() {
  const newLocale = currentLocale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  locale.value = newLocale
  localStorage.setItem('sm-locale', newLocale)
}
</script>

<style scoped>
.mobile-tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  height: var(--sm-mobile-tabbar-height);
  background: var(--sm-bg-card);
  border-top: 1px solid var(--sm-border-light);
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding: 0;
  box-sizing: border-box;
}

.mobile-tabbar__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  height: 100%;
  cursor: pointer;
  color: var(--sm-text-secondary);
  transition: color 0.2s ease;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
}

.mobile-tabbar__item.is-active {
  color: var(--sm-color-primary);
}

.mobile-tabbar__icon {
  margin-bottom: 2px;
}

.mobile-tabbar__label {
  font-size: var(--sm-font-size-xs);
  line-height: 1;
  white-space: nowrap;
}

/* ---- Drawer 样式 ---- */
.mobile-tabbar__drawer-title {
  font-size: var(--sm-font-size-lg);
  font-weight: 600;
  color: var(--sm-text-primary);
}

.mobile-tabbar__drawer-body {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--sm-spacing-md);
  padding: var(--sm-spacing-sm) 0 var(--sm-spacing-lg);
}

.mobile-tabbar__drawer-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--sm-spacing-sm) 0;
  cursor: pointer;
  color: var(--sm-text-regular);
  border-radius: var(--sm-radius-lg);
  transition:
    background 0.2s ease,
    color 0.2s ease;
  -webkit-tap-highlight-color: transparent;
}

.mobile-tabbar__drawer-item:active {
  background: var(--sm-bg-hover);
}

.mobile-tabbar__drawer-item.is-active {
  color: var(--sm-color-primary);
}

.mobile-tabbar__drawer-item-icon {
  margin-bottom: var(--sm-spacing-xs);
}

.mobile-tabbar__drawer-item-label {
  font-size: var(--sm-font-size-xs);
  line-height: 1.2;
  text-align: center;
  color: inherit;
}

.mobile-tabbar__drawer-divider {
  grid-column: 1 / -1;
  height: 1px;
  background: var(--sm-border-lighter);
  margin: var(--sm-spacing-xs) 0;
}
</style>

<style>
/* Drawer 全局样式覆盖（不使用 scoped） */
.mobile-tabbar__drawer .el-drawer__header {
  margin-bottom: 0;
  padding: var(--sm-spacing-md) var(--sm-spacing-lg);
  border-bottom: 1px solid var(--sm-border-lighter);
}

.mobile-tabbar__drawer .el-drawer__body {
  padding: 0 var(--sm-spacing-lg);
}
</style>
