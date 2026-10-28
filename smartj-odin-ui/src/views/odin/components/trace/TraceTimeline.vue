<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ChatDotRound,
  Aim,
  MagicStick,
  Document,
  DataAnalysis,
  Timer,
} from '@element-plus/icons-vue'
import type { TraceData, TracePhase, TraceStepDetail } from '@/types/odin/trace'

/**
 * 全链路溯源时间线：NL解析 → 意图识别 → 推理 → SQL生成 → 执行 → 耗时。
 * 由 TraceDrawer 承载，数据来自 useOdinTraceStore().traceData。
 */
const props = defineProps<{
  traceData: TraceData | null
}>()

const { t } = useI18n()

type StageStatus = 'success' | 'error' | 'running' | 'skipped'

interface StageGroup {
  key: string
  title: string
  icon: typeof ChatDotRound
  phases: TracePhase[]
  steps: TraceStepDetail[]
  status: StageStatus
  durationMs: number
}

/** 步骤按阶段归组；未识别阶段兜底挂到「执行」段，保证不丢数据 */
const stepGroups = computed<Record<string, TraceStepDetail[]>>(() => {
  const groups: Record<string, TraceStepDetail[]> = {
    nlParse: [],
    intentResolve: [],
    reason: [],
    sqlTranslate: [],
    execute: [],
  }
  const phaseToGroup: Record<TracePhase, string> = {
    nl_parse: 'nlParse',
    intent_resolve: 'intentResolve',
    ontology_map: 'reason',
    reason: 'reason',
    sql_translate: 'sqlTranslate',
    execute: 'execute',
    format: 'execute',
  }
  for (const step of props.traceData?.steps ?? []) {
    const group = phaseToGroup[step.phase] ?? 'execute'
    groups[group].push(step)
  }
  return groups
})

function stageStatus(steps: TraceStepDetail[]): StageStatus {
  if (steps.length === 0) return 'skipped'
  if (steps.some((s) => s.status === 'error')) return 'error'
  if (steps.every((s) => s.status === 'skipped')) return 'skipped'
  return 'success'
}

const stages = computed<StageGroup[]>(() => {
  const defs: Array<{ key: string; title: string; icon: typeof ChatDotRound; phases: TracePhase[] }> = [
    { key: 'nlParse', title: t('odinWorkspace.traceStages.nlParse'), icon: ChatDotRound, phases: ['nl_parse'] },
    { key: 'intentResolve', title: t('odinWorkspace.traceStages.intentResolve'), icon: Aim, phases: ['intent_resolve'] },
    { key: 'reason', title: t('odinWorkspace.traceStages.reason'), icon: MagicStick, phases: ['ontology_map', 'reason'] },
    { key: 'sqlTranslate', title: t('odinWorkspace.traceStages.sqlTranslate'), icon: Document, phases: ['sql_translate'] },
    { key: 'execute', title: t('odinWorkspace.traceStages.execute'), icon: DataAnalysis, phases: ['execute', 'format'] },
  ]
  const built: StageGroup[] = defs.map((d) => {
    const steps = stepGroups.value[d.key] ?? []
    return {
      key: d.key,
      title: d.title,
      icon: d.icon,
      phases: d.phases,
      steps,
      status: stageStatus(steps),
      durationMs: steps.reduce((sum, s) => sum + (Number(s.durationMs) || 0), 0),
    }
  })
  // 第六段：耗时汇总（合成节点，展示总耗时与步骤分布）
  const allSteps = props.traceData?.steps ?? []
  built.push({
    key: 'timing',
    title: t('odinWorkspace.traceStages.timing'),
    icon: Timer,
    phases: [],
    steps: allSteps,
    status: stageStatus(allSteps),
    durationMs:
      Number(props.traceData?.totalDurationMs)
      || allSteps.reduce((sum, s) => sum + (Number(s.durationMs) || 0), 0),
  })
  return built
})

/** 耗时条宽度：相对最大阶段耗时归一化 */
const maxStageMs = computed(() =>
  Math.max(...stages.value.map((s) => s.durationMs), 1),
)

