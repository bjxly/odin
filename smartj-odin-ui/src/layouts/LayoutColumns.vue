<template>
  <div class="sm-layout-columns">
    <!-- 桌面/平板：正常布局 -->
    <template v-if="!isMobile">
      <!-- 左侧一级图标栏 -->
      <ColumnBar />
      <!-- 二级菜单栏 -->
      <div
        v-if="activeMenuChildren.length"
        class="sm-submenu-bar"
        :style="{ background: submenuBackground }"
      >
        <div class="sm-submenu-title">{{ activeMenuItem?.meta?.title }}</div>
        <template
          v-for="child in activeMenuChildren"
          :key="child.item.path"
        >
          <!-- 目录节点：作为分组标题，不可点击 -->
          <div
            v-if="child.isDir"
            class="sm-submenu-section"
          >
            <span>{{ child.item.meta?.title }}</span>
          </div>
          <!-- 叶子节点：可点击导航 -->
          <div
            v-else
            class="sm-submenu-item"
            :class="{ active: route.path === child.item.path, nested: child.nested }"
            @click="handleSubmenuClick(child.item)"
          >
            <el-icon v-if="themeStore.showNavIcon && child.item.meta?.icon"><component :is="child.item.meta.icon" /></el-icon>
            <span>{{ child.item.meta?.title }}</span>
          </div>
        </template>
      </div>
      <!-- 右侧区域 -->
      <div
        class="sm-columns-right"
        :style="{ marginLeft: totalLeft + 'px' }"
      >
        <!-- 顶部栏 -->
        <NavHeader
          :show-logo="false"
          :offset-left="totalLeft"
          @open-config="showConfig = true"
        />
        <!-- 布局配置面板 -->
        <LayoutConfig v-model:visible="showConfig" />
        <!-- 标签栏 -->
        <TagsView
          v-if="themeStore.showTagsView"
          :style="{ left: totalLeft + 'px' }"
        />
        <!-- 内容区 -->
        <main
          class="sm-layout-main"
          :style="mainStyle"
        >
          <router-view v-slot="{ Component, route: routeData }">
            <keep-alive :include="keepAliveStore.cachedViews">
              <component
                :is="Component"
                :key="routeData.fullPath"
              />
            </keep-alive>
          </router-view>
        </main>
      </div>
    </template>
    <!-- 移动端：简化布局 -->
    <template v-else>
      <main class="sm-mobile-main">
        <router-view v-slot="{ Component, route: routeData }">
          <keep-alive :include="keepAliveStore.cachedViews">
            <component
              :is="Component"
              :key="routeData.fullPath"
            />
          </keep-alive>
        </router-view>
      </main>
      <MobileTabBar />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useThemeStore } from '@/stores/theme'
import { useKeepAliveStore } from '@/stores/keepAlive'
import { useWatermark } from '@/composables/useWatermark'
import { useResponsive } from '@/composables/useResponsive'
import LayoutConfig from '@/components/LayoutConfig.vue'
import NavHeader from '@/layouts/components/NavHeader.vue'
import ColumnBar from '@/layouts/components/ColumnBar.vue'
import TagsView from './components/TagsView.vue'
import MobileTabBar from './components/MobileTabBar.vue'
import { useUserStore } from '@/stores/user'
import type { MenuTreeVO } from '@/types/sys'

const route = useRoute()
const router = useRouter()
const themeStore = useThemeStore()
const keepAliveStore = useKeepAliveStore()
const showConfig = ref(false)
const { createWatermark, removeWatermark } = useWatermark()
const { isMobile } = useResponsive()
const userStore = useUserStore()

const mainStyle = computed(() => {
  const top = themeStore.showTagsView ? 94 : 56
  return {
    paddingTop: top + 'px',
    minHeight: `calc(100vh - ${top}px)`,
  }
})

// 二级菜单栏背景色：与一级栏（ColumnBar）同色系，承接渐变终止色
const submenuBackground = computed(() => themeStore.navBgEnd)

/** 过滤掉按钮权限项和隐藏项 */
function getVisibleChildren(item: MenuTreeVO): MenuTreeVO[] {
  return (item.children || []).filter((child) => child.path && child.meta?.hidden !== true)
}

