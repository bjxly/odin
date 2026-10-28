<template>
  <component :is="layoutComponent" />
  <!-- 全局溯源抽屉：由 useOdinTraceStore().openTrace(traceId) 唤起 -->
  <TraceDrawer />
  <!-- 全局 ODIN AI 副驾：仅 ODIN 业务路由（/odin/**）显示浮动入口，AI 助手页自身除外 -->
  <OdinAiFab v-if="isOdinRoute && !isAssistantRoute" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useThemeStore } from '@/stores/theme'
import TraceDrawer from '@/views/odin/components/trace/TraceDrawer.vue'
import OdinAiFab from '@/views/odin/components/assistant/OdinAiFab.vue'
import LayoutHorizontal from './LayoutHorizontal.vue'
import LayoutSidebar from './LayoutSidebar.vue'
import LayoutSidebarTop from './LayoutSidebarTop.vue'
import LayoutColumns from './LayoutColumns.vue'

const themeStore = useThemeStore()
const route = useRoute()

/** ODIN 业务路由（菜单种子中全部 ODIN 页面路径均以 /odin 开头） */
const isOdinRoute = computed(() => route.path.startsWith('/odin'))

/** AI 助手页：页面内已有完整对话能力，隐藏右下角副驾入口 */
const isAssistantRoute = computed(() => route.path.startsWith('/odin/assistant'))

const layoutMap: Record<string, any> = {
  horizontal: LayoutHorizontal,
  sidebar: LayoutSidebar,
  'sidebar-top': LayoutSidebarTop,
  columns: LayoutColumns,
}

const layoutComponent = computed(() => {
  return layoutMap[themeStore.layoutMode] || LayoutHorizontal
})
</script>
