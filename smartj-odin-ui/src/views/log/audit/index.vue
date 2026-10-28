<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <div class="sm-card">
      <div class="sm-query-container">
        <el-form
          :model="queryForm"
          inline
        >
          <el-form-item :label="$t('logAudit.target')">
            <el-input
              v-model="queryForm.target"
              :placeholder="$t('common.pleaseInput') + $t('logAudit.target')"
              clearable
              style="width: 160px"
            />
          </el-form-item>
          <el-form-item :label="$t('logAudit.operType')">
            <el-input
              v-model="queryForm.action"
              :placeholder="$t('common.pleaseInput') + $t('logAudit.operType')"
              clearable
              style="width: 160px"
            />
          </el-form-item>
          <el-form-item :label="$t('logAudit.operName')">
            <el-input
              v-model="queryForm.operatorName"
              :placeholder="$t('common.pleaseInput') + $t('logAudit.operName')"
              clearable
              style="width: 160px"
            />
          </el-form-item>
          <el-form-item :label="$t('logAudit.operTime')">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="-"
              :start-placeholder="$t('logAudit.startDate')"
              :end-placeholder="$t('logAudit.endDate')"
              value-format="YYYY-MM-DD HH:mm:ss"
              clearable
              style="width: 260px"
            />
          </el-form-item>
        </el-form>
        <div class="sm-query-actions">
          <el-button-group>
            <el-button
              type="primary"
              @click="handleQuery"
            >
              {{ $t('common.query') }}
            </el-button>
            <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
          </el-button-group>
        </div>
      </div>
    </div>

    <!-- 表格区 -->
    <div class="sm-card">
      <div class="sm-table-header">
        <span class="sm-table-title">{{ $t('logAudit.title') }}</span>
      </div>
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
      >
        <el-table-column
          type="index"
          :label="$t('common.index')"
          width="60"
          align="center"
        />
        <el-table-column
          prop="traceId"
          :label="$t('logAudit.traceId')"
          width="180"
        />
        <el-table-column
          prop="target"
          :label="$t('logAudit.target')"
          min-width="120"
        />
        <el-table-column
          prop="targetId"
          :label="$t('logAudit.targetId')"
        />
        <el-table-column
          prop="action"
          :label="$t('logAudit.operType')"
          width="150"
          align="center"
        />
        <el-table-column
          prop="operatorName"
          :label="$t('logAudit.operName')"
          width="120"
          align="center"
        />
        <el-table-column
          prop="createTime"
          :label="$t('logAudit.operTime')"
          width="170"
          align="center"
        />
        <el-table-column
          :label="$t('common.operation')"
          width="100"
          fixed="right"
          align="center"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="handleDetail(row)"
            >
              {{ $t('common.detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="sm-table-pagination">
        <el-pagination
          v-model:current-page="queryForm.current"
          v-model:page-size="queryForm.size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
        />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      :title="$t('logAudit.detailTitle')"
      width="750px"
      destroy-on-close
    >
      <el-descriptions
        :column="2"
        border
      >
        <el-descriptions-item :label="$t('logAudit.traceId')">{{ detailRow?.traceId || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('logAudit.operName')">{{ detailRow?.operatorName }}</el-descriptions-item>
        <el-descriptions-item :label="$t('logAudit.target')">{{ detailRow?.target }}</el-descriptions-item>
        <el-descriptions-item :label="$t('logAudit.targetId')">{{ detailRow?.targetId || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('logAudit.operType')">{{ detailRow?.action }}</el-descriptions-item>
        <el-descriptions-item :label="$t('logAudit.operTime')">{{ detailRow?.createTime }}</el-descriptions-item>
      </el-descriptions>
      <div
        v-if="detailRow?.beforeData"
        style="margin-top: 16px"
      >
        <h4 style="margin-bottom: 8px; font-weight: 600">{{ $t('logAudit.beforeData') }}</h4>
        <pre class="sm-log-pre">{{ formatJson(detailRow.beforeData) }}</pre>
      </div>
      <div
        v-if="detailRow?.afterData"
        style="margin-top: 16px"
      >
        <h4 style="margin-bottom: 8px; font-weight: 600">{{ $t('logAudit.afterData') }}</h4>
        <pre class="sm-log-pre">{{ formatJson(detailRow.afterData) }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getAuditLogPage } from '@/api/log/audit'
import type { AuditLogVO } from '@/types/sys'

useI18n()
const loading = ref(false)
const tableData = ref<AuditLogVO[]>([])
const total = ref(0)
const detailVisible = ref(false)
const detailRow = ref<AuditLogVO | null>(null)
const dateRange = ref<[string, string] | null>(null)

const queryForm = reactive({
  current: 1,
  size: 10,
  target: '',
  action: '',
  operatorName: '',
  startTime: '',
  endTime: '',
})

watch(dateRange, (val) => {
  if (val) {
    queryForm.startTime = val[0]
    queryForm.endTime = val[1]
  } else {
    queryForm.startTime = ''
    queryForm.endTime = ''
  }
})

watch(() => [queryForm.current, queryForm.size], loadData)

function formatJson(str: string): string {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await getAuditLogPage(queryForm)
    tableData.value = res.records
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function handleQuery() {
  queryForm.current = 1
  loadData()
}

function handleReset() {
  queryForm.target = ''
  queryForm.action = ''
  queryForm.operatorName = ''
  dateRange.value = null
  handleQuery()
}

function handleDetail(row: AuditLogVO) {
  detailRow.value = row
  detailVisible.value = true
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.sm-log-pre {
  background-color: var(--sm-bg-page, #f5f7fa);
  border: 1px solid var(--sm-border-lighter, #ebeef5);
  border-radius: 4px;
  padding: 12px;
  max-height: 300px;
  overflow: auto;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
