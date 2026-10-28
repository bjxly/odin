<template>
  <div class="sm-calendar">
    <!-- 头部导航 -->
    <div class="sm-calendar__header">
      <div class="sm-calendar__nav">
        <button class="sm-calendar__nav-btn" @click="prevMonth" :disabled="prevDisabled">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z" />
          </svg>
        </button>
        <span class="sm-calendar__title" @click="showYearPicker = !showYearPicker">
          {{ currentYear }}年 {{ currentMonth + 1 }}月
        </span>
        <button class="sm-calendar__nav-btn" @click="nextMonth" :disabled="nextDisabled">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" />
          </svg>
        </button>
      </div>
      <button v-if="showToday" class="sm-calendar__today-btn" @click="goToday">今天</button>
    </div>

    <!-- 年份快速选择 -->
    <div v-if="showYearPicker" class="sm-calendar__year-picker">
      <div class="sm-calendar__year-grid">
        <button
          v-for="y in yearRange"
          :key="y"
          class="sm-calendar__year-item"
          :class="{ 'sm-calendar__year-item--active': y === currentYear }"
          @click="selectYear(y)"
        >
          {{ y }}
        </button>
      </div>
    </div>

    <!-- 星期表头 -->
    <div class="sm-calendar__weekdays">
      <span
        v-for="d in weekDays"
        :key="d"
        class="sm-calendar__weekday"
      >
        {{ d }}
      </span>
    </div>

    <!-- 日期网格 -->
    <div class="sm-calendar__grid">
      <div
        v-for="(cell, idx) in calendarCells"
        :key="idx"
        class="sm-calendar__cell"
        :class="{
          'sm-calendar__cell--other-month': !cell.isCurrentMonth,
          'sm-calendar__cell--today': cell.isToday,
          'sm-calendar__cell--selected': isSelected(cell),
          'sm-calendar__cell--range': isInRange(cell),
          'sm-calendar__cell--range-start': isRangeStart(cell),
          'sm-calendar__cell--range-end': isRangeEnd(cell),
          'sm-calendar__cell--disabled': cell.disabled,
        }"
        @click="handleCellClick(cell)"
      >
        <slot name="cell" :cell="cell" :date="cell.date">
          <span class="sm-calendar__date">{{ cell.day }}</span>
        </slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface CalendarCell {
  date: Date
  day: number
  isCurrentMonth: boolean
  isToday: boolean
  disabled: boolean
}

const props = withDefaults(
  defineProps<{
    modelValue?: Date | [Date, Date] | Date[]
    mode?: string
    showToday?: boolean
    weekStartsOn?: number
    disabledDate?: (date: Date) => boolean
  }>(),
  {
    showToday: true,
    weekStartsOn: 0,
    mode: 'single',
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: Date | [Date, Date] | Date[]): void
  (e: 'change', value: Date | [Date, Date] | Date[]): void
  (e: 'monthChange', year: number, month: number): void
}>()

const weekDays = ['日', '一', '二', '三', '四', '五', '六']

const currentYear = ref(new Date().getFullYear())
const currentMonth = ref(new Date().getMonth())
const showYearPicker = ref(false)

const prevDisabled = computed(() => false)
const nextDisabled = computed(() => false)

const yearRange = computed(() => {
  const y = currentYear.value
  const result: number[] = []
  for (let i = y - 6; i <= y + 6; i++) {
    result.push(i)
  }
  return result
})

