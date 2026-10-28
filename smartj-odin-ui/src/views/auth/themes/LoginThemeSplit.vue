<template>
  <div class="split-theme">
    <!-- Canvas 粒子背景 -->
    <canvas
      ref="canvasRef"
      class="split-theme__canvas"
    />

    <!-- 居中卡片 -->
    <div class="split-theme__card">
      <!-- 左侧品牌区 -->
      <div class="split-theme__brand">
        <div class="split-theme__brand-content">
          <img
            src="/logo.png"
            alt="logo"
            class="split-theme__logo"
          />
          <h1 class="split-theme__title">{{ appTitle }}</h1>
          <p class="split-theme__subtitle">{{ $t('auth.brand.subtitle') }}</p>
          <div class="split-theme__features">
            <div class="split-theme__feature">
              <span class="split-theme__feature-dot"></span>
              <span>{{ $t('auth.brand.feature1') }}</span>
            </div>
            <div class="split-theme__feature">
              <span class="split-theme__feature-dot"></span>
              <span>{{ $t('auth.brand.feature2') }}</span>
            </div>
            <div class="split-theme__feature">
              <span class="split-theme__feature-dot"></span>
              <span>{{ $t('auth.brand.feature3') }}</span>
            </div>
            <div class="split-theme__feature">
              <span class="split-theme__feature-dot"></span>
              <span>{{ $t('auth.brand.feature4') }}</span>
            </div>
            <div class="split-theme__feature">
              <span class="split-theme__feature-dot"></span>
              <span>{{ $t('auth.brand.feature5') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧表单区 -->
      <div class="split-theme__form">
        <div class="split-theme__form-inner">
          <LoginForm />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import LoginForm from '../components/LoginForm.vue'
import { useParticleCanvas } from '../composables/useParticleCanvas'
import { useAppTitle } from '@/composables/useAppTitle'

const props = withDefaults(defineProps<{ title?: string }>(), {
  title: useAppTitle().value,
})

const appTitle = computed(() => props.title)

const canvasRef = ref<HTMLCanvasElement>()
useParticleCanvas(canvasRef, {
  countDesktop: 160,
  countMobile: 40,
  linkDistance: 130,
  mouseRadius: 140,
  speed: 1.35,
  radiusRange: [1.8, 3.5],
  alphaRange: [0.25, 0.65],
  alphaSpeed: 0.003,
  mouseForce: 0.015,
  lineColorRgb: '59, 130, 246',
  lineMaxOpacity: 0.28,
  lineWidth: 0.7,
  dotColorRgb: '59, 130, 246',
})
</script>

<style scoped>
.split-theme {
  position: relative;
  display: flex;
  min-height: 100vh;
  background: linear-gradient(135deg, #f0f4f8 0%, #e2e8f0 100%);
  --login-tab-border: #e4e7ed;
  --login-tab-color: #909399;
  --login-tab-active: #409eff;
  --login-border: #dcdfe6;
  --login-input-bg: #fff;
  --login-text-secondary: #606266;
  --login-text-hint: #909399;
  --login-social-bg: rgba(0, 0, 0, 0.04);
  --login-social-hover: rgba(64, 158, 255, 0.1);
}

.split-theme__canvas {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

/* 居中卡片容器 */
.split-theme__card {
  position: relative;
  z-index: 1;
  display: flex;
  margin: auto;
  width: 820px;
  min-height: 440px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.12);
  animation: fadeInUp 0.6s cubic-bezier(0.22, 1, 0.36, 1) both;
}

/* 左侧品牌区 */
.split-theme__brand {
  width: 380px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 40px 36px;
  position: relative;
  overflow: hidden;
}

.split-theme__brand::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle at 30% 50%, rgba(255, 255, 255, 0.08) 0%, transparent 50%);
  pointer-events: none;
}

.split-theme__brand-content {
  position: relative;
  z-index: 1;
  color: #fff;
}

.split-theme__logo {
  width: 48px;
  height: 48px;
  margin-bottom: 20px;
  filter: brightness(0) invert(1);
}

.split-theme__title {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 8px;
  letter-spacing: -0.5px;
}

.split-theme__subtitle {
  font-size: 14px;
  opacity: 0.85;
  margin: 0 0 32px;
}

.split-theme__features {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.split-theme__feature {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  opacity: 0.9;
}

.split-theme__feature-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.8);
  flex-shrink: 0;
}

/* 右侧表单区 */
.split-theme__form {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  padding: 40px 36px;
}

.split-theme__form-inner {
  width: 100%;
  max-width: 380px;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(24px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 分屏主题标题颜色 */
:deep(.login-form__title) {
  color: #303133;
}

/* 响应式：小屏隐藏品牌区，卡片变为单栏 */
@media (max-width: 900px) {
  .split-theme__card {
    width: calc(100% - 48px);
    max-width: 440px;
    min-height: auto;
  }
  .split-theme__brand {
    display: none;
  }
  .split-theme__form {
    padding: 32px 24px;
  }
}
</style>
