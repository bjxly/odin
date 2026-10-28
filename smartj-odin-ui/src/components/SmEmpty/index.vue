<template>
  <div class="sm-empty">
    <div class="sm-empty__image">
      <slot name="image">
        <svg
          class="sm-empty__svg"
          :style="{ width: `${imageSize}px`, height: `${imageSize}px` }"
          viewBox="0 0 184 152"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <g opacity="0.6">
            <path
              d="M49 17H135L156 69H28L49 17Z"
              fill="var(--sm-border-lighter)"
              stroke="var(--sm-border-base)"
              stroke-width="1.5"
            />
            <rect
              x="28"
              y="69"
              width="128"
              height="60"
              rx="4"
              fill="var(--sm-bg-card)"
              stroke="var(--sm-border-base)"
              stroke-width="1.5"
            />
            <line
              x1="56"
              y1="85"
              x2="128"
              y2="85"
              stroke="var(--sm-border-lighter)"
              stroke-width="4"
              stroke-linecap="round"
            />
            <line
              x1="56"
              y1="97"
              x2="112"
              y2="97"
              stroke="var(--sm-border-lighter)"
              stroke-width="4"
              stroke-linecap="round"
            />
            <line
              x1="56"
              y1="109"
              x2="96"
              y2="109"
              stroke="var(--sm-border-lighter)"
              stroke-width="4"
              stroke-linecap="round"
            />
          </g>
        </svg>
      </slot>
    </div>
    <div class="sm-empty__desc">{{ desc }}</div>
    <div
      v-if="$slots.default"
      class="sm-empty__actions"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    description?: string
    imageSize?: number
  }>(),
  {
    description: '',
    imageSize: 120,
  }
)

const desc = computed(() => props.description || t('common.noData'))
</script>

<style scoped>
.sm-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 0;
}
.sm-empty__image {
  margin-bottom: 16px;
}
.sm-empty__svg {
  display: block;
}
.sm-empty__desc {
  font-size: var(--sm-font-size-base);
  color: var(--sm-text-secondary);
  margin-bottom: 16px;
}
.sm-empty__actions {
  display: flex;
  gap: 8px;
}
</style>
