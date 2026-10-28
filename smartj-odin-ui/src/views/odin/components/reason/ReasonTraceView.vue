<template>
  <div class="reason-trace-view">
    <el-collapse v-model="activeNames">
      <!-- 1. 触发规则 -->
      <el-collapse-item :title="t('odinReason.rulesFired')" name="rules">
        <el-table v-if="trace.rulesFired?.length" :data="trace.rulesFired" border size="small">
          <el-table-column prop="ruleName" :label="t('odinReason.ruleName')" min-width="120" align="left" />
          <el-table-column :label="t('odinReason.ruleType')" width="100" align="center">
            <template #default="{ row }">
              <el-tag size="small" :type="ruleTypeTag(row.ruleType)">{{ row.ruleType }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="t('odinReason.conditionResult')" width="90" align="center">
            <template #default="{ row }">
              <el-tag size="small" :type="row.conditionResult ? 'success' : 'info'">
                {{ row.conditionResult ? '✓' : '✗' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="action" :label="t('odinReason.action')" min-width="100" align="left" show-overflow-tooltip />
          <el-table-column prop="detail" :label="t('odinReason.detail')" min-width="140" align="left" show-overflow-tooltip />
        </el-table>
        <el-empty v-else :description="t('common.noData')" :image-size="40" />
      </el-collapse-item>

      <!-- 2. 虚拟类展开 -->
      <el-collapse-item :title="t('odinReason.expansions')" name="expansions">
        <div v-if="trace.expansions?.length" class="expansion-list">
          <div v-for="(step, i) in trace.expansions" :key="i" class="expansion-item">
            <el-tag size="small" effect="plain">{{ step.from }}</el-tag>
            <span class="expansion-arrow">→</span>
            <el-tag size="small" type="success" effect="plain">{{ step.to }}</el-tag>
            <span class="expansion-reason">{{ step.reason }}</span>
          </div>
        </div>
        <el-empty v-else :description="t('common.noData')" :image-size="40" />
      </el-collapse-item>

      <!-- 3. 派生过滤 -->
      <el-collapse-item :title="t('odinReason.derivedFilters')" name="filters">
        <el-table v-if="trace.derivedFilters?.length" :data="trace.derivedFilters" border size="small">
          <el-table-column prop="property" :label="t('odinMapping.propertyCode')" min-width="120" align="left" />
          <el-table-column prop="op" :label="t('odinReason.action')" width="80" align="center" />
          <el-table-column :label="'Value'" min-width="100" align="left">
            <template #default="{ row }">
              {{ row.value != null ? JSON.stringify(row.value) : '—' }}
            </template>
          </el-table-column>
          <el-table-column prop="source" :label="t('odinReason.detail')" min-width="120" align="left" show-overflow-tooltip />
        </el-table>
        <el-empty v-else :description="t('common.noData')" :image-size="40" />
      </el-collapse-item>

      <!-- 4. 推导关系 -->
      <el-collapse-item :title="t('odinReason.inferredRelations')" name="relations">
        <div v-if="trace.inferredRelations?.length" class="relation-list">
          <div v-for="(rel, i) in trace.inferredRelations" :key="i" class="relation-item">
            <el-tag size="small" effect="plain">{{ rel.fromClass }}</el-tag>
            <span class="relation-arrow">—[{{ rel.label || rel.name }}]→</span>
            <el-tag size="small" type="success" effect="plain">{{ rel.toClass }}</el-tag>
          </div>
        </div>
        <el-empty v-else :description="t('common.noData')" :image-size="40" />
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ReasonTraceResult } from '@/types/odin/reason'

defineProps<{
  trace: ReasonTraceResult
}>()

const { t } = useI18n()
const activeNames = ref(['rules', 'expansions', 'filters', 'relations'])

function ruleTypeTag(type: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  switch (type) {
    case 'derivation': return 'success'
    case 'constraint': return 'warning'
    case 'aggregation': return ''
    default: return 'info'
  }
}
</script>

<style scoped>
.reason-trace-view {
  width: 100%;
}

.expansion-list,
.relation-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.expansion-item,
.relation-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: var(--sm-font-size-sm, 13px);
}

.expansion-arrow,
.relation-arrow {
  color: var(--sm-text-secondary);
  font-weight: 500;
}

.expansion-reason {
  color: var(--sm-text-secondary);
  font-size: 12px;
  margin-left: 8px;
}
</style>
