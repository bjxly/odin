<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

/**
 * 解析路径徽章：rule=蓝 / llm=紫 / cache=灰 / local=橙 / concept=绿 / clarification=黄
 */
const { t } = useI18n()

const props = withDefaults(defineProps<{
  path?: string | null
}>(), {
  path: '',
})

interface PathMeta {
  label: string
  type: 'primary' | 'success' | 'warning' | 'danger' | 'info' | ''
  tip: string
}

const meta = computed<PathMeta | null>(() => {
  const p = String(props.path ?? '').trim().toLowerCase()
  if (!p) return null
  const map: Record<string, PathMeta> = {
    rule: { label: t('odinAssistant.pathRule'), type: 'primary', tip: '零 LLM 调用，纯规则解析' },
    llm: { label: t('odinAssistant.pathLlm'), type: '', tip: 'LLM 意图链解析' },
    cache: { label: t('odinAssistant.pathCache'), type: 'info', tip: '命中查询计划缓存' },
    local: { label: t('odinAssistant.pathLocal'), type: 'warning', tip: '离线兜底处理' },
    concept: { label: t('odinAssistant.pathConcept'), type: 'success', tip: '概念知识库问答' },
    clarification: { label: t('odinAssistant.pathClarification'), type: 'warning', tip: '需要用户澄清' },
    meta: { label: t('odinAssistant.pathMeta'), type: 'info', tip: '元信息技能' },
  }
  return map[p] ?? { label: p, type: 'info', tip: p }
})
</script>

<template>
  <el-tooltip v-if="meta" :content="meta.tip" placement="top" :show-after="200">
    <el-tag :type="meta.type || undefined" size="small" effect="light" class="parse-path-badge">
      {{ meta.label }}
    </el-tag>
  </el-tooltip>
</template>

<style scoped>
.parse-path-badge {
  font-weight: 600;
  letter-spacing: 0.2px;
}
</style>
