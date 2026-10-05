<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  ElCard,
  ElButton,
  ElTag,
  ElInput,
  ElSelect,
  ElOption,
  ElEmpty,
  ElDivider,
  ElAlert,
  ElIcon,
  ElTable,
  ElTableColumn,
  ElDialog,
  ElForm,
  ElFormItem,
  ElSwitch,
  ElPopconfirm,
  ElMessage,
} from 'element-plus'
import {
  MagicStick,
  View,
  Plus,
  EditPen,
  Delete,
  Aim,
} from '@element-plus/icons-vue'
import ReasonTraceView from '@/components/reason/ReasonTraceView.vue'
import { reasonExplain } from '@/api/reason'
import { useOntologyStore } from '@/stores/ontology'
import { useQueryStore } from '@/stores/query'
import { useTraceStore } from '@/stores/trace'
import { FILTER_OPS } from '@/types/query'
import type { FilterOp, ReasonExplainVO } from '@/types/query'
import type { OntoRule } from '@/types/ontology'

const onto = useOntologyStore()
const qs = useQueryStore()
/** P3：全链路溯源抽屉由 trace store 驱动（TraceDrawer 挂在 App.vue） */
const traceStore = useTraceStore()

/** 规则展示形状（由 ontologyStore.rules 派生，不再硬编码） */
interface RuleView {
  id: string
  numericId?: number
  name: string
  description: string
  condition: string
  action: string
  targetClass: string
  enabled: boolean
}

const rules = computed<RuleView[]>(() =>
  onto.rules.map((r) => ({
    id: r.id,
    numericId: r.numericId,
    name: r.label || r.id,
    description: r.description || '—',
    condition: r.condition || '—',
    action: r.action || '—',
    targetClass: r.classId,
    enabled: r.enabled !== false,
  })),
)

const classOptions = computed(() =>
  onto.classes.map((c) => ({ id: c.id, label: c.label, numericId: c.numericId })),
)

function classLabel(classId: string) {
  return onto.classes.find((c) => c.id === classId)?.label || classId || '—'
}

// 规则统计
const ruleStats = computed(() => {
  const total = rules.value.length
  const enabled = rules.value.filter((r) => r.enabled).length
  return { total, enabled, disabled: total - enabled }
})

// --- 规则 CRUD：全部走 ontologyStore → 后端 API ---
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const ruleForm = reactive({
  name: '',
  label: '',
  classId: '',
  description: '',
  condition: '',
  action: '',
  enabled: true,
})

function openCreate() {
  editingId.value = null
  Object.assign(ruleForm, {
    name: '',
    label: '',
    classId: classOptions.value[0]?.id ?? '',
    description: '',
    condition: '',
    action: '',
    enabled: true,
  })
  dialogVisible.value = true
}

function openEdit(rule: RuleView) {
  const raw = onto.rules.find((r) => r.id === rule.id)
  if (!raw) return
  editingId.value = raw.numericId ?? null
  Object.assign(ruleForm, {
    name: raw.id,
    label: raw.label || '',
    classId: raw.classId,
    description: raw.description || '',
    condition: raw.condition || '',
    action: raw.action || '',
    enabled: raw.enabled !== false,
  })
  dialogVisible.value = true
}

/** OntoRule / 表单 → 后端 OntRuleVO 请求体 */
function rulePayload(src: {
  name: string
  label: string
  classId: string
  description: string
  condition: string
  action: string
  enabled: boolean
}) {
  return {
    name: src.name,
    label: src.label || src.name,
    class_id: onto.classes.find((c) => c.id === src.classId)?.numericId,
    description: src.description,
    condition_json: src.condition,
    action_json: src.action,
    enabled: src.enabled,
  }
}

async function submitRule() {
  if (onto.currentOntologyId == null) {
    ElMessage.warning('请先选择本体')
    return
  }
  const name = (ruleForm.name || ruleForm.label).trim()
  if (!name) {
    ElMessage.warning('请填写规则名')
    return
  }
  saving.value = true
  try {
    const payload = rulePayload({ ...ruleForm, name })
    if (editingId.value != null) {
      await onto.updateRuleApi(onto.currentOntologyId, editingId.value, payload)
      ElMessage.success('规则已更新')
    } else {
      await onto.createRule(onto.currentOntologyId, payload)
      ElMessage.success('规则已创建')
    }
    dialogVisible.value = false
  } catch {
    ElMessage.error('规则保存失败')
  } finally {
    saving.value = false
  }
}

