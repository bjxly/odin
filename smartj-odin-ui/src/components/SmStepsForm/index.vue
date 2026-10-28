<template>
  <div class="sm-steps-form">
    <!-- 步骤条 -->
    <div class="sm-steps-form__steps">
      <el-steps
        :active="currentStep"
        :align-center="alignCenter"
        finish-status="success"
      >
        <el-step
          v-for="(step, idx) in steps"
          :key="idx"
          :title="step.title"
          :description="step.description"
        />
      </el-steps>
    </div>

    <!-- 表单区域 -->
    <div class="sm-steps-form__body">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="currentRules"
        :label-width="labelWidth"
      >
        <slot
          :name="`step-${currentStep}`"
          :form="formData"
          :step="currentStep"
        />
      </el-form>
    </div>

    <!-- 底部操作 -->
    <div class="sm-steps-form__footer">
      <el-button
        v-if="currentStep > 0"
        @click="prevStep"
      >
        {{ $t('cmp.prevStep') }}
      </el-button>
      <el-button
        v-if="currentStep < steps.length - 1"
        type="primary"
        @click="nextStep"
      >
        {{ $t('cmp.nextStep') }}
      </el-button>
      <el-button
        v-if="currentStep === steps.length - 1"
        type="primary"
        :loading="loading"
        @click="handleFinish"
      >
        {{ finishLabel }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'

const { t } = useI18n()

export interface StepConfig {
  title: string
  description?: string
  rules?: FormRules
}

const props = withDefaults(
  defineProps<{
    steps: StepConfig[]
    modelValue: Record<string, any>
    labelWidth?: string
    alignCenter?: boolean
    finishText?: string
    loading?: boolean
  }>(),
  {
    labelWidth: '100px',
    alignCenter: false,
    finishText: '',
    loading: false,
  }
)

const finishLabel = computed(() => props.finishText || t('cmp.finish'))

const emit = defineEmits<{
  (e: 'update:modelValue', val: Record<string, any>): void
  (e: 'finish'): void
  (e: 'step-change', step: number): void
}>()

const formRef = ref<FormInstance>()
const currentStep = ref(0)
const formData = reactive({ ...props.modelValue })

const currentRules = computed(() => {
  return props.steps[currentStep.value]?.rules || {}
})

function prevStep() {
  if (currentStep.value > 0) {
    currentStep.value--
    emit('step-change', currentStep.value)
  }
}

async function nextStep() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    currentStep.value++
    emit('step-change', currentStep.value)
    emit('update:modelValue', { ...formData })
  } catch {
    // validation failed
  }
}

async function handleFinish() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    emit('update:modelValue', { ...formData })
    emit('finish')
  } catch {
    // validation failed
  }
}
</script>

<style scoped>
.sm-steps-form {
  padding: 0;
}
.sm-steps-form__steps {
  margin-bottom: 32px;
  padding: 0 20px;
}
.sm-steps-form__body {
  min-height: 200px;
}
.sm-steps-form__footer {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 32px;
  padding-top: 16px;
  border-top: 1px solid var(--sm-border-lighter);
}
</style>
