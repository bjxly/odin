<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'

export interface ClassNodeProperty {
  id: string
  /** 展示名（属性名称，无名称时回退编码） */
  name: string
  /** 属性编码（name 后括号展示） */
  code?: string
  dataType: string
  required?: boolean
}

export interface ClassNodeData {
  id: string
  name: string
  code: string
  description?: string
  domain?: string
  isVirtual?: boolean
  isUnmapped?: boolean
  isFocused?: boolean
  /** 布局方向：TB 纵向（连线在上下），LR 横向（连线在左右） */
  direction?: 'TB' | 'LR'
  properties: ClassNodeProperty[]
  mappedSources: string[]
}

const props = defineProps<{
  data: ClassNodeData
  id: string
}>()

const visibleProps = computed(() => props.data.properties.slice(0, 5))
const moreCount = computed(() => Math.max(0, props.data.properties.length - 5))

const isLR = computed(() => props.data.direction === 'LR')
const targetPos = computed(() => (isLR.value ? Position.Left : Position.Top))
const sourcePos = computed(() => (isLR.value ? Position.Right : Position.Bottom))

const rootClass = computed(() => ({
  'class-node': true,
  'is-unmapped': props.data.isUnmapped,
  'is-virtual': props.data.isVirtual,
  'is-focused': props.data.isFocused,
}))
</script>

<template>
  <div :class="rootClass">
    <Handle type="target" :position="targetPos" />

    <header class="cn-header">
      <div class="cn-title">
        <strong>{{ data.name }}</strong>
        <code class="cn-code">{{ data.code }}</code>
      </div>
      <div class="cn-badges">
        <span v-if="data.domain" class="badge domain">{{ data.domain }}</span>
        <span v-if="data.isVirtual" class="badge virtual">V</span>
        <span v-if="data.isUnmapped" class="badge unmapped">!</span>
      </div>
    </header>

    <p v-if="data.description" class="cn-desc">{{ data.description }}</p>

    <ul v-if="visibleProps.length" class="cn-props">
      <li v-for="p in visibleProps" :key="p.id" class="cn-prop">
        <span class="prop-name">{{ p.name }}<template v-if="p.code && p.code !== p.name">（{{ p.code }}）</template></span>
        <span v-if="p.dataType" class="prop-type">{{ p.dataType }}</span>
      </li>
      <li v-if="moreCount" class="cn-more">+{{ moreCount }}</li>
    </ul>
    <p v-else class="cn-empty">—</p>

    <footer v-if="data.mappedSources.length" class="cn-footer">
      <span v-for="s in data.mappedSources" :key="s" class="source-chip">{{ s }}</span>
    </footer>

    <Handle type="source" :position="sourcePos" />
  </div>
</template>

<style scoped>
.class-node {
  width: 220px;
  background: var(--sm-bg-card);
  border: 2px solid var(--sm-color-primary);
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 12px;
  color: var(--sm-text-primary);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s, background 0.15s;
}

.class-node:hover {
  box-shadow: 0 4px 16px rgba(64, 158, 255, 0.18);
}

/* 未映射态：灰色斜纹 */
.class-node.is-unmapped {
  border-color: var(--sm-border-base);
  border-style: dashed;
  background: repeating-linear-gradient(
    135deg,
    var(--sm-bg-card),
    var(--sm-bg-card) 6px,
    var(--sm-bg-page) 6px,
    var(--sm-bg-page) 12px
  );
}

/* 虚拟/规则类：绿色点线 */
.class-node.is-virtual {
  border-color: var(--sm-color-success);
  border-style: dotted;
  background: color-mix(in srgb, var(--sm-color-success) 4%, var(--sm-bg-card));
}

/* 聚焦态：高亮橙 */
.class-node.is-focused {
  border-color: var(--sm-color-warning);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--sm-color-warning) 25%, transparent);
}

.cn-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 6px;
  margin-bottom: 6px;
}

.cn-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.cn-title strong {
  font-size: 13px;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cn-code {
  font-size: 10px;
  color: var(--sm-text-secondary);
  background: var(--sm-bg-page);
  border-radius: 3px;
  padding: 0 4px;
  width: fit-content;
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.cn-badges {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.badge {
  font-size: 9px;
  padding: 1px 5px;
  border-radius: 999px;
  line-height: 1.4;
  font-weight: 600;
}

.badge.domain {
  background: color-mix(in srgb, var(--sm-color-primary) 12%, transparent);
  color: var(--sm-color-primary);
}

.badge.virtual {
  background: color-mix(in srgb, var(--sm-color-success) 15%, transparent);
  color: var(--sm-color-success);
}

.badge.unmapped {
  background: color-mix(in srgb, var(--sm-color-warning) 15%, transparent);
  color: var(--sm-color-warning);
}

.cn-desc {
  margin: 0 0 6px;
  color: var(--sm-text-secondary);
  font-size: 11px;
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.cn-props {
  list-style: none;
  margin: 0 0 6px;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.cn-prop {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  font-size: 11px;
  color: var(--sm-text-regular);
}

.prop-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.prop-type {
  flex-shrink: 0;
  font-size: 10px;
  color: var(--sm-text-secondary);
  font-family: 'JetBrains Mono', Consolas, monospace;
}

.cn-more,
.cn-empty {
  color: var(--sm-text-placeholder);
  font-size: 10px;
  margin: 0;
}

.cn-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding-top: 6px;
  border-top: 1px dashed var(--sm-border-light);
}

.source-chip {
  font-size: 9px;
  background: color-mix(in srgb, var(--sm-color-primary) 8%, transparent);
  color: var(--sm-color-primary);
  border: 1px solid color-mix(in srgb, var(--sm-color-primary) 20%, transparent);
  border-radius: 4px;
  padding: 1px 5px;
  white-space: nowrap;
}
</style>