async function toggleRule(rule: RuleView, enabled: boolean | string | number) {
  const next = enabled === true || enabled === 'true'
  const raw = onto.rules.find((r) => r.id === rule.id)
  if (onto.currentOntologyId == null || rule.numericId == null || !raw) return
  try {
    await onto.updateRuleApi(
      onto.currentOntologyId,
      rule.numericId,
      rulePayload({
        name: raw.id,
        label: raw.label || '',
        classId: raw.classId,
        description: raw.description || '',
        condition: raw.condition || '',
        action: raw.action || '',
        enabled: next,
      }),
    )
  } catch {
    ElMessage.error('规则状态更新失败')
  }
}

async function removeRule(rule: RuleView) {
  if (onto.currentOntologyId == null || rule.numericId == null) return
  try {
    await onto.deleteRuleApi(onto.currentOntologyId, rule.numericId)
    ElMessage.success('规则已删除')
  } catch {
    ElMessage.error('规则删除失败')
  }
}

// --- 推理测试：走 queryStore.dryRun（后端生成物理 SQL） ---
const testQuery = ref('')
const testResult = ref<{
  expandedRules: string[]
  inferredClasses: string[]
  sqlPreview: string
  notes: string[]
} | null>(null)
const testing = ref(false)

const nlExample = computed(() => qs.nlExamples[0] ?? '查询本体类')

/** 从规则动作中解析出被推断的类名 */
function inferTargetClass(r: OntoRule): string {
  const m = /assert\s+([A-Za-z_][\w.]*)/i.exec(r.action ?? '')
  const name = m?.[1]
  const byName = name
    ? onto.classes.find((c) => c.id === name || c.label === name)
    : undefined
  return byName?.label || name || classLabel(r.classId)
}

const runReasonTest = async () => {
  if (!testQuery.value.trim()) return
  testing.value = true
  try {
    // P2：服务端解析入口 parseNaturalLanguage 已改为 async；ReasonLab 属 P3 范围，此处沿用同步本地解析
    const parsed = qs.parseNaturalLanguageLocal(testQuery.value)
    const query = { ...parsed.query, useReasoning: true }
    // 规则为 ontology 级：推理测试展示当前本体全部启用规则
    const hit = onto.rules.filter((r) => r.enabled !== false)
    const res = await qs.dryRunQuery({
      ontology_query: query,
      ontology_id: onto.currentOntologyId ?? undefined,
    })
    const sql = res?.sql || qs.generatedSQL || ''
    testResult.value = {
      expandedRules: hit.map((r) => r.id),
      inferredClasses: [...new Set(hit.map(inferTargetClass))],
      sqlPreview: sql || '未能生成物理 SQL，请检查该类的映射配置',
      notes: [...parsed.notes],
    }
  } finally {
    testing.value = false
  }
}

// --- 关系路径分析：基于 ontologyStore.relations 的真实图遍历 ---
const selectedClass = ref<string | null>(null)
const pathResult = ref<{
  path: string[]
  relations: Array<{ from: string; to: string; label: string }>
} | null>(null)

const analyzePath = (classId: string) => {
  selectedClass.value = classId
  pathResult.value = null
  if (!classId || !onto.classes.some((c) => c.id === classId)) return
  const visited = new Set<string>([classId])
  const path: string[] = [classId]
  const edges: Array<{ from: string; to: string; label: string }> = []
  let cur = classId
  for (let hop = 0; hop < 3; hop++) {
    const rel = onto.relations.find(
      (r) =>
        (r.domain === cur && !visited.has(r.range)) ||
        (r.range === cur && !visited.has(r.domain)),
    )
    if (!rel) break
    const next = rel.domain === cur ? rel.range : rel.domain
    edges.push({ from: cur, to: next, label: rel.label || rel.id })
    path.push(next)
    visited.add(next)
    cur = next
  }
  pathResult.value = { path, relations: edges }
}

