<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.syncFromCrsTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <!-- Step 1: Input credentials -->
    <form
      v-if="currentStep === 'input'"
      id="sync-from-crs-form"
      class="space-y-4"
      @submit.prevent="handlePreview"
    >
      <div class="sync-from-crs-modal__description">
        {{ t('admin.accounts.syncFromCrsDesc') }}
      </div>
      <div class="sync-from-crs-modal__notice sync-from-crs-modal__notice--neutral">
        {{ t('admin.accounts.crsUpdateBehaviorNote') }}
      </div>
      <div class="sync-from-crs-modal__notice sync-from-crs-modal__notice--warning">
        {{ t('admin.accounts.crsVersionRequirement') }}
      </div>

      <div class="grid grid-cols-1 gap-4">
        <div>
          <label for="crs-base-url" class="input-label">{{ t('admin.accounts.crsBaseUrl') }}</label>
          <input
            id="crs-base-url"
            v-model="form.base_url"
            type="text"
            class="input"
            required
            :placeholder="t('admin.accounts.crsBaseUrlPlaceholder')"
          />
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label for="crs-username" class="input-label">{{ t('admin.accounts.crsUsername') }}</label>
            <input id="crs-username" v-model="form.username" type="text" class="input" required autocomplete="username" />
          </div>
          <div>
            <label for="crs-password" class="input-label">{{ t('admin.accounts.crsPassword') }}</label>
            <input
              id="crs-password"
              v-model="form.password"
              type="password"
              class="input"
              required
              autocomplete="current-password"
            />
          </div>
        </div>

        <label class="sync-from-crs-modal__checkbox">
          <input
            v-model="form.sync_proxies"
            type="checkbox"
            class="sync-from-crs-modal__checkbox-input"
          />
          {{ t('admin.accounts.syncProxies') }}
        </label>
      </div>
    </form>

    <!-- Step 2: Preview & select -->
    <div v-else-if="currentStep === 'preview' && previewResult" class="space-y-4">
      <!-- Existing accounts (read-only info) -->
      <div
        v-if="previewResult.existing_accounts.length"
        class="sync-from-crs-modal__section sync-from-crs-modal__section--muted"
      >
        <div class="sync-from-crs-modal__section-title">
          {{ t('admin.accounts.crsExistingAccounts') }}
          <span class="sync-from-crs-modal__section-count">({{ previewResult.existing_accounts.length }})</span>
        </div>
        <div class="sync-from-crs-modal__scroll-list sync-from-crs-modal__scroll-list--compact">
          <div
            v-for="acc in previewResult.existing_accounts"
            :key="acc.crs_account_id"
            class="sync-from-crs-modal__account-row"
          >
            <span :class="getAccountChipClasses('existing')">{{ acc.platform }} / {{ acc.type }}</span>
            <span class="sync-from-crs-modal__account-name">{{ acc.name }}</span>
          </div>
        </div>
      </div>

      <!-- New accounts (selectable) -->
      <div v-if="previewResult.new_accounts.length">
        <div class="mb-2 flex items-center justify-between">
          <div class="sync-from-crs-modal__section-title">
            {{ t('admin.accounts.crsNewAccounts') }}
            <span class="sync-from-crs-modal__section-count">({{ previewResult.new_accounts.length }})</span>
          </div>
          <div class="flex gap-2">
            <button
              type="button"
              class="sync-from-crs-modal__link-button"
              @click="selectAll"
            >{{ t('admin.accounts.crsSelectAll') }}</button>
            <button
              type="button"
              class="sync-from-crs-modal__secondary-link"
              @click="selectNone"
            >{{ t('admin.accounts.crsSelectNone') }}</button>
          </div>
        </div>
        <div class="sync-from-crs-modal__scroll-list">
          <label
            v-for="acc in previewResult.new_accounts"
            :key="acc.crs_account_id"
            class="sync-from-crs-modal__selectable-row"
          >
            <input
              type="checkbox"
              :checked="selectedIds.has(acc.crs_account_id)"
              class="sync-from-crs-modal__checkbox-input"
              @change="toggleSelect(acc.crs_account_id)"
            />
            <span :class="getAccountChipClasses('new')">{{ acc.platform }} / {{ acc.type }}</span>
            <span class="sync-from-crs-modal__account-name sync-from-crs-modal__account-name--strong">{{ acc.name }}</span>
          </label>
        </div>
        <div class="sync-from-crs-modal__selection-count">
          {{ t('admin.accounts.crsSelectedCount', { count: selectedIds.size }) }}
        </div>
      </div>

      <!-- Sync options summary -->
      <div class="sync-from-crs-modal__summary-row">
        <span>{{ t('admin.accounts.syncProxies') }}:</span>
        <span :class="getProxySyncStateClasses()">
          {{ form.sync_proxies ? t('common.yes') : t('common.no') }}
        </span>
      </div>

      <!-- No new accounts -->
      <div
        v-if="!previewResult.new_accounts.length"
        class="sync-from-crs-modal__empty-state"
      >
        {{ t('admin.accounts.crsNoNewAccounts') }}
        <span v-if="previewResult.existing_accounts.length">
          {{ t('admin.accounts.crsWillUpdate', { count: previewResult.existing_accounts.length }) }}
        </span>
      </div>
    </div>

    <!-- Step 3: Result -->
    <div v-else-if="currentStep === 'result' && result" class="space-y-4">
      <div class="sync-from-crs-modal__result-card">
        <div class="sync-from-crs-modal__result-title">
          {{ t('admin.accounts.syncResult') }}
        </div>
        <div class="sync-from-crs-modal__result-summary">
          {{ t('admin.accounts.syncResultSummary', result) }}
        </div>

        <div v-if="errorItems.length" class="mt-2">
          <div class="sync-from-crs-modal__error-title">
            {{ t('admin.accounts.syncErrors') }}
          </div>
          <div class="sync-from-crs-modal__error-log">
            <div v-for="(item, idx) in errorItems" :key="idx" class="whitespace-pre-wrap">
              {{ item.kind }} {{ item.crs_account_id }} — {{ item.action
              }}{{ item.error ? `: ${item.error}` : '' }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <!-- Step 1: Input -->
        <template v-if="currentStep === 'input'">
          <button
            class="btn btn-secondary"
            type="button"
            :disabled="previewing"
            @click="handleClose"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            class="btn btn-primary"
            type="submit"
            form="sync-from-crs-form"
            :disabled="previewing"
          >
            {{ previewing ? t('admin.accounts.crsPreviewing') : t('admin.accounts.crsPreview') }}
          </button>
        </template>

        <!-- Step 2: Preview -->
        <template v-else-if="currentStep === 'preview'">
          <button
            class="btn btn-secondary"
            type="button"
            :disabled="syncing"
            @click="handleBack"
          >
            {{ t('admin.accounts.crsBack') }}
          </button>
          <button
            class="btn btn-primary"
            type="button"
            :disabled="syncing || hasNewButNoneSelected"
            @click="handleSync"
          >
            {{ syncing ? t('admin.accounts.syncing') : t('admin.accounts.syncNow') }}
          </button>
        </template>

        <!-- Step 3: Result -->
        <template v-else-if="currentStep === 'result'">
          <button class="btn btn-secondary" type="button" @click="handleClose">
            {{ t('common.close') }}
          </button>
        </template>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { PreviewFromCRSResult } from '@/api/admin/accounts'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'synced'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

type Step = 'input' | 'preview' | 'result'
const currentStep = ref<Step>('input')
const previewing = ref(false)
const syncing = ref(false)
const previewResult = ref<PreviewFromCRSResult | null>(null)
const selectedIds = ref(new Set<string>())
const result = ref<Awaited<ReturnType<typeof adminAPI.accounts.syncFromCrs>> | null>(null)
let syncFromCrsRequestSequence = 0

const form = reactive({
  base_url: '',
  username: '',
  password: '',
  sync_proxies: true
})

const hasNewButNoneSelected = computed(() => {
  if (!previewResult.value) return false
  return previewResult.value.new_accounts.length > 0 && selectedIds.value.size === 0
})

const errorItems = computed(() => {
  if (!result.value?.items) return []
  return result.value.items.filter(
    (i) => i.action === 'failed' || (i.action === 'skipped' && i.error !== 'not selected')
  )
})

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getAccountChipClasses = (kind: 'existing' | 'new') => {
  return joinClassNames(
    'theme-chip theme-chip--compact inline-flex text-[10px] font-semibold',
    kind === 'new' ? 'theme-chip--success' : 'theme-chip--info'
  )
}

const getProxySyncStateClasses = () => {
  return joinClassNames(
    'sync-from-crs-modal__summary-value',
    form.sync_proxies ? 'sync-from-crs-modal__summary-value--enabled' : 'sync-from-crs-modal__summary-value--disabled'
  )
}

const getErrorMessage = (error: unknown, fallbackMessage: string) => {
  return error instanceof Error && error.message ? error.message : fallbackMessage
}

const resetSyncFromCrsState = () => {
  currentStep.value = 'input'
  previewResult.value = null
  selectedIds.value = new Set()
  result.value = null
  form.base_url = ''
  form.username = ''
  form.password = ''
  form.sync_proxies = true
}

const invalidateSyncFromCrsRequests = () => {
  syncFromCrsRequestSequence += 1
  previewing.value = false
  syncing.value = false
}

const isActiveSyncFromCrsRequest = (requestSequence: number) => (
  requestSequence === syncFromCrsRequestSequence && props.show
)

watch(
  () => props.show,
  () => {
    invalidateSyncFromCrsRequests()
    resetSyncFromCrsState()
  }
)

const handleClose = () => {
  if (syncing.value || previewing.value) {
    return
  }
  emit('close')
}

const handleBack = () => {
  currentStep.value = 'input'
  previewResult.value = null
  selectedIds.value = new Set()
}

const selectAll = () => {
  if (!previewResult.value) return
  selectedIds.value = new Set(previewResult.value.new_accounts.map((a) => a.crs_account_id))
}

const selectNone = () => {
  selectedIds.value = new Set()
}

const toggleSelect = (id: string) => {
  const s = new Set(selectedIds.value)
  if (s.has(id)) {
    s.delete(id)
  } else {
    s.add(id)
  }
  selectedIds.value = s
}

const handlePreview = async () => {
  if (!form.base_url.trim() || !form.username.trim() || !form.password.trim()) {
    appStore.showError(t('admin.accounts.syncMissingFields'))
    return
  }

  const requestSequence = ++syncFromCrsRequestSequence
  const payload = {
    base_url: form.base_url.trim(),
    username: form.username.trim(),
    password: form.password
  }
  previewing.value = true
  try {
    const res = await adminAPI.accounts.previewFromCrs(payload)
    if (!isActiveSyncFromCrsRequest(requestSequence)) {
      return
    }
    previewResult.value = res
    // Auto-select all new accounts
    selectedIds.value = new Set(res.new_accounts.map((a) => a.crs_account_id))
    currentStep.value = 'preview'
  } catch (error) {
    if (!isActiveSyncFromCrsRequest(requestSequence)) {
      return
    }
    appStore.showError(getErrorMessage(error, t('admin.accounts.crsPreviewFailed')))
  } finally {
    if (requestSequence === syncFromCrsRequestSequence) {
      previewing.value = false
    }
  }
}

const handleSync = async () => {
  if (!form.base_url.trim() || !form.username.trim() || !form.password.trim()) {
    appStore.showError(t('admin.accounts.syncMissingFields'))
    return
  }

  const requestSequence = ++syncFromCrsRequestSequence
  const payload = {
    base_url: form.base_url.trim(),
    username: form.username.trim(),
    password: form.password,
    sync_proxies: form.sync_proxies,
    selected_account_ids: [...selectedIds.value]
  }
  syncing.value = true
  try {
    const res = await adminAPI.accounts.syncFromCrs(payload)
    if (!isActiveSyncFromCrsRequest(requestSequence)) {
      return
    }
    result.value = res
    currentStep.value = 'result'

    if (res.failed > 0) {
      appStore.showError(t('admin.accounts.syncCompletedWithErrors', res))
    } else {
      appStore.showSuccess(t('admin.accounts.syncCompleted', res))
    }
    emit('synced')
  } catch (error) {
    if (!isActiveSyncFromCrsRequest(requestSequence)) {
      return
    }
    appStore.showError(getErrorMessage(error, t('admin.accounts.syncFailed')))
  } finally {
    if (requestSequence === syncFromCrsRequestSequence) {
      syncing.value = false
    }
  }
}
</script>

<style scoped>
.sync-from-crs-modal__description {
  color: var(--theme-page-muted);
  font-size: 0.875rem;
}

.sync-from-crs-modal__notice,
.sync-from-crs-modal__section,
.sync-from-crs-modal__result-card,
.sync-from-crs-modal__empty-state,
.sync-from-crs-modal__scroll-list {
  border-radius: calc(var(--theme-button-radius) + 2px);
}

.sync-from-crs-modal__notice {
  padding: 0.75rem;
  font-size: 0.75rem;
}

.sync-from-crs-modal__notice--neutral,
.sync-from-crs-modal__section--muted,
.sync-from-crs-modal__empty-state {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.sync-from-crs-modal__notice--warning {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-warning-rgb)) 34%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 9%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 74%, var(--theme-page-text));
}

