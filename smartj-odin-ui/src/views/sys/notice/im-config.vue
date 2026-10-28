<template>
  <div class="sm-page">
    <SmCrudTable
      :data="tableData"
      :columns="columns"
      :loading="loading"
      :show-selection="false"
      :show-index="true"
      :show-pagination="false"
    >
      <template #title>{{ $t('sysImConfig.title') }}</template>
      <template #header-actions>
        <el-button
          type="primary"
          @click="handleAdd"
        >
          {{ $t('sysImConfig.addTitle') }}
        </el-button>
      </template>
      <template #providerId="{ row }">
        <el-tag size="small">{{ row.providerId === 'dingtalk' ? $t('sysImConfig.dingtalk') : row.providerId === 'wecom' ? $t('sysImConfig.wecom') : row.providerId }}</el-tag>
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
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          link
          type="danger"
          @click="handleDelete(row)"
        >
          {{ $t('common.delete') }}
        </el-button>
      </template>
    </SmCrudTable>

    <!-- 编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('sysImConfig.editTitle') : $t('sysImConfig.addTitle')"
      width="600px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="editForm"
        label-width="auto"
      >
        <el-form-item
          :label="$t('sysImConfig.providerId')"
          required
        >
          <el-select
            v-model="editForm.providerId"
            :placeholder="$t('common.pleaseSelect')"
            style="width: 100%"
          >
            <el-option
              :label="$t('sysImConfig.dingtalk')"
              value="dingtalk"
            />
            <el-option
              :label="$t('sysImConfig.wecom')"
              value="wecom"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          :label="$t('sysImConfig.robotName')"
          required
        >
          <el-input
            v-model="editForm.robotName"
            :placeholder="$t('sysImConfig.robotNamePlaceholder')"
            :maxlength="128"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysImConfig.webhookUrl')"
          required
        >
          <el-input
            v-model="editForm.webhookUrl"
            :placeholder="$t('sysImConfig.webhookUrlPlaceholder')"
            :maxlength="512"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sysImConfig.secret')">
          <el-input
            v-model="editForm.secret"
            :placeholder="$t('sysImConfig.secretPlaceholder')"
            type="password"
            show-password
            :maxlength="256"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sysImConfig.businessScene')">
          <el-select
            v-model="editForm.scene"
            :placeholder="$t('common.pleaseSelect') + $t('sysImConfig.scene')"
            style="width: 100%"
          >
            <el-option
              :label="$t('sysImConfig.systemAlert')"
              value="SYSTEM_ALERT"
            />
            <el-option
              :label="$t('sysImConfig.bizNotify')"
              value="BIZ_NOTIFY"
            />
            <el-option
              :label="$t('sysImConfig.auditNotify')"
              value="AUDIT_NOTIFY"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('sysImConfig.atMobiles')">
          <el-input
            v-model="editForm.atMobiles"
            :placeholder="$t('sysImConfig.atMobilesPlaceholder')"
            :maxlength="512"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sysImConfig.status')">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysImConfig.remark')">
          <el-input
            v-model="editForm.remark"
            :placeholder="$t('sysImConfig.remarkPlaceholder')"
            type="textarea"
            :rows="2"
            :maxlength="256"
            show-word-limit="true"
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
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import { getImConfigList, getImConfigById, createImConfig, updateImConfig, deleteImConfig } from '@/api/sys/notice'
import type { ImConfigVO } from '@/types/sys/notice'

const loading = ref(false)
const { t } = useI18n()
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()

const tableData = ref<ImConfigVO[]>([])

const editForm = reactive<ImConfigVO>({
  providerId: 'dingtalk',
  robotName: '',
  webhookUrl: '',
  secret: '',
  atMobiles: '',
  scene: '',
  status: 1,
  tenantId: '1',
  remark: '',
  createBy: '',
  createTime: '',
  updateBy: '',
  updateTime: '',
})

const columns = [
  { prop: 'robotName', label: t('sysImConfig.robotName'), minWidth: 160 },
  { prop: 'providerId', label: t('sysImConfig.providerName'), width: 100 },
  { prop: 'scene', label: t('sysImConfig.scene'), width: 120 },
  { prop: 'webhookUrl', label: 'Webhook', minWidth: 280, showOverflowTooltip: true },
  { prop: 'status', label: t('sysImConfig.status'), width: 80 },
  { prop: 'createTime', label: t('sysImConfig.createTime'), width: 170 },
]

async function loadData() {
  loading.value = true
  try {
    const res = await getImConfigList()
    tableData.value = res || []
  } finally {
    loading.value = false
  }
}

function resetForm() {
  editForm.id = undefined
  editForm.providerId = 'dingtalk'
  editForm.robotName = ''
  editForm.webhookUrl = ''
  editForm.secret = ''
  editForm.atMobiles = ''
  editForm.scene = ''
  editForm.status = 1
  editForm.remark = ''
}

function handleAdd() {
  resetForm()
  isEdit.value = false
  dialogVisible.value = true
}

async function handleEdit(row: ImConfigVO) {
  isEdit.value = true
  const res = await getImConfigById(row.id!)
  Object.assign(editForm, res)
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editForm.providerId || !editForm.robotName || !editForm.webhookUrl) {
    ElMessage.warning(t('sysImConfig.pleaseFillRequired'))
    return
  }
  // 手机号列表校验：非空时必须为字符串数组的 JSON
  if (editForm.atMobiles && editForm.atMobiles.trim()) {
    let parsed: unknown
    try {
      parsed = JSON.parse(editForm.atMobiles)
    } catch {
      ElMessage.warning(t('sysImConfig.atMobilesInvalid'))
      return
    }
    if (!Array.isArray(parsed) || parsed.some((v) => typeof v !== 'string')) {
      ElMessage.warning(t('sysImConfig.atMobilesInvalid'))
      return
    }
  }
  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateImConfig({ ...editForm })
      ElMessage.success(t('sysImConfig.updateSuccess'))
    } else {
      await createImConfig({ ...editForm })
      ElMessage.success(t('sysImConfig.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: ImConfigVO) {
  try {
    await ElMessageBox.confirm(t('sysImConfig.confirmDelete', { name: row.robotName }), t('sysImConfig.deleteConfirm'), { type: 'warning' })
    await deleteImConfig(row.id!)
    ElMessage.success(t('sysImConfig.deleteSuccess'))
    loadData()
  } catch {
    // 取消
  }
}

onMounted(() => {
  loadData()
})
</script>
