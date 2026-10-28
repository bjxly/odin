<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item :label="$t('sysUser.username')">
        <el-input
          v-model="queryForm.username"
          :placeholder="$t('common.pleaseInput') + $t('sysUser.username')"
          clearable
        />
      </el-form-item>
      <el-form-item :label="$t('sysUser.nickname')">
        <el-input
          v-model="queryForm.nickname"
          :placeholder="$t('common.pleaseInput') + $t('sysUser.nickname')"
          clearable
        />
      </el-form-item>
      <el-form-item :label="$t('sysUser.mobile')">
        <el-input
          v-model="queryForm.mobile"
          :placeholder="$t('common.pleaseInput') + $t('sysUser.mobile')"
          clearable
        />
      </el-form-item>
      <el-form-item :label="$t('sysUser.status')">
        <el-select
          v-model="queryForm.status"
          :placeholder="$t('common.pleaseSelect')"
          clearable
        >
          <el-option
            :label="$t('common.enabled')"
            :value="1"
          />
          <el-option
            :label="$t('common.disabled')"
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
      :operate-width="240"
      @selection-change="handleSelectionChange"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('platform.list') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sys:user:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
        </el-button>
        <el-button
          v-permission="'sys:user:delete'"
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
          {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
      <template #operate="{ row }">
        <el-button
          v-permission="'platform:user:authorize'"
          link
          type="primary"
          @click="handleAuthorize(row)"
        >
          {{ $t('platform.authorize') }}
        </el-button>
        <el-button
          v-permission="'sys:user:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-permission="'sys:user:resetPwd'"
          link
          type="warning"
          @click="handleResetPwd(row)"
        >
          {{ $t('common.resetPassword') }}
        </el-button>
        <el-button
          v-permission="'sys:user:delete'"
          link
          type="danger"
          :disabled="row.username === 'admin' || isSelf(row)"
          @click="handleDelete(row)"
        >
          {{ $t('common.delete') }}
        </el-button>
      </template>
    </SmCrudTable>

    <!-- 租户授权弹窗 -->
    <AuthorizationDialog
      v-model="authVisible"
      :user-id="authUserId"
      @success="loadData"
    />

    <!-- 新增/编辑账号弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('platform.editTitle') : $t('platform.addTitle')"
      width="560px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-width="80px"
      >
        <el-form-item
          :label="$t('sysUser.username')"
          prop="username"
        >
          <el-input
            v-model="editForm.username"
            :placeholder="$t('common.pleaseInput') + $t('sysUser.username')"
            :disabled="isEdit"
            :maxlength="50"
            :minlength="2"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysUser.nickname')"
          prop="nickname"
        >
          <el-input
            v-model="editForm.nickname"
            :placeholder="$t('common.pleaseInput') + $t('sysUser.nickname')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          v-if="!isEdit"
          :label="$t('sysUser.password')"
          prop="password"
        >
          <el-input
            v-model="editForm.password"
            type="password"
            :placeholder="$t('common.pleaseInput') + $t('sysUser.password')"
            :maxlength="100"
            :minlength="passwordMinLength"
            show-password
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysUser.mobile')"
          prop="mobile"
        >
          <el-input
            v-model="editForm.mobile"
            :placeholder="$t('sysUser.mobilePlaceholder')"
            :maxlength="11"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysUser.email')"
          prop="email"
        >
          <el-input
            v-model="editForm.email"
            :placeholder="$t('sysUser.emailPlaceholder')"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysUser.gender')">
          <el-radio-group v-model="editForm.gender">
            <el-radio :value="1">{{ $t('sysUser.male') }}</el-radio>
            <el-radio :value="2">{{ $t('sysUser.female') }}</el-radio>
            <el-radio :value="0">{{ $t('sysUser.unknown') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysUser.status')">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysUser.remark')">
          <el-input
            v-model="editForm.remark"
            type="textarea"
            :rows="2"
            :placeholder="$t('sysUser.remarkPlaceholder')"
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
          @click="handleSubmit"
        >
          {{ $t('common.confirmBtn') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码弹窗 -->
    <el-dialog
      v-model="resetPwdVisible"
      :title="$t('sysUser.resetPwdTitle')"
      width="400px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="resetPwdFormRef"
        :model="resetPwdForm"
        :rules="resetPwdRules"
        label-width="80px"
      >
        <el-form-item
          :label="$t('sysUser.newPassword')"
          prop="password"
        >
          <el-input
            v-model="resetPwdForm.password"
            type="password"
            :placeholder="$t('sysUser.newPasswordPlaceholder')"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPwdVisible = false">{{ $t('common.cancelBtn') }}</el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          @click="handleResetPwdSubmit"
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
import AuthorizationDialog from './AuthorizationDialog.vue'
import { getPlatformUserPage } from '@/api/sys/platform'
import { createUser, updateUser, deleteUsers, resetPassword } from '@/api/sys/user'
import type { UserVO, UserDTO } from '@/types/sys'
import { usePasswordMinLength } from '@/composables/usePasswordMinLength'
import { usePasswordComplexity } from '@/composables/usePasswordComplexity'
import { useUserStore } from '@/stores/user'

const loading = ref(false)
const submitLoading = ref(false)
const { t } = useI18n()
const userStore = useUserStore()
const passwordMinLength = usePasswordMinLength()
const passwordComplexityRequired = usePasswordComplexity()
const tableData = ref<UserVO[]>([])
const total = ref(0)
const selectedIds = ref<string[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const resetPwdVisible = ref(false)

const editFormRef = ref<FormInstance>()
const resetPwdFormRef = ref<FormInstance>()

// 授权弹窗
const authVisible = ref(false)
const authUserId = ref<string | undefined>(undefined)

const queryForm = reactive({
  current: 1,
  size: 10,
  username: '',
  nickname: '',
  mobile: '',
  status: undefined as number | undefined,
})

const columns = computed<TableColumn[]>(() => [
  { prop: 'username', label: t('sysUser.username'), minWidth: 120, align: 'left' },
  { prop: 'nickname', label: t('sysUser.nickname'), minWidth: 120, align: 'left' },
  { prop: 'mobile', label: t('sysUser.mobile'), minWidth: 130, align: 'left' },
  { prop: 'status', label: t('sysUser.status'), width: 100, slot: 'status' },
  { prop: 'createTime', label: t('sysUser.createTime'), width: 170 },
])

const editForm = reactive<UserDTO>({
  id: undefined,
  username: '',
  nickname: '',
  password: '',
  mobile: '',
  email: '',
  gender: 1,
  status: 1,
  remark: '',
})

const resetPwdForm = reactive({ password: '', userId: '' })

/** 密码复杂度校验：开关开启时要求同时包含大写、小写、数字和特殊字符 */
function validateComplexity(_rule: any, value: string, callback: any) {
  if (!passwordComplexityRequired.value || !value) {
    callback()
    return
  }
  const ok = /[a-z]/.test(value) && /[A-Z]/.test(value) && /\d/.test(value) && /[^a-zA-Z0-9]/.test(value)
  callback(ok ? undefined : new Error(t('sysUser.passwordComplexity')))
}

const editRules = computed<FormRules>(() => ({
  username: [{ required: true, message: t('common.pleaseInput') + t('sysUser.username'), trigger: 'blur' }],
  nickname: [{ required: true, message: t('common.pleaseInput') + t('sysUser.nickname'), trigger: 'blur' }],
  password: [
    { required: true, message: t('common.pleaseInput') + t('sysUser.password'), trigger: 'blur' },
    { min: passwordMinLength.value, message: t('sysUser.passwordMinLength', { min: passwordMinLength.value }), trigger: 'blur' },
    { validator: validateComplexity, trigger: 'blur' },
  ],
  mobile: [
    {
      validator(_: any, value: string, callback: any) {
        // 校验手机号必须为数字，非数字字符实时过滤
        const filtered = value?.replace(/[^0-9]/g, '') ?? ''
        if (filtered !== value) {
          editForm.mobile = filtered
        }
        if (filtered === '') {
          callback()
          return
        }
        if (/^\d+$/.test(filtered)) {
          callback()
          return
        }
        callback(new Error(t('sysUser.phoneOnlyDigits')))
      },
      trigger: 'change',
    },
    {
      validator(_: any, value: string, callback: any) {
        // 校验手机号格式
        if (value == null || value === '') {
          callback()
          return
        }
        if (/^1[3-9]\d{9}$/.test(value)) {
          callback()
          return
        }
        callback(new Error(t('sysUser.phoneInvalid')))
      },
      trigger: 'blur',
    },
  ],
  email: [
    {
      validator(_: any, v: string, callback: any) {
        // 校验邮箱格式
        if (v == null || v === '') {
          callback()
          return
        }
        if (/[\w.%+-]+@[\w.-]+\.[a-zA-Z]{2,}/.test(v)) {
          callback()
          return
        }
        callback(new Error(t('sysUser.emailInvalid')))
      },
      trigger: 'blur',
    },
  ],
}))

const resetPwdRules = computed<FormRules>(() => ({
  password: [
    { required: true, message: t('sysUser.newPasswordPlaceholder'), trigger: 'blur' },
    { min: passwordMinLength.value, message: t('sysUser.passwordMinLength', { min: passwordMinLength.value }), trigger: 'blur' },
    { validator: validateComplexity, trigger: 'blur' },
  ],
}))

async function loadData() {
  loading.value = true
  try {
    const res = await getPlatformUserPage(queryForm)
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
  queryForm.username = ''
  queryForm.nickname = ''
  queryForm.mobile = ''
  queryForm.status = undefined
  handleQuery()
}

function handleSelectionChange(rows: UserVO[]) {
  selectedIds.value = rows.map((r) => r.id)
}

/** 判断是否为当前登录用户自身 */
function isSelf(row: UserVO) {
  return String(row.id) === userStore.userInfo?.id
}

/** 打开租户授权弹窗 */
function handleAuthorize(row: UserVO) {
  authUserId.value = row.id
  authVisible.value = true
}

function handleAdd() {
  isEdit.value = false
  Object.assign(editForm, { id: undefined, username: '', nickname: '', password: '', mobile: '', email: '', gender: 1, status: 1, remark: '' })
  dialogVisible.value = true
}

function handleEdit(row: UserVO) {
  isEdit.value = true
  Object.assign(editForm, { ...row, password: '' })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateUser(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createUser(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: UserVO) {
  await ElMessageBox.confirm(t('sysUser.confirmDeleteUser'), t('common.tip'), { type: 'warning' })
  await deleteUsers([row.id])
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

async function handleBatchDelete() {
  if (userStore.userInfo?.id && selectedIds.value.some((id) => String(id) === userStore.userInfo?.id)) {
    ElMessage.warning(t('sysUser.cannotDeleteSelf'))
    return
  }
  await ElMessageBox.confirm(t('sysUser.confirmDeleteUsers', { count: selectedIds.value.length }), t('common.tip'), { type: 'warning' })
  await deleteUsers(selectedIds.value)
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

function handleResetPwd(row: UserVO) {
  resetPwdForm.userId = row.id
  resetPwdForm.password = ''
  resetPwdVisible.value = true
}

async function handleResetPwdSubmit() {
  if (!resetPwdFormRef.value) return
  const valid = await resetPwdFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    await resetPassword(resetPwdForm.userId, resetPwdForm.password)
    ElMessage.success(t('sysUser.resetPwdSuccess'))
    resetPwdVisible.value = false
  } finally {
    submitLoading.value = false
  }
}

watch(() => [queryForm.current, queryForm.size], loadData)

onMounted(() => {
  loadData()
})
</script>
