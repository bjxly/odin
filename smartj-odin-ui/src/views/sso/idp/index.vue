<template>
  <div class="sm-page">
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item :label="$t('sso.idpName')">
        <el-input
          v-model="queryForm.idpName"
          :placeholder="$t('common.pleaseInput') + $t('sso.idpName')"
          clearable
          style="width: 160px"
        />
      </el-form-item>
      <el-form-item :label="$t('sso.idpType')">
        <el-select
          v-model="queryForm.idpType"
          :placeholder="$t('common.pleaseSelect') + $t('sso.idpType')"
          clearable
          style="width: 200px"
        >
          <el-option
            v-for="item of idpTypeLabels"
            :key="item.dictValue"
            :label="item.dictLabel"
            :value="item.dictValue"
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
      :operate-width="180"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sso.idpTitle') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sso:idp:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
        </el-button>
      </template>
      <template #idpType="{ row }">
        <el-tag size="small">{{ row.idpName }}</el-tag>
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
          v-permission="'sso:idp:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-permission="'sso:idp:delete'"
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
      width="600px"
      :draggable="true"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="auto"
      >
        <el-form-item
          :label="$t('sso.idpName')"
          prop="idpName"
        >
          <el-input
            v-model="form.idpName"
            :placeholder="$t('common.pleaseInput') + $t('sso.idpName')"
            :maxlength="100"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.idpType')"
          prop="idpType"
        >
          <el-select
            v-model="form.idpType"
            :placeholder="$t('common.pleaseSelect') + $t('sso.idpType')"
          >
            <el-option
              v-for="item of idpTypeLabels"
              :key="item.dictValue"
              :label="item.dictLabel"
              :value="item.dictValue"
            />
          </el-select>
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
          :label="$t('sso.loginPolicy')"
          prop="bindPolicy"
        >
          <el-select
            v-model="form.bindPolicy"
            :placeholder="$t('common.pleaseSelect') + $t('sso.loginPolicy')"
          >
            <el-option
              value="AUTO_CREATE"
              :label="$t('sso.autoCreateUser')"
            />
            <el-option
              value="PROMPT_BIND"
              :label="$t('sso.guideBindExisting')"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          label="Client ID"
          prop="clientId"
        >
          <el-input
            v-model="form.clientId"
            :placeholder="$t('common.pleaseInput') + 'Client ID'"
            :maxlength="255"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          label="Client Secret"
          prop="clientSecret"
        >
          <el-input
            v-model="form.clientSecret"
            :placeholder="$t('common.pleaseInput') + 'Client Secret'"
            :maxlength="512"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sso.authEndpoint')">
          <el-input
            v-model="form.authEndpoint"
            :placeholder="$t('common.pleaseInput') + $t('sso.authEndpoint')"
            :maxlength="500"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sso.tokenEndpoint')">
          <el-input
            v-model="form.tokenEndpoint"
            :placeholder="$t('common.pleaseInput') + $t('sso.tokenEndpoint')"
            :maxlength="500"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sso.userinfoEndpoint')">
          <el-input
            v-model="form.userinfoEndpoint"
            :placeholder="$t('common.pleaseInput') + $t('sso.userinfoEndpoint')"
            :maxlength="500"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sso.logoutEndpoint')">
          <el-input
            v-model="form.logoutEndpoint"
            :placeholder="$t('common.pleaseInput') + $t('sso.logoutEndpoint')"
            :maxlength="500"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sso.metadataUrl')">
          <el-input
            v-model="form.metadataUrl"
            :placeholder="$t('common.pleaseInput') + $t('sso.metadataUrl')"
            :maxlength="500"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sso.logoUrl')">
          <el-input
            v-model="form.logoUrl"
            :placeholder="$t('common.pleaseInput') + $t('sso.logoUrl')"
            :maxlength="500"
            :show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('common.status')"
          prop="status"
        >
          <el-radio-group v-model="form.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('common.remark')">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="2"
            :placeholder="$t('common.pleaseInput') + $t('common.remark')"
            :maxlength="500"
            :show-word-limit="true"
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getSsoIdps, createSsoIdp, updateSsoIdp, deleteSsoIdp } from '@/api/sso'
import type { SsoIdpConfigVO } from '@/types/sso'
import { getDictDataByType } from '@/api/sys/dict'
import type { DictDataVO } from '@/types/sys'

