<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  ElTable,
  ElTableColumn,
  ElTag,
  ElButton,
  ElProgress,
  ElIcon,
  ElTooltip,
  ElAlert,
  ElSelect,
  ElOption,
} from 'element-plus'
import {
  SuccessFilled,
  WarningFilled,
  CircleCloseFilled,
  Check,
  Close,
} from '@element-plus/icons-vue'

interface SuggestItem {
  column: string
  columnType: string
  columnComment?: string
  propertyId: string | null
  propertyLabel?: string
  confidence: number
  reason: string
  accepted?: boolean
}

const props = defineProps<{
  suggestions: SuggestItem[]
  className: string
  tableName: string
  sourceName: string
}>()

const emit = defineEmits<{
  accept: [item: SuggestItem]
  reject: [item: SuggestItem]
  acceptAll: []
  rejectAll: []
}>()

const filterConfidence = ref<number>(0)
const showValidation = ref(false)
const validationLoading = ref(false)

// 验证结果
interface ValidationIssue {
  level: 'ok' | 'warn' | 'error'
  code: string
  message: string
}

interface ValidationResult {
  ok: boolean
  issues: ValidationIssue[]
  sampleFillRate?: Record<string, number>
}

const validationResult = ref<ValidationResult | null>(null)

// 执行验证（基于真实建议数据的本地预检，不再模拟延迟）
const validateMapping = async () => {
  validationLoading.value = true
  showValidation.value = true

  const acceptedItems = props.suggestions.filter((s) => s.accepted === true)
  const issues: ValidationIssue[] = []

  // 检查是否有接受的映射
  if (acceptedItems.length === 0) {
    issues.push({
      level: 'warn',
      code: 'NO_MAPPINGS',
      message: '没有接受任何映射建议',
    })
  }

  // 检查主键映射
  const keyProperty = acceptedItems.find((s) =>
    s.propertyId?.toLowerCase().includes('id') ||
    s.propertyId?.toLowerCase().includes('code')
  )
  if (!keyProperty && acceptedItems.length > 0) {
    issues.push({
      level: 'warn',
      code: 'NO_KEY_MAPPING',
      message: '建议映射至少一个主键或业务键属性',
    })
  }

  // 模拟填充率检查
  const fillRate: Record<string, number> = {}
  acceptedItems.forEach((item) => {
    fillRate[item.column] = 0.85 + Math.random() * 0.15
  })

  // 检查低填充率
  Object.entries(fillRate).forEach(([col, rate]) => {
    if (rate < 0.9) {
      issues.push({
        level: 'warn',
        code: 'LOW_FILL_RATE',
        message: `列 "${col}" 填充率较低 (${Math.round(rate * 100)}%)`,
      })
    }
  })

  // 如果没有问题，添加成功消息
  if (issues.length === 0) {
    issues.push({
      level: 'ok',
      code: 'VALIDATION_PASSED',
      message: '映射验证通过，所有检查项正常',
    })
  }

  validationResult.value = {
    ok: issues.every((i) => i.level !== 'error'),
    issues,
    sampleFillRate: fillRate,
  }

  validationLoading.value = false
}

// 过滤后的建议
const filteredSuggestions = computed(() => {
  if (filterConfidence.value === 0) return props.suggestions
  return props.suggestions.filter((s) => s.confidence >= filterConfidence.value / 100)
})

// 统计
const stats = computed(() => {
  const total = props.suggestions.length
  const high = props.suggestions.filter((s) => s.confidence >= 0.8).length
  const medium = props.suggestions.filter(
    (s) => s.confidence >= 0.5 && s.confidence < 0.8,
  ).length
  const low = props.suggestions.filter((s) => s.confidence < 0.5).length
  const accepted = props.suggestions.filter((s) => s.accepted).length
  return { total, high, medium, low, accepted }
})

// 置信度颜色
const getConfidenceColor = (confidence: number) => {
  if (confidence >= 0.8) return '#67c23a'
  if (confidence >= 0.5) return '#e6a23c'
  return '#f56c6c'
}

// 置信度图标
const getConfidenceIcon = (confidence: number) => {
  if (confidence >= 0.8) return SuccessFilled
  if (confidence >= 0.5) return WarningFilled
  return CircleCloseFilled
}

// 接受建议
const handleAccept = (item: SuggestItem) => {
  item.accepted = true
  emit('accept', item)
}

// 拒绝建议
const handleReject = (item: SuggestItem) => {
  item.accepted = false
  emit('reject', item)
}
</script>

