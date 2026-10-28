<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { useOdinDraftStore } from '@/stores/odin/draft'
import type { AiDraftVO } from '@/types/odin'

/**
 * ODIN AI 草稿审阅面板（HITL）。
 *
 * 左侧列出草稿，展开后逐项显示 AI 建议内容（afterSnapshot），
 * 支持单项批准/驳回，以及整草稿发布（后端单事务落地）与整体驳回。
 */
const { t } = useI18n()
const draftStore = useOdinDraftStore()

/** 当前展开的草稿ID（accordion 模式，v-model 值可能为 string | number | 数组） */
const expanded = computed<string | number | Array<string | number>>({
  get: () => draftStore.activeDraftId,
  set: (v) => {
    const target = Array.isArray(v) ? v[0] : v
    void draftStore.selectDraft(target ? String(target) : '')
    draftStore.markRead()
  },
})

/** 草稿状态筛选选项 */
const statusOptions = [
  { value: 'draft', labelKey: 'odinAssistant.draftStatusDraft' },
  { value: 'published', labelKey: 'odinAssistant.draftStatusPublished' },
  { value: 'rejected', labelKey: 'odinAssistant.draftStatusRejected' },
]

/** 草稿类型筛选选项（值为后端 draftType 编码，直接展示） */
const typeOptions = ['ontology', 'mapping', 'datasource']

/** 筛选条件变更：回到第一页并重载 */
function onFilterChange(): void {
  draftStore.query.current = 1
  void draftStore.loadDrafts()
}

/** 草稿状态标签色 */
function statusTag(status: string): 'success' | 'danger' | 'warning' | 'info' {
  if (status === 'published') return 'success'
  if (status === 'rejected') return 'danger'
  if (status === 'draft') return 'warning'
  return 'info'
}

/** 草稿状态文案（复用/扩展既有 i18n 扁平键） */
function statusText(status: string): string {
  if (status === 'published') return t('odinAssistant.draftStatusPublished')
  if (status === 'rejected') return t('odinAssistant.draftStatusRejected')
  if (status === 'approved') return t('odinAssistant.draftStatusApproved')
  return t('odinAssistant.draftStatusDraft')
}

/** 明细项审阅状态标签色 */
function reviewTag(status: string): 'success' | 'danger' | 'info' {
  if (status === 'approved') return 'success'
  if (status === 'rejected') return 'danger'
  return 'info'
}

/** 明细项审阅状态文案（复用既有键） */
function reviewText(status: string): string {
  if (status === 'approved') return t('odinAssistant.reviewApproved')
  if (status === 'rejected') return t('odinAssistant.reviewRejected')
  return t('odinAssistant.reviewPending')
}

/** 快照 JSON 美化（解析失败原样返回） */
function pretty(json?: string): string {
  if (!json) return '-'
  try {
    return JSON.stringify(JSON.parse(json), null, 2)
  } catch {
    return json
  }
}

/** 审阅单项 */
async function onReview(itemId: string, reviewStatus: 'approved' | 'rejected'): Promise<void> {
  await draftStore.reviewItem(itemId, reviewStatus)
  ElMessage.success(t('common.updateSuccess'))
}

