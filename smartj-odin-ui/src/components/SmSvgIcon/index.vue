<template>
  <i
    v-if="isSvgIcon"
    class="sm-svg-icon"
    :style="iconStyle"
    v-html="svgContent"
  />
  <i
    v-else-if="isFontClassIcon"
    class="sm-font-icon iconfont"
    :class="fontIconClass"
    :style="fontIconStyle"
  />
  <el-icon
    v-else
    :size="size"
    :color="color"
  >
    <component :is="name" />
  </el-icon>
</template>

<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue'
import { useIconfontStyle } from '@/composables/useIconfontStyle'

const props = withDefaults(
  defineProps<{
    name: string
    size?: number | string
    color?: string
    iconfontCssUrl?: string
  }>(),
  {
    size: 16,
    iconfontCssUrl: () => import.meta.env.VITE_ICONFONT_CSS_URL || '/iconfont/iconfont.css',
  }
)

const svgCache = new Map<string, string>()
const isSvgIcon = computed(() => props.name.startsWith('svg-'))
const svgContent = ref('')

const iconStyle = computed(() => ({
  width: typeof props.size === 'number' ? `${props.size}px` : props.size,
  height: typeof props.size === 'number' ? `${props.size}px` : props.size,
  color: props.color || 'inherit',
  display: 'inline-flex',
  alignItems: 'center',
  justifyContent: 'center',
}))

const isFontClassIcon = computed(() => props.name.startsWith('ifc-'))
const fontIconClass = computed(() => `icon-${props.name.replace(/^ifc-/, '')}`)
const fontIconStyle = computed(() => ({
  fontSize: typeof props.size === 'number' ? `${props.size}px` : props.size,
  color: props.color || 'inherit',
  display: 'inline-flex',
  alignItems: 'center',
  justifyContent: 'center',
  fontStyle: 'normal',
}))

watchEffect(() => {
  if (isFontClassIcon.value && props.iconfontCssUrl) {
    useIconfontStyle(props.iconfontCssUrl)
  }
})

watch(
  () => props.name,
  async (name) => {
    if (!name.startsWith('svg-')) return
    const key = name.replace('svg-', '')
    if (svgCache.has(key)) {
      svgContent.value = svgCache.get(key)!
      return
    }
    try {
      const resp = await fetch(`/src/assets/icons/${key}.svg`)
      if (resp.ok) {
        const text = await resp.text()
        svgCache.set(key, text)
        svgContent.value = text
      }
    } catch {
      /* ignore */
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.sm-svg-icon :deep(svg) {
  width: 100%;
  height: 100%;
  fill: currentColor;
}

.sm-font-icon {
  font-style: normal;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
