<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Guide, Refresh, DocumentCopy } from '@element-plus/icons-vue'
import TraceTimeline from './TraceTimeline.vue'
import { useOdinTraceStore } from '@/stores/odin/trace'

/**
 * 全链路溯源抽屉（挂载于 MainLayout，全局唯一）。
 *
 * 由 useOdinTraceStore 驱动：任何位置调用 `openTrace(traceId)` 即可唤起，
 * 助手消息、查询工作台、洞察卡片无需各自实现查看器。
 */
const traceStore = useOdinTraceStore()
const { t } = useI18n()

const visible = computed<boolean>({
  get: () => traceStore.visible,
  set: (v) => {
    if (!v) traceStore.closeTrace()
  },
})

/** 强制重新拉取当前链路（绕过缓存） */
async function reload() {
  if (traceStore.currentTraceId) {
    await traceStore.loadTraceData(traceStore.currentTraceId)
  }
}

async function copyTraceId() {
  const id = traceStore.currentTraceId
  if (!id) return
  try {
    await navigator.clipboard.writeText(id)
    ElMessage.success(t('odinWorkspace.traceCopied'))
  } catch {
    // 剪贴板不可用时静默忽略
  }
}

function handleClose() {
  traceStore.closeTrace()
}
</script>

<template>
  <el-drawer
    v-model="visible"
    size="480px"
    direction="rtl"
    :destroy-on-close="false"
    class="odin-trace-drawer"
  >
    <template #header>
      <div class="drawer-head">
        <div class="head-title">
          <span class="head-glyph">
            <el-icon :size="15"><Guide /></el-icon>
          </span>
          <span>{{ t('odinWorkspace.traceTitle') }}</span>
        </div>
        <div class="head-actions">
          <el-button
            size="small"
            text
            :icon="DocumentCopy"
            :disabled="!traceStore.currentTraceId"
            @click="copyTraceId"
          >
            {{ t('odinWorkspace.traceCopyId') }}
          </el-button>
          <el-button
            size="small"
            text
            :icon="Refresh"
            :loading="traceStore.loading"
            :disabled="!traceStore.currentTraceId"
            @click="reload"
          >
            {{ t('odinWorkspace.traceRefresh') }}
          </el-button>
        </div>
      </div>
    </template>

    <div v-loading="traceStore.loading" class="drawer-body">
      <TraceTimeline :trace-data="traceStore.traceData" />
    </div>

    <template #footer>
      <div class="drawer-foot">
        <el-button @click="handleClose">{{ t('odinWorkspace.traceClose') }}</el-button>
      </div>
    </template>
  </el-drawer>
</template>

<style scoped>
.drawer-head {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}

.head-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: var(--sm-font-size-lg);
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--sm-text-primary);
}

.head-glyph {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: var(--sm-radius-md);
  color: var(--sm-color-primary);
  background: color-mix(in srgb, var(--sm-color-primary) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--sm-color-primary) 25%, transparent);
}

.head-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
}

.drawer-body {
  min-height: 200px;
}

.drawer-foot {
  display: flex;
  justify-content: flex-end;
}
</style>

<style>
/* 非 scoped：覆盖 Element Plus 抽屉默认布局 */
.odin-trace-drawer .el-drawer__header {
  margin-bottom: 0;
  padding: 14px 20px;
  border-bottom: 1px solid var(--sm-border-light);
  background: color-mix(in srgb, var(--sm-color-primary) 4%, var(--sm-bg-card));
}

.odin-trace-drawer .el-drawer__body {
  padding: 16px 20px 24px;
  background: var(--sm-bg-page);
}

.odin-trace-drawer .el-drawer__footer {
  padding: 10px 20px;
  border-top: 1px solid var(--sm-border-light);
}
</style>
