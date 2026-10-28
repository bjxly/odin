<template>
  <div class="sm-layout-horizontal">
    <!-- 桌面/平板：正常布局 -->
    <template v-if="!isMobile">
      <!-- 顶部导航栏 -->
      <NavHeader
        :show-logo="true"
        :offset-left="0"
        @open-config="showConfig = true"
      >
        <el-menu
          :default-active="activeMenu"
          mode="horizontal"
          background-color="transparent"
          text-color="#ffffff"
          active-text-color="#ffd04b"
          :ellipsis="false"
          router
          popper-class="sm-horiz-popper"
          class="sm-horizontal-menu"
        >
          <HorizMenuItem
            v-for="item in filteredMenus"
            :key="item.path"
            :item="item"
          />
        </el-menu>
      </NavHeader>
      <!-- 布局配置面板 -->
      <LayoutConfig v-model:visible="showConfig" />
      <!-- 标签栏 -->
      <TagsView v-if="themeStore.showTagsView" />
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
import { useRoute } from 'vue-router'
import { useThemeStore } from '@/stores/theme'
import { useKeepAliveStore } from '@/stores/keepAlive'
import { useWatermark } from '@/composables/useWatermark'
import { useResponsive } from '@/composables/useResponsive'
import LayoutConfig from '@/components/LayoutConfig.vue'
import NavHeader from '@/layouts/components/NavHeader.vue'
import TagsView from './components/TagsView.vue'
import MobileTabBar from './components/MobileTabBar.vue'
import { useUserStore } from '@/stores/user'
import HorizMenuItem from './components/HorizMenuItem.vue'

const route = useRoute()
const themeStore = useThemeStore()
const keepAliveStore = useKeepAliveStore()
const showConfig = ref(false)
const { createWatermark, removeWatermark } = useWatermark()
const { isMobile } = useResponsive()
const userStore = useUserStore()

/** 过滤顶级菜单 */
const filteredMenus = computed(() => {
  return (userStore.menus || []).filter((item) => item.path && item.meta?.hidden !== true)
})

const mainStyle = computed(() => {
  const top = themeStore.showTagsView ? 94 : 56
  return {
    paddingTop: top + 'px',
    minHeight: `calc(100vh - ${top}px)`,
  }
})

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

// 当前激活菜单
const activeMenu = computed(() => route.path)
</script>

<style scoped>
.sm-layout-horizontal {
  height: 100vh;
  min-height: 100vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background-color: var(--sm-bg-page);
}

.sm-horizontal-menu {
  border: none;
  --el-menu-hover-text-color: #ffd04b;
  --el-menu-hover-bg-color: rgba(255, 255, 255, 0.08);
}

.sm-layout-main {
  flex: 1;
  min-height: 0;
  width: 100%;
  background-color: var(--sm-bg-page);
}

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

/* 横向菜单下拉弹出面板：覆盖白色文字，适配浅色/暗色主题 */
:deep(.el-menu--horizontal .el-menu--popup-container) {
  --el-menu-text-color: var(--sm-text-primary);
  --el-menu-active-color: var(--sm-color-primary);
  --el-menu-hover-bg-color: var(--sm-bg-hover);
}

:deep(.el-menu--horizontal .el-menu--popup-container .el-menu-item),
:deep(.el-menu--horizontal .el-menu--popup-container .el-sub-menu__title) {
  color: var(--sm-text-primary) !important;
}

:deep(.el-menu--horizontal .el-menu--popup-container .el-menu-item:hover),
:deep(.el-menu--horizontal .el-menu--popup-container .el-sub-menu__title:hover) {
  background-color: var(--sm-bg-hover) !important;
}

:deep(.el-menu--horizontal .el-menu--popup-container .el-menu-item.is-active) {
  color: var(--sm-color-primary) !important;
}
</style>

<!-- 全局样式：处理被 teleport 到 body 下的横向菜单弹出面板 -->
<style>
.sm-horiz-popper {
  --el-menu-text-color: var(--sm-text-primary) !important;
  --el-menu-active-color: var(--sm-color-primary) !important;
  --el-menu-hover-bg-color: var(--sm-bg-hover) !important;
}

.sm-horiz-popper .el-menu-item,
.sm-horiz-popper .el-sub-menu__title {
  color: var(--sm-text-primary) !important;
}

.sm-horiz-popper .el-menu-item:hover,
.sm-horiz-popper .el-sub-menu__title:hover {
  background-color: var(--sm-bg-hover) !important;
}

.sm-horiz-popper .el-menu-item.is-active {
  color: var(--sm-color-primary) !important;
}
</style>
