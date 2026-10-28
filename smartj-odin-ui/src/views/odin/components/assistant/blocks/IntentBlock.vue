<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MessageBlock, IntentBlockPayload } from '@/types/odin'

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()

const intent = computed<IntentBlockPayload | null>(() => {
  const p = props.block?.payload as IntentBlockPayload | null | undefined
  if (!p || typeof p !== 'object') return null
  return p
})

const className = computed(() => intent.value?.className || intent.value?.classCode || '—')
const properties = computed<string[]>(() => intent.value?.properties ?? [])
const filters = computed(() => intent.value?.filters ?? [])
const joins = computed(() => intent.value?.joins ?? [])

function filterText(f: { property: string; op: string; value?: unknown }): string {
  const val = f.value == null ? 'NULL' : Array.isArray(f.value) ? `[${f.value.join(', ')}]` : String(f.value)
  return `${f.property} ${f.op} ${val}`
}
</script>

<template>
  <div class="intent-block">
    <div class="block-head">
      <span class="block-title">{{ block.title || t('odinAssistant.blockIntent') }}</span>
      <el-tag size="small" type="primary" effect="plain">AST</el-tag>
    </div>

    <div v-if="!intent" class="empty">—</div>

    <div v-else class="intent-grid">
      <div class="row">
        <span class="k">{{ t('odinAssistant.provClass') }}</span>
        <el-tag size="small" type="success" effect="dark">{{ className }}</el-tag>
      </div>

      <div v-if="properties.length" class="row">
        <span class="k">投影</span>
        <el-tag v-for="p in properties" :key="p" size="small" effect="plain">{{ p }}</el-tag>
      </div>

      <div class="row">
        <span class="k">过滤</span>
        <template v-if="filters.length">
          <el-tag
            v-for="(f, i) in filters"
            :key="i"
            size="small"
            type="warning"
            effect="plain"
            class="filter-tag"
          >
            {{ filterText(f) }}
          </el-tag>
        </template>
        <span v-else class="muted">无</span>
      </div>

      <div v-if="joins.length" class="row">
        <span class="k">关系</span>
        <el-tag v-for="(j, i) in joins" :key="i" size="small" type="primary" effect="plain">
          {{ j.relation }}
        </el-tag>
      </div>
    </div>
  </div>
</template>

<style scoped>
.intent-block {
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 10px;
  background: linear-gradient(180deg, var(--el-color-primary-light-9), var(--el-fill-color-blank));
  padding: 10px 12px;
}
.block-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.block-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.intent-grid {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  font-size: 12px;
}
.k {
  min-width: 48px;
  color: var(--el-text-color-secondary);
  font-weight: 600;
}
.muted {
  color: var(--el-text-color-placeholder);
}
.filter-tag {
  font-family: 'JetBrains Mono', Consolas, monospace;
}
.empty {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}
</style>
