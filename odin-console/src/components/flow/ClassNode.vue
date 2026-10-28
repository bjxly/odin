<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'

export interface ClassNodeProp {
  id: string
  label: string
  isKey?: boolean
  sensitive?: boolean
}

export interface ClassNodeData {
  id: string
  label: string
  description?: string
  domain?: string
  mappedSources: string[]
  isVirtual?: boolean
  props: ClassNodeProp[]
}

const props = defineProps<{
  data: ClassNodeData
  id: string
}>()

const visibleProps = computed(() => props.data.props.slice(0, 5))
const moreCount = computed(() => Math.max(0, props.data.props.length - 5))
const unmapped = computed(() => props.data.mappedSources.length === 0)

const rootClass = computed(() => ({
  'class-node': true,
  unmapped: unmapped.value,
  virtual: !!props.data.isVirtual,
}))
</script>

<template>
  <!-- 不要 click.stop：会拦掉 Vue Flow 的 node-click，导致抽屉偶发打不开 -->
  <div :class="rootClass">
    <Handle type="target" :position="Position.Left" />

    <header class="cn-header">
      <div class="cn-title">
        <strong>{{ data.label }}</strong>
        <code>{{ data.id }}</code>
      </div>
      <div class="cn-badges">
        <span v-if="data.isVirtual" class="badge virtual">虚拟</span>
        <span v-if="unmapped" class="badge warn">未映射</span>
      </div>
    </header>

    <p v-if="data.description" class="cn-desc">{{ data.description }}</p>

    <ul class="cn-props">
      <li v-for="p in visibleProps" :key="p.id" class="cn-prop">
        <span v-if="p.isKey" class="icon" title="业务键">🔑</span>
        <span v-else-if="p.sensitive" class="icon" title="敏感">🛡</span>
        <span v-else class="icon placeholder"></span>
        <span class="prop-label">{{ p.label }}</span>
      </li>
      <li v-if="moreCount" class="cn-more">+{{ moreCount }} 个属性</li>
      <li v-if="!visibleProps.length && !data.isVirtual" class="cn-empty">暂无属性</li>
    </ul>

    <footer class="cn-footer">
      <span v-for="s in data.mappedSources" :key="s" class="source-chip">{{ s }}</span>
      <span v-if="unmapped && !data.isVirtual" class="source-chip empty">待映射</span>
    </footer>

    <Handle type="source" :position="Position.Right" />
  </div>
</template>

<style scoped>
.class-node {
  min-width: 200px;
  max-width: 240px;
  background: #fff;
  border: 2px solid #c0c6d4;
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 12px;
  color: #1f2937;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.06);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s, opacity 0.15s;
}

.class-node:hover {
  border-color: #4c8bf5;
  box-shadow: 0 4px 14px rgba(76, 139, 245, 0.18);
}

.class-node.unmapped {
  border-style: dashed;
  background: repeating-linear-gradient(
    135deg,
    #fafafa,
    #fafafa 6px,
    #f3f4f6 6px,
    #f3f4f6 12px
  );
}

.class-node.virtual {
  border-color: #94a3b8;
  border-style: dotted;
  background: #f8fafc;
}

.cn-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 6px;
}

.cn-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.cn-title strong {
  font-size: 14px;
  line-height: 1.2;
}

.cn-title code {
  font-size: 11px;
  color: #64748b;
  background: #f1f5f9;
  border-radius: 4px;
  padding: 0 4px;
  width: fit-content;
}

.cn-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
}

.badge {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 999px;
  line-height: 1.4;
}

.badge.virtual {
  background: #e2e8f0;
  color: #334155;
}

.badge.warn {
  background: #fef3c7;
  color: #b45309;
}

.cn-desc {
  margin: 0 0 6px;
  color: #6b7280;
  font-size: 11px;
  line-height: 1.35;
  /* 最多两行：保证节点高度可被布局算法估算（estimateNodeSize 的 descH） */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.cn-props {
  list-style: none;
  margin: 0 0 8px;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.cn-prop {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #374151;
}

.icon {
  width: 14px;
  text-align: center;
  flex-shrink: 0;
}

.icon.placeholder {
  width: 14px;
}

.prop-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cn-more,
.cn-empty {
  color: #9ca3af;
  font-size: 11px;
}

.cn-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding-top: 6px;
  border-top: 1px dashed #e5e7eb;
}

.source-chip {
  font-size: 10px;
  background: #eff6ff;
  color: #1d4ed8;
  border: 1px solid #bfdbfe;
  border-radius: 4px;
  padding: 1px 6px;
}

.source-chip.empty {
  background: #fff7ed;
  color: #c2410c;
  border-color: #fed7aa;
}
</style>
