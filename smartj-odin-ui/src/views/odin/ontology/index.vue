<script setup lang="ts">
import { ref, computed, markRaw, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { VueFlow, useVueFlow, type Node, type Edge } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Connection,
  Refresh,
  FolderOpened,
  Upload,
  ZoomIn,
  ZoomOut,
  FullScreen,
  Search,
  SetUp,
} from '@element-plus/icons-vue'

import OntologySwitcher from '@/views/odin/components/OntologySwitcher.vue'
import ClassNode from '@/views/odin/components/flow/ClassNode.vue'
import RelationEdge from '@/views/odin/components/flow/RelationEdge.vue'
import ClassDrawer from '@/views/odin/components/ontology/ClassDrawer.vue'
import RuleDrawer from '@/views/odin/components/ontology/RuleDrawer.vue'
import ClassFormDialog from '@/views/odin/components/ontology/ClassFormDialog.vue'
import PropertyFormDialog from '@/views/odin/components/ontology/PropertyFormDialog.vue'
import RelationFormDialog from '@/views/odin/components/ontology/RelationFormDialog.vue'
import RuleFormDialog from '@/views/odin/components/ontology/RuleFormDialog.vue'
import ImportExportDialog from '@/views/odin/components/ontology/ImportExportDialog.vue'

import { useOdinOntologyStore } from '@/stores/odin/ontology'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
import { useOntologyLayout } from '@/composables/useOntologyLayout'
import { createOntProperty, updateOntProperty, createOntRule, updateOntRule, deleteOntRule } from '@/api/odin/ontology'
import type { OntClassVO, OntClassDTO, OntPropertyVO, OntPropertyDTO, OntRelationVO, OntRelationDTO, OntRuleVO, OntRuleDTO } from '@/types/odin'
import type { ClassNodeData } from '@/views/odin/components/flow/ClassNode.vue'

import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

const { t } = useI18n()
const route = useRoute()
const ontStore = useOdinOntologyStore()
const wsStore = useOdinWorkspaceStore()
const { fitView, zoomIn, zoomOut } = useVueFlow()

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const nodeTypes = { classNode: markRaw(ClassNode) } as any
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const edgeTypes = { relationEdge: markRaw(RelationEdge) } as any

// ===== 左侧面板 =====
const leftTab = ref<'graph' | 'history'>('graph')
const searchKeyword = ref('')

/** 搜索过滤（左侧类列表用） */
const filteredClasses = computed(() => {
  if (!searchKeyword.value) return ontStore.classes
  const kw = searchKeyword.value.toLowerCase()
  return ontStore.classes.filter(
    (c) => c.name.toLowerCase().includes(kw) || c.code.toLowerCase().includes(kw),
  )
})

// ===== 画布过滤器（图例即过滤器）=====
const showMapped = ref(true)    // 已映射类（有属性）
const showUnmapped = ref(true)  // 未映射类（无属性）
const showVirtual = ref(true)   // 虚拟/规则类
const focusMode = ref(false)    // 聚焦模式：仅显示聚焦类及其邻域
const layoutDirection = ref<'TB' | 'LR'>('LR')

/** graph nodes 携带的属性名（string[]），用于填充节点属性预览 */
const graphPropMap = computed(() => {
  const m = new Map<string, string[]>()
  ontStore.graph?.nodes?.forEach((n) => {
    if (n.properties?.length) m.set(n.id, n.properties)
  })
  return m
})

/** graph nodes 属性明细：类节点ID → 属性编码 → 名称/类型，用于节点内显示「名称(编码)」 */
const graphPropDetailMap = computed(() => {
  const m = new Map<string, Map<string, { name?: string; dataType?: string }>>()
  ontStore.graph?.nodes?.forEach((n) => {
    if (!n.propertyDetails?.length) return
    const inner = new Map<string, { name?: string; dataType?: string }>()
    n.propertyDetails.forEach((pd) => inner.set(pd.code, { name: pd.name, dataType: pd.dataType }))
    m.set(n.id, inner)
  })
  return m
})

