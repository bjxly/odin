<script setup lang="ts">
import { reactive, ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { OntologyVO, OntologyDTO } from '@/types/odin'

/**
 * 本体新增/编辑弹窗。
 *
 * 除基础字段外，额外提供「助手预置问题」的动态列表编辑：
 * 每条为一个问题字符串，提交时收敛为空串过滤后的 string[]，
 * 由后端序列化到 odin_ont_definition.preset_questions JSON 列。
 */
const props = defineProps<{
  visible: boolean
  ontologyData?: OntologyVO | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: [data: OntologyDTO]
}>()

const { t } = useI18n()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const isEdit = computed(() => !!props.ontologyData?.id)
const title = computed(() =>
  isEdit.value ? t('odinOntology.ontEditTitle') : t('odinOntology.ontAddTitle'),
)

const formRef = ref<FormInstance>()
const form = reactive<OntologyDTO>({
  id: undefined,
  code: '',
  name: '',
  description: '',
  version: '1.0',
  status: 1,
  remark: '',
})

// 预置问题列表（编辑态用可增删的行数组）
const questions = ref<string[]>([])

const rules = computed<FormRules>(() => ({
  code: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.ontCode'), trigger: 'blur' }],
  name: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.ontName'), trigger: 'blur' }],
}))

function resetForm() {
  Object.assign(form, {
    id: undefined,
    code: '',
    name: '',
    description: '',
    version: '1.0',
    status: 1,
    remark: '',
  })
  questions.value = []
}

function fillFromOntology(o: OntologyVO) {
  Object.assign(form, {
    id: o.id,
    code: o.code,
    name: o.name,
    description: o.description ?? '',
    version: o.version ?? '1.0',
    status: o.status ?? 1,
    remark: o.remark ?? '',
  })
  questions.value = [...(o.presetQuestions ?? [])]
}

watch(() => props.visible, (v) => {
  if (v) {
    if (props.ontologyData) fillFromOntology(props.ontologyData)
    else resetForm()
    formRef.value?.clearValidate()
  }
})

function addQuestion() {
  questions.value.push('')
}

function removeQuestion(idx: number) {
  questions.value.splice(idx, 1)
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  const presetQuestions = questions.value.map((q) => q.trim()).filter((q) => q !== '')
  emit('save', { ...form, presetQuestions })
  dialogVisible.value = false
}
</script>

<template>
  <el-dialog v-model="dialogVisible" :title="title" width="600px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item :label="t('odinOntology.ontCode')" prop="code">
        <el-input v-model="form.code" :placeholder="t('common.pleaseInput') + t('odinOntology.ontCode')" :maxlength="64" :disabled="isEdit" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.ontName')" prop="name">
        <el-input v-model="form.name" :placeholder="t('common.pleaseInput') + t('odinOntology.ontName')" :maxlength="128" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.ontVersion')" prop="version">
        <el-input v-model="form.version" :maxlength="32" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.description')" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="3" :maxlength="500" show-word-limit />
      </el-form-item>

      <el-divider content-position="left">{{ t('odinOntology.presetQuestions') }}</el-divider>
      <div class="preset-hint">{{ t('odinOntology.presetQuestionsHint') }}</div>
      <div v-for="(q, idx) in questions" :key="idx" class="preset-row">
        <el-input
          :model-value="q"
          :maxlength="200"
          :placeholder="t('odinOntology.presetQuestionPlaceholder')"
          @update:model-value="(val: string) => (questions[idx] = val)"
        />
        <el-button :icon="Delete" circle plain type="danger" @click="removeQuestion(idx)" />
      </div>
      <el-button :icon="Plus" plain type="primary" size="small" @click="addQuestion">
        {{ t('odinOntology.addPresetQuestion') }}
      </el-button>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
      <el-button type="primary" @click="handleSubmit">{{ t('common.confirmBtn') }}</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.preset-hint {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin: 0 0 8px 100px;
}
.preset-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding-left: 100px;
}
</style>
