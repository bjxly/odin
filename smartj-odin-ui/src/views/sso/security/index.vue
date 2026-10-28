<template>
  <div class="sm-page">
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="t('sso.username')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.userName"
          :placeholder="t('sso.enterUsername')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="t('sso.lockReason')"
        label-width="auto"
      >
        <el-select
          v-model="queryForm.lockReason"
          :placeholder="t('sso.selectLockReason')"
          clearable
        >
          <el-option
            :label="t('sso.bruteForce')"
            value="BRUTE_FORCE"
          />
          <el-option
            :label="t('sso.manualLock')"
            value="ADMIN_MANUAL"
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
      :show-index="true"
      :operate-width="120"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sso.accountLockTitle') }}</template>
      <template #lockReason="{ row }">
        <el-tag
          size="small"
          :type="row.lockReason === 'BRUTE_FORCE' ? 'danger' : 'warning'"
        >
          {{ row.lockReason === 'BRUTE_FORCE' ? $t('sso.bruteForce') : $t('sso.manualLock') }}
        </el-tag>
      </template>
      <template #unlockTime="{ row }">
        {{ row.unlockTime || '--' }}
      </template>
      <template #operate="{ row }">
        <el-button
          v-if="!row.unlockedBy"
          v-permission="'sso:account:unlock'"
          link
          type="primary"
          @click="handleUnlock(row)"
        >
          {{ $t('sso.unlock') }}
        </el-button>
        <span
          v-else
          style="color: #909399"
        >
          {{ $t('sso.unlocked') }}
        </span>
      </template>
    </SmCrudTable>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getSsoAccountLocks, unlockSsoAccount } from '@/api/sso'
import type { SsoAccountLockVO } from '@/types/sso'

const loading = ref(false)
const { t } = useI18n()
const tableData = ref<SsoAccountLockVO[]>([])
const total = ref(0)
const queryForm = reactive({ current: 1, size: 10, userName: undefined, lockReason: undefined })

const columns = computed<TableColumn[]>(() => [
  { prop: 'username', label: t('sso.username'), width: 120, align: 'left' },
  { prop: 'lockReason', label: t('sso.lockReason'), slot: 'lockReason' },
  { prop: 'failedCount', label: t('sso.failedCount'), width: 120, align: 'right' },
  { prop: 'lockTime', label: t('sso.lockTime'), width: 170 },
  { prop: 'unlockTime', label: t('sso.autoUnlock'), width: 170, slot: 'unlockTime' },
  { prop: 'unlockedBy', label: t('sso.unlockedBy'), width: 130, align: 'left' },
])

async function loadData() {
  loading.value = true
  try {
    const res = await getSsoAccountLocks({
      page: queryForm.current,
      size: queryForm.size,
      userName: queryForm.userName,
      lockReason: queryForm.lockReason,
    })
    tableData.value = res.records || []
    total.value = Number(res.totalRow)
  } catch {
    tableData.value = []
  } finally {
    loading.value = false
  }
}

async function handleUnlock(row: SsoAccountLockVO) {
  await ElMessageBox.confirm(t('sso.confirmUnlock', { name: row.username }), t('common.tip'), { type: 'warning' })
  try {
    await unlockSsoAccount(row.id, 'admin')
    ElMessage.success(t('sso.unlockSuccess'))
    loadData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || t('sso.unlockFailed'))
  }
}

function handleQuery() {
  queryForm.current = 1
  loadData()
}

function handleReset() {
  queryForm.current = 1
  queryForm.size = 10
  queryForm.userName = undefined
  queryForm.lockReason = undefined
  loadData()
}

onMounted(() => loadData())
</script>