/** 类的邻接关系（聚焦模式用） */
const neighborMap = computed(() => {
  const m = new Map<string, Set<string>>()
  const add = (a: string, b: string) => {
    if (!m.has(a)) m.set(a, new Set())
    m.get(a)!.add(b)
  }
  ontStore.relations.forEach((r) => {
    add(r.sourceClassId, r.targetClassId)
    add(r.targetClassId, r.sourceClassId)
  })
  return m
})

/** 画布可视节点模型（类 + 规则虚拟节点） */
type VisualKind = 'class' | 'rule'
interface VisualNode {
  id: string
  name: string
  code: string
  description?: string
  kind: VisualKind
  isVirtual: boolean
  isUnmapped: boolean
  properties: string[]
}

const visualNodes = computed<VisualNode[]>(() => {
  const kw = searchKeyword.value.trim().toLowerCase()
  const focused = ontStore.focusedClassId
  const neighbors = focused ? (neighborMap.value.get(focused) ?? new Set<string>()) : new Set<string>()
  const result: VisualNode[] = []

  for (const c of ontStore.classes) {
    if (kw && !c.name.toLowerCase().includes(kw) && !c.code.toLowerCase().includes(kw)) continue
    const isUnmapped = (c.propertyCount ?? 0) === 0
    if (isUnmapped && !showUnmapped.value) continue
    if (!isUnmapped && !showMapped.value) continue
    if (focusMode.value && focused && c.id !== focused && !neighbors.has(c.id)) continue
    result.push({
      id: c.id,
      name: c.name,
      code: c.code,
      description: c.description,
      kind: 'class',
      isVirtual: false,
      isUnmapped,
      properties: graphPropMap.value.get(c.id) ?? [],
    })
  }

  // 规则作为虚拟节点进入画布（聚焦模式下隐藏，规则不属于类邻域）
  if (showVirtual.value && !(focusMode.value && focused)) {
    for (const r of ontStore.rules) {
      if (kw && !r.name.toLowerCase().includes(kw) && !r.code.toLowerCase().includes(kw)) continue
      result.push({
        id: `rule:${r.id}`,
        name: r.name,
        code: r.code,
        description: r.description,
        kind: 'rule',
        isVirtual: true,
        isUnmapped: false,
        properties: [],
      })
    }
  }

  return result
})

const visibleIds = computed(() => new Set(visualNodes.value.map((v) => v.id)))

// ===== 画布数据 =====
const layoutDirty = ref(false)

const nodes = computed<Node<ClassNodeData>[]>(() => {
  return visualNodes.value.map((v) => ({
    id: v.id,
    type: 'classNode',
    position: ontStore.layoutData.get(v.id) ?? { x: 0, y: 0 },
    data: {
      id: v.id,
      name: v.name,
      code: v.code,
      description: v.description,
      domain: undefined,
      isVirtual: v.isVirtual,
      isUnmapped: v.isUnmapped,
      isFocused: ontStore.focusedClassId === v.id,
      direction: layoutDirection.value,
      properties: v.properties.map((p, i) => {
        const d = graphPropDetailMap.value.get(v.id)?.get(p)
        return { id: `${v.id}-p${i}`, code: p, name: d?.name || p, dataType: d?.dataType || '' }
      }),
      mappedSources: [],
    },
    draggable: true,
    connectable: false,
  }))
})

const edges = computed<Edge[]>(() => {
  return ontStore.relations
    .filter((r) => visibleIds.value.has(r.sourceClassId) && visibleIds.value.has(r.targetClassId))
    .map((r) => ({
      id: r.id,
      type: 'relationEdge',
      source: r.sourceClassId,
      target: r.targetClassId,
      label: `${r.name} (${r.cardinality})`,
      data: { isCrossSource: false, cardinality: r.cardinality },
    }))
})