const calendarCells = computed(() => {
  const cells: CalendarCell[] = []
  const firstDay = new Date(currentYear.value, currentMonth.value, 1)
  const lastDay = new Date(currentYear.value, currentMonth.value + 1, 0)
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  // 计算前月填充
  const startDayOfWeek = firstDay.getDay()
  const prevFill = (startDayOfWeek - props.weekStartsOn + 7) % 7
  const prevLast = new Date(currentYear.value, currentMonth.value, 0).getDate()

  for (let i = prevFill - 1; i >= 0; i--) {
    const d = new Date(currentYear.value, currentMonth.value - 1, prevLast - i)
    cells.push({
      date: d,
      day: prevLast - i,
      isCurrentMonth: false,
      isToday: d.getTime() === today.getTime(),
      disabled: props.disabledDate ? props.disabledDate(d) : false,
    })
  }

  // 当前月
  for (let i = 1; i <= lastDay.getDate(); i++) {
    const d = new Date(currentYear.value, currentMonth.value, i)
    cells.push({
      date: d,
      day: i,
      isCurrentMonth: true,
      isToday: d.getTime() === today.getTime(),
      disabled: props.disabledDate ? props.disabledDate(d) : false,
    })
  }

  // 后月填充
  const remaining = 42 - cells.length
  for (let i = 1; i <= remaining; i++) {
    const d = new Date(currentYear.value, currentMonth.value + 1, i)
    cells.push({
      date: d,
      day: i,
      isCurrentMonth: false,
      isToday: d.getTime() === today.getTime(),
      disabled: props.disabledDate ? props.disabledDate(d) : false,
    })
  }

  return cells
})

function isSelected(cell: CalendarCell): boolean {
  if (!props.modelValue) return false
  if (props.mode === 'single') {
    const mv = props.modelValue as Date
    return cell.date.toDateString() === mv.toDateString()
  }
  if (props.mode === 'range' && Array.isArray(props.modelValue)) {
    const [start, end] = props.modelValue as [Date, Date]
    if (start && end) {
      return cell.date.toDateString() === start.toDateString() ||
             cell.date.toDateString() === end.toDateString()
    }
    if (start) return cell.date.toDateString() === start.toDateString()
  }
  if (props.mode === 'multiple' && Array.isArray(props.modelValue)) {
    return (props.modelValue as Date[]).some(
      (d) => cell.date.toDateString() === d.toDateString(),
    )
  }
  return false
}

function isRangeStart(cell: CalendarCell): boolean {
  if (props.mode !== 'range' || !Array.isArray(props.modelValue)) return false
  const [start] = props.modelValue as [Date, Date]
  return start ? cell.date.toDateString() === start.toDateString() : false
}

function isRangeEnd(cell: CalendarCell): boolean {
  if (props.mode !== 'range' || !Array.isArray(props.modelValue)) return false
  const [, end] = props.modelValue as [Date, Date]
  return end ? cell.date.toDateString() === end.toDateString() : false
}

function isInRange(cell: CalendarCell): boolean {
  if (props.mode !== 'range' || !Array.isArray(props.modelValue)) return false
  const [start, end] = props.modelValue as [Date, Date]
  if (!start || !end) return false
  const t = cell.date.getTime()
  return t > start.getTime() && t < end.getTime()
}

function handleCellClick(cell: CalendarCell) {
  if (cell.disabled) return

  if (props.mode === 'single') {
    const v = new Date(cell.date)
    emit('update:modelValue', v)
    emit('change', v)
  } else if (props.mode === 'range') {
    const current = props.modelValue as [Date, Date] | undefined
    if (!current || !current[0] || (current[0] && current[1])) {
      // 新的选择开始
      const v: [Date, Date] = [new Date(cell.date), null as unknown as Date]
      emit('update:modelValue', v)
      emit('change', v)
    } else {
      // 选择结束
      const start = current[0]
      const end = new Date(cell.date)
      if (end < start) {
        emit('update:modelValue', [end, start])
        emit('change', [end, start])
      } else {
        emit('update:modelValue', [start, end])
        emit('change', [start, end])
      }
    }
  } else if (props.mode === 'multiple') {
    const selected = [...(props.modelValue as Date[] || [])]
    const idx = selected.findIndex(
      (d) => d.toDateString() === cell.date.toDateString(),
    )
    if (idx >= 0) {
      selected.splice(idx, 1)
    } else {
      selected.push(new Date(cell.date))
    }
    emit('update:modelValue', selected)
    emit('change', selected)
  }
}

function prevMonth() {
  if (currentMonth.value === 0) {
    currentMonth.value = 11
    currentYear.value--
  } else {
    currentMonth.value--
  }
  emit('monthChange', currentYear.value, currentMonth.value)
}

