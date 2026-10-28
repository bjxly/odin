<template>
  <div class="wms-page">
    <!-- 左侧控制面板 -->
    <div class="wms-panel">
      <div class="panel-header">
        <h3>岩心仓储 3D 管理</h3>
      </div>
      <div class="panel-body">
        <!-- 视角控制 -->
        <div class="section">
          <div class="section-title">视角控制</div>
          <div class="btn-group">
            <el-button
              size="small"
              @click="sceneRef?.resetView()"
            >
              自由
            </el-button>
            <el-button
              size="small"
              @click="sceneRef?.setTopView()"
            >
              俯视
            </el-button>
            <el-button
              size="small"
              @click="sceneRef?.setFrontView()"
            >
              正面
            </el-button>
            <el-button
              size="small"
              @click="sceneRef?.setSideView()"
            >
              侧面
            </el-button>
          </div>
        </div>

        <!-- 仓库配置 -->
        <div class="section">
          <div class="section-title">仓库配置</div>
          <div class="config-item">
            <label>货架层数</label>
            <el-input-number
              v-model="config.rackLevels"
              :min="2"
              :max="12"
              size="small"
              @change="onConfigChange"
            />
          </div>
          <div class="config-item">
            <label>层高 (m)</label>
            <el-slider
              v-model="config.levelHeight"
              :min="0.8"
              :max="2.0"
              :step="0.1"
              @change="onConfigChange"
            />
          </div>
          <div class="config-item">
            <label>通道宽度 (m)</label>
            <el-slider
              v-model="config.aisleWidth"
              :min="2.5"
              :max="5.0"
              :step="0.5"
              @change="onConfigChange"
            />
          </div>
        </div>

        <!-- 库存统计 -->
        <div class="section">
          <div class="section-title">库存统计</div>
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-value">{{ stats.totalSlots }}</span>
              <span class="stat-label">总库位</span>
            </div>
            <div class="stat-item">
              <span class="stat-value stat-occupied">{{ stats.occupiedSlots }}</span>
              <span class="stat-label">已占用</span>
            </div>
            <div class="stat-item">
              <span class="stat-value stat-empty">{{ stats.emptySlots }}</span>
              <span class="stat-label">空闲</span>
            </div>
            <div class="stat-item">
              <span class="stat-value stat-boxes">{{ stats.totalBoxes }}</span>
              <span class="stat-label">岩心盒</span>
            </div>
          </div>
          <div class="utilization-bar">
            <div
              class="utilization-fill"
              :style="{ width: stats.utilization + '%' }"
            ></div>
          </div>
          <div class="utilization-text">利用率 {{ stats.utilization }}%</div>
        </div>

        <!-- 选项卡：库存 / 操作 -->
        <div class="section">
          <el-tabs
            v-model="activeTab"
            stretch
          >
            <el-tab-pane
              label="库存列表"
              name="inventory"
            >
              <div class="inventory-list">
                <div
                  v-for="item in occupiedSlots"
                  :key="item.palletId"
                  class="inventory-item"
                >
                  <span class="inv-rack">{{ item.rackId }}</span>
                  <span class="inv-pos">{{ item.level + 1 }}层{{ item.column + 1 }}列</span>
                  <span class="inv-spec">{{ item.boxSpec }}</span>
                  <span class="inv-count">×{{ item.boxCount }}</span>
                </div>
                <div
                  v-if="occupiedSlots.length === 0"
                  class="empty-text"
                >
                  暂无库存
                </div>
              </div>
            </el-tab-pane>

            <el-tab-pane
              label="入库操作"
              name="putaway"
            >
              <div class="operation-form">
                <div class="config-item">
                  <label>岩心盒规格</label>
                  <el-select
                    v-model="putAwayForm.boxSpec"
                    size="small"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="spec in boxSpecs"
                      :key="spec.name"
                      :label="spec.name"
                      :value="spec.name"
                    />
                  </el-select>
                </div>
                <div class="config-item">
                  <label>堆叠数量</label>
                  <el-input-number
                    v-model="putAwayForm.boxCount"
                    :min="1"
                    :max="10"
                    size="small"
                  />
                </div>
                <div class="config-item">
                  <label>目标货位</label>
                  <el-select
                    v-model="putAwayForm.slotKey"
                    size="small"
                    style="width: 100%"
                    placeholder="选择空货位"
                  >
                    <el-option
                      v-for="slot in availableSlots"
                      :key="slot.rackId + '-' + slot.level + '-' + slot.column"
                      :label="slot.label"
                      :value="slot.rackId + '-' + slot.level + '-' + slot.column"
                    />
                  </el-select>
                </div>
                <el-button
                  type="primary"
                  size="small"
                  :loading="animating"
                  :disabled="!putAwayForm.slotKey"
                  style="width: 100%"
                  @click="handlePutAway"
                >
                  执行入库
                </el-button>
              </div>
            </el-tab-pane>

            <el-tab-pane
              label="出库操作"
              name="pick"
            >
              <div class="operation-form">
                <div class="config-item">
                  <label>选择货位</label>
                  <el-select
                    v-model="pickForm.slotKey"
                    size="small"
                    style="width: 100%"
                    placeholder="选择已占货位"
                  >
                    <el-option
                      v-for="item in occupiedSlots"
                      :key="item.palletId"
                      :label="`${item.rackId} - ${item.level + 1}层${item.column + 1}列 (${item.boxSpec}×${item.boxCount})`"
                      :value="item.rackId + '-' + item.level + '-' + item.column"
                    />
                  </el-select>
                </div>
                <el-button
                  type="warning"
                  size="small"
                  :loading="animating"
                  :disabled="!pickForm.slotKey"
                  style="width: 100%"
                  @click="handlePick"
                >
                  执行出库
                </el-button>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>

        <!-- 选中详情 -->
        <div
          v-if="selection"
          class="section"
        >
          <div class="section-title">选中详情</div>
          <div class="selection-detail">
            <div>
              货架:
              <strong>{{ selection.rackId }}</strong>
            </div>
            <div>分类: {{ selection.categoryLabel }}</div>
            <div v-if="selection.level !== undefined">位置: {{ selection.level! + 1 }}层 {{ selection.column! + 1 }}列</div>
            <div v-if="selection.boxSpec">规格: {{ selection.boxSpec }}</div>
            <div v-if="selection.boxCount">盒数: {{ selection.boxCount }}</div>
            <div v-if="selection.projectNo">项目编号: {{ selection.projectNo }}</div>
            <div v-if="selection.holeNo">钻孔编号: {{ selection.holeNo }}</div>
            <div v-if="selection.depthFrom !== undefined">深度: {{ selection.depthFrom }}m ~ {{ selection.depthTo }}m</div>
            <div v-if="selection.coreLength !== undefined">岩心长度: {{ selection.coreLength }}m</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧 3D 视图 -->
    <div
      ref="viewportRef"
      class="wms-viewport"
    >
      <canvas
        ref="canvasRef"
        class="wms-canvas"
      ></canvas>
      <div class="viewport-hint">鼠标左键旋转 · 滚轮缩放 · 右键平移 · 点击货架查看详情 · 悬停岩心盒查看地质信息</div>

      <!-- ★ Tooltip 浮层 -->
      <div
        v-if="tooltip"
        class="wms-tooltip"
        :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
      >
        <div class="tooltip-header">{{ tooltip.slotLabel }}</div>
        <div class="tooltip-row">
          <span class="tooltip-label">项目编号</span>
          <span class="tooltip-value">{{ tooltip.projectNo }}</span>
        </div>
        <div class="tooltip-row">
          <span class="tooltip-label">钻孔编号</span>
          <span class="tooltip-value">{{ tooltip.holeNo }}</span>
        </div>
        <div class="tooltip-row">
          <span class="tooltip-label">起止深度</span>
          <span class="tooltip-value">{{ tooltip.depthFrom }}m ~ {{ tooltip.depthTo }}m</span>
        </div>
        <div class="tooltip-row">
          <span class="tooltip-label">岩心长度</span>
          <span class="tooltip-value tooltip-highlight">{{ tooltip.coreLength }}m</span>
        </div>
        <div class="tooltip-row">
          <span class="tooltip-label">规格 / 盒数</span>
          <span class="tooltip-value">{{ tooltip.boxSpec }} × {{ tooltip.boxCount }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { WarehouseScene, type WarehouseConfig, type SelectionInfo, type CoreBoxSpec, type TooltipInfo } from './WarehouseScene'
import warehouseData from './warehouse-data.json'

// ==================== Reactive State ====================

const canvasRef = ref<HTMLCanvasElement>()
const viewportRef = ref<HTMLDivElement>()
const sceneRef = ref<WarehouseScene>()
const activeTab = ref('inventory')
const animating = ref(false)
const selection = ref<SelectionInfo | null>(null)
const tooltip = ref<TooltipInfo | null>(null)

const config = reactive<WarehouseConfig>({ ...warehouseData.warehouse })

const putAwayForm = reactive({
  boxSpec: warehouseData.boxSpecs[0].name,
  boxCount: warehouseData.palletConfig.defaultBoxes,
  slotKey: '',
})

const pickForm = reactive({ slotKey: '' })

// ==================== Computed ====================

const stats = computed(() => sceneRef.value?.getStats() || { totalSlots: 0, occupiedSlots: 0, emptySlots: 0, totalBoxes: 0, utilization: 0 })
const availableSlots = computed(() => sceneRef.value?.getAvailableSlots() || [])
const occupiedSlots = computed(() => sceneRef.value?.getOccupiedSlots() || [])
const boxSpecs = computed<CoreBoxSpec[]>(() => warehouseData.boxSpecs)

// ==================== Config Watcher (debounced rebuild) ====================

let rebuildTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => ({ ...config }),
  () => {
    if (rebuildTimer) clearTimeout(rebuildTimer)
    rebuildTimer = setTimeout(() => {
      sceneRef.value?.updateConfig({ ...config })
    }, 300)
  },
  { deep: true }
)

