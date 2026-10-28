<script setup lang="ts">
import { reactive, ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import type { OntClassVO, OntClassDTO } from '@/types/odin'
import { useOdinOntologyStore } from '@/stores/odin/ontology'

const props = defineProps<{
  visible: boolean
  ontologyId: string
  classData?: OntClassVO | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  save: [data: OntClassDTO]
}>()

const { t } = useI18n()
const store = useOdinOntologyStore()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const isEdit = computed(() => !!props.classData?.id)
const title = computed(() => isEdit.value ? t('odinOntology.classEditTitle') : t('odinOntology.classAddTitle'))

const formRef = ref<FormInstance>()
const form = reactive<OntClassDTO>({
  id: undefined,
  ontologyId: '',
  code: '',
  name: '',
  description: '',
  parentClassId: '',
  icon: '',
  sortOrder: 0,
  identityKeys: '',
  status: 1,
  remark: '',
})

const rules = computed<FormRules>(() => ({
  code: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.classCode'), trigger: 'blur' }],
  name: [{ required: true, message: t('common.pleaseInput') + t('odinOntology.className'), trigger: 'blur' }],
}))

/** 可选父类（排除自身） */
const parentOptions = computed(() =>
  store.classes.filter((c) => c.id !== props.classData?.id),
)

watch(() => props.visible, (v) => {
  if (v) {
    if (props.classData) {
      Object.assign(form, {
        id: props.classData.id,
        ontologyId: props.classData.ontologyId || props.ontologyId,
        code: props.classData.code,
        name: props.classData.name,
        description: props.classData.description ?? '',
        parentClassId: props.classData.parentClassId ?? '',
        icon: props.classData.icon ?? '',
        sortOrder: props.classData.sortOrder ?? 0,
        identityKeys: props.classData.identityKeys ?? '',
        status: props.classData.status ?? 1,
        remark: props.classData.remark ?? '',
      })
    } else {
      Object.assign(form, {
        id: undefined,
        ontologyId: props.ontologyId,
        code: '',
        name: '',
        description: '',
        parentClassId: '',
        icon: '',
        sortOrder: 0,
        identityKeys: '',
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
      <el-form-item :label="t('odinOntology.classCode')" prop="code">
        <el-input v-model="form.code" :placeholder="t('common.pleaseInput') + t('odinOntology.classCode')" :maxlength="64" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.className')" prop="name">
        <el-input v-model="form.name" :placeholder="t('common.pleaseInput') + t('odinOntology.className')" :maxlength="128" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.parentClass')" prop="parentClassId">
        <el-select v-model="form.parentClassId" clearable :placeholder="t('odinOntology.rootClass')" style="width: 100%">
          <el-option v-for="c in parentOptions" :key="c.id" :label="`${c.name}（${c.code}）`" :value="c.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('odinOntology.description')" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="3" :maxlength="500" show-word-limit />
      </el-form-item>
      <el-form-item :label="t('odinOntology.icon')" prop="icon">
        <el-input v-model="form.icon" placeholder="emoji or icon name" :maxlength="64" />
      </el-form-item>
      <el-form-item :label="t('odinOntology.identityKeys')" prop="identityKeys">
        <el-input v-model="form.identityKeys" placeholder="key1,key2" :maxlength="256" />
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
