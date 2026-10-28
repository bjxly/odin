<template>
  <div class="sm-page">
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sso.clientName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.clientName"
          :placeholder="$t('common.pleaseInput') + $t('sso.clientName')"
          clearable
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
      :operate-width="180"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sso.clientTitle') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sso:client:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
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
          v-permission="'sso:client:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-permission="'sso:client:delete'"
          link
          type="danger"
          @click="handleDelete(row)"
        >
          {{ $t('common.delete') }}
        </el-button>
      </template>
    </SmCrudTable>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="560px"
      :close-on-click-modal="false"
      :draggable="true"
      @closed="handleDialogClosed"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="auto"
      >
        <el-form-item
          :label="$t('sso.clientId')"
          prop="clientId"
        >
          <el-input
            v-model="form.clientId"
            :placeholder="$t('common.pleaseInput') + $t('sso.clientId')"
            :disabled="!!editId"
            :maxlength="100"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.clientName')"
          prop="clientName"
        >
          <el-input
            v-model="form.clientName"
            :placeholder="$t('common.pleaseInput') + $t('sso.clientName')"
            :maxlength="200"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          v-if="!editId"
          :label="$t('sso.clientSecret')"
          prop="clientSecret"
        >
          <el-input
            v-model="form.clientSecret"
            :placeholder="$t('common.pleaseInput') + $t('sso.clientSecret')"
            :maxlength="255"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.protocol')"
          prop="protocol"
        >
          <el-select
            v-model="form.protocol"
            :placeholder="$t('common.pleaseSelect') + $t('sso.protocol')"
          >
            <el-option
              label="OAuth 2.0"
              value="OAUTH2"
            />
            <el-option
              label="OIDC"
              value="OIDC"
            />
            <el-option
              label="SAML 2.0"
              value="SAML2"
            />
            <el-option
              label="CAS"
              value="CAS"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          :label="$t('sso.grantTypes')"
          prop="grantTypes"
        >
          <el-input
            v-model="form.grantTypes"
            :placeholder="$t('common.pleaseInput') + $t('sso.grantTypes')"
            :maxlength="500"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.redirectUris')"
          prop="redirectUris"
        >
          <el-input
            v-model="form.redirectUris"
            type="textarea"
            :rows="2"
            :placeholder="$t('common.pleaseInput') + $t('sso.redirectUris')"
            :maxlength="2000"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysRole.status')"
          prop="status"
        >
          <el-radio-group v-model="form.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          :label="$t('sso.tokenTtl')"
          prop="tokenTtl"
        >
          <el-input-number
            v-model="form.tokenTtl"
            :min="60"
            :step="3600"
            style="width: 100%"
          >
            <template #suffix>
              <span>{{ $t('sso.seconds') }}</span>
            </template>
          </el-input-number>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getSsoClients, createSsoClient, updateSsoClient, deleteSsoClient } from '@/api/sso'
import type { SsoClientVO } from '@/types/sso'

const loading = ref(false)
const { t } = useI18n()
const tableData = ref<SsoClientVO[]>([])
const total = ref(0)
const queryForm = reactive({ clientName: '', current: 1, size: 10 })

const columns = computed<TableColumn[]>(() => [
  { prop: 'clientId', label: t('sso.clientId'), width: 160, align: 'left' },
  { prop: 'clientName', label: t('sso.clientName'), width: 140, align: 'left' },
  { prop: 'protocol', label: t('sso.protocol'), width: 120, align: 'left' },
  { prop: 'grantTypes', label: t('sso.grantTypes'), minWidth: 180, align: 'left' },
  { prop: 'redirectUris', label: t('sso.redirectUris'), minWidth: 200, align: 'left' },
  { prop: 'status', label: t('common.status'), width: 100, slot: 'status' },
  { prop: 'createTime', label: t('common.createTime'), width: 160 },
])

async function loadData() {
  loading.value = true
  try {
    const res = await getSsoClients({
      page: queryForm.current,
      size: queryForm.size,
      clientName: queryForm.clientName || undefined,
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
  queryForm.clientName = ''
  handleQuery()
}

// 弹窗
const dialogVisible = ref(false)
const dialogTitle = ref('')
const editId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const form = reactive<SsoClientVO>({
  clientId: '',
  clientName: '',
  protocol: 'OAUTH2',
  grantTypes: 'authorization_code,refresh_token',
  redirectUris: '',
  tokenTtl: 7200,
  // refresh_ttl字段，过期时间默认3600秒
  refreshTtl: 3600,
  mfaRequired: 0,
  status: 1,
})

const rules = computed<FormRules>(() => ({
  clientId: [{ required: true, message: t('common.pleaseInput') + t('sso.clientId'), trigger: 'blur' }],
  clientName: [{ required: true, message: t('common.pleaseInput') + t('sso.clientName'), trigger: 'blur' }],
  protocol: [{ required: true, message: t('common.pleaseSelect') + t('sso.protocol'), trigger: 'change' }],
  grantTypes: [{ required: true, message: t('common.pleaseInput') + t('sso.grantTypes'), trigger: 'blur' }],
  redirectUris: [{ required: true, message: t('common.pleaseInput') + t('sso.redirectUris'), trigger: 'blur' }],
}))

function handleAdd() {
  dialogTitle.value = t('sso.addClient')
  editId.value = null
  Object.assign(form, {
    id: undefined,
    clientId: '',
    clientName: '',
    clientSecret: '',
    protocol: 'OAUTH2',
    grantTypes: 'authorization_code,refresh_token',
    redirectUris: '',
    tokenTtl: 7200,
    status: 1,
  })
  dialogVisible.value = true
}

function handleEdit(row: SsoClientVO) {
  dialogTitle.value = t('sso.editClient')
  editId.value = row.id!
  Object.assign(form, {
    ...row,
    clientSecret: undefined,
    mfaRequired: 0,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    if (editId.value) {
      await updateSsoClient(editId.value, form)
      ElMessage.success(t('sso.updateSuccess'))
    } else {
      await createSsoClient(form)
      ElMessage.success(t('sso.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.fail'))
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: SsoClientVO) {
  await ElMessageBox.confirm(t('sso.confirmDeleteClient', { name: row.clientName }), t('common.tip'), { type: 'warning' })
  try {
    await deleteSsoClient(row.id!)
    ElMessage.success(t('sso.deleteSuccess'))
    loadData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || t('common.fail'))
  }
}

function handleDialogClosed() {
  formRef.value?.resetFields()
}

onMounted(() => loadData())
</script>
