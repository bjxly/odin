<template>
  <el-drawer
    :model-value="visible"
    :title="title"
    :size="width"
    :direction="direction"
    :close-on-click-modal="!loading"
    :close-on-press-escape="!loading"
    destroy-on-close
    @update:model-value="$emit('update:visible', $event)"
  >
    <div class="sm-drawer-form__body">
      <slot />
    </div>
    <template #footer>
      <div class="sm-drawer-form__footer">
        <slot name="footer">
          <el-button @click="$emit('update:visible', false)">{{ $t('common.cancelBtn') }}</el-button>
          <el-button
            type="primary"
            :loading="loading"
            @click="$emit('confirm')"
          >
            {{ $t('common.confirmBtn') }}
          </el-button>
        </slot>
      </div>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    visible: boolean
    title?: string
    width?: number | string
    loading?: boolean
    direction?: 'rtl' | 'ltr' | 'ttb' | 'btt'
  }>(),
  {
    title: '',
    width: '560px',
    loading: false,
    direction: 'rtl',
  }
)

defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'confirm'): void
}>()
</script>

<style scoped>
.sm-drawer-form__body {
  padding: 0 4px;
}
.sm-drawer-form__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
