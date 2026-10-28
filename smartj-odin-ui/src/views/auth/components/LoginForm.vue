<template>
  <div class="login-form">
    <!-- 标题（可选） -->
    <div
      v-if="title"
      class="login-form__header"
    >
      <h2 class="login-form__title">{{ title }}</h2>
    </div>

    <!-- 认证方式切换 Tab（多种登录方式时才显示） -->
    <div
      v-if="visibleTabs.length > 1"
      class="login-form__tabs"
    >
      <div
        v-for="tab in visibleTabs"
        :key="tab.key"
        :class="['login-form__tab', { active: activeTab === tab.key }]"
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
      </div>
    </div>

    <!-- ========== 密码登录 ========== -->
    <el-form
      v-show="activeTab === 'password'"
      ref="pwdFormRef"
      :model="pwdForm"
      :rules="pwdRules"
      size="large"
      @keyup.enter="handlePasswordLogin"
    >
      <el-form-item prop="username">
        <el-input
          v-model="pwdForm.username"
          :placeholder="$t('auth.usernamePlaceholder')"
          prefix-icon="User"
        />
      </el-form-item>

      <el-form-item prop="password">
        <el-input
          v-model="pwdForm.password"
          type="password"
          :placeholder="$t('auth.passwordPlaceholder')"
          prefix-icon="Lock"
          show-password
        />
      </el-form-item>

      <el-form-item prop="captchaCode">
        <div class="login-form__captcha-row">
          <el-input
            v-model="pwdForm.captchaCode"
            :placeholder="$t('auth.captchaPlaceholder')"
            prefix-icon="Key"
            class="login-form__captcha-input"
          />
          <img
            v-if="captchaImage"
            :src="captchaImage"
            :alt="$t('auth.captcha')"
            class="login-form__captcha-img"
            @click="refreshCaptcha"
          />
          <div
            v-else
            class="login-form__captcha-placeholder"
            @click="refreshCaptcha"
          >
            {{ $t('auth.loading') }}
          </div>
        </div>
      </el-form-item>

      <el-form-item v-if="isConfigEnabled(ssoConfig, 'sso.features.remember-me-enabled')">
        <el-checkbox v-model="pwdForm.rememberMe">{{ $t('auth.rememberMe') }}</el-checkbox>
      </el-form-item>

      <el-form-item>
        <el-button
          type="primary"
          :loading="pwdLoading"
          class="login-form__btn"
          @click="handlePasswordLogin"
        >
          {{ $t('auth.login') }}
        </el-button>
      </el-form-item>
    </el-form>

    <!-- ========== 短信登录 ========== -->
    <el-form
      v-show="activeTab === 'sms'"
      ref="smsFormRef"
      :model="smsForm"
      :rules="smsRules"
      size="large"
    >
      <el-form-item prop="phone">
        <el-input
          v-model="smsForm.phone"
          :placeholder="$t('auth.phonePlaceholder')"
          prefix-icon="Iphone"
        />
      </el-form-item>

      <el-form-item
        v-if="smsCountdown <= 0"
        prop="smsCaptchaCode"
      >
        <div class="login-form__captcha-row">
          <el-input
            v-model="smsForm.smsCaptchaCode"
            :placeholder="$t('auth.graphicCaptchaPlaceholder')"
            prefix-icon="Key"
            class="login-form__captcha-input"
          />
          <img
            v-if="smsCaptchaImage"
            :src="smsCaptchaImage"
            :alt="$t('auth.captcha')"
            class="login-form__captcha-img"
            @click="refreshSmsCaptcha()"
          />
          <div
            v-else
            class="login-form__captcha-placeholder"
            @click="refreshSmsCaptcha()"
          >
            {{ $t('auth.loading') }}
          </div>
        </div>
      </el-form-item>

      <el-form-item prop="smsCode">
        <div class="login-form__captcha-row">
          <el-input
            v-model="smsForm.smsCode"
            :placeholder="$t('auth.smsCodePlaceholder')"
            prefix-icon="Message"
            class="login-form__captcha-input"
            maxlength="6"
          />
          <el-button
            :disabled="smsCountdown > 0 || !canSendSms"
            class="login-form__sms-btn"
            @click="handleSendSms"
          >
            {{ smsCountdown > 0 ? `${smsCountdown}s` : $t('auth.sendSms') }}
          </el-button>
        </div>
      </el-form-item>

      <el-form-item>
        <el-button
          type="primary"
          :loading="smsLoading"
          class="login-form__btn"
          @click="handleSmsLogin"
        >
          {{ $t('auth.login') }}
        </el-button>
      </el-form-item>
    </el-form>

    <!-- ========== 扫码登录 ========== -->
    <div
      v-show="activeTab === 'qr'"
      class="login-form__qr"
    >
      <div class="login-form__qr-box">
        <el-icon
          :size="64"
          class="login-form__qr-icon"
        >
          <Picture />
        </el-icon>
        <p class="login-form__qr-text">{{ $t('auth.qrScanTip') }}</p>
        <p class="login-form__qr-hint">{{ $t('auth.qrDeveloping') }}</p>
      </div>
    </div>

    <!-- ========== 底部链接 & 社交登录 ========== -->
    <div class="login-form__footer">
      <div
        v-if="showRegistration || showForgotPassword"
        class="login-form__footer-links"
      >
        <span
          v-if="showRegistration"
          class="login-form__footer-link"
          @click="goRegister"
        >
          {{ $t('auth.register') }}
        </span>
        <span
          v-if="showRegistration && showForgotPassword"
          class="login-form__footer-divider"
        >
          |
        </span>
        <span
          v-if="showForgotPassword"
          class="login-form__footer-link"
          @click="goForgot"
        >
          {{ $t('auth.forgotPassword') }}
        </span>
      </div>

      <div
        v-if="isConfigEnabled(ssoConfig, 'sso.features.social-login-enabled') && idps.length"
        class="login-form__social"
      >
        <div class="login-form__social-title">{{ $t('auth.otherLogin') }}</div>
        <div class="login-form__social-icons">
          <div
            v-for="idp in idps"
            :key="idp.id"
            class="login-form__social-icon"
            :title="idp.idpName"
            @click="handleSocialLogin(idp.idpType)"
          >
            <span
              v-if="socialIconSvgMap[idp.idpType]"
              v-html="socialIconSvgMap[idp.idpType]"
            />
            <span
              v-else
              class="login-form__social-text-icon"
            >
              {{ idp.idpName.charAt(0) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'
import { getSsoCaptcha, ssoPasswordLogin, ssoSmsLogin, sendSmsCode, getSsoSm2PublicKey, getSsoPublicIdps } from '@/api/sso'
import { getLocalCaptcha, localLogin, getLocalPublicKey } from '@/api/auth'
import { fetchPublicConfig } from '@/composables/usePublicConfig'
import { isConfigEnabled } from '@/utils/config'
import { SSO_ENABLED } from '@/config/sso'
import type { SsoAuthLoginResponse, SsoPublicConfig, SsoSm2PublicKey, SsoIdpPublicInfo } from '@/types/sso'

defineProps<{
  title?: string
}>()

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const userStore = useUserStore()
const authStore = useAuthStore()

// ==================== SSO 公开配置 ====================
const ssoConfig = ref<SsoPublicConfig | null>(null)
const sm2Config = ref<SsoSm2PublicKey | null>(null)
const idps = ref<SsoIdpPublicInfo[]>([])

type TabKey = 'password' | 'sms' | 'qr'
interface TabItem {
  key: TabKey
  label: string
}

const visibleTabs = computed<TabItem[]>(() => {
  const tabs: TabItem[] = [{ key: 'password', label: t('auth.passwordLogin') }]
  if (isConfigEnabled(ssoConfig.value, 'sso.features.sms-login-enabled')) {
    tabs.push({ key: 'sms', label: t('auth.smsLogin') })
  }
  if (isConfigEnabled(ssoConfig.value, 'sso.features.qr-code-login-enabled')) {
    tabs.push({ key: 'qr', label: t('auth.qrLogin') })
  }
  return tabs
})

const activeTab = ref<TabKey>('password')

/** 是否显示注册链接（受后端配置控制） */
const showRegistration = computed(() => isConfigEnabled(ssoConfig.value, 'sso.features.registration-enabled'))
/** 是否显示忘记密码链接（受后端配置控制） */
const showForgotPassword = computed(() => isConfigEnabled(ssoConfig.value, 'sso.features.forgot-password-enabled'))

watch(visibleTabs, (tabs) => {
  if (!tabs.find((t) => t.key === activeTab.value)) {
    activeTab.value = tabs[0]?.key ?? 'password'
  }
  if (activeTab.value === 'sms' && !smsCaptchaImage.value) {
    refreshSmsCaptcha()
  }
})

function switchTab(key: TabKey) {
  activeTab.value = key
  if (key === 'sms' && !smsCaptchaImage.value) {
    refreshSmsCaptcha()
  }
}

// ==================== 社交登录图标映射 ====================
const socialIconSvgMap: Record<string, string> = {
  WECHAT:
    '<svg viewBox="0 0 24 24" class="login-form__social-svg"><path fill="currentColor" d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1-.023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.952-7.062-6.122zm-2.18 2.769c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982z"/></svg>',
  WECOM:
    '<svg viewBox="0 0 24 24" class="login-form__social-svg"><rect x="3" y="3" width="8" height="8" rx="1.5" fill="currentColor"/><rect x="13" y="13" width="8" height="8" rx="1.5" fill="currentColor" opacity="0.6"/><rect x="13" y="3" width="8" height="8" rx="1.5" fill="currentColor" opacity="0.4"/></svg>',
  DINGTALK:
    '<svg viewBox="0 0 24 24" class="login-form__social-svg"><path fill="currentColor" d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm4.49 7.617l-3.205 2.56a.194.194 0 0 0 0 .307l3.205 2.561c.142.114.036.345-.14.345h-1.902a.389.389 0 0 1-.242-.09l-2.573-2.057a.194.194 0 0 0-.306.146v3.472c0 .108-.087.195-.195.195H9.368a.195.195 0 0 1-.195-.195V7.39c0-.108.087-.195.195-.195h1.764c.108 0 .195.087.195.195v3.472c0 .166.186.263.324.166l2.554-2.077a.389.389 0 0 1 .242-.09h1.902c.177 0 .283.23.141.345z"/></svg>',
  FEISHU:
    '<svg viewBox="0 0 24 24" class="login-form__social-svg"><path fill="currentColor" d="M16.75 2c-1.02 0-1.97.29-2.77.82L6.44 7.91c-1.02.68-1.61 1.8-1.61 3.01v7.56c0 .83.67 1.5 1.5 1.5h1.77c.83 0 1.5-.67 1.5-1.5v-5.44l6.19-3.72c.54-.33 1.18-.32 1.71.02.53.34.82.93.78 1.57l-.24 3.89c-.03.43.31.8.75.8h1.77c.45 0 .81-.39.78-.84l-.23-4.14c-.06-1.14-.31-2.26-.73-3.31C19.99 3.04 17.98 2 16.75 2z"/></svg>',
}

// ==================== 验证码 ====================
const captchaImage = ref('')
const captchaId = ref('')

async function refreshCaptcha() {
  try {
    const data = SSO_ENABLED ? await getSsoCaptcha() : await getLocalCaptcha()
    captchaId.value = data.captchaId
    captchaImage.value = data.image
  } catch {
    // ignore
  }
}

// ==================== 密码登录 ====================
const pwdFormRef = ref<FormInstance>()
const pwdLoading = ref(false)
const pwdForm = reactive({
  username: '',
  password: '',
  captchaCode: '',
  rememberMe: true,
})
const pwdRules = computed<FormRules>(() => ({
  username: [{ required: true, message: t('auth.usernamePlaceholder'), trigger: 'blur' }],
  password: [{ required: true, message: t('auth.passwordPlaceholder'), trigger: 'blur' }],
  captchaCode: [{ required: true, message: t('auth.captchaPlaceholder'), trigger: 'blur' }],
}))

async function handlePasswordLogin() {
  if (!pwdFormRef.value) return
  const valid = await pwdFormRef.value.validate().catch(() => false)
  if (!valid) return

  pwdLoading.value = true
  try {
    const loginData = {
      username: pwdForm.username,
      password: pwdForm.password,
      captchaId: captchaId.value,
      captchaCode: pwdForm.captchaCode,
      rememberMe: pwdForm.rememberMe,
      authMethod: 'PASSWORD',
    }
    const sm2Opts = sm2Config.value
      ? {
          enabled: sm2Config.value.enabled === 'true',
          publicKey: sm2Config.value.sm2PublicKey,
        }
      : undefined
    const res = SSO_ENABLED
      ? await ssoPasswordLogin(loginData, sm2Opts)
      : await localLogin(loginData, sm2Opts)
    await handleLoginSuccess(res)
  } catch {
    refreshCaptcha()
    pwdForm.captchaCode = ''
  } finally {
    pwdLoading.value = false
  }
}

// ==================== 短信登录 ====================
const smsFormRef = ref<FormInstance>()
const smsLoading = ref(false)
const smsCountdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null
const smsCaptchaImage = ref('')
const smsCaptchaId = ref('')
const smsForm = reactive({
  phone: '',
  smsCaptchaCode: '',
  smsCode: '',
})

// 判断是否可以发送短信（手机号和图形验证码都已填写）
const canSendSms = computed(() => {
  return /^1[3-9]\d{9}$/.test(smsForm.phone) && smsForm.smsCaptchaCode.trim().length > 0
})
const smsRules = computed<FormRules>(() => ({
  phone: [
    { required: true, message: t('auth.phonePlaceholder'), trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: t('auth.phoneInvalid'), trigger: 'blur' },
  ],
  // 倒计时期间图形验证码已消费，无需校验；倒计时结束后恢复必填
  smsCaptchaCode: [{ required: smsCountdown.value <= 0, message: t('auth.graphicCaptchaPlaceholder'), trigger: 'blur' }],
  smsCode: [{ required: true, message: t('auth.smsCodePlaceholder'), trigger: 'blur' }],
}))

async function refreshSmsCaptcha() {
  try {
    const data = await getSsoCaptcha()
    smsCaptchaId.value = data.captchaId
    smsCaptchaImage.value = data.image
  } catch {
    // ignore
  }
}

async function handleSendSms() {
  if (!smsFormRef.value) return
  const valid = await smsFormRef.value.validateField(['phone', 'smsCaptchaCode']).catch(() => false)
  if (!valid) return
  try {
    await sendSmsCode({
      phone: smsForm.phone,
      scope: 'PRIMARY_LOGIN',
      captchaId: smsCaptchaId.value,
      captchaCode: smsForm.smsCaptchaCode,
    })
    ElMessage.success(t('auth.smsSent'))
    smsForm.smsCaptchaCode = ''
    refreshSmsCaptcha()
    smsCountdown.value = 60
    countdownTimer = setInterval(() => {
      smsCountdown.value--
      if (smsCountdown.value <= 0 && countdownTimer) {
        clearInterval(countdownTimer)
        countdownTimer = null
      }
    }, 1000)
  } catch {
    smsForm.smsCaptchaCode = ''
    refreshSmsCaptcha()
  }
}

async function handleSmsLogin() {
  if (!smsFormRef.value) return
  // 短信登录只需校验手机号和短信验证码，图形验证码在发送短信时已消费
  const valid = await smsFormRef.value.validateField(['phone', 'smsCode']).catch(() => false)
  if (!valid) return

  smsLoading.value = true
  try {
    const res = await ssoSmsLogin({
      phone: smsForm.phone,
      smsCode: smsForm.smsCode,
      rememberMe: true,
    })
    await handleLoginSuccess(res)
  } finally {
    smsLoading.value = false
  }
}

// ==================== 登录成功处理 ====================
async function handleLoginSuccess(res: SsoAuthLoginResponse) {
  userStore.setToken(res.token, res.refreshToken)
  // 持久化当前租户 ID（与 token 相同的 remember-me 存储策略）
  authStore.setTenantId(String(res.user.tenantId))
  userStore.setUserInfo({
    id: String(res.user.id),
    username: res.user.username,
    nickname: res.user.nickname,
    tenantId: String(res.user.tenantId),
    tenantName: res.user.tenantName,
    permissions: res.user.permissions,
  })

  // 密码过期提示：正常进入首页，仅弹出警告提醒用户修改密码
  if (res.passwordExpired) {
    ElMessage.warning(t('auth.passwordExpiredWarn', { days: res.passwordDaysSinceChange ?? 0 }))
  } else {
    ElMessage.success(t('auth.loginSuccess'))
  }
  const redirect = (route.query.redirect as string) || '/'
  router.push(redirect)
}

// ==================== 社交登录 ====================
function handleSocialLogin(provider: string) {
  window.location.href = `/sso/social/${provider.toLowerCase()}`
}

// ==================== 跳转 ====================
function goRegister() {
  router.push('/sso/register')
}
function goForgot() {
  router.push('/sso/forgot-password')
}

// ==================== 生命周期 ====================
onMounted(() => {
  refreshCaptcha()
  if (SSO_ENABLED) {
    fetchPublicConfig().then((cfg) => {
      ssoConfig.value = cfg
    })
    getSsoSm2PublicKey().then((cfg) => {
      sm2Config.value = cfg
    })
    getSsoPublicIdps().then((list) => {
      idps.value = list
    })
  } else {
    // 无 SSO 模式：使用本地默认配置（功能精简），SM2 走主应用端点
    fetchPublicConfig().then((cfg) => {
      ssoConfig.value = cfg
    })
    getLocalPublicKey().then((cfg) => {
      sm2Config.value = cfg
    })
  }
})

onBeforeUnmount(() => {
  if (countdownTimer) clearInterval(countdownTimer)
})
</script>

<style scoped>
.login-form__header {
  text-align: center;
  margin-bottom: 24px;
}

.login-form__title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  letter-spacing: 0.5px;
}

/* ====== Tab ====== */
.login-form__tabs {
  display: flex;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--login-tab-border, rgba(0, 0, 0, 0.08));
}

