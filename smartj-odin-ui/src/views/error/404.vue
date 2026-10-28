<template>
  <div class="error-page">
    <div class="error-content">
      <h1 class="error-code">404</h1>
      <p class="error-desc">{{ $t('error.pageNotFound') }}</p>
      <el-button
        type="primary"
        @click="goHome"
      >
        {{ $t('error.backHome') }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goHome() {
  router.push('/')
}
</script>

<style scoped>
.error-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #fff;
}

.error-content {
  text-align: center;
}

.error-code {
  font-size: 120px;
  font-weight: 700;
  color: #dcdfe6;
  margin: 0;
  line-height: 1;
}

.error-desc {
  font-size: 18px;
  color: #606266;
  margin: 24px 0 32px;
}
</style>
