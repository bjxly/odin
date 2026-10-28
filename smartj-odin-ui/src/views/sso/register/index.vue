<template>
  <div class="sso-page">
    <div
      class="sso-card"
      style="max-width: 460px; margin: 60px auto"
    >
      <div class="sso-card-header">
        <h2>{{ $t('sso.registerHeading') }}</h2>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        size="large"
        label-width="80px"
      >
        <el-form-item
          :label="$t('sso.username')"
          prop="username"
        >
          <el-input
            v-model="form.username"
            :placeholder="$t('sso.enterUsername')"
            :maxlength="50"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.nickname')"
          prop="nickname"
        >
          <el-input
            v-model="form.nickname"
            :placeholder="$t('sso.enterNickname')"
            :maxlength="50"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.password')"
          prop="password"
        >
          <el-input
            v-model="form.password"
            type="password"
            :placeholder="$t('sso.enterPassword')"
            show-password
            :maxlength="128"
            :minlength="8"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.confirmPassword')"
          prop="confirmPassword"
        >
          <el-input
            v-model="form.confirmPassword"
            type="password"
            :placeholder="$t('sso.enterConfirmPw')"
            show-password
            :maxlength="50"
            :minlength="8"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.mobile')"
          prop="mobile"
        >
          <el-input
            v-model="form.mobile"
            :placeholder="$t('sso.enterMobile')"
            :maxlength="11"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sso.email')"
          prop="email"
        >
          <el-input
            v-model="form.email"
            :placeholder="$t('sso.enterEmail')"
            :maxlength="100"
            show-word-limit="true"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            style="width: 100%"
            @click="handleSubmit"
          >
            {{ $t('sso.submitRegister') }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="sso-card-footer">
        {{ $t('sso.hasAccount') }}
        <el-link
          type="primary"
          @click="router.push('/login')"
        >
          {{ $t('sso.loginNow') }}
        </el-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { submitRegistration } from '@/api/sso'
import { usePasswordMinLength } from '@/composables/usePasswordMinLength'
import { usePasswordComplexity } from '@/composables/usePasswordComplexity'

const router = useRouter()
const { t } = useI18n()
const formRef = ref<FormInstance>()
const loading = ref(false)
const passwordMinLength = usePasswordMinLength()
const passwordComplexityRequired = usePasswordComplexity()

const form = reactive({
  username: '',
  nickname: '',
  password: '',
  confirmPassword: '',
  mobile: '',
  email: '',
})

const validateConfirm = (_rule: any, value: string, callback: any) => {
  if (value !== form.password) {
    callback(new Error(t('sso.pwMismatch')))
  } else {
    callback()
  }
}

/** 密码复杂度校验：开关开启时要求同时包含大写、小写、数字和特殊字符 */
const validateComplexity = (_rule: any, value: string, callback: any) => {
  if (!passwordComplexityRequired.value || !value) {
    callback()
    return
  }
  const ok = /[a-z]/.test(value) && /[A-Z]/.test(value) && /\d/.test(value) && /[^a-zA-Z0-9]/.test(value)
  callback(ok ? undefined : new Error(t('sso.pwComplexity')))
}

const rules = computed<FormRules>(() => ({
  username: [
    { required: true, message: t('sso.enterUsername'), trigger: 'blur' },
    { min: 3, max: 20, message: t('sso.usernameLenHint'), trigger: 'blur' },
  ],
  nickname: [{ required: true, message: t('sso.enterNickname'), trigger: 'blur' }],
  password: [
    { required: true, message: t('sso.enterPassword'), trigger: 'blur' },
    { min: passwordMinLength.value, message: t('sso.pwMinLength', { min: passwordMinLength.value }), trigger: 'blur' },
    { validator: validateComplexity, trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: t('sso.enterConfirmPw'), trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' },
  ],
  mobile: [
    { required: true, message: t('sso.enterMobile'), trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: t('sso.phoneFormatInvalid'), trigger: 'blur' },
  ],
}))

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await submitRegistration({
      username: form.username,
      password: form.password,
      nickname: form.nickname,
      mobile: form.mobile,
      email: form.email,
    })
    ElMessage.success(t('sso.registerSuccess'))
    router.push('/sso/login')
  } catch (e: any) {
    ElMessage.error(e?.message || t('sso.registerFailed'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.sso-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding-top: 60px;
}
.sso-card {
  background: #fff;
  border-radius: 8px;
  padding: 32px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}
.sso-card-header {
  text-align: center;
  margin-bottom: 24px;
}
.sso-card-header h2 {
  font-size: 20px;
  color: #303133;
}
.sso-card-footer {
  text-align: center;
  margin-top: 16px;
  color: #909399;
}
</style>
