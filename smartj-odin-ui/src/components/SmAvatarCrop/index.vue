<template>
  <div class="sm-avatar-crop">
    <div
      class="sm-avatar-crop__preview"
      :class="{ 'sm-avatar-crop__preview--circle': shape === 'circle' }"
      :style="{ width: size + 'px', height: size + 'px' }"
      @click="openFilePicker"
    >
      <img
        v-if="modelValue"
        :src="modelValue"
        class="sm-avatar-crop__image"
      />
      <div
        v-else
        class="sm-avatar-crop__placeholder"
      >
        <el-icon :size="size * 0.35"><Plus /></el-icon>
      </div>
      <div class="sm-avatar-crop__overlay">
        <el-icon :size="20"><UploadFilled /></el-icon>
      </div>
    </div>

    <input
      ref="fileInputRef"
      type="file"
      accept="image/*"
      style="display: none"
      @change="onFileChange"
    />

    <el-dialog
      v-model="dialogVisible"
      :title="$t('cmp.cropAvatar')"
      width="480px"
      :close-on-click-modal="false"
      destroy-on-close
      @opened="initCanvas"
    >
      <div class="sm-avatar-crop__canvas-wrapper">
        <canvas
          ref="canvasRef"
          class="sm-avatar-crop__canvas"
          @mousedown="onCanvasMouseDown"
          @mousemove="onCanvasMouseMove"
          @mouseup="onCanvasMouseUp"
          @mouseleave="onCanvasMouseUp"
        />
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="onConfirm"
        >
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { Plus, UploadFilled } from '@element-plus/icons-vue'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    shape?: 'circle' | 'square'
    size?: number
  }>(),
  {
    modelValue: '',
    shape: 'circle',
    size: 128,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
}>()

const fileInputRef = ref<HTMLInputElement>()
const canvasRef = ref<HTMLCanvasElement>()
const dialogVisible = ref(false)

let img = new Image()
let imgLoaded = false
let cropX = 0
let cropY = 0
let cropSize = 0
let isDragging = false
let dragStartX = 0
let dragStartY = 0
let imgOffsetX = 0
let imgOffsetY = 0
let imgDrawW = 0
let imgDrawH = 0
let scale = 1

const CANVAS_W = 420
const CANVAS_H = 320

function openFilePicker() {
  fileInputRef.value?.click()
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (ev) => {
    img = new Image()
    img.onload = () => {
      imgLoaded = true
      dialogVisible.value = true
    }
    img.src = ev.target?.result as string
  }
  reader.readAsDataURL(file)
  input.value = ''
}

function initCanvas() {
  nextTick(() => {
    const canvas = canvasRef.value
    if (!canvas || !imgLoaded) return
    canvas.width = CANVAS_W
    canvas.height = CANVAS_H

    scale = Math.min(CANVAS_W / img.width, CANVAS_H / img.height)
    imgDrawW = img.width * scale
    imgDrawH = img.height * scale
    imgOffsetX = (CANVAS_W - imgDrawW) / 2
    imgOffsetY = (CANVAS_H - imgDrawH) / 2

    cropSize = Math.min(imgDrawW, imgDrawH) * 0.8
    cropX = (CANVAS_W - cropSize) / 2
    cropY = (CANVAS_H - cropSize) / 2

    drawCanvas()
  })
}

