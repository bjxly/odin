<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmQRCode 二维码组件示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">基于 qrcode 库的 Canvas 二维码生成器，支持自定义大小、颜色、容错等级、Logo 叠加，参考 NaiveUI QRCode 设计。</p>

      <!-- 基础用法 -->
      <h3 class="text-base font-medium mt-6 mb-3">基础用法</h3>
      <div class="flex items-start gap-4 flex-wrap">
        <div class="flex flex-col items-center gap-2">
          <SmQRCode :value="demoUrl" :size="160" />
          <span class="text-xs text-sm-text-secondary">默认 160x160</span>
        </div>
        <div class="flex flex-col items-center gap-2">
          <SmQRCode :value="demoUrl" :size="120" />
          <span class="text-xs text-sm-text-secondary">120x120</span>
        </div>
        <div class="flex flex-col items-center gap-2">
          <SmQRCode :value="demoUrl" :size="200" />
          <span class="text-xs text-sm-text-secondary">200x200</span>
        </div>
      </div>

      <!-- 自定义颜色 -->
      <h3 class="text-base font-medium mt-6 mb-3">自定义颜色</h3>
      <div class="flex items-start gap-4 flex-wrap">
        <div class="flex flex-col items-center gap-2">
          <SmQRCode :value="demoUrl" :size="140" fg-color="#165dff" />
          <span class="text-xs text-sm-text-secondary">蓝色</span>
        </div>
        <div class="flex flex-col items-center gap-2">
          <SmQRCode :value="demoUrl" :size="140" fg-color="#f53f3f" />
          <span class="text-xs text-sm-text-secondary">红色</span>
        </div>
        <div class="flex flex-col items-center gap-2">
          <SmQRCode :value="demoUrl" :size="140" fg-color="#00b42a" bg-color="#f0fff4" />
          <span class="text-xs text-sm-text-secondary">绿色+浅底</span>
        </div>
      </div>

      <!-- 容错等级 -->
      <h3 class="text-base font-medium mt-6 mb-3">容错等级</h3>
      <div class="flex items-start gap-4 flex-wrap mb-4">
        <el-radio-group v-model="qrcodeLevel" size="small">
          <el-radio-button label="L">L (7%)</el-radio-button>
          <el-radio-button label="M">M (15%)</el-radio-button>
          <el-radio-button label="Q">Q (25%)</el-radio-button>
          <el-radio-button label="H">H (30%)</el-radio-button>
        </el-radio-group>
      </div>
      <SmQRCode :value="demoUrl" :size="180" :level="qrcodeLevel" />

      <!-- 动态输入 -->
      <h3 class="text-base font-medium mt-6 mb-3">动态内容</h3>
      <el-input
        v-model="qrcodeText"
        placeholder="输入文字或链接"
        style="max-width: 340px; margin-bottom: 12px"
        clearable
      />
      <div class="flex items-start gap-4 flex-wrap">
        <SmQRCode ref="dynamicQrRef" :value="qrcodeText" :size="180" />
        <el-button size="small" @click="downloadQR">下载二维码</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmQRCode from '@/components/SmQRCode/index.vue'

const demoUrl = ref('https://www.naiveui.com')
const qrcodeLevel = ref('M')
const qrcodeText = ref('https://www.smartj.com')
const dynamicQrRef = ref<InstanceType<typeof SmQRCode>>()

function downloadQR() {
  const dataUrl = dynamicQrRef.value?.toDataURL('image/png')
  if (!dataUrl) return
  const link = document.createElement('a')
  link.download = 'qrcode.png'
  link.href = dataUrl
  link.click()
}
</script>
