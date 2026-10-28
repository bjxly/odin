<template>
  <div class="sm-page profile-page">
    <el-row :gutter="20">
      <!-- 左侧：基本信息 -->
      <el-col :span="14">
        <div class="sm-card">
          <h3 class="card-header-title">{{ $t('sysProfile.basicInfo') }}</h3>
          <el-form
            ref="profileFormRef"
            :model="profileForm"
            :rules="profileRules"
            label-width="80px"
            style="max-width: 480px"
          >
            <el-form-item :label="$t('sysProfile.username')">
              <el-input
                :model-value="profile.username"
                disabled
              />
            </el-form-item>
            <el-form-item
              :label="$t('sysProfile.nickname')"
              prop="nickname"
            >
              <el-input
                v-model="profileForm.nickname"
                :placeholder="$t('sysProfile.pleaseInputNickname')"
              />
            </el-form-item>
            <el-form-item
              :label="$t('sysProfile.mobile')"
              prop="mobile"
            >
              <el-input
                v-model="profileForm.mobile"
                :placeholder="$t('sysProfile.pleaseInputMobile')"
              />
            </el-form-item>
            <el-form-item
              :label="$t('sysProfile.email')"
              prop="email"
            >
              <el-input
                v-model="profileForm.email"
                :placeholder="$t('sysProfile.pleaseInputEmail')"
              />
            </el-form-item>
            <el-form-item :label="$t('sysProfile.gender')">
              <el-radio-group v-model="profileForm.gender">
                <el-radio :value="1">{{ $t('sysProfile.male') }}</el-radio>
                <el-radio :value="2">{{ $t('sysProfile.female') }}</el-radio>
                <el-radio :value="0">{{ $t('sysProfile.unknown') }}</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('sysProfile.source')">
              <el-tag
                :type="profile.source === 'LOCAL' || !profile.source ? '' : 'success'"
                size="small"
              >
                {{ sourceLabel(profile.source) }}
              </el-tag>
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                :loading="saving"
                @click="handleSaveProfile"
              >
                {{ $t('sysProfile.saveProfile') }}
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-col>

      <!-- 右侧：安全设置 + 社交绑定 -->
      <el-col :span="10">
        <!-- 修改密码 -->
        <div
          class="sm-card"
          style="margin-bottom: 20px"
        >
          <h3 class="card-header-title">{{ $t('sysProfile.changePassword') }}</h3>
          <div
            v-if="isThirdParty"
            class="text-secondary"
          >
            <el-alert
              :title="$t('sysProfile.thirdPartyNoPwd')"
              type="info"
              show-icon
              :closable="false"
            />
          </div>
          <el-form
            v-else
            ref="pwdFormRef"
            :model="pwdForm"
            :rules="pwdRules"
            label-width="80px"
            style="max-width: 360px"
          >
            <el-form-item
              :label="$t('sysProfile.oldPassword')"
              prop="oldPassword"
            >
              <el-input
                v-model="pwdForm.oldPassword"
                type="password"
                :placeholder="$t('sysProfile.pleaseInputOldPwd')"
                show-password
              />
            </el-form-item>
            <el-form-item
              :label="$t('sysProfile.newPassword')"
              prop="newPassword"
            >
              <el-input
                v-model="pwdForm.newPassword"
                type="password"
                :placeholder="$t('sysProfile.pleaseInputNewPwd')"
                show-password
              />
            </el-form-item>
            <el-form-item
              :label="$t('sysProfile.confirmPassword')"
              prop="confirmPassword"
            >
              <el-input
                v-model="pwdForm.confirmPassword"
                type="password"
                :placeholder="$t('sysProfile.pleaseConfirmNewPwd')"
                show-password
              />
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                :loading="changingPwd"
                @click="handleChangePassword"
              >
                {{ $t('sysProfile.changePassword') }}
              </el-button>
            </el-form-item>
          </el-form>
        </div>

        <!-- 社交账号绑定 -->
        <div class="sm-card">
          <h3 class="card-header-title">{{ $t('sysProfile.socialAccounts') }}</h3>
          <div
            v-if="socialAccounts.length === 0"
            class="text-secondary"
            style="padding: 20px 0; text-align: center"
          >
            {{ $t('sysProfile.noSocialAccounts') }}
          </div>
          <div
            v-else
            class="social-list"
          >
            <div
              v-for="item in socialAccounts"
              :key="item.id"
              class="social-item"
            >
              <div class="social-item-left">
                <el-avatar
                  :size="36"
                  :src="item.avatarUrl"
                >
                  <el-icon :size="18"><UserFilled /></el-icon>
                </el-avatar>
                <div class="social-item-info">
                  <div class="social-item-name">{{ item.nickname || item.openId }}</div>
                  <div class="social-item-type">
                    <el-tag
                      size="small"
                      type="info"
                    >
                      {{ idpLabel(item.idpType) }}
                    </el-tag>
                    <span style="font-size: 12px; color: var(--el-text-color-placeholder); margin-left: 8px">
                      {{ $t('sysProfile.boundAt') }} {{ item.createTime?.substring(0, 10) }}
                    </span>
                  </div>
                </div>
              </div>
              <el-popconfirm
                :title="$t('sysProfile.confirmUnbind')"
                :confirm-button-text="$t('common.confirmBtn')"
                :cancel-button-text="$t('common.cancelBtn')"
                @confirm="handleUnbind(item.id)"
              >
                <template #reference>
                  <el-button
                    link
                    type="danger"
                    size="small"
                  >
                    {{ $t('sysProfile.unbind') }}
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { UserFilled } from '@element-plus/icons-vue'
import { getUserProfile, updateUserProfile, changePassword } from '@/api/sys/user'
import { unbindSocialAccount } from '@/api/sso'
import type { UserVO, SocialAccountVO } from '@/types/sys'
import { usePasswordMinLength } from '@/composables/usePasswordMinLength'
import { usePasswordComplexity } from '@/composables/usePasswordComplexity'

