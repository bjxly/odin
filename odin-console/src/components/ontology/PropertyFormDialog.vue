<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
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
} from 'element-plus'
import { useOntologyStore } from '@/stores/ontology'

const props = defineProps<{
  open: boolean
  classId: string
  mode: 'create' | 'edit'
  propId?: string | null
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
}>()

const store = useOntologyStore()

const ranges = [
  'string',
  'text',
  'int',
  'bigint',
  'decimal',
  'float',
  'bool',
  'date',
  'datetime',
  'enum',
  'json',
  'uuid',
]

const form = reactive({
  id: '',
  label: '',
  range: 'string',
  isKey: false,
  sensitive: false,
  synonymsText: '',
})

const isEdit = () => props.mode === 'edit' && !!props.propId

watch(
  () => [props.open, props.propId, props.mode],
  () => {
    if (!props.open) return
    if (isEdit() && props.propId) {
      const cls = store.classes.find((c) => c.id === props.classId)
      const p = cls?.properties.find((x) => x.id === props.propId)
      if (!p) return
      form.id = p.id
      form.label = p.label
      form.range = p.range
      form.isKey = !!p.isKey
      form.sensitive = !!p.sensitive
      form.synonymsText = (p.synonyms ?? []).join(', ')
    } else {
      form.id = ''
      form.label = ''
      form.range = 'string'
      form.isKey = false
      form.sensitive = false
      form.synonymsText = ''
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

async function onSubmit() {
  const cls = store.classes.find((c) => c.id === props.classId)
  const oid = store.currentOntologyId
  const synonyms = parseSynonyms()
  const pid = form.id.trim()
  const label = form.label.trim() || pid
  /** OntoProperty → 后端 OntPropertyVO 请求体 */
  const payload = {
    name: pid,
    label,
    range: form.range,
    is_key: form.isKey,
    sensitive: form.sensitive,
  }
  saving.value = true
  try {
    if (isEdit() && props.propId) {
      // 属性的数值主键不在本地态中，先向后端查询再匹配
      if (oid != null && cls?.numericId != null) {
        const remote = (await store.fetchProperties(oid, cls.numericId)) ?? []
        const match = remote.find((p) => p.name === props.propId)
        if (match) await store.updatePropertyApi(oid, cls.numericId, match.id, payload)
      }
      store.updateProperty(props.classId, props.propId as string, {
        label,
        range: form.range,
        isKey: form.isKey,
        sensitive: form.sensitive,
        synonyms,
      })
      if (oid != null) store.dirty = false
      ElMessage.success('属性已更新')
    } else {
      if (!pid) throw new Error('属性 ID 不能为空')
      if (cls?.properties.some((p) => p.id === pid)) {
        throw new Error(`属性 ID 已存在：${pid}`)
      }
      if (oid != null && cls?.numericId != null) {
        await store.createProperty(oid, cls.numericId, payload)
      }
      store.addProperty(props.classId, {
        id: pid,
        label,
        range: form.range,
        isKey: form.isKey,
        sensitive: form.sensitive,
        synonyms,
      })
      if (oid != null) store.dirty = false
      ElMessage.success('属性已添加')
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
    :title="isEdit() ? '编辑属性' : `为 ${classId} 添加属性`"
    width="460px"
    @update:model-value="emit('update:open', $event)"
  >
    <ElForm :model="form" label-width="88px">
      <ElFormItem label="属性 ID" required>
        <ElInput
          v-model="form.id"
          :disabled="isEdit()"
          :placeholder="`如 ${classId}.name`"
        />
      </ElFormItem>
      <ElFormItem label="显示名" required>
        <ElInput v-model="form.label" placeholder="如 客户名称" />
      </ElFormItem>
      <ElFormItem label="类型" required>
        <ElSelect v-model="form.range">
          <ElOption v-for="r in ranges" :key="r" :label="r" :value="r" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="业务键">
        <ElSwitch v-model="form.isKey" />
      </ElFormItem>
      <ElFormItem label="敏感">
        <ElSwitch v-model="form.sensitive" />
      </ElFormItem>
      <ElFormItem label="同义词">
        <ElInput v-model="form.synonymsText" placeholder="逗号分隔" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="emit('update:open', false)">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="onSubmit">
        {{ isEdit() ? '保存' : '添加' }}
      </ElButton>
    </template>
  </ElDialog>
</template>