// --- P3 规则试运行：POST /reason/explain（只推理不翻译、不执行 SQL） ---
interface ExplainFilter {
  property: string
  op: FilterOp
  value: string
}

const explainClass = ref<string>('')
const explainFilters = reactive<ExplainFilter[]>([])
const explainProps = ref<Array<{ name: string; label: string }>>([])
const explainResult = ref<ReasonExplainVO | null>(null)
const explaining = ref(false)
const explainError = ref('')

/** 注意：后端把推理轨迹嵌在 reason_trace 字段内，不是平铺 */
const explainTrace = computed(() => explainResult.value?.reason_trace ?? null)
const explainTargets = computed(() => explainResult.value?.expanded_targets ?? [])
const explainRules = computed(() => explainTrace.value?.rules_fired ?? [])
const explainFired = computed(() => explainRules.value.filter((r) => r.condition_result).length)
const explainHasResult = computed(() => !!explainResult.value)

/** ont_class.class_type：后端默认 normal，虚拟类为 virtual */
const CLASS_TYPE_LABELS: Record<string, string> = {
  normal: '普通类',
  physical: '实体类',
  virtual: '虚拟类',
  abstract: '抽象类',
}

const explainClassType = computed(() => {
  const t = String(explainResult.value?.class_type ?? '')
  return t ? CLASS_TYPE_LABELS[t] ?? t : ''
})

function addExplainFilter() {
  explainFilters.push({ property: '', op: 'eq', value: '' })
}

function removeExplainFilter(index: number) {
  explainFilters.splice(index, 1)
}

/** 切类：清空上次结果，并按需拉取该类的属性名作为过滤条件下拉项 */
async function onExplainClassChange(classId: string) {
  explainResult.value = null
  explainError.value = ''
  explainProps.value = []
  const cls = onto.classes.find((c) => c.id === classId)
  if (!cls) return
  if (cls.properties?.length) {
    explainProps.value = cls.properties.map((p) => ({ name: p.id, label: p.label || p.id }))
    return
  }
  if (onto.currentOntologyId == null || cls.numericId == null) return
  const list = await onto.fetchProperties(onto.currentOntologyId, cls.numericId)
  explainProps.value = (list ?? [])
    .map((p: any) => ({
      name: String(p?.name ?? p?.id ?? ''),
      label: String(p?.label ?? p?.name ?? p?.id ?? ''),
    }))
    .filter((p: { name: string }) => !!p.name)
}

async function runExplain() {
  if (onto.currentOntologyId == null) {
    ElMessage.warning('请先在顶部选择本体')
    return
  }
  if (!explainClass.value) {
    ElMessage.warning('请选择要试运行的本体类')
    return
  }
  explaining.value = true
  explainError.value = ''
  try {
    const filters = explainFilters
      .filter((f) => f.property.trim())
      .map((f) => {
        const v = f.value.trim()
        return { property: f.property.trim(), op: f.op, value: v || undefined }
      })
    explainResult.value = await reasonExplain({
      ontology_id: onto.currentOntologyId,
      class_name: explainClass.value,
      filters,
    })
    const fired = explainResult.value?.reason_trace?.rules_fired?.filter((r) => r.condition_result).length ?? 0
    ElMessage.success(`推理完成：命中 ${fired} 条规则，展开 ${explainTargets.value.length} 个目标类`)
  } catch (e: any) {
    explainResult.value = null
    explainError.value = e?.message || '规则试运行失败'
  } finally {
    explaining.value = false
  }
}

function resetExplain() {
  explainResult.value = null
  explainError.value = ''
  explainFilters.splice(0, explainFilters.length)
}

/** 打开 P3 全链路溯源抽屉（reason/explain 已落库 query_trace，可按 trace_id 查看） */
function openTrace(traceId?: string) {
  const id = (traceId || explainResult.value?.trace_id || '').trim()
  if (!id) {
    ElMessage.warning('本次试运行未返回 trace_id，无法查看全链路')
    return
  }
  void traceStore.open(id)
}

onMounted(async () => {
  if (onto.currentOntologyId != null && !onto.rules.length) {
    await onto.fetchRules(onto.currentOntologyId)
  }
})

