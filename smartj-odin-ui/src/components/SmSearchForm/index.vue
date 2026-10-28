<template>
  <div
    class="sm-search-form"
    :style="{ '--sm-field-width': fieldWidth }"
  >
    <el-form
      :model="model"
      :label-width="labelWidth"
      @submit.prevent="handleSearch"
    >
      <div class="sm-search-form__row">
        <div class="sm-search-form__fields">
          <slot />
        </div>
        <div class="sm-search-form__actions">
          <el-button-group>
            <el-button
              type="primary"
              @click="handleSearch"
            >
              <el-icon><Search /></el-icon>
              {{ $t('common.search') }}
            </el-button>
            <el-button
              :title="$t('common.reset')"
              @click="handleReset"
            >
              <el-icon><Refresh /></el-icon>
            </el-button>
            <el-button
              v-if="hasExpandSlot"
              link
              type="SmTreeSelect primary"
              @click="expanded = !expanded"
            >
              <!--              {{ expanded ? '收起' : '展开' }}-->
              <el-icon
                class="sm-search-form__arrow"
                :class="{ 'is-expanded': expanded }"
              >
                <ArrowDown />
              </el-icon>
            </el-button>
          </el-button-group>
        </div>
      </div>
      <template v-if="expanded && hasExpandSlot">
        <div class="sm-search-form__row">
          <div class="sm-search-form__fields">
            <slot name="expand" />
          </div>
        </div>
      </template>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, useSlots } from 'vue'
import { Search, Refresh, ArrowDown } from '@element-plus/icons-vue'

withDefaults(
  defineProps<{
    model: Record<string, any>
    labelWidth?: string
    fieldWidth?: string
  }>(),
  {
    labelWidth: 'auto',
    fieldWidth: '200px',
  }
)

const emit = defineEmits<{
  (e: 'search'): void
  (e: 'reset'): void
}>()

const slots = useSlots()
const expanded = ref(false)
const hasExpandSlot = !!slots.expand

function handleSearch() {
  emit('search')
}

function handleReset() {
  emit('reset')
  emit('search')
}
</script>

<style scoped>
.sm-search-form {
  padding: 16px;
  background: var(--sm-bg-card);
  border-radius: var(--sm-radius-md);
  margin-bottom: 16px;
}
.sm-search-form__row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}
.sm-search-form__fields {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 0 16px;
}
.sm-search-form__actions {
  flex-shrink: 0;
}
.sm-search-form__actions :deep(.el-button-group) {
  display: inline-flex;
  align-items: center;
}
.sm-search-form__actions :deep(.el-button-group .el-button) {
  vertical-align: top;
}
.sm-search-form__actions :deep(.el-button-group .el-button--link) {
  margin-left: 4px;
}
/* 表单字段默认宽度 — 通过 prop fieldWidth 可自定义 */
.sm-search-form__fields :deep(.el-input),
.sm-search-form__fields :deep(.el-select),
.sm-search-form__fields :deep(.el-autocomplete),
.sm-search-form__fields :deep(.el-date-picker),
.sm-search-form__fields :deep(.el-time-picker),
.sm-search-form__fields :deep(.el-cascader) {
  width: var(--sm-field-width, 200px);
}
.sm-search-form__arrow {
  transition: transform 0.3s;
  margin-left: 4px;
}
.sm-search-form__arrow.is-expanded {
  transform: rotate(180deg);
}
</style>
