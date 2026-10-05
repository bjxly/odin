<script setup lang="ts">
import { computed } from 'vue'
import { BaseEdge, EdgeLabelRenderer, getBezierPath } from '@vue-flow/core'

const props = defineProps<{
  id: string
  sourceX: number
  sourceY: number
  targetX: number
  targetY: number
  sourcePosition: string
  targetPosition: string
  label?: string
  markerEnd?: string
  data?: { crossSource?: boolean; cardinality?: string }
}>()

const path = computed(() => {
  const [edgePath] = getBezierPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    targetX: props.targetX,
    targetY: props.targetY,
    sourcePosition: props.sourcePosition as never,
    targetPosition: props.targetPosition as never,
  })
  return edgePath
})

const labelX = computed(() => (props.sourceX + props.targetX) / 2)
const labelY = computed(() => (props.sourceY + props.targetY) / 2)
const crossSource = computed(() => !!props.data?.crossSource)
const label = computed(() => props.label ?? '')
</script>

<template>
  <BaseEdge
    :id="id"
    :path="path"
    :marker-end="markerEnd"
    :style="{
      stroke: crossSource ? '#f59e0b' : '#94a3b8',
      strokeWidth: crossSource ? 2 : 1.5,
      strokeDasharray: crossSource ? '6 4' : undefined,
    }"
  />
  <EdgeLabelRenderer>
    <div
      :style="{
        position: 'absolute',
        transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
        pointerEvents: 'all',
      }"
      class="edge-label"
      :class="{ cross: crossSource }"
    >
      {{ label }}
    </div>
  </EdgeLabelRenderer>
</template>

<style scoped>
.edge-label {
  font-size: 11px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  padding: 1px 6px;
  color: #475569;
  white-space: nowrap;
}

.edge-label.cross {
  border-color: #fcd34d;
  background: #fffbeb;
  color: #b45309;
}
</style>
