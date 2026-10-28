<template>
  <div class="sso-page">
    <div class="sso-card" style="max-width: 520px; margin: 60px auto;">
      <div class="sso-card-header">
        <h2>{{ $t('sso.consentHeading') }}</h2>
        <p class="sso-subtitle">{{ $t('sso.consentRequest', { name: clientName }) }}</p>
      </div>

      <el-descriptions :column="1" border class="sso-desc">
        <el-descriptions-item :label="$t('sso.appName')">{{ clientName }}</el-descriptions-item>
        <el-descriptions-item :label="$t('sso.requestScopes')">{{ scopes.join('、') }}</el-descriptions-item>
      </el-descriptions>

      <div class="sso-consent-actions">
        <el-button size="large" @click="handleDeny">{{ $t('sso.deny') }}</el-button>
        <el-button type="primary" size="large" :loading="loading" @click="handleApprove">
          {{ $t('sso.approve') }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { post } from '@/api/request'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const loading = ref(false)

const clientName = ref('未知应用')
const scopes = ref<string[]>(['openid', 'profile'])

onMounted(() => {
  clientName.value = (route.query.client_name as string) || clientName.value
  if (route.query.scope) {
    scopes.value = (route.query.scope as string).split(',')
  }
})

async function handleApprove() {
  loading.value = true
  try {
    const consentId = route.query.consent_id as string
    await post('/sso/consent/approve', { consentId, approved: true })
    ElMessage.success(t('sso.consentSuccess'))
    const redirect = (route.query.redirect_uri as string) || '/'
    window.location.href = redirect
  } catch (e: any) {
    ElMessage.error(e?.message || t('sso.consentFailed'))
  } finally {
    loading.value = false
  }
}

function handleDeny() {
  router.push('/')
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
  margin-bottom: 20px;
}
.sso-card-header h2 {
  font-size: 20px;
  color: #303133;
}
.sso-subtitle {
  color: #909399;
  margin-top: 8px;
}
.sso-desc {
  margin-bottom: 24px;
}
.sso-consent-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
}
</style>
