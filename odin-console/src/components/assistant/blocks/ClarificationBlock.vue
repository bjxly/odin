<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElButton, ElIcon, ElInput, ElTag } from 'element-plus'
import { Position, ChatLineSquare } from '@element-plus/icons-vue'
import type { BlockVO, ChatAction, ClarificationOptionVO, ClarificationPayload } from '@/types/assistant'

/**
 * clarification 块：意图不确定时的澄清交互。
 * options 渲染为可点按钮（点击即以该文本重发），free_input 为真时提供自由输入框。
 */
const props = defineProps<{ block: BlockVO }>()
const emit = defineEmits<{ action: [ChatAction] }>()

const p = computed<ClarificationPayload>(() => (props.block?.payload ?? {}) as ClarificationPayload)

/** 统一为 { label, value, kind }：后端为对象，也兼容纯字符串数组 */
const options = computed<ClarificationOptionVO[]>(() => {
  const raw = p.value.options ?? []
  return raw
    .map((o) => {
      if (typeof o === 'string') return { label: o, value: o, kind: '' }
      const label = String(o?.label ?? o?.value ?? '').trim()
      return { label, value: String(o?.value ?? label).trim(), kind: String(o?.kind ?? '') }
    })
    .filter((o) => !!o.label)
})

const prompt = computed(
  () => String(p.value.prompt ?? '').trim() || '我不太确定你的问题指向，请选择一个选项，或直接补充说明：',
)
const freeInput = computed(() => p.value.free_input !== false)
const picked = ref('')
const custom = ref('')

function choose(opt: ClarificationOptionVO) {
  const text = opt.value || opt.label || ''
  if (!text) return
  picked.value = opt.label ?? text
  emit('action', { type: 'clarify-option', value: text })
}

function sendCustom() {
  const text = custom.value.trim()
  if (!text) return
  emit('action', { type: 'clarify-option', value: text })
  custom.value = ''
}
</script>

<template>
  <div class="clarification-block">
    <div class="block-head">
      <ElIcon class="head-icon"><ChatLineSquare /></ElIcon>
      <span class="block-title">{{ block.title || '需要澄清' }}</span>
      <ElTag size="small" type="info" effect="plain">clarification</ElTag>
    </div>

    <p class="prompt">{{ prompt }}</p>

    <div v-if="options.length" class="options">
      <ElButton
        v-for="(o, i) in options"
        :key="i"
        class="option-btn"
        :type="picked === o.label ? 'primary' : 'default'"
        :plain="picked !== o.label"
        @click="choose(o)"
      >
        {{ o.label }}
      </ElButton>
    </div>

    <div v-if="freeInput" class="free-input">
      <ElInput
        v-model="custom"
        placeholder="也可以直接补充说明你的问题…"
        size="default"
        clearable
        @keyup.enter="sendCustom"
      >
        <template #append>
          <ElButton :icon="Position" :disabled="!custom.trim()" @click="sendCustom">发送</ElButton>
        </template>
      </ElInput>
    </div>
  </div>
</template>

<style scoped>
.clarification-block {
  border: 1px solid #fbd38d;
  border-radius: 10px;
  background: linear-gradient(180deg, #fffaf0, #ffffff);
  padding: 10px 12px;
}

.block-head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.head-icon {
  color: #dd6b20;
}

.block-title {
  font-size: 13px;
  font-weight: 700;
  color: #7b341e;
}

.prompt {
  margin: 0 0 10px;
  font-size: 13px;
  line-height: 1.7;
  color: #4a5568;
}

.options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.option-btn {
  height: auto;
  padding: 7px 12px;
  white-space: normal;
  word-break: break-word;
  text-align: left;
  justify-content: flex-start;
  line-height: 1.5;
  max-width: 100%;
}

.free-input {
  margin-top: 10px;
}
</style>
