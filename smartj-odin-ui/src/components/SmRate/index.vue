<template>
  <div class="sm-rate" :class="{ 'sm-rate--readonly': readonly }">
    <span
      v-for="i in count"
      :key="i"
      class="sm-rate__item"
      :class="{
        'sm-rate__item--active': i <= innerValue,
        'sm-rate__item--half': allowHalf && i - 0.5 === innerValue,
      }"
      :style="itemStyle"
      @click="handleClick(i, $event)"
      @mousemove="handleMove(i, $event)"
      @mouseleave="handleLeave"
    >
      <template v-if="slots.default">
        <span class="sm-rate__char sm-rate__char--empty">
          <slot name="default" :index="i" :active="false" />
        </span>
        <span
          class="sm-rate__char sm-rate__char--active"
          :style="{ width: activeWidth(i) }"
        >
          <slot name="default" :index="i" :active="true" />
        </span>
      </template>
      <template v-else>
        <span class="sm-rate__char sm-rate__char--empty">
          <slot name="empty-icon">
            <svg viewBox="0 0 24 24" width="1em" height="1em" fill="currentColor">
              <path d="M12 2l3.09 6.26L22 9.27l-5 4.87L18.18 22 12 18.56 5.82 22 7 14.14 2 9.27l6.91-1.01L12 2z" />
            </svg>
          </slot>
        </span>
        <span
          class="sm-rate__char sm-rate__char--active"
          :style="{ width: activeWidth(i) }"
        >
          <slot name="active-icon">
            <svg viewBox="0 0 24 24" width="1em" height="1em" fill="currentColor">
              <path d="M12 2l3.09 6.26L22 9.27l-5 4.87L18.18 22 12 18.56 5.82 22 7 14.14 2 9.27l6.91-1.01L12 2z" />
            </svg>
          </slot>
        </span>
      </template>
    </span>
    <span v-if="showValue" class="sm-rate__text" :style="{ color: textColor || activeColor }">
      {{ innerValue.toFixed(allowHalf ? 1 : 0) }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, useSlots } from 'vue'

const slots = useSlots()

const props = withDefaults(
  defineProps<{
    modelValue?: number
    count?: number
    allowHalf?: boolean
    size?: number | string
    activeColor?: string
    voidColor?: string
    readonly?: boolean
    clearable?: boolean
    showValue?: boolean
    textColor?: string
  }>(),
  {
    modelValue: 0,
    count: 5,
    allowHalf: false,
    size: 20,
    activeColor: '#f7ba2a',
    voidColor: '#c0c4cc',
    readonly: false,
    clearable: false,
    showValue: false,
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
  (e: 'change', value: number): void
}>()

const innerValue = ref(props.modelValue)
const hoverValue = ref(0)

watch(
  () => props.modelValue,
  (val) => {
    innerValue.value = val
  },
)

const itemStyle = computed(() => ({
  fontSize: typeof props.size === 'number' ? `${props.size}px` : props.size,
  color: props.voidColor,
  '--sm-rate-active-color': props.activeColor,
  cursor: props.readonly ? 'default' : 'pointer',
}))

function activeWidth(index: number): string {
  if (!props.allowHalf) return '100%'
  if (index <= Math.floor(innerValue.value)) return '100%'
  if (index - 0.5 === innerValue.value) return '50%'
  return '0%'
}

function calcValue(index: number, event: MouseEvent): number {
  if (!props.allowHalf) return index
  const el = event.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  const x = event.clientX - rect.left
  return x < rect.width / 2 ? index - 0.5 : index
}

function handleClick(index: number, event: MouseEvent) {
  if (props.readonly) return
  const val = calcValue(index, event)
  if (props.clearable && val === innerValue.value) {
    innerValue.value = 0
    emit('update:modelValue', 0)
    emit('change', 0)
  } else {
    innerValue.value = val
    emit('update:modelValue', val)
    emit('change', val)
  }
}

function handleMove(index: number, event: MouseEvent) {
  if (props.readonly) return
  hoverValue.value = calcValue(index, event)
}

function handleLeave() {
  hoverValue.value = 0
}
</script>

<style scoped>
.sm-rate {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  vertical-align: middle;
}
.sm-rate--readonly {
  pointer-events: none;
}
.sm-rate__item {
  position: relative;
  display: inline-flex;
  transition: transform 0.15s;
}
.sm-rate__item:hover {
  transform: scale(1.15);
}
.sm-rate__item--active {
  color: var(--sm-rate-active-color, #f7ba2a);
}
.sm-rate__char {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s;
}
.sm-rate__char--active {
  position: absolute;
  left: 0;
  top: 0;
  overflow: hidden;
  white-space: nowrap;
  color: var(--sm-rate-active-color, #f7ba2a);
}
.sm-rate__char--empty {
  color: inherit;
}
.sm-rate__text {
  margin-left: 8px;
  font-size: 14px;
  font-variant-numeric: tabular-nums;
}
</style>
