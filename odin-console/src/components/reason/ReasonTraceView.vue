<script setup lang="ts">
import { computed } from 'vue'
import { ElTag } from 'element-plus'
import { opLabel, valueText } from '@/utils/format'
import type { ReasonTraceVO } from '@/types/query'

/**
 * 推理轨迹视图：把后端 ReasonTrace 的四段可解释信息统一渲染。
 * 被「助手 reasoning 块」「链路溯源抽屉」「ReasonLab 规则试运行」三处复用。
 */
const props = withDefaults(
  defineProps<{
    trace?: ReasonTraceVO | null
    /** 无任何轨迹时的占位文案 */
    emptyText?: string
    /** 紧凑模式（助手气泡内使用） */
    compact?: boolean
  }>(),
  {
    trace: null,
    emptyText: '本次查询未产生推理轨迹（无规则命中、无虚拟类展开）',
    compact: false,
  },
)

const expansion = computed(() => props.trace?.expansion ?? [])
const rules = computed(() => props.trace?.rules_fired ?? [])
const derived = computed(() => props.trace?.derived_filters ?? [])
const inferred = computed(() => props.trace?.inferred_relations ?? [])

const firedCount = computed(() => rules.value.filter((r) => r.condition_result).length)
const isEmpty = computed(
  () => !expansion.value.length && !rules.value.length && !derived.value.length && !inferred.value.length,
)
</script>

<template>
  <div class="reason-trace" :class="{ compact }">
    <div v-if="isEmpty" class="empty">{{ emptyText }}</div>

    <template v-else>
      <!-- 规则命中 -->
      <section v-if="rules.length" class="sec">
        <h5>
          规则命中
          <ElTag size="small" :type="firedCount ? 'success' : 'info'" effect="plain">
            {{ firedCount }}/{{ rules.length }} 生效
          </ElTag>
        </h5>
        <ul class="list">
          <li v-for="(r, i) in rules" :key="`${r.rule_id}-${i}`" :class="{ miss: !r.condition_result }">
            <div class="line1">
              <ElTag size="small" :type="r.condition_result ? 'success' : 'info'" :effect="r.condition_result ? 'dark' : 'plain'">
                {{ r.condition_result ? '命中' : '未命中' }}
              </ElTag>
              <span class="name">{{ r.name || `规则 #${r.rule_id}` }}</span>
              <ElTag v-if="r.rule_type" size="small" type="warning" effect="plain">{{ r.rule_type }}</ElTag>
              <ElTag v-if="r.derived" size="small" type="primary" effect="plain">派生</ElTag>
            </div>
            <div v-if="r.action" class="line2">
              <span class="lbl">action</span>
              <code>{{ r.action }}</code>
            </div>
            <div v-if="r.detail" class="line2">
              <span class="lbl">detail</span>
              <span class="txt">{{ r.detail }}</span>
            </div>
          </li>
        </ul>
      </section>

      <!-- 虚拟类展开 -->
      <section v-if="expansion.length" class="sec">
        <h5>虚拟类展开</h5>
        <ul class="list">
          <li v-for="(e, i) in expansion" :key="i">
            <div class="line1 chain">
              <code class="node">{{ e.from }}</code>
              <span class="arrow">→</span>
              <code class="node to">{{ e.to }}</code>
              <ElTag v-if="e.reason" size="small" effect="plain">{{ e.reason }}</ElTag>
            </div>
          </li>
        </ul>
      </section>

      <!-- 派生过滤 -->
      <section v-if="derived.length" class="sec">
        <h5>派生过滤（规则注入）</h5>
        <ul class="list">
          <li v-for="(f, i) in derived" :key="i">
            <div class="line1 chain">
              <code class="node">{{ f.property }}</code>
              <span class="arrow">{{ opLabel(f.op) }}</span>
              <code class="node to">{{ valueText(f.value) }}</code>
              <ElTag v-if="f.source" size="small" type="warning" effect="plain">来源：{{ f.source }}</ElTag>
            </div>
          </li>
        </ul>
      </section>

      <!-- 推导关系 -->
      <section v-if="inferred.length" class="sec">
        <h5>推导关系</h5>
        <ul class="list">
          <li v-for="(r, i) in inferred" :key="i">
            <div class="line1 chain">
              <code class="node">{{ r.from_class }}</code>
              <span class="arrow">—{{ r.label || r.name }}→</span>
              <code class="node to">{{ r.to_class }}</code>
            </div>
          </li>
        </ul>
      </section>
    </template>
  </div>
</template>

<style scoped>
.reason-trace {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.reason-trace.compact {
  gap: 8px;
}

.sec h5 {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0 0 6px;
  font-size: 12px;
  font-weight: 700;
  color: #2b6cb0;
  letter-spacing: 0.3px;
}

.list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.list li {
  border-left: 3px solid #4c8bf5;
  background: #f7fbff;
  border-radius: 0 6px 6px 0;
  padding: 6px 10px;
}

.list li.miss {
  border-left-color: #cbd5e0;
  background: #f8fafc;
  opacity: 0.85;
}

.line1 {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  font-size: 12px;
}

.line1 .name {
  font-weight: 600;
  color: #1a365d;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.line2 {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-top: 3px;
  font-size: 11px;
  color: #4a5568;
}

.line2 .lbl {
  color: #a0aec0;
  font-weight: 600;
  min-width: 34px;
}

.line2 code {
  background: #edf2f7;
  border-radius: 4px;
  padding: 1px 5px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  color: #2c5282;
}

.chain {
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.chain .node {
  background: #e6f0fb;
  color: #2c5282;
  border-radius: 4px;
  padding: 1px 6px;
  font-weight: 600;
}

.chain .node.to {
  background: #d6f5e3;
  color: #22694a;
}

.chain .arrow {
  color: #4c8bf5;
  font-weight: 700;
}

.empty {
  font-size: 12px;
  color: #a0aec0;
  padding: 6px 0;
}
</style>
