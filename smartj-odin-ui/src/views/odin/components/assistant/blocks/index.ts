import type { Component } from 'vue'
import { defineAsyncComponent } from 'vue'
import { registerBlock, registerBlocks } from './registry'
import TextBlock from './TextBlock.vue'
import IntentBlock from './IntentBlock.vue'
import ReasoningBlock from './ReasoningBlock.vue'
import SqlBlock from './SqlBlock.vue'
import TableBlock from './TableBlock.vue'
import ChartBlock from './ChartBlock.vue'
import ProvenanceBlock from './ProvenanceBlock.vue'
import ClarificationBlock from './ClarificationBlock.vue'
import UnknownBlock from './UnknownBlock.vue'

/**
 * 数据集图表卡片（懒加载）：dataset 模式（引用 datasetId）渲染交互式 ECharts 工具栏卡片；
 * 无法解析数据集时内部委派给 EChartsBlock 兼容旧版整块 payload。
 * 只有消息流里出现可出图的 chart 块时才下载 echarts chunk。
 */
const DatasetChartCard = defineAsyncComponent(() => import('./DatasetChartCard.vue'))

/**
 * 内置块渲染器集合。
 * 扩展方式：import { registerBlock } from './registry'; registerBlock('pivot', PivotBlock)
 */
export const BUILTIN_BLOCKS: Record<string, Component> = {
  text: TextBlock,
  intent: IntentBlock,
  reasoning: ReasoningBlock,
  sql: SqlBlock,
  table: TableBlock,
  chart: ChartBlock,
  provenance: ProvenanceBlock,
  clarification: ClarificationBlock,
}

// 模块加载即完成注册（副作用）
registerBlocks(BUILTIN_BLOCKS)

// 覆盖 chart 为数据集图表卡片（EChartsBlock/ChartBlock 保留为旧 payload 兼容与降级层）
registerBlock('chart', DatasetChartCard)

export { ChartBlock, DatasetChartCard, UnknownBlock }
export * from './registry'