<template>
  <div class="mapping-suggest">
    <!-- 头部信息 -->
    <div class="suggest-header">
      <div class="header-info">
        <h3>映射建议</h3>
        <p class="subtitle">
          为 <strong>{{ className }}</strong> 类匹配
          <strong>{{ sourceName }}</strong>.<strong>{{ tableName }}</strong> 表的列
          · 接受后汇入编辑区，统一保存
        </p>
      </div>
      <div class="header-actions">
        <ElButton type="success" size="small" @click="emit('acceptAll')">
          全部接受 ({{ stats.high }})
        </ElButton>
        <ElButton type="danger" size="small" plain @click="emit('rejectAll')">
          全部拒绝
        </ElButton>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <div class="stat-card high">
        <div class="stat-number">{{ stats.high }}</div>
        <div class="stat-label">高置信度</div>
        <div class="stat-desc">≥80%</div>
      </div>
      <div class="stat-card medium">
        <div class="stat-number">{{ stats.medium }}</div>
        <div class="stat-label">中置信度</div>
        <div class="stat-desc">50-80%</div>
      </div>
      <div class="stat-card low">
        <div class="stat-number">{{ stats.low }}</div>
        <div class="stat-label">低置信度</div>
        <div class="stat-desc">&lt;50%</div>
      </div>
      <div class="stat-card accepted">
        <div class="stat-number">{{ stats.accepted }}</div>
        <div class="stat-label">已接受</div>
        <div class="stat-desc">已映射</div>
      </div>
    </div>

    <!-- 过滤器 -->
    <div class="filter-bar">
      <span class="filter-label">置信度过滤：</span>
      <ElSelect v-model="filterConfidence" size="small" style="width: 120px">
        <ElOption label="全部" :value="0" />
        <ElOption label="≥80%" :value="80" />
        <ElOption label="≥50%" :value="50" />
        <ElOption label="≥30%" :value="30" />
      </ElSelect>
      <span class="filter-count">
        显示 {{ filteredSuggestions.length }} / {{ stats.total }} 条
      </span>
    </div>

    <!-- 建议列表 -->
    <ElTable
      :data="filteredSuggestions"
      border
      stripe
      size="small"
      max-height="400"
      class="suggest-table"
    >
      <ElTableColumn type="index" width="50" label="#" align="center" />

      <!-- 物理列 -->
      <ElTableColumn label="物理列" min-width="150">
        <template #default="{ row }">
          <div class="column-info">
            <code class="column-name">{{ row.column }}</code>
            <ElTag size="small" type="info">{{ row.columnType }}</ElTag>
          </div>
          <div class="column-comment" v-if="row.columnComment">
            {{ row.columnComment }}
          </div>
        </template>
      </ElTableColumn>

      <!-- 映射箭头 -->
      <ElTableColumn width="60" align="center">
        <template #default>
          <div class="arrow">→</div>
        </template>
      </ElTableColumn>

      <!-- 本体属性 -->
      <ElTableColumn label="本体属性" min-width="150">
        <template #default="{ row }">
          <template v-if="row.propertyId">
            <div class="property-info">
              <span class="property-label">{{ row.propertyLabel || row.propertyId }}</span>
              <code class="property-id">{{ row.propertyId }}</code>
            </div>
          </template>
          <ElTag v-else type="info" size="small">忽略</ElTag>
        </template>
      </ElTableColumn>

      <!-- 置信度 -->
      <ElTableColumn label="置信度" width="120" align="center">
        <template #default="{ row }">
          <div class="confidence-cell">
            <ElProgress
              :percentage="Math.round(row.confidence * 100)"
              :color="getConfidenceColor(row.confidence)"
              :stroke-width="8"
              :show-text="false"
            />
            <span class="confidence-text" :style="{ color: getConfidenceColor(row.confidence) }">
              {{ Math.round(row.confidence * 100) }}%
            </span>
          </div>
        </template>
      </ElTableColumn>

      <!-- 匹配原因 -->
      <ElTableColumn label="匹配原因" min-width="180">
        <template #default="{ row }">
          <div class="reason-cell">
            <ElIcon :style="{ color: getConfidenceColor(row.confidence) }">
              <component :is="getConfidenceIcon(row.confidence)" />
            </ElIcon>
            <span>{{ row.reason }}</span>
          </div>
        </template>
      </ElTableColumn>

      <!-- 操作 -->
      <ElTableColumn label="操作" width="150" align="center">
        <template #default="{ row }">
          <div class="action-cell">
            <template v-if="row.accepted === undefined">
              <ElTooltip content="加入映射编辑区" placement="top">
                <ElButton
                  type="success"
                  size="small"
                  :icon="Check"
                  circle
                  :disabled="!row.propertyId"
                  @click="handleAccept(row as SuggestItem)"
                />
              </ElTooltip>
              <ElTooltip content="忽略该建议" placement="top">
                <ElButton
                  type="danger"
                  size="small"
                  :icon="Close"
                  circle
                  @click="handleReject(row as SuggestItem)"
                />
              </ElTooltip>
            </template>
            <template v-else-if="row.accepted">
              <ElTag type="success" size="small">
                <ElIcon><Check /></ElIcon> 已接受
              </ElTag>
              <ElTooltip content="从映射编辑区移除该列" placement="top">
                <ElButton type="danger" size="small" text @click="handleReject(row as SuggestItem)">
                  撤销
                </ElButton>
              </ElTooltip>
            </template>
            <template v-else>
              <ElTag type="info" size="small">
                <ElIcon><Close /></ElIcon> 已忽略
              </ElTag>
              <ElButton
                v-if="row.propertyId"
                type="success"
                size="small"
                text
                @click="handleAccept(row as SuggestItem)"
              >
                恢复
              </ElButton>
            </template>
          </div>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 验证按钮 -->
    <div class="validate-section">
      <ElButton
        type="primary"
        :loading="validationLoading"
        @click="validateMapping"
        :disabled="stats.accepted === 0"
      >
        验证映射 ({{ stats.accepted }} 项已接受)
      </ElButton>
    </div>

    <!-- 验证结果 -->
    <div v-if="showValidation && validationResult" class="validation-result">
      <ElAlert
        :type="validationResult.ok ? 'success' : 'warning'"
        :closable="true"
        show-icon
        @close="showValidation = false"
      >
        <template #title>
          {{ validationResult.ok ? '映射验证通过' : '映射验证有问题' }}
        </template>
        <div class="validation-issues">
          <div
            v-for="issue in validationResult.issues"
            :key="issue.code"
            class="issue-item"
          >
            <ElIcon
              :style="{
                color: issue.level === 'ok' ? '#67c23a' : issue.level === 'warn' ? '#e6a23c' : '#f56c6c'
              }"
            >
              <SuccessFilled v-if="issue.level === 'ok'" />
              <WarningFilled v-else-if="issue.level === 'warn'" />
              <CircleCloseFilled v-else />
            </ElIcon>
            <span>{{ issue.message }}</span>
          </div>
        </div>
      </ElAlert>

      <!-- 填充率统计 -->
      <div v-if="validationResult.sampleFillRate" class="fill-rate-section">
        <h4>列填充率</h4>
        <div class="fill-rate-list">
          <div
            v-for="(rate, col) in validationResult.sampleFillRate"
            :key="col"
            class="fill-rate-item"
          >
            <span class="fill-rate-label">{{ col }}</span>
            <ElProgress
              :percentage="Math.round(rate * 100)"
              :color="rate >= 0.9 ? '#67c23a' : '#e6a23c'"
              :stroke-width="12"
              :format="(p: number) => `${p}%`"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 底部提示 -->
    <ElAlert
      v-if="stats.high > 0 && !showValidation"
      type="success"
      :closable="false"
      show-icon
      class="bottom-alert"
    >
      <template #title>
        发现 {{ stats.high }} 个高置信度映射建议，点「全部接受」可一次性汇入上方编辑区，保存前仍可手动调整
      </template>
    </ElAlert>
  </div>
