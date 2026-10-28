<template>
  <div
    ref="containerRef"
    class="sm-video-player"
    @mouseenter="showControls = true"
    @mouseleave="onContainerMouseLeave"
  >
    <video
      ref="videoRef"
      class="sm-video-player__video"
      :src="src"
      :poster="poster"
      :autoplay="autoplay"
      :muted="muted"
      preload="metadata"
      @loadedmetadata="onLoadedMetadata"
      @timeupdate="onTimeUpdate"
      @ended="onEnded"
      @click="togglePlay"
    />

    <div
      v-if="!isPlaying && !showControls"
      class="sm-video-player__play-overlay"
      @click="togglePlay"
    >
      <el-icon :size="56"><VideoPlay /></el-icon>
    </div>

    <transition name="sm-video-fade">
      <div
        v-show="showControls"
        class="sm-video-player__controls"
      >
        <el-button
          class="sm-video-player__btn"
          text
          @click="togglePlay"
        >
          <el-icon :size="18">
            <VideoPlay v-if="!isPlaying" />
            <VideoPause v-else />
          </el-icon>
        </el-button>

        <el-slider
          v-model="progress"
          class="sm-video-player__progress"
          :show-tooltip="true"
          :format-tooltip="formatTime"
          :max="100"
          :step="0.1"
          @input="onSeek"
        />

        <span class="sm-video-player__time">{{ formatTime(currentTime) }} / {{ formatTime(duration) }}</span>

        <el-popover
          trigger="click"
          :width="140"
          placement="top"
        >
          <template #reference>
            <el-button
              class="sm-video-player__btn"
              text
            >
              <el-icon :size="18"><Microphone /></el-icon>
            </el-button>
          </template>
          <div class="sm-video-player__volume">
            <el-slider
              v-model="volume"
              vertical
              height="100px"
              :max="100"
              @input="onVolumeChange"
            />
          </div>
        </el-popover>

        <el-button
          class="sm-video-player__btn"
          text
          @click="toggleFullscreen"
        >
          <el-icon :size="18"><FullScreen /></el-icon>
        </el-button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue'
import { VideoPlay, VideoPause, Microphone, FullScreen } from '@element-plus/icons-vue'

const props = withDefaults(
  defineProps<{
    src?: string
    poster?: string
    autoplay?: boolean
    muted?: boolean
  }>(),
  {
    src: '',
    poster: '',
    autoplay: false,
    muted: false,
  }
)

const containerRef = ref<HTMLDivElement>()
const videoRef = ref<HTMLVideoElement>()
const isPlaying = ref(false)
const showControls = ref(true)
const progress = ref(0)
const currentTime = ref(0)
const duration = ref(0)
const volume = ref(props.muted ? 0 : 80)

let controlsTimer: ReturnType<typeof setTimeout> | null = null

watch(
  () => props.src,
  () => {
    isPlaying.value = false
    progress.value = 0
    currentTime.value = 0
    duration.value = 0
  }
)

onBeforeUnmount(() => {
  if (controlsTimer) clearTimeout(controlsTimer)
})

function togglePlay() {
  const video = videoRef.value
  if (!video) return
  if (video.paused) {
    video.play()
    isPlaying.value = true
    startControlsTimer()
  } else {
    video.pause()
    isPlaying.value = false
    showControls.value = true
  }
}

function onLoadedMetadata() {
  const video = videoRef.value
  if (!video) return
  duration.value = video.duration
  video.volume = volume.value / 100
}

function onTimeUpdate() {
  const video = videoRef.value
  if (!video) return
  currentTime.value = video.currentTime
  if (duration.value > 0) {
    progress.value = (video.currentTime / duration.value) * 100
  }
}

function onEnded() {
  isPlaying.value = false
  showControls.value = true
}

function onSeek(val: number | string) {
  const video = videoRef.value
  if (!video || !duration.value) return
  const pct = Number(val) / 100
  video.currentTime = pct * duration.value
}

function onVolumeChange(val: number | string) {
  const video = videoRef.value
  if (!video) return
  const v = Number(val) / 100
  video.volume = v
  video.muted = v === 0
}

function toggleFullscreen() {
  const el = containerRef.value
  if (!el) return
  if (document.fullscreenElement) {
    document.exitFullscreen()
  } else {
    el.requestFullscreen()
  }
}

function formatTime(seconds: number): string {
  if (!seconds || isNaN(seconds)) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

function startControlsTimer() {
  if (controlsTimer) clearTimeout(controlsTimer)
  controlsTimer = setTimeout(() => {
    if (isPlaying.value) {
      showControls.value = false
    }
  }, 3000)
}

function onContainerMouseLeave() {
  if (isPlaying.value) {
    startControlsTimer()
  } else {
    showControls.value = true
  }
}
</script>

<style scoped>
.sm-video-player {
  position: relative;
  display: inline-block;
  background: #000;
  border-radius: var(--sm-radius-lg);
  overflow: hidden;
  line-height: 0;
}

.sm-video-player__video {
  display: block;
  width: 100%;
  cursor: pointer;
}

.sm-video-player__play-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.3);
  cursor: pointer;
  color: rgba(255, 255, 255, 0.85);
  transition: background 0.2s;
}

.sm-video-player__play-overlay:hover {
  background: rgba(0, 0, 0, 0.45);
}

.sm-video-player__controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  gap: var(--sm-spacing-xs);
  padding: var(--sm-spacing-sm) var(--sm-spacing-md);
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.7));
}

.sm-video-player__btn {
  color: #fff !important;
  padding: 4px !important;
  min-height: auto !important;
}

.sm-video-player__btn:hover {
  color: var(--sm-color-primary) !important;
}

.sm-video-player__progress {
  flex: 1;
  --el-slider-runway-bg-color: rgba(255, 255, 255, 0.2);
  --el-slider-main-bg-color: var(--sm-color-primary);
}

.sm-video-player__time {
  font-size: var(--sm-font-size-xs);
  color: rgba(255, 255, 255, 0.85);
  white-space: nowrap;
  user-select: none;
  line-height: 1;
}

.sm-video-player__volume {
  display: flex;
  justify-content: center;
  padding: var(--sm-spacing-xs) 0;
}

/* fade transition */
.sm-video-fade-enter-active,
.sm-video-fade-leave-active {
  transition: opacity 0.25s ease;
}

.sm-video-fade-enter-from,
.sm-video-fade-leave-to {
  opacity: 0;
}

:deep(.sm-video-player__progress .el-slider__bar) {
  background-color: var(--sm-color-primary);
}

:deep(.sm-video-player__progress .el-slider__button) {
  border-color: var(--sm-color-primary);
  width: 12px;
  height: 12px;
}
</style>