// ===== 布局 =====
const { layout: computeLayout } = useOntologyLayout({
  nodeSpacingX: 260,
  layerSpacingY: 180,
  nodeWidth: 220,
  nodeHeight: 140,
  collisionDetection: true,
})

/** 自动布局；markDirty=false 用于加载/切换方向等自动触发，不提示"未保存" */
function runAutoLayout(markDirty = true) {
  if (!visualNodes.value.length) {
    if (markDirty) ElMessage.info(t('odinOntology.noClasses'))
    return
  }
  const layoutNodes = visualNodes.value.map((v) => ({ id: v.id, label: v.name }))
  const layoutEdges = ontStore.relations
    .filter((r) => visibleIds.value.has(r.sourceClassId) && visibleIds.value.has(r.targetClassId))
    .map((r) => ({ source: r.sourceClassId, target: r.targetClassId }))
  const result = computeLayout(layoutNodes, layoutEdges, layoutDirection.value)
  ontStore.layoutData = new Map(result.positions)
  if (markDirty) layoutDirty.value = true
  fitViewLater(400)
}

async function saveLayoutNow() {
  const ontId = wsStore.activeOntologyId
  if (!ontId) return
  await ontStore.saveLayout(ontId)
  layoutDirty.value = false
  ElMessage.success(t('odinOntology.layoutSaved'))
}

function fitViewLater(duration?: number) {
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      fitView(duration ? { padding: 0.12, duration } : { padding: 0.12 })
    })
  })
}

// ===== 节点交互 =====
const drawerVisible = ref(false)
const drawerClassId = ref('')
const ruleDrawerVisible = ref(false)
const drawerRuleId = ref('')

function onNodeClick(payload: { node: Node }) {
  // 规则虚拟节点：单独打开规则详情抽屉
  if (payload.node.id.startsWith('rule:')) {
    drawerRuleId.value = payload.node.id
    ruleDrawerVisible.value = true
    return
  }
  drawerClassId.value = payload.node.id
  drawerVisible.value = true
  ontStore.focusClass(payload.node.id)
}

function onNodeDblClick(payload: { node: Node }) {
  if (payload.node.id.startsWith('rule:')) return
  ontStore.focusClass(payload.node.id)
  fitViewLater(300)
}

function onPaneClick() {
  ontStore.focusClass('')
}

function onNodeDragStop(payload: { node: Node }) {
  ontStore.layoutData.set(payload.node.id, {
    x: payload.node.position.x,
    y: payload.node.position.y,
  })
  layoutDirty.value = true
}

// ===== 弹窗控制 =====
const classDlgVisible = ref(false)
const classDlgData = ref<OntClassVO | null>(null)
const propDlgVisible = ref(false)
const propDlgClassId = ref('')
const propDlgData = ref<OntPropertyVO | null>(null)
const relDlgVisible = ref(false)
const relDlgData = ref<OntRelationVO | null>(null)
const ruleDlgVisible = ref(false)
const ruleDlgData = ref<OntRuleVO | null>(null)
const importExportVisible = ref(false)
const savingLayout = ref(false)

function openNewClass() {
  classDlgData.value = null
  classDlgVisible.value = true
}

function openEditClass(cls: OntClassVO) {
  classDlgData.value = cls
  classDlgVisible.value = true
}

function openNewRelation() {
  relDlgData.value = null
  relDlgVisible.value = true
}

function openNewRule() {
  ruleDlgData.value = null
  ruleDlgVisible.value = true
}

function openEditRule(rule: OntRuleVO) {
  ruleDlgData.value = rule
  ruleDlgVisible.value = true
  ruleDrawerVisible.value = false
}

function openAddProperty(classId: string) {
  propDlgClassId.value = classId
  propDlgData.value = null
  propDlgVisible.value = true
}

function openEditProperty(prop: OntPropertyVO) {
  propDlgClassId.value = prop.classId
  propDlgData.value = prop
  propDlgVisible.value = true
}