/** 判断是否为目录节点（有可见子菜单且自身无 component） */
function isDirectory(item: MenuTreeVO): boolean {
  return getVisibleChildren(item).length > 0 && !item.component
}

/** 二级面板菜单项（含目录分组标题和叶子导航项） */
interface SubMenuItem {
  item: MenuTreeVO
  isDir: boolean
  nested: boolean
}

/** 递归展开子菜单，保留目录节点作为分组标题，叶子节点作为可点击项 */
function getFlattenedWithDirs(item: MenuTreeVO, nested = false): SubMenuItem[] {
  const children = getVisibleChildren(item)
  if (!children.length) return []
  const result: SubMenuItem[] = []
  for (const child of children) {
    if (isDirectory(child)) {
      // 目录节点：自身作为分组标题，子项标记为 nested
      result.push({ item: child, isDir: true, nested: false })
      result.push(...getFlattenedWithDirs(child, true))
    } else {
      result.push({ item: child, isDir: false, nested })
    }
  }
  return result
}

/** 过滤顶级菜单 */
const filteredMenus = computed(() => {
  return (userStore.menus || []).filter((item) => item.path && item.meta?.hidden !== true)
})

// 当前激活的一级菜单（匹配所有子项路径，避免切换子路由时丢失激活状态）
const activeMenuItem = computed(() => {
  return filteredMenus.value.find((item) => {
    const children = getVisibleChildren(item)
    if (children.length && !item.component) {
      return children.some((child) => {
        const prefix = child.path.split('/').slice(0, 2).join('/')
        return route.path.startsWith(prefix)
      })
    }
    const prefix = item.path.split('/').slice(0, 2).join('/')
    return route.path.startsWith(prefix)
  })
})

// 当前激活一级菜单的展开项（含目录分组标题和叶子导航项）
const activeMenuChildren = computed(() => {
  if (!activeMenuItem.value) return []
  return getFlattenedWithDirs(activeMenuItem.value)
})

// 总左侧偏移量（一级栏64px + 二级栏140px）
const totalLeft = computed(() => 64 + (activeMenuChildren.value.length ? 140 : 0))

function handleSubmenuClick(child: MenuTreeVO) {
  if (child.path) {
    router.push(child.path)
  }
}

// 水印
watch(
  () => themeStore.watermark,
  (val) => {
    if (val) createWatermark('SmartX 调度平台')
    else removeWatermark()
  },
  { immediate: true }
)

onMounted(() => {
  themeStore.initTheme()
})

onUnmounted(() => {
  removeWatermark()
})
</script>

<style scoped>
.sm-layout-columns {
  min-height: 100vh;
}

.sm-columns-right {
  /* margin-left 由行内样式根据 totalLeft 动态控制 */
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.sm-layout-main {
  flex: 1;
  min-height: 0;
  background-color: var(--sm-bg-page);
}

/* === 二级菜单栏 === */
.sm-submenu-bar {
  position: fixed;
  left: 64px;
  top: 0;
  bottom: 0;
  width: 140px;
  z-index: 51;
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  flex-direction: column;
  padding-top: var(--sm-header-height);
  box-sizing: border-box;
  overflow: hidden;
}

.sm-submenu-title {
  padding: 14px 16px;
  font-size: 13px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.85);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  white-space: nowrap;
}

.sm-submenu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 11px 16px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  transition:
    background-color 0.2s,
    color 0.2s;
  white-space: nowrap;
}

.sm-submenu-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: #ffffff;
}

.sm-submenu-item.active {
  background-color: var(--sm-color-primary);
  color: #ffffff;
}

/* 目录分组标题（不可点击） */
.sm-submenu-section {
  padding: 12px 16px 4px;
  font-size: 11px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.35);
  text-transform: uppercase;
  letter-spacing: 1px;
  white-space: nowrap;
  cursor: default;
}

/* 嵌套子项：缩进 + 左侧连接线 */
.sm-submenu-item.nested {
  padding-left: 28px;
  position: relative;
}

.sm-submenu-item.nested::before {
  content: '';
  position: absolute;
  left: 14px;
  top: 0;
  bottom: 0;
  width: 2px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 1px;
}
</style>
