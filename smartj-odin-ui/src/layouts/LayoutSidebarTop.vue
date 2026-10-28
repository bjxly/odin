<template>
  <div class="sm-layout-sidebar-top">
    <!-- 桌面/平板：正常布局 -->
    <template v-if="!isMobile">
      <!-- 顶部栏（全宽，含Logo） -->
      <NavHeader
        :show-logo="true"
        :offset-left="0"
        @open-config="showConfig = true"
      />
      <!-- 布局配置面板 -->
      <LayoutConfig v-model:visible="showConfig" />
      <!-- 标签栏 -->
      <TagsView
        v-if="themeStore.showTagsView"
        :style="{ left: (sideCollapse ? 64 : 220) + 'px' }"
      />
      <!-- 下半部分：左侧菜单 + 内容 -->
      <div class="sm-sidebar-top-body">
        <!-- 左侧菜单占位（SideMenu是fixed定位，用占位div撑开空间） -->
        <div
          class="sm-sidebar-top-left"
          :style="{ width: sideCollapse ? '64px' : '220px' }"
        >
          <SideMenu
            v-model:collapse="sideCollapse"
            :show-logo="false"
            class="sm-sidebar-top-menu"
          />
        </div>
        <!-- 内容区 -->
        <main
          class="sm-layout-main"
          :style="mainStyle"
        >
          <router-view v-slot="{ Component, route }">
            <keep-alive :include="keepAliveStore.cachedViews">
              <component
                :is="Component"
                :key="route.fullPath"
              />
            </keep-alive>
          </router-view>
        </main>
      </div>
    </template>
    <!-- 移动端：简化布局 -->
    <template v-else>
      <main class="sm-mobile-main">
        <router-view v-slot="{ Component, route }">
          <keep-alive :include="keepAliveStore.cachedViews">
            <component
              :is="Component"
              :key="route.fullPath"
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
import { useThemeStore } from '@/stores/theme'
import { useKeepAliveStore } from '@/stores/keepAlive'
import { useWatermark } from '@/composables/useWatermark'
import { useResponsive } from '@/composables/useResponsive'
import LayoutConfig from '@/components/LayoutConfig.vue'
import NavHeader from '@/layouts/components/NavHeader.vue'
import SideMenu from '@/layouts/components/SideMenu.vue'
import TagsView from './components/TagsView.vue'
import MobileTabBar from './components/MobileTabBar.vue'

const themeStore = useThemeStore()
const keepAliveStore = useKeepAliveStore()
const showConfig = ref(false)
const sideCollapse = ref(false)
const { createWatermark, removeWatermark } = useWatermark()
const { isMobile, isTablet } = useResponsive()

// 平板模式下自动折叠侧边栏
watch(
  isTablet,
  (val) => {
    sideCollapse.value = val
  },
  { immediate: true }
)

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
</script>

<style scoped>
.sm-layout-sidebar-top {
  height: 100vh;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sm-sidebar-top-body {
  flex: 1;
  min-height: 0;
  display: flex;
}

.sm-sidebar-top-left {
  flex-shrink: 0;
  transition: width 0.3s;
}

.sm-sidebar-top-menu {
  top: 56px;
  height: calc(100vh - 56px);
}

.sm-layout-main {
  flex: 1;
  min-width: 0;
  background-color: var(--sm-bg-page);
}
</style>
