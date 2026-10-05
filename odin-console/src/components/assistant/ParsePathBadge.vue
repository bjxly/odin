<script setup lang="ts">
import { computed } from 'vue'
import { ElTag, ElTooltip } from 'element-plus'
import { parsePathBadge } from '@/types/assistant'

/**
 * parse_path 徽章：
 * rule=规则解析·零LLM / llm=LLM意图链 / cache=计划缓存 /
 * concept=概念知识库 / clarification=澄清 / meta=元信息技能 / local=离线兜底
 */
const props = withDefaults(
  defineProps<{
    path?: string | null
    size?: 'large' | 'default' | 'small'
    effect?: 'dark' | 'light' | 'plain'
    /** 无 path 时是否隐藏（false 则不渲染任何内容） */
    showEmpty?: boolean
  }>(),
  {
    path: '',
    size: 'small',
    effect: 'light',
    showEmpty: false,
  },
)

const meta = computed(() => parsePathBadge(props.path))
</script>

<template>
  <ElTooltip v-if="meta" :content="meta.tip" placement="top" :show-after="200">
    <ElTag :type="meta.type" :size="size" :effect="effect" class="parse-path-badge">
      {{ meta.label }}
    </ElTag>
  </ElTooltip>
  <ElTag v-else-if="showEmpty" :size="size" :effect="effect" type="info" class="parse-path-badge">
    未标注路径
  </ElTag>
</template>

<style scoped>
.parse-path-badge {
  font-weight: 600;
  letter-spacing: 0.2px;
}
</style>
