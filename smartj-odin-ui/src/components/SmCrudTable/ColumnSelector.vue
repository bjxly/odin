<template>
  <el-popover
    placement="bottom-end"
    :width="280"
    trigger="click"
  >
    <template #reference>
      <el-button
        :icon="Setting"
        :title="$t('cmp.columnSettings')"
      ></el-button>
    </template>

    <div class="sm-column-selector">
      <div class="sm-column-selector__header">
        <span class="sm-column-selector__title">{{ $t('cmp.columnDisplaySettings') }}</span>
        <el-button
          link
          type="primary"
          size="small"
          @click="$emit('reset')"
        >
          {{ $t('common.reset') }}
        </el-button>
      </div>

      <div
        ref="sortListRef"
        class="sm-column-selector__list"
      >
        <div
          v-for="cfg in configs"
          :key="cfg.prop"
          class="sm-column-selector__item"
          :data-prop="cfg.prop"
        >
          <el-icon class="sm-column-selector__drag-icon"><Rank /></el-icon>
          <el-checkbox
            :model-value="cfg.visible"
            size="small"
            @change="$emit('toggle', cfg.prop)"
          >
            {{ cfg.label }}
          </el-checkbox>
        </div>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue'
import { Setting, Rank } from '@element-plus/icons-vue'
import Sortable from 'sortablejs'

const props = defineProps<{
  configs: Array<{ prop: string; label: string; visible: boolean; order: number }>
}>()

const emit = defineEmits<{
  (e: 'toggle', prop: string): void
  (e: 'update-order', orderedProps: string[]): void
  (e: 'reset'): void
}>()

const sortListRef = ref<HTMLElement>()
let sortableInstance: Sortable | null = null

function initSortable() {
  if (!sortListRef.value) return
  if (sortableInstance) {
    sortableInstance.destroy()
  }
  sortableInstance = Sortable.create(sortListRef.value, {
    animation: 200,
    handle: '.sm-column-selector__drag-icon',
    ghostClass: 'sm-column-selector__ghost',
    onEnd: () => {
      if (!sortListRef.value) return
      const items = sortListRef.value.querySelectorAll('.sm-column-selector__item')
      const orderedProps = Array.from(items).map((el) => (el as HTMLElement).dataset.prop || '')
      emit('update-order', orderedProps)
    },
  })
}

watch(
  () => props.configs,
  () => {
    nextTick(initSortable)
  },
  { deep: true }
)

onMounted(() => {
  nextTick(initSortable)
})
</script>

<style scoped>
.sm-column-selector__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0 10px;
}
.sm-column-selector__title {
  font-size: var(--sm-font-size-base);
  font-weight: 600;
  color: var(--sm-text-primary);
}
.sm-column-selector__list {
  max-height: 320px;
  overflow-y: auto;
}
.sm-column-selector__item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 4px;
  border-radius: var(--sm-radius-sm);
  cursor: default;
  margin-bottom: 2px;
  transition: background 0.15s;
}
.sm-column-selector__item:hover {
  background: var(--sm-bg-hover);
}
.sm-column-selector__drag-icon {
  cursor: grab;
  color: var(--sm-text-secondary);
  flex-shrink: 0;
  font-size: 14px;
}
.sm-column-selector__drag-icon:active {
  cursor: grabbing;
}
.sm-column-selector__ghost {
  opacity: 0.4;
  background: var(--sm-color-primary);
}
</style>
