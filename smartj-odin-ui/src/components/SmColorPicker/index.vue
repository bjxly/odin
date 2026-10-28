<template>
  <div class="sm-color-picker">
    <el-color-picker
      :model-value="modelValue"
      :predefine="predefine"
      show-alpha
      @change="onColorChange"
    />
    <div
      v-if="historyColors.length > 0 || (predefine && predefine.length > 0)"
      class="sm-color-picker__swatches"
    >
      <div
        v-if="historyColors.length > 0"
        class="sm-color-picker__group"
      >
        <div class="sm-color-picker__label">最近使用</div>
        <div class="sm-color-picker__list">
          <div
            v-for="color in historyColors"
            :key="color"
            class="sm-color-picker__swatch"
            :class="{ 'is-active': modelValue === color }"
            :style="{ backgroundColor: color }"
            :title="color"
            @click="onSwatchClick(color)"
          />
        </div>
      </div>
      <div
        v-if="predefine && predefine.length > 0"
        class="sm-color-picker__group"
      >
        <div class="sm-color-picker__label">预设颜色</div>
        <div class="sm-color-picker__list">
          <div
            v-for="color in predefine"
            :key="color"
            class="sm-color-picker__swatch"
            :class="{ 'is-active': modelValue === color }"
            :style="{ backgroundColor: color }"
            :title="color"
            @click="onSwatchClick(color)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const STORAGE_KEY = 'sm-color-picker-history'
const MAX_HISTORY = 10

const props = withDefaults(
  defineProps<{
    modelValue?: string
    predefine?: string[]
    showHistory?: boolean
  }>(),
  {
    modelValue: '',
    predefine: () => [],
    showHistory: true,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
  (e: 'change', val: string): void
}>()

const historyColors = ref<string[]>([])

onMounted(() => {
  if (props.showHistory) {
    loadHistory()
  }
})

function loadHistory() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      historyColors.value = JSON.parse(raw).slice(0, MAX_HISTORY)
    }
  } catch {
    historyColors.value = []
  }
}

function saveHistory() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(historyColors.value))
  } catch {
    // storage full or unavailable
  }
}

function addToHistory(color: string) {
  if (!color || !props.showHistory) return
  const idx = historyColors.value.indexOf(color)
  if (idx !== 0) {
    if (idx > 0) historyColors.value.splice(idx, 1)
    historyColors.value.unshift(color)
    if (historyColors.value.length > MAX_HISTORY) {
      historyColors.value = historyColors.value.slice(0, MAX_HISTORY)
    }
    saveHistory()
  }
}

function onColorChange(val: string | null) {
  const color = val || ''
  emit('update:modelValue', color)
  emit('change', color)
  if (color) addToHistory(color)
}

function onSwatchClick(color: string) {
  emit('update:modelValue', color)
  emit('change', color)
  addToHistory(color)
}
</script>

<style scoped>
.sm-color-picker {
  display: inline-flex;
  flex-direction: column;
  gap: var(--sm-spacing-sm);
}

.sm-color-picker__swatches {
  display: flex;
  flex-direction: column;
  gap: var(--sm-spacing-sm);
  padding: var(--sm-spacing-sm);
  background: var(--sm-bg-card);
  border: 1px solid var(--sm-border-light);
  border-radius: var(--sm-radius-md);
  min-width: 200px;
}

.sm-color-picker__group {
  display: flex;
  flex-direction: column;
  gap: var(--sm-spacing-xs);
}

.sm-color-picker__label {
  font-size: var(--sm-font-size-xs);
  color: var(--sm-text-secondary);
  line-height: 1;
}

.sm-color-picker__list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--sm-spacing-xs);
}

.sm-color-picker__swatch {
  width: 20px;
  height: 20px;
  border-radius: var(--sm-radius-sm);
  border: 2px solid transparent;
  cursor: pointer;
  transition:
    border-color 0.2s,
    transform 0.15s;
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.1);
}

.sm-color-picker__swatch:hover {
  transform: scale(1.15);
}

.sm-color-picker__swatch.is-active {
  border-color: var(--sm-color-primary);
}
</style>