function barWidth(stage: StageGroup): string {
  if (!stage.durationMs) return '2%'
  return `${Math.max(2, Math.min(100, (stage.durationMs / maxStageMs.value) * 100)).toFixed(1)}%`
}

const statusTextMap: Record<StageStatus, string> = {
  success: 'odinWorkspace.traceSuccess',
  error: 'odinWorkspace.traceError',
  running: 'odinWorkspace.traceRunning',
  skipped: 'odinWorkspace.traceSkipped',
}

const statusTagType: Record<StageStatus, 'success' | 'danger' | 'warning' | 'info'> = {
  success: 'success',
  error: 'danger',
  running: 'warning',
  skipped: 'info',
}

const timelineColor: Record<StageStatus, string> = {
  success: 'var(--sm-color-success)',
  error: 'var(--sm-color-danger)',
  running: 'var(--sm-color-warning)',
  skipped: 'var(--sm-text-placeholder)',
}

function formatMs(ms?: number): string {
  const n = Number(ms ?? 0)
  if (!Number.isFinite(n) || n <= 0) return '-'
  return n >= 1000 ? `${(n / 1000).toFixed(2)} s` : `${Math.round(n)} ms`
}

function formatTime(v?: string): string {
  if (!v) return ''
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  return d.toLocaleString('zh-CN', { hour12: false })
}

const overallStatus = computed<StageStatus>(() => {
  const s = String(props.traceData?.status ?? '')
  if (s === 'success' || s === 'SUCCESS') return 'success'
  if (s === 'failed' || s === 'error' || s === 'FAILED') return 'error'
  if (s === 'running' || s === 'RUNNING') return 'running'
  return 'skipped'
})
</script>

<template>
  <div v-if="!traceData" class="trace-empty">
    <el-empty :description="t('odinWorkspace.traceEmpty')" :image-size="72" />
  </div>

  <div v-else class="trace-timeline">
    <!-- 概要头 -->
    <div class="trace-head">
      <div class="head-line">
        <el-tag size="small" :type="statusTagType[overallStatus]" effect="dark">
          {{ t(statusTextMap[overallStatus]) }}
        </el-tag>
        <el-tag size="small" type="info" effect="plain">
          {{ t('odinWorkspace.traceTotalDuration') }} {{ formatMs(traceData.totalDurationMs) }}
        </el-tag>
        <el-tag size="small" type="info" effect="plain">
          {{ t('odinWorkspace.traceStepCount', { count: traceData.steps?.length ?? 0 }) }}
        </el-tag>
        <span v-if="traceData.createdAt" class="created">
          {{ t('odinWorkspace.traceCreatedAt', { time: formatTime(traceData.createdAt) }) }}
        </span>
      </div>
      <div class="trace-id">
        <span class="lbl">trace_id</span>
        <code>{{ traceData.traceId }}</code>
      </div>
      <div v-if="traceData.naturalLanguage" class="nl-line">
        <span class="lbl">{{ t('odinWorkspace.traceNaturalLanguage') }}</span>
        <span class="nl">{{ traceData.naturalLanguage }}</span>
      </div>
    </div>

    <!-- 六阶段时间线 -->
    <el-timeline class="stage-timeline">
      <el-timeline-item
        v-for="stage in stages"
        :key="stage.key"
        :color="timelineColor[stage.status]"
        :hollow="stage.status === 'skipped'"
        size="large"
      >
        <template #dot>
          <span class="stage-dot" :class="`dot-${stage.status}`">
            <el-icon :size="13"><component :is="stage.icon" /></el-icon>
          </span>
        </template>

        <div class="stage-card">
          <div class="stage-head">
            <h4>{{ stage.title }}</h4>
            <el-tag size="small" :type="statusTagType[stage.status]" effect="plain">
              {{ t(statusTextMap[stage.status]) }}
            </el-tag>
            <span class="stage-ms">{{ formatMs(stage.durationMs) }}</span>
          </div>

          <!-- 耗时条 -->
          <div class="stage-bar-track">
            <div
              class="stage-bar"
              :class="`bar-${stage.status}`"
              :style="{ width: barWidth(stage) }"
            />
          </div>

          <!-- 步骤详情（可展开） -->
          <el-collapse v-if="stage.steps.length" class="stage-collapse">
            <el-collapse-item :name="stage.key">
              <template #title>
                <span class="collapse-title">
                  {{ t('odinWorkspace.traceDetail') }} ·
                  {{ t('odinWorkspace.traceStepCount', { count: stage.steps.length }) }}
                </span>
              </template>
              <div
                v-for="step in stage.steps"
                :key="step.stepIndex"
                class="step-item"
                :class="{ 'step-error': step.status === 'error' }"
              >
                <div class="step-line">
                  <span class="step-index">#{{ step.stepIndex }}</span>
                  <span class="step-label">{{ step.label }}</span>
                  <el-tag size="small" effect="plain" :type="statusTagType[step.status === 'error' ? 'error' : step.status === 'skipped' ? 'skipped' : 'success']">
                    {{ t(`odinWorkspace.tracePhases.${step.phase}`) }}
                  </el-tag>
                  <span class="step-ms">{{ formatMs(step.durationMs) }}</span>
                </div>
                <div v-if="step.input" class="step-kv">
                  <span class="k">{{ t('odinWorkspace.traceInput') }}</span>
                  <pre class="v">{{ step.input }}</pre>
                </div>
                <div v-if="step.output" class="step-kv">
                  <span class="k">{{ t('odinWorkspace.traceOutput') }}</span>
                  <pre class="v">{{ step.output }}</pre>
                </div>
                <div v-if="step.errorMessage" class="step-err">
                  {{ t('odinWorkspace.traceErrorDetail') }}：{{ step.errorMessage }}
                </div>
              </div>
            </el-collapse-item>
          </el-collapse>
          <div v-else class="stage-none">{{ t('odinWorkspace.traceNoStep') }}</div>
        </div>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<style scoped>
