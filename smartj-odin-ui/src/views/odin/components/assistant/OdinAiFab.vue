<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { ChatDotRound } from '@element-plus/icons-vue'
import CopilotChatPanel from './CopilotChatPanel.vue'
import DraftReviewPanel from './DraftReviewPanel.vue'
import { useOdinDraftStore } from '@/stores/odin/draft'

/**
 * 全局「ODIN AI」浮动入口。
 *
 * 右下角常驻按钮（点击开合抽屉），抽屉为非模态（`:modal="false"`，不遮挡页面操作），
 * 内含「对话」与「草稿审阅」双 Tab；有新草稿时按钮显示角标，仅存在历史待审阅草稿时显示小圆点。
 */
const { t } = useI18n()
const draftStore = useOdinDraftStore()

const visible = ref(false)
const activeTab = ref<'chat' | 'draft'>('chat')

/** 角标数：本轮新增草稿数（切到审阅 Tab 后清零） */
const badge = computed(() => draftStore.unreadCount)

/** 无新增草稿但存在待审阅草稿时，用小圆点提示 */
const showDot = computed(() => badge.value === 0 && draftStore.hasPending)

/** 点击 FAB：切换抽屉开合，打开时回到对话 Tab */
function toggle(): void {
  visible.value = !visible.value
  if (visible.value) {
    activeTab.value = 'chat'
  }
}

/** 切到草稿审阅 Tab 时清除未读角标并刷新列表 */
function onTabChange(name: string | number): void {
  if (name === 'draft') {
    draftStore.markRead()
    void draftStore.loadDrafts()
  }
}

/** 新草稿到达时轻提示，引导用户前往审阅 */
watch(badge, (n, old) => {
  if (n > (old ?? 0)) {
    ElMessage.info(t('odinAssistant.copilotDraftCreated'))
  }
})
</script>

<template>
  <button
    :class="['odin-ai-fab', { 'odin-ai-fab--active': visible }]"
    type="button"
    :title="t('odinAssistant.copilotFab')"
    @click="toggle"
  >
    <el-badge :value="badge" :is-dot="showDot" :hidden="badge === 0 && !showDot" :max="9">
      <el-icon :size="20"><ChatDotRound /></el-icon>
    </el-badge>
    <span class="odin-ai-fab__label">ODIN AI</span>
  </button>

  <el-drawer
    v-model="visible"
    class="odin-ai-drawer"
    :title="t('odinAssistant.copilotTitle')"
    direction="rtl"
    size="640px"
    :modal="false"
    :append-to-body="true"
  >
    <el-tabs v-model="activeTab" class="odin-ai-drawer__tabs" @tab-change="onTabChange">
      <el-tab-pane name="chat" :label="t('odinAssistant.copilotChatTab')">
        <CopilotChatPanel v-if="visible" class="odin-ai-drawer__pane" />
      </el-tab-pane>
      <el-tab-pane name="draft" :label="t('odinAssistant.copilotDraftTab')">
        <DraftReviewPanel v-if="visible && activeTab === 'draft'" class="odin-ai-drawer__pane" />
      </el-tab-pane>
    </el-tabs>
  </el-drawer>
</template>

<style scoped>
.odin-ai-fab {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 1200;
  display: flex;
  gap: 6px;
  align-items: center;
  padding: 10px 16px;
  cursor: pointer;
  background: var(--sm-color-primary);
  border: none;
  border-radius: 999px;
  box-shadow: 0 4px 16px rgb(0 0 0 / 18%);
}

.odin-ai-fab:hover {
  filter: brightness(1.08);
}

.odin-ai-fab--active {
  filter: brightness(0.92);
}

.odin-ai-fab__label {
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  letter-spacing: 0.5px;
}

.odin-ai-fab :deep(.el-badge__content) {
  border: none;
}

.odin-ai-drawer__tabs {
  height: 100%;
}

.odin-ai-drawer__pane {
  height: calc(100vh - 160px);
}
</style>

<!-- 抽屉为 append-to-body，scoped 无法命中其内部结构，此处用全局样式压缩头部与 Tab 之间的空白 -->
<style>
.odin-ai-drawer .el-drawer__header {
  padding-bottom: 10px;
  margin-bottom: 0;
  font-size: 15px;
}

.odin-ai-drawer .el-drawer__body {
  padding-top: 0;
}

.odin-ai-drawer .el-tabs__header {
  margin-bottom: 8px;
}
</style>
