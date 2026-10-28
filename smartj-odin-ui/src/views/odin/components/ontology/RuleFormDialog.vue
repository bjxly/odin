<script setup lang="ts">
import { reactive, ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import { useOdinOntologyStore } from '@/stores/odin/ontology'
import ConditionGroupEditor from './ConditionGroupEditor.vue'
import type { GroupNode } from './conditionTree'
import { newGroup, parseCondition, parseTargetClass, serializeCondition } from './conditionTree'
import type { OntRuleVO, OntRuleDTO } from '@/types/odin'

/**
 * 规则新增/编辑弹窗。
 *
 * 规则的类关联信息全部编码在 expression JSON（condition + action）里，
 * 这里用递归条件组编辑器（AND/OR/NOT + 子条件/子组）替代单条件构造器，
 * 提交时序列化为分组 JSON（单条件自动折叠为旧版叶子结构），
 * 编辑时反解析回填，避免手填写错结构。
 */
const props = defineProps<{
  visible: boolean
  ontologyId: string
  ruleData?: OntRuleVO | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: [data: OntRuleDTO]
}>()

const { t } = useI18n()
const store = useOdinOntologyStore()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const isEdit = computed(() => !!props.ruleData?.id)
const title = computed(() => (isEdit.value ? t('odinOntology.ruleEditTitle') : t('odinOntology.ruleAddTitle')))

const RULE_TYPES = ['derivation', 'constraint', 'enrichment']
const OPERATORS = ['eq', 'ne', 'gt', 'gte', 'lt', 'lte', 'in', 'like']
const ACTION_TYPES = ['expand_class']

const formRef = ref<FormInstance>()
const form = reactive<OntRuleDTO>({
  id: undefined,
  ontologyId: '',
  code: '',
  name: '',
  description: '',
  ruleType: 'derivation',
  priority: 100,
  enabled: 1,
  status: 1,
})

// 条件组编辑器（递归模型）
const rootGroup = ref<GroupNode>(newGroup())
// 无法识别的旧表达式原样保留标记（避免编辑动作时丢失条件）
const legacyExpression = ref('')
// 动作构造器
const actionType = ref('expand_class')
const actionTarget = ref('')

const rules = computed<FormRules>(() => ({
  code: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.ruleCode'), trigger: 'blur' }],
  name: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.ruleName'), trigger: 'blur' }],
  ruleType: [{ required: true, message: t('common.pleaseSelect') + t('odinOntology.ruleType'), trigger: 'change' }],
}))

/** 可选条件属性：本体内类属性（去重），标签为「类名.属性名称（编码）」，值仍为编码 */
const propertyOptions = computed<{ value: string; label: string }[]>(() => {
  const seen = new Set<string>()
  const list: { value: string; label: string }[] = []
  for (const n of store.graph?.nodes ?? []) {
    const className = n.label || ''
    for (const p of n.properties ?? []) {
      if (seen.has(p)) continue
      seen.add(p)
      const pd = n.propertyDetails?.find((d) => d.code === p)
      const propName = pd?.name && pd.name !== p ? pd.name : ''
      list.push({ value: p, label: `${className || p}.${propName || p}（${p}）` })
    }
  }
  return list.sort((a, b) => a.label.localeCompare(b.label, 'zh-CN'))
})

/** 可选目标类：本体内类编码 */
const classOptions = computed(() => store.classes)

function resetForm() {
  Object.assign(form, {
    id: undefined,
    ontologyId: props.ontologyId,
    code: '',
    name: '',
    description: '',
    ruleType: 'derivation',
    priority: 100,
    enabled: 1,
    status: 1,
  })
  rootGroup.value = newGroup()
  legacyExpression.value = ''
  actionType.value = 'expand_class'
  actionTarget.value = ''
}

function fillFromRule(r: OntRuleVO) {
  Object.assign(form, {
    id: r.id,
    ontologyId: r.ontologyId || props.ontologyId,
    code: r.code,
    name: r.name,
    description: r.description ?? '',
    ruleType: r.ruleType ?? 'derivation',
    priority: r.priority ?? 100,
    enabled: r.enabled ?? 1,
    status: r.status ?? 1,
  })
  rootGroup.value = parseCondition(r.expression)
  legacyExpression.value = !r.expression ? '' : (rootGroup.value.items.length || rootGroup.value.groups.length) ? '' : r.expression
  actionType.value = parseActionType(r.expression)
  actionTarget.value = parseTargetClass(r.expression)
}

/** expression JSON 字符串 → action.type（默认 expand_class） */
function parseActionType(exprJson?: string | null): string {
  if (!exprJson) return 'expand_class'
  try {
    return String(JSON.parse(exprJson)?.action?.type ?? 'expand_class')
  } catch {
    return 'expand_class'
  }
}

watch(() => props.visible, (v) => {
  if (v) {
    if (props.ruleData) fillFromRule(props.ruleData)
    else resetForm()
    formRef.value?.clearValidate()
  }
})

function handleSubmit() {
  formRef.value?.validate((valid) => {
    if (!valid) return
    const serialized = serializeCondition(rootGroup.value, actionType.value, actionTarget.value)
    // 条件组完全未填时：保留原表达式（可只改动作/基本信息）；否则以条件组为准
    const expression = serialized ?? legacyExpression.value ?? ''
    emit('save', { ...form, expression })
    dialogVisible.value = false
  })
}
</script>

<template>
  <el-dialog v-model="dialogVisible" :title="title" width="600px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item :label="t('odinOntology.ruleCode')" prop="code">
        <el-input v-model="form.code" :maxlength="64" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.ruleName')" prop="name">
        <el-input v-model="form.name" :maxlength="128" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.ruleType')" prop="ruleType">
        <el-select v-model="form.ruleType" style="width: 100%">
          <el-option v-for="rt in RULE_TYPES" :key="rt" :label="t(`odinOntology.ruleType_${rt}`)" :value="rt" />
        </el-select>
      </el-form-item>

      <el-divider content-position="left">{{ t('odinOntology.ruleCondition') }}</el-divider>
      <el-form-item>
        <ConditionGroupEditor
          :group="rootGroup"
          :depth="0"
          :property-options="propertyOptions"
          :operators="OPERATORS"
        />
      </el-form-item>

      <el-divider content-position="left">{{ t('odinOntology.ruleAction') }}</el-divider>
      <el-form-item :label="t('odinOntology.ruleActionType')">
        <el-select v-model="actionType" style="width: 100%">
          <el-option v-for="at in ACTION_TYPES" :key="at" :label="at" :value="at" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('odinOntology.ruleActionTarget')">
        <el-select v-model="actionTarget" filterable allow-create default-first-option style="width: 100%">
          <el-option
            v-for="c in classOptions"
            :key="c.id"
            :label="`${c.name}（${c.code}）`"
            :value="c.code"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('odinOntology.rulePriority')">
        <el-input-number v-model="form.priority" :min="0" :max="9999" controls-position="right" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.description')" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="2" :maxlength="500" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
      <el-button type="primary" @click="handleSubmit">{{ t('common.confirmBtn') }}</el-button>
    </template>
  </el-dialog>
</template>
