<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmStepsForm 分步表单示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">将复杂表单拆分为多步骤，逐步引导用户填写。</p>

      <SmStepsForm
        v-model="formData"
        :steps="steps"
        label-width="100px"
        :loading="submitting"
        @finish="handleFinish"
      >
        <template #step-0="{ form }">
          <el-form-item
            label="项目名称"
            prop="name"
          >
            <el-input
              v-model="form.name"
              placeholder="请输入项目名称"
            />
          </el-form-item>
          <el-form-item
            label="项目类型"
            prop="type"
          >
            <el-select
              v-model="form.type"
              placeholder="请选择类型"
              style="width: 100%"
            >
              <el-option
                label="Web 应用"
                value="web"
              />
              <el-option
                label="移动应用"
                value="mobile"
              />
              <el-option
                label="数据服务"
                value="data"
              />
            </el-select>
          </el-form-item>
        </template>

        <template #step-1="{ form }">
          <el-form-item
            label="负责人"
            prop="owner"
          >
            <el-input
              v-model="form.owner"
              placeholder="请输入负责人"
            />
          </el-form-item>
          <el-form-item
            label="截止日期"
            prop="deadline"
          >
            <el-date-picker
              v-model="form.deadline"
              type="date"
              placeholder="请选择截止日期"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item
            label="优先级"
            prop="priority"
          >
            <el-radio-group v-model="form.priority">
              <el-radio value="high">高</el-radio>
              <el-radio value="medium">中</el-radio>
              <el-radio value="low">低</el-radio>
            </el-radio-group>
          </el-form-item>
        </template>

        <template #step-2="{ form }">
          <el-form-item
            label="项目描述"
            prop="desc"
          >
            <el-input
              v-model="form.desc"
              type="textarea"
              :rows="4"
              placeholder="请输入项目描述"
            />
          </el-form-item>
          <el-form-item label="备注">
            <el-input
              v-model="form.remark"
              type="textarea"
              :rows="2"
              placeholder="备注（选填）"
            />
          </el-form-item>
        </template>
      </SmStepsForm>
    </div>

    <div
      v-if="result"
      class="sm-card mt-4"
    >
      <h3 class="text-base font-semibold mb-3">提交结果</h3>
      <SmCodeEditor
        :model-value="result"
        language="json"
        :height="200"
        readonly
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import SmStepsForm from '@/components/SmStepsForm/index.vue'
import SmCodeEditor from '@/components/SmCodeEditor/index.vue'
import type { StepConfig } from '@/components/SmStepsForm/index.vue'

const submitting = ref(false)
const result = ref('')

const formData = reactive({
  name: '',
  type: '',
  owner: '',
  deadline: '',
  priority: 'medium',
  desc: '',
  remark: '',
})

const steps: StepConfig[] = [
  {
    title: '基本信息',
    description: '项目名称和类型',
    rules: {
      name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
      type: [{ required: true, message: '请选择项目类型', trigger: 'change' }],
    },
  },
  {
    title: '分配信息',
    description: '负责人和排期',
    rules: {
      owner: [{ required: true, message: '请输入负责人', trigger: 'blur' }],
      deadline: [{ required: true, message: '请选择截止日期', trigger: 'change' }],
    },
  },
  {
    title: '完成',
    description: '描述和备注',
    rules: {
      desc: [{ required: true, message: '请输入项目描述', trigger: 'blur' }],
    },
  },
]

function handleFinish() {
  submitting.value = true
  setTimeout(() => {
    submitting.value = false
    result.value = JSON.stringify({ ...formData }, null, 2)
    ElMessage.success('提交成功！')
  }, 1000)
}
</script>