// ===== 弹窗回调 =====
async function handleClassSave(data: OntClassDTO) {
  if (data.id) {
    await ontStore.updateClass(data)
    ElMessage.success(t('common.updateSuccess'))
  } else {
    await ontStore.addClass(data)
    ElMessage.success(t('common.createSuccess'))
  }
  refreshGraph()
}

async function handlePropertySave(data: OntPropertyDTO) {
  if (data.id) {
    await updateOntProperty(data)
    ElMessage.success(t('common.updateSuccess'))
  } else {
    await createOntProperty(data)
    ElMessage.success(t('common.createSuccess'))
  }
  refreshGraph()
}

async function handleRelationSave(data: OntRelationDTO) {
  if (data.id) {
    await ontStore.updateRelation(data)
    ElMessage.success(t('common.updateSuccess'))
  } else {
    await ontStore.addRelation(data)
    ElMessage.success(t('common.createSuccess'))
  }
  refreshGraph()
}

async function handleRuleSave(data: OntRuleDTO) {
  if (data.id) {
    await updateOntRule(data)
    ElMessage.success(t('common.updateSuccess'))
  } else {
    await createOntRule(data)
    ElMessage.success(t('common.createSuccess'))
  }
  refreshGraph()
}

async function handleRuleDelete(rule: OntRuleVO) {
  try {
    await ElMessageBox.confirm(
      t('odinOntology.confirmDeleteRule', { name: rule.name }),
      t('common.tip'),
      { type: 'warning' },
    )
  } catch {
    return
  }
  await deleteOntRule(rule.id)
  ElMessage.success(t('common.deleteSuccess'))
  ruleDrawerVisible.value = false
  refreshGraph()
}

// ===== 变更历史 =====
const expandedChange = ref<string>('')

function toggleChange(id: string) {
  expandedChange.value = expandedChange.value === id ? '' : id
}

function actionLabel(action: string): string {
  const map: Record<string, string> = {
    create: t('odinOntology.changeActionCreate'),
    update: t('odinOntology.changeActionUpdate'),
    delete: t('odinOntology.changeActionDelete'),
    publish: t('odinOntology.changeActionPublish'),
    revert: t('odinOntology.changeActionRevert'),
  }
  return map[action] ?? action
}

function actionTagType(action: string): 'success' | 'warning' | 'danger' | 'info' | 'primary' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'primary'> = {
    create: 'success',
    update: 'warning',
    delete: 'danger',
    publish: 'primary',
    revert: 'info',
  }
  return map[action] ?? 'info'
}

function formatJson(v: unknown): string {
  if (v == null || v === '') return '—'
  if (typeof v === 'string') {
    try { return JSON.stringify(JSON.parse(v), null, 2) } catch { return v }
  }
  try { return JSON.stringify(v, null, 2) } catch { return String(v) }
}

// ===== 加载与刷新 =====
async function refreshGraph() {
  const ontId = wsStore.activeOntologyId
  if (!ontId) return
  await ontStore.loadGraph(ontId)
}

async function loadData() {
  await wsStore.loadOntologies()
  const ontId = wsStore.activeOntologyId
  if (ontId) {
    await ontStore.loadGraph(ontId)
    await ontStore.loadChangeLog(ontId)
    runAutoLayout(false)
  }
}

watch(() => wsStore.activeOntologyId, (id) => {
  if (id) {
    ontStore.loadGraph(id).then(() => runAutoLayout(false))
    ontStore.loadChangeLog(id)
  }
})

// 切换布局方向 → 重新布局
watch(layoutDirection, () => runAutoLayout(false))

// 类/规则数量变化（加载或增删后）→ 重新布局
watch(() => [ontStore.classes.length, ontStore.rules.length], () => {
  if (ontStore.classes.length) runAutoLayout(false)
})

onMounted(async () => {
  await loadData()
  // 支持从映射看板深链聚焦类（?classId=xxx）
  const cid = route.query.classId
  if (typeof cid === 'string' && cid) {
    drawerClassId.value = cid
    drawerVisible.value = true
    ontStore.focusClass(cid)
  }
})
</script>

