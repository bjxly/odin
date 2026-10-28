<template>
  <div class="tags-view-container">
    <el-scrollbar class="tags-view-wrapper">
      <router-link
        v-for="tag in sortedTags"
        :key="tag.path"
        :to="tag.fullPath"
        class="tags-view-item"
        :class="{ active: isActive(tag) }"
        @contextmenu.prevent="openMenu(tag, $event)"
      >
        {{ $t(tag.title) }}
        <el-icon
          v-if="!tag.affix"
          class="tags-view-close"
          @click.prevent.stop="closeTag(tag)"
        >
          <Close />
        </el-icon>
      </router-link>
    </el-scrollbar>

    <!-- 虚拟触发元素 -->
    <div
      ref="triggerRef"
      class="tags-view-menu-trigger"
      :style="triggerStyle"
    ></div>

    <!-- 右键菜单使用 el-dropdown -->
    <el-dropdown
      ref="dropdownRef"
      trigger="click"
      :virtual-ref="triggerRef"
      virtual-triggering
      popper-class="tags-view-dropdown"
      @command="handleMenuCommand"
      @visible-change="onDropdownVisibleChange"
    >
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="refresh">
            <el-icon><Refresh /></el-icon>
            <span>{{ $t('layout.refresh') }}</span>
          </el-dropdown-item>
          <el-dropdown-item
            v-if="!selectedTag?.affix"
            command="close"
          >
            <el-icon><Close /></el-icon>
            <span>{{ $t('layout.close') }}</span>
          </el-dropdown-item>
          <el-dropdown-item command="closeOthers">
            <el-icon><CircleClose /></el-icon>
            <span>{{ $t('layout.closeOthers') }}</span>
          </el-dropdown-item>
          <el-dropdown-item command="closeAll">
            <el-icon><FolderDelete /></el-icon>
            <span>{{ $t('layout.closeAll') }}</span>
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Close, Refresh, CircleClose, FolderDelete } from '@element-plus/icons-vue'
import { useTagsViewStore } from '@/stores/tagsView'
import type { TagView } from '@/stores/tagsView'

const route = useRoute()
const router = useRouter()
const tagsViewStore = useTagsViewStore()

const triggerRef = ref<HTMLElement>()
const dropdownRef = ref()
const selectedTag = ref<TagView | null>(null)
const triggerStyle = ref({
  position: 'fixed' as const,
  left: '0px',
  top: '0px',
  width: '0px',
  height: '0px',
  visibility: 'hidden' as const,
})

const sortedTags = computed(() => {
  const affix = tagsViewStore.visitedTags.filter((t) => t.affix)
  const normal = tagsViewStore.visitedTags.filter((t) => !t.affix)
  return [...affix, ...normal]
})

function isActive(tag: TagView): boolean {
  return tag.path === route.path
}

watch(
  () => route.path,
  () => {
    tagsViewStore.addTag(route, router)
  },
  { immediate: true }
)

onMounted(() => {
  const routes = router.getRoutes()
  tagsViewStore.initAffixTags(routes.map((r) => ({ path: r.path, name: r.name as string, meta: r.meta as Record<string, any> })))
  tagsViewStore.addTag(route, router)
})

function openMenu(tag: TagView, e: MouseEvent) {
  selectedTag.value = tag
  // 将虚拟触发元素定位到右键点击位置
  triggerStyle.value = {
    position: 'fixed',
    left: e.clientX + 'px',
    top: e.clientY + 'px',
    width: '0px',
    height: '0px',
    visibility: 'hidden',
  }
  // 在下一帧触发 dropdown 的点击以打开菜单
  nextTick(() => {
    triggerRef.value?.click()
  })
}

function handleMenuCommand(command: string) {
  switch (command) {
    case 'refresh':
      refreshSelected()
      break
    case 'close':
      closeSelected()
      break
    case 'closeOthers':
      closeOthers()
      break
    case 'closeAll':
      closeAll()
      break
  }
}

function onDropdownVisibleChange(visible: boolean) {
  if (!visible) {
    selectedTag.value = null
  }
}

function closeTag(tag: TagView) {
  if (tag.affix) return
  const isCurrentActive = tag.path === route.path
  const lastTag = tagsViewStore.removeTag(tag.path)
  if (isCurrentActive && lastTag) {
    router.push(lastTag.fullPath)
  } else if (isCurrentActive && !lastTag) {
    router.push('/')
  }
}

function refreshSelected() {
  if (selectedTag.value) {
    router.push(selectedTag.value.fullPath)
  }
}

function closeSelected() {
  if (selectedTag.value && !selectedTag.value.affix) {
    closeTag(selectedTag.value)
  }
}

function closeOthers() {
  if (selectedTag.value) {
    tagsViewStore.removeOthers(selectedTag.value.path)
    if (selectedTag.value.path !== route.path) {
      router.push(selectedTag.value.fullPath)
    }
  }
}

function closeAll() {
  const lastTag = tagsViewStore.removeAll()
  if (lastTag) {
    router.push(lastTag.fullPath)
  } else {
    router.push('/')
  }
}
</script>
