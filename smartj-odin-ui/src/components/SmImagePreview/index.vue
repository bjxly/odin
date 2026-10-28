<template>
  <div class="sm-image-preview">
    <div
      class="sm-image-preview__trigger"
      @click="openPreview"
    >
      <slot />
    </div>
    <Teleport to="body">
      <Transition name="sm-image-preview-fade">
        <div
          v-if="visible"
          class="sm-image-preview__overlay"
          :style="{ zIndex }"
          @click.self="close"
          @wheel.prevent="onWheel"
        >
          <!-- 关闭按钮 -->
          <div
            class="sm-image-preview__close"
            @click="close"
          >
            <el-icon :size="28"><Close /></el-icon>
          </div>

          <!-- 计数器 -->
          <div class="sm-image-preview__counter">{{ currentIndex + 1 }} / {{ urlList.length }}</div>

          <!-- 左箭头 -->
          <div
            v-if="urlList.length > 1"
            class="sm-image-preview__arrow sm-image-preview__arrow--left"
            @click.stop="prev"
          >
            <el-icon :size="32"><ArrowLeft /></el-icon>
          </div>

          <!-- 图片 -->
          <div
            class="sm-image-preview__canvas"
            :style="canvasStyle"
            @mousedown="onMouseDown"
            @mousemove="onMouseMove"
            @mouseup="onMouseUp"
            @mouseleave="onMouseUp"
          >
            <img
              :src="currentUrl"
              :style="imgStyle"
              class="sm-image-preview__img"
              draggable="false"
              @load="resetTransform"
            />
          </div>

          <!-- 右箭头 -->
          <div
            v-if="urlList.length > 1"
            class="sm-image-preview__arrow sm-image-preview__arrow--right"
            @click.stop="next"
          >
            <el-icon :size="32"><ArrowRight /></el-icon>
          </div>

          <!-- 底部工具栏 -->
          <div class="sm-image-preview__toolbar">
            <el-button-group>
              <el-button
                text
                @click="zoomOut"
              >
                <el-icon><ZoomOut /></el-icon>
              </el-button>
              <el-button text>{{ Math.round(scale * 100) }}%</el-button>
              <el-button
                text
                @click="zoomIn"
              >
                <el-icon><ZoomIn /></el-icon>
              </el-button>
              <el-button
                text
                @click="rotateLeft"
              >
                <el-icon><RefreshLeft /></el-icon>
              </el-button>
              <el-button
                text
                @click="rotateRight"
              >
                <el-icon><RefreshRight /></el-icon>
              </el-button>
            </el-button-group>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Close, ArrowLeft, ArrowRight, ZoomOut, ZoomIn, RefreshLeft, RefreshRight } from '@element-plus/icons-vue'

const props = withDefaults(
  defineProps<{
    urlList: string[]
    initialIndex?: number
    zIndex?: number
  }>(),
  {
    initialIndex: 0,
    zIndex: 3000,
  }
)

const visible = ref(false)
const currentIndex = ref(props.initialIndex)
const scale = ref(1)
const rotation = ref(0)
const translateX = ref(0)
const translateY = ref(0)

// Pan state
let isDragging = false
let dragStartX = 0
let dragStartY = 0
let dragTranslateX = 0
let dragTranslateY = 0

const currentUrl = computed(() => props.urlList[currentIndex.value] || '')

const canvasStyle = computed(() => ({
  cursor: scale.value > 1 ? (isDragging ? 'grabbing' : 'grab') : 'default',
}))

const imgStyle = computed(() => ({
  transform: `translate(${translateX.value}px, ${translateY.value}px) scale(${scale.value}) rotate(${rotation.value}deg)`,
  transition: isDragging ? 'none' : 'transform 0.3s',
}))

function openPreview() {
  visible.value = true
  currentIndex.value = props.initialIndex
}

function close() {
  visible.value = false
  resetTransform()
}