<template>
  <div class="ontology-workspace">
    <!-- 左侧面板 -->
    <aside class="left-panel">
      <el-tabs v-model="leftTab" class="panel-tabs">
        <!-- 图谱 Tab -->
        <el-tab-pane :label="t('odinOntology.graphTab')" name="graph">
          <div class="panel-content">
            <!-- 搜索 -->
            <el-input
              v-model="searchKeyword"
              :placeholder="t('odinOntology.searchPlaceholder')"
              clearable
              size="default"
              :prefix-icon="Search"
              class="panel-search"
            />

            <!-- 统计摘要 -->
            <div class="stats-grid">
              <div class="stat-item">
                <span class="stat-value">{{ ontStore.stats.classCount }}</span>
                <span class="stat-label">{{ t('odinOntology.statsClassCount') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ ontStore.stats.relationCount }}</span>
                <span class="stat-label">{{ t('odinOntology.statsRelationCount') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ ontStore.stats.propertyCount }}</span>
                <span class="stat-label">{{ t('odinOntology.statsPropertyCount') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ ontStore.stats.coverageRate }}<i>%</i></span>
                <span class="stat-label">{{ t('odinOntology.statsCoverageRate') }}</span>
              </div>
            </div>

            <!-- 领域分组 -->
            <div class="section-title">
              <span class="section-bar" />{{ t('odinOntology.domainGroup') }}
              <em class="section-count">{{ filteredClasses.length }}</em>
            </div>
            <div class="class-list">
              <div
                v-for="c in filteredClasses"
                :key="c.id"
                class="class-item"
                :class="{ active: ontStore.focusedClassId === c.id }"
                @click="drawerClassId = c.id; drawerVisible = true; ontStore.focusClass(c.id)"
              >
                <span
                  class="class-dot"
                  :class="(c.propertyCount ?? 0) === 0 ? 'dot-unmapped' : 'dot-mapped'"
                />
                <span class="class-text">
                  <span class="class-name">{{ c.name }}</span>
                  <span class="class-code">{{ c.code }}</span>
                </span>
                <span class="class-badge" :title="t('odinOntology.statsPropertyCount')">
                  {{ c.propertyCount ?? 0 }}
                </span>
              </div>
              <el-empty v-if="!filteredClasses.length" :image-size="40" :description="t('odinOntology.noClasses')" />
            </div>
          </div>
        </el-tab-pane>

        <!-- 变更历史 Tab -->
        <el-tab-pane :label="t('odinOntology.changeHistoryTab')" name="history">
          <div class="panel-content">
            <el-timeline v-if="ontStore.changeLog.length">
              <el-timeline-item
                v-for="log in ontStore.changeLog"
                :key="log.id"
                :timestamp="log.createTime"
                placement="top"
                size="small"
              >
                <div class="change-item" @click="toggleChange(log.id)">
                  <el-tag size="small" :type="actionTagType(log.action)">
                    {{ actionLabel(log.action) }}
                  </el-tag>
                  <span class="change-entity">{{ log.entityType }}: {{ log.entityId }}</span>
                </div>
                <div v-if="expandedChange === log.id" class="change-detail">
                  <div v-if="log.operator" class="change-meta">
                    {{ t('odinOntology.changeOperator') }}: {{ log.operatorName || log.operator }}
                  </div>
                  <div v-if="log.reason" class="change-meta">
                    {{ t('odinOntology.changeReason') }}: {{ log.reason }}
                  </div>
                  <div class="diff-grid">
                    <div>
                      <div class="diff-label">{{ t('odinOntology.changeBefore') }}</div>
                      <pre class="diff-block">{{ formatJson(log.beforeSnapshot) }}</pre>
                    </div>
                    <div>
                      <div class="diff-label">{{ t('odinOntology.changeAfter') }}</div>
                      <pre class="diff-block">{{ formatJson(log.afterSnapshot) }}</pre>
                    </div>
                  </div>
                </div>
              </el-timeline-item>
            </el-timeline>
            <el-empty v-else :description="t('odinOntology.changeNoData')" :image-size="48" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </aside>

    <!-- 右侧画布 -->
    <main class="canvas-area">
      <!-- 工具条 -->
      <div class="canvas-toolbar">
        <OntologySwitcher />
        <el-divider direction="vertical" />
        <el-button v-permission="'odin:ontology:create'" size="small" type="primary" :icon="Plus" @click="openNewClass">
          {{ t('odinOntology.toolbarNewClass') }}
        </el-button>
        <el-button v-permission="'odin:ontology:create'" size="small" :icon="Connection" @click="openNewRelation">
          {{ t('odinOntology.toolbarNewRelation') }}
        </el-button>
        <el-button v-permission="'odin:ontology:create'" size="small" :icon="SetUp" @click="openNewRule">
          {{ t('odinOntology.toolbarNewRule') }}
        </el-button>
        <el-divider direction="vertical" />
        <el-radio-group v-model="layoutDirection" size="small">
          <el-radio-button value="LR">{{ t('odinOntology.layoutHorizontal') }}</el-radio-button>
          <el-radio-button value="TB">{{ t('odinOntology.layoutVertical') }}</el-radio-button>
        </el-radio-group>
        <el-button size="small" :icon="Refresh" @click="runAutoLayout(true)">
          {{ t('odinOntology.toolbarAutoLayout') }}
        </el-button>
        <el-button
          v-permission="'odin:ontology:update'"
          size="small"
          :icon="FolderOpened"
          :loading="savingLayout"
          :disabled="!layoutDirty"
          @click="saveLayoutNow"
        >
          {{ t('odinOntology.toolbarSaveLayout') }}
        </el-button>
        <el-tag v-if="layoutDirty" size="small" type="warning" effect="plain">
          {{ t('odinOntology.layoutDirty') }}
        </el-tag>
        <el-divider direction="vertical" />
        <el-button v-permission="['odin:ontology:create', 'odin:ontology:query']" size="small" :icon="Upload" @click="importExportVisible = true">
          {{ t('odinOntology.toolbarImportExport') }}
        </el-button>
        <el-divider direction="vertical" />
        <el-button size="small" :icon="FullScreen" @click="fitView({ padding: 0.12 })" />
        <el-button size="small" :icon="ZoomIn" @click="zoomIn()" />
        <el-button size="small" :icon="ZoomOut" @click="zoomOut()" />
      </div>

      <!-- Vue Flow 画布 -->
      <VueFlow
        v-loading="ontStore.loading"
        :nodes="nodes"
        :edges="edges"
        :node-types="nodeTypes"
        :edge-types="edgeTypes"
        :min-zoom="0.15"
        :max-zoom="2"
        fit-view-on-init
        @node-click="onNodeClick"
        @node-double-click="onNodeDblClick"
        @pane-click="onPaneClick"
        @node-drag-stop="onNodeDragStop"
      >
        <Background :gap="20" />
        <Controls />
        <MiniMap pannable zoomable />
      </VueFlow>

      <!-- 图例（即过滤器：勾选/取消可显隐对应节点） -->
      <div class="legend">
        <span class="legend-hint">{{ t('odinOntology.legendFilterHint') }}</span>
        <el-checkbox v-model="showMapped" size="small">
          <i class="lg lg-normal" />{{ t('odinOntology.legendNormal') }}
        </el-checkbox>
        <el-checkbox v-model="showUnmapped" size="small">
          <i class="lg lg-unmapped" />{{ t('odinOntology.legendUnmapped') }}
        </el-checkbox>
        <el-checkbox v-model="showVirtual" size="small">
          <i class="lg lg-virtual" />{{ t('odinOntology.legendVirtual') }}
        </el-checkbox>
        <el-checkbox v-model="focusMode" size="small" :disabled="!ontStore.focusedClassId">
          <i class="lg lg-focused" />{{ t('odinOntology.legendFocused') }}
        </el-checkbox>
      </div>
    </main>

    <!-- 抽屉与弹窗 -->
    <ClassDrawer
      v-model:visible="drawerVisible"
      :class-id="drawerClassId"
      @edit-class="openEditClass"
      @add-property="openAddProperty"
      @edit-property="openEditProperty"
      @add-relation="openNewRelation"
      @refresh="refreshGraph"
    />
    <RuleDrawer
      v-model:visible="ruleDrawerVisible"
      :node-id="drawerRuleId"
      @edit="openEditRule"
      @delete="handleRuleDelete"
    />
    <ClassFormDialog
      v-model:visible="classDlgVisible"
      :ontology-id="wsStore.activeOntologyId"
      :class-data="classDlgData"
      @save="handleClassSave"
    />
    <PropertyFormDialog
      v-model:visible="propDlgVisible"
      :class-id="propDlgClassId"
      :property-data="propDlgData"
      @save="handlePropertySave"
    />
    <RelationFormDialog
      v-model:visible="relDlgVisible"
      :ontology-id="wsStore.activeOntologyId"
      :relation-data="relDlgData"
      @save="handleRelationSave"
    />
    <RuleFormDialog
      v-model:visible="ruleDlgVisible"
      :ontology-id="wsStore.activeOntologyId"
      :rule-data="ruleDlgData"
      @save="handleRuleSave"
    />
    <ImportExportDialog
      v-model:visible="importExportVisible"
      :ontology-id="wsStore.activeOntologyId"
      @imported="refreshGraph"
    />
  </div>