.sync-from-crs-modal__checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--theme-page-text);
  font-size: 0.875rem;
}

.sync-from-crs-modal__checkbox-input {
  border: 1px solid var(--theme-input-border);
  border-radius: 0.375rem;
  accent-color: var(--theme-accent);
}

.sync-from-crs-modal__section {
  padding: 0.75rem;
}

.sync-from-crs-modal__section-title,
.sync-from-crs-modal__result-title {
  color: var(--theme-page-text);
  font-size: 0.875rem;
  font-weight: 600;
}

.sync-from-crs-modal__section-count,
.sync-from-crs-modal__selection-count,
.sync-from-crs-modal__summary-row {
  color: var(--theme-page-muted);
  font-size: 0.75rem;
}

.sync-from-crs-modal__scroll-list {
  max-height: 12rem;
  overflow: auto;
  border: 1px solid var(--theme-card-border);
  background: var(--theme-surface);
  padding: 0.5rem;
}

.sync-from-crs-modal__scroll-list--compact {
  max-height: 8rem;
  border: none;
  background: transparent;
  padding: 0;
}

.sync-from-crs-modal__account-row,
.sync-from-crs-modal__selectable-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: calc(var(--theme-button-radius) - 2px);
}

.sync-from-crs-modal__account-row {
  padding: 0.125rem 0;
}

