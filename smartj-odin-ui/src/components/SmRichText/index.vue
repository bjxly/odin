<template>
  <div
    class="sm-rich-text"
    :class="{ 'is-readonly': readonly }"
  >
    <div
      v-if="!readonly && toolbar.length"
      class="sm-rich-text__toolbar"
    >
      <template
        v-for="item in toolbar"
        :key="item"
      >
        <button
          v-if="item === 'bold'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('bold') }"
          :title="$t('cmp.rteBold')"
          @click="editorInstance?.chain().focus().toggleBold().run()"
        >
          <strong>B</strong>
        </button>
        <button
          v-else-if="item === 'italic'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('italic') }"
          :title="$t('cmp.rteItalic')"
          @click="editorInstance?.chain().focus().toggleItalic().run()"
        >
          <em>I</em>
        </button>
        <button
          v-else-if="item === 'underline'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('underline') }"
          :title="$t('cmp.rteUnderline')"
          @click="editorInstance?.chain().focus().toggleUnderline().run()"
        >
          <u>U</u>
        </button>
        <button
          v-else-if="item === 'strike'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('strike') }"
          :title="$t('cmp.rteStrikethrough')"
          @click="editorInstance?.chain().focus().toggleStrike().run()"
        >
          <s>S</s>
        </button>
        <span
          v-else-if="item === '|'"
          class="sm-rich-text__sep"
        />
        <button
          v-else-if="item === 'heading'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('heading') }"
          :title="$t('cmp.rteHeading')"
          @click="editorInstance?.chain().focus().toggleHeading({ level: 2 }).run()"
        >
          H
        </button>
        <button
          v-else-if="item === 'bulletList'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('bulletList') }"
          :title="$t('cmp.rteBulletList')"
          @click="editorInstance?.chain().focus().toggleBulletList().run()"
        >
          •
        </button>
        <button
          v-else-if="item === 'orderedList'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('orderedList') }"
          :title="$t('cmp.rteOrderedList')"
          @click="editorInstance?.chain().focus().toggleOrderedList().run()"
        >
          1.
        </button>
        <button
          v-else-if="item === 'blockquote'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('blockquote') }"
          :title="$t('cmp.rteBlockquote')"
          @click="editorInstance?.chain().focus().toggleBlockquote().run()"
        >
          "
        </button>
        <button
          v-else-if="item === 'codeBlock'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('codeBlock') }"
          :title="$t('cmp.rteCodeBlock')"
          @click="editorInstance?.chain().focus().toggleCodeBlock().run()"
        >
          &lt;/&gt;
        </button>
        <button
          v-else-if="item === 'horizontalRule'"
          class="sm-rich-text__btn"
          :title="$t('cmp.rteHorizontalRule')"
          @click="editorInstance?.chain().focus().setHorizontalRule().run()"
        >
          —
        </button>
        <span
          v-else-if="item === '|2'"
          class="sm-rich-text__sep"
        />
        <button
          v-else-if="item === 'image'"
          class="sm-rich-text__btn"
          :title="$t('cmp.rteImage')"
          @click="triggerImageUpload"
        >
          🖼
        </button>
        <button
          v-else-if="item === 'link'"
          class="sm-rich-text__btn"
          :class="{ 'is-active': editorInstance?.isActive('link') }"
          :title="$t('cmp.rteLink')"
          @click="toggleLink"
        >
          🔗
        </button>
        <button
          v-else-if="item === 'undo'"
          class="sm-rich-text__btn"
          :title="$t('cmp.rteUndo')"
          @click="editorInstance?.chain().focus().undo().run()"
        >
          ↩
        </button>
        <button
          v-else-if="item === 'redo'"
          class="sm-rich-text__btn"
          :title="$t('cmp.rteRedo')"
          @click="editorInstance?.chain().focus().redo().run()"
        >
          ↪
        </button>
      </template>
    </div>
    <EditorContent
      v-if="editorInstance"
      :editor="editorInstance"
      class="sm-rich-text__editor"
      :style="editorStyle"
    />
    <!-- Hidden file input for image upload -->
    <input
      v-if="!readonly"
      ref="fileInputRef"
      type="file"
      accept="image/*"
      style="display: none"
      @change="onImageSelected"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import Underline from '@tiptap/extension-underline'
import { ElMessage } from 'element-plus'
import { TOKEN_KEY } from '@/utils/constants'

export type ToolbarItem =
  | 'bold'
  | 'italic'
  | 'underline'
  | 'strike'
  | 'heading'
  | 'bulletList'
  | 'orderedList'
  | 'blockquote'
  | 'codeBlock'
  | 'horizontalRule'
  | 'image'
  | 'link'
  | '|'
  | '|2'
  | 'undo'
  | 'redo'