watch(
  () => onto.currentOntologyId,
  async (id) => {
    testResult.value = null
    pathResult.value = null
    selectedClass.value = null
    resetExplain()
    explainClass.value = ''
    explainProps.value = []
    if (id != null) await onto.fetchRules(id)
  },
)
</script>

<template>
  <div class="reason-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>推理验证</h1>
        <p class="subtitle">测试和验证本体推理规则</p>
      </div>
      <div class="header-stats">
        <div class="stat-item">
          <span class="stat-value">{{ ruleStats.total }}</span>
          <span class="stat-label">规则总数</span>
        </div>
        <div class="stat-item">
          <span class="stat-value enabled">{{ ruleStats.enabled }}</span>
          <span class="stat-label">已启用</span>
        </div>
      </div>
    </div>

    <div class="content-area">
      <!-- 左侧：规则列表 -->
      <div class="rules-panel">
        <div class="panel-title">
          <h3>推理规则</h3>
          <ElButton type="primary" size="small" :icon="Plus" @click="openCreate">
            新建规则
          </ElButton>
        </div>
        <div v-loading="onto.loading" class="rules-list">
          <ElEmpty
            v-if="!rules.length"
            description="暂无推理规则"
            :image-size="80"
          />
          <div
            v-for="rule in rules"
            :key="rule.id"
            class="rule-card"
            :class="{ disabled: !rule.enabled }"
          >
            <div class="rule-header">
              <span class="rule-name">{{ rule.name }}</span>
              <div class="rule-header-right">
                <ElSwitch
                  :model-value="rule.enabled"
                  size="small"
                  @change="toggleRule(rule, $event)"
                />
                <ElTag
                  :type="rule.enabled ? 'success' : 'info'"
                  size="small"
                  effect="plain"
                >
                  {{ rule.enabled ? '启用' : '禁用' }}
                </ElTag>
                <ElButton
                  link
                  type="primary"
                  size="small"
                  :icon="EditPen"
                  @click="openEdit(rule)"
                />
                <ElPopconfirm
                  title="确认删除该推理规则？"
                  @confirm="removeRule(rule)"
                >
                  <template #reference>
                    <ElButton link type="danger" size="small" :icon="Delete" />
                  </template>
                </ElPopconfirm>
              </div>
            </div>
            <p class="rule-desc">{{ rule.description }}</p>
            <div class="rule-condition">
              <code>{{ rule.condition }}</code>
            </div>
            <div class="rule-action">
              <span class="label">动作：</span>
              <code>{{ rule.action }}</code>
            </div>
            <div class="rule-action rule-target">
              <span class="label">目标类：</span>
              <ElTag size="small" effect="plain" type="primary">
                {{ classLabel(rule.targetClass) }}
              </ElTag>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：测试区域 -->
      <div class="test-panel">
        <!-- 推理测试 -->
        <ElCard class="test-card">
          <template #header>
            <div class="card-header">
              <ElIcon><MagicStick /></ElIcon>
              <span>推理测试</span>
            </div>
          </template>

          <div class="test-input">
            <ElInput
              v-model="testQuery"
              :placeholder="'输入查询条件，如：' + nlExample"
              type="textarea"
              :rows="3"
            />
            <ElButton
              type="primary"
              :loading="testing"
              @click="runReasonTest"
              :disabled="!testQuery"
            >
              测试推理
            </ElButton>
          </div>

          <div v-if="testResult" class="test-result">
            <ElDivider>推理结果</ElDivider>

            <div class="result-section">
              <h4>触发的规则</h4>
              <div class="rules-tags">
                <ElTag
                  v-for="ruleId in testResult.expandedRules"
                  :key="ruleId"
                  type="success"
                  effect="plain"
                >
                  {{ rules.find((r) => r.id === ruleId)?.name || ruleId }}
                </ElTag>
                <span v-if="testResult.expandedRules.length === 0" class="no-result">
                  未触发任何规则
                </span>
              </div>
            </div>

            <div class="result-section">
              <h4>推断的类</h4>
              <div class="rules-tags">
                <ElTag
                  v-for="cls in testResult.inferredClasses"
                  :key="cls"
                  type="primary"
                  effect="plain"
                >
                  {{ cls }}
                </ElTag>
                <span v-if="testResult.inferredClasses.length === 0" class="no-result">
                  无推断类
                </span>
              </div>
            </div>

            <div class="result-section">
              <h4>SQL 预览</h4>
              <pre class="sql-preview"><code>{{ testResult.sqlPreview }}</code></pre>
            </div>

            <div v-if="testResult.notes.length" class="result-section">
              <h4>解析说明</h4>
              <ElAlert type="info" :closable="false">
                <div v-for="(note, i) in testResult.notes" :key="i" class="note-line">
                  {{ note }}
                </div>
              </ElAlert>
            </div>
          </div>
        </ElCard>

        <!-- P3 规则试运行 / 单步解释 -->
        <ElCard class="test-card">
          <template #header>
            <div class="card-header">
              <ElIcon><Aim /></ElIcon>
              <span>规则试运行 · 单步解释</span>
              <ElTag size="small" type="info" effect="plain">只推理不执行 SQL</ElTag>
              <div class="card-header-actions">
                <ElButton
                  size="small"
                  text
                  :icon="Delete"
                  :disabled="!explainFilters.length && !explainHasResult"
                  @click="resetExplain"
                >
                  清空
                </ElButton>
              </div>
            </div>
          </template>

          <div class="explain-form">
            <div class="explain-row">
              <span class="fld-label">本体类</span>
              <ElSelect
                v-model="explainClass"
                placeholder="选择要试运行的类"
                filterable
                style="flex: 1"
                @change="onExplainClassChange"
              >
                <ElOption
                  v-for="c in classOptions"
                  :key="c.id"
                  :label="c.label"
                  :value="c.id"
                />
              </ElSelect>
            </div>

            <div class="explain-filters">
              <div class="filters-head">
                <span class="fld-label">过滤条件</span>
                <ElButton size="small" text type="primary" :icon="Plus" @click="addExplainFilter">
                  添加条件
                </ElButton>
              </div>
              <div v-if="!explainFilters.length" class="no-result">
                无条件也可试运行（仅按类评估规则与虚拟类展开）
              </div>
              <div v-for="(f, i) in explainFilters" :key="i" class="filter-row">
                <ElSelect
                  v-model="f.property"
                  placeholder="属性"
                  filterable
                  allow-create
                  default-first-option
                  class="f-prop"
                >
                  <ElOption
                    v-for="p in explainProps"
                    :key="p.name"
                    :label="p.label === p.name ? p.name : `${p.label}（${p.name}）`"
                    :value="p.name"
                  />
                </ElSelect>
                <ElSelect v-model="f.op" class="f-op">
                  <ElOption v-for="o in FILTER_OPS" :key="o.value" :label="o.label" :value="o.value" />
                </ElSelect>
                <ElInput
                  v-model="f.value"
                  class="f-val"
                  placeholder="值"
                  :disabled="!FILTER_OPS.find((o) => o.value === f.op)?.needValue"
                />
                <ElButton :icon="Delete" text type="danger" @click="removeExplainFilter(i)" />
              </div>
            </div>

            <div class="explain-actions">
              <ElButton
                type="primary"
                :loading="explaining"
                :disabled="!explainClass"
                @click="runExplain"
              >
                试运行推理
              </ElButton>
              <span class="explain-tip">返回 rules_fired / expansion / derived_filters / inferred_relations</span>
            </div>
          </div>

          <ElAlert
            v-if="explainError"
            type="error"
            :closable="false"
            :title="explainError"
            class="explain-err"
          />

          <div v-if="explainHasResult" class="test-result">
            <ElDivider>推理结果</ElDivider>

            <div class="explain-summary">
              <ElTag size="small" type="success" effect="dark">{{ explainResult?.class_name }}</ElTag>
              <ElTag v-if="explainClassType" size="small" effect="plain">{{ explainClassType }}</ElTag>
              <ElTag size="small" :type="explainFired ? 'success' : 'info'" effect="plain">
                规则命中 {{ explainFired }}/{{ explainRules.length }}
              </ElTag>
              <ElTag size="small" type="primary" effect="plain">
                展开目标 {{ explainTargets.length }}
              </ElTag>
              <ElButton
                v-if="explainResult?.trace_id"
                link
                type="primary"
                size="small"
                :icon="View"
                @click="openTrace(explainResult?.trace_id)"
              >
                查看全链路
              </ElButton>
              <code v-if="explainResult?.trace_id" class="explain-trace-id">
                trace_id：{{ explainResult?.trace_id }}
              </code>
            </div>

            <div class="result-section">
              <h4>推理轨迹（命中情况 · condition · action）</h4>
              <ReasonTraceView
                :trace="explainTrace"
                empty-text="该类未触发任何规则，也没有虚拟类展开"
              />
            </div>

            <div class="result-section">
              <h4>展开落地目标</h4>
              <ElTable v-if="explainTargets.length" :data="explainTargets" border size="small">
                <ElTableColumn prop="class_name" label="类名" min-width="120" />
                <ElTableColumn prop="source_table" label="物理表" min-width="120" />
                <ElTableColumn label="数据源" min-width="170">
                  <template #default="{ row }">
                    <span>{{ row.datasource_name || (row.datasource_id ? `#${row.datasource_id}` : '—') }}</span>
                    <ElTag v-if="row.datasource_type" size="small" effect="plain" class="ds-type">
                      {{ row.datasource_type }}
                    </ElTag>
                  </template>
                </ElTableColumn>
              </ElTable>
              <span v-else class="no-result">无展开目标（该类可能直接映射到物理表）</span>
            </div>
          </div>
        </ElCard>

        <!-- 路径分析 -->
        <ElCard class="test-card">
          <template #header>
            <div class="card-header">
              <ElIcon><View /></ElIcon>
              <span>关系路径分析</span>
            </div>
          </template>

          <div class="path-input">
            <ElSelect
              v-model="selectedClass"
              placeholder="选择一个类"
              @change="analyzePath"
              style="width: 100%"
            >
              <ElOption
                v-for="c in classOptions"
                :key="c.id"
                :label="c.label"
                :value="c.id"
              />
            </ElSelect>
          </div>

          <div v-if="pathResult" class="path-result">
            <div class="path-visual">
              <div
                v-for="(node, index) in pathResult.path"
                :key="node"
                class="path-node-wrapper"
              >
                <div class="path-node">{{ node }}</div>
                <div
                  v-if="index < pathResult.path.length - 1"
                  class="path-arrow"
                >
                  <span>{{ pathResult.relations[index]?.label }}</span>
                  →
                </div>
              </div>
            </div>

            <ElTable :data="pathResult.relations" border size="small" class="relations-table">
              <ElTableColumn prop="from" label="起点" width="120" />
              <ElTableColumn prop="label" label="关系" />
              <ElTableColumn prop="to" label="终点" width="120" />
            </ElTable>
          </div>
        </ElCard>
      </div>
    </div>

    <!-- 规则编辑对话框 -->
    <ElDialog
      v-model="dialogVisible"
      :title="editingId != null ? '编辑推理规则' : '新建推理规则'"
      width="560px"
      append-to-body
    >
      <ElForm label-width="90px">
        <ElFormItem label="规则名">
          <ElInput v-model="ruleForm.name" placeholder="英文标识，如 vip_customer" />
        </ElFormItem>
        <ElFormItem label="显示名">
          <ElInput v-model="ruleForm.label" placeholder="如：VIP 客户" />
        </ElFormItem>
        <ElFormItem label="所属类">
          <ElSelect v-model="ruleForm.classId" placeholder="选择本体类" style="width: 100%">
            <ElOption
              v-for="c in classOptions"
              :key="c.id"
              :label="c.label"
              :value="c.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="ruleForm.description" type="textarea" :rows="2" />
        </ElFormItem>
        <ElFormItem label="条件">
          <ElInput
            v-model="ruleForm.condition"
            type="textarea"
            :rows="2"
            placeholder='如 Customer.level = "A"'
          />
        </ElFormItem>
        <ElFormItem label="动作">
          <ElInput v-model="ruleForm.action" placeholder="如 assert VipCustomer" />
        </ElFormItem>
        <ElFormItem label="启用">
          <ElSwitch v-model="ruleForm.enabled" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="saving" @click="submitRule">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.reason-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #f0f7ff 0%, #e8f4fd 100%);
  overflow-y: auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 32px;
  background: linear-gradient(135deg, #2b6cb0 0%, #3182ce 100%);
  color: white;
}

