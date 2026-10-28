<template>
  <div
    ref="wrapperRef"
    class="sm-tree-select"
  >
    <el-popover
      :visible="visible"
      placement="bottom-start"
      :width="popoverWidth"
      trigger="click"
      :show-arrow="false"
      :offset="4"
      :popper-style="{ padding: '0' }"
      @update:visible="onPopoverVisibleChange"
    >
      <template #reference>
        <div
          class="sm-tree-select__trigger"
          :class="{ 'is-active': visible, 'is-placeholder': !selectedLabel && !hasTags }"
        >
          <span
            v-if="multiple && selectedArray.length"
            class="sm-tree-select__tags"
          >
            <el-tag
              v-for="item in displayTags"
              :key="item.id"
              size="small"
              closable
              disable-transitions
              @click.stop
              @close="removeTag(item)"
            >
              {{ item.label }}
            </el-tag>
            <el-tag
              v-if="hiddenCount > 0"
              size="small"
              type="info"
              disable-transitions
            >
              +{{ hiddenCount }}
            </el-tag>
          </span>
          <span
            v-else
            class="sm-tree-select__text"
          >
            {{ selectedLabel || placeholder }}
          </span>
          <el-icon
            class="sm-tree-select__arrow"
            :class="{ 'is-reverse': visible }"
          >
            <ArrowDown />
          </el-icon>
        </div>
      </template>

      <div class="sm-tree-select__dropdown">
        <div class="sm-tree-select__search">
          <el-input
            v-model="searchText"
            :prefix-icon="Search"
            :placeholder="$t('common.search')"
            clearable
            size="small"
          />
        </div>
        <div class="sm-tree-select__tree-wrap">
          <el-tree
            ref="treeRef"
            :data="data"
            :props="treeProps"
            :node-key="nodeKey"
            default-expand-all
            :expand-on-click-node="false"
            :check-strictly="checkStrictly"
            :filter-node-method="filterNode"
            :show-checkbox="multiple"
            highlight-current
            @node-click="onNodeClick"
            @check-change="onCheckChange"
          />
          <div
            v-if="isEmpty"
            class="sm-tree-select__empty"
          >
            {{ $t('common.noData') }}
          </div>
        </div>
      </div>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { ArrowDown, Search } from '@element-plus/icons-vue'
import type ElTree from 'element-plus/es/components/tree/src/tree.type'

interface TreeNode {
  id: string | number
  label: string
  children?: TreeNode[]
  [key: string]: any
}

const props = withDefaults(
  defineProps<{
    data: TreeNode[]
    modelValue?: string | number | (string | number)[]
    multiple?: boolean
    checkStrictly?: boolean
    placeholder?: string
    maxTagCount?: number
  }>(),
  {
    multiple: false,
    checkStrictly: false,
    placeholder: '请选择',
    maxTagCount: 3,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string | number | (string | number)[] | undefined): void
  (e: 'change', value: string | number | (string | number)[] | undefined): void
}>()

const nodeKey = 'id'
const treeProps = { label: 'label', children: 'children' }

const visible = ref(false)
const searchText = ref('')
const popoverWidth = ref(240)
const wrapperRef = ref<HTMLElement>()
const treeRef = ref()

// --- Lookup map for id -> label ---
const nodeMap = computed(() => {
  const map = new Map<string | number, string>()
  const walk = (list: TreeNode[]) => {
    for (const node of list) {
      map.set(node.id, node.label)
      if (node.children?.length) walk(node.children)
    }
  }
  walk(props.data)
  return map
})

// --- Selected display (single mode) ---
const selectedLabel = computed(() => {
  if (props.multiple) return ''
  if (props.modelValue === undefined || props.modelValue === null || props.modelValue === '') return ''
  return nodeMap.value.get(props.modelValue as string | number) ?? ''
})

// --- Selected display (multiple mode) ---
const selectedArray = computed(() => {
  if (!props.multiple || !Array.isArray(props.modelValue)) return []
  return props.modelValue.map((id) => ({ id, label: nodeMap.value.get(id) ?? String(id) }))
})

const hasTags = computed(() => props.multiple && selectedArray.value.length > 0)
const displayTags = computed(() => selectedArray.value.slice(0, props.maxTagCount))
const hiddenCount = computed(() => Math.max(0, selectedArray.value.length - props.maxTagCount))

// --- Check if filtered tree is empty ---
const isEmpty = computed(() => {
  if (!searchText.value || !treeRef.value) return false
  // After filtering, check if all nodes are hidden
  const store = treeRef.value.store
  if (!store) return false
  const root = store.root
  if (!root) return false
  return !hasVisibleNode(root)
})

