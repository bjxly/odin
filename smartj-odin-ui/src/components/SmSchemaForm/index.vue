<template>
  <div class="sm-schema-form">
    <!-- 工具描述 -->
    <div
      v-if="schema.description"
      class="sm-schema-form__desc"
    >
      {{ schema.description }}
    </div>

    <el-form
      ref="formRef"
      :model="engine.formData"
      :label-position="labelPosition"
      :label-width="labelWidth"
      :disabled="readonly"
    >
      <!-- 按分组渲染 -->
      <template
        v-for="group in groupedFields"
        :key="group.name"
      >
        <div
          v-if="group.name"
          class="sm-schema-form__group"
        >
          <div class="sm-schema-form__group-title">{{ group.name }}</div>
        </div>
        <el-row :gutter="gutter">
          <el-col
            v-for="field in group.fields"
            :key="field.fieldKey"
            :span="field.span || 24"
          >
            <FieldRenderer
              :field="field"
              :model-value="engine.formData[field.fieldKey]"
              :readonly="readonly || !!field.isReadonly"
              :visible="engine.isFieldVisible(field)"
              :error="engine.fieldErrors[field.fieldKey]"
              @update:model-value="engine.formData[field.fieldKey] = $event"
            >
              <template
                v-for="(_, slotName) in $slots"
                :key="slotName"
                #[slotName]="slotProps"
              >
                <slot
                  :name="slotName"
                  v-bind="slotProps"
                />
              </template>
            </FieldRenderer>
          </el-col>
        </el-row>
      </template>

      <!-- 操作按钮 -->
      <div
        v-if="showSubmit || showReset"
        class="sm-schema-form__actions"
      >
        <slot name="submit-button">
          <el-button
            v-if="showSubmit"
            type="primary"
            @click="handleSubmit"
          >
            {{ submitLabel }}
          </el-button>
        </slot>
        <el-button
          v-if="showReset"
          @click="handleReset"
        >
          {{ $t('common.reset') }}
        </el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance } from 'element-plus'
import type { FormSchema, FormField } from '@/types/schema'
import { useFormEngine } from './useFormEngine'
import FieldRenderer from './FieldRenderer.vue'
import type { ValidationError } from './types'

const props = withDefaults(
  defineProps<{
    schema: FormSchema
    modelValue?: Record<string, any>
    readonly?: boolean
    showSubmit?: boolean
    submitText?: string
    showReset?: boolean
    labelPosition?: 'left' | 'right' | 'top'
    gutter?: number
    labelWidth?: string
  }>(),
  {
    readonly: false,
    showSubmit: true,
    submitText: '',
    showReset: true,
    labelPosition: 'right',
    gutter: 16,
    labelWidth: '100px',
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', data: Record<string, any>): void
  (e: 'submit', data: Record<string, any>): void
  (e: 'reset'): void
  (e: 'validate-fail', errors: ValidationError[]): void
}>()

const { t } = useI18n()

const formRef = ref<FormInstance>()

const submitLabel = computed(() => props.submitText || t('cmp.schemaSubmit'))

const engine = useFormEngine(
  computed(() => props.schema),
  props.modelValue
)

// 监听modelValue变化
watch(
  () => props.modelValue,
  (val) => {
    if (val) {
      engine.setFormData(val)
    }
  },
  { deep: true }
)

// 按分组整理字段
const groupedFields = computed(() => {
  const fields = (props.schema.fields || []).filter((f) => (f.fieldLevel || 0) === 0).sort((a, b) => (a.displayOrder || 0) - (b.displayOrder || 0))

  const groups: { name: string; fields: FormField[] }[] = []
  const groupMap = new Map<string, FormField[]>()

  for (const field of fields) {
    const groupName = field.fieldGroup || ''
    if (!groupMap.has(groupName)) {
      groupMap.set(groupName, [])
    }
    groupMap.get(groupName)!.push(field)
  }

  groupMap.forEach((fields, name) => {
    groups.push({ name, fields })
  })

  return groups
})

// 提交
async function handleSubmit() {
  const { valid, errors } = await engine.validate()
  if (valid) {
    const data = engine.getFormData()
    emit('submit', data)
    emit('update:modelValue', data)
  } else {
    emit('validate-fail', errors)
  }
}

// 重置
function handleReset() {
  engine.resetForm()
  emit('reset')
}

// 暴露方法
defineExpose({
  validate: engine.validate,
  getFormData: engine.getFormData,
  setFormData: engine.setFormData,
  resetForm: engine.resetForm,
  setFieldReadonly: engine.setFieldReadonly,
  setFieldVisible: engine.setFieldVisible,
})
</script>

<style scoped>
.sm-schema-form {
  width: 100%;
}
.sm-schema-form__desc {
  margin-bottom: 16px;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  line-height: 1.6;
}
.sm-schema-form__group {
  margin-bottom: 8px;
  margin-top: 16px;
}
.sm-schema-form__group-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  padding-left: 8px;
  border-left: 3px solid var(--el-color-primary);
}
.sm-schema-form__actions {
  margin-top: 24px;
  text-align: center;
}
</style>
