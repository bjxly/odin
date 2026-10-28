import type { Component } from 'vue'
import { defineAsyncComponent } from 'vue'
import { registerBlockRenderer, registerBlockRenderers } from './registry'
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
 * ECharts 版 chart 渲染器。
 *
 * 用 `defineAsyncComponent` 而非静态 import：EChartsBlock 内部 `import('@/utils/echarts')`
 * 拉取 echarts/zrender，异步包装保证只有消息流里真正出现 chart 块时才会下载该 chunk，
 * 不会把 ~300KB 的图表库拖进首屏（见 vite.config.ts 的 `echarts` 分组）。
 */
const EChartsBlock = defineAsyncComponent(() => import('./EChartsBlock.vue'))

/**
 * 内置块渲染器集合（#66 新 SQL 生成架构：退役 thinking/plan/derivation/hop）。
 *
 * 扩展方式（无需改动 BlockRenderer / 消息流）：
 *   import { registerBlockRenderer } from '@/components/assistant/blocks/registry'
 *   registerBlockRenderer('pivot', PivotBlock)
 * 同名注册即覆盖，下方已用这一机制把 chart 换成 ECharts 实现。
 */
export const BUILTIN_BLOCK_RENDERERS: Record<string, Component> = {
  text: TextBlock,
  intent: IntentBlock,
  reasoning: ReasoningBlock,
  sql: SqlBlock,
  table: TableBlock,
  chart: ChartBlock,
  provenance: ProvenanceBlock,
  clarification: ClarificationBlock,
}

// 模块加载即完成注册（副作用），BlockRenderer 只需 import 本文件
registerBlockRenderers(BUILTIN_BLOCK_RENDERERS)

/**
 * 覆盖内置的 CSS 版 chart 渲染器。
 * ChartBlock 仍保留在 BUILTIN_BLOCK_RENDERERS 中，作为 ECharts 加载失败时的降级渲染器
 * （EChartsBlock 内部捕获 import 异常后会直接渲染 `<ChartBlock />`）。
 */
registerBlockRenderer('chart', EChartsBlock)

export { ChartBlock, EChartsBlock, UnknownBlock }
export * from './registry'