function hasVisibleNode(node: any): boolean {
  if (node.visible) return true
  if (node.childNodes) {
    return node.childNodes.some((child: any) => hasVisibleNode(child))
  }
  return false
}

// --- Filtering ---
watch(searchText, (val) => {
  treeRef.value?.filter(val)
})

function filterNode(value: string, data: TreeNode): boolean {
  if (!value) return true
  return data.label.includes(value)
}

// --- Single select ---
function onNodeClick(data: TreeNode) {
  if (props.multiple) return
  emit('update:modelValue', data.id)
  emit('change', data.id)
  visible.value = false
}

// --- Multiple select ---
function onCheckChange() {
  if (!props.multiple || !treeRef.value) return
  const keys = treeRef.value.getCheckedKeys(false) as (string | number)[]
  emit('update:modelValue', keys)
  emit('change', keys)
}

// --- Remove tag in multiple mode ---
function removeTag(item: { id: string | number }) {
  if (!props.multiple || !Array.isArray(props.modelValue)) return
  const newVal = props.modelValue.filter((v) => v !== item.id)
  emit('update:modelValue', newVal)
  emit('change', newVal)
  nextTick(() => {
    treeRef.value?.setCheckedKeys(newVal as (string | number)[])
  })
}

// --- Sync checked keys when popover opens ---
function onPopoverVisibleChange(val: boolean) {
  visible.value = val
  if (val && props.multiple && Array.isArray(props.modelValue)) {
    nextTick(() => {
      treeRef.value?.setCheckedKeys(props.modelValue as (string | number)[])
    })
  }
}

// --- Measure trigger width to match popover ---
watch(visible, (val) => {
  if (val && wrapperRef.value) {
    popoverWidth.value = Math.max(wrapperRef.value.offsetWidth, 240)
  }
})

// --- Expose methods ---
defineExpose({
  /** Programmatically open/close the dropdown */
  toggle: (val?: boolean) => {
    visible.value = val ?? !visible.value
  },
  /** Clear current selection */
  clear: () => {
    if (props.multiple) {
      emit('update:modelValue', [])
      emit('change', [])
      nextTick(() => treeRef.value?.setCheckedKeys([]))
    } else {
      emit('update:modelValue', undefined)
      emit('change', undefined)
    }
  },
})
</script>

<style scoped>
.sm-tree-select {
  display: inline-block;
  width: 100%;
}

/* ---- Trigger ---- */
.sm-tree-select__trigger {
  display: flex;
  align-items: center;
  min-height: 32px;
  padding: 4px 30px 4px 11px;
  border: 1px solid var(--sm-border-base);
  border-radius: var(--sm-radius-md);
  background: var(--sm-bg-card);
  cursor: pointer;
  position: relative;
  transition: border-color 0.2s;
}

.sm-tree-select__trigger:hover {
  border-color: var(--sm-color-primary);
}

.sm-tree-select__trigger.is-active {
  border-color: var(--sm-color-primary);
  box-shadow: 0 0 0 1px var(--sm-color-primary);
}

.sm-tree-select__trigger.is-placeholder .sm-tree-select__text {
  color: var(--sm-text-placeholder);
}

.sm-tree-select__text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--sm-font-size-base);
  color: var(--sm-text-primary);
  line-height: 22px;
}

/* ---- Tags (multiple mode) ---- */
.sm-tree-select__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  flex: 1;
  min-width: 0;
}

/* ---- Arrow icon ---- */
.sm-tree-select__arrow {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--sm-text-secondary);
  transition: transform 0.3s;
  font-size: 14px;
}

.sm-tree-select__arrow.is-reverse {
  transform: translateY(-50%) rotate(180deg);
}

/* ---- Dropdown panel ---- */
.sm-tree-select__dropdown {
  padding: 0;
}

.sm-tree-select__search {
  padding: 8px 12px;
  border-bottom: 1px solid var(--sm-border-lighter);
}

.sm-tree-select__tree-wrap {
  max-height: 280px;
  overflow-y: auto;
  padding: 4px 0;
}

.sm-tree-select__tree-wrap :deep(.el-tree-node__content) {
  height: 32px;
  padding-right: 12px;
}

.sm-tree-select__tree-wrap :deep(.el-tree-node__content:hover) {
  background: var(--sm-bg-hover);
}

.sm-tree-select__tree-wrap :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background: var(--sm-bg-hover);
  color: var(--sm-color-primary);
}

.sm-tree-select__empty {
  padding: 20px 0;
  text-align: center;
  color: var(--sm-text-secondary);
  font-size: var(--sm-font-size-sm);
}
</style>
