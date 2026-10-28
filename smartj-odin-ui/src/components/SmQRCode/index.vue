<template>
  <div
    class="sm-qrcode"
    :style="{ width: size + 'px', height: size + 'px' }"
  >
    <canvas
      v-show="!errored"
      ref="canvasRef"
      class="sm-qrcode__canvas"
    />
    <div
      v-if="errored"
      class="sm-qrcode__error"
    >
      <svg
        viewBox="0 0 24 24"
        width="32"
        height="32"
        fill="currentColor"
      >
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
      </svg>
      <span>{{ $t('cmp.qrcodeFailed') }}</span>
    </div>
    <!-- Logo 叠加 -->
    <img
      v-if="logo && !errored"
      :src="logo"
      class="sm-qrcode__logo"
      :style="{ width: logoSize + 'px', height: logoSize + 'px' }"
      @error="onLogoError"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue'

const props = withDefaults(
  defineProps<{
    value?: string
    size?: number
    fgColor?: string
    bgColor?: string
    level?: string
    logo?: string
    logoSize?: number
    margin?: number
  }>(),
  {
    value: '',
    size: 200,
    fgColor: '#000000',
    bgColor: '#ffffff',
    level: 'M',
    logoSize: 44,
    margin: 2,
  }
)

const emit = defineEmits<{
  (e: 'error', err: Error): void
  (e: 'rendered'): void
}>()

const canvasRef = ref<HTMLCanvasElement>()
const errored = ref(false)
const logoVisible = ref(true)

const QR_LEVELS: Record<string, 'L' | 'M' | 'Q' | 'H'> = {
  L: 'L',
  M: 'M',
  Q: 'Q',
  H: 'H',
}

async function generate() {
  if (!props.value || !canvasRef.value) return

  try {
    // 动态导入 qrcode，避免 SSR 问题
    // @ts-expect-error - qrcode has no type declarations
    const QRCode = await import('qrcode')
    errored.value = false

    const canvas = canvasRef.value
    canvas.width = props.size
    canvas.height = props.size
    canvas.style.width = props.size + 'px'
    canvas.style.height = props.size + 'px'

    await QRCode.toCanvas(canvas, props.value, {
      width: props.size,
      margin: props.margin,
      color: {
        dark: props.fgColor,
        light: props.bgColor,
      },
      errorCorrectionLevel: QR_LEVELS[props.level] || 'M',
    })

    emit('rendered')
  } catch (err) {
    errored.value = true
    emit('error', err instanceof Error ? err : new Error(String(err)))
  }
}

function onLogoError() {
  logoVisible.value = false
}

watch(
  () => [props.value, props.size, props.fgColor, props.bgColor, props.level, props.margin],
  () => {
    nextTick(generate)
  }
)

onMounted(() => {
  nextTick(generate)
})

defineExpose({
  /** 获取 canvas（可用于 toDataURL 下载等） */
  getCanvas: () => canvasRef.value,
  /** 获取 data URL */
  toDataURL: (type?: string, quality?: number) => canvasRef.value?.toDataURL(type, quality),
  /** 重新生成 */
  regenerate: generate,
})
</script>

<style scoped>
.sm-qrcode {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
}
.sm-qrcode__canvas {
  display: block;
  width: 100%;
  height: 100%;
}
.sm-qrcode__logo {
  position: absolute;
  border-radius: 6px;
  box-shadow: 0 0 0 4px #fff;
  background: #fff;
  object-fit: contain;
}
.sm-qrcode__error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #c0c4cc;
  font-size: 13px;
}
</style>
