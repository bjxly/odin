<template>
  <div class="sm-page">
    <div class="sm-card" style="padding: 16px">
      <el-tabs v-model="activeTab">
        <!-- ============ 监控规则 ============ -->
        <el-tab-pane :label="t('odinInsight.tabMonitors')" name="monitors">
          <SmSearchForm :model="monitorQuery" @search="handleMonitorQuery" @reset="handleMonitorReset">
            <el-form-item :label="t('odinInsight.monitorName')" prop="name">
              <el-input v-model="monitorQuery.name" clearable :placeholder="t('common.pleaseInput') + t('odinInsight.monitorName')" />
            </el-form-item>
            <el-form-item :label="t('odinInsight.ontology')" prop="ontologyId">
              <el-select v-model="monitorQuery.ontologyId" clearable :placeholder="t('common.pleaseSelect')" style="width: 200px">
                <el-option v-for="o in ontologyOptions" :key="o.id" :label="o.name" :value="o.id" />
              </el-select>
            </el-form-item>
            <el-form-item :label="t('odinInsight.enabled')" prop="enabled">
              <el-select v-model="monitorQuery.enabled" clearable :placeholder="t('common.pleaseSelect')" style="width: 120px">
                <el-option :label="t('common.yes')" :value="1" />
                <el-option :label="t('common.no')" :value="0" />
              </el-select>
            </el-form-item>
          </SmSearchForm>
          <SmCrudTable
            v-model:page="monitorQuery.current"
            v-model:page-size="monitorQuery.size"
            :data="monitorData"
            :columns="monitorColumns"
            :total="monitorTotal"
            :loading="monitorLoading"
            :show-selection="true"
            :operate-width="220"
            @page-change="loadMonitors"
            @size-change="loadMonitors"
            @selection-change="handleMonitorSelectionChange"
          >
            <template #title>{{ t('odinInsight.monitorList') }}</template>
            <template #header-actions>
              <el-button v-permission="'odin:insight:create'" type="primary" @click="handleAddMonitor">
                {{ t('common.add') }}
              </el-button>
              <el-button v-permission="'odin:insight:delete'" type="danger" :disabled="!selectedMonitorIds.length" @click="handleBatchDeleteMonitor">
                {{ t('common.batchDelete') }}
              </el-button>
            </template>
            <template #enabled="{ row }">
              <el-switch
                v-permission="'odin:insight:update'"
                :model-value="row.enabled === 1"
                @change="(v: boolean) => handleToggleEnabled(row, v)"
              />
            </template>
            <template #status="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                {{ row.status === 1 ? t('common.enabled') : t('common.disabled') }}
              </el-tag>
            </template>
            <template #operate="{ row }">
              <el-button v-permission="'odin:insight:update'" link type="primary" size="small" @click="handleEditMonitor(row)">
                {{ t('common.edit') }}
              </el-button>
              <el-button v-permission="'odin:insight:run'" link type="primary" size="small" :loading="executingId === row.id" @click="handleExecuteMonitor(row)">
                {{ t('odinInsight.execute') }}
              </el-button>
              <el-button v-permission="'odin:insight:delete'" link type="danger" size="small" @click="handleDeleteMonitor(row)">
                {{ t('common.delete') }}
              </el-button>
            </template>
          </SmCrudTable>
        </el-tab-pane>

        <!-- ============ 告警记录 ============ -->
        <el-tab-pane :label="t('odinInsight.tabAlerts')" name="alerts">
          <SmSearchForm :model="alertQuery" @search="handleAlertQuery" @reset="handleAlertReset">
            <el-form-item :label="t('odinInsight.severity')" prop="severity">
              <el-select v-model="alertQuery.severity" clearable :placeholder="t('common.pleaseSelect')" style="width: 140px">
                <el-option :label="t('odinInsight.sevInfo')" value="info" />
                <el-option :label="t('odinInsight.sevWarning')" value="warning" />
                <el-option :label="t('odinInsight.sevCritical')" value="critical" />
              </el-select>
            </el-form-item>
            <el-form-item :label="t('odinInsight.handled')" prop="handled">
              <el-select v-model="alertQuery.handled" clearable :placeholder="t('common.pleaseSelect')" style="width: 140px">
                <el-option :label="t('odinInsight.handledYes')" :value="1" />
                <el-option :label="t('odinInsight.handledNo')" :value="0" />
              </el-select>
            </el-form-item>
          </SmSearchForm>
          <SmCrudTable
            v-model:page="alertQuery.current"
            v-model:page-size="alertQuery.size"
            :data="alertData"
            :columns="alertColumns"
            :total="alertTotal"
            :loading="alertLoading"
            :operate-width="180"
            @page-change="loadAlerts"
            @size-change="loadAlerts"
          >
            <template #title>{{ t('odinInsight.alertList') }}</template>
            <template #severity="{ row }">
              <el-tag :type="severityType(row.severity)" size="small">
                {{ severityLabel(row.severity) }}
              </el-tag>
            </template>
            <template #handled="{ row }">
              <el-tag :type="row.handled === 1 ? 'success' : 'warning'" size="small">
                {{ row.handled === 1 ? t('odinInsight.handledYes') : t('odinInsight.handledNo') }}
              </el-tag>
            </template>
            <template #operate="{ row }">
              <el-button link type="primary" size="small" @click="handleViewAlert(row)">
                {{ t('odinInsight.viewDetail') }}
              </el-button>
              <el-button v-if="row.handled !== 1" v-permission="'odin:insight:handle'" link type="primary" size="small" @click="handleOpenHandle(row)">
                {{ t('odinInsight.handle') }}
              </el-button>
            </template>
          </SmCrudTable>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 监控规则 新增/编辑 -->
    <el-dialog
      v-model="monitorDialogVisible"
      :title="isEditMonitor ? t('odinInsight.editMonitor') : t('odinInsight.addMonitor')"
      width="720px"
      destroy-on-close
    >
      <el-form ref="monitorFormRef" :model="monitorForm" :rules="monitorRules" label-width="140px">
        <el-form-item :label="t('odinInsight.monitorName')" prop="name">
          <el-input v-model="monitorForm.name" :maxlength="128" />
        </el-form-item>
        <el-form-item :label="t('odinInsight.description')" prop="description">
          <el-input v-model="monitorForm.description" type="textarea" :rows="2" :maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('odinInsight.ontology')" prop="ontologyId">
          <el-select v-model="monitorForm.ontologyId" :placeholder="t('common.pleaseSelect')" style="width: 100%">
            <el-option v-for="o in ontologyOptions" :key="o.id" :label="o.name" :value="o.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('odinInsight.ruleExpression')" prop="ruleExpression">
          <el-input v-model="monitorForm.ruleExpression" type="textarea" :rows="4" :maxlength="2000" show-word-limit placeholder="SQL / Expression" />
        </el-form-item>
        <el-form-item :label="t('odinInsight.scheduleCron')" prop="scheduleCron">
          <el-input v-model="monitorForm.scheduleCron" placeholder="0 0 * * * ?" :maxlength="64" />
        </el-form-item>
        <el-form-item :label="t('odinInsight.thresholds')" prop="thresholds">
          <el-input v-model="monitorForm.thresholds" type="textarea" :rows="3" placeholder='JSON: {"warn": 100, "critical": 500}' />
        </el-form-item>
        <el-form-item :label="t('odinInsight.notificationChannels')" prop="notificationChannels">
          <el-input v-model="monitorForm.notificationChannels" type="textarea" :rows="2" placeholder='JSON: ["email","dingtalk"]' />
        </el-form-item>
        <el-form-item :label="t('odinInsight.enabled')" prop="enabled">
          <el-switch v-model="monitorForm.enabledBool" />
        </el-form-item>
        <el-form-item :label="t('odinInsight.status')" prop="status">
          <el-radio-group v-model="monitorForm.status">
            <el-radio :value="1">{{ t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('odinInsight.remark')" prop="remark">
          <el-input v-model="monitorForm.remark" type="textarea" :rows="2" :maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="monitorDialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
        <el-button type="primary" :loading="monitorSubmitLoading" @click="handleSubmitMonitor">{{ t('common.confirmBtn') }}</el-button>
      </template>
    </el-dialog>

    <!-- 告警详情 -->
    <el-dialog v-model="alertDetailVisible" :title="t('odinInsight.alertDetail')" width="720px" destroy-on-close>
      <el-descriptions v-if="alertDetail" :column="1" border size="small">
        <el-descriptions-item :label="t('odinInsight.monitorName')">{{ alertDetail.monitorName || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('odinInsight.triggerTime')">{{ alertDetail.triggerTime || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('odinInsight.severity')">
          <el-tag :type="severityType(alertDetail.severity)" size="small">{{ severityLabel(alertDetail.severity) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinInsight.alertTitle')">{{ alertDetail.title || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('odinInsight.alertDetail')">
          <pre class="pre-block">{{ alertDetail.detail || '-' }}</pre>
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinInsight.metrics')">
          <pre class="pre-block">{{ formatJson(alertDetail.metricValues) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinInsight.handled')">
          <el-tag :type="alertDetail.handled === 1 ? 'success' : 'warning'" size="small">
            {{ alertDetail.handled === 1 ? t('odinInsight.handledYes') : t('odinInsight.handledNo') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="alertDetail.handled === 1" :label="t('odinInsight.handleNote')">
          {{ alertDetail.handleNote || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="alertDetail.handledTime" :label="t('odinInsight.handledTime')">
          {{ alertDetail.handledTime }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 处理告警 -->
    <el-dialog v-model="handleDialogVisible" :title="t('odinInsight.handleAlert')" width="560px" destroy-on-close>
      <el-form ref="handleFormRef" :model="handleForm" :rules="handleRules" label-width="100px">
        <el-form-item :label="t('odinInsight.alertTitle')">
          <span>{{ handleTarget?.title || '-' }}</span>
        </el-form-item>
        <el-form-item :label="t('odinInsight.handleNote')" prop="handleNote">
          <el-input v-model="handleForm.handleNote" type="textarea" :rows="4" :maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleDialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
        <el-button type="primary" :loading="handleSubmitting" @click="handleSubmitHandle">{{ t('common.confirmBtn') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import type {
  InsightMonitorVO,
  InsightMonitorDTO,
  InsightMonitorQuery,
  InsightAlertVO,
  AlertQuery,
} from '@/types/odin/insight'
import {
  getMonitorPage,
  getMonitorById,
  createMonitor,
  updateMonitor,
  deleteMonitor,
  runMonitor,
  getAlertPage,
  handleAlert,
} from '@/api/odin/insight'
import { getOntologyList } from '@/api/odin/ontology'
import type { OntologyVO } from '@/types/odin/ontology'

const { t } = useI18n()

const activeTab = ref('monitors')
const ontologyOptions = ref<OntologyVO[]>([])

async function loadOntologyOptions() {
  try {
    ontologyOptions.value = await getOntologyList()
  } catch {
    ontologyOptions.value = []
  }
}

// ===== 监控规则 =====
const monitorLoading = ref(false)
const monitorData = ref<InsightMonitorVO[]>([])
const monitorTotal = ref(0)
const selectedMonitorIds = ref<string[]>([])
const monitorQuery = reactive<InsightMonitorQuery>({
  current: 1,
  size: 10,
  name: '',
  ontologyId: undefined,
  enabled: undefined,
})

const monitorColumns = computed<TableColumn[]>(() => [
  { prop: 'name', label: t('odinInsight.monitorName'), minWidth: 160, align: 'left' },
  { prop: 'ontologyName', label: t('odinInsight.ontology'), minWidth: 120, align: 'left' },
  { prop: 'scheduleCron', label: t('odinInsight.scheduleCron'), width: 140, align: 'left' },
  { prop: 'enabled', label: t('odinInsight.enabled'), width: 90, slot: 'enabled' },
  { prop: 'alertCount', label: t('odinInsight.alertCount'), width: 90, align: 'right' },
  { prop: 'status', label: t('odinInsight.status'), width: 80, slot: 'status' },
  { prop: 'createTime', label: t('odinInsight.createTime'), width: 170 },
])

async function loadMonitors() {
  monitorLoading.value = true
  try {
    const res = await getMonitorPage(monitorQuery)
    monitorData.value = res.records
    monitorTotal.value = res.total
  } finally {
    monitorLoading.value = false
  }
}

function handleMonitorQuery() {
  monitorQuery.current = 1
  loadMonitors()
}

function handleMonitorReset() {
  monitorQuery.name = ''
  monitorQuery.ontologyId = undefined
  monitorQuery.enabled = undefined
}

function handleMonitorSelectionChange(rows: InsightMonitorVO[]) {
  selectedMonitorIds.value = rows.map((r) => r.id)
}

const monitorDialogVisible = ref(false)
const isEditMonitor = ref(false)
const monitorSubmitLoading = ref(false)
const monitorFormRef = ref<FormInstance>()
const monitorForm = reactive<InsightMonitorDTO & { enabledBool: boolean }>({
  id: undefined,
  name: '',
  description: '',
  ontologyId: '',
  ruleExpression: '',
  scheduleCron: '',
  thresholds: '',
  notificationChannels: '',
  enabled: 1,
  enabledBool: true,
  status: 1,
  remark: '',
})

const monitorRules = computed<FormRules>(() => ({
  name: [{ required: true, message: t('common.pleaseInput') + t('odinInsight.monitorName'), trigger: 'blur' }],
  ontologyId: [{ required: true, message: t('common.pleaseSelect') + t('odinInsight.ontology'), trigger: 'change' }],
  ruleExpression: [{ required: true, message: t('common.pleaseInput') + t('odinInsight.ruleExpression'), trigger: 'blur' }],
  scheduleCron: [{ required: true, message: t('common.pleaseInput') + t('odinInsight.scheduleCron'), trigger: 'blur' }],
}))

watch(
  () => monitorForm.enabledBool,
  (v) => {
    monitorForm.enabled = v ? 1 : 0
  }
)

function resetMonitorForm() {
  monitorForm.id = undefined
  monitorForm.name = ''
  monitorForm.description = ''
  monitorForm.ontologyId = ''
  monitorForm.ruleExpression = ''
  monitorForm.scheduleCron = ''
  monitorForm.thresholds = ''
  monitorForm.notificationChannels = ''
  monitorForm.enabled = 1
  monitorForm.enabledBool = true
  monitorForm.status = 1
  monitorForm.remark = ''
}

function handleAddMonitor() {
  isEditMonitor.value = false
  resetMonitorForm()
  monitorDialogVisible.value = true
}

async function handleEditMonitor(row: InsightMonitorVO) {
  isEditMonitor.value = true
  const detail = await getMonitorById(row.id)
  Object.assign(monitorForm, {
    id: detail.id,
    name: detail.name,
    description: detail.description ?? '',
    ontologyId: detail.ontologyId ?? '',
    ruleExpression: detail.ruleExpression ?? '',
    scheduleCron: detail.scheduleCron ?? '',
    thresholds: detail.thresholds ?? '',
    notificationChannels: detail.notificationChannels ?? '',
    enabled: detail.enabled ?? 1,
    enabledBool: detail.enabled === 1,
    status: detail.status,
    remark: detail.remark ?? '',
  })
  monitorDialogVisible.value = true
}

async function handleSubmitMonitor() {
  const valid = await monitorFormRef.value?.validate().catch(() => false)
  if (!valid) return
  monitorSubmitLoading.value = true
  try {
    const { enabledBool: _enabledBool, ...payload } = monitorForm
    if (isEditMonitor.value) {
      await updateMonitor(payload)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createMonitor(payload)
      ElMessage.success(t('common.createSuccess'))
    }
    monitorDialogVisible.value = false
    loadMonitors()
  } finally {
    monitorSubmitLoading.value = false
  }
}

async function handleDeleteMonitor(row: InsightMonitorVO) {
  await ElMessageBox.confirm(t('odinInsight.confirmDeleteMonitor'), t('common.tip'), { type: 'warning' })
  await deleteMonitor(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadMonitors()
}

async function handleBatchDeleteMonitor() {
  if (!selectedMonitorIds.value.length) return
  await ElMessageBox.confirm(t('odinInsight.confirmBatchDeleteMonitor'), t('common.tip'), { type: 'warning' })
  // 后端未提供批量接口，串行逐条删除
  for (const id of selectedMonitorIds.value) {
    await deleteMonitor(id)
  }
  ElMessage.success(t('common.deleteSuccess'))
  selectedMonitorIds.value = []
  loadMonitors()
}

const executingId = ref('')

async function handleExecuteMonitor(row: InsightMonitorVO) {
  if (executingId.value) return
  executingId.value = row.id
  try {
    await runMonitor(row.id)
    ElMessage.success(t('odinInsight.executeOk'))
    // 执行后刷新告警列表
    if (activeTab.value === 'alerts') await loadAlerts()
  } finally {
    executingId.value = ''
  }
}

async function handleToggleEnabled(row: InsightMonitorVO, val: boolean) {
  try {
    await updateMonitor({ id: row.id, name: row.name, enabled: val ? 1 : 0 })
    row.enabled = val ? 1 : 0
    ElMessage.success(val ? t('odinInsight.enabledOn') : t('odinInsight.enabledOff'))
  } catch {
    ElMessage.error(t('common.operationFail'))
  }
}

// ===== 告警 =====
const alertLoading = ref(false)
const alertData = ref<InsightAlertVO[]>([])
const alertTotal = ref(0)
const alertQuery = reactive<AlertQuery>({
  current: 1,
  size: 10,
  severity: undefined,
  handled: undefined,
})

const alertColumns = computed<TableColumn[]>(() => [
  { prop: 'monitorName', label: t('odinInsight.monitorName'), minWidth: 160, align: 'left' },
  { prop: 'triggerTime', label: t('odinInsight.triggerTime'), width: 170 },
  { prop: 'severity', label: t('odinInsight.severity'), width: 100, slot: 'severity' },
  { prop: 'title', label: t('odinInsight.alertTitle'), minWidth: 200, align: 'left' },
  { prop: 'handled', label: t('odinInsight.handled'), width: 100, slot: 'handled' },
])

async function loadAlerts() {
  alertLoading.value = true
  try {
    const res = await getAlertPage(alertQuery)
    alertData.value = res.records
    alertTotal.value = res.total
  } finally {
    alertLoading.value = false
  }
}

function handleAlertQuery() {
  alertQuery.current = 1
  loadAlerts()
}

function handleAlertReset() {
  alertQuery.severity = undefined
  alertQuery.handled = undefined
}

function severityType(sev: string): 'info' | 'warning' | 'danger' | 'success' {
  switch (sev) {
    case 'critical':
      return 'danger'
    case 'warning':
      return 'warning'
    case 'info':
      return 'info'
    default:
      return 'info'
  }
}

function severityLabel(sev: string): string {
  switch (sev) {
    case 'critical':
      return t('odinInsight.sevCritical')
    case 'warning':
      return t('odinInsight.sevWarning')
    case 'info':
      return t('odinInsight.sevInfo')
    default:
      return sev
  }
}

function formatJson(v: unknown): string {
  if (v == null || v === '') return '-'
  if (typeof v === 'string') return v
  try {
    return JSON.stringify(v, null, 2)
  } catch {
    return String(v)
  }
}

const alertDetailVisible = ref(false)
const alertDetail = ref<InsightAlertVO | null>(null)

function handleViewAlert(row: InsightAlertVO) {
  alertDetail.value = row
  alertDetailVisible.value = true
}

const handleDialogVisible = ref(false)
const handleSubmitting = ref(false)
const handleTarget = ref<InsightAlertVO | null>(null)
const handleFormRef = ref<FormInstance>()
const handleForm = reactive({ handleNote: '' })

const handleRules = computed<FormRules>(() => ({
  handleNote: [{ required: true, message: t('common.pleaseInput') + t('odinInsight.handleNote'), trigger: 'blur' }],
}))

function handleOpenHandle(row: InsightAlertVO) {
  handleTarget.value = row
  handleForm.handleNote = ''
  handleDialogVisible.value = true
}

async function handleSubmitHandle() {
  const valid = await handleFormRef.value?.validate().catch(() => false)
  if (!valid || !handleTarget.value) return
  handleSubmitting.value = true
  try {
    await handleAlert(handleTarget.value.id, { handleNote: handleForm.handleNote })
    ElMessage.success(t('odinInsight.handleSuccess'))
    handleDialogVisible.value = false
    loadAlerts()
  } finally {
    handleSubmitting.value = false
  }
}

// Tab 切换时懒加载
watch(activeTab, (tab) => {
  if (tab === 'alerts' && !alertData.value.length) loadAlerts()
})

onMounted(async () => {
  await loadOntologyOptions()
  await loadMonitors()
})
</script>

<style scoped>
.pre-block {
  margin: 0;
  padding: 8px 12px;
  background: var(--sm-bg-page);
  border-radius: 4px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 260px;
  overflow: auto;
}
</style>