// ==================== Lifecycle ====================

onMounted(() => {
  if (!canvasRef.value) return
  const ws = new WarehouseScene(canvasRef.value, {
    warehouse: { ...config },
    rackSections: warehouseData.rackSections,
    inventory: JSON.parse(JSON.stringify(warehouseData.inventory)),
    boxSpecs: warehouseData.boxSpecs,
    palletConfig: warehouseData.palletConfig,
  })
  ws.onSelectionChange = (info) => {
    selection.value = info
  }
  ws.onTooltipShow = (info) => {
    // Adjust coordinates relative to viewport
    if (viewportRef.value) {
      const rect = viewportRef.value.getBoundingClientRect()
      tooltip.value = { ...info, x: info.x - rect.left + 16, y: info.y - rect.top - 10 }
    } else {
      tooltip.value = info
    }
  }
  ws.onTooltipHide = () => {
    tooltip.value = null
  }
  ws.init()
  sceneRef.value = ws
})

onBeforeUnmount(() => {
  sceneRef.value?.dispose()
})

// ==================== Handlers ====================

function onConfigChange() {
  // Triggered by @change on input components; the watcher handles debounce
}

function handlePutAway() {
  if (!putAwayForm.slotKey) return
  const parts = putAwayForm.slotKey.split('-')
  const rackId = parts[0]
  const level = parseInt(parts[1])
  const column = parseInt(parts[2])
  animating.value = true
  sceneRef.value!.executePutAway(rackId, level, column, putAwayForm.boxSpec, putAwayForm.boxCount).then(() => {
    animating.value = false
    putAwayForm.slotKey = ''
    ElMessage.success('入库完成')
  })
}