.trace-timeline {
  display: flex;
  flex-direction: column;
  gap: var(--sm-spacing-lg);
}

.trace-empty {
  padding: 40px 0;
}

/* ============ 概要头 ============ */
.trace-head {
  padding: 12px 14px;
  background: color-mix(in srgb, var(--sm-color-primary) 5%, var(--sm-bg-card));
  border: 1px solid color-mix(in srgb, var(--sm-color-primary) 22%, var(--sm-border-light));
  border-radius: var(--sm-radius-lg);
}

.head-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.created {
  margin-left: auto;
  font-size: 11px;
  color: var(--sm-text-secondary);
}

.trace-id,
.nl-line {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-top: 8px;
  font-size: var(--sm-font-size-xs);
}

.trace-id .lbl,
.nl-line .lbl {
  flex-shrink: 0;
  font-weight: 600;
  color: var(--sm-text-secondary);
}

.trace-id code {
  font-family: 'JetBrains Mono', Consolas, monospace;
  color: var(--sm-color-primary);
  word-break: break-all;
}

.nl-line .nl {
  color: var(--sm-text-primary);
  line-height: 1.6;
}

/* ============ 阶段时间线 ============ */
.stage-timeline {
  padding-left: 4px;
}

.stage-timeline :deep(.el-timeline-item) {
  padding-bottom: 14px;
}

.stage-timeline :deep(.el-timeline-item__node) {
  background: transparent;
  border: none;
}

.stage-timeline :deep(.el-timeline-item__tail) {
  border-left: 2px solid var(--sm-border-light);
}

.stage-dot {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  margin-left: -13px;
  border-radius: 50%;
  color: #fff;
  box-shadow: 0 2px 6px color-mix(in srgb, var(--sm-text-primary) 20%, transparent);
}

.dot-success {
  background: linear-gradient(135deg, var(--sm-color-success), color-mix(in srgb, var(--sm-color-success) 60%, var(--sm-color-primary)));
}

.dot-error {
  background: linear-gradient(135deg, var(--sm-color-danger), color-mix(in srgb, var(--sm-color-danger) 60%, var(--sm-color-warning)));
}

.dot-running {
  background: linear-gradient(135deg, var(--sm-color-warning), color-mix(in srgb, var(--sm-color-warning) 60%, var(--sm-color-danger)));
  animation: pulse 1.4s ease-in-out infinite;
}

