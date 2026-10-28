import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import {
  getDraftDetail,
  getDraftPage,
  publishDraft,
  rejectDraft,
  reviewDraftItem,
} from '@/api/odin/assistant'
import type { AiDraftItemVO, AiDraftVO, DraftQuery } from '@/types/odin'

/**
 * ODIN AI 草稿审阅 store。
 *
 * 负责草稿列表加载、明细项审阅（批准/驳回）、草稿发布与整体驳回；
 * `unreadCount` 由对话 SSE 的 `draft` 事件累加，用于 FAB 角标提示。
 */
export const useOdinDraftStore = defineStore('odinDraft', () => {
  const drafts = ref<AiDraftVO[]>([])
  const total = ref(0)
  const loading = ref(false)
  const activeDraftId = ref('')
  const activeItems = ref<AiDraftItemVO[]>([])
  const itemsLoading = ref(false)
  const publishing = ref(false)
  /** 未读草稿提示计数（打开审阅面板后清零） */
  const unreadCount = ref(0)

  const query = ref<DraftQuery>({ current: 1, size: 20 })

  /** 是否存在待审阅草稿（角标兜底判断） */
  const hasPending = computed(() => drafts.value.some((d) => d.status === 'draft'))

  /** 加载草稿列表 */
  async function loadDrafts(): Promise<void> {
    loading.value = true
    try {
      const page = await getDraftPage(query.value)
      drafts.value = page?.records ?? []
      total.value = page?.total ?? 0
    } catch (e) {
      console.error('[odinDraft] loadDrafts failed', e)
    } finally {
      loading.value = false
    }
  }

  /** 选中草稿并加载明细项 */
  async function selectDraft(id: string): Promise<void> {
    activeDraftId.value = id || ''
    if (!id) {
      activeItems.value = []
      return
    }
    itemsLoading.value = true
    try {
      const detail = await getDraftDetail(id)
      activeItems.value = detail?.items ?? []
    } catch (e) {
      console.error('[odinDraft] selectDraft failed', e)
      activeItems.value = []
    } finally {
      itemsLoading.value = false
    }
  }

  /** 审阅单条明细项后刷新明细与列表 */
  async function reviewItem(
    itemId: string,
    reviewStatus: 'approved' | 'rejected',
    comment?: string,
  ): Promise<void> {
    await reviewDraftItem(activeDraftId.value, itemId, { reviewStatus, comment })
    await Promise.all([selectDraft(activeDraftId.value), loadDrafts()])
  }

  /** 发布草稿（后端单事务落地已批准明细项） */
  async function publish(id: string): Promise<void> {
    publishing.value = true
    try {
      await publishDraft(id)
      await Promise.all([selectDraft(id), loadDrafts()])
    } finally {
      publishing.value = false
    }
  }

  /** 整体驳回草稿 */
  async function reject(id: string, reason?: string): Promise<void> {
    await rejectDraft(id, reason)
    await Promise.all([selectDraft(id), loadDrafts()])
  }

  /** SSE draft 事件回调：累加未读并刷新列表 */
  function notifyNewDraft(): void {
    unreadCount.value += 1
    void loadDrafts()
  }

  /** 清除未读角标 */
  function markRead(): void {
    unreadCount.value = 0
  }

  return {
    drafts,
    total,
    loading,
    activeDraftId,
    activeItems,
    itemsLoading,
    publishing,
    unreadCount,
    hasPending,
    query,
    loadDrafts,
    selectDraft,
    reviewItem,
    publish,
    reject,
    notifyNewDraft,
    markRead,
  }
})
