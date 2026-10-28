<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sysOperLog.module')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.module"
          :placeholder="$t('common.pleaseInput') + $t('sysOperLog.module')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysOperLog.operatorName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.operatorName"
          :placeholder="$t('common.pleaseInput') + $t('sysOperLog.operatorName')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysOperLog.createTime')"
        label-width="auto"
      >
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          :range-separator="$t('sysOperLog.to')"
          :start-placeholder="$t('sysOperLog.startTime')"
          :end-placeholder="$t('sysOperLog.endTime')"
          value-format="YYYY-MM-DD HH:mm:ss"
          clearable
          :disabled-date="disableDate"
        />
      </el-form-item>
    </SmSearchForm>

    <!-- 表格区 -->
    <div class="sm-card">
      <div class="sm-table-header">
        <span class="sm-table-title">{{ $t('sysOperLog.title') }}</span>
        <div class="sm-table-header-actions">
          <el-button-group>
            <el-button
              v-permission="'sys:log:delete'"
              type="danger"
              @click="handleClean"
            >
              {{ $t('sysOperLog.clean') }}
            </el-button>
          </el-button-group>
        </div>
      </div>
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
      >
        <el-table-column
          type="index"
          label="#"
          width="60"
          align="center"
        />
        <el-table-column
          prop="module"
          :label="$t('sysOperLog.module')"
          min-width="120"
        />
        <el-table-column
          prop="action"
          :label="$t('sysOperLog.action')"
          width="120"
          align="center"
        />
        <el-table-column
          prop="requestMethod"
          :label="$t('sysOperLog.requestMethod')"
          width="140"
          align="center"
        >
          <template #default="{ row }">
            <el-tag
              v-if="row.requestMethod"
              :type="methodTagType(row.requestMethod)"
              size="small"
            >
              {{ row.requestMethod }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="operatorName"
          :label="$t('sysOperLog.operatorName')"
          width="120"
          align="center"
        />
        <el-table-column
          prop="operatorIp"
          :label="$t('sysOperLog.operatorIp')"
          width="140"
          align="center"
        />
        <el-table-column
          prop="responseCode"
          :label="$t('sysOperLog.responseCode')"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.responseCode === 200 ? 'success' : 'danger'"
              size="small"
            >
              {{ row.responseCode === 200 ? $t('sysOperLog.success') : $t('sysOperLog.fail') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="duration"
          :label="$t('sysOperLog.duration')"
          width="120"
          align="right"
        />
        <el-table-column
          prop="createTime"
          :label="$t('sysOperLog.createTime')"
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
      :title="$t('sysOperLog.detailTitle')"
      width="700px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-descriptions
        :column="2"
        border
      >
        <el-descriptions-item :label="$t('sysOperLog.module')">{{ detailRow?.module }}</el-descriptions-item>
        <el-descriptions-item :label="$t('sysOperLog.action')">{{ detailRow?.action }}</el-descriptions-item>
        <el-descriptions-item :label="$t('sysOperLog.requestMethod')">{{ detailRow?.requestMethod }}</el-descriptions-item>
        <el-descriptions-item :label="$t('sysOperLog.operationStatus')">
          <el-tag
            :type="detailRow?.responseCode === 200 ? 'success' : 'danger'"
            size="small"
          >
            {{ detailRow?.responseCode === 200 ? $t('sysOperLog.success') : $t('sysOperLog.fail') }}（{{ detailRow?.responseCode }}）
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('sysOperLog.operatorName')">{{ detailRow?.operatorName }}</el-descriptions-item>
        <el-descriptions-item :label="$t('sysOperLog.operatorIp')">{{ detailRow?.operatorIp }}</el-descriptions-item>
        <el-descriptions-item :label="$t('sysOperLog.cost')">{{ detailRow?.duration }} ms</el-descriptions-item>
        <el-descriptions-item :label="$t('sysOperLog.traceId')">{{ detailRow?.traceId || '-' }}</el-descriptions-item>
        <el-descriptions-item
          :label="$t('sysOperLog.requestUrl')"
          :span="2"
        >
          {{ detailRow?.requestUrl }}
        </el-descriptions-item>
        <el-descriptions-item
          :label="$t('sysOperLog.method')"
          :span="2"
        >
          {{ detailRow?.method }}
        </el-descriptions-item>
        <el-descriptions-item
          :label="$t('sysOperLog.createTime')"
          :span="2"
        >
          {{ detailRow?.createTime }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOperLogPage, cleanOperLog } from '@/api/log/oper'
import type { OperLogVO } from '@/types/sys'

const { t } = useI18n()
const loading = ref(false)
const tableData = ref<OperLogVO[]>([])
const total = ref(0)
const detailVisible = ref(false)
const detailRow = ref<OperLogVO | null>(null)
const dateRange = ref<[string, string] | null>(null)

const queryForm = reactive({
  current: 1,
  size: 10,
  module: '',
  operatorName: '',
  startTime: '',
  endTime: '',
})

watch(dateRange, (val) => {
  if (val) {
    let satrtTime = val[0].split(' ')[0]
    queryForm.startTime = satrtTime + ' 00:00:00'
    // 对操作时间进行过滤，结束时间时分秒为 23:59:59
    let endDate = val[1].split(' ')[0]
    queryForm.endTime = endDate + ' 23:59:59'
  } else {
    queryForm.startTime = ''
    queryForm.endTime = ''
  }
})

watch(() => [queryForm.current, queryForm.size], loadData)

function methodTagType(method: string) {
  const map: Record<string, string> = {
    GET: 'success',
    POST: '',
    PUT: 'warning',
    DELETE: 'danger',
  }
  return map[method] || 'info'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getOperLogPage(queryForm)
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
  queryForm.module = ''
  queryForm.operatorName = ''
  dateRange.value = null
  handleQuery()
}

function handleDetail(row: OperLogVO) {
  detailRow.value = row
  detailVisible.value = true
}

async function handleClean() {
  await ElMessageBox.confirm(t('sysOperLog.cleanConfirm'), t('common.tip'), { type: 'warning' })
  await cleanOperLog(0)
  ElMessage.success(t('sysOperLog.cleanSuccess'))
  loadData()
}

function disableDate(time: Date) {
  return time.getTime() > Date.now()
}

onMounted(() => {
  loadData()
})
</script>