const loading = ref(false)
const { t } = useI18n()
const tableData = ref<SsoIdpConfigVO[]>([])
const total = ref(0)
const queryForm = reactive({ idpName: '', idpType: '', current: 1, size: 10 })

const columns = computed<TableColumn[]>(() => [
  { prop: 'idpName', label: t('sso.idpName'), minWidth: 160, align: 'left' },
  { prop: 'idpType', label: t('sso.idpType'), width: 100, slot: 'idpType' },
  { prop: 'protocol', label: t('sso.protocol'), width: 90, align: 'left' },
  { prop: 'status', label: t('common.status'), width: 100, slot: 'status' },
  { prop: 'createTime', label: t('common.createTime'), width: 170 },
  { prop: 'remark', label: t('common.remark'), minWidth: 150, align: 'left' },
])
const idpTypeLabels = ref<DictDataVO[]>([])

async function loadData() {
  loading.value = true
  try {
    const res = await getSsoIdps({
      page: queryForm.current,
      size: queryForm.size,
      idpName: queryForm.idpName || undefined,
      idpType: queryForm.idpType || undefined,
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
  queryForm.idpName = ''
  queryForm.idpType = ''
  handleQuery()
}

// 弹窗
const dialogVisible = ref(false)
const dialogTitle = ref('')
const editId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const submitLoading = ref(false)
const form = reactive<SsoIdpConfigVO>({
  idpName: '',
  idpType: 'OAUTH2',
  protocol: 'OAUTH2',
  clientId: '',
  clientSecret: '',
  status: 1,
  bindPolicy: '',
})

const rules = computed<FormRules>(() => ({
  idpName: [{ required: true, message: t('common.pleaseInput') + t('sso.idpName'), trigger: 'blur' }],
  idpType: [{ required: true, message: t('common.pleaseSelect') + t('sso.idpType'), trigger: 'change' }],
  protocol: [{ required: true, message: t('common.pleaseSelect') + t('sso.protocol'), trigger: 'change' }],
  bindPolicy: [{ required: true, message: t('common.pleaseSelect') + t('sso.bindPolicy'), trigger: 'change' }],
}))

function handleEditorFormReset() {
  Object.assign(form, {
    id: undefined,
    idpName: '',
    idpType: 'OAUTH2',
    protocol: 'OAUTH2',
    clientId: '',
    clientSecret: '',
    authEndpoint: '',
    tokenEndpoint: '',
    userinfoEndpoint: '',
    logoutEndpoint: '',
    metadataUrl: '',
    logoUrl: '',
    status: 1,
    remark: '',
    bindPolicy: undefined,
    createTime: '',
  })
}

function handleAdd() {
  dialogTitle.value = t('sso.addIdp')
  editId.value = null
  handleEditorFormReset()
  dialogVisible.value = true
}

function handleEdit(row: SsoIdpConfigVO) {
  dialogTitle.value = t('sso.editIdp')
  editId.value = row.id!
  handleEditorFormReset()
  Object.assign(form, { ...row })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    if (editId.value) {
      await updateSsoIdp(editId.value, form)
      ElMessage.success(t('sso.updateSuccess'))
    } else {
      await createSsoIdp(form)
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

async function handleDelete(row: SsoIdpConfigVO) {
  await ElMessageBox.confirm(t('sso.confirmDeleteIdp', { name: row.idpName }), t('common.tip'), { type: 'warning' })
  try {
    await deleteSsoIdp(row.id!)
    ElMessage.success(t('sso.deleteSuccess'))
    loadData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || t('common.fail'))
  }
}

function handleDialogClosed() {
  formRef.value?.resetFields()
}

async function loadIdpType() {
  // 加载Idp类型字典数据
  let res = await getDictDataByType('IdpType')
  idpTypeLabels.value = res
}

onMounted(() => {
  loadIdpType()
  loadData()
})
</script>
