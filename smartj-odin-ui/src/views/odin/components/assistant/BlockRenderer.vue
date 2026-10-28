<script setup lang="ts">
import { computed, provide } from 'vue'
// 副作用导入：完成内置渲染器注册
import './blocks'
import { getBlockComponent } from './blocks/registry'
import { BlockContextKey } from './blocks/blockContext'
import UnknownBlock from './blocks/UnknownBlock.vue'
import type { ChartHint, MessageBlock, QueryDataset } from '@/types/odin'

/**
 * 块分发器：按 block.type 从注册表取渲染器，未注册则回落 UnknownBlock。
 * 交互（澄清选项、打开溯源等）统一以 `action` 事件向上冒泡。
 *
 * 消息级上下文（datasets/chartHints/streaming）通过 provide 透传，供图表卡片等需要跨块数据的渲染器 inject，
 * 不作为普通块组件的 props，避免向每个渲染器注入无关属性。
 */
const props = defineProps<{
  block: MessageBlock
  datasets?: QueryDataset[]
  chartHints?: Record<string, ChartHint>
  streaming?: boolean
}>()
const emit = defineEmits<{ action: [{ type: string; value?: string; traceId?: string }] }>()

const renderer = computed(() => getBlockComponent(props.block?.type) ?? UnknownBlock)

// props 本身是响应式代理，直接 provide 即可随父级消息更新保持响应
provide(BlockContextKey, props)

function onAction(a: { type: string; value?: string; traceId?: string }) {
  emit('action', a)
}
</script>

<template>
  <component :is="renderer" :block="props.block" @action="onAction" />
</template>
