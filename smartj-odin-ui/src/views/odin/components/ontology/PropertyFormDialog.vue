<script setup lang="ts">
import { reactive, ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import type { OntPropertyVO, OntPropertyDTO } from '@/types/odin'

const props = defineProps<{
  visible: boolean
  classId: string
  propertyData?: OntPropertyVO | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: [data: OntPropertyDTO]
}>()

const { t } = useI18n()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const isEdit = computed(() => !!props.propertyData?.id)
const title = computed(() => isEdit.value ? t('odinOntology.propertyEditTitle') : t('odinOntology.propertyAddTitle'))

const DATA_TYPES = ['string', 'number', 'boolean', 'date', 'json', 'enum', 'text', 'integer', 'float']

const formRef = ref<FormInstance>()
const form = reactive<OntPropertyDTO>({
  id: undefined,
  classId: '',
  code: '',
  name: '',
  dataType: 'string',
  description: '',
  required: 0,
  defaultValue: '',
  enumValues: '',
  unit: '',
  sortOrder: 0,
  status: 1,
})

const rules = computed<FormRules>(() => ({
  code: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.propertyCode'), trigger: 'blur' }],
  name: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.propertyName'), trigger: 'blur' }],
  dataType: [{ required: true, message: t('common.pleaseSelect') + t('odinOntology.dataType'), trigger: 'change' }],
}))

const showEnum = computed(() => form.dataType === 'enum')

watch(() => props.visible, (v) => {
  if (v) {
    if (props.propertyData) {
      Object.assign(form, {
        id: props.propertyData.id,
        classId: props.propertyData.classId || props.classId,
        code: props.propertyData.code,
        name: props.propertyData.name,
        dataType: props.propertyData.dataType ?? 'string',
        description: props.propertyData.description ?? '',
        required: props.propertyData.required ?? 0,
        defaultValue: props.propertyData.defaultValue ?? '',
        enumValues: props.propertyData.enumValues ?? '',
        unit: props.propertyData.unit ?? '',
        sortOrder: props.propertyData.sortOrder ?? 0,
        status: props.propertyData.status ?? 1,
      })
    } else {
      Object.assign(form, {
        id: undefined,
        classId: props.classId,
        code: '',
        name: '',
        dataType: 'string',
        description: '',
        required: 0,
        defaultValue: '',
        enumValues: '',
        unit: '',
        sortOrder: 0,
        status: 1,
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
  <el-dialog v-model="dialogVisible" :title="title" width="520px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item :label="t('odinOntology.propertyCode')" prop="code">
        <el-input v-model="form.code" :maxlength="64" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.propertyName')" prop="name">
        <el-input v-model="form.name" :maxlength="128" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.dataType')" prop="dataType">
        <el-select v-model="form.dataType" style="width: 100%">
          <el-option v-for="dt in DATA_TYPES" :key="dt" :label="dt" :value="dt" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('odinOntology.required')" prop="required">
        <el-switch v-model="form.required" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item v-if="showEnum" :label="t('odinOntology.enumValues')" prop="enumValues">
        <el-input v-model="form.enumValues" type="textarea" :rows="2" placeholder="val1,val2,val3" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.defaultValue')" prop="defaultValue">
        <el-input v-model="form.defaultValue" :maxlength="256" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.description')" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="2" :maxlength="500" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.unit')" prop="unit">
        <el-input v-model="form.unit" :maxlength="32" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.sortOrder')" prop="sortOrder">
        <el-input-number v-model="form.sortOrder" :min="0" :max="9999" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
      <el-button type="primary" @click="handleSubmit">{{ t('common.confirmBtn') }}</el-button>
    </template>
  </el-dialog>
</template>
