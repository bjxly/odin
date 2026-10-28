<template>
  <el-card header="SmJsonEditor 示例">
    <el-row :gutter="24">
      <el-col :span="12">
        <h4>可编辑模式</h4>
        <SmJsonEditor
          v-model="editableJson"
          height="280px"
        />
      </el-col>
      <el-col :span="12">
        <h4>只读模式</h4>
        <SmJsonEditor
          v-model="readonlyJson"
          readonly
          height="280px"
        />
      </el-col>
    </el-row>

    <el-divider />

    <h4>编辑器当前值</h4>
    <el-input
      type="textarea"
      :model-value="editableJson"
      :rows="3"
      readonly
    />
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmJsonEditor from '@/components/SmJsonEditor/index.vue'

const sampleData = {
  name: 'SmartX Scheduler',
  version: '1.0.0',
  description: 'A modern task scheduling platform',
  features: ['DAG workflow', 'cron scheduling', 'failover recovery'],
  config: {
    maxRetries: 3,
    timeout: 30000,
    enabled: true,
  },
}

const editableJson = ref(JSON.stringify(sampleData, null, 2))
const readonlyJson = ref(JSON.stringify(sampleData, null, 2))
</script>

<style scoped>
h4 {
  margin: 0 0 12px;
}
</style>
