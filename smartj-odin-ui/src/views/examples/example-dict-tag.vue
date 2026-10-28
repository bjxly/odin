<template>
  <div class="example-page">
    <h2>SmDictTag 示例</h2>
    <p class="example-desc">展示字典标签组件，通过 dictType 和 value 自动匹配标签样式。</p>

    <el-card shadow="never">
      <template #header>
        <span>任务状态 (task_status)</span>
      </template>
      <el-space wrap>
        <SmDictTag
          v-for="i in [0, 1, 2, 3, 4]"
          :key="i"
          dict-type="task_status"
          :value="i"
        />
      </el-space>
    </el-card>

    <el-card
      shadow="never"
      style="margin-top: 16px"
    >
      <template #header>
        <span>任务类型 (task_type)</span>
      </template>
      <el-space wrap>
        <SmDictTag
          v-for="t in ['shell', 'http', 'sql', 'python']"
          :key="t"
          dict-type="task_type"
          :value="t"
        />
      </el-space>
    </el-card>

    <el-card
      shadow="never"
      style="margin-top: 16px"
    >
      <template #header>
        <span>触发类型 (trigger_type)</span>
      </template>
      <el-space wrap>
        <SmDictTag
          v-for="t in ['manual', 'cron', 'api', 'subflow']"
          :key="t"
          dict-type="trigger_type"
          :value="t"
        />
      </el-space>
    </el-card>

    <el-card
      shadow="never"
      style="margin-top: 16px"
    >
      <template #header>
        <span>表格中使用</span>
      </template>
      <el-table
        :data="tableData"
        border
        stripe
      >
        <el-table-column
          prop="id"
          label="ID"
          width="80"
          align="center"
        />
        <el-table-column
          prop="name"
          label="任务名称"
          min-width="160"
        />
        <el-table-column
          label="状态"
          width="120"
          align="center"
        >
          <template #default="{ row }">
            <SmDictTag
              dict-type="task_status"
              :value="row.status"
            />
          </template>
        </el-table-column>
        <el-table-column
          label="任务类型"
          width="120"
          align="center"
        >
          <template #default="{ row }">
            <SmDictTag
              dict-type="task_type"
              :value="row.type"
            />
          </template>
        </el-table-column>
        <el-table-column
          label="触发方式"
          width="120"
          align="center"
        >
          <template #default="{ row }">
            <SmDictTag
              dict-type="trigger_type"
              :value="row.trigger"
            />
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import SmDictTag from '@/components/SmDictTag/index.vue'

const tableData = [
  { id: 1, name: '数据备份任务', status: 2, type: 'shell', trigger: 'cron' },
  { id: 2, name: '健康检查', status: 1, type: 'http', trigger: 'cron' },
  { id: 3, name: '日报生成', status: 3, type: 'python', trigger: 'manual' },
  { id: 4, name: '日志清理', status: 0, type: 'shell', trigger: 'api' },
  { id: 5, name: 'ETL 流程', status: 4, type: 'sql', trigger: 'subflow' },
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
