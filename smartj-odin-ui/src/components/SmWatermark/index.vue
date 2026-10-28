<template>
  <Teleport to="body">
    <div
      v-if="enabled && hasContent"
      class="sm-watermark"
      :style="watermarkStyle"
    />
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'

const props = withDefaults(
  defineProps<{
    text?: string | string[]
    fontSize?: number
    opacity?: number
    rotate?: number
    gap?: [number, number]
    zIndex?: number
    enabled?: boolean
  }>(),
  {
    text: () => [],
    fontSize: 14,
    opacity: 0.15,
    rotate: -20,
    gap: () => [200, 150] as [number, number],
    zIndex: 9999,
    enabled: true,
  }
)

const hasContent = computed(() => {
  const t = props.text
  if (!t) return false
  if (Array.isArray(t)) return t.length > 0
  return t.length > 0
})

const textArray = computed(() => {
  if (Array.isArray(props.text)) return props.text
  return props.text ? [props.text] : []
})

function generateDataUrl(): string {
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')!
  const { fontSize, rotate, gap, opacity } = props
  const texts = textArray.value

  // Calculate canvas size based on text content
  const lineHeight = fontSize * 1.5
  ctx.font = `${fontSize}px Arial, "Microsoft YaHei", sans-serif`
  canvas.width = gap[0]
  canvas.height = gap[1]

  ctx.fillStyle = `rgba(180, 180, 180, ${opacity})`
  ctx.font = `${fontSize}px Arial, "Microsoft YaHei", sans-serif`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'

  const centerX = canvas.width / 2
  const centerY = canvas.height / 2
  const totalHeight = texts.length * lineHeight
  const startY = centerY - totalHeight / 2 + lineHeight / 2

  ctx.save()
  ctx.translate(centerX, centerY)
  ctx.rotate((rotate * Math.PI) / 180)

  texts.forEach((text, i) => {
    ctx.fillText(text, 0, startY - centerY + i * lineHeight)
  })

  ctx.restore()

  return canvas.toDataURL()
}

const bgImage = ref('')

function updateImage() {
  if (hasContent.value) {
    bgImage.value = generateDataUrl()
  } else {
    bgImage.value = ''
  }
}

const watermarkStyle = computed(() => ({
  backgroundImage: bgImage.value ? `url(${bgImage.value})` : 'none',
  zIndex: props.zIndex,
}))

watch(() => [props.text, props.fontSize, props.opacity, props.rotate, props.gap], updateImage, { deep: true })

onMounted(updateImage)
</script>

<style scoped>
.sm-watermark {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  background-repeat: repeat;
}
</style>
