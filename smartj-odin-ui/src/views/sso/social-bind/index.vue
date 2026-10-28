<template>
  <div class="social-bind-container">
    <div class="social-bind-card">
      <!-- 社交信息展示 -->
      <div class="social-info">
        <el-avatar :size="64" :src="bindInfo.avatarUrl">
          <el-icon :size="32"><UserFilled /></el-icon>
        </el-avatar>
        <div class="social-info-text">
          <div class="social-info-title">{{ bindInfo.nickname || $t('sso.socialUser') }}</div>
          <div class="social-info-type">
            <el-tag size="small">{{ idpLabel }}</el-tag>
            {{ $t('sso.firstLogin') }}
          </div>
        </div>
      </div>

      <el-divider />

      <!-- 步骤 1: 选择操作方式 -->
      <div v-if="step === 'choose'" class="step-content">
        <p class="step-desc">{{ $t('sso.bindHint') }}</p>
        <div class="action-options">
          <div class="action-option" @click="step = 'bind'">
            <el-icon :size="28"><Link /></el-icon>
            <span class="action-label">{{ $t('sso.bindExisting') }}</span>
            <span class="action-hint">{{ $t('sso.bindExistingHint') }}</span>
          </div>
          <div class="action-option" @click="handleCreateAndBind">
            <el-icon :size="28"><Plus /></el-icon>
            <span class="action-label">{{ $t('sso.createNew') }}</span>
            <span class="action-hint">{{ $t('sso.createNewHint') }}</span>
          </div>
        </div>
      </div>

      <!-- 步骤 2: 绑定已有账号 -->
      <div v-if="step === 'bind'" class="step-content">
        <p class="step-desc">{{ $t('sso.bindHint') }}</p>
        <el-form ref="bindFormRef" :model="bindForm" :rules="bindRules" label-width="0">
          <el-form-item prop="username">
            <el-input v-model="bindForm.username" :placeholder="$t('sso.enterUsernameOrPhone')" clearable />
          </el-form-item>
          <el-form-item prop="password">
            <el-input v-model="bindForm.password" type="password" :placeholder="$t('sso.loginPassword')" show-password @keyup.enter="handleBind" />
          </el-form-item>
        </el-form>
        <div class="step-actions">
          <el-button @click="step = 'choose'">{{ $t('sso.back') }}</el-button>
          <el-button type="primary" :loading="loading" @click="handleBind">{{ $t('sso.confirmBind') }}</el-button>
        </div>
        <div v-if="errorMsg" class="error-msg">
          <el-alert :title="errorMsg" type="error" show-icon :closable="false" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { type FormInstance, type FormRules } from 'element-plus'
import { UserFilled, Link, Plus } from '@element-plus/icons-vue'
import { bindSocialAccount, createAndBindSocial } from '@/api/sso'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const step = ref<'choose' | 'bind'>('choose')
const loading = ref(false)
const errorMsg = ref('')

const bindInfo = reactive({
  nickname: '',
  avatarUrl: '',
  idpType: '',
})

const idpLabel = computed(() => {
  const map: Record<string, string> = {
    GITHUB: 'GitHub',
    WECHAT: '微信',
    WECOM: '企业微信',
    DINGTALK: '钉钉',
    FEISHU: '飞书',
  }
  return map[bindInfo.idpType] || bindInfo.idpType
})

const bindFormRef = ref<FormInstance>()
const bindForm = reactive({
  username: '',
  password: '',
})

const bindRules: FormRules = {
  username: [{ required: true, message: t('sso.enterUsername'), trigger: 'blur' }],
  password: [{ required: true, message: t('sso.enterPassword'), trigger: 'blur' }],
}

const tempToken = ref('')

onMounted(() => {
  tempToken.value = route.query.token as string
  if (!tempToken.value) {
    errorMsg.value = t('sso.missingToken')
    return
  }
  // 尝试解析 token 中的信息（后端暂存后无法直接读取，展示默认值）
  // 实际社交信息会在绑定接口中由后端校验
  // 这里从 URL query 尝试获取额外参数（如果后端在重定向时附带）
  bindInfo.nickname = (route.query.nickname as string) || t('sso.socialUser')
  bindInfo.avatarUrl = (route.query.avatarUrl as string) || ''
  bindInfo.idpType = (route.query.idpType as string) || ''
})

async function handleBind() {
  if (!bindFormRef.value) return
  const valid = await bindFormRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  errorMsg.value = ''
  try {
    const res = await bindSocialAccount({
      tempToken: tempToken.value,
      username: bindForm.username,
      password: bindForm.password,
    })
    const userStore = useUserStore()
    userStore.setToken(res.token, res.refreshToken)
    router.replace('/')
  } catch (e: any) {
    errorMsg.value = e?.message || e?.msg || t('sso.bindFailed')
  } finally {
    loading.value = false
  }
}

async function handleCreateAndBind() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await createAndBindSocial({ tempToken: tempToken.value })
    const userStore = useUserStore()
    userStore.setToken(res.token, res.refreshToken)
    router.replace('/')
  } catch (e: any) {
    errorMsg.value = e?.message || e?.msg || t('sso.createFailed')
    loading.value = false
  }
}
</script>

<style scoped>
.social-bind-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--el-bg-color-page);
}

.social-bind-card {
  width: 420px;
  padding: 32px;
  background: var(--el-bg-color);
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.social-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.social-info-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.social-info-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.social-info-type {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.step-content {
  margin-top: 4px;
}

.step-desc {
  font-size: 14px;
  color: var(--el-text-color-regular);
  margin-bottom: 16px;
}

.action-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s;
}

.action-option:hover {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.action-label {
  font-size: 15px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  flex: 1;
}

.action-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.step-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 16px;
}

.error-msg {
  margin-top: 12px;
}
</style>
