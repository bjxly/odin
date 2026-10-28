<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
import OntologyFormDialog from '@/views/odin/components/ontology/OntologyFormDialog.vue'
import { createOntology } from '@/api/odin/ontology'
import type { OntologyDTO } from '@/types/odin'

/**
 * 全局本体切换器：下拉选择当前活跃本体。
 *
 * 默认与 useOdinWorkspaceStore.activeOntologyId 双向绑定；
 * 也可通过 v-model 外部受控（此时以 props 为准，变更同步回 store）。
 * 存在未保存变更时显示 dirty 标记。
 */
const props = defineProps<{
  modelValue?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const workspaceStore = useOdinWorkspaceStore()
const { t } = useI18n()

/** 外部受控优先，否则跟随 store */
const selectedId = computed<string>({
  get: () => props.modelValue ?? workspaceStore.activeOntologyId,
  set: (id: string) => {
    workspaceStore.setActive(id)
    emit('update:modelValue', id)
  },
})

const options = computed(() => workspaceStore.ontologies)

onMounted(() => {
  if (options.value.length === 0) {
    void workspaceStore.loadOntologies()
  }
})

// 新建本体入口
const ontDlgVisible = ref(false)

function openNewOntology() {
  ontDlgVisible.value = true
}

async function handleCreate(data: OntologyDTO) {
  await createOntology(data)
  ElMessage.success(t('common.createSuccess'))
  await workspaceStore.loadOntologies()
  const created = workspaceStore.ontologies.find((o) => o.code === data.code)
  if (created) workspaceStore.setActive(created.id)
}
</script>

<template>
  <div class="ontology-switcher">
    <el-tooltip
      :content="t('odinWorkspace.dirty')"
      :disabled="!workspaceStore.dirty"
      placement="bottom"
    >
      <span class="dirty-badge" :class="{ 'is-dirty': workspaceStore.dirty }" />
    </el-tooltip>
    <el-select
      v-model="selectedId"
      class="switcher-select"
      size="small"
      :placeholder="t('odinWorkspace.selectOntology')"
      :loading="workspaceStore.loading"
      :no-data-text="t('odinWorkspace.noOntology')"
      :title="t('odinWorkspace.ontologySwitcher')"
    >
      <template #prefix>
        <span class="switcher-label">{{ t('odinWorkspace.activeOntology') }}</span>
      </template>
      <el-option
        v-for="o in options"
        :key="o.id"
        :label="o.name"
        :value="o.id"
      >
        <span class="option-name">{{ o.name }}</span>
        <span v-if="o.version" class="option-version">{{ o.version }}</span>
      </el-option>
      <template #footer>
        <button class="switcher-new" type="button" @click="openNewOntology">
          <el-icon :size="14"><Plus /></el-icon>
          {{ t('odinWorkspace.newOntology') }}
        </button>
      </template>
    </el-select>

    <OntologyFormDialog v-model:visible="ontDlgVisible" @save="handleCreate" />
  </div>
</template>

<style scoped>
.ontology-switcher {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

/* dirty 标记：未保存变更时点亮 */
.dirty-badge {
  width: 8px;
  height: 8px;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--sm-border-base);
  transition: background 0.25s ease, box-shadow 0.25s ease;
}

.dirty-badge.is-dirty {
  background: var(--sm-color-warning);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--sm-color-warning) 22%, transparent);
  animation: dirty-pulse 1.6s ease-in-out infinite;
}

@keyframes dirty-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--sm-color-warning) 22%, transparent);
  }
  50% {
    box-shadow: 0 0 0 5px color-mix(in srgb, var(--sm-color-warning) 10%, transparent);
  }
}

/* compact 风格 */
.switcher-select {
  width: 200px;
}

.switcher-select :deep(.el-select__wrapper) {
  min-height: 26px;
  padding: 2px 8px;
  border-radius: var(--sm-radius-md);
  background: color-mix(in srgb, var(--sm-color-primary) 5%, var(--sm-bg-card));
  border: 1px solid color-mix(in srgb, var(--sm-color-primary) 22%, var(--sm-border-light));
  box-shadow: none;
}

.switcher-select :deep(.el-select__selected-item) {
  font-size: var(--sm-font-size-xs);
  font-weight: 600;
  color: var(--sm-text-primary);
}

.switcher-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.06em;
  color: var(--sm-text-secondary);
  white-space: nowrap;
}

.option-name {
  float: left;
  font-size: var(--sm-font-size-sm);
}

.option-version {
  float: right;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 11px;
  color: var(--sm-text-secondary);
}

.switcher-new {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  padding: 6px 0;
  font-size: var(--sm-font-size-sm);
  color: var(--sm-color-primary);
  cursor: pointer;
  background: transparent;
  border: 1px dashed color-mix(in srgb, var(--sm-color-primary) 40%, transparent);
  border-radius: var(--sm-radius-md);
  transition: background 0.2s ease;
}

.switcher-new:hover {
  background: color-mix(in srgb, var(--sm-color-primary) 8%, transparent);
}
</style>
