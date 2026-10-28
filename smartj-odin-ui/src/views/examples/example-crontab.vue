<template>
  <div class="example-page">
    <h2>SmCrontab 示例</h2>
    <p class="example-desc">展示 Cron 表达式生成器，通过 v-model 双向绑定表达式。</p>

    <el-card shadow="never">
      <template #header>
        <span>Cron 表达式生成器</span>
      </template>
      <SmCrontab v-model="cronExpression" />
    </el-card>

    <el-card
      shadow="never"
      style="margin-top: 16px"
    >
      <template #header>
        <span>当前绑定值</span>
      </template>
      <el-descriptions
        :column="1"
        border
      >
        <el-descriptions-item label="Cron 表达式">
          <el-tag
            type="primary"
            size="large"
            effect="dark"
          >
            {{ cronExpression }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="说明">
          <span>{{ cronDesc }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card
      shadow="never"
      style="margin-top: 16px"
    >
      <template #header>
        <span>常用表达式</span>
      </template>
      <el-space wrap>
        <el-button
          v-for="item in presets"
          :key="item.expr"
          size="small"
          @click="cronExpression = item.expr"
        >
          {{ item.label }}
        </el-button>
      </el-space>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import SmCrontab from '@/components/SmCrontab/index.vue'

const cronExpression = ref('0 * * * * *')

const cronDesc = computed(() => {
  const parts = cronExpression.value.split(/\s+/)
  if (parts.length < 6) return '无效表达式'
  const [sec, min, hour, day, month, week] = parts
  const segs: string[] = []
  if (sec !== '*') segs.push(`${sec}秒`)
  if (min !== '*') segs.push(`${min}分`)
  if (hour !== '*') segs.push(`${hour}时`)
  if (day !== '*') segs.push(`${day}日`)
  if (month !== '*') segs.push(`${month}月`)
  if (week !== '*') segs.push(`周${week}`)
  return segs.length > 0 ? `每 ${segs.join(' ')}` : '每秒执行'
})

const presets = [
  { label: '每秒', expr: '* * * * * *' },
  { label: '每分钟', expr: '0 * * * * *' },
  { label: '每小时', expr: '0 0 * * * *' },
  { label: '每天零点', expr: '0 0 0 * * *' },
  { label: '每周一零点', expr: '0 0 0 * * 1' },
  { label: '每月1号零点', expr: '0 0 0 1 * *' },
]
</script>

<style scoped>
.example-page {
  padding: 20px;
}
.example-desc {
  color: var(--el-text-color-secondary);
  margin-bottom: 16px;
}
</style>
