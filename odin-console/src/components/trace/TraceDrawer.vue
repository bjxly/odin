<script setup lang="ts">
import { computed } from 'vue'
import { ElButton, ElDrawer, ElIcon, ElTag } from 'element-plus'
import { Refresh, Share } from '@element-plus/icons-vue'
import TraceTimeline from './TraceTimeline.vue'
import { useTraceStore } from '@/stores/trace'

/**
 * 全链路溯源抽屉。
 *
 * 由 trace store 驱动：任何位置调用 `useTraceStore().open(traceId)` 即可唤起，
 * 因此 ProvenanceCard、QueryLab 结果区、助手消息都无需各自实现查看器。
 */
const ts = useTraceStore()

const visible = computed<boolean>({
  get: () => ts.visible,
  set: (v) => {
    if (!v) ts.close()
  },
})

async function reload() {
  if (ts.traceId) await ts.loadTrace(ts.traceId, true)
}

async function copyTraceId() {
  if (!ts.traceId) return
  try {
    await navigator.clipboard.writeText(ts.traceId)
  } catch {
    // 剪贴板不可用时静默忽略
  }
}
</script>

<template>
  <ElDrawer
    v-model="visible"
    size="760px"
    direction="rtl"
    :destroy-on-close="false"
    class="trace-drawer"
  >
    <template #header>
      <div class="drawer-head">
        <div class="head-title">
          <ElIcon><Share /></ElIcon>
          <span>全链路溯源</span>
          <ElTag v-if="ts.trace?.parse_path" size="small" type="primary" effect="plain">
            {{ ts.trace.parse_path }}
          </ElTag>
        </div>
        <div class="head-actions">
          <ElButton size="small" text :icon="Share" @click="copyTraceId">复制 trace_id</ElButton>
          <ElButton size="small" text :icon="Refresh" :loading="ts.loading" @click="reload">
            刷新
          </ElButton>
        </div>
      </div>
    </template>

    <div v-loading="ts.loading" class="drawer-body">
      <div v-if="ts.error" class="err-box">
        <div class="err-title">链路加载失败</div>
        <div class="err-msg">{{ ts.error }}</div>
        <ElButton size="small" type="primary" plain @click="reload">重试</ElButton>
      </div>
      <TraceTimeline v-else :trace="ts.trace" />
    </div>
  </ElDrawer>
</template>

<style scoped>
.drawer-head {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 10px;
}

.head-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
  color: #1a365d;
}

.head-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;
}

.drawer-body {
  min-height: 200px;
}

.err-box {
  border: 1px solid #feb2b2;
  background: #fff5f5;
  border-radius: 10px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
}

.err-title {
  font-size: 14px;
  font-weight: 700;
  color: #c53030;
}

.err-msg {
  font-size: 12px;
  color: #742a2a;
  word-break: break-all;
}
</style>

<style>
/* 抽屉头部占满宽度（非 scoped，覆盖 Element Plus 默认布局） */
.trace-drawer .el-drawer__header {
  margin-bottom: 0;
  padding: 14px 20px;
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(135deg, #f0f7ff, #ffffff);
}

.trace-drawer .el-drawer__body {
  padding: 16px 20px 24px;
  background: #f8fafc;
}
</style>
