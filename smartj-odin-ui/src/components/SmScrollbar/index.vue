<template>
  <div
    class="sm-scrollbar"
    :style="wrapperStyle"
  >
    <el-scrollbar
      :height="native ? undefined : height"
      :max-height="native ? undefined : maxHeight"
      :always="always"
      :native="native"
    >
      <slot />
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    height?: number | string
    maxHeight?: number | string
    native?: boolean
    always?: boolean
  }>(),
  {
    native: false,
    always: false,
  }
)

const wrapperStyle = computed(() => {
  const style: Record<string, string> = {}
  if (props.native) {
    if (props.height) style.height = typeof props.height === 'number' ? `${props.height}px` : props.height
    if (props.maxHeight) style.maxHeight = typeof props.maxHeight === 'number' ? `${props.maxHeight}px` : props.maxHeight
    style.overflow = 'auto'
  }
  return style
})
</script>

<style scoped>
.sm-scrollbar {
  width: 100%;
}
</style>
