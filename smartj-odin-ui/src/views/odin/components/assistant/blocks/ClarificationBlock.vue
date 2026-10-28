<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Position, ChatLineSquare } from '@element-plus/icons-vue'
import type { MessageBlock, ClarificationBlockPayload } from '@/types/odin'

export interface ClarifyAction {
  type: 'clarify-option'
  value: string
}

const { t } = useI18n()
const props = defineProps<{ block: MessageBlock }>()
const emit = defineEmits<{ action: [ClarifyAction] }>()

const p = computed<ClarificationBlockPayload>(() => (props.block?.payload ?? {}) as ClarificationBlockPayload)

interface OptItem { label: string; value: string }

const options = computed<OptItem[]>(() => {
  const raw = p.value.options ?? []
  return raw
    .map((o) => {
      if (typeof o === 'string') return { label: o, value: o }
      const label = String(o?.label ?? o?.value ?? '').trim()
      return { label, value: String(o?.value ?? label).trim() }
    })
    .filter((o) => !!o.label)
})

const prompt = computed(() => String(p.value.prompt ?? '').trim() || t('odinAssistant.clarifyPrompt'))
const freeInput = computed(() => p.value.freeInput !== false)
const picked = ref('')
const custom = ref('')

function choose(opt: OptItem) {
  const text = opt.value || opt.label || ''
  if (!text) return
  picked.value = opt.label
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
      <el-icon class="head-icon"><ChatLineSquare /></el-icon>
      <span class="block-title">{{ block.title || t('odinAssistant.blockClarification') }}</span>
      <el-tag size="small" type="info" effect="plain">clarification</el-tag>
    </div>

    <p class="prompt">{{ prompt }}</p>

    <div v-if="options.length" class="options">
      <el-button
        v-for="(o, i) in options"
        :key="i"
        class="option-btn"
        :type="picked === o.label ? 'primary' : 'default'"
        :plain="picked !== o.label"
        @click="choose(o)"
      >
        {{ o.label }}
      </el-button>
    </div>

    <div v-if="freeInput" class="free-input">
      <el-input
        v-model="custom"
        :placeholder="t('odinAssistant.clarifyFreeInput')"
        size="default"
        clearable
        @keyup.enter="sendCustom"
      >
        <template #append>
          <el-button :icon="Position" :disabled="!custom.trim()" @click="sendCustom">
            {{ t('odinAssistant.clarifySend') }}
          </el-button>
        </template>
      </el-input>
    </div>
  </div>
</template>

<style scoped>
.clarification-block {
  border: 1px solid var(--el-color-warning-light-5);
  border-radius: 10px;
  background: linear-gradient(180deg, var(--el-color-warning-light-9), var(--el-fill-color-blank));
  padding: 10px 12px;
}
.block-head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}
.head-icon {
  color: var(--el-color-warning);
}
.block-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.prompt {
  margin: 0 0 10px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--el-text-color-regular);
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
