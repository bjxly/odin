<template>
  <div class="sm-page">
    <SmCrudTable
      v-model:page="queryForm.current"
      v-model:page-size="queryForm.size"
      :data="tableData"
      :columns="columns"
      :total="total"
      :loading="loading"
      :show-selection="false"
      :show-index="true"
      :show-pagination="false"
    >
      <template #title>{{ $t('sysNoticeTemplate.title') }}</template>
      <template #header-actions>
        <el-button
          type="primary"
          @click="handleAdd"
        >
          {{ $t('sysNoticeTemplate.addTitle') }}
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
      <template #channel="{ row }">
        <el-tag size="small">
          {{ row.channel === 'EMAIL' ? $t('sysNotice.email') : row.channel === 'SMS' ? $t('sysNotice.sms') : row.channel === 'SITE' ? $t('sysNotice.site') : row.channel }}
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
          type="warning"
          @click="handlePreview(row)"
        >
          {{ $t('sysNoticeTemplate.preview') }}
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
      :title="isEdit ? $t('sysNoticeTemplate.editTitle') : $t('sysNoticeTemplate.addTitle')"
      width="650px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="editForm"
        label-width="auto"
      >
        <el-form-item
          :label="$t('sysNoticeTemplate.templateCode')"
          required
        >
          <el-input
            v-model="editForm.templateCode"
            :placeholder="$t('sysNoticeTemplate.templateCodePlaceholder')"
            :maxlength="64"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysNoticeTemplate.templateName')"
          required
        >
          <el-input
            v-model="editForm.templateName"
            :placeholder="$t('sysNoticeTemplate.templateNamePlaceholder')"
            :maxlength="128"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysNoticeTemplate.channel')"
          required
        >
          <el-select
            v-model="editForm.channel"
            :placeholder="$t('common.pleaseSelect') + $t('sysNoticeTemplate.channel')"
            style="width: 100%"
          >
            <el-option
              :label="$t('sysNotice.email')"
              value="EMAIL"
            />
            <el-option
              :label="$t('sysNotice.sms')"
              value="SMS"
            />
            <el-option
              :label="$t('sysNotice.site')"
              value="SITE"
            />
            <el-option
              :label="$t('sysNotice.im')"
              value="IM"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('sysNoticeTemplate.titleTemplate')">
          <el-input
            v-model="editForm.titleTemplate"
            :placeholder="$t('sysNoticeTemplate.titleTemplatePlaceholder')"
            :maxlength="256"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysNoticeTemplate.contentTemplate')"
          required
        >
          <el-input
            v-model="editForm.contentTemplate"
            type="textarea"
            :rows="5"
            :placeholder="$t('sysNoticeTemplate.contentTemplatePlaceholder')"
            :maxlength="512"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sysNoticeTemplate.variableList')">
          <el-input
            v-model="editForm.variables"
            :placeholder="$t('sysNoticeTemplate.variablesPlaceholder')"
            :maxlength="512"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item :label="$t('sysNoticeTemplate.status')">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysNoticeTemplate.remark')">
          <el-input
            v-model="editForm.remark"
            :placeholder="$t('sysNoticeTemplate.remarkPlaceholder')"
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

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewVisible"
      :title="$t('sysNoticeTemplate.previewTitle')"
      width="550px"
    >
      <el-form label-width="auto">
        <el-form-item :label="$t('sysNoticeTemplate.previewLabelCode')">
          <span>{{ previewTemplate?.templateCode }}</span>
        </el-form-item>
        <el-form-item :label="$t('sysNoticeTemplate.previewLabelVariables')">
          <el-input
            v-model="previewVariablesJson"
            type="textarea"
            :rows="3"
            :placeholder="$t('sysNoticeTemplate.previewVariablesPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="$t('sysNoticeTemplate.previewLabelTitle')">
          <el-input
            :model-value="previewTitle"
            readonly
          />
        </el-form-item>
        <el-form-item :label="$t('sysNoticeTemplate.previewLabelContent')">
          <el-input
            :model-value="previewContent"
            readonly
            type="textarea"
            :rows="4"
          />
        </el-form-item>
      </el-form>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import { getNoticeTemplateList, getNoticeTemplateById, createNoticeTemplate, updateNoticeTemplate, deleteNoticeTemplate, previewNoticeTemplate } from '@/api/sys/notice'
import type { NoticeTemplateVO } from '@/types/sys/notice'

const loading = ref(false)
const { t } = useI18n()
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()

const tableData = ref<NoticeTemplateVO[]>([])
const total = ref(0)

const queryForm = reactive({ current: 1, size: 100 })

const editForm = reactive<NoticeTemplateVO>({
  templateCode: '',
  templateName: '',
  channel: 'SITE',
  titleTemplate: '',
  contentTemplate: '',
  variables: '',
  status: 1,
  tenantId: '1',
  remark: '',
  delFlag: 0,
  createBy: '',
  createTime: '',
  updateBy: '',
  updateTime: '',
})

