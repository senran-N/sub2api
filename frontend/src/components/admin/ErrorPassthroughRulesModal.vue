<template>
  <BaseDialog
    :show="show"
    :title="t('admin.errorPassthrough.title')"
    width="extra-wide"
    @close="$emit('close')"
  >
    <div class="error-passthrough-rules-modal__content">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <p class="error-passthrough-rules-modal__subtitle text-sm">
          {{ t('admin.errorPassthrough.description') }}
        </p>
        <button @click="showCreateModal = true" class="btn btn-primary btn-sm">
          <Icon name="plus" size="sm" class="mr-1" />
          {{ t('admin.errorPassthrough.createRule') }}
        </button>
      </div>

      <!-- Rules Table -->
      <div v-if="loading" class="error-passthrough-rules-modal__status-state">
        <Icon name="refresh" size="lg" class="error-passthrough-rules-modal__status-icon animate-spin" />
      </div>

      <div v-else-if="rules.length === 0" class="error-passthrough-rules-modal__empty-state">
        <div class="error-passthrough-rules-modal__empty-icon-wrap">
          <Icon name="shield" size="lg" class="error-passthrough-rules-modal__empty-icon" />
        </div>
        <h4 class="error-passthrough-rules-modal__empty-title">
          {{ t('admin.errorPassthrough.noRules') }}
        </h4>
        <p class="error-passthrough-rules-modal__subtitle text-sm">
          {{ t('admin.errorPassthrough.createFirstRule') }}
        </p>
      </div>

      <div v-else class="error-passthrough-rules-modal__table-shell">
        <table class="error-passthrough-rules-modal__table min-w-full">
          <thead class="error-passthrough-rules-modal__table-head sticky top-0">
            <tr>
              <th class="error-passthrough-rules-modal__table-header">
                {{ t('admin.errorPassthrough.columns.priority') }}
              </th>
              <th class="error-passthrough-rules-modal__table-header">
                {{ t('admin.errorPassthrough.columns.name') }}
              </th>
              <th class="error-passthrough-rules-modal__table-header">
                {{ t('admin.errorPassthrough.columns.conditions') }}
              </th>
              <th class="error-passthrough-rules-modal__table-header">
                {{ t('admin.errorPassthrough.columns.platforms') }}
              </th>
              <th class="error-passthrough-rules-modal__table-header">
                {{ t('admin.errorPassthrough.columns.behavior') }}
              </th>
              <th class="error-passthrough-rules-modal__table-header">
                {{ t('admin.errorPassthrough.columns.status') }}
              </th>
              <th class="error-passthrough-rules-modal__table-header">
                {{ t('admin.errorPassthrough.columns.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="error-passthrough-rules-modal__table-body">
            <tr v-for="rule in rules" :key="rule.id" class="error-passthrough-rules-modal__table-row">
              <td class="error-passthrough-rules-modal__table-cell error-passthrough-rules-modal__table-cell--nowrap">
                <span class="theme-chip theme-chip--compact theme-chip--neutral inline-flex h-5 w-5 items-center justify-center">
                  {{ rule.priority }}
                </span>
              </td>
              <td class="error-passthrough-rules-modal__table-cell">
                <div class="error-passthrough-rules-modal__rule-name text-sm font-medium">{{ rule.name }}</div>
                <div v-if="rule.description" class="error-passthrough-rules-modal__rule-description mt-0.5 max-w-xs truncate text-xs">
                  {{ rule.description }}
                </div>
              </td>
              <td class="error-passthrough-rules-modal__table-cell">
                <div class="flex max-w-48 flex-wrap gap-1">
                  <span
                    v-for="code in rule.error_codes.slice(0, 3)"
                    :key="code"
                    class="theme-chip theme-chip--compact theme-chip--danger"
                  >
                    {{ code }}
                  </span>
                  <span
                    v-if="rule.error_codes.length > 3"
                    class="error-passthrough-rules-modal__count text-xs"
                  >
                    +{{ rule.error_codes.length - 3 }}
                  </span>
                  <span
                    v-for="keyword in rule.keywords.slice(0, 1)"
                    :key="keyword"
                    class="theme-chip theme-chip--compact theme-chip--neutral"
                  >
                    "{{ keyword.length > 10 ? keyword.substring(0, 10) + '...' : keyword }}"
                  </span>
                  <span
                    v-if="rule.keywords.length > 1"
                    class="error-passthrough-rules-modal__count text-xs"
                  >
                    +{{ rule.keywords.length - 1 }}
                  </span>
                </div>
                <div class="error-passthrough-rules-modal__rule-description mt-0.5 text-xs">
                  {{ t('admin.errorPassthrough.matchMode.' + rule.match_mode) }}
                </div>
              </td>
              <td class="error-passthrough-rules-modal__table-cell">
                <div v-if="rule.platforms.length === 0" class="error-passthrough-rules-modal__rule-description text-xs">
                  {{ t('admin.errorPassthrough.allPlatforms') }}
                </div>
                <div v-else class="flex flex-wrap gap-1">
                  <span
                    v-for="platform in rule.platforms.slice(0, 2)"
                    :key="platform"
                    class="theme-chip theme-chip--compact theme-chip--accent"
                  >
                    {{ platform }}
                  </span>
                  <span v-if="rule.platforms.length > 2" class="error-passthrough-rules-modal__count text-xs">
                    +{{ rule.platforms.length - 2 }}
                  </span>
                </div>
              </td>
              <td class="error-passthrough-rules-modal__table-cell">
                <div class="space-y-0.5 text-xs">
                  <div class="flex items-center gap-1">
                    <Icon
                      :name="rule.passthrough_code ? 'checkCircle' : 'xCircle'"
                      size="xs"
                      :class="getBehaviorIconClasses(rule.passthrough_code, 'success')"
                    />
                    <span class="error-passthrough-rules-modal__behavior-text">
                      {{ t('admin.errorPassthrough.code') }}:
                      {{ rule.passthrough_code ? t('admin.errorPassthrough.passthrough') : (rule.response_code || '-') }}
                    </span>
                  </div>
                  <div class="flex items-center gap-1">
                    <Icon
                      :name="rule.passthrough_body ? 'checkCircle' : 'xCircle'"
                      size="xs"
                      :class="getBehaviorIconClasses(rule.passthrough_body, 'success')"
                    />
                    <span class="error-passthrough-rules-modal__behavior-text">
                      {{ t('admin.errorPassthrough.body') }}:
                      {{ rule.passthrough_body ? t('admin.errorPassthrough.passthrough') : t('admin.errorPassthrough.custom') }}
                    </span>
                  </div>
                  <div v-if="rule.skip_monitoring" class="flex items-center gap-1">
                    <Icon
                      name="checkCircle"
                      size="xs"
                      :class="getBehaviorIconClasses(true, 'warning')"
                    />
                    <span class="error-passthrough-rules-modal__behavior-text">
                      {{ t('admin.errorPassthrough.skipMonitoring') }}
                    </span>
                  </div>
                </div>
              </td>
              <td class="error-passthrough-rules-modal__table-cell">
                <button
                  @click="toggleEnabled(rule)"
                  :class="getRuleToggleTrackClasses(rule.enabled)"
                >
                  <span
                    :class="[
                      'error-passthrough-rules-modal__toggle-thumb pointer-events-none inline-block h-3 w-3 transform rounded-full ring-0 transition duration-200 ease-in-out',
                      rule.enabled ? 'translate-x-3' : 'translate-x-0'
                    ]"
                  />
                </button>
              </td>
              <td class="error-passthrough-rules-modal__table-cell">
                <div class="flex items-center gap-1">
                  <button
                    @click="handleEdit(rule)"
                    :class="getActionButtonClasses('info')"
                    :title="t('common.edit')"
                  >
                    <Icon name="edit" size="sm" />
                  </button>
                  <button
                    @click="handleDelete(rule)"
                    :class="getActionButtonClasses('danger')"
                    :title="t('common.delete')"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button @click="$emit('close')" class="btn btn-secondary">
          {{ t('common.close') }}
        </button>
      </div>
    </template>

    <!-- Create/Edit Modal -->
    <BaseDialog
      :show="showCreateModal || showEditModal"
      :title="showEditModal ? t('admin.errorPassthrough.editRule') : t('admin.errorPassthrough.createRule')"
      width="wide"
      @close="closeFormModal"
    >
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <!-- Basic Info -->
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
          <div>
            <label class="input-label">{{ t('admin.errorPassthrough.form.name') }}</label>
            <input
              v-model="form.name"
              type="text"
              required
              class="input"
              :placeholder="t('admin.errorPassthrough.form.namePlaceholder')"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.errorPassthrough.form.priority') }}</label>
            <input
              v-model.number="form.priority"
              type="number"
              min="0"
              class="input"
            />
            <p class="input-hint">{{ t('admin.errorPassthrough.form.priorityHint') }}</p>
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.errorPassthrough.form.description') }}</label>
          <input
            v-model="form.description"
            type="text"
            class="input"
            :placeholder="t('admin.errorPassthrough.form.descriptionPlaceholder')"
          />
        </div>

        <!-- Match Conditions -->
        <div class="error-passthrough-rules-modal__form-section">
          <h4 class="error-passthrough-rules-modal__form-section-title mb-2 text-sm font-medium">
            {{ t('admin.errorPassthrough.form.matchConditions') }}
          </h4>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label class="input-label text-xs">{{ t('admin.errorPassthrough.form.errorCodes') }}</label>
              <input
                v-model="errorCodesInput"
                type="text"
                class="input text-sm"
                :placeholder="t('admin.errorPassthrough.form.errorCodesPlaceholder')"
              />
              <p class="input-hint text-xs">{{ t('admin.errorPassthrough.form.errorCodesHint') }}</p>
            </div>
            <div>
              <label class="input-label text-xs">{{ t('admin.errorPassthrough.form.keywords') }}</label>
              <textarea
                v-model="keywordsInput"
                rows="2"
                class="input font-mono text-xs"
                :placeholder="t('admin.errorPassthrough.form.keywordsPlaceholder')"
              />
              <p class="input-hint text-xs">{{ t('admin.errorPassthrough.form.keywordsHint') }}</p>
            </div>
          </div>

          <div class="mt-3">
            <label class="input-label text-xs">{{ t('admin.errorPassthrough.form.matchMode') }}</label>
            <div class="mt-1 space-y-2">
              <label
                v-for="option in matchModeOptions"
                :key="option.value"
                class="error-passthrough-rules-modal__option-row flex cursor-pointer items-start gap-2"
              >
                <input
                  type="radio"
                  :value="option.value"
                  v-model="form.match_mode"
                  class="error-passthrough-rules-modal__radio-input mt-0.5 h-3.5 w-3.5"
                />
                <div class="flex-1">
                  <span class="error-passthrough-rules-modal__option-label text-xs font-medium">{{ option.label }}</span>
                  <p class="error-passthrough-rules-modal__option-description text-xs">{{ option.description }}</p>
                </div>
              </label>
            </div>
          </div>

          <div class="mt-3">
            <label class="input-label text-xs">{{ t('admin.errorPassthrough.form.platforms') }}</label>
            <div class="flex flex-wrap gap-3">
              <label
                v-for="platform in platformOptions"
                :key="platform.value"
                class="error-passthrough-rules-modal__checkbox-row inline-flex items-center gap-1.5"
              >
                <input
                  type="checkbox"
                  :value="platform.value"
                  v-model="form.platforms"
                  class="error-passthrough-rules-modal__checkbox-input h-3.5 w-3.5 rounded"
                />
                <span class="error-passthrough-rules-modal__checkbox-label text-xs">{{ platform.label }}</span>
              </label>
            </div>
            <p class="input-hint mt-1 text-xs">{{ t('admin.errorPassthrough.form.platformsHint') }}</p>
          </div>
        </div>

        <!-- Response Behavior -->
        <div class="error-passthrough-rules-modal__form-section">
          <h4 class="error-passthrough-rules-modal__form-section-title mb-2 text-sm font-medium">
            {{ t('admin.errorPassthrough.form.responseBehavior') }}
          </h4>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label class="error-passthrough-rules-modal__checkbox-row flex items-center gap-1.5">
                <input
                  type="checkbox"
                  v-model="form.passthrough_code"
                  class="error-passthrough-rules-modal__checkbox-input h-3.5 w-3.5 rounded"
                />
                <span class="error-passthrough-rules-modal__checkbox-label text-xs font-medium">
                  {{ t('admin.errorPassthrough.form.passthroughCode') }}
                </span>
              </label>
              <div v-if="!form.passthrough_code" class="mt-2">
                <label class="input-label text-xs">{{ t('admin.errorPassthrough.form.responseCode') }}</label>
                <input
                  v-model.number="form.response_code"
                  type="number"
                  min="100"
                  max="599"
                  class="input text-sm"
                  placeholder="422"
                />
              </div>
            </div>
            <div>
              <label class="error-passthrough-rules-modal__checkbox-row flex items-center gap-1.5">
                <input
                  type="checkbox"
                  v-model="form.passthrough_body"
                  class="error-passthrough-rules-modal__checkbox-input h-3.5 w-3.5 rounded"
                />
                <span class="error-passthrough-rules-modal__checkbox-label text-xs font-medium">
                  {{ t('admin.errorPassthrough.form.passthroughBody') }}
                </span>
              </label>
              <div v-if="!form.passthrough_body" class="mt-2">
                <label class="input-label text-xs">{{ t('admin.errorPassthrough.form.customMessage') }}</label>
                <input
                  v-model="form.custom_message"
                  type="text"
                  class="input text-sm"
                  :placeholder="t('admin.errorPassthrough.form.customMessagePlaceholder')"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Skip Monitoring -->
        <div class="error-passthrough-rules-modal__checkbox-row flex items-center gap-1.5">
          <input
            type="checkbox"
            v-model="form.skip_monitoring"
            class="error-passthrough-rules-modal__checkbox-input error-passthrough-rules-modal__checkbox-input--warning h-3.5 w-3.5 rounded"
          />
          <span class="error-passthrough-rules-modal__checkbox-label text-xs font-medium">
            {{ t('admin.errorPassthrough.form.skipMonitoring') }}
          </span>
        </div>
        <p class="input-hint text-xs -mt-3">{{ t('admin.errorPassthrough.form.skipMonitoringHint') }}</p>

        <!-- Enabled -->
        <div class="error-passthrough-rules-modal__checkbox-row flex items-center gap-1.5">
          <input
            type="checkbox"
            v-model="form.enabled"
            class="error-passthrough-rules-modal__checkbox-input h-3.5 w-3.5 rounded"
          />
          <span class="error-passthrough-rules-modal__checkbox-label text-xs font-medium">
            {{ t('admin.errorPassthrough.form.enabled') }}
          </span>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button @click="closeFormModal" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button @click="handleSubmit" :disabled="submitting" class="btn btn-primary">
            <Icon v-if="submitting" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ showEditModal ? t('common.update') : t('common.create') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.errorPassthrough.deleteRule')"
      :message="t('admin.errorPassthrough.deleteConfirm', { name: deletingRule?.name })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { ErrorPassthroughRule } from '@/api/admin/errorPassthrough'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

// eslint-disable-next-line @typescript-eslint/no-unused-vars
void emit // suppress unused warning - emit is used via $emit in template

const { t } = useI18n()
const appStore = useAppStore()

const rules = ref<ErrorPassthroughRule[]>([])
const loading = ref(false)
const submitting = ref(false)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const editingRule = ref<ErrorPassthroughRule | null>(null)
const deletingRule = ref<ErrorPassthroughRule | null>(null)

// Form inputs for arrays
const errorCodesInput = ref('')
const keywordsInput = ref('')

const form = reactive({
  name: '',
  enabled: true,
  priority: 0,
  match_mode: 'any' as 'any' | 'all',
  platforms: [] as string[],
  passthrough_code: true,
  response_code: null as number | null,
  passthrough_body: true,
  custom_message: null as string | null,
  skip_monitoring: false,
  description: null as string | null
})

const matchModeOptions = computed(() => [
  { value: 'any', label: t('admin.errorPassthrough.matchMode.any'), description: t('admin.errorPassthrough.matchMode.anyHint') },
  { value: 'all', label: t('admin.errorPassthrough.matchMode.all'), description: t('admin.errorPassthrough.matchMode.allHint') }
])

const platformOptions = [
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'antigravity', label: 'Antigravity' }
]

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getBehaviorIconClasses = (enabled: boolean, tone: 'success' | 'warning' = 'success') => {
  if (!enabled) {
    return 'error-passthrough-rules-modal__behavior-icon error-passthrough-rules-modal__behavior-icon--inactive'
  }

  return joinClassNames(
    'error-passthrough-rules-modal__behavior-icon',
    tone === 'warning'
      ? 'error-passthrough-rules-modal__behavior-icon--warning'
      : 'error-passthrough-rules-modal__behavior-icon--success'
  )
}

const getRuleToggleTrackClasses = (enabled: boolean) => {
  return joinClassNames(
    'error-passthrough-rules-modal__toggle-track',
    enabled
      ? 'error-passthrough-rules-modal__toggle-track--enabled'
      : 'error-passthrough-rules-modal__toggle-track--disabled'
  )
}

const getActionButtonClasses = (tone: 'info' | 'danger') => {
  return joinClassNames(
    'error-passthrough-rules-modal__action-button',
    tone === 'info'
      ? 'error-passthrough-rules-modal__action-button--info'
      : 'error-passthrough-rules-modal__action-button--danger'
  )
}

// Load rules when dialog opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    loadRules()
  }
})

