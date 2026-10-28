<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useOdinOntologyStore } from '@/stores/odin/ontology'
import type { OntRuleVO } from '@/types/odin'

/**
 * 规则详情抽屉：点击图谱上的规则虚拟节点后单独展示该规则。
 *
 * 规则不直连类，expression（JSON：condition + action）中的
 * target_class / 条件属性是仅有的类关联信息，这里解析后友好展示。
 */
const props = defineProps<{
  visible: boolean
  /** 规则节点 ID，形如 `rule:{id}` */
  nodeId: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  edit: [rule: OntRuleVO]
  delete: [rule: OntRuleVO]
}>()

const { t } = useI18n()
const store = useOdinOntologyStore()

const drawerVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const rule = computed<OntRuleVO | null>(() => {
  if (!props.nodeId) return null
  const id = props.nodeId.replace(/^rule:/, '')
  return store.rules.find((r) => r.id === id) ?? null
})

/** 解析 expression JSON（非法或为空返回 null） */
function parseExpr(raw?: string): Record<string, any> | null {
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

const condition = computed(() => parseExpr(rule.value?.expression)?.condition ?? null)
const action = computed(() => parseExpr(rule.value?.expression)?.action ?? null)

/** 规则派生目标类编码（action.target_class，规则与类仅有的显式关联） */
const targetClasses = computed(() => {
  const codes = new Set<string>()
  const target = action.value?.target_class ?? action.value?.targetClass
  if (target) codes.add(String(target))
  return [...codes]
})

function handleEdit() {
  if (rule.value) emit('edit', rule.value)
}

function handleDelete() {
  if (rule.value) emit('delete', rule.value)
}
</script>

<template>
  <el-drawer
    v-model="drawerVisible"
    class="rule-drawer"
    :title="t('odinOntology.ruleDrawerTitle')"
    direction="rtl"
    size="420px"
  >
    <template v-if="rule">
      <el-descriptions :column="1" border>
        <el-descriptions-item :label="t('odinOntology.drawerRuleName')">
          {{ rule.name }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinOntology.drawerRuleCode')">
          <code>{{ rule.code }}</code>
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinOntology.drawerRuleType')">
          <el-tag size="small" type="success">{{ rule.ruleType }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinOntology.drawerRulePriority')">
          {{ rule.priority ?? '—' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinOntology.drawerRuleEnabled')">
          {{ rule.enabled ? t('common.yes') : t('common.no') }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinOntology.drawerRuleTargetClass')">
          <template v-if="targetClasses.length">
            <el-tag v-for="c in targetClasses" :key="c" size="small" class="rule-class-tag">
              {{ c }}
            </el-tag>
          </template>
          <span v-else>—</span>
        </el-descriptions-item>
        <el-descriptions-item :label="t('odinOntology.drawerDesc')">
          {{ rule.description || '—' }}
        </el-descriptions-item>
      </el-descriptions>

      <div v-if="condition || action" class="rule-expr-block">
        <div v-if="condition" class="rule-expr-item">
          <span class="rule-expr-label">{{ t('odinOntology.drawerRuleCondition') }}</span>
          <pre class="rule-expr-json">{{ JSON.stringify(condition, null, 2) }}</pre>
        </div>
        <div v-if="action" class="rule-expr-item">
          <span class="rule-expr-label">{{ t('odinOntology.drawerRuleAction') }}</span>
          <pre class="rule-expr-json">{{ JSON.stringify(action, null, 2) }}</pre>
        </div>
      </div>
    </template>
    <el-empty v-else :image-size="60" :description="t('odinOntology.drawerNoRules')" />

    <template v-if="rule" #footer>
      <div class="rule-drawer-footer">
        <el-button size="small" @click="handleDelete">{{ t('common.delete') }}</el-button>
        <el-button size="small" type="primary" @click="handleEdit">{{ t('common.edit') }}</el-button>
      </div>
    </template>
  </el-drawer>
</template>

<style scoped>
.rule-expr-block {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 16px;
}

.rule-expr-label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.rule-expr-json {
  padding: 10px 12px;
  margin: 0;
  overflow-x: auto;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
  word-break: break-all;
  background: var(--el-fill-color-light);
  border-radius: 6px;
}

.rule-class-tag {
  margin-right: 6px;
}

.rule-drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
