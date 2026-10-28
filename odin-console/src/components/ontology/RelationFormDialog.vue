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
} from 'element-plus'
import { useOntologyStore } from '@/stores/ontology'

const props = defineProps<{
  open: boolean
  /** 预填起点类 */
  domainId?: string | null
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
}>()

const store = useOntologyStore()

const form = reactive({
  domain: '',
  range: '',
  label: '',
  cardinality: 'N:1',
  inverse: '',
  crossSource: false,
})

const classOptions = computed(() =>
  store.classes.map((c) => ({ id: c.id, label: `${c.label}（${c.id}）` })),
)

const cardinalities = ['1:1', '1:N', 'N:1', 'N:M']

watch(
  () => [props.open, props.domainId],
  () => {
    if (!props.open) return
    form.domain = props.domainId ?? ''
    form.range = ''
    form.label = ''
    form.cardinality = 'N:1'
    form.inverse = ''
    form.crossSource = false
  },
  { immediate: true },
)

const saving = ref(false)

async function onSubmit() {
  const oid = store.currentOntologyId
  const label = form.label.trim()
  const name = `${form.domain}.${label || 'rel'}.${form.range}`.replace(/\s+/g, '')
  saving.value = true
  try {
    if (!form.domain) throw new Error('请选择起点类')
    if (!form.range) throw new Error('请选择终点类')
    if (!store.classExists(form.domain)) throw new Error(`起点类不存在：${form.domain}`)
    if (!store.classExists(form.range)) throw new Error(`终点类不存在：${form.range}`)
    if (store.relations.some((r) => r.id === name)) throw new Error(`关系已存在：${name}`)
    if (oid != null) {
      await store.createRelation(oid, {
        name,
        label: label || `${form.domain}→${form.range}`,
        from_class: form.domain,
        to_class: form.range,
        cardinality: form.cardinality,
        inverse: form.inverse || undefined,
        cross_source: form.crossSource,
      })
      store.dirty = false
    } else {
      store.addRelation({
        domain: form.domain,
        range: form.range,
        label: form.label,
        cardinality: form.cardinality,
        inverse: form.inverse || undefined,
        crossSource: form.crossSource,
      })
    }
    ElMessage.success('关系已创建')
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
    title="新建对象属性（关系）"
    width="480px"
    @update:model-value="emit('update:open', $event)"
  >
    <ElForm :model="form" label-width="88px">
      <ElFormItem label="起点类" required>
        <ElSelect v-model="form.domain" filterable placeholder="domain">
          <ElOption v-for="c in classOptions" :key="c.id" :label="c.label" :value="c.id" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="终点类" required>
        <ElSelect v-model="form.range" filterable placeholder="range">
          <ElOption v-for="c in classOptions" :key="c.id" :label="c.label" :value="c.id" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="关系名" required>
        <ElInput v-model="form.label" placeholder="如 所属客户 / 明细" />
      </ElFormItem>
      <ElFormItem label="基数">
        <ElSelect v-model="form.cardinality">
          <ElOption v-for="c in cardinalities" :key="c" :label="c" :value="c" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="反向名">
        <ElInput v-model="form.inverse" placeholder="可选，如 历史订单" />
      </ElFormItem>
      <ElFormItem label="跨源">
        <ElSwitch v-model="form.crossSource" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="emit('update:open', false)">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="onSubmit">创建</ElButton>
    </template>
  </ElDialog>
</template>
