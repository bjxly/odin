<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { ArrowRight, CopyDocument, Document } from '@element-plus/icons-vue'
import type { MessageBlock, SqlBlockPayload } from '@/types/odin'

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()
const expanded = ref(false)

const rawPayload = computed(() => (props.block?.payload ?? {}) as SqlBlockPayload)

interface SqlStep {
  purpose: string
  sql: string
  datasourceLabel: string
}

const steps = computed<SqlStep[]>(() => {
  const p = rawPayload.value
  if (Array.isArray(p.steps) && p.steps.length) {
    return p.steps
      .map((s) => ({
        purpose: String(s.purpose ?? '').trim() || t('odinAssistant.blockSql'),
        sql: String(s.sql ?? ''),
        datasourceLabel: String(s.datasourceLabel ?? '').trim(),
      }))
      .filter((s) => s.sql.trim().length > 0)
  }
  const sql = String(p.sql ?? '').trim()
  if (!sql) return []
  return [{
    purpose: String(p.explanation ?? '').trim() || t('odinAssistant.blockSql'),
    sql,
    datasourceLabel: String(p.datasourceName ?? '').trim(),
  }]
})

async function onCopy(sql: string) {
  if (!sql) return
  try {
    await navigator.clipboard.writeText(sql)
    ElMessage.success(t('odinAssistant.sqlCopied'))
  } catch {
    ElMessage.warning(t('odinAssistant.sqlCopyFail'))
  }
}

/** 简易 SQL 关键字高亮（不引入 highlight.js 避免 bundle 膨胀） */
function highlightSql(sql: string): string {
  const escaped = sql.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  const keywords = /\b(SELECT|FROM|WHERE|JOIN|LEFT|RIGHT|INNER|OUTER|ON|AND|OR|NOT|IN|IS|NULL|AS|ORDER|BY|GROUP|HAVING|LIMIT|OFFSET|INSERT|INTO|VALUES|UPDATE|SET|DELETE|CREATE|TABLE|ALTER|DROP|INDEX|UNION|ALL|DISTINCT|COUNT|SUM|AVG|MAX|MIN|CASE|WHEN|THEN|ELSE|END|EXISTS|BETWEEN|LIKE|ASC|DESC|WITH|RECURSIVE|CAST|COALESCE|NULLIF|TOP|FETCH|NEXT|ROWS|ONLY|FIRST|CROSS|NATURAL|FULL|USING|EXCEPT|INTERSECT|OVER|PARTITION|ROW_NUMBER|RANK|DENSE_RANK|LAG|LEAD)\b/gi
  const strings = /('[^']*')/g
  const numbers = /\b(\d+\.?\d*)\b/g
  let result = escaped
  result = result.replace(strings, '<span class="sql-str">$1</span>')
  result = result.replace(keywords, '<span class="sql-kw">$1</span>')
  result = result.replace(numbers, '<span class="sql-num">$1</span>')
  return result
}
</script>

<template>
  <div class="sql-block">
    <div class="sql-block__header" @click="expanded = !expanded">
      <el-icon class="sql-block__icon"><Document /></el-icon>
      <span class="sql-block__count">{{ t('odinAssistant.sqlSteps', { n: steps.length }) }}</span>
      <el-icon class="sql-block__arrow" :class="{ 'is-expanded': expanded }"><ArrowRight /></el-icon>
    </div>

    <div v-show="expanded" class="sql-block__body">
      <div v-if="!steps.length" class="sql-block__empty">{{ t('odinAssistant.sqlEmpty') }}</div>
      <div v-for="(step, i) in steps" :key="i" class="sql-block__step">
        <div class="sql-block__step-header">
          <span class="sql-block__purpose">{{ step.purpose }}</span>
          <el-tag v-if="step.datasourceLabel" size="small" type="info" effect="plain">
            {{ step.datasourceLabel }}
          </el-tag>
          <el-button
            class="sql-block__copy"
            link
            size="small"
            :icon="CopyDocument"
            @click.stop="onCopy(step.sql)"
          >
            {{ t('odinAssistant.sqlCopy') }}
          </el-button>
        </div>
        <!-- eslint-disable-next-line vue/no-v-html -->
        <pre class="sql-block__code"><code v-html="highlightSql(step.sql)"></code></pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sql-block {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
}
.sql-block__header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
  background: var(--el-fill-color-lighter);
  font-size: 13px;
  color: var(--el-text-color-secondary);
  user-select: none;
}
.sql-block__header:hover {
  background: var(--el-fill-color-light);
}
.sql-block__icon {
  color: var(--el-color-primary);
}
.sql-block__count {
  font-weight: 600;
}
.sql-block__arrow {
  transition: transform 0.2s;
  margin-left: auto;
}
.sql-block__arrow.is-expanded {
  transform: rotate(90deg);
}
.sql-block__body {
  padding: 12px 14px;
  border-top: 1px solid var(--el-border-color-lighter);
}
.sql-block__empty {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.sql-block__step {
  margin-bottom: 12px;
}
.sql-block__step:last-child {
  margin-bottom: 0;
}
.sql-block__step-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 13px;
}
.sql-block__purpose {
  font-weight: 500;
  color: var(--el-text-color-primary);
}
.sql-block__copy {
  margin-left: auto;
}
.sql-block__code {
  background: var(--el-fill-color);
  padding: 10px 12px;
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.6;
  overflow-x: auto;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--el-text-color-primary);
  font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
}
.sql-block__code :deep(.sql-kw) {
  color: var(--el-color-primary);
  font-weight: 600;
}
.sql-block__code :deep(.sql-str) {
  color: var(--el-color-success);
}
.sql-block__code :deep(.sql-num) {
  color: var(--el-color-warning);
}
</style>