function handlePick() {
  if (!pickForm.slotKey) return
  const parts = pickForm.slotKey.split('-')
  const rackId = parts[0]
  const level = parseInt(parts[1])
  const column = parseInt(parts[2])
  animating.value = true
  sceneRef.value!.executePick(rackId, level, column).then(() => {
    animating.value = false
    pickForm.slotKey = ''
    ElMessage.success('出库完成')
  })
}
</script>

<style scoped>
.wms-page {
  display: flex;
  height: 100%;
  background: #0f172a;
  overflow: hidden;
}

/* ===== Left Panel ===== */
.wms-panel {
  width: 320px;
  min-width: 320px;
  background: #1e293b;
  border-right: 1px solid #334155;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.panel-header {
  padding: 16px;
  border-bottom: 1px solid #334155;
}
.panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #f1f5f9;
}
.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.section {
  margin-bottom: 16px;
}
.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #94a3b8;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Buttons */
.btn-group {
  display: flex;
  gap: 6px;
}

/* Config items */
.config-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.config-item label {
  font-size: 13px;
  color: #cbd5e1;
  white-space: nowrap;
  margin-right: 12px;
}
.config-item :deep(.el-slider) {
  flex: 1;
}

/* Stats */
.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 8px;
}
.stat-item {
  background: #0f172a;
  border-radius: 6px;
  padding: 10px;
  text-align: center;
}
.stat-value {
  display: block;
  font-size: 22px;
  font-weight: 700;
  color: #60a5fa;
}
.stat-occupied {
  color: #f59e0b;
}
.stat-empty {
  color: #34d399;
}
.stat-boxes {
  color: #a78bfa;
}
.stat-label {
  font-size: 11px;
  color: #64748b;
}