const loadRules = async () => {
  loading.value = true
  try {
    rules.value = await adminAPI.errorPassthrough.list()
  } catch (error) {
    appStore.showError(t('admin.errorPassthrough.failedToLoad'))
    console.error('Error loading rules:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.enabled = true
  form.priority = 0
  form.match_mode = 'any'
  form.platforms = []
  form.passthrough_code = true
  form.response_code = null
  form.passthrough_body = true
  form.custom_message = null
  form.skip_monitoring = false
  form.description = null
  errorCodesInput.value = ''
  keywordsInput.value = ''
}

const closeFormModal = () => {
  showCreateModal.value = false
  showEditModal.value = false
  editingRule.value = null
  resetForm()
}

const handleEdit = (rule: ErrorPassthroughRule) => {
  editingRule.value = rule
  form.name = rule.name
  form.enabled = rule.enabled
  form.priority = rule.priority
  form.match_mode = rule.match_mode
  form.platforms = [...rule.platforms]
  form.passthrough_code = rule.passthrough_code
  form.response_code = rule.response_code
  form.passthrough_body = rule.passthrough_body
  form.custom_message = rule.custom_message
  form.skip_monitoring = rule.skip_monitoring
  form.description = rule.description
  errorCodesInput.value = rule.error_codes.join(', ')
  keywordsInput.value = rule.keywords.join('\n')
  showEditModal.value = true
}

const handleDelete = (rule: ErrorPassthroughRule) => {
  deletingRule.value = rule
  showDeleteDialog.value = true
}

const parseErrorCodes = (): number[] => {
  if (!errorCodesInput.value.trim()) return []
  return errorCodesInput.value
    .split(/[,\s]+/)
    .map(s => parseInt(s.trim(), 10))
    .filter(n => !isNaN(n) && n > 0)
}

const parseKeywords = (): string[] => {
  if (!keywordsInput.value.trim()) return []
  return keywordsInput.value
    .split('\n')
    .map(s => s.trim())
    .filter(s => s.length > 0)
}

const handleSubmit = async () => {
  if (!form.name.trim()) {
    appStore.showError(t('admin.errorPassthrough.nameRequired'))
    return
  }

  const errorCodes = parseErrorCodes()
  const keywords = parseKeywords()

  if (errorCodes.length === 0 && keywords.length === 0) {
    appStore.showError(t('admin.errorPassthrough.conditionsRequired'))
    return
  }

  submitting.value = true
  try {
    const data = {
      name: form.name.trim(),
      enabled: form.enabled,
      priority: form.priority,
      error_codes: errorCodes,
      keywords: keywords,
      match_mode: form.match_mode,
      platforms: form.platforms,
      passthrough_code: form.passthrough_code,
      response_code: form.passthrough_code ? null : form.response_code,
      passthrough_body: form.passthrough_body,
      custom_message: form.passthrough_body ? null : form.custom_message,
      skip_monitoring: form.skip_monitoring,
      description: form.description?.trim() || null
    }

    if (showEditModal.value && editingRule.value) {
      await adminAPI.errorPassthrough.update(editingRule.value.id, data)
      appStore.showSuccess(t('admin.errorPassthrough.ruleUpdated'))
    } else {
      await adminAPI.errorPassthrough.create(data)
      appStore.showSuccess(t('admin.errorPassthrough.ruleCreated'))
    }

    closeFormModal()
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToSave'))
    console.error('Error saving rule:', error)
  } finally {
    submitting.value = false
  }
}

const toggleEnabled = async (rule: ErrorPassthroughRule) => {
  try {
    await adminAPI.errorPassthrough.toggleEnabled(rule.id, !rule.enabled)
    rule.enabled = !rule.enabled
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToToggle'))
    console.error('Error toggling rule:', error)
  }
}

const confirmDelete = async () => {
  if (!deletingRule.value) return

  try {
    await adminAPI.errorPassthrough.delete(deletingRule.value.id)
    appStore.showSuccess(t('admin.errorPassthrough.ruleDeleted'))
    showDeleteDialog.value = false
    deletingRule.value = null
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToDelete'))
    console.error('Error deleting rule:', error)
  }
}
</script>

<style scoped>
.error-passthrough-rules-modal__subtitle,
.error-passthrough-rules-modal__rule-description,
.error-passthrough-rules-modal__count,
.error-passthrough-rules-modal__behavior-text,
.error-passthrough-rules-modal__option-description {
  color: var(--theme-page-muted);
}

.error-passthrough-rules-modal__content {
  display: flex;
  flex-direction: column;
  gap: var(--theme-table-layout-gap);
}

.error-passthrough-rules-modal__status-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) 0;
  color: var(--theme-page-muted);
}

