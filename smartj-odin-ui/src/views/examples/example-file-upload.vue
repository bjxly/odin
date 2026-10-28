<template>
  <el-card header="SmFileUpload 示例">
    <el-row :gutter="24">
      <el-col :span="12">
        <h4>拖拽模式</h4>
        <SmFileUpload
          v-model="dragFiles"
          action="#"
          accept=".jpg,.png,.pdf"
          :max-size="5"
          :limit="3"
          drag
        />
      </el-col>
      <el-col :span="12">
        <h4>按钮模式</h4>
        <SmFileUpload
          v-model="buttonFiles"
          action="#"
          accept=".jpg,.png,.pdf"
          :max-size="5"
          :limit="3"
          :drag="false"
        />
      </el-col>
    </el-row>

    <el-divider />

    <h4>文件列表状态</h4>
    <el-descriptions
      :column="2"
      border
      size="small"
    >
      <el-descriptions-item label="拖拽模式文件数">{{ dragFiles.length }}</el-descriptions-item>
      <el-descriptions-item label="按钮模式文件数">{{ buttonFiles.length }}</el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { UploadFiles } from 'element-plus'
import SmFileUpload from '@/components/SmFileUpload/index.vue'

const dragFiles = ref<UploadFiles>([])
const buttonFiles = ref<UploadFiles>([])
</script>

<style scoped>
h4 {
  margin: 0 0 12px;
}
</style>