.sync-from-crs-modal__selectable-row {
  cursor: pointer;
  padding: 0.5rem;
  transition: background-color 0.18s ease;
}

.sync-from-crs-modal__selectable-row:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 60%, var(--theme-surface));
}

.sync-from-crs-modal__account-name {
  min-width: 0;
  flex: 1 1 0%;
  color: var(--theme-page-muted);
  font-size: 0.8125rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sync-from-crs-modal__account-name--strong,
.sync-from-crs-modal__result-summary {
  color: var(--theme-page-text);
}

.sync-from-crs-modal__link-button,
.sync-from-crs-modal__secondary-link {
  font-size: 0.75rem;
  font-weight: 600;
  transition: color 0.18s ease;
}

.sync-from-crs-modal__link-button {
  color: var(--theme-accent);
}

.sync-from-crs-modal__secondary-link {
  color: var(--theme-page-muted);
}

.sync-from-crs-modal__link-button:hover,
.sync-from-crs-modal__link-button:focus-visible {
  color: color-mix(in srgb, var(--theme-accent) 74%, var(--theme-accent-strong));
  outline: none;
}

.sync-from-crs-modal__secondary-link:hover,
.sync-from-crs-modal__secondary-link:focus-visible {
  color: var(--theme-page-text);
  outline: none;
}

.sync-from-crs-modal__selection-count {
  margin-top: 0.25rem;
}

.sync-from-crs-modal__summary-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.sync-from-crs-modal__summary-value {
  font-weight: 600;
}

.sync-from-crs-modal__summary-value--enabled {
  color: rgb(var(--theme-success-rgb));
}

.sync-from-crs-modal__summary-value--disabled {
  color: var(--theme-page-muted);
}

.sync-from-crs-modal__empty-state,
.sync-from-crs-modal__result-card {
  padding: 1rem;
}

.sync-from-crs-modal__empty-state {
  text-align: center;
  font-size: 0.875rem;
}

.sync-from-crs-modal__result-card {
  border: 1px solid var(--theme-card-border);
  background: var(--theme-surface);
}

.sync-from-crs-modal__result-summary {
  font-size: 0.875rem;
}

.sync-from-crs-modal__error-title {
  color: rgb(var(--theme-danger-rgb));
  font-size: 0.875rem;
  font-weight: 600;
}

.sync-from-crs-modal__error-log {
  max-height: 12rem;
  overflow: auto;
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
  color: var(--theme-page-text);
  font-family: var(--theme-font-mono);
  font-size: 0.75rem;
  margin-top: 0.5rem;
  padding: 0.75rem;
}
</style>
