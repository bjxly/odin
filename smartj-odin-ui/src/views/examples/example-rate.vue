<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmRate 评分组件示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">支持半星、自定义颜色、字符、只读、清除等，参考 NaiveUI Rate 设计。</p>

      <!-- 基础用法 -->
      <h3 class="text-base font-medium mt-6 mb-3">基础用法</h3>
      <div class="flex items-center gap-4 flex-wrap">
        <SmRate v-model="rate1" />
        <span class="text-sm text-sm-text-secondary">{{ rate1 }} 分</span>
      </div>

      <!-- 半星模式 -->
      <h3 class="text-base font-medium mt-6 mb-3">半星模式</h3>
      <div class="flex items-center gap-4 flex-wrap">
        <SmRate v-model="rate2" allow-half :size="24" />
        <span class="text-sm text-sm-text-secondary">{{ rate2 }} 分</span>
      </div>

      <!-- 不同大小 -->
      <h3 class="text-base font-medium mt-6 mb-3">不同大小</h3>
      <div class="flex flex-col gap-2">
        <div class="flex items-center gap-4">
          <SmRate v-model="rate3" :size="16" />
          <span class="text-xs text-sm-text-secondary">小 (16px)</span>
        </div>
        <div class="flex items-center gap-4">
          <SmRate v-model="rate3" :size="24" />
          <span class="text-xs text-sm-text-secondary">中 (24px)</span>
        </div>
        <div class="flex items-center gap-4">
          <SmRate v-model="rate3" :size="32" />
          <span class="text-xs text-sm-text-secondary">大 (32px)</span>
        </div>
      </div>

      <!-- 自定义颜色 -->
      <h3 class="text-base font-medium mt-6 mb-3">自定义颜色</h3>
      <div class="flex flex-col gap-2">
        <div class="flex items-center gap-2">
          <SmRate v-model="rate4" active-color="#f53f3f" />
          <span class="text-xs text-sm-text-secondary">红色</span>
        </div>
        <div class="flex items-center gap-2">
          <SmRate v-model="rate4" active-color="#00b42a" />
          <span class="text-xs text-sm-text-secondary">绿色</span>
        </div>
        <div class="flex items-center gap-2">
          <SmRate v-model="rate4" active-color="#0fc6c2" />
          <span class="text-xs text-sm-text-secondary">青色</span>
        </div>
      </div>

      <!-- 不同数量 -->
      <h3 class="text-base font-medium mt-6 mb-3">不同数量</h3>
      <div class="flex items-center gap-2">
        <SmRate v-model="rate5" :count="3" />
      </div>
      <div class="flex items-center gap-2 mt-2">
        <SmRate v-model="rate6" :count="10" :size="16" />
      </div>

      <!-- 只读 + 显示数值 -->
      <h3 class="text-base font-medium mt-6 mb-3">只读 + 显示数值</h3>
      <SmRate :model-value="4.5" allow-half readonly show-value />

      <!-- 可清除 -->
      <h3 class="text-base font-medium mt-6 mb-3">可清除（再点一次取消）</h3>
      <SmRate v-model="rate7" clearable />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmRate from '@/components/SmRate/index.vue'

const rate1 = ref(3)
const rate2 = ref(2.5)
const rate3 = ref(3)
const rate4 = ref(4)
const rate5 = ref(2)
const rate6 = ref(7)
const rate7 = ref(3)
</script>
