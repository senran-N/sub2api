<template>
  <div class="flex items-start gap-4">
    <!-- Preview Box -->
    <div class="flex-shrink-0">
      <div
        class="image-upload__preview flex items-center justify-center overflow-hidden border-2 border-dashed"
        :class="[previewSizeClass, { 'image-upload__preview--filled border-solid': !!modelValue }]"
      >
        <!-- SVG mode: render inline -->
        <span
          v-if="mode === 'svg' && modelValue"
          class="image-upload__content [&>svg]:h-full [&>svg]:w-full"
          :class="innerSizeClass"
          v-html="sanitizedValue"
        ></span>
        <!-- Image mode: show as img -->
        <img
          v-else-if="mode === 'image' && modelValue"
          :src="modelValue"
          alt=""
          class="h-full w-full object-contain"
        />
        <!-- Empty placeholder -->
        <svg
          v-else
          class="image-upload__placeholder"
          :class="placeholderSizeClass"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
      </div>
    </div>

    <!-- Controls -->
    <div class="flex-1 space-y-2">
      <div class="flex items-center gap-2">
        <label class="btn btn-secondary btn-sm cursor-pointer">
          <input
            :id="resolvedInputId"
            :name="resolvedInputName"
            type="file"
            :accept="acceptTypes"
            :aria-label="uploadLabel"
            class="hidden"
            @change="handleUpload"
          />
          <Icon name="upload" size="sm" class="mr-1.5" :stroke-width="2" />
          {{ uploadLabel }}
        </label>
        <button
          v-if="modelValue"
          type="button"
          class="image-upload__remove btn btn-secondary btn-sm"
          @click="$emit('update:modelValue', '')"
        >
          <Icon name="trash" size="sm" class="mr-1.5" :stroke-width="2" />
          {{ removeLabel }}
        </button>
      </div>
      <p v-if="hint" class="image-upload__hint text-xs">{{ hint }}</p>
      <p v-if="error" class="image-upload__error text-xs">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeSvg } from '@/utils/sanitize'

let imageUploadInputIdCounter = 0

const props = withDefaults(defineProps<{
  modelValue: string
  mode?: 'image' | 'svg'
  size?: 'sm' | 'md'
  uploadLabel?: string
  removeLabel?: string
  hint?: string
  inputId?: string
  inputName?: string
  maxSize?: number // bytes
}>(), {
  mode: 'image',
  size: 'md',
  uploadLabel: 'Upload',
  removeLabel: 'Remove',
  hint: '',
  inputId: undefined,
  inputName: undefined,
  maxSize: 300 * 1024,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const error = ref('')

const acceptTypes = computed(() => props.mode === 'svg' ? '.svg' : 'image/*')
const generatedInputId = `image-upload-input-${++imageUploadInputIdCounter}`
const resolvedInputId = computed(() => props.inputId ?? generatedInputId)
const resolvedInputName = computed(() => props.inputName ?? `${resolvedInputId.value}-file`)

const sanitizedValue = computed(() =>
  props.mode === 'svg' ? sanitizeSvg(props.modelValue ?? '') : ''
)

const previewSizeClass = computed(() => props.size === 'sm' ? 'h-14 w-14' : 'h-20 w-20')
const innerSizeClass = computed(() => props.size === 'sm' ? 'h-7 w-7' : 'h-12 w-12')
const placeholderSizeClass = computed(() => props.size === 'sm' ? 'h-5 w-5' : 'h-8 w-8')

function handleUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  error.value = ''

  if (!file) return

  if (props.maxSize && file.size > props.maxSize) {
    error.value = `File too large (${(file.size / 1024).toFixed(1)} KB), max ${(props.maxSize / 1024).toFixed(0)} KB`
    input.value = ''
    return
  }

  const reader = new FileReader()
  if (props.mode === 'svg') {
    reader.onload = (e) => {
      const text = e.target?.result as string
      if (text) emit('update:modelValue', text.trim())
    }
    reader.readAsText(file)
  } else {
    if (!file.type.startsWith('image/')) {
      error.value = 'Please select an image file'
      input.value = ''
      return
    }
    reader.onload = (e) => {
      emit('update:modelValue', e.target?.result as string)
    }
    reader.readAsDataURL(file)
  }

  reader.onerror = () => {
    error.value = 'Failed to read file'
  }
  input.value = ''
}
</script>

<style scoped>
.image-upload__preview {
  border-radius: calc(var(--theme-surface-radius) + 2px);
  border-color: var(--theme-input-border);
  background: color-mix(in srgb, var(--theme-input-bg) 82%, var(--theme-surface-soft) 18%);
}

.image-upload__preview--filled {
  background: var(--theme-surface);
}

.image-upload__content {
  color: var(--theme-page-text);
}

.image-upload__placeholder {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.image-upload__remove {
  color: rgb(var(--theme-danger-rgb));
}

.image-upload__remove:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 72%, var(--theme-accent-strong));
}

.image-upload__hint {
  color: var(--theme-page-muted);
}

.image-upload__error {
  color: rgb(var(--theme-danger-rgb));
}
</style>
