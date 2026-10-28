<template>
  <div class="crontab-field">
    <el-radio-group
      v-model="mode"
      @change="onModeChange"
    >
      <el-radio value="every">{{ $t('cmp.cronEvery') }}{{ unit }}</el-radio>
      <el-radio value="range">{{ $t('cmp.cronRange') }}</el-radio>
      <el-radio value="interval">{{ $t('cmp.cronInterval') }}</el-radio>
      <el-radio value="specify">{{ $t('cmp.cronSpecify') }}</el-radio>
      <el-radio
        v-if="allowNone"
        value="none"
      >
        {{ $t('cmp.cronUnspecified') }}
      </el-radio>
    </el-radio-group>

    <div class="crontab-field__inputs">
      <template v-if="mode === 'range'">
        <el-input-number
          v-model="rangeStart"
          :min="min"
          :max="max"
          @change="emitValue"
        />
        <span>{{ $t('cmp.cronTo') }}</span>
        <el-input-number
          v-model="rangeEnd"
          :min="min"
          :max="max"
          @change="emitValue"
        />
      </template>
      <template v-else-if="mode === 'interval'">
        {{ $t('cmp.cronFrom') }}
        <el-input-number
          v-model="intervalStart"
          :min="min"
          :max="max"
          @change="emitValue"
        />
        {{ $t('cmp.cronFromText') }}
        <el-input-number
          v-model="intervalStep"
          :min="1"
          :max="max"
          @change="emitValue"
        />
        {{ unit }}
      </template>
      <template v-else-if="mode === 'specify'">
        <el-checkbox-group
          v-model="specifyValues"
          @change="emitValue"
        >
          <el-checkbox
            v-for="n in range"
            :key="n"
            :value="n"
            :label="String(n)"
          />
        </el-checkbox-group>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  modelValue: string
  min: number
  max: number
  unit: string
  allowNone?: boolean
}>()

const emit = defineEmits<{ (e: 'update:modelValue', val: string): void }>()

const mode = ref('every')
const rangeStart = ref(props.min)
const rangeEnd = ref(props.max)
const intervalStart = ref(props.min)
const intervalStep = ref(1)
const specifyValues = ref<string[]>([])

const range = computed(() => {
  const arr = []
  for (let i = props.min; i <= props.max; i++) arr.push(i)
  return arr
})

watch(
  () => props.modelValue,
  (val) => {
    if (!val || val === '*') {
      mode.value = 'every'
      return
    }
    if (val === '?') {
      mode.value = 'none'
      return
    }
    if (val.includes('-')) {
      mode.value = 'range'
      const [s, e] = val.split('-')
      rangeStart.value = Number(s)
      rangeEnd.value = Number(e)
    } else if (val.includes('/')) {
      mode.value = 'interval'
      const [s, step] = val.split('/')
      intervalStart.value = s === '*' ? props.min : Number(s)
      intervalStep.value = Number(step)
    } else if (val.includes(',')) {
      mode.value = 'specify'
      specifyValues.value = val.split(',').map(String)
    } else {
      mode.value = 'specify'
      specifyValues.value = [val]
    }
  },
  { immediate: true }
)

function onModeChange() {
  emitValue()
}

function emitValue() {
  let val = ''
  switch (mode.value) {
    case 'every':
      val = '*'
      break
    case 'none':
      val = '?'
      break
    case 'range':
      val = `${rangeStart.value}-${rangeEnd.value}`
      break
    case 'interval':
      val = `${intervalStart.value}/${intervalStep.value}`
      break
    case 'specify':
      val =
        specifyValues.value
          .map(Number)
          .sort((a, b) => a - b)
          .join(',') || '*'
      break
  }
  emit('update:modelValue', val)
}
</script>

<style scoped>
.crontab-field {
  padding: 8px 0;
}
.crontab-field__inputs {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
