<template>
  <div class="social-callback-container">
    <el-result :icon="statusIcon" :title="statusTitle" :sub-title="statusSubTitle">
      <template #extra>
        <el-button v-if="status === 'error'" type="primary" @click="goLogin">
          {{ $t('sso.backToLogin') }}
        </el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { exchangeSocialCode } from '@/api/sso'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const status = ref<'loading' | 'success' | 'error'>('loading')
const statusIcon = ref<'SuccessFilled' | 'WarningFilled'>('SuccessFilled')
const statusTitle = ref(t('sso.completingLogin'))
const statusSubTitle = ref(t('sso.pleaseWait'))

onMounted(async () => {
  const code = route.query.code as string

  if (!code) {
    status.value = 'error'
    statusIcon.value = 'WarningFilled'
    statusTitle.value = t('sso.missingParams')
    statusSubTitle.value = t('sso.missingCode')
    return
  }

  try {
    const res = await exchangeSocialCode(code)
    const userStore = useUserStore()
    userStore.setToken(res.token, res.refreshToken)
    // 跳转首页，触发路由守卫加载用户信息和菜单
    router.replace('/')
  } catch {
    status.value = 'error'
    statusIcon.value = 'WarningFilled'
    statusTitle.value = t('sso.loginFailed')
    statusSubTitle.value = t('sso.codeExpired')
  }
})

function goLogin() {
  router.replace('/sso/login')
}
</script>

<style scoped>
.social-callback-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--el-bg-color-page);
}
</style>