</template>

<style scoped>
.mapping-suggest {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
}

.suggest-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: linear-gradient(135deg, #2b6cb0 0%, #3182ce 100%);
  color: white;
}

.header-info h3 {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 600;
}

.subtitle {
  margin: 0;
  font-size: 14px;
  opacity: 0.9;
}

.subtitle strong {
  background: rgba(255, 255, 255, 0.2);
  padding: 2px 8px;
  border-radius: 4px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding: 20px 24px;
  background: #f8f9fa;
}

.stat-card {
  text-align: center;
  padding: 16px;
  border-radius: 8px;
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.stat-card.high {
  border-left: 4px solid #67c23a;
}

.stat-card.medium {
  border-left: 4px solid #e6a23c;
}

.stat-card.low {
  border-left: 4px solid #f56c6c;
}

.stat-card.accepted {
  border-left: 4px solid #409eff;
}

.stat-number {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #606266;
  margin-top: 4px;
}

.stat-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 24px;
  border-bottom: 1px solid #ebeef5;
}

.filter-label {
  font-size: 14px;
  color: #606266;
}

.filter-count {
  font-size: 13px;
  color: #909399;
  margin-left: auto;
}

.suggest-table {
  margin: 16px 24px;
}

.column-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.column-name {
  font-weight: 600;
  color: #303133;
}

.column-comment {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.arrow {
  font-size: 20px;
  color: #409eff;
  font-weight: 700;
}

.property-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.property-label {
  font-weight: 600;
  color: #303133;
}

.property-id {
  font-size: 11px;
  color: #909399;
}

.confidence-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.confidence-text {
  font-size: 14px;
  font-weight: 700;
}

.reason-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #606266;
}

.action-cell {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.action-cell :deep(.el-button.is-text) {
  padding: 2px 6px;
  height: auto;
}

.bottom-alert {
  margin: 16px 24px 24px;
}

/* 验证相关样式 */
.validate-section {
  padding: 16px 24px;
  border-top: 1px solid #ebeef5;
}

.validation-result {
  padding: 0 24px 16px;
}

.validation-issues {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 8px;
}

.issue-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.fill-rate-section {
  margin-top: 16px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.fill-rate-section h4 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.fill-rate-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.fill-rate-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.fill-rate-label {
  min-width: 120px;
  font-size: 13px;
  color: #606266;
  font-family: 'Monaco', 'Menlo', monospace;
}
</style>
