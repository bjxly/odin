<template>
  <div
    class="sm-sidebar-menu"
    :style="{ width: collapse ? '64px' : '220px', background: themeStore.navBackground }"
  >
    <!-- Logo 区 -->
    <div
      v-if="showLogo"
      class="sidebar-logo"
    >
      <img
        v-if="!collapse"
        src="/logo.png"
        alt="logo"
        class="h-7 w-auto"
      />
      <img
        v-else
        src="/logo.png"
        alt="logo"
        class="h-6 w-auto"
      />
      <span
        v-if="!collapse"
        class="text-white font-bold ml-2 text-sm"
      >
        {{ appTitle }}
      </span>
    </div>
    <!-- 垂直菜单 -->
    <el-scrollbar class="flex-1">
      <el-menu
        :default-active="activeMenu"
        :collapse="collapse"
        router
        background-color="transparent"
        text-color="rgba(255,255,255,0.8)"
        active-text-color="#ffd04b"
        :collapse-transition="false"
        class="border-none"
      >
        <template
          v-for="item in filteredMenus"
          :key="item.path"
        >
          <SideMenuItem
            :item="item"
            :collapse="collapse"
            :depth="0"
          />
        </template>
      </el-menu>
    </el-scrollbar>
    <!-- 折叠按钮 -->
    <div
      class="collapse-btn"
      @click="$emit('update:collapse', !collapse)"
    >
      <el-icon class="text-white"><component :is="collapse ? Expand : Fold" /></el-icon>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Fold, Expand } from '@element-plus/icons-vue'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user'
import { useAppTitle } from '@/composables/useAppTitle'
import SideMenuItem from './SideMenuItem.vue'

const appTitle = useAppTitle()

withDefaults(
  defineProps<{
    collapse?: boolean
    showLogo?: boolean
  }>(),
  {
    collapse: false,
    showLogo: true,
  }
)

defineEmits<{
  (e: 'update:collapse', val: boolean): void
}>()

const themeStore = useThemeStore()
const userStore = useUserStore()
const route = useRoute()

/** 过滤顶级菜单：只显示有 path 且未隐藏的菜单 */
const filteredMenus = computed(() => {
  return (userStore.menus || []).filter((item) => item.path && item.meta?.hidden !== true)
})

const activeMenu = computed(() => route.path)
</script>
