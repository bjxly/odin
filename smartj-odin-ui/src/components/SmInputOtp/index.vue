<template>
  <div class="sm-input-otp" :class="{ 'sm-input-otp--disabled': disabled }">
    <div class="sm-input-otp__grid" :style="{ gap: gap + 'px' }">
      <input
        v-for="(_, idx) in length"
        :key="idx"
        :ref="(el) => setInputRef(el as HTMLInputElement, idx)"
        type="text"
        :maxlength="1"
        :disabled="disabled"
        :inputmode="numericOnly ? 'numeric' : 'text'"
        class="sm-input-otp__input"
        :class="{
          'sm-input-otp__input--active': currentIndex === idx,
          'sm-input-otp__input--filled': !!chars[idx],
          'sm-input-otp__input--error': !!errorMsg,
        }"
        :style="inputStyle"
        :value="chars[idx] || ''"
        @input="handleInput(idx, $event)"
        @keydown="handleKeydown(idx, $event)"
        @focus="currentIndex = idx"
        @paste="handlePaste"
      >
    </div>
    <p v-if="errorMsg" class="sm-input-otp__error">{{ errorMsg }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    length?: number
    size?: string
    gap?: number
    disabled?: boolean
    numericOnly?: boolean
    mask?: boolean
  }>(),
  {
    modelValue: '',
    length: 6,
    size: '40px',
    gap: 8,
    disabled: false,
    numericOnly: true,
    mask: false,
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'complete', value: string): void
  (e: 'change', value: string): void
}>()

const chars = ref<string[]>(Array(props.length).fill(''))
const inputRefs = ref<(HTMLInputElement | null)[]>([])
const currentIndex = ref(0)
const errorMsg = ref('')

function setInputRef(el: HTMLInputElement | null, idx: number) {
  inputRefs.value[idx] = el
}

const inputStyle = computed(() => ({
  width: props.size,
  height: props.size,
  fontSize: `calc(${props.size} * 0.42)`,
}))

watch(
  () => props.modelValue,
  (val) => {
    const arr = (val || '').split('')
    chars.value = Array.from({ length: props.length }, (_, i) => arr[i] || '')
  },
)

function syncValue() {
  const val = chars.value.join('')
  emit('update:modelValue', val)
  emit('change', val)
  if (val.length === props.length) {
    emit('complete', val)
  }
  errorMsg.value = ''
}

function handleInput(idx: number, event: Event) {
  const target = event.target as HTMLInputElement
  const raw = target.value

  // 过滤非法字符
  let char = ''
  if (props.numericOnly) {
    const match = raw.match(/\d/)
    char = match ? match[0] : ''
  } else {
    char = raw.charAt(raw.length - 1) || ''
  }

  // 处理粘贴多个字符
  if (raw.length > 1) {
    fillFromString(raw)
    return
  }

  if (char) {
    chars.value[idx] = char
    syncValue()
    if (idx < props.length - 1) {
      focusInput(idx + 1)
    } else {
      target.blur()
    }
  } else {
    chars.value[idx] = ''
    syncValue()
  }
}

function handleKeydown(idx: number, event: KeyboardEvent) {
  if (event.key === 'Backspace' || event.key === 'Delete') {
    event.preventDefault()
    if (chars.value[idx]) {
      chars.value[idx] = ''
      syncValue()
    } else if (idx > 0) {
      chars.value[idx - 1] = ''
      syncValue()
      focusInput(idx - 1)
    }
  } else if (event.key === 'ArrowLeft' && idx > 0) {
    event.preventDefault()
    focusInput(idx - 1)
  } else if (event.key === 'ArrowRight' && idx < props.length - 1) {
    event.preventDefault()
    focusInput(idx + 1)
  }
}

function handlePaste(event: ClipboardEvent) {
  event.preventDefault()
  const data = event.clipboardData?.getData('text') || ''
  fillFromString(data)
}

function fillFromString(str: string) {
  const filtered = props.numericOnly ? str.replace(/\D/g, '') : str.replace(/\s/g, '')
  const arr = filtered.split('')
  for (let i = 0; i < props.length; i++) {
    chars.value[i] = arr[i] || ''
  }
  syncValue()
  const nextFocus = Math.min(arr.length, props.length - 1)
  focusInput(nextFocus)
  if (arr.length >= props.length) {
    inputRefs.value[props.length - 1]?.blur()
  }
}

function focusInput(idx: number) {
  nextTick(() => {
    inputRefs.value[idx]?.focus()
    inputRefs.value[idx]?.select()
  })
}

defineExpose({
  focus: () => focusInput(currentIndex.value),
  clear: () => {
    chars.value = Array(props.length).fill('')
    syncValue()
    focusInput(0)
  },
  setError: (msg: string) => {
    errorMsg.value = msg
  },
})
</script>

<style scoped>
.sm-input-otp {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
}
.sm-input-otp--disabled {
  opacity: 0.5;
  pointer-events: none;
}
.sm-input-otp__grid {
  display: flex;
}
.sm-input-otp__input {
  border: 1.5px solid var(--el-border-color, #dcdfe6);
  border-radius: 8px;
  text-align: center;
  outline: none;
  background: var(--el-bg-color, #fff);
  color: var(--el-text-color-primary, #303133);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  transition: all 0.2s;
  padding: 0;
  box-sizing: border-box;
  caret-color: var(--el-color-primary, #409eff);
}
.sm-input-otp__input:focus {
  border-color: var(--el-color-primary, #409eff);
  box-shadow: 0 0 0 2px var(--el-color-primary-light-8, #d9ecff);
}
.sm-input-otp__input--active {
  border-color: var(--el-color-primary, #409eff);
  box-shadow: 0 0 0 2px var(--el-color-primary-light-8, #d9ecff);
}
.sm-input-otp__input--filled {
  border-color: var(--el-color-primary-light-5, #a0cfff);
  background: var(--el-color-primary-light-9, #ecf5ff);
}
.sm-input-otp__input--error {
  border-color: var(--el-color-danger, #f56c6c);
  box-shadow: 0 0 0 2px var(--el-color-danger-light-8, #fde2e2);
}
.sm-input-otp__input:disabled {
  background: var(--el-fill-color-light, #f5f7fa);
  cursor: not-allowed;
}
.sm-input-otp__error {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--el-color-danger, #f56c6c);
}
</style>
