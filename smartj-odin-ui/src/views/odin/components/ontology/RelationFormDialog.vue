<script setup lang="ts">
import { reactive, ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import type { OntRelationVO, OntRelationDTO } from '@/types/odin'
import { useOdinOntologyStore } from '@/stores/odin/ontology'

const props = defineProps<{
  visible: boolean
  ontologyId: string
  relationData?: OntRelationVO | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: [data: OntRelationDTO]
}>()

const { t } = useI18n()
const store = useOdinOntologyStore()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const isEdit = computed(() => !!props.relationData?.id)
const title = computed(() => isEdit.value ? t('odinOntology.relationEditTitle') : t('odinOntology.relationAddTitle'))

const CARDINALITIES = ['1:1', '1:N', 'N:1', 'N:M']

const formRef = ref<FormInstance>()
const form = reactive<OntRelationDTO>({
  id: undefined,
  ontologyId: '',
  code: '',
  name: '',
  description: '',
  sourceClassId: '',
  targetClassId: '',
  cardinality: '1:N',
  joinKeys: '',
  inverseName: '',
  status: 1,
  remark: '',
})

const rules = computed<FormRules>(() => ({
  code: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.relationCode'), trigger: 'blur' }],
  name: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.relationName'), trigger: 'blur' }],
  sourceClassId: [{ required: true, message: t('common.pleaseSelect') + t('odinOntology.sourceClass'), trigger: 'change' }],
  targetClassId: [{ required: true, message: t('common.pleaseSelect') + t('odinOntology.targetClass'), trigger: 'change' }],
  cardinality: [{ required: true, message: t('common.pleaseSelect') + t('odinOntology.cardinality'), trigger: 'change' }],
}))

const classOptions = computed(() => store.classes)

watch(() => props.visible, (v) => {
  if (v) {
    if (props.relationData) {
      Object.assign(form, {
        id: props.relationData.id,
        ontologyId: props.relationData.ontologyId || props.ontologyId,
        code: props.relationData.code,
        name: props.relationData.name,
        description: props.relationData.description ?? '',
        sourceClassId: props.relationData.sourceClassId,
        targetClassId: props.relationData.targetClassId,
        cardinality: props.relationData.cardinality ?? '1:N',
        joinKeys: props.relationData.joinKeys ?? '',
        inverseName: props.relationData.inverseName ?? '',
        status: props.relationData.status ?? 1,
        remark: props.relationData.remark ?? '',
      })
    } else {
      Object.assign(form, {
        id: undefined,
        ontologyId: props.ontologyId,
        code: '',
        name: '',
        description: '',
        sourceClassId: '',
        targetClassId: '',
        cardinality: '1:N',
        joinKeys: '',
        inverseName: '',
        status: 1,
        remark: '',
      })
    }
    formRef.value?.clearValidate()
  }
})

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  emit('save', { ...form })
  dialogVisible.value = false
}
</script>

<template>
  <el-dialog v-model="dialogVisible" :title="title" width="560px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item :label="t('odinOntology.relationCode')" prop="code">
        <el-input v-model="form.code" :maxlength="64" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.relationName')" prop="name">
        <el-input v-model="form.name" :maxlength="128" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.sourceClass')" prop="sourceClassId">
        <el-select v-model="form.sourceClassId" filterable style="width: 100%">
          <el-option v-for="c in classOptions" :key="c.id" :label="`${c.name}（${c.code}）`" :value="c.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('odinOntology.targetClass')" prop="targetClassId">
        <el-select v-model="form.targetClassId" filterable style="width: 100%">
          <el-option v-for="c in classOptions" :key="c.id" :label="`${c.name}（${c.code}）`" :value="c.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('odinOntology.cardinality')" prop="cardinality">
        <el-select v-model="form.cardinality" style="width: 100%">
          <el-option v-for="card in CARDINALITIES" :key="card" :label="card" :value="card" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('odinOntology.joinKeys')" prop="joinKeys">
        <el-input v-model="form.joinKeys" placeholder="sourceCol=targetCol" :maxlength="256" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.inverseName')" prop="inverseName">
        <el-input v-model="form.inverseName" :maxlength="128" />
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