</template>

<style scoped>
.ontology-workspace {
  display: grid;
  grid-template-columns: 268px 1fr;
  height: 100%;
  min-height: 0;
  background: var(--sm-bg-page);
}

/* ===== 左侧面板 ===== */
.left-panel {
  background: var(--sm-bg-card);
  border-right: 1px solid var(--sm-border-light);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 14px;
}

.panel-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background: var(--sm-border-lighter);
}

.panel-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
}

.panel-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.panel-content {
  height: 100%;
  overflow-y: auto;
  padding: 14px;
}

.panel-search {
  margin-bottom: 14px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  color: var(--sm-text-regular);
  letter-spacing: 0.02em;
  margin: 4px 0 10px;
}

.section-bar {
  display: inline-block;
  width: 3px;
  height: 12px;
  border-radius: 2px;
  background: var(--sm-color-primary);
}

.section-count {
  margin-left: auto;
  font-style: normal;
  font-size: 11px;
  font-weight: 500;
  color: var(--sm-text-placeholder);
  background: var(--sm-bg-page);
  border-radius: 999px;
  padding: 0 7px;
  line-height: 18px;
}

.class-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.class-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: var(--sm-radius-md);
  border: 1px solid transparent;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, transform 0.1s;
  font-size: 12px;
}

.class-item:hover {
  background: var(--sm-bg-hover);
  border-color: var(--sm-border-lighter);
}

