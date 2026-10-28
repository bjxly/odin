<template>
  <div class="reason-lab">
    <!-- 页头统计条 -->
    <div class="lab-header">
      <OntologySwitcher v-model="activeOntologyId" />
      <div class="header-stats">
        <div class="stat-item">
          <span class="stat-value">{{ rules.length }}</span>
          <span class="stat-label">{{ t('odinReason.rules') }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-value stat-enabled">{{ enabledCount }}</span>
          <span class="stat-label">{{ t('odinReason.enabled') }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ lastExecTime || '—' }}</span>
          <span class="stat-label">{{ t('odinReason.lastExec') }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ successRate }}%</span>
          <span class="stat-label">{{ t('odinReason.successRate') }}</span>
        </div>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="lab-body">
      <!-- 左侧规则面板 -->
      <aside class="lab-aside">
        <div class="aside-header">
          <el-button v-permission="'odin:reason:save'" type="primary" size="small" style="width: 100%" @click="handleAddRule">
            + {{ t('odinReason.addRule') }}
          </el-button>
        </div>
        <div v-loading="rulesLoading" class="aside-list">
          <template v-for="group in groupedRules" :key="group.type">
            <div class="group-label">{{ group.type }}</div>
            <div
              v-for="rule in group.items"
              :key="rule.id"
              class="rule-item"
              :class="{ 'is-active': selectedRuleId === rule.id }"
              @click="selectedRuleId = rule.id"
            >
              <span class="rule-name">{{ rule.name }}</span>
              <el-switch
                v-permission="'odin:reason:save'"
                v-model="rule.enabled"
                :active-value="1"
                :inactive-value="0"
                size="small"
                @change="(val: number) => handleToggle(rule, val)"
                @click.stop
              />
              <div class="rule-actions">
                <el-icon v-permission="'odin:reason:save'" class="action-icon" @click.stop="handleEditRule(rule)"><Edit /></el-icon>
                <el-icon v-permission="'odin:reason:save'" class="action-icon danger" @click.stop="handleDeleteRule(rule)"><Delete /></el-icon>
              </div>
            </div>
          </template>
          <el-empty v-if="!rulesLoading && rules.length === 0" :description="t('common.noData')" :image-size="40" />
        </div>
      </aside>

      <!-- 右侧三张卡片 -->
      <main class="lab-main">
        <!-- Card 1: 推理测试 -->
        <el-card class="lab-card" shadow="never">
          <template #header>
            <span class="card-title">{{ t('odinReason.reasonTest') }}</span>
          </template>
          <div class="card-body">
            <el-form :inline="true" size="small">
              <el-form-item :label="t('odinReason.targetClass')">
                <el-select v-model="testClassCode" filterable :placeholder="t('common.pleaseSelect')" style="width: 180px">
                  <el-option v-for="c in classOptions" :key="c.id" :label="c.name" :value="c.code" />
                </el-select>
              </el-form-item>
            </el-form>
            <!-- 动态条件 -->
            <div class="condition-list">
              <div v-for="(cond, i) in testConditions" :key="i" class="condition-row">
                <el-input v-model="cond.property" size="small" :placeholder="t('odinMapping.propertyCode')" style="width: 120px" />
                <el-select v-model="cond.op" size="small" style="width: 90px">
                  <el-option v-for="op in OPERATORS" :key="op" :label="op" :value="op" />
                </el-select>
                <el-input v-model="cond.value" size="small" placeholder="value" style="width: 120px" />
                <el-button link type="danger" size="small" @click="testConditions.splice(i, 1)">×</el-button>
              </div>
              <el-button size="small" @click="testConditions.push({ property: '', op: '=', value: '' })">
                + {{ t('common.add') }}
              </el-button>
            </div>
            <el-button type="primary" size="small" :loading="explainLoading" :disabled="!testClassCode" @click="handleExplain">
              {{ t('odinReason.simulate') }}
            </el-button>
            <ReasonTraceView v-if="traceResult" :trace="traceResult" class="trace-result" />
          </div>
        </el-card>

        <!-- Card 2: 规则试运行 -->
        <el-card class="lab-card" shadow="never">
          <template #header>
            <span class="card-title">{{ t('odinReason.dryRun') }}</span>
          </template>
          <div class="card-body">
            <div class="dryrun-info">
              <span v-if="selectedRule">{{ t('odinReason.ruleName') }}: <strong>{{ selectedRule.name }}</strong></span>
              <span v-else class="hint">{{ t('odinReason.selectRuleHint') }}</span>
            </div>
            <el-button
              type="primary"
              size="small"
              :disabled="!selectedRule"
              :loading="dryRunLoading"
              @click="handleDryRun"
            >
              {{ t('odinReason.explain') }}
            </el-button>
            <ReasonTraceView v-if="dryRunResult" :trace="dryRunResult" class="trace-result" />
          </div>
        </el-card>

        <!-- Card 3: 关系路径分析 -->
        <el-card class="lab-card" shadow="never">
          <template #header>
            <span class="card-title">{{ t('odinReason.pathAnalysis') }}</span>
          </template>
          <div class="card-body">
            <el-form :inline="true" size="small">
              <el-form-item :label="t('odinReason.sourceClass')">
                <el-select v-model="pathSource" filterable :placeholder="t('common.pleaseSelect')" style="width: 160px">
                  <el-option v-for="c in classOptions" :key="c.id" :label="c.name" :value="c.code" />
                </el-select>
              </el-form-item>
              <el-form-item :label="t('odinReason.targetClass')">
                <el-select v-model="pathTarget" filterable :placeholder="t('common.pleaseSelect')" style="width: 160px">
                  <el-option v-for="c in classOptions" :key="c.id" :label="c.name" :value="c.code" />
                </el-select>
              </el-form-item>
              <el-form-item :label="t('odinReason.maxHops')">
                <el-input-number v-model="maxHops" :min="1" :max="5" size="small" controls-position="right" style="width: 90px" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" size="small" :loading="pathLoading" :disabled="!pathSource || !pathTarget" @click="handlePathAnalysis">
                  {{ t('odinReason.analyzePath') }}
                </el-button>
              </el-form-item>
            </el-form>
            <!-- 路径可视化 -->
            <div v-if="pathResult" class="path-visual">
              <template v-for="(node, i) in pathResult.nodes" :key="node.classCode">
                <div class="path-node">
                  <span class="path-node-label">{{ node.className || node.classCode }}</span>
                  <span v-if="node.reason" class="path-node-reason">{{ node.reason }}</span>
                </div>
                <span v-if="i < pathResult.nodes.length - 1" class="path-arrow">→</span>
              </template>
              <div class="path-meta">
                {{ t('odinReason.hops') }}: {{ pathResult.hops }} |
                {{ pathResult.relations?.join(', ') || '—' }}
              </div>
            </div>
          </div>
        </el-card>
      </main>
    </div>

    <!-- 规则新增/编辑弹窗 -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="isEditRule ? t('odinReason.editRule') : t('odinReason.addRule')"
      width="600px"
      destroy-on-close
    >
      <el-form ref="ruleFormRef" :model="ruleForm" :rules="ruleFormRules" label-width="110px">
        <el-form-item :label="t('odinReason.ruleName')" prop="name">
          <el-input v-model="ruleForm.name" :placeholder="t('common.pleaseInput')" :maxlength="128" />
        </el-form-item>
        <el-form-item :label="t('odinReason.ruleType')" prop="ruleType">
          <el-select v-model="ruleForm.ruleType" :placeholder="t('common.pleaseSelect')" style="width: 100%">
            <el-option label="Derivation" value="derivation" />
            <el-option label="Constraint" value="constraint" />
            <el-option label="Aggregation" value="aggregation" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('odinReason.sourceClass')" prop="expression">
          <el-input v-model="ruleForm.conditionJson" type="textarea" :rows="3" placeholder="Condition JSON / expression" />
        </el-form-item>
        <el-form-item :label="t('odinReason.action')">
          <el-input v-model="ruleForm.actionJson" type="textarea" :rows="3" placeholder="Action JSON / expression" />
        </el-form-item>
        <el-form-item :label="t('odinReason.priority')" prop="priority">
          <el-input-number v-model="ruleForm.priority" :min="0" :max="999" controls-position="right" />
        </el-form-item>
        <el-form-item :label="t('odinReason.description')" prop="description">
          <el-input v-model="ruleForm.description" type="textarea" :rows="2" :maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('odinReason.enabled')" prop="enabled">
          <el-switch v-model="ruleForm.enabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
        <el-button type="primary" :loading="ruleSubmitLoading" @click="handleRuleSubmit">{{ t('common.confirmBtn') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Edit, Delete } from '@element-plus/icons-vue'
import OntologySwitcher from '@/views/odin/components/OntologySwitcher.vue'
import ReasonTraceView from '@/views/odin/components/reason/ReasonTraceView.vue'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
import {
  getRulePage,
  createRule,
  updateRule,
  deleteRules,
  toggleRule,
  explainRule,
  analyzePath,
} from '@/api/odin/reason'
import { getOntClassTree } from '@/api/odin/ontology'
import type { ReasonRuleVO, ReasonRuleDTO, ReasonTraceResult, ReasonPath } from '@/types/odin/reason'
import type { OntClassVO } from '@/types/odin/ontology'

const { t } = useI18n()
const workspaceStore = useOdinWorkspaceStore()

const OPERATORS = ['=', '!=', '>', '<', '>=', '<=', 'contains', 'startsWith', 'in']

const activeOntologyId = ref(workspaceStore.activeOntologyId)
const rules = ref<ReasonRuleVO[]>([])
const rulesLoading = ref(false)
const selectedRuleId = ref('')
const classOptions = ref<OntClassVO[]>([])

// Stats
const enabledCount = computed(() => rules.value.filter((r) => r.enabled === 1).length)
const lastExecTime = ref('')
const successRate = ref(100)

const selectedRule = computed(() => rules.value.find((r) => r.id === selectedRuleId.value) ?? null)

interface RuleGroup {
  type: string
  items: ReasonRuleVO[]
}

const groupedRules = computed<RuleGroup[]>(() => {
  const map = new Map<string, ReasonRuleVO[]>()
  for (const r of rules.value) {
    const key = r.ruleType || 'other'
    if (!map.has(key)) map.set(key, [])
    map.get(key)!.push(r)
  }
  return [...map.entries()].map(([type, items]) => ({ type, items }))
})

// ===== Load rules =====
async function loadRules() {
  if (!activeOntologyId.value) return
  rulesLoading.value = true
  try {
    const res = await getRulePage({ current: 1, size: 500, ontologyId: activeOntologyId.value })
    rules.value = res.records
  } finally {
    rulesLoading.value = false
  }
}

async function loadClasses() {
  if (!activeOntologyId.value) return
  try {
    const tree = await getOntClassTree(activeOntologyId.value)
    classOptions.value = flattenClasses(tree)
  } catch { classOptions.value = [] }
}

function flattenClasses(tree: OntClassVO[]): OntClassVO[] {
  const out: OntClassVO[] = []
  for (const node of tree) {
    out.push(node)
    if (node.children?.length) out.push(...flattenClasses(node.children))
  }
  return out
}

// ===== Card 1: Explain =====
const testClassCode = ref('')
const testConditions = ref<Array<{ property: string; op: string; value: string }>>([])
const explainLoading = ref(false)
const traceResult = ref<ReasonTraceResult | null>(null)

async function handleExplain() {
  if (!activeOntologyId.value || !testClassCode.value) return
  explainLoading.value = true
  traceResult.value = null
  try {
    const filters = testConditions.value
      .filter((c) => c.property && c.op)
      .map((c) => ({ property: c.property, op: c.op, value: c.value || undefined }))
    traceResult.value = await explainRule({
      ontologyId: activeOntologyId.value,
      classCode: testClassCode.value,
      filters,
    })
    lastExecTime.value = new Date().toLocaleTimeString()
  } catch (e) {
    ElMessage.error((e as Error).message || 'Explain failed')
    successRate.value = Math.max(0, successRate.value - 10)
  } finally {
    explainLoading.value = false
  }
}

// ===== Card 2: Dry Run =====
const dryRunLoading = ref(false)
const dryRunResult = ref<ReasonTraceResult | null>(null)

async function handleDryRun() {
  if (!selectedRule.value || !activeOntologyId.value) return
  dryRunLoading.value = true
  dryRunResult.value = null
  try {
    dryRunResult.value = await explainRule({
      ontologyId: activeOntologyId.value,
      classCode: selectedRule.value.code || testClassCode.value || '',
    })
    lastExecTime.value = new Date().toLocaleTimeString()
  } catch (e) {
    ElMessage.error((e as Error).message || 'Dry run failed')
  } finally {
    dryRunLoading.value = false
  }
}

// ===== Card 3: Path =====
const pathSource = ref('')
const pathTarget = ref('')
const maxHops = ref(3)
const pathLoading = ref(false)
const pathResult = ref<ReasonPath | null>(null)

async function handlePathAnalysis() {
  if (!activeOntologyId.value || !pathSource.value || !pathTarget.value) return
  pathLoading.value = true
  pathResult.value = null
  try {
    pathResult.value = await analyzePath({
      ontologyId: activeOntologyId.value,
      sourceClass: pathSource.value,
      targetClass: pathTarget.value,
    })
    lastExecTime.value = new Date().toLocaleTimeString()
  } catch (e) {
    ElMessage.error((e as Error).message || 'Path analysis failed')
  } finally {
    pathLoading.value = false
  }
}

// ===== Rule CRUD =====
const ruleDialogVisible = ref(false)
const isEditRule = ref(false)
const ruleSubmitLoading = ref(false)
const ruleFormRef = ref<FormInstance>()
const ruleForm = reactive<ReasonRuleDTO>({
  id: undefined,
  ontologyId: '',
  name: '',
  ruleType: 'derivation',
  conditionJson: '',
  actionJson: '',
  priority: 100,
  description: '',
  enabled: 1,
})

const ruleFormRules = computed<FormRules>(() => ({
  name: [{ required: true, message: t('common.pleaseInput') + t('odinReason.ruleName'), trigger: 'blur' }],
  ruleType: [{ required: true, message: t('common.pleaseSelect') + t('odinReason.ruleType'), trigger: 'change' }],
}))

function handleAddRule() {
  isEditRule.value = false
  Object.assign(ruleForm, {
    id: undefined,
    ontologyId: activeOntologyId.value,
    name: '',
    ruleType: 'derivation',
    conditionJson: '',
    actionJson: '',
    priority: 100,
    description: '',
    enabled: 1,
  })
  ruleDialogVisible.value = true
}

function handleEditRule(rule: ReasonRuleVO) {
  isEditRule.value = true
  Object.assign(ruleForm, {
    id: rule.id,
    ontologyId: rule.ontologyId,
    name: rule.name,
    ruleType: rule.ruleType,
    conditionJson: rule.conditionJson ?? '',
    actionJson: rule.actionJson ?? '',
    priority: rule.priority ?? 100,
    description: rule.description ?? '',
    enabled: rule.enabled,
  })
  ruleDialogVisible.value = true
}

async function handleRuleSubmit() {
  const valid = await ruleFormRef.value?.validate().catch(() => false)
  if (!valid) return
  ruleSubmitLoading.value = true
  try {
    ruleForm.ontologyId = activeOntologyId.value
    if (isEditRule.value) {
      await updateRule(ruleForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createRule(ruleForm)
      ElMessage.success(t('common.createSuccess'))
    }
    ruleDialogVisible.value = false
    await loadRules()
  } finally {
    ruleSubmitLoading.value = false
  }
}

async function handleDeleteRule(rule: ReasonRuleVO) {
  await ElMessageBox.confirm(t('odinReason.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteRules([rule.id])
  ElMessage.success(t('common.deleteSuccess'))
  await loadRules()
}

async function handleToggle(rule: ReasonRuleVO, val: number) {
  try {
    await toggleRule(rule.id, val)
    ElMessage.success(val === 1 ? t('odinReason.enabled') : t('odinReason.disabled'))
  } catch {
    rule.enabled = val === 1 ? 0 : 1
  }
}

// ===== Watchers =====
watch(activeOntologyId, (id) => {
  if (!id) return
  workspaceStore.setActive(id)
  selectedRuleId.value = ''
  traceResult.value = null
  dryRunResult.value = null
  pathResult.value = null
  loadRules()
  loadClasses()
})

onMounted(async () => {
  if (!workspaceStore.ontologies.length) {
    await workspaceStore.loadOntologies()
  }
  activeOntologyId.value = workspaceStore.activeOntologyId
  await Promise.all([loadRules(), loadClasses()])
})
</script>

<style scoped>
.reason-lab {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.lab-header {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 12px 20px;
  background: var(--sm-bg-card);
  border-bottom: 1px solid var(--sm-border-light);
}

.header-stats {
  display: flex;
  gap: 24px;
  margin-left: auto;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 18px;
  font-weight: 700;
  color: var(--sm-text-primary);
}

.stat-enabled {
  color: var(--el-color-success);
}

.stat-label {
  font-size: 11px;
  color: var(--sm-text-secondary);
}

.lab-body {
  flex: 1;
  display: grid;
  grid-template-columns: 320px 1fr;
  overflow: hidden;
}

.lab-aside {
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--sm-border-light);
  background: var(--sm-bg-card);
  overflow: hidden;
}

.aside-header {
  padding: 12px;
  border-bottom: 1px solid var(--sm-border-lighter);
}

.aside-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.group-label {
  padding: 8px 12px 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--sm-text-secondary);
}

.rule-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background 0.15s;
}

.rule-item:hover {
  background: var(--sm-bg-page);
}

.rule-item.is-active {
  background: color-mix(in srgb, var(--sm-color-primary) 8%, transparent);
  border-right: 3px solid var(--sm-color-primary);
}

.rule-name {
  flex: 1;
  font-size: var(--sm-font-size-sm, 13px);
  color: var(--sm-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rule-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.15s;
}

.rule-item:hover .rule-actions {
  opacity: 1;
}

.action-icon {
  font-size: 14px;
  cursor: pointer;
  color: var(--sm-text-secondary);
}

.action-icon:hover {
  color: var(--sm-color-primary);
}

.action-icon.danger:hover {
  color: var(--el-color-danger);
}

.lab-main {
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.lab-card {
  border-radius: var(--sm-radius-md, 8px);
}

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--sm-text-primary);
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.condition-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.condition-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.trace-result {
  margin-top: 12px;
}

.dryrun-info {
  font-size: var(--sm-font-size-sm, 13px);
  color: var(--sm-text-primary);
}

.dryrun-info .hint {
  color: var(--sm-text-secondary);
}

/* Path visualization */
.path-visual {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  padding: 12px 0;
}

.path-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 14px;
  background: linear-gradient(135deg, color-mix(in srgb, var(--sm-color-primary) 12%, var(--sm-bg-card)), color-mix(in srgb, var(--sm-color-primary) 5%, var(--sm-bg-card)));
  border: 1px solid color-mix(in srgb, var(--sm-color-primary) 25%, transparent);
  border-radius: 8px;
}

.path-node-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--sm-text-primary);
}

.path-node-reason {
  font-size: 11px;
  color: var(--sm-text-secondary);
  margin-top: 2px;
}

.path-arrow {
  font-size: 18px;
  color: var(--sm-color-primary);
  font-weight: 700;
  margin: 0 4px;
}

.path-meta {
  width: 100%;
  margin-top: 8px;
  font-size: 12px;
  color: var(--sm-text-secondary);
}
</style>
