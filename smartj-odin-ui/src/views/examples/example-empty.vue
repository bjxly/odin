<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmEmpty 空状态占位示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">统一的空状态展示组件，可自定义图片、描述和操作区。</p>

      <div class="grid grid-cols-3 gap-4">
        <div class="border rounded p-4">
          <h3 class="text-sm font-semibold mb-2">默认空状态</h3>
          <SmEmpty />
        </div>
        <div class="border rounded p-4">
          <h3 class="text-sm font-semibold mb-2">自定义描述</h3>
          <SmEmpty description="暂无用户数据，请先创建用户" />
        </div>
        <div class="border rounded p-4">
          <h3 class="text-sm font-semibold mb-2">带操作按钮</h3>
          <SmEmpty description="暂无角色数据">
            <el-button
              type="primary"
              size="small"
              @click="onAdd"
            >
              新增角色
            </el-button>
          </SmEmpty>
        </div>
      </div>
    </div>

    <div class="sm-card mt-4">
      <h3 class="text-base font-semibold mb-3">自定义图片大小</h3>
      <SmEmpty
        description="数据加载中..."
        :image-size="80"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import SmEmpty from '@/components/SmEmpty/index.vue'

function onAdd() {
  ElMessage.info('点击了新增按钮')
}
</script>
