<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sysSms.providerIdLabel')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.providerId"
          :placeholder="$t('common.pleaseInput') + $t('sysSms.providerId')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysSms.enabled')"
        label-width="auto"
      >
        <el-select
          v-model="queryForm.enabled"
          :placeholder="$t('sysNotice.all')"
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
      :show-selection="false"
      :show-index="true"
      :operate-width="230"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sysSms.title') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sys:sms:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
        </el-button>
      </template>
      <template #isDefault="{ row }">
        <el-tag
          :type="row.isDefault === 1 ? 'success' : 'info'"
          size="small"
        >
          {{ row.isDefault === 1 ? $t('common.yes') : $t('common.no') }}
        </el-tag>
      </template>
      <template #enabled="{ row }">
        <el-tag
          :type="row.enabled === 1 ? 'success' : 'danger'"
          size="small"
        >
          {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
      <template #hasSecret="{ row }">
        <el-tag
          :type="row.hasSecret ? 'success' : 'info'"
          size="small"
        >
          {{ row.hasSecret ? $t('sysSms.configured') : $t('sysSms.notConfigured') }}
        </el-tag>
      </template>
      <template #operate="{ row }">
        <el-button
          v-permission="'sys:sms:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-if="row.enabled === 0"
          v-permission="'sys:sms:update'"
          link
          type="success"
          @click="handleEnable(row)"
        >
          {{ $t('common.enabled') }}
        </el-button>
        <el-button
          link
          type="success"
          @click="openTestDialog(row)"
        >
          {{ $t('sysSms.testSend') }}
        </el-button>
        <el-button
          v-permission="'sys:sms:delete'"
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
      :title="isEdit ? $t('sysSms.editTitle') : $t('sysSms.addTitle')"
      width="640px"
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
          :label="$t('sysSms.providerId')"
          prop="providerId"
        >
          <el-input
            v-model="editForm.providerId"
            :placeholder="$t('sysSms.providerIdPlaceholder')"
            :maxlength="32"
            show-word-limit
          />
        </el-form-item>

        <el-form-item :label="$t('sysSms.accessKeyId')">
          <el-input
            v-model="editForm.accessKeyId"
            :placeholder="$t('sysSms.accessKeyIdPlaceholder')"
            :maxlength="256"
            show-word-limit
          />
        </el-form-item>

        <el-form-item :label="$t('sysSms.accessKeySecret')">
          <el-input
            v-model="editForm.accessKeySecret"
            type="password"
            show-password
            :placeholder="$t('sysSms.accessKeySecretPlaceholder')"
            :maxlength="256"
          >
            <template #suffix>
              <span style="color: #909399; font-size: 12px; margin-right: 4px">{{ editForm.accessKeySecret ? editForm.accessKeySecret.length : 0 }} / 256</span>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item :label="$t('sysSms.signName')">
          <el-input
            v-model="editForm.signName"
            :placeholder="$t('sysSms.signNamePlaceholder')"
            :maxlength="64"
            show-word-limit
          />
        </el-form-item>

        <el-form-item :label="$t('sysSms.region')">
          <el-input
            v-model="editForm.region"
            :placeholder="$t('sysSms.regionPlaceholder')"
            :maxlength="32"
            show-word-limit
          />
        </el-form-item>

        <el-form-item :label="$t('sysSms.appId')">
          <el-input
            v-model="editForm.appId"
            :placeholder="$t('sysSms.appIdPlaceholder')"
            :maxlength="64"
            show-word-limit
          />
        </el-form-item>

        <el-form-item :label="$t('sysSms.sender')">
          <el-input
            v-model="editForm.sender"
            :placeholder="$t('sysSms.senderPlaceholder')"
            :maxlength="32"
            show-word-limit
          />
        </el-form-item>

        <el-form-item :label="$t('sysSms.templateConfig')">
          <div class="template-mapper">
            <div
              v-for="(item, index) in templateEntries"
              :key="index"
              class="template-mapper__row"
            >
              <el-input
                v-model="item.key"
                :placeholder="$t('sysSms.templateKeyPlaceholder')"
                :maxlength="64"
                :show-word-limit="true"
                style="flex: 1"
              />
              <el-input
                v-model="item.value"
                :placeholder="$t('sysSms.templateValuePlaceholder')"
                :maxlength="64"
                :show-word-limit="true"
                style="flex: 1"
              />
              <el-button
                type="danger"
                link
                @click="removeTemplateEntry(index)"
              >
                {{ $t('common.delete') }}
              </el-button>
            </div>
            <el-button
              type="primary"
              link
              @click="addTemplateEntry"
            >
              + {{ $t('sysSms.templateConfig') }}
            </el-button>
          </div>
        </el-form-item>

        <el-form-item :label="$t('sysSms.isDefault')">
          <el-radio-group v-model="editForm.isDefault">
            <el-radio :value="1">{{ $t('common.yes') }}</el-radio>
            <el-radio :value="0">{{ $t('common.no') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="$t('sysSms.enabled')">
          <el-radio-group v-model="editForm.enabled">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="$t('sysSms.remark')">
          <el-input
            v-model="editForm.remark"
            type="textarea"
            :rows="2"
            :placeholder="$t('common.pleaseInput') + $t('sysSms.remark')"
            :maxlength="500"
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

    <!-- 测试短信弹窗 -->
    <el-dialog
      v-model="testDialogVisible"
      :title="$t('sysSms.testTitle')"
      width="460px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form label-width="auto">
        <el-form-item :label="$t('sysSms.providerIdLabel')">
          <el-input
            :model-value="testForm.providerId"
            disabled
          />
        </el-form-item>
        <el-form-item :label="$t('sysSms.phone')">
          <el-input
            v-model="testForm.phone"
            :placeholder="$t('sysSms.pleaseInputPhone')"
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('sysSms.templateCode')">
          <el-select
            v-model="testForm.templateCode"
            :placeholder="$t('sysSms.selectTemplate')"
            style="width: 100%"
          >
            <el-option
              v-for="entry in testTemplateOptions"
              :key="entry.key"
              :label="`${entry.key}（${entry.value}）`"
              :value="entry.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testDialogVisible = false">{{ $t('common.cancelBtn') }}</el-button>
        <el-button
          type="primary"
          :loading="testLoading"
          @click="handleTestSend"
        >
          {{ $t('sysSms.sendTest') }}
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
import { getSmsConfigPage, getSmsConfigById, createSmsConfig, updateSmsConfig, deleteSmsConfig, enableSmsConfig, testSmsSend } from '@/api/sys/sms'
import type { SysSmsConfigVO, SysSmsConfigDTO } from '@/types/sys'

const loading = ref(false)
const { t } = useI18n()
const submitLoading = ref(false)
const tableData = ref<SysSmsConfigVO[]>([])
const total = ref(0)
const dialogVisible = ref(false)
const isEdit = ref(false)

const editFormRef = ref<FormInstance>()

// 模板映射 key-value 编辑
interface TemplateEntry {
  key: string
  value: string
}
const templateEntries = ref<TemplateEntry[]>([])

function addTemplateEntry() {
  templateEntries.value.push({ key: '', value: '' })
}

function removeTemplateEntry(index: number) {
  templateEntries.value.splice(index, 1)
}

/** JSON 字符串 -> key-value 数组 */
function parseTemplateConfig(json: string): TemplateEntry[] {
  if (!json) return []
  try {
    const obj = JSON.parse(json)
    return Object.entries(obj).map(([k, v]) => ({ key: k, value: String(v) }))
  } catch {
    return []
  }
}

/** key-value 数组 -> JSON 字符串 */
function buildTemplateConfig(entries: TemplateEntry[]): string {
  const obj: Record<string, string> = {}
  for (const e of entries) {
    if (e.key.trim()) obj[e.key.trim()] = e.value.trim()
  }
  return Object.keys(obj).length > 0 ? JSON.stringify(obj) : ''
}

// 测试短信
const testDialogVisible = ref(false)
const testLoading = ref(false)
const testTemplateOptions = ref<TemplateEntry[]>([])
const testForm = reactive({
  providerId: '',
  phone: '',
  templateCode: '',
})

function openTestDialog(row: SysSmsConfigVO) {
  testForm.providerId = row.providerId
  testForm.phone = ''
  // 解析模板映射作为下拉选项
  const entries = parseTemplateConfig(row.templateConfig)
  testTemplateOptions.value = entries
  testForm.templateCode = entries.length > 0 ? entries[0].value : ''
  testDialogVisible.value = true
}

async function handleTestSend() {
  if (!testForm.phone.trim()) {
    ElMessage.warning(t('sysSms.pleaseInputPhone'))
    return
  }
  if (!testForm.templateCode) {
    ElMessage.warning(t('sysSms.pleaseSelectTemplate'))
    return
  }
  testLoading.value = true
  try {
    await testSmsSend({
      providerId: testForm.providerId,
      phone: testForm.phone.trim(),
      templateCode: testForm.templateCode,
    })
    ElMessage.success(t('sysSms.testSuccess'))
    testDialogVisible.value = false
  } finally {
    testLoading.value = false
  }
}

const queryForm = reactive({
  current: 1,
  size: 10,
  providerId: '',
  enabled: undefined as number | undefined,
})

const columns = computed<TableColumn[]>(() => [
  { prop: 'providerId', label: t('sysSms.providerIdLabel'), width: 100, align: 'left' },
  { prop: 'signName', label: t('sysSms.signName'), width: 140, align: 'left' },
  { prop: 'accessKeyId', label: t('sysSms.accessKey'), minWidth: 140, align: 'left' },
  { prop: 'hasSecret', label: t('sysSms.hasSecret'), width: 110, slot: 'hasSecret' },
  { prop: 'remark', label: t('sysSms.remark'), minWidth: 120, align: 'left' },
  { prop: 'isDefault', label: t('sysSms.defaultProvider'), width: 80, slot: 'isDefault' },
  { prop: 'enabled', label: t('sysSms.enabled'), width: 100, slot: 'enabled' },
  { prop: 'createTime', label: t('sysSms.createTime'), width: 170 },
])

const editForm = reactive<SysSmsConfigDTO>({
  id: undefined,
  providerId: '',
  accessKeyId: '',
  accessKeySecret: '',
  signName: '',
  region: '',
  appId: '',
  sender: '',
  templateConfig: '',
  isDefault: 0,
  enabled: 1,
  remark: '',
})

const editRules = computed<FormRules>(() => ({
  providerId: [{ required: true, message: t('common.pleaseInput') + t('sysSms.providerId'), trigger: 'blur' }],
}))

async function loadData() {
  loading.value = true
  try {
    const res = await getSmsConfigPage(queryForm)
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
  queryForm.providerId = ''
  queryForm.enabled = undefined
  handleQuery()
}

function handleAdd() {
  isEdit.value = false
  Object.assign(editForm, {
    id: undefined,
    providerId: '',
    accessKeyId: '',
    accessKeySecret: '',
    signName: '',
    region: '',
    appId: '',
    sender: '',
    templateConfig: '',
    isDefault: 0,
    enabled: 1,
    remark: '',
  })
  templateEntries.value = []
  dialogVisible.value = true
}

async function handleEdit(row: SysSmsConfigVO) {
  isEdit.value = true
  const detail = await getSmsConfigById(row.id)
  Object.assign(editForm, {
    id: detail.id,
    providerId: detail.providerId,
    accessKeyId: detail.accessKeyId || '',
    accessKeySecret: '',
    signName: detail.signName || '',
    region: detail.region || '',
    appId: detail.appId || '',
    sender: detail.sender || '',
    templateConfig: detail.templateConfig || '',
    isDefault: detail.isDefault,
    enabled: detail.enabled,
    remark: detail.remark || '',
  })
  templateEntries.value = parseTemplateConfig(detail.templateConfig)
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  // 将 key-value 映射同步回 editForm.templateConfig
  editForm.templateConfig = buildTemplateConfig(templateEntries.value)
  // json空值校验
  if (editForm.templateConfig === '') {
    // 字符串空串替换为 undefined
    editForm.templateConfig = undefined
  }

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateSmsConfig(editForm)
      ElMessage.success(t('sysSms.updateSuccess'))
    } else {
      await createSmsConfig(editForm)
      ElMessage.success(t('sysSms.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: SysSmsConfigVO) {
  await ElMessageBox.confirm(t('sysSms.confirmDelete'), t('sysSms.confirmDeleteTip'), { type: 'warning' })
  await deleteSmsConfig(row.id)
  ElMessage.success(t('sysSms.deleteSuccess'))
  loadData()
}

async function handleEnable(row: SysSmsConfigVO) {
  await ElMessageBox.confirm(t('sysSms.confirmEnable', { name: row.providerId }), t('sysSms.enableConfirm'), { type: 'warning' })
  await enableSmsConfig(row.id)
  ElMessage.success(t('sysSms.enableSuccess'))
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.template-mapper {
  width: 100%;
}

.template-mapper__row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
</style>