.error-passthrough-rules-modal__status-icon,
.error-passthrough-rules-modal__empty-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 78%, transparent);
}

.error-passthrough-rules-modal__empty-state {
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) var(--theme-table-mobile-card-padding);
  text-align: center;
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 80%, transparent);
  border-radius: calc(var(--theme-button-radius) + 4px);
  background: color-mix(in srgb, var(--theme-surface-soft) 76%, var(--theme-surface));
}

.error-passthrough-rules-modal__empty-icon-wrap {
  margin: 0 auto calc(var(--theme-table-mobile-card-padding) * 0.75);
  display: flex;
  height: var(--theme-empty-icon-size);
  width: var(--theme-empty-icon-size);
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, var(--theme-surface));
}

.error-passthrough-rules-modal__empty-title,
.error-passthrough-rules-modal__rule-name,
.error-passthrough-rules-modal__form-section-title,
.error-passthrough-rules-modal__option-label,
.error-passthrough-rules-modal__checkbox-label {
  color: var(--theme-page-text);
}

.error-passthrough-rules-modal__table-shell {
  max-height: var(--theme-proxy-quality-table-max-height);
  overflow: auto;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.error-passthrough-rules-modal__table {
  border-collapse: separate;
  border-spacing: 0;
}

.error-passthrough-rules-modal__table-head {
  background: var(--theme-table-head-bg);
}

.error-passthrough-rules-modal__table-header {
  padding: calc(var(--theme-button-padding-y) * 0.8) calc(var(--theme-button-padding-x) * 0.6);
  text-align: left;
  font-size: var(--theme-table-head-font-size);
  font-weight: 500;
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
  color: var(--theme-table-head-text);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.error-passthrough-rules-modal__table-body {
  background: var(--theme-surface);
}

.error-passthrough-rules-modal__table-cell {
  padding: calc(var(--theme-button-padding-y) * 0.8) calc(var(--theme-button-padding-x) * 0.6);
}

.error-passthrough-rules-modal__table-cell--nowrap {
  white-space: nowrap;
}

.error-passthrough-rules-modal__table-row td {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 70%, transparent);
}

.error-passthrough-rules-modal__table-body tr:first-child td {
  border-top: none;
}

.error-passthrough-rules-modal__table-row:hover {
  background: var(--theme-table-row-hover);
}

.error-passthrough-rules-modal__behavior-icon--inactive {
  color: color-mix(in srgb, var(--theme-page-muted) 74%, transparent);
}

.error-passthrough-rules-modal__behavior-icon--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.error-passthrough-rules-modal__behavior-icon--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.error-passthrough-rules-modal__toggle-track {
  position: relative;
  display: inline-flex;
  height: 1rem;
  width: 1.75rem;
  flex-shrink: 0;
  cursor: pointer;
  border-radius: 999px;
  border: 2px solid transparent;
  transition: background-color 0.2s ease;
  outline: none;
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.error-passthrough-rules-modal__toggle-track--enabled {
  background: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-accent-strong));
}

.error-passthrough-rules-modal__toggle-track--disabled {
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.error-passthrough-rules-modal__toggle-thumb {
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.error-passthrough-rules-modal__action-button {
  border-radius: calc(var(--theme-button-radius) - 4px);
  padding: 0.25rem;
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
  transition: color 0.2s ease, background-color 0.2s ease;
}

.error-passthrough-rules-modal__action-button--info:hover {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.error-passthrough-rules-modal__action-button--danger:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.error-passthrough-rules-modal__form-section {
  border-radius: calc(var(--theme-surface-radius) + 2px);
  padding: var(--theme-table-mobile-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 72%, var(--theme-surface));
}

.error-passthrough-rules-modal__option-row,
.error-passthrough-rules-modal__checkbox-row {
  color: var(--theme-page-text);
}

.error-passthrough-rules-modal__radio-input,
.error-passthrough-rules-modal__checkbox-input {
  accent-color: var(--theme-accent);
}

.error-passthrough-rules-modal__checkbox-input--warning {
  accent-color: rgb(var(--theme-warning-rgb));
}
</style>
