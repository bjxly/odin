<script setup lang="ts">
import { computed } from 'vue'
// 副作用导入：完成内置渲染器注册
import './blocks'
import { getBlockRenderer } from './blocks/registry'
import UnknownBlock from './blocks/UnknownBlock.vue'
import type { BlockVO, ChatAction } from '@/types/assistant'

/**
 * 块分发器：按 block.type 从注册表取渲染器，未注册则回落 UnknownBlock。
 * 交互（澄清选项、打开溯源等）统一以 `action` 事件向上冒泡，由页面处理。
 */
const props = defineProps<{ block: BlockVO }>()
const emit = defineEmits<{ action: [ChatAction] }>()

const renderer = computed(() => getBlockRenderer(props.block?.type) ?? UnknownBlock)

function onAction(a: ChatAction) {
  emit('action', a)
}
</script>

<template>
  <component :is="renderer" :block="block" @action="onAction" />
</template>
