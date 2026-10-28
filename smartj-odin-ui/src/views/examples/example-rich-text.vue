<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmRichText 富文本编辑器示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">基于 Tiptap 的所见即所得富文本编辑器，支持加粗、斜体、图片、链接等。</p>

      <SmRichText
        v-model="content"
        :height="400"
        placeholder="请输入公告内容..."
      />
    </div>

    <div class="sm-card mt-4">
      <h3 class="text-base font-semibold mb-3">输出内容（HTML）</h3>
      <SmCodeEditor
        :model-value="content"
        language="html"
        :height="200"
        readonly
      />
    </div>

    <div class="sm-card mt-4">
      <h3 class="text-base font-semibold mb-3">只读模式</h3>
      <SmRichText
        :model-value="content"
        :height="200"
        readonly
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmRichText from '@/components/SmRichText/index.vue'
import SmCodeEditor from '@/components/SmCodeEditor/index.vue'

const content = ref(
  '<h2>欢迎使用 SmartJ</h2><p>这是一个基于 <strong>Tiptap</strong> 的富文本编辑器。</p><ul><li><p>支持加粗、斜体、下划线</p></li><li><p>支持图片和链接插入</p></li><li><p>支持撤销/重做</p></li></ul>'
)
</script>
