<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.dataImportTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <form id="import-data-form" class="space-y-4" @submit.prevent="handleImport">
      <div class="import-data-modal__description text-sm">
        {{ t('admin.accounts.dataImportHint') }}
      </div>
      <div
        class="import-data-modal__warning text-xs"
      >
        {{ t('admin.accounts.dataImportWarning') }}
      </div>

      <div>
        <label class="input-label">{{ t('admin.accounts.dataImportFile') }}</label>
        <div
          class="import-data-modal__file-picker flex items-center justify-between gap-3"
        >
          <div class="min-w-0">
            <div class="import-data-modal__file-name truncate text-sm">
              {{ fileName || t('admin.accounts.dataImportSelectFile') }}
            </div>
            <div class="import-data-modal__description text-xs">JSON (.json)</div>
          </div>
          <button type="button" class="btn btn-secondary shrink-0" @click="openFilePicker">
            {{ t('common.chooseFile') }}
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          accept="application/json,.json"
          @change="handleFileChange"
        />
      </div>

      <div
        v-if="result"
        class="import-data-modal__result space-y-2"
      >
        <div class="import-data-modal__result-title text-sm font-medium">
          {{ t('admin.accounts.dataImportResult') }}
        </div>
        <div class="import-data-modal__result-summary text-sm">
          {{ t('admin.accounts.dataImportResultSummary', result) }}
        </div>

        <div v-if="errorItems.length" class="mt-2">
          <div class="import-data-modal__errors-title text-sm font-medium">
            {{ t('admin.accounts.dataImportErrors') }}
          </div>
          <div
            class="import-data-modal__errors-list mt-2 overflow-auto font-mono text-xs"
          >
            <div v-for="(item, idx) in errorItems" :key="idx" class="whitespace-pre-wrap">
              {{ item.kind }} {{ item.name || item.proxy_key || '-' }} - {{ item.message }}
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" type="button" :disabled="importing" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          type="submit"
          form="import-data-form"
          :disabled="importing"
        >
          {{ importing ? t('admin.accounts.dataImporting') : t('admin.accounts.dataImportButton') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminDataImportResult } from '@/types'
import { resolveErrorMessage } from '@/utils/errorMessage'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'imported'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const importing = ref(false)
const file = ref<File | null>(null)
const result = ref<AdminDataImportResult | null>(null)

const fileInput = ref<HTMLInputElement | null>(null)
const fileName = computed(() => file.value?.name || '')

const errorItems = computed(() => result.value?.errors || [])

watch(
  () => props.show,
  (open) => {
    if (open) {
      file.value = null
      result.value = null
      if (fileInput.value) {
        fileInput.value.value = ''
      }
    }
  }
)

const openFilePicker = () => {
  fileInput.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  file.value = target.files?.[0] || null
}

const handleClose = () => {
  if (importing.value) return
  emit('close')
}

const readFileAsText = async (sourceFile: File): Promise<string> => {
  if (typeof sourceFile.text === 'function') {
    return sourceFile.text()
  }

  if (typeof sourceFile.arrayBuffer === 'function') {
    const buffer = await sourceFile.arrayBuffer()
    return new TextDecoder().decode(buffer)
  }

  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error || new Error('Failed to read file'))
    reader.readAsText(sourceFile)
  })
}

const handleImport = async () => {
  if (!file.value) {
    appStore.showError(t('admin.accounts.dataImportSelectFile'))
    return
  }

  importing.value = true
  try {
    const text = await readFileAsText(file.value)
    const dataPayload = JSON.parse(text)

    const res = await adminAPI.accounts.importData({
      data: dataPayload,
      skip_default_group_bind: true
    })

    result.value = res

    const msgParams: Record<string, unknown> = {
      account_created: res.account_created,
      account_failed: res.account_failed,
      proxy_created: res.proxy_created,
      proxy_reused: res.proxy_reused,
      proxy_failed: res.proxy_failed,
    }
    if (res.account_failed > 0 || res.proxy_failed > 0) {
      appStore.showError(t('admin.accounts.dataImportCompletedWithErrors', msgParams))
    } else {
      appStore.showSuccess(t('admin.accounts.dataImportSuccess', msgParams))
      emit('imported')
    }
  } catch (error) {
    if (error instanceof SyntaxError) {
      appStore.showError(t('admin.accounts.dataImportParseFailed'))
    } else {
      appStore.showError(resolveErrorMessage(error, t('admin.accounts.dataImportFailed')))
    }
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.import-data-modal__description {
  color: var(--theme-page-muted);
}

.import-data-modal__warning {
  padding: calc(var(--theme-markdown-block-padding) * 0.75);
  border: 1px solid color-mix(in srgb, rgb(var(--theme-warning-rgb)) 26%, var(--theme-card-border));
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.import-data-modal__file-picker {
  padding:
    calc(var(--theme-markdown-block-padding) * 0.75)
    var(--theme-markdown-block-padding);
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.import-data-modal__file-name,
.import-data-modal__result-title {
  color: var(--theme-page-text);
}

.import-data-modal__result {
  padding: var(--theme-markdown-block-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  border-radius: var(--theme-select-panel-radius);
}

.import-data-modal__result-summary {
  color: color-mix(in srgb, var(--theme-page-text) 84%, transparent);
}

.import-data-modal__errors-title {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.import-data-modal__errors-list {
  max-height: var(--theme-search-dropdown-max-height);
  padding: calc(var(--theme-markdown-block-padding) * 0.75);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}
</style>
