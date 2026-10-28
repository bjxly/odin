<template>
  <!-- 有可见子菜单且自身无 component：渲染为可展开的子菜单目录 -->
  <el-sub-menu
    v-if="hasVisibleChildren && !item.component"
    :index="item.path"
  >
    <template #title>
      <el-icon v-if="themeStore.showNavIcon && item.meta?.icon">
        <component :is="item.meta.icon" />
      </el-icon>
      <span>{{ $t(item.meta?.title ?? '') }}</span>
    </template>
    <!-- 递归渲染子菜单 -->
    <SideMenuItem
      v-for="child in visibleChildren"
      :key="child.path"
      :item="child"
      :collapse="collapse"
      :depth="depth + 1"
    />
  </el-sub-menu>
  <!-- 无可见子菜单或自身有 component：渲染为叶子菜单项 -->
  <el-menu-item
    v-else
    :index="item.path"
  >
    <el-icon v-if="themeStore.showNavIcon && item.meta?.icon">
      <component :is="item.meta.icon" />
    </el-icon>
    <template #title>{{ $t(item.meta?.title || '') }}</template>
  </el-menu-item>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useThemeStore } from '@/stores/theme'
import type { MenuTreeVO } from '@/types/sys'

defineOptions({ name: 'SideMenuItem' })

const props = withDefaults(
  defineProps<{
    item: MenuTreeVO
    collapse?: boolean
    depth?: number
  }>(),
  {
    collapse: false,
    depth: 0,
  }
)

const themeStore = useThemeStore()

/** 过滤出可显示的子菜单：有 path 且未隐藏 */
const visibleChildren = computed<MenuTreeVO[]>(() => {
  return (props.item.children || []).filter((child) => child.path && child.meta?.hidden !== true)
})

/** 是否有可显示的子菜单 */
const hasVisibleChildren = computed(() => visibleChildren.value.length > 0)
</script>
