<template>
  <div class="light-theme">
    <!-- Canvas 粒子背景 -->
    <canvas
      ref="canvasRef"
      class="light-theme__canvas"
    />
    <!-- 装饰圆 -->
    <div class="light-theme__circle light-theme__circle--1"></div>
    <div class="light-theme__circle light-theme__circle--2"></div>

    <!-- 登录卡片 -->
    <div class="light-theme__card">
      <div class="light-theme__header">
        <img
          src="/logo.png"
          alt="logo"
          class="light-theme__logo"
        />
        <LoginForm :title="appTitle" />
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
.light-theme {
  position: relative;
  display: flex;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e8f0 100%);
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

.light-theme__canvas {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

/* 装饰圆 */
.light-theme__circle {
  position: fixed;
  border-radius: 50%;
  pointer-events: none;
}

.light-theme__circle--1 {
  top: -100px;
  right: -100px;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(64, 158, 255, 0.08) 0%, transparent 70%);
}

.light-theme__circle--2 {
  bottom: -150px;
  left: -150px;
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, rgba(103, 126, 234, 0.06) 0%, transparent 70%);
}

/* 卡片 */
.light-theme__card {
  position: relative;
  z-index: 1;
  margin: auto;
  width: 440px;
  padding: 40px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.08);
  animation: fadeInUp 0.6s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.light-theme__header {
  text-align: center;
}

.light-theme__logo {
  width: 56px;
  height: 56px;
  margin-bottom: 16px;
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

/* 浅色主题标题颜色 */
:deep(.login-form__title) {
  color: #303133;
}

/* 浅色主题按钮 */
:deep(.login-form__btn) {
  border-radius: 8px;
}

/* 响应式 */
@media (max-width: 768px) {
  .light-theme__card {
    width: calc(100% - 48px);
    padding: 32px 24px;
  }
}
</style>