.class-item.active {
  background: color-mix(in srgb, var(--sm-color-primary) 10%, transparent);
  border-color: color-mix(in srgb, var(--sm-color-primary) 35%, transparent);
}

.class-item.active .class-name {
  color: var(--sm-color-primary);
}

.class-dot {
  flex-shrink: 0;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-mapped {
  background: var(--sm-color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--sm-color-primary) 15%, transparent);
}

.dot-unmapped {
  background: var(--sm-text-placeholder);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--sm-text-placeholder) 15%, transparent);
}

.class-text {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
  flex: 1;
}

.class-name {
  font-weight: 500;
  color: var(--sm-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.class-code {
  font-size: 10px;
  color: var(--sm-text-placeholder);
  font-family: 'JetBrains Mono', Consolas, monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.class-badge {
  flex-shrink: 0;
  min-width: 20px;
  height: 18px;
  padding: 0 6px;
  border-radius: 999px;
  background: var(--sm-bg-page);
  border: 1px solid var(--sm-border-lighter);
  color: var(--sm-text-secondary);
  font-size: 10px;
  font-weight: 600;
  line-height: 16px;
  text-align: center;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 16px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 10px 6px;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--sm-color-primary) 5%, var(--sm-bg-page)),
    var(--sm-bg-page)
  );
  border-radius: var(--sm-radius-md);
  border: 1px solid var(--sm-border-lighter);
}

