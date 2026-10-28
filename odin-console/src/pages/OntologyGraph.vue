<script setup lang="ts">
import { computed, markRaw, onMounted, ref, watch } from 'vue'
import { VueFlow, useVueFlow, type Node, type Edge } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import { ElButton, ElTag, ElTooltip, ElDivider, ElMessage } from 'element-plus'
import ClassNode from '@/components/flow/ClassNode.vue'
import RelationEdge from '@/components/flow/RelationEdge.vue'
import ClassDrawer from '@/components/ontology/ClassDrawer.vue'
import ClassFormDialog from '@/components/ontology/ClassFormDialog.vue'
import RelationFormDialog from '@/components/ontology/RelationFormDialog.vue'
import ImportExportDialog from '@/components/ontology/ImportExportDialog.vue'
import MappingVersionDialog from '@/components/mapping/MappingVersionDialog.vue'
import { useOntologyStore } from '@/stores/ontology'
import type { LayoutPos } from '@/types/ontology'
import type { ClassNodeData } from '@/components/flow/ClassNode.vue'

import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

const store = useOntologyStore()
const { fitView } = useVueFlow()

const nodeTypes = {
  classNode: markRaw(ClassNode),
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
} as any
const edgeTypes = {
  relation: markRaw(RelationEdge),
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
} as any

const classDlg = ref(false)
const relDlg = ref(false)
const importExportDlg = ref(false)
const mappingVersionDlg = ref(false)
/** 「保存布局」按钮 loading */
const savingLayout = ref(false)

/** 仅由「类/过滤/布局」驱动，不依赖 focusId，避免点选时整图重挂导致抽屉打不开 */
const nodes = computed<Node<ClassNodeData>[]>(() => {
  return store.filteredClasses.map((c) => ({
    id: c.id,
    type: 'classNode',
    position: store.layout[c.id] ?? { x: 0, y: 0 },
    data: {
      id: c.id,
      label: c.label,
      description: c.description,
      domain: c.domain,
      mappedSources: c.mappedSources,
      isVirtual: !!c.abstract,
      props: c.properties.map((p) => ({
        id: p.id,
        label: p.label,
        isKey: p.isKey,
        sensitive: p.sensitive,
      })),
    },
    draggable: true,
    connectable: false,
  }))
})

const edges = computed<Edge[]>(() => {
  return store.visibleRelations.map((r) => ({
    id: r.id,
    type: 'relation',
    source: r.domain,
    target: r.range,
    label: `${r.label} ${r.cardinality}`,
    data: { crossSource: r.crossSource, cardinality: r.cardinality },
  }))
})

// ---- 纯 TS 布局（不依赖第三方布局库）----
/** 节点估算宽度 w=260 / 基准高度 h=160，水平间距 80 / 垂直间距 100 */
const NODE_W = 260
const NODE_H = 160
const GAP_X = 80
const GAP_Y = 100
const PITCH_X = NODE_W + GAP_X
const PITCH_Y = NODE_H + GAP_Y
/** 节点矩形实际宽度（ClassNode max-width）与碰撞留白 */
const NODE_BOX_W = 240
const MIN_GAP = 16
/** 补位搜索步长 */
const STEP = 40
const MARGIN = 40

interface NodeSize {
  w: number
  h: number
}

interface Placed {
  pos: LayoutPos
  size: NodeSize
}

/** 按节点内容估算渲染尺寸（描述最多两行、属性最多 5 条 + 溢出提示） */
function estimateNodeSize(c: {
  description?: string
  abstract?: boolean
  properties: unknown[]
  mappedSources: string[]
}): NodeSize {
  const propCount = c.properties?.length ?? 0
  const rows = propCount > 5 ? 6 : propCount > 0 ? propCount : c.abstract ? 0 : 1
  const srcCount = c.mappedSources?.length ?? 0
  const chips = srcCount > 0 ? srcCount : c.abstract ? 0 : 1
  const chipRows = chips ? Math.ceil(chips / 3) : 0
  const headerH = 42
  const descH = c.description ? 36 : 0
  const propsH = rows ? rows * 16 + (rows - 1) * 3 + 8 : 0
  const footerH = 7 + (chipRows ? chipRows * 16 + (chipRows - 1) * 4 : 0)
  return { w: NODE_BOX_W, h: 24 + headerH + descH + propsH + footerH }
}

