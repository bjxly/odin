<template>
  <div class="sm-page">
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item :label="$t('sso.username')">
        <el-input
          v-model="queryForm.username"
          :placeholder="$t('common.pleaseInput') + $t('sso.username')"
          clearable
          style="width: 160px"
        />
      </el-form-item>
      <el-form-item
        :label="$t('sso.sessionClientName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.clientId"
          :placeholder="$t('common.pleaseInput') + $t('sso.clientId')"
          clearable
          style="width: 160px"
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
      :operate-width="100"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sso.sessionTitle') }}</template>
      <template #status="{ row }">
        <el-tag
          :type="row.status === 1 ? 'success' : 'info'"
          size="small"
        >
          {{ row.status === 1 ? $t('sso.online') : $t('sso.offline') }}
        </el-tag>
      </template>
      <template #loginTime="{ row }">
        {{ row.loginTime || '--' }}
      </template>
      <template #lastAccess="{ row }">
        {{ row.lastAccess || '--' }}
      </template>
      <template #expiresAt="{ row }">
        {{ row.expiresAt || '--' }}
      </template>
      <template #operate="{ row }">
        <el-button
          v-if="row.status === 1"
          v-permission="'sso:session:kick'"
          link
          type="danger"
          @click="handleKick(row)"
        >
          {{ $t('sso.kickOffline') }}
        </el-button>
        <span
          v-else
          style="color: #909399"
        >
          {{ $t('sso.offline') }}
        </span>
      </template>
    </SmCrudTable>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getSsoSessions, kickSsoSession } from '@/api/sso'
import type { SsoSessionVO } from '@/types/sso'

const loading = ref(false)
const { t } = useI18n()
const tableData = ref<SsoSessionVO[]>([])
const total = ref(0)
const queryForm = reactive({ username: '', clientId: '', current: 1, size: 10 })

const columns = computed<TableColumn[]>(() => [
  { prop: 'username', label: t('sso.username'), align: 'left' },
  { prop: 'clientId', label: t('sso.sessionClientName'), align: 'left' },
  { prop: 'loginIp', label: t('sso.loginIp'), width: 140, align: 'right' },
  { prop: 'loginTime', label: t('sso.loginTime'), width: 170, slot: 'loginTime' },
  { prop: 'lastAccess', label: t('sso.lastAccessTime'), width: 170, slot: 'lastAccess' },
  { prop: 'expiresAt', label: t('sso.expireTime'), width: 170, slot: 'expiresAt' },
  { prop: 'status', label: t('common.status'), width: 100, slot: 'status' },
])

async function loadData() {
  loading.value = true
  try {
    const res = await getSsoSessions({
      page: queryForm.current,
      size: queryForm.size,
      username: queryForm.username || undefined,
      clientId: queryForm.clientId || undefined,
    })
    tableData.value = res.records || []
    total.value = Number(res.totalRow)
  } catch {
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handleQuery() {
  queryForm.current = 1
  loadData()
}

function handleReset() {
  queryForm.username = ''
  queryForm.clientId = ''
  handleQuery()
}

async function handleKick(row: SsoSessionVO) {
  await ElMessageBox.confirm(t('sso.kickConfirm', { name: row.username }), t('common.tip'), { type: 'warning' })
  await kickSsoSession(row.id)
  ElMessage.success(t('sso.kickSuccess'))
  loadData()
}

onMounted(() => loadData())
</script>