.stat-value {
  font-size: 18px;
  font-weight: 700;
  line-height: 1.1;
  color: var(--sm-color-primary);
}

.stat-value i {
  font-style: normal;
  font-size: 11px;
  font-weight: 600;
  margin-left: 1px;
}

.stat-label {
  font-size: 10px;
  color: var(--sm-text-secondary);
}

/* ===== 变更历史 ===== */
.change-item {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 12px;
}

.change-entity {
  color: var(--sm-text-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.change-detail {
  margin-top: 6px;
  padding: 6px;
  background: var(--sm-bg-page);
  border-radius: var(--sm-radius-md);
  font-size: 11px;
}

.change-meta {
  color: var(--sm-text-secondary);
  margin-bottom: 4px;
}

.diff-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  margin-top: 6px;
}

.diff-label {
  font-weight: 600;
  font-size: 10px;
  color: var(--sm-text-secondary);
  margin-bottom: 2px;
}

.diff-block {
  margin: 0;
  padding: 4px 6px;
  background: var(--sm-bg-card);
  border: 1px solid var(--sm-border-lighter);
  border-radius: 3px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 10px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 120px;
  overflow: auto;
}

/* ===== 右侧画布 ===== */
.canvas-area {
  position: relative;
  min-width: 0;
  min-height: 0;
}

.canvas-toolbar {
  position: absolute;
  top: 10px;
  left: 10px;
  right: 10px;
  z-index: 6;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  background: color-mix(in srgb, var(--sm-bg-card) 95%, transparent);
  border: 1px solid var(--sm-border-light);
  border-radius: var(--sm-radius-lg);
  padding: 8px 12px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.06);
  backdrop-filter: blur(6px);
}

/* 图例（过滤器） */
.legend {
  position: absolute;
  left: 10px;
  bottom: 10px;
  z-index: 5;
  display: flex;
  align-items: center;
  gap: 14px;
  background: color-mix(in srgb, var(--sm-bg-card) 94%, transparent);
  border: 1px solid var(--sm-border-light);
  border-radius: var(--sm-radius-lg);
  padding: 6px 14px;
  font-size: 11px;
  color: var(--sm-text-secondary);
  backdrop-filter: blur(4px);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.06);
}

.legend-hint {
  font-size: 11px;
  color: var(--sm-text-placeholder);
  padding-right: 4px;
  border-right: 1px solid var(--sm-border-lighter);
}

.legend :deep(.el-checkbox) {
  height: 22px;
  margin-right: 0;
}

.legend :deep(.el-checkbox__label) {
  display: inline-flex;
  align-items: center;
  font-size: 11px;
  color: var(--sm-text-secondary);
  padding-left: 6px;
}

.legend .lg {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 3px;
  margin-right: 4px;
  vertical-align: -2px;
  border: 1.5px solid var(--sm-text-secondary);
}

.lg-normal {
  background: var(--sm-bg-card);
  border-color: var(--sm-color-primary);
}

.lg-unmapped {
  background: var(--sm-bg-page);
  border-style: dashed;
  border-color: var(--sm-border-base);
}

.lg-virtual {
  background: color-mix(in srgb, var(--sm-color-success) 8%, var(--sm-bg-card));
  border-style: dotted;
  border-color: var(--sm-color-success);
}

.lg-focused {
  background: color-mix(in srgb, var(--sm-color-warning) 10%, var(--sm-bg-card));
  border-color: var(--sm-color-warning);
}

/* Vue Flow 选中态 */
:deep(.vue-flow__node.selected .class-node) {
  border-color: var(--sm-color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--sm-color-primary) 20%, transparent);
}
</style>
