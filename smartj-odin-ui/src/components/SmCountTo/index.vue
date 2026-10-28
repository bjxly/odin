<template>
  <span class="sm-count-to">{{ displayValue }}</span>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    endVal: number
    startVal?: number
    duration?: number
    decimals?: number
    prefix?: string
    suffix?: string
    separator?: string
    autoplay?: boolean
  }>(),
  {
    startVal: 0,
    duration: 2000,
    decimals: 0,
    prefix: '',
    suffix: '',
    separator: ',',
    autoplay: true,
  }
)

const emit = defineEmits<{
  (e: 'mounted'): void
  (e: 'start'): void
  (e: 'pause'): void
  (e: 'resume'): void
}>()

const displayValue = ref(formatValue(props.startVal))
let rafId: number | null = null
let startTime = 0
let paused = false
let remaining = 0
let currentValue = 0

function formatValue(val: number): string {
  const fixed = val.toFixed(props.decimals)
  const [intPart, decPart] = fixed.split('.')
  const formatted = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, props.separator)
  return `${props.prefix}${formatted}${decPart ? '.' + decPart : ''}${props.suffix}`
}

function easeOutExpo(t: number): number {
  return t === 1 ? 1 : 1 - Math.pow(2, -10 * t)
}

function animate(timestamp: number) {
  if (!startTime) startTime = timestamp
  const elapsed = timestamp - startTime + (props.duration - remaining)
  const progress = Math.min(elapsed / props.duration, 1)
  const eased = easeOutExpo(progress)
  currentValue = props.startVal + (props.endVal - props.startVal) * eased
  displayValue.value = formatValue(currentValue)

  if (progress < 1) {
    rafId = requestAnimationFrame(animate)
  } else {
    displayValue.value = formatValue(props.endVal)
  }
}

function start() {
  if (props.endVal === props.startVal) {
    displayValue.value = formatValue(props.endVal)
    return
  }
  paused = false
  startTime = 0
  remaining = props.duration
  emit('start')
  rafId = requestAnimationFrame(animate)
}

function pause() {
  if (paused || !rafId) return
  cancelAnimationFrame(rafId)
  rafId = null
  paused = true
  remaining = props.duration - (performance.now() - startTime)
  emit('pause')
}

function resume() {
  if (!paused) return
  start()
  emit('resume')
}

function stop() {
  if (rafId) {
    cancelAnimationFrame(rafId)
    rafId = null
  }
}

watch(
  () => props.endVal,
  () => {
    if (props.autoplay) start()
  }
)

onMounted(() => {
  emit('mounted')
  if (props.autoplay) start()
})

onUnmounted(() => {
  stop()
})

defineExpose({ start, pause, resume, stop })
</script>

<style scoped>
.sm-count-to {
  font-variant-numeric: tabular-nums;
}
</style>