/** 两个节点矩形（含最小留白）是否冲突 */
function boxesOverlap(a: LayoutPos, sa: NodeSize, b: LayoutPos, sb: NodeSize): boolean {
  return (
    Math.abs(a.x - b.x) < (sa.w + sb.w) / 2 + MIN_GAP &&
    Math.abs(a.y - b.y) < (sa.h + sb.h) / 2 + MIN_GAP
  )
}

/**
 * 分层布局：按关系 BFS 深度分层（左 → 右），同层垂直居中排布；
 * 无关系的孤立类统一放到下方网格，保证任意两节点矩形（含间距）不重叠。
 */
function computeLayeredPositions(): Record<string, LayoutPos> {
  const ids = store.classes.map((c) => c.id)
  const idSet = new Set(ids)
  const outEdges = new Map<string, string[]>()
  const adj = new Map<string, string[]>()
  const inDeg = new Map<string, number>()
  ids.forEach((id) => {
    outEdges.set(id, [])
    adj.set(id, [])
    inDeg.set(id, 0)
  })

  const rels = store.relations.filter(
    (r) => r.domain !== r.range && idSet.has(r.domain) && idSet.has(r.range),
  )
  for (const r of rels) {
    outEdges.get(r.domain)!.push(r.range)
    adj.get(r.domain)!.push(r.range)
    adj.get(r.range)!.push(r.domain)
    inDeg.set(r.range, (inDeg.get(r.range) ?? 0) + 1)
  }

  const isolated = ids.filter((id) => (adj.get(id)?.length ?? 0) === 0)
  const isolatedSet = new Set(isolated)
  const connected = ids.filter((id) => !isolatedSet.has(id))

  // BFS 分层：先按有向边从根（无入边）扩散，再对环 / 反向分量用无向边补种
  const depth = new Map<string, number>()
  const bfs = (seeds: string[], edges: Map<string, string[]>) => {
    const queue: string[] = []
    for (const s of seeds) {
      if (!depth.has(s)) {
        depth.set(s, 0)
        queue.push(s)
      }
    }
    while (queue.length) {
      const cur = queue.shift()!
      for (const next of edges.get(cur) ?? []) {
        if (!depth.has(next)) {
          depth.set(next, (depth.get(cur) ?? 0) + 1)
          queue.push(next)
        }
      }
    }
  }
  const roots = connected.filter((id) => (inDeg.get(id) ?? 0) === 0)
  bfs(roots.length ? roots : connected.slice(0, 1), outEdges)
  for (const id of connected) {
    if (!depth.has(id)) bfs([id], adj)
  }

  // 同层按邻居重心排序，减少连线交叉
  const layers = new Map<number, string[]>()
  for (const id of connected) {
    const d = depth.get(id) ?? 0
    if (!layers.has(d)) layers.set(d, [])
    layers.get(d)!.push(id)
  }
  const maxDepth = layers.size ? Math.max(...layers.keys()) : -1
  const neighbours = new Map<string, string[]>()
  connected.forEach((id) => neighbours.set(id, []))
  for (const r of rels) {
    if (isolatedSet.has(r.domain) || isolatedSet.has(r.range)) continue
    neighbours.get(r.range)?.push(r.domain)
    neighbours.get(r.domain)?.push(r.range)
  }
  const orderIdx = new Map<string, number>()
  for (let d = 0; d <= maxDepth; d++) {
    const layer = layers.get(d)
    if (!layer) continue
    if (d > 0) {
      const ranked = layer
        .map((id, i) => {
          const ps = (neighbours.get(id) ?? [])
            .map((p) => orderIdx.get(p))
            .filter((v): v is number => typeof v === 'number')
          const bary = ps.length ? ps.reduce((a, b) => a + b, 0) / ps.length : i
          return { id, bary, i }
        })
        .sort((a, b) => a.bary - b.bary || a.i - b.i)
      ranked.forEach((it, i) => {
        layer[i] = it.id
      })
    }
    layer.forEach((id, i) => orderIdx.set(id, i))
  }

  // 摆位：每层一列，列内垂直居中；行距按最高节点动态放大，确保不重叠
  const sizeOf = new Map<string, NodeSize>()
  for (const c of store.classes) sizeOf.set(c.id, estimateNodeSize(c))
  const maxH = ids.reduce((m, id) => Math.max(m, sizeOf.get(id)?.h ?? NODE_H), NODE_H)
  const rowPitch = Math.max(PITCH_Y, maxH + MIN_GAP)
  const positions: Record<string, LayoutPos> = {}
  let maxRows = 1
  for (let d = 0; d <= maxDepth; d++) maxRows = Math.max(maxRows, layers.get(d)?.length ?? 0)
  for (let d = 0; d <= maxDepth; d++) {
    const layer = layers.get(d) ?? []
    const span = (layer.length - 1) * rowPitch
    layer.forEach((id, i) => {
      positions[id] = {
        x: MARGIN + d * PITCH_X,
        y: MARGIN + Math.round(i * rowPitch - span / 2),
      }
    })
  }

  // 整体平移到正坐标区
  const placed = Object.values(positions)
  if (placed.length) {
    const shiftY = MARGIN - Math.min(...placed.map((p) => p.y))
    placed.forEach((p) => {
      p.y += shiftY
    })
  }

  // 孤立类：底部网格排布
  if (isolated.length) {
    const maxY = placed.length ? Math.max(...placed.map((p) => p.y)) : MARGIN - rowPitch
    const baseY = maxY + rowPitch
    const cols = Math.max(3, maxRows)
    isolated.forEach((id, i) => {
      positions[id] = {
        x: MARGIN + (i % cols) * PITCH_X,
        y: baseY + Math.floor(i / cols) * rowPitch,
      }
    })
  }
  return positions
}

