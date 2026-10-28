<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('tenant.tenantName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.name"
          :placeholder="$t('common.pleaseInput') + $t('tenant.tenantName')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('tenant.tenantCode')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.code"
          :placeholder="$t('common.pleaseInput') + $t('tenant.tenantCode')"
          clearable
        />
      </el-form-item>
      <el-form-item :label="$t('tenant.status')">
        <el-select
          v-model="queryForm.status"
          :placeholder="$t('sysNotice.all')"
          clearable
        >
          <el-option
            :label="$t('tenant.enable')"
            :value="1"
          />
          <el-option
            :label="$t('tenant.disable')"
            :value="0"
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
      :show-selection="true"
      :show-index="true"
      :operate-width="230"
      @selection-change="handleSelectionChange"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('tenant.list') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sys:tenant:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
        </el-button>
        <el-button
          v-permission="'sys:tenant:delete'"
          type="danger"
          :disabled="!selectedIds.length"
          @click="handleBatchDelete"
        >
          {{ $t('common.batchDelete') }}
        </el-button>
      </template>
      <template #status="{ row }">
        <el-tag
          :type="row.status === 1 ? 'success' : 'danger'"
          size="small"
        >
          {{ row.status === 1 ? $t('tenant.enable') : $t('tenant.disable') }}
        </el-tag>
      </template>
      <template #operate="{ row }">
        <el-button
          v-permission="'platform:tenant:role'"
          link
          type="primary"
          @click="handleRoleAuth(row)"
        >
          {{ $t('platform.roleAuth') }}
        </el-button>
        <el-button
          v-permission="'sys:tenant:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-permission="'sys:tenant:delete'"
          link
          type="danger"
          :disabled="row.id === '1'"
          @click="handleDelete(row)"
        >
          {{ $t('common.delete') }}
        </el-button>
      </template>
    </SmCrudTable>

    <!-- 角色权限弹窗（平台超管跨租户维护目标租户角色） -->
    <TenantRoleDialog
      v-model="roleAuthVisible"
      :tenant-id="roleAuthTenantId"
      :tenant-name="roleAuthTenantName"
    />

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('tenant.editTitle') : $t('tenant.addTitle')"
      width="560px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-width="auto"
      >
        <el-form-item
          :label="$t('tenant.tenantName')"
          prop="name"
        >
          <el-input
            v-model="editForm.name"
            :placeholder="$t('common.pleaseInput') + $t('tenant.tenantName')"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('tenant.tenantCode')"
          prop="code"
        >
          <el-input
            v-model="editForm.code"
            :placeholder="$t('common.pleaseInput') + $t('tenant.tenantCode')"
            :disabled="isEdit"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          v-if="!isEdit"
          :label="$t('tenant.adminRole')"
          prop="roleId"
        >
          <el-select
            v-model="editForm.roleId"
            :placeholder="$t('tenant.adminRoleTip')"
            :loading="roleLoading"
            :disabled="roleLoading"
            :no-data-text="$t('tenant.noAvailableRole')"
            filterable
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="role in selectableRoles"
              :key="role.id"
              :label="role.roleName"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="!isEdit">
          <el-checkbox v-model="editForm.associateCurrentUser">
            {{ $t('tenant.associateCurrentUser') }}
          </el-checkbox>
          <div class="tenant-association-tip">{{ $t('tenant.associateCurrentUserTip') }}</div>
        </el-form-item>
        <el-form-item
          :label="$t('tenant.contactName')"
          prop="contact"
        >
          <el-input
            v-model="editForm.contact"
            :placeholder="$t('common.pleaseInput') + $t('tenant.contactName')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('tenant.contactPhone')"
          prop="phone"
        >
          <el-input
            v-model="editForm.phone"
            :placeholder="$t('common.pleaseInput') + $t('tenant.contactPhone')"
            :maxlength="11"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('tenant.expireTime')">
          <el-date-picker
            v-model="editForm.expireTime"
            type="date"
            :placeholder="$t('common.pleaseSelect') + $t('tenant.expireTime')"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('tenant.address')">
          <el-input
            v-model="editForm.address"
            :placeholder="$t('common.pleaseInput') + $t('tenant.address')"
            :maxlength="200"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('tenant.status')">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">{{ $t('tenant.enable') }}</el-radio>
            <el-radio :value="0">{{ $t('tenant.disable') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('tenant.remark')">
          <el-input
            v-model="editForm.remark"
            type="textarea"
            :rows="2"
            :placeholder="$t('common.pleaseInput') + $t('tenant.remark')"
            :maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancelBtn') }}</el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          :disabled="!isEdit && (roleLoading || !editForm.roleId)"
          @click="handleSubmit"
        >
          {{ $t('common.confirmBtn') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getTenantPage, getTenantAdminRoles, createTenant, updateTenant, deleteTenants } from '@/api/sys/tenant'
import TenantRoleDialog from './TenantRoleDialog.vue'
import type { TenantVO, TenantDTO, RoleVO } from '@/types/sys'

const loading = ref(false)
const { t } = useI18n()
const submitLoading = ref(false)
const tableData = ref<TenantVO[]>([])
const total = ref(0)
const selectedIds = ref<string[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
// 角色权限入口（仅平台超管在系统租户可见，服务端二次校验）
const roleAuthVisible = ref(false)
const roleAuthTenantId = ref<string | undefined>(undefined)
const roleAuthTenantName = ref('')
const roleLoading = ref(false)
const roleList = ref<RoleVO[]>([])
const selectableRoles = computed(() => roleList.value.filter((role) => role.roleCode.toLowerCase() !== 'super_admin' && role.status === 1))

const editFormRef = ref<FormInstance>()

const queryForm = reactive({
  current: 1,
  size: 10,
  name: '',
  code: '',
  status: undefined as number | undefined,
})

const columns = computed<TableColumn[]>(() => [
  { prop: 'name', label: t('tenant.tenantName'), minWidth: 150, align: 'left' },
  { prop: 'code', label: t('tenant.tenantCode'), minWidth: 120, align: 'left' },
  { prop: 'contact', label: t('tenant.contactName'), minWidth: 100, align: 'left' },
  { prop: 'phone', label: t('tenant.contactPhone'), minWidth: 130, align: 'left' },
  { prop: 'address', label: t('tenant.address'), minWidth: 150, align: 'left' },
  { prop: 'status', label: t('tenant.status'), width: 100, slot: 'status' },
  { prop: 'expireTime', label: t('tenant.expireTime'), width: 170 },
  { prop: 'createTime', label: t('tenant.createTime'), width: 170 },
])

watch(() => [queryForm.current, queryForm.size], loadData)

const editForm = reactive<TenantDTO>({
  id: undefined,
  roleId: undefined,
  associateCurrentUser: true,
  name: '',
  code: '',
  contact: '',
  phone: '',
  expireTime: '',
  address: '',
  status: 1,
  remark: '',
})

const editRules = computed<FormRules>(() => ({
  name: [{ required: true, message: t('common.pleaseInput') + t('tenant.tenantName'), trigger: 'blur' }],
  code: [{ required: true, message: t('common.pleaseInput') + t('tenant.tenantCode'), trigger: 'blur' }],
  roleId: [{ required: !isEdit.value, message: t('tenant.adminRoleRequired'), trigger: 'change' }],
  phone: [
    {
      validator(_: any, value: string, callback: any) {
        // 校验手机号必须为数字，非数字字符实时过滤
        const filtered = value?.replace(/[^0-9]/g, '') ?? ''
        if (filtered !== value) {
          editForm.phone = filtered
        }
        if (filtered === '') {
          callback()
          return
        }
        const regex = /^\d+$/g
        if (regex.test(filtered)) {
          callback()
          return
        }
        callback(new Error(t('tenant.phoneOnlyDigits')))
      },
      trigger: 'change',
    },
    {
      validator(_: any, value: string, callback: any) {
        // 校验手机号格式
        if (value == null || value === '') {
          callback()
        }
        const regex = /^(1[3-9]\d{9})?$/g
        if (regex.test(value)) {
          callback()
        }
        callback(new Error(t('tenant.phoneInvalid')))
      },
      trigger: 'blur',
    },
  ],
}))

async function loadData() {
  loading.value = true
  try {
    const res = await getTenantPage(queryForm)
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
  queryForm.name = ''
  queryForm.code = ''
  queryForm.status = undefined
  handleQuery()
}

function handleSelectionChange(rows: TenantVO[]) {
  selectedIds.value = rows.map((r) => r.id)
}

/** 打开目标租户的角色权限维护弹窗 */
function handleRoleAuth(row: TenantVO) {
  roleAuthTenantId.value = row.id
  roleAuthTenantName.value = row.name
  roleAuthVisible.value = true
}

async function handleAdd() {
  isEdit.value = false
  Object.assign(editForm, {
    id: undefined,
    roleId: undefined,
    associateCurrentUser: true,
    name: '',
    code: '',
    contact: '',
    phone: '',
    expireTime: '',
    address: '',
    status: 1,
    remark: '',
  })
  roleList.value = []
  dialogVisible.value = true
  roleLoading.value = true
  try {
    roleList.value = await getTenantAdminRoles()
  } catch {
    // 请求层统一提示错误，保持角色列表为空，禁止无角色提交。
  } finally {
    roleLoading.value = false
  }
}

function handleEdit(row: TenantVO) {
  isEdit.value = true
  Object.assign(editForm, { ...row, roleId: undefined, associateCurrentUser: undefined })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return
  if (!isEdit.value && !selectableRoles.value.some((role) => role.id === editForm.roleId)) {
    ElMessage.warning(t('tenant.adminRoleRequired'))
    return
  }
  // 处理过期时间，时分秒为23:59:59
  if (editForm.expireTime && editForm.expireTime !== '') {
    let date = editForm.expireTime.split(' ')[0]
    editForm.expireTime = date + ' 23:59:59'
  }
  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateTenant(editForm)
      ElMessage.success(t('tenant.updateSuccess'))
    } else {
      await createTenant(editForm)
      ElMessage.success(t('tenant.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: TenantVO) {
  await ElMessageBox.confirm(t('tenant.confirmDelete'), t('tenant.confirmDeleteTip'), { type: 'warning' })
  // loading状态
  loading.value = true
  await deleteTenants([row.id])
  ElMessage.success(t('tenant.deleteSuccess'))
  loadData()
}

async function handleBatchDelete() {
  await ElMessageBox.confirm(t('tenant.confirmBatchDelete', { count: selectedIds.value.length }), t('tenant.confirmDeleteTip'), { type: 'warning' })
  await deleteTenants(selectedIds.value)
  ElMessage.success(t('tenant.deleteSuccess'))
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.tenant-association-tip {
  width: 100%;
  color: var(--sm-text-secondary);
  font-size: 12px;
  line-height: 1.5;
}
</style>
