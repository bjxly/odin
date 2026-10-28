<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmCalendar 日历组件示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">支持单选、范围选择、多选、自定义日期格渲染，参考 NaiveUI Calendar 设计。</p>

      <div class="flex flex-wrap gap-6">
        <!-- 单选模式 -->
        <div>
          <h3 class="text-base font-medium mb-3">单选模式</h3>
          <div class="text-xs text-sm-text-secondary mb-2">
            已选：{{ singleDate ? singleDate.toLocaleDateString() : '未选择' }}
          </div>
          <SmCalendar v-model="singleDate" mode="single" />
        </div>

        <!-- 范围选择模式 -->
        <div>
          <h3 class="text-base font-medium mb-3">范围选择模式</h3>
          <div class="text-xs text-sm-text-secondary mb-2">
            {{ rangeDate?.[0]?.toLocaleDateString() || '未选择' }}
            ~
            {{ rangeDate?.[1]?.toLocaleDateString() || '未选择' }}
          </div>
          <SmCalendar v-model="rangeDate" mode="range" />
        </div>
      </div>

      <!-- 多选模式 -->
      <div class="mt-6">
        <h3 class="text-base font-medium mb-3">多选模式</h3>
        <div class="text-xs text-sm-text-secondary mb-2">
          已选 {{ multiDate.length }} 天：
          {{ multiDate.map((d) => d.toLocaleDateString()).join('、') || '无' }}
        </div>
        <SmCalendar v-model="multiDate" mode="multiple" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmCalendar from '@/components/SmCalendar/index.vue'

const singleDate = ref<Date>()
const rangeDate = ref<[Date, Date]>()
const multiDate = ref<Date[]>([])
</script>