function nextMonth() {
  if (currentMonth.value === 11) {
    currentMonth.value = 0
    currentYear.value++
  } else {
    currentMonth.value++
  }
  emit('monthChange', currentYear.value, currentMonth.value)
}

function goToday() {
  const now = new Date()
  currentYear.value = now.getFullYear()
  currentMonth.value = now.getMonth()
  emit('monthChange', currentYear.value, currentMonth.value)
}

function selectYear(year: number) {
  currentYear.value = year
  showYearPicker.value = false
  emit('monthChange', currentYear.value, currentMonth.value)
}
</script>

<style scoped>
.sm-calendar {
  width: 320px;
  border: 1px solid var(--el-border-color-light, #e4e7ed);
  border-radius: 8px;
  padding: 12px;
  background: var(--el-bg-color, #fff);
}
.sm-calendar__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.sm-calendar__nav {
  display: flex;
  align-items: center;
  gap: 4px;
}
.sm-calendar__nav-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  color: var(--el-text-color-regular, #606266);
  transition: background 0.15s;
}
.sm-calendar__nav-btn:hover:not(:disabled) {
  background: var(--el-fill-color-light, #f5f7fa);
}
.sm-calendar__nav-btn:disabled {
  color: var(--el-text-color-placeholder, #c0c4cc);
  cursor: not-allowed;
}
.sm-calendar__title {
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  padding: 2px 8px;
  border-radius: 4px;
  min-width: 100px;
  text-align: center;
  user-select: none;
}
.sm-calendar__title:hover {
  background: var(--el-fill-color-light, #f5f7fa);
}
.sm-calendar__today-btn {
  font-size: 12px;
  padding: 4px 12px;
  border: 1px solid var(--el-border-color, #dcdfe6);
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  color: var(--el-color-primary, #409eff);
  transition: all 0.15s;
}
.sm-calendar__today-btn:hover {
  border-color: var(--el-color-primary, #409eff);
  background: var(--el-color-primary-light-9, #ecf5ff);
}
/* 年份选择器 */
.sm-calendar__year-picker {
  margin-bottom: 8px;
}
.sm-calendar__year-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 4px;
}
.sm-calendar__year-item {
  padding: 6px 0;
  text-align: center;
  font-size: 13px;
  border: none;
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  color: var(--el-text-color-regular, #606266);
  transition: all 0.15s;
}
.sm-calendar__year-item:hover {
  background: var(--el-fill-color-light, #f5f7fa);
}
.sm-calendar__year-item--active {
  background: var(--el-color-primary, #409eff) !important;
  color: #fff !important;
}
/* 星期表头 */
.sm-calendar__weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  margin-bottom: 4px;
}
.sm-calendar__weekday {
  text-align: center;
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
  padding: 4px 0;
  font-weight: 500;
}
/* 日期网格 */
.sm-calendar__grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
}
.sm-calendar__cell {
  position: relative;
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.12s;
}
.sm-calendar__cell:hover {
  background: var(--el-fill-color-light, #f5f7fa);
}
.sm-calendar__cell--other-month {
  color: var(--el-text-color-placeholder, #c0c4cc);
}
.sm-calendar__cell--today {
  font-weight: 700;
  color: var(--el-color-primary, #409eff);
}
.sm-calendar__cell--selected {
  background: var(--el-color-primary, #409eff) !important;
  color: #fff !important;
  font-weight: 600;
}
.sm-calendar__cell--range {
  background: var(--el-color-primary-light-9, #ecf5ff);
  border-radius: 0;
}
.sm-calendar__cell--range-start {
  background: var(--el-color-primary, #409eff) !important;
  color: #fff !important;
  border-radius: 6px 0 0 6px;
}
.sm-calendar__cell--range-end {
  background: var(--el-color-primary, #409eff) !important;
  color: #fff !important;
  border-radius: 0 6px 6px 0;
}
.sm-calendar__cell--disabled {
  color: var(--el-text-color-placeholder, #c0c4cc);
  cursor: not-allowed;
  opacity: 0.5;
}
.sm-calendar__date {
  position: relative;
  z-index: 1;
}
</style>
