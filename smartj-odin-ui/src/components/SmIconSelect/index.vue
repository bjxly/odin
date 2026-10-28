<template>
  <div class="sm-icon-select">
    <el-popover
      :visible="visible"
      placement="bottom-start"
      :width="420"
      trigger="click"
      @show="onShow"
      @hide="visible = false"
    >
      <template #reference>
        <div
          class="sm-icon-select__trigger"
          @click="visible = !visible"
        >
          <el-icon
            v-if="modelValue"
            class="sm-icon-select__icon"
            :size="18"
          >
            <component :is="modelValue" />
          </el-icon>
          <el-icon
            v-else
            class="sm-icon-select__placeholder"
            :size="18"
          >
            <component :is="defaultIcon" />
          </el-icon>
          <span class="sm-icon-select__text">{{ modelValue || placeholderText }}</span>
          <el-icon
            class="sm-icon-select__arrow"
            :class="{ 'is-open': visible }"
          >
            <ArrowDown />
          </el-icon>
        </div>
      </template>
      <div class="sm-icon-select__panel">
        <el-input
          v-model="keyword"
          :placeholder="$t('cmp.iconSearch')"
          clearable
          size="small"
          class="sm-icon-select__search"
        />
        <SmScrollbar height="300px">
          <div class="sm-icon-select__grid">
            <div
              v-for="name in filteredIcons"
              :key="name"
              class="sm-icon-select__item"
              :class="{ 'is-active': modelValue === name }"
              :title="name"
              @click="onSelect(name)"
            >
              <el-icon :size="22">
                <component :is="name" />
              </el-icon>
            </div>
          </div>
        </SmScrollbar>
      </div>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, shallowRef, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowDown, Picture } from '@element-plus/icons-vue'
import SmScrollbar from '@/components/SmScrollbar/index.vue'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    modelValue?: string
    placeholder?: string
  }>(),
  {
    placeholder: '',
  }
)

const placeholderText = computed(() => props.placeholder || t('cmp.iconPlaceholder'))

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
}>()

const visible = ref(false)
const keyword = ref('')
const icons = shallowRef<string[]>([])
const defaultIcon = Picture

// Dynamically collect available icon names from @element-plus/icons-vue
onMounted(async () => {
  try {
    const mod = await import('@element-plus/icons-vue')
    icons.value = Object.keys(mod).filter((k) => k !== 'default')
  } catch {
    // fallback to empty
  }
})

const filteredIcons = computed(() => {
  if (!keyword.value) return icons.value
  const kw = keyword.value.toLowerCase()
  return icons.value.filter((name) => name.toLowerCase().includes(kw))
})

function onSelect(name: string) {
  emit('update:modelValue', name)
  visible.value = false
}

function onShow() {
  keyword.value = ''
}
</script>

<style scoped>
.sm-icon-select {
  width: 100%;
}
.sm-icon-select__trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  height: 32px;
  border: 1px solid var(--sm-border-base);
  border-radius: var(--sm-radius-md);
  cursor: pointer;
  background: var(--sm-bg-card);
  transition: border-color 0.2s;
  user-select: none;
}
.sm-icon-select__trigger:hover {
  border-color: var(--sm-color-primary);
}
.sm-icon-select__icon {
  color: var(--sm-color-primary);
  flex-shrink: 0;
}
.sm-icon-select__placeholder {
  color: var(--sm-text-placeholder);
  flex-shrink: 0;
}
.sm-icon-select__text {
  flex: 1;
  font-size: var(--sm-font-size-base);
  color: var(--sm-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sm-icon-select__arrow {
  flex-shrink: 0;
  color: var(--sm-text-secondary);
  transition: transform 0.2s;
}
.sm-icon-select__arrow.is-open {
  transform: rotate(180deg);
}
.sm-icon-select__panel {
  margin: -12px;
}
.sm-icon-select__search {
  margin-bottom: 8px;
  padding: 0 4px;
}
.sm-icon-select__grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 4px;
  padding: 0 4px;
}
.sm-icon-select__item {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  border-radius: var(--sm-radius-sm);
  cursor: pointer;
  color: var(--sm-text-regular);
  transition:
    background-color 0.15s,
    color 0.15s;
}
.sm-icon-select__item:hover {
  background-color: var(--sm-bg-hover);
  color: var(--sm-color-primary);
}
.sm-icon-select__item.is-active {
  background-color: var(--sm-color-primary);
  color: #fff;
}
</style>
