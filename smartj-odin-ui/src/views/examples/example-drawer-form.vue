<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmDrawerForm 抽屉表单示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">侧边滑出的抽屉表单，适合宽表单和便捷编辑场景。</p>

      <div class="flex gap-3">
        <el-button
          type="primary"
          @click="openDrawer('add')"
        >
          新增用户
        </el-button>
        <el-button @click="openDrawer('edit')">编辑用户（带数据）</el-button>
      </div>
    </div>

    <div class="sm-card mt-4">
      <h3 class="text-base font-semibold mb-3">当前数据</h3>
      <SmCodeEditor
        :model-value="JSON.stringify(userData, null, 2)"
        language="json"
        :height="150"
        readonly
      />
    </div>

    <SmDrawerForm
      v-model:visible="drawerVisible"
      :title="drawerTitle"
      :loading="saving"
      @confirm="handleSave"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
      >
        <el-form-item
          label="用户名"
          prop="username"
        >
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
          />
        </el-form-item>
        <el-form-item
          label="昵称"
          prop="nickname"
        >
          <el-input
            v-model="form.nickname"
            placeholder="请输入昵称"
          />
        </el-form-item>
        <el-form-item
          label="邮箱"
          prop="email"
        >
          <el-input
            v-model="form.email"
            placeholder="请输入邮箱"
          />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input
            v-model="form.mobile"
            placeholder="请输入手机号"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="3"
            placeholder="备注信息"
          />
        </el-form-item>
      </el-form>
    </SmDrawerForm>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import SmDrawerForm from '@/components/SmDrawerForm/index.vue'
import SmCodeEditor from '@/components/SmCodeEditor/index.vue'

const drawerVisible = ref(false)
const drawerTitle = ref('')
const saving = ref(false)
const formRef = ref<FormInstance>()
const mode = ref<'add' | 'edit'>('add')

const form = reactive({
  username: '',
  nickname: '',
  email: '',
  mobile: '',
  status: 1,
  remark: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱', trigger: 'blur' }],
}

const userData = reactive({})

function openDrawer(type: 'add' | 'edit') {
  mode.value = type
  drawerTitle.value = type === 'add' ? '新增用户' : '编辑用户'
  if (type === 'add') {
    Object.assign(form, { username: '', nickname: '', email: '', mobile: '', status: 1, remark: '' })
  } else {
    Object.assign(form, { username: 'zhangsan', nickname: '张三', email: 'zhangsan@smartj.com', mobile: '13800138000', status: 1, remark: '示例数据' })
  }
  drawerVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  setTimeout(() => {
    saving.value = false
    drawerVisible.value = false
    Object.assign(userData, { ...form })
    ElMessage.success(mode.value === 'add' ? '新增成功' : '更新成功')
  }, 800)
}
</script>
