<template>
  <div
    class="column-bar"
    :style="{ background: themeStore.navBackground }"
  >
    <!-- Logo -->
    <div class="column-logo">
      <el-icon class="text-white text-2xl"><Timer /></el-icon>
    </div>
    <!-- 一级菜单图标 -->
    <div
      v-for="item in filteredMenus"
      :key="item.path"
      class="column-item"
      :class="{ active: isActive(item) }"
      @click="handleClick(item)"
    >
      <el-icon class="text-white text-lg"><component :is="item.meta?.icon" /></el-icon>
      <span class="column-item-text">{{ $t(item.meta?.title ?? '') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Timer } from '@element-plus/icons-vue'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user'
import type { MenuTreeVO } from '@/types/sys'

const themeStore = useThemeStore()
const userStore = useUserStore()
const route = useRoute()
const router = useRouter()

/** 过滤掉按钮权限项（path 为 null）和隐藏项 */
function getVisibleChildren(item: MenuTreeVO): MenuTreeVO[] {
  return (item.children || []).filter((child) => child.path && child.meta?.hidden !== true)
}

/** 过滤顶级菜单 */
const filteredMenus = computed(() => {
  return (userStore.menus || []).filter((item) => item.path && item.meta?.hidden !== true)
})

function isActive(item: MenuTreeVO): boolean {
  const children = getVisibleChildren(item)
  if (children.length && !item.component) {
    return children.some((child) => {
      const prefix = child.path.split('/').slice(0, 2).join('/')
      return route.path.startsWith(prefix)
    })
  }
  const prefix = item.path.split('/').slice(0, 2).join('/')
  return route.path.startsWith(prefix)
}

function handleClick(item: MenuTreeVO) {
  // 有可显示子菜单时跳转第一个子项，否则跳转自身
  const children = getVisibleChildren(item)
  if (children.length && !item.component) {
    router.push(children[0].path)
  } else {
    router.push(item.path)
  }
}
</script>
