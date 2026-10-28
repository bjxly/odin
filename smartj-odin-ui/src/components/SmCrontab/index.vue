<template>
  <div class="sm-crontab">
    <el-input
      v-model="expression"
      :placeholder="$t('cmp.cronPlaceholder')"
      @change="onInputChange"
    >
      <template #prepend>Cron</template>
    </el-input>

    <el-tabs
      v-model="activeTab"
      class="sm-crontab__tabs"
    >
      <el-tab-pane
        :label="$t('cmp.cronSecond')"
        name="second"
      >
        <CrontabField
          v-model="fields.second"
          :min="0"
          :max="59"
          :unit="$t('cmp.cronSecond')"
        />
      </el-tab-pane>
      <el-tab-pane
        :label="$t('cmp.cronMinute')"
        name="minute"
      >
        <CrontabField
          v-model="fields.minute"
          :min="0"
          :max="59"
          :unit="$t('cmp.cronMinute')"
        />
      </el-tab-pane>
      <el-tab-pane
        :label="$t('cmp.cronHour')"
        name="hour"
      >
        <CrontabField
          v-model="fields.hour"
          :min="0"
          :max="23"
          :unit="$t('cmp.cronHour')"
        />
      </el-tab-pane>
      <el-tab-pane
        :label="$t('cmp.cronDay')"
        name="day"
      >
        <CrontabField
          v-model="fields.day"
          :min="1"
          :max="31"
          :unit="$t('cmp.cronDay')"
          allow-none
        />
      </el-tab-pane>
      <el-tab-pane
        :label="$t('cmp.cronMonth')"
        name="month"
      >
        <CrontabField
          v-model="fields.month"
          :min="1"
          :max="12"
          :unit="$t('cmp.cronMonth')"
        />
      </el-tab-pane>
      <el-tab-pane
        :label="$t('cmp.cronWeek')"
        name="week"
      >
        <CrontabField
          v-model="fields.week"
          :min="1"
          :max="7"
          :unit="$t('cmp.cronWeek')"
          allow-none
        />
      </el-tab-pane>
    </el-tabs>

    <div class="sm-crontab__preview">
      <div class="sm-crontab__expression">
        {{ $t('cmp.cronExpression') }}
        <code>{{ expression }}</code>
      </div>
      <div class="sm-crontab__next-runs">
        <div class="sm-crontab__next-title">{{ $t('cmp.cronNextRuns') }}</div>
        <div
          v-for="(t, i) in nextRuns"
          :key="i"
          class="sm-crontab__next-item"
        >
          {{ t }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import CrontabField from './CrontabField.vue'

const { locale } = useI18n()

const props = withDefaults(
  defineProps<{
    modelValue?: string
  }>(),
  {
    modelValue: '0 * * * * *',
  }
)

const emit = defineEmits<{ (e: 'update:modelValue', val: string): void }>()

const activeTab = ref('second')

const fields = reactive({
  second: '0',
  minute: '*',
  hour: '*',
  day: '*',
  month: '*',
  week: '*',
})

const expression = ref(props.modelValue)

function parseExpression(val: string) {
  const parts = val.split(/\s+/)
  if (parts.length >= 6) {
    fields.second = parts[0]
    fields.minute = parts[1]
    fields.hour = parts[2]
    fields.day = parts[3]
    fields.month = parts[4]
    fields.week = parts[5]
  }
}
parseExpression(props.modelValue)

watch(
  fields,
  () => {
    const val = `${fields.second} ${fields.minute} ${fields.hour} ${fields.day} ${fields.month} ${fields.week}`
    expression.value = val
    emit('update:modelValue', val)
  },
  { deep: true }
)

function onInputChange(val: string) {
  parseExpression(val)
  emit('update:modelValue', val)
}

const nextRuns = computed(() => {
  const runs: string[] = []
  const now = new Date()
  for (let i = 0; i < 5; i++) {
    const d = new Date(now.getTime() + (i + 1) * 60000)
    runs.push(d.toLocaleString(locale.value))
  }
  return runs
})
</script>

<style scoped>
.sm-crontab {
  padding: 16px;
  background: var(--sm-bg-card);
  border-radius: var(--sm-radius-md);
}
.sm-crontab__tabs {
  margin-top: 16px;
}
.sm-crontab__preview {
  margin-top: 16px;
  padding: 12px;
  background: var(--sm-bg-page);
  border-radius: var(--sm-radius-md);
}
.sm-crontab__expression {
  font-size: 14px;
  color: var(--sm-text-regular);
  margin-bottom: 8px;
}
.sm-crontab__expression code {
  font-weight: 600;
  color: var(--sm-color-primary);
}
.sm-crontab__next-title {
  font-size: 13px;
  color: var(--sm-text-secondary);
  margin-bottom: 4px;
}
.sm-crontab__next-item {
  font-size: 13px;
  color: var(--sm-text-regular);
  line-height: 1.8;
}
</style>