const columns = [
  { prop: 'templateCode', label: t('sysNoticeTemplate.templateCode'), minWidth: 180 },
  { prop: 'templateName', label: t('sysNoticeTemplate.templateName'), minWidth: 150 },
  { prop: 'channel', label: t('sysNoticeTemplate.channel'), width: 100 },
  { prop: 'contentTemplate', label: t('sysNoticeTemplate.contentTemplate'), minWidth: 300, showOverflowTooltip: true },
  { prop: 'status', label: t('sysNoticeTemplate.status'), width: 80 },
  { prop: 'createTime', label: t('sysNoticeTemplate.createTime'), width: 170 },
]

async function loadData() {
  loading.value = true
  try {
    const res = await getNoticeTemplateList()
    tableData.value = res || []
    total.value = (res || []).length
  } finally {
    loading.value = false
  }
}

function resetForm() {
  editForm.id = undefined
  editForm.templateCode = ''
  editForm.templateName = ''
  editForm.channel = 'SITE'
  editForm.titleTemplate = ''
  editForm.contentTemplate = ''
  editForm.variables = ''
  editForm.status = 1
  editForm.remark = ''
}

function handleAdd() {
  resetForm()
  isEdit.value = false
  dialogVisible.value = true
}

async function handleEdit(row: NoticeTemplateVO) {
  isEdit.value = true
  const res = await getNoticeTemplateById(row.id!)
  Object.assign(editForm, res)
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editForm.templateCode || !editForm.templateName || !editForm.contentTemplate) {
    ElMessage.warning(t('sysNoticeTemplate.pleaseFillRequired'))
    return
  }
  // 变量列表校验：非空时必须为字符串数组的 JSON
  if (editForm.variables && editForm.variables.trim()) {
    let parsed: unknown
    try {
      parsed = JSON.parse(editForm.variables)
    } catch {
      ElMessage.warning(t('sysNoticeTemplate.variablesInvalid'))
      return
    }
    if (!Array.isArray(parsed) || parsed.some((v) => typeof v !== 'string')) {
      ElMessage.warning(t('sysNoticeTemplate.variablesInvalid'))
      return
    }
  }
  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateNoticeTemplate({ ...editForm })
      ElMessage.success(t('sysNoticeTemplate.updateSuccess'))
    } else {
      await createNoticeTemplate({ ...editForm })
      ElMessage.success(t('sysNoticeTemplate.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: NoticeTemplateVO) {
  try {
    await ElMessageBox.confirm(t('sysNoticeTemplate.confirmDelete', { name: row.templateName }), t('sysNoticeTemplate.deleteConfirm'), { type: 'warning' })
    await deleteNoticeTemplate(row.id!)
    ElMessage.success(t('sysNoticeTemplate.deleteSuccess'))
    loadData()
  } catch {
    // 取消
  }
}

// 预览相关
const previewVisible = ref(false)
const previewTemplate = ref<NoticeTemplateVO | null>(null)
const previewVariablesJson = ref('')
const previewTitle = ref('')
const previewContent = ref('')

function handlePreview(row: NoticeTemplateVO) {
  previewTemplate.value = row
  // 兼容历史脏数据：变量列表非法时按空数组处理，避免预览弹窗打不开
  let keys: string[] = []
  if (row.variables) {
    try {
      const parsed = JSON.parse(row.variables)
      if (Array.isArray(parsed)) {
        keys = parsed.filter((k): k is string => typeof k === 'string')
      }
    } catch {
      keys = []
    }
  }
  previewVariablesJson.value = JSON.stringify(
    keys.reduce((acc: Record<string, string>, k) => {
      acc[k] = ''
      return acc
    }, {}),
    null,
    2
  )
  previewTitle.value = ''
  previewContent.value = ''
  previewVisible.value = true
}

async function renderPreview(silent = false) {
  if (!previewTemplate.value) return
  // eslint-disable-next-line no-useless-assignment
  let variables: Record<string, unknown> = {}
  try {
    variables = JSON.parse(previewVariablesJson.value)
  } catch {
    // 实时输入过程中解析失败不弹警告，保留上次渲染结果
    if (!silent) ElMessage.warning(t('sysNoticeTemplate.previewJsonInvalid'))
    return
  }
  const res = await previewNoticeTemplate({
    templateCode: previewTemplate.value.templateCode,
    variables,
  })
  previewTitle.value = res.title
  previewContent.value = res.content
}

// 监听变量值输入，防抖自动刷新预览正文
let previewDebounceTimer: ReturnType<typeof setTimeout> | null = null
watch(previewVariablesJson, () => {
  if (!previewVisible.value) return
  if (previewDebounceTimer) clearTimeout(previewDebounceTimer)
  previewDebounceTimer = setTimeout(() => renderPreview(true), 300)
})

// 监听预览弹窗
watch(previewVisible, (val) => {
  if (val && previewTemplate.value) {
    renderPreview()
  } else if (!val && previewDebounceTimer) {
    clearTimeout(previewDebounceTimer)
    previewDebounceTimer = null
  }
})

onMounted(() => {
  loadData()
})
</script>
