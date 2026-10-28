<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmIconSelect 图标选择器示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">基于 Element Plus Icons 的图标可视化选择器，支持搜索。</p>

      <el-form
        label-width="100px"
        class="max-w-lg"
      >
        <el-form-item label="菜单图标">
          <SmIconSelect
            v-model="selectedIcon"
            placeholder="请选择图标"
          />
        </el-form-item>
        <el-form-item label="当前选中">
          <div class="flex items-center gap-2">
            <el-icon
              v-if="selectedIcon"
              :size="24"
            >
              <component :is="selectedIcon" />
            </el-icon>
            <span class="text-sm">{{ selectedIcon || '未选择' }}</span>
          </div>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmIconSelect from '@/components/SmIconSelect/index.vue'

const selectedIcon = ref('HomeFilled')
</script>
