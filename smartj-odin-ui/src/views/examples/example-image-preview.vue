<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmImagePreview 图片预览示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">支持缩放、旋转、拖拽的全屏图片预览组件，键盘快捷键操作。</p>

      <div class="flex gap-4 flex-wrap">
        <SmImagePreview
          v-for="(url, i) in images"
          :key="i"
          :url-list="images"
          :initial-index="i"
        >
          <img
            :src="url"
            class="w-32 h-32 object-cover rounded cursor-pointer border"
          />
        </SmImagePreview>
      </div>
      <p class="text-xs text-sm-text-secondary mt-3">快捷键：← → 切换图片 | 滚轮缩放 | Esc 关闭</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import SmImagePreview from '@/components/SmImagePreview/index.vue'

const images = [
  'https://fuss10.elemecdn.com/a/3f/3302e58f9a181d2509f3dc0fa68b0jpeg.jpeg',
  'https://fuss10.elemecdn.com/1/34/19aa98b1fcb2781c4fba33d850549jpeg.jpeg',
  'https://fuss10.elemecdn.com/0/6f/e35ff375812e6b0020b6b0f0e4e0bjpeg.jpeg',
  'https://fuss10.elemecdn.com/9/bb/e27858e973f5d7d3904835f46abbdjpeg.jpeg',
]
</script>