/* Utilization bar */
.utilization-bar {
  height: 6px;
  background: #334155;
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 4px;
}
.utilization-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  border-radius: 3px;
  transition: width 0.3s;
}
.utilization-text {
  font-size: 12px;
  color: #94a3b8;
  text-align: right;
}

/* Tabs */
:deep(.el-tabs__item) {
  color: #94a3b8;
  font-size: 13px;
}
:deep(.el-tabs__item.is-active) {
  color: #60a5fa;
}
:deep(.el-tabs__nav-wrap::after) {
  background-color: #334155;
}

/* Inventory list */
.inventory-list {
  max-height: 220px;
  overflow-y: auto;
}
.inventory-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #cbd5e1;
  transition: background 0.15s;
}
.inventory-item:hover {
  background: #334155;
}
.inv-rack {
  background: #1e40af;
  color: #fff;
  padding: 1px 6px;
  border-radius: 3px;
  font-weight: 600;
  font-size: 11px;
}
.inv-pos {
  color: #94a3b8;
}
.inv-spec {
  color: #f59e0b;
}
.inv-count {
  margin-left: auto;
  color: #60a5fa;
  font-weight: 600;
}
.empty-text {
  text-align: center;
  color: #475569;
  padding: 20px 0;
  font-size: 13px;
}

/* Operation form */
.operation-form {
  padding-top: 4px;
}

/* Selection detail */
.selection-detail {
  background: #0f172a;
  border-radius: 6px;
  padding: 10px 12px;
  font-size: 13px;
  color: #cbd5e1;
  line-height: 1.8;
}
.selection-detail strong {
  color: #60a5fa;
}

/* ===== Right Viewport ===== */
.wms-viewport {
  flex: 1;
  position: relative;
  overflow: hidden;
}
.wms-canvas {
  width: 100%;
  height: 100%;
  display: block;
  outline: none;
}
.viewport-hint {
  position: absolute;
  bottom: 12px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.6);
  color: #94a3b8;
  padding: 6px 16px;
  border-radius: 16px;
  font-size: 12px;
  pointer-events: none;
}

/* ★ Tooltip */
.wms-tooltip {
  position: absolute;
  z-index: 100;
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  min-width: 220px;
  pointer-events: none;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(8px);
}
.tooltip-header {
  font-size: 14px;
  font-weight: 700;
  color: #60a5fa;
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid #334155;
}
.tooltip-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 2px 0;
  font-size: 12px;
}
.tooltip-label {
  color: #64748b;
}
.tooltip-value {
  color: #e2e8f0;
  font-weight: 500;
}
.tooltip-highlight {
  color: #f59e0b;
  font-weight: 700;
}

/* Dark theme overrides for Element Plus */
:deep(.el-input-number) {
  width: 120px;
}
:deep(.el-select .el-input__wrapper) {
  background-color: #0f172a;
  box-shadow: 0 0 0 1px #334155 inset;
}
:deep(.el-input__wrapper) {
  background-color: #0f172a;
  box-shadow: 0 0 0 1px #334155 inset;
}
:deep(.el-input__inner) {
  color: #e2e8f0;
}
</style>