const props = withDefaults(
  defineProps<{
    modelValue?: string
    height?: number | string
    placeholder?: string
    readonly?: boolean
    toolbar?: ToolbarItem[]
    uploadAction?: string
  }>(),
  {
    modelValue: '',
    height: 400,
    placeholder: '',
    readonly: false,
    toolbar: () =>
      [
        'heading',
        '|',
        'bold',
        'italic',
        'underline',
        'strike',
        '|',
        'bulletList',
        'orderedList',
        'blockquote',
        'codeBlock',
        '|',
        'image',
        'link',
        '|2',
        'undo',
        'redo',
      ] as ToolbarItem[],
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
}>()

const { t } = useI18n()

const fileInputRef = ref<HTMLInputElement>()

const editor = useEditor({
  content: props.modelValue,
  editable: !props.readonly,
  extensions: [
    StarterKit,
    Image.configure({ inline: true }),
    Link.configure({ openOnClick: false }),
    Placeholder.configure({ placeholder: props.placeholder || t('cmp.rtePlaceholder') }),
    Underline,
  ],
  onUpdate: ({ editor: ed }) => {
    emit('update:modelValue', ed.getHTML())
  },
})

const editorInstance = computed(() => editor.value)

const editorStyle = computed(() => ({
  height: typeof props.height === 'number' ? `${props.height}px` : props.height,
}))

// Two-way sync for modelValue
watch(
  () => props.modelValue,
  (val) => {
    const currentHtml = editor.value?.getHTML()
    if (val !== currentHtml) {
      editor.value?.commands.setContent(val || '')
    }
  }
)

watch(
  () => props.readonly,
  (val) => {
    editor.value?.setEditable(!val)
  }
)

// Image upload handling
function triggerImageUpload() {
  if (props.uploadAction) {
    fileInputRef.value?.click()
  } else {
    const url = prompt(t('cmp.rteImagePrompt'))
    if (url) {
      editor.value?.chain().focus().setImage({ src: url }).run()
    }
  }
}

async function onImageSelected(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file || !props.uploadAction) return

  try {
    const formData = new FormData()
    formData.append('file', file)
    const res = await fetch(props.uploadAction, {
      method: 'POST',
      body: formData,
      headers: {
        Authorization: `Bearer ${localStorage.getItem(TOKEN_KEY) || sessionStorage.getItem(TOKEN_KEY) || ''}`,
      },
    })
    const data = await res.json()
    if (data.code === 0 && data.data) {
      const url = typeof data.data === 'string' ? data.data : data.data.url || data.data.fileUrl
      if (url) {
        editor.value?.chain().focus().setImage({ src: url }).run()
      }
    } else {
      ElMessage.warning(t('cmp.rteUploadFailed'))
    }
  } catch {
    ElMessage.warning(t('cmp.rteUploadNetworkError'))
  } finally {
    // Reset file input
    if (fileInputRef.value) fileInputRef.value.value = ''
  }
}

function toggleLink() {
  if (editor.value?.isActive('link')) {
    editor.value.chain().focus().unsetLink().run()
  } else {
    const url = prompt(t('cmp.rteLinkPrompt'), 'https://')
    if (url) {
      editor.value?.chain().focus().setLink({ href: url }).run()
    }
  }
}

onUnmounted(() => {
  editor.value?.destroy()
})
</script>

<style scoped>
.sm-rich-text {
  border: 1px solid var(--sm-border-base);
  border-radius: var(--sm-radius-md);
  overflow: hidden;
  background: var(--sm-bg-card);
}
.sm-rich-text.is-readonly {
  border-color: transparent;
  background: transparent;
}
.sm-rich-text__toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--sm-border-lighter);
  flex-wrap: wrap;
  background: var(--sm-bg-header);
}
.sm-rich-text__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 28px;
  border: none;
  border-radius: 3px;
  background: transparent;
  cursor: pointer;
  font-size: 13px;
  color: var(--sm-text-regular);
  transition:
    background 0.15s,
    color 0.15s;
}
.sm-rich-text__btn:hover {
  background: var(--sm-bg-hover);
  color: var(--sm-color-primary);
}
.sm-rich-text__btn.is-active {
  background: var(--sm-color-primary);
  color: #fff;
}
.sm-rich-text__sep {
  width: 1px;
  height: 20px;
  background: var(--sm-border-base);
  margin: 0 4px;
}
.sm-rich-text__editor {
  overflow-y: auto;
}
.sm-rich-text__editor :deep(.tiptap) {
  padding: 12px 16px;
  outline: none;
  min-height: 100px;
}
.sm-rich-text__editor :deep(.tiptap p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  color: var(--sm-text-placeholder);
  float: left;
  height: 0;
  pointer-events: none;
}
.sm-rich-text__editor :deep(.tiptap) h2 {
  font-size: 18px;
  font-weight: 600;
  margin: 8px 0;
}
.sm-rich-text__editor :deep(.tiptap) blockquote {
  border-left: 3px solid var(--sm-border-base);
  padding-left: 12px;
  color: var(--sm-text-secondary);
  margin: 8px 0;
}
.sm-rich-text__editor :deep(.tiptap) pre {
  background: var(--sm-bg-header);
  border-radius: 4px;
  padding: 12px;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 13px;
}
.sm-rich-text__editor :deep(.tiptap) img {
  max-width: 100%;
  border-radius: 4px;
}
</style>
