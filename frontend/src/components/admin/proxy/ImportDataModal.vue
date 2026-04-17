<template>
  <BaseDialog
    :show="show"
    :title="t('admin.proxies.dataImportTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <form id="import-proxy-data-form" class="space-y-4" @submit.prevent="handleImport">
      <div class="import-data-modal__description text-sm">
        {{ t('admin.proxies.dataImportHint') }}
      </div>
      <div
        class="import-data-modal__warning import-data-modal__warning-surface text-xs"
      >
        {{ t('admin.proxies.dataImportWarning') }}
      </div>

      <div>
        <label class="input-label">{{ t('admin.proxies.dataImportFile') }}</label>
        <div
          class="import-data-modal__file-picker import-data-modal__file-picker-surface flex items-center justify-between gap-3"
        >
          <div class="min-w-0">
            <div class="import-data-modal__file-name truncate text-sm">
              {{ fileName || t('admin.proxies.dataImportSelectFile') }}
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
        class="import-data-modal__result import-data-modal__result-surface space-y-2"
      >
        <div class="import-data-modal__result-title text-sm font-medium">
          {{ t('admin.proxies.dataImportResult') }}
        </div>
        <div class="import-data-modal__result-summary text-sm">
          {{ t('admin.proxies.dataImportResultSummary', result) }}
        </div>

        <div v-if="errorItems.length" class="mt-2">
          <div class="import-data-modal__errors-title text-sm font-medium">
            {{ t('admin.proxies.dataImportErrors') }}
          </div>
          <div
            class="import-data-modal__errors-list import-data-modal__errors-list-surface mt-2 overflow-auto font-mono text-xs"
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
          form="import-proxy-data-form"
          :disabled="importing"
        >
          {{ importing ? t('admin.proxies.dataImporting') : t('admin.proxies.dataImportButton') }}
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
let importRequestSequence = 0

const fileInput = ref<HTMLInputElement | null>(null)
const fileName = computed(() => file.value?.name || '')

const errorItems = computed(() => result.value?.errors || [])

const resetImportState = () => {
  file.value = null
  result.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const invalidateImportRequests = () => {
  importRequestSequence += 1
  importing.value = false
}

const isActiveImportRequest = (requestSequence: number) => (
  requestSequence === importRequestSequence && props.show
)

watch(
  () => props.show,
  () => {
    invalidateImportRequests()
    resetImportState()
  }
)

const openFilePicker = () => {
  fileInput.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  invalidateImportRequests()
  file.value = target.files?.[0] || null
  result.value = null
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
    appStore.showError(t('admin.proxies.dataImportSelectFile'))
    return
  }

  const requestSequence = ++importRequestSequence
  const sourceFile = file.value
  result.value = null
  importing.value = true
  try {
    const text = await readFileAsText(sourceFile)
    if (!isActiveImportRequest(requestSequence)) {
      return
    }
    const dataPayload = JSON.parse(text)

    const res = await adminAPI.proxies.importData({ data: dataPayload })
    if (!isActiveImportRequest(requestSequence)) {
      return
    }

    result.value = res

    const msgParams: Record<string, unknown> = {
      proxy_created: res.proxy_created,
      proxy_reused: res.proxy_reused,
      proxy_failed: res.proxy_failed
    }

    if (res.proxy_failed > 0) {
      appStore.showError(t('admin.proxies.dataImportCompletedWithErrors', msgParams))
    } else {
      appStore.showSuccess(t('admin.proxies.dataImportSuccess', msgParams))
      emit('imported')
    }
  } catch (error) {
    if (!isActiveImportRequest(requestSequence)) {
      return
    }
    if (error instanceof SyntaxError) {
      appStore.showError(t('admin.proxies.dataImportParseFailed'))
    } else {
      appStore.showError(resolveErrorMessage(error, t('admin.proxies.dataImportFailed')))
    }
  } finally {
    if (requestSequence === importRequestSequence) {
      importing.value = false
    }
  }
}
</script>

<style scoped>
.import-data-modal__description {
  color: var(--theme-page-muted);
}

.import-data-modal__warning {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-warning-rgb)) 26%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.import-data-modal__warning-surface {
  border-radius: var(--theme-import-data-modal-warning-radius);
  padding: var(--theme-import-data-modal-warning-padding);
}

.import-data-modal__file-picker {
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.import-data-modal__file-picker-surface {
  border-radius: var(--theme-import-data-modal-file-picker-radius);
  padding:
    var(--theme-import-data-modal-file-picker-padding-y)
    var(--theme-import-data-modal-file-picker-padding-x);
}

.import-data-modal__file-name,
.import-data-modal__result-title {
  color: var(--theme-page-text);
}

.import-data-modal__result {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
}

.import-data-modal__result-surface {
  border-radius: var(--theme-import-data-modal-result-radius);
  padding: var(--theme-import-data-modal-result-padding);
}

.import-data-modal__result-summary {
  color: color-mix(in srgb, var(--theme-page-text) 84%, transparent);
}

.import-data-modal__errors-title {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.import-data-modal__errors-list {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.import-data-modal__errors-list-surface {
  max-height: var(--theme-import-data-modal-errors-max-height);
  border-radius: var(--theme-import-data-modal-errors-radius);
  padding: var(--theme-import-data-modal-errors-padding);
}
</style>