function drawCanvas() {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  ctx.clearRect(0, 0, CANVAS_W, CANVAS_H)

  // draw dimmed image
  ctx.save()
  ctx.globalAlpha = 0.4
  ctx.drawImage(img, imgOffsetX, imgOffsetY, imgDrawW, imgDrawH)
  ctx.restore()

  // clip crop area and draw full image
  ctx.save()
  ctx.beginPath()
  if (props.shape === 'circle') {
    ctx.arc(cropX + cropSize / 2, cropY + cropSize / 2, cropSize / 2, 0, Math.PI * 2)
  } else {
    ctx.rect(cropX, cropY, cropSize, cropSize)
  }
  ctx.clip()
  ctx.drawImage(img, imgOffsetX, imgOffsetY, imgDrawW, imgDrawH)
  ctx.restore()

  // draw crop border
  ctx.save()
  ctx.strokeStyle = '#409eff'
  ctx.lineWidth = 2
  if (props.shape === 'circle') {
    ctx.beginPath()
    ctx.arc(cropX + cropSize / 2, cropY + cropSize / 2, cropSize / 2, 0, Math.PI * 2)
    ctx.stroke()
  } else {
    ctx.strokeRect(cropX, cropY, cropSize, cropSize)
  }
  ctx.restore()

  // draw grid lines
  ctx.save()
  ctx.strokeStyle = 'rgba(255,255,255,0.3)'
  ctx.lineWidth = 1
  const step = cropSize / 3
  for (let i = 1; i < 3; i++) {
    ctx.beginPath()
    ctx.moveTo(cropX + step * i, cropY)
    ctx.lineTo(cropX + step * i, cropY + cropSize)
    ctx.stroke()
    ctx.beginPath()
    ctx.moveTo(cropX, cropY + step * i)
    ctx.lineTo(cropX + cropSize, cropY + step * i)
    ctx.stroke()
  }
  ctx.restore()
}

function onCanvasMouseDown(e: MouseEvent) {
  const rect = canvasRef.value?.getBoundingClientRect()
  if (!rect) return
  const mx = e.clientX - rect.left
  const my = e.clientY - rect.top
  if (mx >= cropX && mx <= cropX + cropSize && my >= cropY && my <= cropY + cropSize) {
    isDragging = true
    dragStartX = mx - cropX
    dragStartY = my - cropY
  }
}

function onCanvasMouseMove(e: MouseEvent) {
  if (!isDragging) return
  const rect = canvasRef.value?.getBoundingClientRect()
  if (!rect) return
  const mx = e.clientX - rect.left
  const my = e.clientY - rect.top
  cropX = Math.max(imgOffsetX, Math.min(mx - dragStartX, imgOffsetX + imgDrawW - cropSize))
  cropY = Math.max(imgOffsetY, Math.min(my - dragStartY, imgOffsetY + imgDrawH - cropSize))
  drawCanvas()
}

function onCanvasMouseUp() {
  isDragging = false
}

function onConfirm() {
  // compute source crop rect on original image
  const sx = (cropX - imgOffsetX) / scale
  const sy = (cropY - imgOffsetY) / scale
  const sw = cropSize / scale
  const sh = cropSize / scale

  const outCanvas = document.createElement('canvas')
  outCanvas.width = props.size
  outCanvas.height = props.size
  const ctx = outCanvas.getContext('2d')
  if (!ctx) return

  if (props.shape === 'circle') {
    ctx.beginPath()
    ctx.arc(props.size / 2, props.size / 2, props.size / 2, 0, Math.PI * 2)
    ctx.clip()
  }

  ctx.drawImage(img, sx, sy, sw, sh, 0, 0, props.size, props.size)
  const base64 = outCanvas.toDataURL('image/png')
  emit('update:modelValue', base64)
  dialogVisible.value = false
}
</script>

<style scoped>
.sm-avatar-crop {
  display: inline-block;
}

.sm-avatar-crop__preview {
  position: relative;
  overflow: hidden;
  border: 2px dashed var(--sm-border-base);
  border-radius: var(--sm-radius-md);
  cursor: pointer;
  background: var(--sm-bg-header);
  transition: border-color 0.2s;
}

.sm-avatar-crop__preview--circle {
  border-radius: 50%;
}

.sm-avatar-crop__preview:hover {
  border-color: var(--sm-color-primary);
}

.sm-avatar-crop__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.sm-avatar-crop__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--sm-text-placeholder);
}

.sm-avatar-crop__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  color: #fff;
  opacity: 0;
  transition: opacity 0.2s;
}

.sm-avatar-crop__preview:hover .sm-avatar-crop__overlay {
  opacity: 1;
}

.sm-avatar-crop__canvas-wrapper {
  display: flex;
  justify-content: center;
}

.sm-avatar-crop__canvas {
  display: block;
  border: 1px solid var(--sm-border-light);
  border-radius: var(--sm-radius-md);
  cursor: move;
}
</style>
