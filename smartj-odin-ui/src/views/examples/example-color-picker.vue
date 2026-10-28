<template>
  <el-card header="SmColorPicker 示例">
    <el-row
      :gutter="24"
      align="middle"
    >
      <el-col :span="8">
        <h4>选择颜色</h4>
        <SmColorPicker
          v-model="color"
          :predefine="presetColors"
          show-history
        />
      </el-col>
      <el-col :span="16">
        <h4>当前值</h4>
        <el-descriptions
          :column="1"
          border
          size="small"
        >
          <el-descriptions-item label="颜色值">
            <el-tag
              :color="color"
              style="margin-right: 8px; vertical-align: middle"
            >
              &nbsp;&nbsp;&nbsp;&nbsp;
            </el-tag>
            <code>{{ color || '（未选择）' }}</code>
          </el-descriptions-item>
        </el-descriptions>

        <div
          class="color-preview-box"
          :style="{ backgroundColor: color || '#ffffff' }"
        >
          {{ color || '请选择颜色' }}
        </div>
      </el-col>
    </el-row>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmColorPicker from '@/components/SmColorPicker/index.vue'

const color = ref('#409EFF')

const presetColors = ['#ff4500', '#ff8c00', '#ffd700', '#90ee90', '#00ced1', '#1e90ff', '#c71585', '#409EFF']
</script>

<style scoped>
h4 {
  margin: 0 0 12px;
}
.color-preview-box {
  margin-top: 16px;
  width: 100%;
  height: 80px;
  border-radius: 8px;
  border: 1px solid var(--el-border-color-lighter);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
  transition: background-color 0.3s;
}
</style>
