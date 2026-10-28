<template>
  <div class="example-page">
    <h2>SmSearchForm 示例</h2>
    <p class="example-desc">展示搜索表单组件，支持基础字段和可展开的高级字段。</p>

    <SmSearchForm
      :model="formModel"
      @search="handleSearch"
      @reset="handleReset"
    >
      <el-form-item label="名称">
        <el-input
          v-model="formModel.name"
          placeholder="请输入名称"
          clearable
        />
      </el-form-item>
      <el-form-item label="状态">
        <el-select
          v-model="formModel.status"
          placeholder="请选择状态"
          clearable
        >
          <el-option
            label="待执行"
            value="0"
          />
          <el-option
            label="运行中"
            value="1"
          />
          <el-option
            label="成功"
            value="2"
          />
          <el-option
            label="失败"
            value="3"
          />
          <el-option
            label="已终止"
            value="4"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="类型">
        <el-select
          v-model="formModel.type"
          placeholder="请选择类型"
          clearable
        >
          <el-option
            label="Shell"
            value="shell"
          />
          <el-option
            label="HTTP"
            value="http"
          />
          <el-option
            label="SQL"
            value="sql"
          />
          <el-option
            label="Python"
            value="python"
          />
        </el-select>
      </el-form-item>

      <template #expand>
        <el-form-item label="触发方式">
          <el-select
            v-model="formModel.trigger"
            placeholder="请选择"
            clearable
          >
            <el-option
              label="手动"
              value="manual"
            />
            <el-option
              label="定时"
              value="cron"
            />
            <el-option
              label="API"
              value="api"
            />
            <el-option
              label="子流程"
              value="subflow"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="创建人">
          <el-input
            v-model="formModel.creator"
            placeholder="请输入创建人"
            clearable
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="formModel.description"
            placeholder="请输入描述"
            clearable
            style="width: 200px"
          />
        </el-form-item>
      </template>
    </SmSearchForm>

    <SmCrudTable
      :data="filteredData"
      :columns="columns"
      :total="filteredData.length"
      :show-selection="true"
    >
      <template #title>任务列表</template>
      <template #header-actions>
        <el-button type="primary">新增任务</el-button>
      </template>
      <template #status="{ row }">
        <SmDictTag
          dict-type="task_status"
          :value="row.status"
        />
      </template>
      <template #operate>
        <el-button
          link
          type="primary"
          size="small"
        >
          编辑
        </el-button>
        <el-button
          link
          type="danger"
          size="small"
        >
          删除
        </el-button>
      </template>
    </SmCrudTable>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed } from 'vue'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import SmDictTag from '@/components/SmDictTag/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'

const formModel = reactive({
  name: '',
  status: '',
  type: '',
  trigger: '',
  creator: '',
  description: '',
})

const columns: TableColumn[] = [
  { prop: 'id', label: 'ID', width: 80 },
  { prop: 'name', label: '任务名称', minWidth: 160 },
  { prop: 'status', label: '状态', width: 120, slot: 'status' },
  { prop: 'type', label: '类型', width: 100 },
  { prop: 'trigger', label: '触发方式', width: 100 },
  { prop: 'creator', label: '创建人', width: 100 },
  { prop: 'createTime', label: '创建时间', width: 180 },
]

const mockData = [
  { id: 1, name: '数据备份任务', status: 2, type: 'shell', trigger: 'cron', creator: 'admin', createTime: '2025-01-10 08:00:00' },
  { id: 2, name: '健康检查', status: 1, type: 'http', trigger: 'cron', creator: 'admin', createTime: '2025-01-11 09:30:00' },
  { id: 3, name: '日报生成', status: 3, type: 'python', trigger: 'manual', creator: 'zhangsan', createTime: '2025-01-12 10:00:00' },
  { id: 4, name: '日志清理', status: 0, type: 'shell', trigger: 'api', creator: 'lisi', createTime: '2025-01-13 14:00:00' },
  { id: 5, name: 'ETL 流程', status: 4, type: 'sql', trigger: 'subflow', creator: 'wangwu', createTime: '2025-01-14 16:00:00' },
]

const filteredData = computed(() => {
  return mockData.filter((row) => {
    if (formModel.name && !row.name.includes(formModel.name)) return false
    if (formModel.status && row.status !== Number(formModel.status)) return false
    if (formModel.type && row.type !== formModel.type) return false
    if (formModel.trigger && row.trigger !== formModel.trigger) return false
    if (formModel.creator && !row.creator.includes(formModel.creator)) return false
    return true
  })
})

function handleReset() {
  for (const key of Object.keys(formModel)) {
    ;(formModel as any)[key] = Array.isArray((formModel as any)[key]) ? [] : ''
  }
  handleSearch()
}

function handleSearch() {
  // TODO: 实际查询逻辑待接入
}
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
