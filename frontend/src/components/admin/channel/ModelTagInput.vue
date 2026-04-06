<template>
  <div>
    <div class="model-tag-input">
      <div class="flex flex-wrap gap-1.5">
        <span
          v-for="(model, idx) in models"
          :key="idx"
          :class="joinClassNames(getPlatformTagClass(props.platform || ''), 'model-tag-input__tag')"
        >
          {{ model }}
          <button
            type="button"
            class="model-tag-input__remove-button"
            @click="removeModel(idx)"
          >
            <Icon name="x" size="xs" />
          </button>
        </span>
        <input
          ref="inputRef"
          v-model="inputValue"
          type="text"
          class="model-tag-input__field"
          :placeholder="models.length === 0 ? placeholder : ''"
          @keydown.enter.prevent="addModel"
          @keydown.tab.prevent="addModel"
          @keydown.delete="handleBackspace"
          @paste="handlePaste"
        />
      </div>
    </div>
    <p class="model-tag-input__hint">
      {{ t('admin.channels.form.modelInputHint') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getPlatformTagClass } from './types'

const { t } = useI18n()

const props = defineProps<{
  models: string[]
  placeholder?: string
  platform?: string
}>()

const emit = defineEmits<{
  'update:models': [models: string[]]
}>()

const inputValue = ref('')
const inputRef = ref<HTMLInputElement>()

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

function addModel() {
  const val = inputValue.value.trim()
  if (!val) return
  if (!props.models.includes(val)) {
    emit('update:models', [...props.models, val])
  }
  inputValue.value = ''
}

function removeModel(idx: number) {
  const next = [...props.models]
  next.splice(idx, 1)
  emit('update:models', next)
}

function handleBackspace() {
  if (inputValue.value === '' && props.models.length > 0) {
    removeModel(props.models.length - 1)
  }
}

function handlePaste(e: ClipboardEvent) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text') || ''
  const items = text.split(/[,\n;]+/).map(s => s.trim()).filter(Boolean)
  if (items.length === 0) return
  emit('update:models', [...new Set([...props.models, ...items])])
  inputValue.value = ''
}
</script>

<style scoped>
.model-tag-input {
  min-height: 2.5rem;
  border: 1px solid var(--theme-input-border);
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: var(--theme-input-bg);
  padding: 0.5rem;
}

.model-tag-input__tag {
  align-items: center;
  gap: 0.35rem;
  font-size: 0.875rem;
  min-width: 0;
}

.model-tag-input__remove-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  color: inherit;
  margin-left: 0.125rem;
  padding: 0.125rem;
  transition: background-color 0.18s ease;
}

.model-tag-input__remove-button:hover,
.model-tag-input__remove-button:focus-visible {
  background: color-mix(in srgb, var(--theme-accent) 16%, transparent);
  outline: none;
}

.model-tag-input__field {
  min-width: 120px;
  flex: 1 1 0%;
  border: none;
  background: transparent;
  color: var(--theme-input-text);
  font-size: 0.875rem;
  outline: none;
}

.model-tag-input__field::placeholder {
  color: var(--theme-input-placeholder);
}

.model-tag-input__hint {
  margin-top: 0.25rem;
  color: var(--theme-page-muted);
  font-size: 0.75rem;
}
</style>
