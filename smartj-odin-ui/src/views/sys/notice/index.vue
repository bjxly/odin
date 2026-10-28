<template>
  <div class="sm-page">
    <!-- 检索区：通道 + 阅读状态筛选 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item :label="$t('sysNotice.channel')">
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
      <el-form-item :label="$t('sysNotice.status')">
        <el-select
          v-model="queryForm.readStatus"
          :placeholder="$t('sysNotice.all')"
          clearable
          style="width: 120px"
        >
          <el-option
            :label="$t('sysNotice.unread')"
            :value="0"
          />
          <el-option
            :label="$t('sysNotice.read')"
            :value="1"
          />
        </el-select>
      </el-form-item>
    </SmSearchForm>

    <!-- 表格区 -->
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
      <template #title>{{ $t('sysNotice.myMessages') }}</template>
      <template #header-actions>
        <el-button
          type="primary"
          :disabled="unreadCount === 0"
          @click="handleMarkAllRead"
        >
          {{ $t('sysNotice.markAllRead') }}
        </el-button>
      </template>
      <template #readStatus="{ row }">
        <el-tag
          :type="row.readStatus === 1 ? 'success' : 'warning'"
          size="small"
        >
          {{ row.readStatus === 1 ? $t('sysNotice.read') : $t('sysNotice.unread') }}
        </el-tag>
      </template>
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
          {{ row.channel === 'EMAIL' ? $t('sysNotice.email') : row.channel === 'SMS' ? $t('sysNotice.sms') : row.channel === 'SITE' ? $t('sysNotice.site') : row.channel }}
        </el-tag>
      </template>
      <template #operate="{ row }">
        <el-button
          v-if="row.readStatus === 0"
          link
          type="primary"
          @click="handleMarkRead(row)"
        >
          {{ $t('sysNotice.markRead') }}
        </el-button>
      </template>
    </SmCrudTable>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import { getNoticePage, getUnreadCount, markNoticeRead, markAllNoticeRead } from '@/api/sys/notice'
import type { NoticeLogVO } from '@/types/sys/notice'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const { t } = useI18n()

// 查询参数
const queryForm = reactive({
  current: 1,
  size: 20,
  channel: '',
  readStatus: undefined as number | undefined,
  userId: userStore.userInfo?.id || '',
})

// 表格数据
const tableData = ref<NoticeLogVO[]>([])
const total = ref(0)
const loading = ref(false)
const unreadCount = ref(0)

// 表格列定义
const columns = [
  { prop: 'title', label: t('sysNotice.title'), minWidth: 200 },
  { prop: 'channel', label: t('sysNotice.channel'), width: 100 },
  { prop: 'content', label: t('sysNotice.content'), minWidth: 300, showOverflowTooltip: true },
  { prop: 'sendStatus', label: t('sysNotice.sendStatus'), width: 130 },
  { prop: 'readStatus', label: t('sysNotice.readStatus'), width: 130 },
  { prop: 'receiver', label: t('sysNotice.receiver'), width: 180 },
  { prop: 'createTime', label: t('sysNotice.sendTime'), width: 170 },
]

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      pageNumber: queryForm.current,
      pageSize: queryForm.size,
      userId: queryForm.userId,
    }
    if (queryForm.channel) params.channel = queryForm.channel
    if (queryForm.readStatus !== undefined) params.readStatus = queryForm.readStatus

    const res = await getNoticePage(params)
    tableData.value = res.records
    total.value = res.total
  } finally {
    loading.value = false
  }
}

// 查询
function handleQuery() {
  queryForm.current = 1
  loadData()
}

// 重置
function handleReset() {
  queryForm.channel = ''
  queryForm.readStatus = undefined
  queryForm.current = 1
  loadData()
}

// 标记单条已读
async function handleMarkRead(row: NoticeLogVO) {
  try {
    await markNoticeRead(row.id)
    ElMessage.success(t('sysNotice.markReadSuccess'))
    loadData()
    loadUnreadCount()
  } catch {
    ElMessage.error(t('sysNotice.operateFailed'))
  }
}

// 标记全部已读
async function handleMarkAllRead() {
  try {
    await ElMessageBox.confirm(t('sysNotice.markAllReadConfirm'), t('common.tip'), { type: 'info' })
    await markAllNoticeRead(queryForm.userId)
    ElMessage.success(t('sysNotice.markAllReadSuccess'))
    loadData()
    loadUnreadCount()
  } catch {
    // 取消操作
  }
}

// 获取未读数量
async function loadUnreadCount() {
  const res = await getUnreadCount(queryForm.userId)
  unreadCount.value = res || 0
}

onMounted(() => {
  loadData()
  loadUnreadCount()
})
</script>
