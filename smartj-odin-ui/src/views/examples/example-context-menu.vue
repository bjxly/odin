<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmContextMenu 右键菜单示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">在下方区域右键点击，或点击表格行右键试试。</p>

      <!-- 简单右键区域 -->
      <div
        class="sm-card p-8 text-center cursor-context-menu select-none mb-4"
        style="border: 2px dashed var(--sm-border-base)"
        @contextmenu="onAreaContextMenu"
      >
        <p class="text-sm-text-secondary">在此区域右键点击</p>
      </div>

      <!-- 表格右键 -->
      <h3 class="text-base font-semibold mb-3">表格右键菜单</h3>
      <el-table
        :data="tableData"
        border
        stripe
        @row-contextmenu="onRowContextMenu"
      >
        <el-table-column
          type="index"
          label="#"
          width="60"
        />
        <el-table-column
          prop="name"
          label="名称"
        />
        <el-table-column
          prop="status"
          label="状态"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.status === 1 ? 'success' : 'danger'"
              size="small"
            >
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>

      <!-- 右键菜单组件 -->
      <SmContextMenu
        ref="contextMenuRef"
        :options="menuOptions"
        :row-data="currentRow"
      />
    </div>

    <div class="sm-card mt-4">
      <h3 class="text-base font-semibold mb-3">操作日志</h3>
      <div class="text-sm max-h-40 overflow-auto">
        <div
          v-for="(log, i) in logs"
          :key="i"
          class="py-1 border-b border-sm-border-lighter"
        >
          {{ log }}
        </div>
        <SmEmpty
          v-if="!logs.length"
          description="暂无操作日志"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import SmContextMenu from '@/components/SmContextMenu/index.vue'
import SmEmpty from '@/components/SmEmpty/index.vue'
import type { ContextMenuOption } from '@/components/SmContextMenu/index.vue'

const contextMenuRef = ref<InstanceType<typeof SmContextMenu>>()
const currentRow = ref<any>(null)
const logs = ref<string[]>([])

const tableData = [
  { id: 1, name: '用户管理', status: 1 },
  { id: 2, name: '角色管理', status: 1 },
  { id: 3, name: '日志管理', status: 0 },
]

const menuOptions: ContextMenuOption[] = [
  { label: '查看详情', action: (row) => addLog(`查看: ${row?.name || '区域'}`) },
  { label: '编辑', action: (row) => addLog(`编辑: ${row?.name || '区域'}`) },
  { label: '─', divided: true },
  {
    label: '更多操作',
    children: [
      { label: '复制', action: (row) => addLog(`复制: ${row?.name || '区域'}`) },
      { label: '导出', action: (row) => addLog(`导出: ${row?.name || '区域'}`) },
    ],
  },
  { label: '─', divided: true },
  { label: '删除', disabled: true, action: () => {} },
]

function addLog(msg: string) {
  logs.value.unshift(`[${new Date().toLocaleTimeString()}] ${msg}`)
  ElMessage.info(msg)
}

function onAreaContextMenu(e: MouseEvent) {
  currentRow.value = { name: '区域' }
  contextMenuRef.value?.show(e)
}

function onRowContextMenu(row: any, _column: any, e: MouseEvent) {
  currentRow.value = row
  contextMenuRef.value?.show(e)
}
</script>
