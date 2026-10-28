<template>
  <div class="dark-theme">
    <!-- Canvas 粒子背景 -->
    <canvas
      ref="canvasRef"
      class="dark-theme__canvas"
    />
    <!-- 网格装饰层 -->
    <div class="dark-theme__grid"></div>
    <!-- 径向光晕 -->
    <div class="dark-theme__glow"></div>

    <!-- 登录卡片 -->
    <div class="dark-theme__card">
      <div class="dark-theme__decoration"></div>
      <LoginForm :title="appTitle" />
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
  countDesktop: 200,
  countMobile: 50,
  linkDistance: 130,
  mouseRadius: 150,
  speed: 1.35,
  radiusRange: [1.8, 3.5],
  alphaRange: [0.2, 0.8],
  alphaSpeed: 0.003,
  mouseForce: 0.02,
  lineColorRgb: '100, 180, 255',
  lineMaxOpacity: 0.35,
  lineWidth: 0.6,
  dotColorRgb: '140, 200, 255',
})
</script>

<style scoped>
.dark-theme {
  position: relative;
  display: flex;
  min-height: 100vh;
  background: linear-gradient(160deg, #0a1628 0%, #0d2137 40%, #0f2b46 100%);
  --login-tab-border: rgba(100, 160, 255, 0.12);
  --login-tab-color: #64748b;
  --login-tab-active: #60a5fa;
  --login-border: rgba(100, 160, 255, 0.2);
  --login-input-bg: rgba(255, 255, 255, 0.06);
  --login-text-secondary: #94a3b8;
  --login-text-hint: #64748b;
  --login-social-bg: rgba(255, 255, 255, 0.06);
  --login-social-hover: rgba(37, 99, 235, 0.2);
}

.dark-theme__canvas {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

.dark-theme__grid {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background-image: linear-gradient(rgba(59, 130, 246, 0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(59, 130, 246, 0.04) 1px, transparent 1px);
  background-size: 60px 60px;
}

.dark-theme__glow {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 700px;
  height: 700px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(37, 99, 235, 0.12) 0%, rgba(37, 99, 235, 0.04) 40%, transparent 70%);
  z-index: 0;
  pointer-events: none;
}

.dark-theme__card {
  position: relative;
  z-index: 1;
  margin: auto;
  width: 440px;
  padding: 36px 40px 24px;
  background: rgba(15, 25, 50, 0.65);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(100, 160, 255, 0.15);
  border-radius: 16px;
  box-shadow:
    0 8px 32px rgba(0, 0, 0, 0.4),
    0 0 40px rgba(59, 130, 246, 0.08);
  animation: fadeInUp 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.dark-theme__decoration {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 50%;
  height: 3px;
  border-radius: 0 0 3px 3px;
  background: linear-gradient(90deg, #3b82f6, #06b6d4);
  box-shadow: 0 0 12px rgba(59, 130, 246, 0.5);
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(32px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 暗色主题下 LoginForm 标题颜色 */
:deep(.login-form__title) {
  color: #f1f5f9;
}

/* 暗色主题下输入框适配 */
:deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06);
  box-shadow: 0 0 0 1px rgba(100, 160, 255, 0.15) inset;
  border-radius: 8px;
  transition:
    box-shadow 0.25s ease,
    background 0.25s ease;
}

:deep(.el-input__wrapper:hover) {
  background: rgba(255, 255, 255, 0.08);
  box-shadow: 0 0 0 1px rgba(100, 160, 255, 0.25) inset;
}

:deep(.el-input__wrapper.is-focus) {
  background: rgba(255, 255, 255, 0.08);
  box-shadow:
    0 0 0 1px rgba(59, 130, 246, 0.6) inset,
    0 0 0 3px rgba(59, 130, 246, 0.12);
}

:deep(.el-input__inner) {
  color: #e2e8f0;
}

:deep(.el-input__inner::placeholder) {
  color: #64748b;
}

:deep(.el-input__prefix .el-icon),
:deep(.el-input__suffix .el-icon) {
  color: #64748b;
}

:deep(.el-checkbox__label) {
  color: #94a3b8;
}

:deep(.el-checkbox__inner) {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(100, 160, 255, 0.25);
}

:deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background: #2563eb;
  border-color: #2563eb;
}

/* 暗色主题下按钮渐变 */
:deep(.login-form__btn) {
  border: none;
  border-radius: 8px;
  background: linear-gradient(90deg, #2563eb, #06b6d4);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    background 0.3s ease;
}

:deep(.login-form__btn:hover) {
  transform: translateY(-2px);
  box-shadow: 0 6px 24px rgba(37, 99, 235, 0.5);
  background: linear-gradient(90deg, #1d4ed8, #0891b2);
}

:deep(.login-form__btn:active) {
  transform: translateY(0);
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.3);
}

/* 响应式 */
@media (max-width: 768px) {
  .dark-theme__card {
    width: calc(100% - 48px);
    padding: 30px 24px 20px;
  }
  .dark-theme__glow {
    width: 400px;
    height: 400px;
  }
}
</style>
