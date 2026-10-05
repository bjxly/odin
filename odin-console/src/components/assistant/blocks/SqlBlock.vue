<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElButton, ElMessage, ElTag } from 'element-plus'
import { ArrowRight, CopyDocument, Document } from '@element-plus/icons-vue'
import { copyText } from '@/utils/format'
import type { BlockVO, SQLBlockPayload, SQLStepItem, SqlPayload } from '@/types/assistant'

/**
 * sql 块（#66 新 SQL 生成架构）：折叠展示一次问答产出的多条查询语句。
 *
 * - 默认收起，标题「N 条查询语句」，点击展开；
 * - 展开后按 step 列出：用途说明 + 数据源标签 + SQL 代码块 + 复制按钮；
 * - 兼容旧契约（单条 SqlPayload.sql）：无 steps 时归一为单步展示。
 */
const props = defineProps<{ block: BlockVO }>()

const rawPayload = computed(() => (props.block?.payload ?? {}) as SQLBlockPayload | SqlPayload)

/** 归一为步骤数组：优先取新契约 steps，回退旧契约单条 sql */
const steps = computed<SQLStepItem[]>(() => {
  const p = rawPayload.value as Partial<SQLBlockPayload> & Partial<SqlPayload>
  if (Array.isArray(p.steps)) {
    return p.steps
      .filter((s): s is SQLStepItem => !!s && typeof s === 'object')
      .map((s) => ({
        purpose: String(s.purpose ?? '').trim() || '查询语句',
        sql: String(s.sql ?? ''),
        datasource_label: String(s.datasource_label ?? '').trim(),
      }))
      .filter((s) => s.sql.trim().length > 0)
  }
  const sql = String(p.sql ?? '').trim()
  if (!sql) return []
  return [
    {
      purpose: String(p.explanation ?? p.class_name ?? '').trim() || '翻译 SQL',
      sql,
      datasource_label:
        String(p.datasource_name ?? '').trim() ||
        (p.datasource_id ? `数据源 #${p.datasource_id}` : ''),
    },
  ]
})

const expanded = ref(false)

async function onCopy(sql: string) {
  if (!sql) return
  const ok = await copyText(sql)
  if (ok) ElMessage.success('SQL 已复制到剪贴板')
  else ElMessage.warning('复制失败，请手动选择文本')
}
</script>

<template>
  <div class="sql-block">
    <div class="sql-block__header" @click="expanded = !expanded">
      <el-icon class="sql-block__icon"><Document /></el-icon>
      <span class="sql-block__count">{{ steps.length }} 条查询语句</span>
      <el-icon class="sql-block__arrow" :class="{ 'is-expanded': expanded }"><ArrowRight /></el-icon>
    </div>

    <div v-show="expanded" class="sql-block__body">
      <div v-if="!steps.length" class="sql-block__empty">本次未产出 SQL（可能是试算或未命中）</div>
      <div v-for="(step, i) in steps" :key="i" class="sql-block__step">
        <div class="sql-block__step-header">
          <span class="sql-block__purpose">{{ step.purpose }}</span>
          <el-tag v-if="step.datasource_label" size="small" type="info" effect="plain">
            {{ step.datasource_label }}
          </el-tag>
          <ElButton
            class="sql-block__copy"
            link
            size="small"
            :icon="CopyDocument"
            @click.stop="onCopy(step.sql)"
          >
            复制
          </ElButton>
        </div>
        <pre class="sql-block__code"><code>{{ step.sql }}</code></pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sql-block {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
  margin: 0;
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
</style>
