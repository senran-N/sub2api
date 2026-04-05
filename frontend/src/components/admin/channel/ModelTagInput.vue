<template>
  <div>
    <div class="min-h-[2.5rem] rounded-lg border border-gray-200 bg-white p-2 dark:border-dark-600 dark:bg-dark-800">
      <div class="flex flex-wrap gap-1.5">
        <span
          v-for="(model, idx) in models"
          :key="idx"
          class="inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-sm"
          :class="getPlatformTagClass(props.platform || '')"
        >
          {{ model }}
          <button
            type="button"
            class="ml-0.5 rounded-full p-0.5 hover:bg-primary-200 dark:hover:bg-primary-800"
            @click="removeModel(idx)"
          >
            <Icon name="x" size="xs" />
          </button>
        </span>
        <input
          ref="inputRef"
          v-model="inputValue"
          type="text"
          class="min-w-[120px] flex-1 border-none bg-transparent text-sm outline-none placeholder:text-gray-400 dark:text-white"
          :placeholder="models.length === 0 ? placeholder : ''"
          @keydown.enter.prevent="addModel"
          @keydown.tab.prevent="addModel"
          @keydown.delete="handleBackspace"
          @paste="handlePaste"
        />
      </div>
    </div>
    <p class="mt-1 text-xs text-gray-400">
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
