<script setup lang="ts">
import { computed } from 'vue'
import { BaseEdge, EdgeLabelRenderer, getBezierPath, Position } from '@vue-flow/core'

const props = defineProps<{
  id: string
  sourceX: number
  sourceY: number
  targetX: number
  targetY: number
  sourcePosition: Position
  targetPosition: Position
  label?: string
  markerEnd?: string
  data?: { isCrossSource?: boolean; cardinality?: string }
}>()

const path = computed(() => {
  const [edgePath] = getBezierPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    targetX: props.targetX,
    targetY: props.targetY,
    sourcePosition: props.sourcePosition,
    targetPosition: props.targetPosition,
  })
  return edgePath
})

const labelX = computed(() => (props.sourceX + props.targetX) / 2)
const labelY = computed(() => (props.sourceY + props.targetY) / 2)
const crossSource = computed(() => !!props.data?.isCrossSource)
const displayLabel = computed(() => props.label ?? '')
</script>

<template>
  <BaseEdge
    :id="id"
    :path="path"
    :marker-end="markerEnd"
    :style="{
      stroke: crossSource ? 'var(--sm-color-warning)' : 'var(--sm-text-secondary)',
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
      :class="{ 'is-cross': crossSource }"
    >
      {{ displayLabel }}
    </div>
  </EdgeLabelRenderer>
</template>

<style scoped>
.edge-label {
  font-size: 11px;
  background: var(--sm-bg-card);
  border: 1px solid var(--sm-border-light);
  border-radius: 4px;
  padding: 1px 6px;
  color: var(--sm-text-regular);
  white-space: nowrap;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.edge-label.is-cross {
  border-color: var(--sm-color-warning);
  background: color-mix(in srgb, var(--sm-color-warning) 8%, var(--sm-bg-card));
  color: var(--sm-color-warning);
}
</style>