.login-form__tab {
  flex: 1;
  text-align: center;
  padding: 8px 0 10px;
  font-size: 14px;
  color: var(--login-tab-color, #909399);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition:
    color 0.25s,
    border-color 0.25s;
}

.login-form__tab:hover {
  color: var(--login-tab-active, #409eff);
}

.login-form__tab.active {
  color: var(--login-tab-active, #409eff);
  border-bottom-color: var(--login-tab-active, #409eff);
}

/* ====== 验证码行 ====== */
.login-form__captcha-row {
  display: flex;
  width: 100%;
  gap: 12px;
}

.login-form__captcha-input {
  flex: 1;
}

.login-form__captcha-img {
  width: 120px;
  height: 40px;
  cursor: pointer;
  border-radius: 6px;
  border: 1px solid var(--login-border, #dcdfe6);
  object-fit: contain;
  background: #fff;
  transition: transform 0.2s ease;
}

.login-form__captcha-img:hover {
  transform: scale(1.05);
}

.login-form__captcha-placeholder {
  width: 120px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 6px;
  border: 1px solid var(--login-border, #dcdfe6);
  background: var(--login-input-bg, transparent);
  color: var(--login-text-secondary, #909399);
  font-size: 12px;
}

/* SMS 发送按钮 */
.login-form__sms-btn {
  width: 120px;
  height: 40px;
  flex-shrink: 0;
  font-size: 13px;
}

/* ====== 登录按钮 ====== */
.login-form__btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 2px;
}

/* ====== 扫码登录 ====== */
.login-form__qr {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 24px 0;
  min-height: 200px;
}

.login-form__qr-box {
  text-align: center;
}

.login-form__qr-icon {
  color: var(--login-text-secondary, #909399);
  margin-bottom: 12px;
}

.login-form__qr-text {
  font-size: 14px;
  color: var(--login-text-secondary, #909399);
  margin: 0 0 6px;
}

.login-form__qr-hint {
  font-size: 12px;
  color: var(--login-text-hint, #c0c4cc);
  margin: 0;
}

/* ====== 底部 ====== */
.login-form__footer {
  margin-top: 8px;
}

.login-form__footer-links {
  text-align: center;
  margin-bottom: 20px;
}

.login-form__footer-link {
  font-size: 13px;
  color: var(--login-text-secondary, #909399);
  cursor: pointer;
  transition: color 0.2s;
}

.login-form__footer-link:hover {
  color: var(--login-tab-active, #409eff);
}

.login-form__footer-divider {
  margin: 0 12px;
  color: var(--login-border, #dcdfe6);
  user-select: none;
}

/* ====== 社交登录 ====== */
.login-form__social {
  text-align: center;
}

.login-form__social-title {
  font-size: 12px;
  color: var(--login-text-hint, #c0c4cc);
  margin-bottom: 12px;
}

.login-form__social-icons {
  display: flex;
  justify-content: center;
  gap: 20px;
}

.login-form__social-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--login-social-bg, rgba(0, 0, 0, 0.04));
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition:
    background 0.2s,
    transform 0.2s;
  color: var(--login-text-secondary, #909399);
}

.login-form__social-icon:hover {
  background: var(--login-social-hover, rgba(64, 158, 255, 0.1));
  color: var(--login-tab-active, #409eff);
  transform: scale(1.1);
}

.login-form__social-svg {
  width: 18px;
  height: 18px;
}

.login-form__social-text-icon {
  font-size: 14px;
  font-weight: 600;
  line-height: 1;
}
</style>
