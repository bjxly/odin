<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sysConfig.configKey')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.configKey"
          :placeholder="$t('common.pleaseInput') + $t('sysConfig.configKey')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysConfig.configName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.configName"
          :placeholder="$t('common.pleaseInput') + $t('sysConfig.configName')"
          clearable
        />
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
      @selection-change="handleSelectionChange"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sysConfig.list') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sys:config:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
        </el-button>
        <el-button
          v-permission="'sys:config:delete'"
          type="danger"
          :disabled="!selectedIds.length || hasSystemSelected"
          @click="handleBatchDelete"
        >
          {{ $t('common.batchDelete') }}
        </el-button>
      </template>
      <template #configType="{ row }">
        <el-tag
          :type="row.configType === 1 ? 'success' : 'info'"
          size="small"
        >
          {{ row.configType === 1 ? $t('common.yes') : $t('common.no') }}
        </el-tag>
      </template>
      <template #operate="{ row }">
        <el-button
          v-permission="'sys:config:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-permission="'sys:config:delete'"
          link
          type="danger"
          :disabled="row.configType === 1 ? true : false"
          @click="handleDelete(row)"
        >
          {{ $t('common.delete') }}
        </el-button>
      </template>
    </SmCrudTable>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('sysConfig.editTitle') : $t('sysConfig.addTitle')"
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
          :label="$t('sysConfig.configName')"
          prop="configName"
        >
          <el-input
            v-model="editForm.configName"
            :placeholder="$t('common.pleaseInput') + $t('sysConfig.configName')"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysConfig.configKey')"
          prop="configKey"
        >
          <el-input
            v-model="editForm.configKey"
            :placeholder="$t('common.pleaseInput') + $t('sysConfig.configKey')"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysConfig.configValue')"
          prop="configValue"
        >
          <el-input
            v-model="editForm.configValue"
            :placeholder="$t('common.pleaseInput') + $t('sysConfig.configValue')"
            :maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysConfig.isSystem')">
          <el-radio-group v-model="editForm.configType">
            <el-radio :value="1">{{ $t('common.yes') }}</el-radio>
            <el-radio :value="0">{{ $t('common.no') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysConfig.remark')">
          <el-input
            v-model="editForm.remark"
            type="textarea"
            :rows="2"
            :placeholder="$t('common.pleaseInput') + $t('sysConfig.remark')"
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getConfigPage, getConfigById, createConfig, updateConfig, deleteConfigs } from '@/api/sys/config'
import type { ConfigVO, ConfigDTO } from '@/types/sys'
import { SSO_PUBLIC_CONFIG_KEY } from '@/utils/constants'

const { t } = useI18n()

const loading = ref(false)
const submitLoading = ref(false)
const tableData = ref<ConfigVO[]>([])
const total = ref(0)
const selectedIds = ref<string[]>([])
const selectedRows = ref<ConfigVO[]>([])
/** 选中的数据中是否包含系统内置参数（configType === 1），包含时禁止批量删除 */
const hasSystemSelected = computed(() => selectedRows.value.some((r) => r.configType === 1))
const dialogVisible = ref(false)
const isEdit = ref(false)

const editFormRef = ref<FormInstance>()

const queryForm = reactive({
  current: 1,
  size: 100,
  configName: '',
  configKey: '',
})

const columns = computed<TableColumn[]>(() => [
  { prop: 'configKey', label: t('sysConfig.configKey'), minWidth: 200, align: 'left' },
  { prop: 'configName', label: t('sysConfig.configName'), minWidth: 180, align: 'left' },
  { prop: 'configValue', label: t('sysConfig.configValue'), minWidth: 150, align: 'left' },
  { prop: 'remark', label: t('sysConfig.remark'), minWidth: 150, align: 'left' },
  { prop: 'configType', label: t('sysConfig.isSystem'), width: 130, slot: 'configType' },
  { prop: 'createTime', label: t('sysConfig.createTime'), width: 170 },
])

const editForm = reactive<ConfigDTO>({
  id: undefined,
  configName: '',
  configKey: '',
  configValue: '',
  configType: undefined,
  remark: '',
})

const editRules = computed<FormRules>(() => ({
  configName: [{ required: true, message: t('common.pleaseInput') + t('sysConfig.configName'), trigger: 'blur' }],
  configKey: [{ required: true, message: t('common.pleaseInput') + t('sysConfig.configKey'), trigger: 'blur' }],
  configValue: [{ required: true, message: t('common.pleaseInput') + t('sysConfig.configValue'), trigger: 'blur' }],
}))

async function loadData() {
  loading.value = true
  try {
    const res = await getConfigPage(queryForm)
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
  queryForm.configName = ''
  queryForm.configKey = ''
  handleQuery()
}

function handleSelectionChange(rows: ConfigVO[]) {
  selectedRows.value = rows
  selectedIds.value = rows.map((r) => r.id)
}

function handleAdd() {
  isEdit.value = false
  Object.assign(editForm, { id: undefined, configName: '', configKey: '', configValue: '', configType: 1, remark: '' })
  dialogVisible.value = true
}

async function handleEdit(row: ConfigVO) {
  isEdit.value = true
  const detail = await getConfigById(row.id)
  Object.assign(editForm, detail)
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateConfig(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createConfig(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    // 系统参数变更后清除前端 SSO 公开配置缓存，使下次访问重新从后端拉取
    sessionStorage.removeItem(SSO_PUBLIC_CONFIG_KEY)
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ConfigVO) {
  await ElMessageBox.confirm(t('sysConfig.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteConfigs(row.id)
  sessionStorage.removeItem(SSO_PUBLIC_CONFIG_KEY)
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

async function handleBatchDelete() {
  await ElMessageBox.confirm(t('sysConfig.confirmDeleteConfigs', { count: selectedIds.value.length }), t('common.tip'), { type: 'warning' })
  for (let id of selectedIds.value) {
    await deleteConfigs(id)
  }
  sessionStorage.removeItem(SSO_PUBLIC_CONFIG_KEY)
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

onMounted(() => {
  loadData()
})
</script>
