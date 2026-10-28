import type { InjectionKey } from 'vue'
import type { ChartHint, MessageBlock, QueryDataset } from '@/types/odin'

/**
 * 块渲染上下文：由 BlockRenderer 从消息级作用域 provide，供需要跨块数据的渲染器（图表卡片）inject。
 *
 * 图表卡片本身只拿到 `chart` 块（含 datasetId），需要回到消息的 datasets/chartHints 才能取到数据集快照
 * 与模型推荐；这些不属于单块 payload，故通过上下文透传，避免污染所有块组件的 props。
 */
export interface BlockContext {
  datasets?: QueryDataset[]
  chartHints?: Record<string, ChartHint>
  streaming?: boolean
  block?: MessageBlock
}

export const BlockContextKey: InjectionKey<BlockContext> = Symbol('OdinBlockContext')
