<template>
  <Teleport to="body">
    <Transition name="sm-context-fade">
      <div
        v-if="visible"
        ref="menuRef"
        class="sm-context-menu"
        :style="menuStyle"
        @click.stop
        @contextmenu.prevent
      >
        <template
          v-for="item in options"
          :key="item.label || item.divided"
        >
          <div
            v-if="item.divided"
            class="sm-context-menu__divider"
          />
          <div
            v-if="item.children && item.children.length"
            class="sm-context-menu__sub"
          >
            <div class="sm-context-menu__item sm-context-menu__item--has-children">
              <span>{{ item.label }}</span>
              <el-icon class="sm-context-menu__arrow"><ArrowRight /></el-icon>
            </div>
            <div class="sm-context-menu__submenu">
              <div
                v-for="(child, cidx) in item.children"
                :key="cidx"
                class="sm-context-menu__item"
                :class="{ 'is-disabled': child.disabled }"
                @click.stop="onItemClick(child)"
              >
                {{ child.label }}
              </div>
            </div>
          </div>
          <div
            v-else
            class="sm-context-menu__item"
            :class="{ 'is-disabled': item.disabled }"
            @click.stop="onItemClick(item)"
          >
            {{ item.label }}
          </div>
        </template>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ArrowRight } from '@element-plus/icons-vue'

export interface ContextMenuOption {
  label: string
  disabled?: boolean
  divided?: boolean
  action?: (row?: any) => void
  children?: ContextMenuOption[]
}

const props = withDefaults(
  defineProps<{
    options?: ContextMenuOption[]
    rowData?: any
  }>(),
  {
    options: () => [],
  }
)

const visible = ref(false)
const menuRef = ref<HTMLElement>()
const position = ref({ x: 0, y: 0 })

const menuStyle = computed(() => ({
  left: `${position.value.x}px`,
  top: `${position.value.y}px`,
}))

function show(e: MouseEvent) {
  e.preventDefault()
  const x = Math.min(e.clientX, window.innerWidth - 180)
  const y = Math.min(e.clientY, window.innerHeight - 200)
  position.value = { x: Math.max(x, 4), y: Math.max(y, 4) }
  visible.value = true
}

function hide() {
  visible.value = false
}

function onItemClick(item: ContextMenuOption) {
  if (item.disabled) return
  item.action?.(props.rowData)
  hide()
}

function onGlobalClick() {
  hide()
}

function onGlobalContextMenu() {
  hide()
}

function onEsc(e: KeyboardEvent) {
  if (e.key === 'Escape') hide()
}

onMounted(() => {
  document.addEventListener('click', onGlobalClick, true)
  document.addEventListener('contextmenu', onGlobalContextMenu, true)
  document.addEventListener('keydown', onEsc)
})

onUnmounted(() => {
  document.removeEventListener('click', onGlobalClick, true)
  document.removeEventListener('contextmenu', onGlobalContextMenu, true)
  document.removeEventListener('keydown', onEsc)
})

defineExpose({ show, hide, visible })
</script>

<script lang="ts">
// Export a directive for easy usage
import type { Directive } from 'vue'

export const vContextmenu: Directive = {
  mounted(el: HTMLElement, binding) {
    const handler = (e: MouseEvent) => {
      const instance = binding.instance
      if (instance && typeof (instance as any).show === 'function') {
        ;(instance as any).show(e)
        // Pass the row data if available
        if (binding.value !== undefined && binding.value !== null) {
          ;(instance as any).rowData = binding.value
        }
      }
    }
    el.addEventListener('contextmenu', handler)
    ;(el as any).__contextmenu_handler = handler
  },
  unmounted(el: HTMLElement) {
    const handler = (el as any).__contextmenu_handler
    if (handler) {
      el.removeEventListener('contextmenu', handler)
    }
  },
}
</script>

<style scoped>
.sm-context-menu {
  position: fixed;
  z-index: 4000;
  min-width: 140px;
  background: var(--sm-bg-card);
  border: 1px solid var(--sm-border-base);
  border-radius: var(--sm-radius-md);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 4px 0;
}
.sm-context-menu__item {
  padding: 8px 16px;
  font-size: var(--sm-font-size-base);
  color: var(--sm-text-primary);
  cursor: pointer;
  white-space: nowrap;
  transition: background-color 0.15s;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sm-context-menu__item:hover:not(.is-disabled) {
  background-color: var(--sm-bg-hover);
  color: var(--sm-color-primary);
}
.sm-context-menu__item.is-disabled {
  color: var(--sm-text-placeholder);
  cursor: not-allowed;
}
.sm-context-menu__item--has-children {
  position: relative;
}
.sm-context-menu__arrow {
  margin-left: 24px;
  font-size: 12px;
}
.sm-context-menu__sub {
  position: relative;
}
.sm-context-menu__submenu {
  display: none;
  position: absolute;
  left: 100%;
  top: 0;
  min-width: 140px;
  background: var(--sm-bg-card);
  border: 1px solid var(--sm-border-base);
  border-radius: var(--sm-radius-md);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 4px 0;
}
.sm-context-menu__sub:hover .sm-context-menu__submenu {
  display: block;
}
.sm-context-menu__divider {
  height: 1px;
  margin: 4px 0;
  background: var(--sm-border-lighter);
}

/* Transition */
.sm-context-fade-enter-active,
.sm-context-fade-leave-active {
  transition: opacity 0.15s;
}
.sm-context-fade-enter-from,
.sm-context-fade-leave-to {
  opacity: 0;
}
</style>
