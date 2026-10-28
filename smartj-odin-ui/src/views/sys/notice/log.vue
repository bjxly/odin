<template>
  <div class="sm-page">
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item :label="$t('sysNoticeLog.channel')">
        <el-select
          v-model="queryForm.channel"
          :placeholder="$t('sysNotice.all')"
          clearable
          style="width: 140px"
        >
          <el-option
            :label="$t('sysNotice.email')"
            value="EMAIL"
          />
          <el-option
            :label="$t('sysNotice.sms')"
            value="SMS"
          />
          <el-option
            :label="$t('sysNotice.site')"
            value="SITE"
          />
          <el-option
            :label="$t('sysNotice.im')"
            value="IM"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('sysNoticeLog.sendStatus')">
        <el-select
          v-model="queryForm.sendStatus"
          :placeholder="$t('sysNotice.all')"
          clearable
          style="width: 120px"
        >
          <el-option
            :label="$t('sysNotice.success')"
            :value="1"
          />
          <el-option
            :label="$t('sysNotice.fail')"
            :value="2"
          />
        </el-select>
      </el-form-item>
    </SmSearchForm>

    <SmCrudTable
      v-model:page="queryForm.current"
      v-model:page-size="queryForm.size"
      :data="tableData"
      :columns="columns"
      :total="total"
      :loading="loading"
      :show-selection="false"
      :show-index="true"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sysNoticeLog.title') }}</template>
      <template #sendStatus="{ row }">
        <el-tag
          :type="row.sendStatus === 1 ? 'success' : row.sendStatus === 2 ? 'danger' : 'info'"
          size="small"
        >
          {{ row.sendStatus === 1 ? $t('sysNotice.success') : row.sendStatus === 2 ? $t('sysNotice.fail') : $t('sysNotice.pending') }}
        </el-tag>
      </template>
      <template #channel="{ row }">
        <el-tag size="small">
          {{
            row.channel === 'EMAIL'
              ? $t('sysNotice.email')
              : row.channel === 'SMS'
                ? $t('sysNotice.sms')
                : row.channel === 'SITE'
                  ? $t('sysNotice.site')
                  : row.channel === 'IM'
                    ? $t('sysNotice.im')
                    : row.channel
          }}
        </el-tag>
      </template>
      <template #errorMessage="{ row }">
        <span
          v-if="row.errorMessage"
          style="color: red; font-size: 12px"
        >
          {{ row.errorMessage.length > 50 ? row.errorMessage.substring(0, 50) + '...' : row.errorMessage }}
        </span>
        <span
          v-else
          style="color: #999"
        >
          -
        </span>
      </template>
      <template #operate="{ row }">
        <el-button
          v-if="row.sendStatus === 2"
          link
          type="primary"
          @click="handleResend(row)"
        >
          {{ $t('sysNoticeLog.resend') }}
        </el-button>
      </template>
    </SmCrudTable>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import { getNoticePage, sendNotice } from '@/api/sys/notice'
import type { NoticeLogVO } from '@/types/sys/notice'

const loading = ref(false)
const { t } = useI18n()
const tableData = ref<NoticeLogVO[]>([])
const total = ref(0)

const queryForm = reactive({
  current: 1,
  size: 20,
  channel: '',
  sendStatus: undefined as number | undefined,
})

const columns = [
  { prop: 'templateCode', label: t('sysNoticeLog.templateCode'), minWidth: 160 },
  { prop: 'channel', label: t('sysNoticeLog.channel'), width: 90 },
  { prop: 'receiver', label: t('sysNoticeLog.receiver'), minWidth: 180 },
  { prop: 'title', label: t('sysNoticeLog.messageTitle'), minWidth: 200 },
  { prop: 'sendStatus', label: t('sysNoticeLog.sendStatus'), width: 130 },
  { prop: 'providerId', label: t('sysNoticeLog.providerId'), width: 100 },
  { prop: 'errorMessage', label: t('sysNoticeLog.errorMessage'), minWidth: 150 },
  { prop: 'createTime', label: t('sysNoticeLog.sendTime'), width: 170 },
]

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      pageNumber: queryForm.current,
      pageSize: queryForm.size,
    }
    if (queryForm.channel) params.channel = queryForm.channel
    if (queryForm.sendStatus !== undefined) params.readStatus = undefined // 用 sendStatus 改为 filter 逻辑
    const res = await getNoticePage(params)
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
  queryForm.channel = ''
  queryForm.sendStatus = undefined
  queryForm.current = 1
  loadData()
}

async function handleResend(row: NoticeLogVO) {
  try {
    await sendNotice({
      templateCode: row.templateCode,
      channel: row.channel,
      receiver: row.receiver,
      userId: row.userId,
      variables: {},
      tenantId: row.tenantId,
    })
    ElMessage.success(t('sysNoticeLog.resendSuccess'))
    loadData()
  } catch {
    ElMessage.error(t('sysNoticeLog.resendFail'))
  }
}

onMounted(() => {
  loadData()
})
</script>