function prev() {
  if (currentIndex.value > 0) {
    currentIndex.value--
    resetTransform()
  }
}

function next() {
  if (currentIndex.value < props.urlList.length - 1) {
    currentIndex.value++
    resetTransform()
  }
}

function zoomIn() {
  scale.value = Math.min(scale.value + 0.25, 5)
}

function zoomOut() {
  scale.value = Math.max(scale.value - 0.25, 0.25)
  if (scale.value <= 1) {
    translateX.value = 0
    translateY.value = 0
  }
}

function rotateLeft() {
  rotation.value -= 90
}

function rotateRight() {
  rotation.value += 90
}

function resetTransform() {
  scale.value = 1
  rotation.value = 0
  translateX.value = 0
  translateY.value = 0
}

function onWheel(e: WheelEvent) {
  const delta = e.deltaY > 0 ? -0.25 : 0.25
  scale.value = Math.max(0.25, Math.min(5, scale.value + delta))
  if (scale.value <= 1) {
    translateX.value = 0
    translateY.value = 0
  }
}

function onMouseDown(e: MouseEvent) {
  if (scale.value <= 1) return
  isDragging = true
  dragStartX = e.clientX
  dragStartY = e.clientY
  dragTranslateX = translateX.value
  dragTranslateY = translateY.value
}

function onMouseMove(e: MouseEvent) {
  if (!isDragging) return
  translateX.value = dragTranslateX + (e.clientX - dragStartX)
  translateY.value = dragTranslateY + (e.clientY - dragStartY)
}

function onMouseUp() {
  isDragging = false
}

// Keyboard support
function onKeydown(e: KeyboardEvent) {
  if (!visible.value) return
  switch (e.key) {
    case 'Escape':
      close()
      break
    case 'ArrowLeft':
      prev()
      break
    case 'ArrowRight':
      next()
      break
    case '+':
    case '=':
      zoomIn()
      break
    case '-':
      zoomOut()
      break
  }
}

watch(visible, (val) => {
  if (val) {
    document.addEventListener('keydown', onKeydown)
    document.body.style.overflow = 'hidden'
  } else {
    document.removeEventListener('keydown', onKeydown)
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
.sm-image-preview__trigger {
  display: inline-block;
  cursor: pointer;
}
.sm-image-preview__overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
}
.sm-image-preview__close {
  position: fixed;
  top: 24px;
  right: 40px;
  color: #fff;
  cursor: pointer;
  z-index: 1;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background 0.2s;
}
.sm-image-preview__close:hover {
  background: rgba(255, 255, 255, 0.1);
}
.sm-image-preview__counter {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  color: #fff;
  font-size: 14px;
  z-index: 1;
}
.sm-image-preview__arrow {
  position: fixed;
  top: 50%;
  transform: translateY(-50%);
  color: #fff;
  cursor: pointer;
  z-index: 1;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background 0.2s;
}
.sm-image-preview__arrow:hover {
  background: rgba(255, 255, 255, 0.1);
}
.sm-image-preview__arrow--left {
  left: 16px;
}
.sm-image-preview__arrow--right {
  right: 16px;
}
.sm-image-preview__canvas {
  display: flex;
  align-items: center;
  justify-content: center;
  max-width: 90vw;
  max-height: 80vh;
  overflow: hidden;
}
.sm-image-preview__img {
  max-width: 90vw;
  max-height: 80vh;
  object-fit: contain;
  user-select: none;
}
.sm-image-preview__toolbar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 8px;
  padding: 4px 8px;
}
.sm-image-preview__toolbar .el-button {
  color: #fff;
}
.sm-image-preview__toolbar .el-button:hover {
  background: rgba(255, 255, 255, 0.1);
}

/* Transition */
.sm-image-preview-fade-enter-active,
.sm-image-preview-fade-leave-active {
  transition: opacity 0.3s;
}
.sm-image-preview-fade-enter-from,
.sm-image-preview-fade-leave-to {
  opacity: 0;
}
</style>
