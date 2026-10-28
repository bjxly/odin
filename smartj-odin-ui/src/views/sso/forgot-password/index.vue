<template>
  <div class="sso-page">
    <div
      class="sso-card"
      style="max-width: 460px; margin: 60px auto"
    >
      <div class="sso-card-header">
        <h2>{{ $t('sso.forgotPasswordHeading') }}</h2>
      </div>

      <el-steps
        :active="step"
        align-center
        class="sso-steps"
      >
        <el-step :title="$t('sso.verifyIdentity')" />
        <el-step :title="$t('sso.resetPassword')" />
        <el-step :title="$t('sso.complete')" />
      </el-steps>

      <!-- Step 1: 验证身份 -->
      <el-form
        v-if="step === 0"
        ref="verifyFormRef"
        :model="verifyForm"
        :rules="verifyRules"
        size="large"
      >
        <el-form-item prop="phone">
          <el-input
            v-model="verifyForm.phone"
            :placeholder="$t('sso.enterPhone')"
            prefix-icon="Phone"
            :maxlength="11"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          v-if="countdown <= 0"
          prop="graphicCaptcha"
        >
          <div class="sso-captcha-row">
            <el-input
              v-model="verifyForm.graphicCaptcha"
              :placeholder="$t('sso.enterGraphicCaptcha')"
              prefix-icon="Key"
              class="sso-captcha-input"
            />
            <img
              v-if="captchaImage"
              :src="captchaImage"
              :alt="$t('sso.enterGraphicCaptcha')"
              class="sso-captcha-img"
              @click="refreshCaptcha()"
            />
            <div
              v-else
              class="sso-captcha-placeholder"
              @click="refreshCaptcha()"
            >
              {{ $t('auth.loading') }}
            </div>
          </div>
        </el-form-item>
        <el-form-item prop="smsCode">
          <div class="sso-sms-row">
            <el-input
              v-model="verifyForm.smsCode"
              :placeholder="$t('sso.enterSmsCode')"
            />
            <el-button
              :disabled="countdown > 0 || !verifyForm.phone"
              @click="handleSendSms"
            >
              {{ countdown > 0 ? `${countdown}s` : $t('sso.getSmsCode') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="verifyLoading"
            style="width: 100%"
            @click="handleVerify"
          >
            {{ $t('sso.nextStep') }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- Step 2: 重置密码 -->
      <el-form
        v-if="step === 1"
        ref="resetFormRef"
        :model="resetForm"
        :rules="resetRules"
        size="large"
      >
        <el-form-item prop="newPassword">
          <el-input
            v-model="resetForm.newPassword"
            type="password"
            :placeholder="$t('sso.inputNewPw')"
            show-password
          />
        </el-form-item>
        <el-form-item prop="confirmPassword">
          <el-input
            v-model="resetForm.confirmPassword"
            type="password"
            :placeholder="$t('sso.confirmNewPw')"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="resetLoading"
            style="width: 100%"
            @click="handleReset"
          >
            {{ $t('sso.resetPw') }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- Step 3: 完成 -->
      <el-result
        v-if="step === 2"
        icon="success"
        :title="$t('sso.resetPwSuccess')"
        :sub-title="$t('sso.resetPwSuccessHint')"
      >
        <template #extra>
          <el-button
            type="primary"
            @click="router.push('/login')"
          >
            {{ $t('sso.backToLogin') }}
          </el-button>
        </template>
      </el-result>

      <div
        v-if="step < 2"
        class="sso-card-footer"
      >
        <el-link
          type="primary"
          @click="router.push('/login')"
        >
          {{ $t('sso.backToLogin') }}
        </el-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { sendSmsCode, forgotPassword, getSsoCaptcha } from '@/api/sso'
import { usePasswordMinLength } from '@/composables/usePasswordMinLength'
import { usePasswordComplexity } from '@/composables/usePasswordComplexity'

const router = useRouter()
const { t } = useI18n()
const step = ref(0)
const passwordMinLength = usePasswordMinLength()
const passwordComplexityRequired = usePasswordComplexity()

// Step 1: 验证
const verifyFormRef = ref<FormInstance>()
const verifyForm = reactive({ phone: '', smsCode: '', graphicCaptcha: '' })
const verifyLoading = ref(false)
const countdown = ref(0)
const captchaId = ref('')
const captchaImage = ref('')
let timer: ReturnType<typeof setInterval> | null = null
const verifyRules = computed<FormRules>(() => ({
  phone: [
    { required: true, message: t('sso.enterPhone'), trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: t('sso.phoneFormatInvalid'), trigger: 'blur' },
  ],
  // 倒计时期间图形验证码已消费，无需校验；倒计时结束后恢复必填
  graphicCaptcha: [{ required: countdown.value <= 0, message: t('sso.enterGraphicCaptcha'), trigger: 'blur' }],
  smsCode: [{ required: true, message: t('sso.enterSmsCode'), trigger: 'blur' }],
}))

async function refreshCaptcha() {
  try {
    const data = await getSsoCaptcha()
    captchaId.value = data.captchaId
    captchaImage.value = data.image
  } catch {
    // ignore
  }
}

onMounted(refreshCaptcha)

async function handleSendSms() {
  const valid = await verifyFormRef.value?.validateField(['phone', 'graphicCaptcha']).catch(() => false)
  if (!valid) return
  try {
    await sendSmsCode({
      phone: verifyForm.phone,
      scope: 'RESET_PWD',
      captchaId: captchaId.value,
      captchaCode: verifyForm.graphicCaptcha,
    })
    ElMessage.success(t('sso.smsCodeSent'))
    verifyForm.graphicCaptcha = ''
    refreshCaptcha()
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0 && timer) clearInterval(timer)
    }, 1000)
  } catch (e: any) {
    verifyForm.graphicCaptcha = ''
    refreshCaptcha()
    ElMessage.error(e?.message || t('sso.sendFailed'))
  }
}

async function handleVerify() {
  // 仅校验手机号和短信验证码，图形验证码在发送短信时已消费
  const valid = await verifyFormRef.value?.validateField(['phone', 'smsCode']).catch(() => false)
  if (!valid) return
  step.value = 1
}

// Step 2: 重置
const resetFormRef = ref<FormInstance>()
const resetForm = reactive({ newPassword: '', confirmPassword: '' })
const resetLoading = ref(false)
const validateConfirm = (_: any, v: string, cb: any) => {
  cb(v !== resetForm.newPassword ? new Error(t('sso.pwMismatch')) : undefined)
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
const resetRules = computed<FormRules>(() => ({
  newPassword: [
    { required: true, message: t('sso.inputNewPw'), trigger: 'blur' },
    { min: passwordMinLength.value, message: t('sso.pwMinLength', { min: passwordMinLength.value }), trigger: 'blur' },
    { validator: validateComplexity, trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: t('sso.confirmNewPw'), trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' },
  ],
}))

async function handleReset() {
  const valid = await resetFormRef.value?.validate().catch(() => false)
  if (!valid) return
  resetLoading.value = true
  try {
    await forgotPassword({
      phone: verifyForm.phone,
      smsCode: verifyForm.smsCode,
      newPassword: resetForm.newPassword,
    })
    step.value = 2
  } catch (e: any) {
    ElMessage.error(e?.message || t('sso.resetFailed'))
  } finally {
    resetLoading.value = false
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
.sso-steps {
  margin-bottom: 32px;
}
.sso-sms-row {
  display: flex;
  gap: 8px;
  width: 100%;
}
.sso-captcha-row {
  display: flex;
  width: 100%;
  gap: 12px;
}
.sso-captcha-input {
  flex: 1;
}
.sso-captcha-img {
  width: 120px;
  height: 40px;
  cursor: pointer;
  border-radius: 6px;
  border: 1px solid #dcdfe6;
  object-fit: contain;
  background: #fff;
  flex-shrink: 0;
  transition: transform 0.2s ease;
}
.sso-captcha-img:hover {
  transform: scale(1.05);
}
.sso-captcha-placeholder {
  width: 120px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 6px;
  border: 1px solid #dcdfe6;
  background: transparent;
  color: #909399;
  font-size: 12px;
  flex-shrink: 0;
}
.sso-card-footer {
  text-align: center;
  margin-top: 16px;
}
</style>