.header-left h1 {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 600;
}

.subtitle {
  margin: 0;
  font-size: 14px;
  opacity: 0.85;
}

.header-stats {
  display: flex;
  gap: 24px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
}

.stat-value.enabled {
  color: #86efac;
}

.stat-label {
  font-size: 12px;
  opacity: 0.8;
  margin-top: 4px;
}

.content-area {
  flex: 1;
  display: flex;
  gap: 24px;
  padding: 24px 32px;
  min-height: 0;
}

.rules-panel {
  width: 320px;
  display: flex;
  flex-direction: column;
}

.rules-panel h3 {
  margin: 0 0 16px;
  font-size: 16px;
  font-weight: 600;
  color: #1a365d;
}

.panel-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.panel-title h3 {
  margin: 0;
}

.rule-header-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.rule-target {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
}

.note-line {
  font-size: 12px;
  line-height: 1.8;
}

.rules-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-card {
  padding: 16px;
  background: white;
  border-radius: 12px;
  border: 1px solid #e4e7ed;
  transition: all 0.2s;
}

.rule-card:hover {
  border-color: #3182ce;
  box-shadow: 0 4px 12px rgba(49, 130, 206, 0.1);
}

.rule-card.disabled {
  opacity: 0.6;
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.rule-name {
  font-weight: 600;
  color: #1a365d;
}

.rule-desc {
  font-size: 13px;
  color: #718096;
  margin: 0 0 10px;
}

.rule-condition {
  background: #f7fafc;
  padding: 8px;
  border-radius: 6px;
  margin-bottom: 8px;
}

.rule-condition code {
  font-size: 12px;
  color: #2d3748;
}

.rule-action {
  font-size: 12px;
  color: #718096;
}

.rule-action .label {
  font-weight: 500;
}

.rule-action code {
  background: #edf2f7;
  padding: 2px 6px;
  border-radius: 4px;
  color: #2d3748;
}

.test-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.test-card {
  border-radius: 12px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #1a365d;
}

.test-input {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.test-result {
  margin-top: 16px;
}

.result-section {
  margin-bottom: 16px;
}

.result-section h4 {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 600;
  color: #2d3748;
}

.rules-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.no-result {
  color: #a0aec0;
  font-size: 13px;
}

/* --- P3 规则试运行 --- */
.card-header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;
}

.explain-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.explain-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.fld-label {
  flex-shrink: 0;
  width: 62px;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
}

.explain-filters {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filters-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filters-head .el-button {
  margin-left: auto;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.f-prop {
  width: 200px;
  flex-shrink: 0;
}

.f-op {
  width: 110px;
  flex-shrink: 0;
}

.f-val {
  flex: 1;
  min-width: 100px;
}

.explain-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.explain-tip {
  font-size: 12px;
  color: #a0aec0;
}

.explain-err {
  margin-top: 12px;
}

.explain-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 14px;
}

.ds-type {
  margin-left: 6px;
}

.explain-trace-id {
  font-size: 11px;
  color: #a0aec0;
  font-family: 'JetBrains Mono', Consolas, monospace;
  word-break: break-all;
}

.sql-preview {
  background: #1a202c;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.6;
}

.path-input {
  margin-bottom: 16px;
}

.path-result {
  margin-top: 16px;
}

.path-visual {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  overflow-x: auto;
  padding: 10px 0;
}

.path-node-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.path-node {
  padding: 10px 16px;
  background: linear-gradient(135deg, #3182ce 0%, #2b6cb0 100%);
  color: white;
  border-radius: 8px;
  font-weight: 600;
  font-size: 14px;
  white-space: nowrap;
}

.path-arrow {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  color: #718096;
  font-size: 12px;
}

.path-arrow span {
  font-size: 10px;
  white-space: nowrap;
}

.relations-table {
  margin-top: 12px;
}
</style>
