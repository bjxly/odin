<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmWatermark 水印组件示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">全局页面水印，支持多行文字、旋转角度、透明度等自定义。</p>

      <el-form
        label-width="100px"
        class="max-w-lg mb-4"
      >
        <el-form-item label="水印文字">
          <el-input
            v-model="text1"
            placeholder="第一行"
            class="mb-2"
          />
          <el-input
            v-model="text2"
            placeholder="第二行（可选）"
          />
        </el-form-item>
        <el-form-item label="启用水印">
          <el-switch v-model="enabled" />
        </el-form-item>
        <el-form-item label="透明度">
          <el-slider
            v-model="opacity"
            :min="0.05"
            :max="0.5"
            :step="0.05"
            show-input
          />
        </el-form-item>
      </el-form>

      <p class="text-xs text-sm-text-secondary">当前水印：{{ watermarkText.join(' | ') }} | 透明度：{{ opacity }}</p>
    </div>

    <SmWatermark
      :text="watermarkText"
      :opacity="opacity"
      :enabled="enabled"
      :rotate="-20"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import SmWatermark from '@/components/SmWatermark/index.vue'

const text1 = ref('SmartJ Admin')
const text2 = ref(new Date().toLocaleDateString())
const enabled = ref(true)
const opacity = ref(0.15)

const watermarkText = computed(() => {
  const arr = [text1.value]
  if (text2.value) arr.push(text2.value)
  return arr
})
</script>
