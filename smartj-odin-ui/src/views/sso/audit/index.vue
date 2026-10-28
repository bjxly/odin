<template>
  <div class="sm-page">
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sso.eventType')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.eventType"
          :placeholder="$t('common.pleaseInput') + $t('sso.eventType')"
          clearable
        />
      </el-form-item>
      <el-form-item :label="$t('sso.username')">
        <el-input
          v-model="queryForm.userName"
          :placeholder="$t('common.pleaseInput') + $t('sso.username')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sso.timeRange')"
        label-width="auto"
      >
        <el-date-picker
          v-model="queryForm.dateRange"
          type="daterange"
          :range-separator="$t('sso.to')"
          :start-placeholder="$t('sso.start')"
          :end-placeholder="$t('sso.end')"
          value-format="YYYY-MM-DD"
          :disabled-date="disableDate"
          style="width: 260px"
        />
      </el-form-item>
    </SmSearchForm>

    <SmCrudTable
      v-model:page="queryForm.current"
      v-model:page-size="queryForm.size"
      :data="tableData"
      :columns="columns"
      :total="total"
      :loading="loading"
      :show-index="true"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sso.auditTitle') }}</template>
      <template #header-actions>
        <el-button
          v-if="false"
          v-permission="'sso:audit:export'"
          @click="handleExport"
        >
          {{ $t('sso.export') }}
        </el-button>
      </template>
      <template #eventCategory="{ row }">
        <el-tag
          size="small"
          :type="categoryTagType(row.eventCategory)"
        >
          {{ categoryLabel(row.eventCategory) }}
        </el-tag>
      </template>
      <template #result="{ row }">
        <el-tag
          size="small"
          :type="row.result === 'SUCCESS' ? 'success' : 'danger'"
        >
          {{ row.result === 'SUCCESS' ? $t('sso.success') : $t('sso.failed') }}
        </el-tag>
      </template>
    </SmCrudTable>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getSsoAuditLogs } from '@/api/sso'
import type { SsoAuditLogVO } from '@/types/sso'

const loading = ref(false)
const { t } = useI18n()
const tableData = ref<SsoAuditLogVO[]>([])
const total = ref(0)
const queryForm = reactive({
  eventType: '',
  userName: '' as string,
  dateRange: null as [string, string] | null,
  current: 1,
  size: 20,
})

const columns = computed<TableColumn[]>(() => [
  { prop: 'eventType', label: t('sso.eventType'), width: 140, align: 'left' },
  { prop: 'eventCategory', label: t('sso.eventCategory'), width: 140, slot: 'eventCategory' },
  { prop: 'username', label: t('sso.operatorName'), width: 100, align: 'left' },
  { prop: 'action', label: t('sysOperLog.action'), minWidth: 150, align: 'left' },
  { prop: 'result', label: t('sso.result'), width: 100, slot: 'result' },
  { prop: 'resultMessage', label: t('sso.resultMessage'), minWidth: 180, align: 'left' },
  { prop: 'operatorIp', label: t('sso.operatorIp'), width: 130, align: 'right' },
  { prop: 'eventTime', label: t('sso.eventTime'), width: 170 },
])

function categoryTagType(cat: string) {
  const map: Record<string, string> = {
    AUTHENTICATION: '',
    AUTHORIZATION: 'warning',
    ADMIN: 'info',
    ANOMALY: 'danger',
  }
  return map[cat] || ''
}

function categoryLabel(cat: string) {
  const map: Record<string, string> = {
    AUTHENTICATION: t('sso.authentication'),
    AUTHORIZATION: t('sso.authorization'),
    ADMIN: t('sso.admin'),
    ANOMALY: t('sso.anomaly'),
  }
  return map[cat] || cat
}

async function loadData() {
  loading.value = true
  try {
    // 对操作时间进行过滤，结束时间时分秒为 23:59:59
    let startTime = undefined
    let endTime = undefined
    if (queryForm.dateRange && queryForm.dateRange.length === 2) {
      let startDate = queryForm.dateRange[0].split(' ')[0]
      startTime = startDate + ' 00:00:00'
      let endDate = queryForm.dateRange[1].split(' ')[0]
      endTime = endDate + ' 23:59:59'
    }

    const res = await getSsoAuditLogs({
      page: queryForm.current,
      size: queryForm.size,
      eventType: queryForm.eventType || undefined,
      userName: queryForm.userName,
      start: startTime,
      end: endTime,
    })
    tableData.value = res.records || []
    total.value = Number(res.totalRow)
  } catch {
    tableData.value = []
  } finally {
    loading.value = false
  }
}

function handleQuery() {
  queryForm.current = 1
  loadData()
}
function handleReset() {
  queryForm.eventType = ''
  queryForm.userName = ''
  queryForm.dateRange = null
  handleQuery()
}

function disableDate(time: Date) {
  return time.getTime() > Date.now()
}

function handleExport() {
  ElMessage.info(t('sso.exportDeveloping'))
}

onMounted(() => loadData())
</script>
