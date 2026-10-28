<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmScrollbar 统一滚动条示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">封装 el-scrollbar 的统一样式滚动容器。</p>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <h3 class="text-base font-semibold mb-2">固定高度滚动</h3>
          <SmScrollbar height="200px">
            <div class="p-4">
              <p
                v-for="i in 20"
                :key="i"
                class="py-2 border-b border-sm-border-lighter text-sm"
              >
                第 {{ i }} 行 — 这是一段足够长的文本内容用来测试滚动条行为。
              </p>
            </div>
          </SmScrollbar>
        </div>
        <div>
          <h3 class="text-base font-semibold mb-2">最大高度滚动</h3>
          <SmScrollbar max-height="200px">
            <div class="p-4">
              <p
                v-for="i in 15"
                :key="i"
                class="py-2 border-b border-sm-border-lighter text-sm"
              >
                第 {{ i }} 行 — Max-height 模式的滚动容器。
              </p>
            </div>
          </SmScrollbar>
        </div>
      </div>
    </div>

    <div class="sm-card mt-4">
      <h3 class="text-base font-semibold mb-3">原生滚动条模式</h3>
      <SmScrollbar
        :height="150"
        :native="true"
      >
        <div class="p-4">
          <p
            v-for="i in 12"
            :key="i"
            class="py-1 text-sm"
          >
            原生滚动条 — 第 {{ i }} 行
          </p>
        </div>
      </SmScrollbar>
    </div>
  </div>
</template>

<script setup lang="ts">
import SmScrollbar from '@/components/SmScrollbar/index.vue'
</script>