.dot-skipped {
  background: var(--sm-bg-header);
  color: var(--sm-text-placeholder);
  border: 1px dashed var(--sm-border-base);
  box-shadow: none;
}

@keyframes pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--sm-color-warning) 45%, transparent);
  }
  50% {
    box-shadow: 0 0 0 7px transparent;
  }
}

/* ============ 阶段卡片 ============ */
.stage-card {
  margin-left: 10px;
  padding: 10px 14px;
  background: var(--sm-bg-card);
  border: 1px solid var(--sm-border-lighter);
  border-radius: var(--sm-radius-lg);
  transition: border-color 0.25s ease;
}

.stage-card:hover {
  border-color: color-mix(in srgb, var(--sm-color-primary) 35%, var(--sm-border-light));
}

.stage-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.stage-head h4 {
  margin: 0;
  font-size: var(--sm-font-size-base);
  font-weight: 700;
  letter-spacing: 0.03em;
  color: var(--sm-text-primary);
}

.stage-ms {
  margin-left: auto;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: var(--sm-font-size-xs);
  font-weight: 700;
  color: var(--sm-color-primary);
}

/* 耗时条 */
.stage-bar-track {
  height: 5px;
  margin: 10px 0 4px;
  background: var(--sm-bg-hover);
  border-radius: 3px;
  overflow: hidden;
}

.stage-bar {
  height: 100%;
  border-radius: 3px;
  transition: width 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

.bar-success {
  background: linear-gradient(90deg, var(--sm-color-primary), var(--sm-color-success));
}

.bar-error {
  background: linear-gradient(90deg, var(--sm-color-danger), var(--sm-color-warning));
}

.bar-running {
  background: linear-gradient(90deg, var(--sm-color-warning), var(--sm-color-primary));
}

.bar-skipped {
  background: var(--sm-border-base);
}

/* 展开详情 */
.stage-collapse {
  margin-top: 4px;
  border-top: none;
  border-bottom: none;
}

.stage-collapse :deep(.el-collapse-item__header) {
  height: 32px;
  line-height: 32px;
  border-bottom: none;
  background: transparent;
}

.stage-collapse :deep(.el-collapse-item__wrap) {
  border-bottom: none;
  background: transparent;
}

.collapse-title {
  font-size: var(--sm-font-size-xs);
  font-weight: 600;
  color: var(--sm-text-secondary);
}

.step-item {
  padding: 8px 10px;
  margin-bottom: 8px;
  background: var(--sm-bg-page);
  border-left: 3px solid var(--sm-color-primary);
  border-radius: 0 var(--sm-radius-md) var(--sm-radius-md) 0;
}

.step-item.step-error {
  border-left-color: var(--sm-color-danger);
}

.step-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  font-size: var(--sm-font-size-xs);
}

.step-index {
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-weight: 700;
  color: var(--sm-text-secondary);
}

.step-label {
  font-weight: 600;
  color: var(--sm-text-primary);
}

.step-ms {
  margin-left: auto;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-weight: 700;
  color: var(--sm-color-primary);
}

.step-kv {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-top: 6px;
  font-size: 11px;
}

.step-kv .k {
  flex-shrink: 0;
  min-width: 34px;
  font-weight: 700;
  color: var(--sm-text-secondary);
}

.step-kv .v {
  flex: 1;
  min-width: 0;
  margin: 0;
  padding: 6px 8px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 11px;
  line-height: 1.6;
  color: var(--sm-text-regular);
  background: var(--sm-bg-card);
  border: 1px solid var(--sm-border-lighter);
  border-radius: var(--sm-radius-md);
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 160px;
  overflow: auto;
}

.step-err {
  margin-top: 6px;
  padding: 6px 8px;
  font-size: 11px;
  line-height: 1.6;
  color: var(--sm-color-danger);
  background: color-mix(in srgb, var(--sm-color-danger) 8%, transparent);
  border-radius: var(--sm-radius-md);
  word-break: break-all;
}

.stage-none {
  margin-top: 4px;
  font-size: var(--sm-font-size-xs);
  color: var(--sm-text-placeholder);
}
</style>