/** 在期望位置附近按 STEP 螺旋搜索第一个空位，保证与已占位节点不重叠 */
function findFreeSlot(desired: LayoutPos, size: NodeSize, taken: Placed[]): LayoutPos {
  const fits = (p: LayoutPos) => !taken.some((t) => boxesOverlap(p, size, t.pos, t.size))
  if (fits(desired)) return { ...desired }
  for (let ring = 1; ring <= 80; ring++) {
    for (let dy = -ring; dy <= ring; dy++) {
      for (let dx = -ring; dx <= ring; dx++) {
        if (Math.max(Math.abs(dx), Math.abs(dy)) !== ring) continue
        const cand = { x: desired.x + dx * STEP, y: desired.y + dy * STEP }
        if (cand.x < 0 || cand.y < 0) continue
        if (fits(cand)) return cand
      }
    }
  }
  return { ...desired }
}

function fitViewLater(duration?: number) {
  // 双 rAF 确保 VueFlow 完成渲染后再 fitView
  requestAnimationFrame(() => {
    requestAnimationFrame(() =>
      fitView(duration ? { padding: 0.12, duration } : { padding: 0.12 }),
    )
  })
}

/**
 * 加载：layout_json / 类自带坐标 → 按坐标摆放（若彼此重叠则就近挪开）；
 * 仍缺坐标的类 → 自动布局补位。结果静默写回后端；写入失败则转为「未保存」。
 */
async function ensureLayout(): Promise<boolean> {
  const list = store.classes
  if (!list.length) return false

  const taken: Placed[] = []
  const patch: Record<string, LayoutPos> = {}
  let autoPos: Record<string, LayoutPos> | null = null
  let repaired = 0

  // 1) 已有坐标：保留设计意图，仅在实际重叠时就近挪开
  for (const c of list) {
    const p = store.layout[c.id]
    if (!p) continue
    const size = estimateNodeSize(c)
    const slot = findFreeSlot({ x: p.x, y: p.y }, size, taken)
    if (slot.x !== p.x || slot.y !== p.y) {
      patch[c.id] = slot
      repaired++
    }
    taken.push({ pos: slot, size })
  }

  // 2) 缺失坐标：按自动布局的期望位置补位
  const missing = list.filter((c) => !store.layout[c.id])
  for (const c of missing) {
    if (!autoPos) autoPos = computeLayeredPositions()
    const size = estimateNodeSize(c)
    const slot = findFreeSlot(autoPos[c.id] ?? { x: MARGIN, y: MARGIN }, size, taken)
    patch[c.id] = slot
    taken.push({ pos: slot, size })
  }

  if (!Object.keys(patch).length) return false
  store.applyLayout(patch, `自动布局：补位 ${missing.length} 个类、修复 ${repaired} 处重叠`, false)
  const ok = await store.saveLayout()
  if (!ok) store.markLayoutDirty(`布局坐标待保存（补位 ${missing.length}、修复 ${repaired}）`)
  return true
}

/** 「自动布局」按钮：整图重算坐标并应用，标记未保存 */
function runAutoLayout() {
  if (!store.classes.length) {
    ElMessage.info('暂无可布局的类')
    return
  }
  const positions = computeLayeredPositions()
  store.applyLayout(positions, `自动布局：重排 ${Object.keys(positions).length} 个类`)
  fitViewLater(400)
}