/** 发布草稿（需二次确认） */
async function onPublish(draft: AiDraftVO): Promise<void> {
  const ok = await ElMessageBox.confirm(
    t('odinAssistant.publishConfirm'),
    t('common.tip'),
    { type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  await draftStore.publish(draft.id)
  ElMessage.success(t('odinAssistant.publishSuccess'))
}

/** 整体驳回草稿（可填原因） */
async function onReject(draft: AiDraftVO): Promise<void> {
  const reason = await ElMessageBox.prompt(
    t('odinAssistant.draftRejectPrompt'),
    t('common.tip'),
    { inputPlaceholder: t('odinAssistant.draftRejectReason'), inputValue: '' },
  )
    .then((r) => r.value as string)
    .catch(() => null)
  if (reason === null) return
  await draftStore.reject(draft.id, reason || undefined)
  ElMessage.success(t('odinAssistant.draftRejectSuccess'))
}

onMounted(() => {
  void draftStore.loadDrafts()
  draftStore.markRead()
})
</script>

<template>
  <div class="draft-review">
    <div class="draft-review__toolbar">
      <el-select
        v-model="draftStore.query.status"
        class="draft-review__filter"
        size="small"
        clearable
        :placeholder="t('odinAssistant.draftStatus')"
        @change="onFilterChange"
      >
        <el-option v-for="o in statusOptions" :key="o.value" :value="o.value" :label="t(o.labelKey)" />
      </el-select>
      <el-select
        v-model="draftStore.query.draftType"
        class="draft-review__filter"
        size="small"
        clearable
        :placeholder="t('odinAssistant.draftType')"
        @change="onFilterChange"
      >
        <el-option v-for="ty in typeOptions" :key="ty" :value="ty" :label="ty" />
      </el-select>
      <span class="draft-review__count">
        {{ t('odinAssistant.draftTotal', { total: draftStore.total }) }}
      </span>
      <el-button link type="primary" :icon="Refresh" :loading="draftStore.loading" @click="draftStore.loadDrafts()">
        {{ t('odinAssistant.draftRefresh') }}
      </el-button>
    </div>

    <div v-loading="draftStore.loading" class="draft-review__body">
      <el-empty
        v-if="!draftStore.drafts.length"
        :description="t('odinAssistant.draftEmpty')"
        :image-size="72"
      />
      <el-collapse v-else v-model="expanded" accordion>
        <el-collapse-item v-for="d in draftStore.drafts" :key="d.id" :name="d.id">
          <template #title>
            <div class="draft-item__title">
              <el-tag size="small" :type="statusTag(d.status)">{{ statusText(d.status) }}</el-tag>
              <span class="draft-item__type">{{ d.draftType }}</span>
              <span class="draft-item__summary">{{ d.summary || '-' }}</span>
            </div>
          </template>

          <div v-loading="draftStore.itemsLoading" class="draft-item__detail">
            <div v-for="it in draftStore.activeItems" :key="it.id" class="draft-item__card">
              <div class="draft-item__card-head">
                <el-tag size="small" type="info">{{ it.itemType || '-' }}</el-tag>
                <el-tag size="small">{{ it.action }}</el-tag>
                <el-tag size="small" :type="reviewTag(it.reviewStatus)">
                  {{ reviewText(it.reviewStatus) }}
                </el-tag>
                <span class="draft-item__ops">
                  <el-button
                    v-if="it.reviewStatus === 'pending'"
                    link
                    type="primary"
                    size="small"
                    @click="onReview(it.id, 'approved')"
                  >
                    {{ t('odinAssistant.approve') }}
                  </el-button>
                  <el-button
                    v-if="it.reviewStatus === 'pending'"
                    link
                    type="danger"
                    size="small"
                    @click="onReview(it.id, 'rejected')"
                  >
                    {{ t('odinAssistant.reject') }}
                  </el-button>
                </span>
              </div>
              <div v-if="it.beforeSnapshot" class="draft-item__snap">
                <div class="draft-item__snap-label">{{ t('odinAssistant.beforeSnapshot') }}</div>
                <pre class="draft-item__json">{{ pretty(it.beforeSnapshot) }}</pre>
              </div>
              <div class="draft-item__snap">
                <div class="draft-item__snap-label">{{ t('odinAssistant.afterSnapshot') }}</div>
                <pre class="draft-item__json">{{ pretty(it.afterSnapshot) }}</pre>
              </div>
              <div v-if="it.reviewerComment" class="draft-item__comment">
                {{ t('odinAssistant.reviewerComment') }}：{{ it.reviewerComment }}
              </div>
            </div>

            <div class="draft-item__footer">
              <el-button
                type="primary"
                size="small"
                :loading="draftStore.publishing"
                :disabled="d.status === 'published' || d.status === 'rejected'"
                @click="onPublish(d)"
              >
                {{ t('odinAssistant.publishDraft') }}
              </el-button>
              <el-button
                type="danger"
                size="small"
                plain
                :disabled="d.status === 'rejected' || d.status === 'published'"
                @click="onReject(d)"
              >
                {{ t('odinAssistant.draftRejectAll') }}
              </el-button>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </div>
  </div>
</template>

<style scoped>
.draft-review {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.draft-review__toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--sm-border-lighter);
}

.draft-review__filter {
  width: 110px;
}

.draft-review__count {
  margin-left: auto;
  font-size: 12px;
  color: var(--sm-text-secondary, #909399);
}

.draft-review__body {
  flex: 1;
  padding-top: 8px;
  overflow-y: auto;
}

.draft-item__title {
  display: flex;
  gap: 8px;
  align-items: center;
  overflow: hidden;
}

.draft-item__type {
  font-size: 12px;
  color: var(--sm-text-secondary, #909399);
}

.draft-item__summary {
  overflow: hidden;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.draft-item__card {
  padding: 8px;
  margin-bottom: 8px;
  background: var(--sm-bg-page);
  border-radius: var(--sm-radius-md);
}

.draft-item__card-head {
  display: flex;
  gap: 6px;
  align-items: center;
}

.draft-item__ops {
  margin-left: auto;
}

.draft-item__json {
  padding: 8px;
  margin: 8px 0 0;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
  background: var(--sm-bg-card);
  border-radius: var(--sm-radius-md);
}

.draft-item__snap {
  margin-top: 8px;
}

.draft-item__snap-label {
  font-size: 12px;
  color: var(--sm-text-secondary, #909399);
}

.draft-item__comment {
  margin-top: 6px;
  font-size: 12px;
  color: var(--sm-text-secondary, #909399);
}

.draft-item__footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  padding: 8px 0;
}
</style>
