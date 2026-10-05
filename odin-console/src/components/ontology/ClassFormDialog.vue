<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import {
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElSelect,
  ElOption,
  ElSwitch,
  ElButton,
  ElMessage,
  ElTag,
} from 'element-plus'
import { useOntologyStore } from '@/stores/ontology'

const props = defineProps<{
  open: boolean
  mode: 'create' | 'edit'
  classId?: string | null
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
}>()

const store = useOntologyStore()

const form = reactive({
  id: '',
  label: '',
  description: '',
  domain: '',
  synonymsText: '',
  subclassOf: '',
  abstract: false,
})

const isEdit = computed(() => props.mode === 'edit' && !!props.classId)

const classOptions = computed(() =>
  store.classes
    .filter((c) => c.id !== props.classId)
    .map((c) => ({ id: c.id, label: `${c.label}（${c.id}）` })),
)

const domainOptions = computed(() => store.domains)

watch(
  () => [props.open, props.classId, props.mode],
  () => {
    if (!props.open) return
    if (isEdit.value && props.classId) {
      const c = store.classes.find((x) => x.id === props.classId)
      if (!c) return
      form.id = c.id
      form.label = c.label
      form.description = c.description ?? ''
      form.domain = c.domain ?? ''
      form.synonymsText = (c.synonyms ?? []).join(', ')
      form.subclassOf = c.subclassOf ?? ''
      form.abstract = !!c.abstract
    } else {
      form.id = ''
      form.label = ''
      form.description = ''
      form.domain = ''
      form.synonymsText = ''
      form.subclassOf = ''
      form.abstract = false
    }
  },
  { immediate: true },
)

function parseSynonyms(): string[] {
  return form.synonymsText
    .split(/[,，]/)
    .map((s) => s.trim())
    .filter(Boolean)
}

const saving = ref(false)

/** 父类名 → 后端数值主键 */
function parentClassId(): number | undefined {
  if (!form.subclassOf) return undefined
  return store.classes.find((c) => c.id === form.subclassOf)?.numericId
}

async function onSubmit() {
  const synonyms = parseSynonyms()
  const id = form.id.trim()
  const label = form.label.trim() || id
  const oid = store.currentOntologyId
  saving.value = true
  try {
    if (isEdit.value && props.classId) {
      const target = store.classes.find((c) => c.id === props.classId)
      if (oid != null && target?.numericId != null) {
        await store.updateClassApi(oid, target.numericId, {
          name: target.id,
          label,
          description: form.description || undefined,
          domain: form.domain || undefined,
          is_abstract: form.abstract,
          parent_class_id: parentClassId(),
        })
        // 同义词后端未建模，仅回写本地态以保持界面一致
        if (synonyms.length) {
          store.updateClass(props.classId, { synonyms })
          store.dirty = false
        }
      } else {
        store.updateClass(props.classId, {
          label,
          description: form.description || undefined,
          domain: form.domain || undefined,
          synonyms,
          subclassOf: form.subclassOf || undefined,
          abstract: form.abstract,
        })
      }
      ElMessage.success('类已更新')
    } else {
      if (!id) throw new Error('类 ID 不能为空')
      if (store.classExists(id)) throw new Error(`类 ID 已存在：${id}`)
      if (oid != null) {
        await store.createClass(oid, {
          name: id,
          label,
          description: form.description || undefined,
          domain: form.domain || undefined,
          is_abstract: form.abstract,
          parent_class_id: parentClassId(),
        })
        if (synonyms.length) {
          store.updateClass(id, { synonyms })
          store.dirty = false
        }
      } else {
        store.addClass({
          id,
          label,
          description: form.description || undefined,
          domain: form.domain || undefined,
          synonyms,
          subclassOf: form.subclassOf || undefined,
          abstract: form.abstract,
        })
      }
      ElMessage.success('类已创建')
    }
    emit('update:open', false)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <ElDialog
    :model-value="open"
    :title="isEdit ? `编辑类：${form.id}` : '新建类'"
    width="480px"
    @update:model-value="emit('update:open', $event)"
  >
    <ElForm :model="form" label-width="88px" size="default">
      <ElFormItem label="类 ID" required>
        <ElInput
          v-model="form.id"
          :disabled="isEdit"
          placeholder="如 Customer（创建后不可改）"
        />
      </ElFormItem>
      <ElFormItem label="显示名" required>
        <ElInput v-model="form.label" placeholder="如 客户" />
      </ElFormItem>
      <ElFormItem label="描述">
        <ElInput v-model="form.description" type="textarea" :rows="2" />
      </ElFormItem>
      <ElFormItem label="领域">
        <ElSelect
          v-model="form.domain"
          filterable
          allow-create
          default-first-option
          placeholder="选择或输入领域"
          clearable
        >
          <ElOption v-for="d in domainOptions" :key="d" :label="d" :value="d" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="同义词">
        <ElInput v-model="form.synonymsText" placeholder="逗号分隔，供 NL/映射推断" />
      </ElFormItem>
      <ElFormItem label="父类">
        <ElSelect v-model="form.subclassOf" clearable placeholder="可选">
          <ElOption
            v-for="c in classOptions"
            :key="c.id"
            :label="c.label"
            :value="c.id"
          />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="虚拟类">
        <ElSwitch v-model="form.abstract" />
        <ElTag v-if="form.abstract" size="small" type="info" style="margin-left: 8px">
          规则/抽象类，无物理映射
        </ElTag>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="emit('update:open', false)">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="onSubmit">
        {{ isEdit ? '保存' : '创建' }}
      </ElButton>
    </template>
  </ElDialog>
</template>
