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
      <el-form-item :label="$t('common.status')">
        <el-select
          v-model="queryForm.status"
          :placeholder="$t('common.pleaseSelect')"
          clearable
          style="width: 140px"
        >
          <el-option
            :label="$t('sso.registrationPending')"
            :value="1"
          />
          <el-option
            :label="$t('sso.registrationApproved')"
            :value="2"
          />
          <el-option
            :label="$t('sso.registrationRejected')"
            :value="3"
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
      :operate-width="160"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sso.registrationTitle') }}</template>
      <template #status="{ row }">
        <el-tag
          :type="statusTagType(row.status)"
          size="small"
        >
          {{ statusLabel(row.status) }}
        </el-tag>
      </template>
      <template #mobile="{ row }">
        {{ row.mobile || '--' }}
      </template>
      <template #email="{ row }">
        {{ row.email || '--' }}
      </template>
      <template #approvedBy="{ row }">
        {{ row.approvedBy || '--' }}
      </template>
      <template #operate="{ row }">
        <template v-if="row.status === 1">
          <el-button
            v-permission="'sso:registration:approve'"
            link
            type="success"
            @click="handleApprove(row)"
          >
            {{ $t('sso.registrationApprove') }}
          </el-button>
          <el-button
            v-permission="'sso:registration:reject'"
            link
            type="danger"
            @click="handleReject(row)"
          >
            {{ $t('sso.registrationReject') }}
          </el-button>
        </template>
        <span
          v-else
          style="color: #909399"
        >
          {{ statusLabel(row.status) }}
        </span>
      </template>
    </SmCrudTable>

    <!-- 审批通过弹窗（含租户选择） -->
    <el-dialog
      v-model="approveDialogVisible"
      :title="$t('sso.registrationApproveTitle')"
      width="420px"
      :close-on-click-modal="false"
    >
      <p style="margin-bottom: 12px; color: #606266">
        {{ $t('sso.registrationApproveHint', { name: currentApproveUser }) }}
      </p>
      <el-form label-width="auto">
        <el-form-item :label="$t('sso.registrationSelectTenant')">
          <el-select
            v-model="approveTenantId"
            :placeholder="$t('sso.registrationSelectTenant')"
            style="width: 100%"
          >
            <el-option
              v-for="tenant in tenantOptions"
              :key="tenant.id"
              :label="tenant.name"
              :value="tenant.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="confirmApprove"
        >
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 拒绝原因弹窗 -->
    <el-dialog
      v-model="rejectDialogVisible"
      :title="$t('sso.registrationRejectReason')"
      width="420px"
      :close-on-click-modal="false"
    >
      <el-input
        v-model="rejectReason"
        type="textarea"
        :rows="3"
        :placeholder="$t('sso.registrationRejectReasonRequired')"
      />
      <template #footer>
        <el-button @click="rejectDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="confirmReject"
        >
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getRegistrationList, approveRegistration, rejectRegistration, getRegistrationTenants } from '@/api/sso'
import type { RegistrationReviewVO } from '@/api/sso'

const loading = ref(false)
const { t } = useI18n()
const tableData = ref<RegistrationReviewVO[]>([])
const total = ref(0)
const queryForm = reactive({ username: '', status: undefined as number | undefined, current: 1, size: 10 })

const rejectDialogVisible = ref(false)
const rejectReason = ref('')
const currentRejectId = ref<number | null>(null)

const approveDialogVisible = ref(false)
const approveTenantId = ref<number | undefined>(undefined)
const currentApproveId = ref<number | null>(null)
const currentApproveUser = ref('')
const tenantOptions = ref<Array<{ id: number; name: string; code: string }>>([])

const columns = computed<TableColumn[]>(() => [
  { prop: 'username', label: t('sso.username'), align: 'left' },
  { prop: 'nickname', label: t('sso.nickname'), align: 'left' },
  { prop: 'mobile', label: t('sso.mobile'), width: 130, slot: 'mobile' },
  { prop: 'email', label: t('sso.email'), width: 180, slot: 'email' },
  { prop: 'status', label: t('common.status'), width: 110, slot: 'status' },
  { prop: 'createTime', label: t('common.createTime'), width: 170 },
  { prop: 'approvedBy', label: t('sso.operatorName'), width: 120, slot: 'approvedBy' },
])

function statusTagType(status: number) {
  switch (status) {
    case 1:
      return 'warning'
    case 2:
      return 'success'
    case 3:
      return 'danger'
    default:
      return 'info'
  }
}

function statusLabel(status: number) {
  switch (status) {
    case 1:
      return t('sso.registrationPending')
    case 2:
      return t('sso.registrationApproved')
    case 3:
      return t('sso.registrationRejected')
    default:
      return String(status)
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await getRegistrationList({
      page: queryForm.current,
      size: queryForm.size,
      username: queryForm.username || undefined,
      status: queryForm.status,
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
  queryForm.status = undefined
  handleQuery()
}

async function handleApprove(row: RegistrationReviewVO) {
  currentApproveId.value = row.id
  currentApproveUser.value = row.username
  approveTenantId.value = undefined
  // 加载租户列表
  try {
    tenantOptions.value = await getRegistrationTenants()
    // 选项加载完后再设置默认值，避免 select 显示原始数字
    if (tenantOptions.value.length > 0 && approveTenantId.value === undefined) {
      approveTenantId.value = tenantOptions.value[0].id
    }
  } catch {
    tenantOptions.value = []
  }
  approveDialogVisible.value = true
}

async function confirmApprove() {
  if (!approveTenantId.value) {
    ElMessage.warning(t('sso.registrationSelectTenant'))
    return
  }
  await approveRegistration(currentApproveId.value!, approveTenantId.value)
  ElMessage.success(t('sso.registrationApproveSuccess'))
  approveDialogVisible.value = false
  loadData()
}

function handleReject(row: RegistrationReviewVO) {
  currentRejectId.value = row.id
  rejectReason.value = ''
  rejectDialogVisible.value = true
}

async function confirmReject() {
  if (!rejectReason.value.trim()) {
    ElMessage.warning(t('sso.registrationRejectReasonRequired'))
    return
  }
  await rejectRegistration(currentRejectId.value!, rejectReason.value)
  ElMessage.success(t('sso.registrationRejectSuccess'))
  rejectDialogVisible.value = false
  loadData()
}

onMounted(() => loadData())
</script>