/** 「保存布局」按钮：立即将当前坐标写回后端 layout_json */
async function saveLayoutNow() {
  savingLayout.value = true
  try {
    const ok = await store.saveLayout()
    if (ok) ElMessage.success('布局坐标已保存')
    else ElMessage.error('布局保存失败，请稍后重试')
  } finally {
    savingLayout.value = false
  }
}

function onNodeClick(payload: { node: Node }) {
  store.focus(payload.node.id)
}

function onPaneClick() {
  store.focus(null)
}

/** 拖动结束：只更新该类坐标并标记未保存变更（不重排其他节点） */
function onNodeDragStop(payload: { node: Node }) {
  store.setLayoutPos(payload.node.id, payload.node.position.x, payload.node.position.y)
}

/**
 * 串行化布局任务：onMounted 与 watch 可能同时触发，
 * 排队执行可避免重复计算与重复写库，且后到的类仍能被补位。
 */
let layoutTask: Promise<boolean> | null = null

function scheduleEnsureLayout(): Promise<boolean> {
  const prev = layoutTask
  const run = async (): Promise<boolean> => {
    if (prev) await prev.catch(() => false)
    return ensureLayout()
  }
  layoutTask = run()
  return layoutTask
}

/** 类集合签名：首次加载 / 新建 / 删除 / 切换本体都会变化 */
const classSignature = computed(() => store.classes.map((c) => c.id).join('|'))

watch(classSignature, async (sig) => {
  if (!sig) return
  const changed = await scheduleEnsureLayout()
  if (changed) fitViewLater(300)
})

onMounted(async () => {
  const changed = await scheduleEnsureLayout()
  fitViewLater(changed ? 300 : undefined)
})
</script>

<template>
  <div class="graph-page">
    <aside class="side">
      <section class="panel">
        <h3>搜索</h3>
        <input
          class="search"
          type="search"
          placeholder="类 / 属性 / 别名…"
          :value="store.search"
          @input="store.setSearch(($event.target as HTMLInputElement).value)"
        />
      </section>

      <section class="panel">
        <h3>过滤</h3>
        <label class="check">
          <input
            type="checkbox"
            :checked="store.showUnmappedOnly"
            @change="store.toggleUnmapped()"
          />
          仅显示未映射
        </label>
      </section>

      <section class="panel">
        <h3>领域</h3>
        <div class="domain-list">
          <ElTag
            :type="store.activeDomain === null ? 'primary' : 'info'"
            effect="plain"
            style="cursor: pointer; margin: 2px"
            @click="store.setDomain(null)"
          >
            全部
          </ElTag>
          <ElTag
            v-for="d in store.domains"
            :key="d"
            :type="store.activeDomain === d ? 'primary' : 'info'"
            effect="plain"
            style="cursor: pointer; margin: 2px"
            @click="store.setDomain(store.activeDomain === d ? null : d)"
          >
            {{ d }}
          </ElTag>
        </div>
      </section>

      <section class="panel">
        <h3>数据源</h3>
        <ul class="source-list">
          <li v-for="s in store.sources" :key="s.id">
            <span class="dot" :class="s.status"></span>
            <span class="sid">{{ s.label || s.id }}</span>
            <span class="stype">{{ s.type }} · {{ s.tables ?? 0 }}表</span>
          </li>
        </ul>
      </section>

      <section class="panel">
        <h3>最近变更</h3>
        <ul v-if="store.changeLog.length" class="change-list">
          <li v-for="c in store.changeLog.slice(0, 8)" :key="c.id">
            {{ c.summary }}
          </li>
        </ul>
        <p v-else class="stat muted">暂无未保存变更</p>
      </section>

      <section class="panel">
        <h3>统计</h3>
        <p class="stat">类 {{ store.filteredClasses.length }} / {{ store.classes.length }}</p>
        <p class="stat">关系 {{ store.visibleRelations.length }}</p>
        <p class="stat">本体版本 v{{ store.version }}</p>
      </section>
    </aside>

    <main class="canvas">
      <!-- 画布左上角工具栏 -->
      <div class="canvas-toolbar">
        <ElButton size="small" type="primary" @click="classDlg = true">新建类</ElButton>
        <ElButton size="small" @click="relDlg = true">新建关系</ElButton>
        <ElButton size="small" @click="runAutoLayout">自动布局</ElButton>
        <ElButton
          size="small"
          type="primary"
          plain
          :loading="savingLayout"
          :disabled="!store.layoutDirty"
          @click="saveLayoutNow"
        >
          保存布局
        </ElButton>
        <ElButton size="small" @click="store.focus(null)">取消选中</ElButton>
        <ElTag v-if="store.layoutDirty" size="small" type="warning" effect="plain">
          布局未保存
        </ElTag>
        <ElDivider direction="vertical" />
        <ElButton size="small" @click="importExportDlg = true">导入导出</ElButton>
        <ElButton size="small" @click="mappingVersionDlg = true">版本管理</ElButton>
      </div>

      <VueFlow
        :nodes="nodes"
        :edges="edges"
        :node-types="nodeTypes"
        :edge-types="edgeTypes"
        :min-zoom="0.2"
        :max-zoom="1.8"
        fit-view-on-init
        @node-click="onNodeClick"
        @pane-click="onPaneClick"
        @node-drag-stop="onNodeDragStop"
      >
        <Background :gap="18" />
        <Controls />
        <MiniMap pannable zoomable />
      </VueFlow>

      <div class="legend">
        <span><i class="lg normal"></i>已映射类</span>
        <span><i class="lg unmapped"></i>未映射</span>
        <span><i class="lg virtual"></i>虚拟/规则类</span>
        <span><i class="lg cross"></i>跨源关系</span>
      </div>

      <ElTooltip
        content="点节点打开详情；拖动节点或「自动布局」会更新坐标，需「保存布局」写入后端"
        placement="top"
      >
        <div class="hint">V2 可编辑 · 坐标持久化到 layout_json</div>
      </ElTooltip>
    </main>

    <ClassDrawer />
    <ClassFormDialog v-model:open="classDlg" mode="create" />
    <RelationFormDialog v-model:open="relDlg" />
    <ImportExportDialog v-model:open="importExportDlg" />
    <MappingVersionDialog v-model:open="mappingVersionDlg" />
  </div>
