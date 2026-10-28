<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  ElDialog,
  ElButton,
  ElTable,
  ElTableColumn,
  ElInput,
  ElAlert,
} from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'

const props = defineProps<{
  open: boolean
  modelValue: Record<string, string>
  propertyLabel?: string
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'update:modelValue', v: Record<string, string>): void
  (e: 'save', v: Record<string, string>): void
}>()

interface Row {
  phys: string
  onto: string
}

const rows = ref<Row[]>([])

watch(
  () => [props.open, props.modelValue],
  () => {
    if (!props.open) return
    rows.value = Object.entries(props.modelValue ?? {}).map(([phys, onto]) => ({ phys, onto }))
    if (!rows.value.length) {
      rows.value = [
        { phys: '', onto: '' },
        { phys: '', onto: '' },
      ]
    }
  },
  { immediate: true, deep: true },
)

const preview = computed(() => {
  const out: Record<string, string> = {}
  for (const r of rows.value) {
    const p = r.phys.trim()
    const o = r.onto.trim()
    if (p && o) out[p] = o
  }
  return out
})

function addRow() {
  rows.value.push({ phys: '', onto: '' })
}

function removeRow(i: number) {
  rows.value.splice(i, 1)
}

function onSave() {
  emit('update:modelValue', preview.value)
  emit('save', preview.value)
  emit('update:open', false)
}
</script>

<template>
  <ElDialog
    :model-value="open"
    :title="`枚举 value_map${propertyLabel ? ` — ${propertyLabel}` : ''}`"
    width="520px"
    @update:model-value="emit('update:open', $event)"
  >
    <ElAlert
      type="info"
      :closable="false"
      title="左列填物理库中的值，右列填本体枚举值（查询时会自动翻译）"
      style="margin-bottom: 12px"
    />

    <ElTable :data="rows" size="small" border>
      <ElTableColumn label="物理值" min-width="140">
        <template #default="{ row }">
          <ElInput v-model="row.phys" size="small" placeholder="如 1 / VIP" />
        </template>
      </ElTableColumn>
      <ElTableColumn label="本体枚举值" min-width="140">
        <template #default="{ row }">
          <ElInput v-model="row.onto" size="small" placeholder="如 A / confirmed" />
        </template>
      </ElTableColumn>
      <ElTableColumn label="" width="56">
        <template #default="{ $index }">
          <ElButton size="small" text type="danger" :icon="Delete" @click="removeRow($index)" />
        </template>
      </ElTableColumn>
    </ElTable>

    <div style="margin-top: 10px; display: flex; gap: 8px; align-items: center">
      <ElButton size="small" :icon="Plus" @click="addRow">加一行</ElButton>
      <span class="preview">预览：{{ Object.keys(preview).length }} 组</span>
    </div>

    <template #footer>
      <ElButton @click="emit('update:open', false)">取消</ElButton>
      <ElButton type="primary" @click="onSave">确定</ElButton>
    </template>
  </ElDialog>
</template>

<style scoped>
.preview {
  font-size: 12px;
  color: #64748b;
}
</style>
