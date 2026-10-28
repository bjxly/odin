<template>
  <div
    class="s-image"
    :style="{ width, height }"
  >
    <el-image
      v-if="url"
      :src="url"
      :fit="fit"
      :style="{ width: '100%', height: '100%' }"
    >
      <template #error>
        <div class="s-image__error">
          <el-icon :size="24"><Picture /></el-icon>
        </div>
      </template>
    </el-image>
    <div
      v-else-if="loading"
      class="s-image__loading"
    >
      <el-icon
        class="is-loading"
        :size="24"
      >
        <Loading />
      </el-icon>
    </div>
    <div
      v-else
      class="s-image__error"
    >
      <el-icon :size="24"><Picture /></el-icon>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { Picture, Loading } from '@element-plus/icons-vue'
import { getPresignUrl } from '@/api/sys/file'

const props = withDefaults(
  defineProps<{
    fileId: number
    width?: string
    height?: string
    fit?: 'fill' | 'contain' | 'cover' | 'none' | 'scale-down'
    /** 预签名URL过期时间（分钟），用于自动刷新计算 */
    expireMinutes?: number
  }>(),
  {
    width: '100%',
    height: '100%',
    fit: 'cover',
    expireMinutes: 30,
  }
)

const url = ref<string>('')
const loading = ref(false)
let refreshTimer: ReturnType<typeof setTimeout> | null = null

async function fetchUrl() {
  if (!props.fileId) return
  loading.value = true
  try {
    url.value = await getPresignUrl(props.fileId)
    scheduleRefresh()
  } catch {
    url.value = ''
  } finally {
    loading.value = false
  }
}

function scheduleRefresh() {
  clearTimer()
  // 提前3分钟刷新URL
  const refreshDelay = Math.max((props.expireMinutes - 3) * 60 * 1000, 60000)
  refreshTimer = setTimeout(() => {
    fetchUrl()
  }, refreshDelay)
}

function clearTimer() {
  if (refreshTimer) {
    clearTimeout(refreshTimer)
    refreshTimer = null
  }
}

onMounted(() => {
  fetchUrl()
})

watch(
  () => props.fileId,
  () => {
    url.value = ''
    fetchUrl()
  }
)

onBeforeUnmount(() => {
  clearTimer()
})
</script>

<style scoped>
.s-image {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.s-image__loading,
.s-image__error {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: var(--el-fill-color-light);
  color: var(--el-text-color-placeholder);
}
</style>