const { t } = useI18n()
const passwordMinLength = usePasswordMinLength()
const passwordComplexityRequired = usePasswordComplexity()

const saving = ref(false)
const changingPwd = ref(false)
const profileFormRef = ref<FormInstance>()
const pwdFormRef = ref<FormInstance>()

const profile = ref<UserVO>({} as UserVO)

const profileForm = reactive({
  nickname: '',
  mobile: '',
  email: '',
  gender: 0 as number | undefined,
})

const profileRules: FormRules = {
  nickname: [{ required: true, message: t('sysProfile.pleaseInputNickname'), trigger: 'blur' }],
}

const isThirdParty = computed(() => {
  return !!(profile.value.source && profile.value.source !== 'LOCAL')
})

const socialAccounts = computed<SocialAccountVO[]>(() => {
  return profile.value.socialAccounts || []
})

const pwdForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const validateConfirmPwd = (_rule: any, value: string, callback: any) => {
  if (value !== pwdForm.newPassword) {
    callback(new Error(t('sysProfile.pwdMismatch')))
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
  callback(ok ? undefined : new Error(t('sysProfile.pwdComplexity')))
}

const pwdRules = computed<FormRules>(() => ({
  oldPassword: [{ required: true, message: t('sysProfile.pleaseInputOldPwd'), trigger: 'blur' }],
  newPassword: [
    { required: true, message: t('sysProfile.pleaseInputNewPwd'), trigger: 'blur' },
    { min: passwordMinLength.value, message: t('sysProfile.pwdMinLength', { min: passwordMinLength.value }), trigger: 'blur' },
    { validator: validateComplexity, trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: t('sysProfile.pleaseConfirmNewPwd'), trigger: 'blur' },
    { validator: validateConfirmPwd, trigger: 'blur' },
  ],
}))

function sourceLabel(source?: string) {
  if (!source || source === 'LOCAL') return t('sysProfile.local')
  const map: Record<string, string> = {
    GITHUB: 'GitHub',
    WECHAT: '微信',
    WECOM: '企业微信',
    DINGTALK: '钉钉',
    FEISHU: '飞书',
    OAUTH2: 'OAuth2',
    OIDC: 'OIDC',
    SAML2: 'SAML2',
    CAS: 'CAS',
  }
  return map[source] || source
}

function idpLabel(type: string) {
  const map: Record<string, string> = {
    GITHUB: 'GitHub',
    WECHAT: '微信',
    WECOM: '企业微信',
    DINGTALK: '钉钉',
    FEISHU: '飞书',
  }
  return map[type] || type
}

async function loadProfile() {
  try {
    profile.value = await getUserProfile()
    profileForm.nickname = profile.value.nickname || ''
    profileForm.mobile = profile.value.mobile || ''
    profileForm.email = profile.value.email || ''
    profileForm.gender = profile.value.gender ?? 0
  } catch {
    // ignore
  }
}

async function handleSaveProfile() {
  if (!profileFormRef.value) return
  const valid = await profileFormRef.value.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    await updateUserProfile({
      id: profile.value.id,
      nickname: profileForm.nickname,
      mobile: profileForm.mobile,
      email: profileForm.email,
      gender: profileForm.gender,
    })
    ElMessage.success(t('sysProfile.updateProfileSuccess'))
  } finally {
    saving.value = false
  }
}

async function handleChangePassword() {
  if (!pwdFormRef.value) return
  const valid = await pwdFormRef.value.validate().catch(() => false)
  if (!valid) return

  changingPwd.value = true
  try {
    await changePassword({
      oldPassword: pwdForm.oldPassword,
      newPassword: pwdForm.newPassword,
    })
    ElMessage.success(t('sysProfile.changePwdSuccess'))
    pwdForm.oldPassword = ''
    pwdForm.newPassword = ''
    pwdForm.confirmPassword = ''
  } finally {
    changingPwd.value = false
  }
}

async function handleUnbind(id: string | number) {
  try {
    await unbindSocialAccount(Number(id))
    ElMessage.success(t('sysProfile.unbindSuccess'))
    await loadProfile()
  } catch {
    // error handled by interceptor
  }
}

onMounted(() => {
  loadProfile()
})
</script>

<style scoped>
.profile-page {
  padding: 0;
}

.card-header-title {
  font-size: 15px;
  font-weight: 600;
  margin: 0 0 16px 0;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.text-secondary {
  color: var(--el-text-color-secondary);
}

.social-list {
  display: flex;
  flex-direction: column;
}

.social-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.social-item:last-child {
  border-bottom: none;
}

.social-item-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.social-item-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.social-item-name {
  font-size: 14px;
  color: var(--el-text-color-primary);
}
</style>