</template>

<style scoped>
.graph-page {
  display: grid;
  grid-template-columns: 220px 1fr;
  height: 100%;
  min-height: 0;
  background: #f1f5f9;
}

.side {
  background: #fff;
  border-right: 1px solid #e2e8f0;
  padding: 12px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.panel h3 {
  margin: 0 0 8px;
  font-size: 12px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.search {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  padding: 6px 8px;
  font-size: 13px;
}

.check {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #374151;
  margin-bottom: 6px;
  cursor: pointer;
}

.domain-list {
  display: flex;
  flex-wrap: wrap;
}

.source-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.source-list li {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot.ok {
  background: #22c55e;
}
.dot.warn {
  background: #f59e0b;
}
.dot.down {
  background: #ef4444;
}

.sid {
  font-weight: 600;
  color: #111827;
}

.stype {
  color: #9ca3af;
}

.change-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.change-list li {
  font-size: 11px;
  color: #475569;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 4px;
  padding: 4px 6px;
}

.stat {
  margin: 0 0 4px;
  font-size: 12px;
  color: #475569;
}

.stat.muted {
  color: #9ca3af;
}

.canvas {
  position: relative;
  min-width: 0;
  min-height: 0;
}

.canvas-toolbar {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 6;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 8px;
  box-shadow: 0 2px 10px rgba(15, 23, 42, 0.06);
}

.legend {
  position: absolute;
  left: 12px;
  bottom: 12px;
  z-index: 5;
  display: flex;
  gap: 12px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 6px 10px;
  font-size: 11px;
  color: #475569;
}

.legend .lg {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 3px;
  margin-right: 4px;
  vertical-align: -2px;
  border: 1.5px solid #94a3b8;
}

.lg.normal {
  background: #fff;
}
.lg.unmapped {
  background: #f3f4f6;
  border-style: dashed;
}
.lg.virtual {
  background: #f8fafc;
  border-style: dotted;
}
.lg.cross {
  background: #fffbeb;
  border-color: #f59e0b;
  height: 0;
  border-width: 2px 0 0;
  border-radius: 0;
  width: 16px;
}

.hint {
  position: absolute;
  right: 12px;
  bottom: 12px;
  z-index: 5;
  font-size: 11px;
  color: #94a3b8;
  background: rgba(255, 255, 255, 0.85);
  padding: 4px 8px;
  border-radius: 6px;
}

/* Vue Flow 选中态：包装层 .selected → 内部 .class-node */
:deep(.vue-flow__node.selected .class-node) {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.2);
}
</style>
