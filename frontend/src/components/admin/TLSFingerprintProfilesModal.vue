<template>
  <BaseDialog
    :show="show"
    :title="t('admin.tlsFingerprintProfiles.title')"
    width="wide"
    @close="$emit('close')"
  >
    <div class="tls-fingerprint-profiles-modal__content">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <p class="tls-fingerprint-profiles-modal__subtitle text-sm">
          {{ t('admin.tlsFingerprintProfiles.description') }}
        </p>
        <button @click="showCreateModal = true" class="btn btn-primary btn-sm">
          <Icon name="plus" size="sm" class="mr-1" />
          {{ t('admin.tlsFingerprintProfiles.createProfile') }}
        </button>
      </div>

      <!-- Profiles Table -->
      <div v-if="loading" class="tls-fingerprint-profiles-modal__status-state">
        <Icon name="refresh" size="lg" class="tls-fingerprint-profiles-modal__status-icon animate-spin" />
      </div>

      <div v-else-if="profiles.length === 0" class="tls-fingerprint-profiles-modal__empty-state">
        <div class="tls-fingerprint-profiles-modal__empty-icon-wrap">
          <Icon name="shield" size="lg" class="tls-fingerprint-profiles-modal__empty-icon" />
        </div>
        <h4 class="tls-fingerprint-profiles-modal__text-strong tls-fingerprint-profiles-modal__empty-title">
          {{ t('admin.tlsFingerprintProfiles.noProfiles') }}
        </h4>
        <p class="tls-fingerprint-profiles-modal__subtitle text-sm">
          {{ t('admin.tlsFingerprintProfiles.createFirstProfile') }}
        </p>
      </div>

      <div v-else class="tls-fingerprint-profiles-modal__table-shell">
        <table class="min-w-full">
          <thead class="tls-fingerprint-profiles-modal__table-head sticky top-0">
            <tr>
              <th class="tls-fingerprint-profiles-modal__table-header">
                {{ t('admin.tlsFingerprintProfiles.columns.name') }}
              </th>
              <th class="tls-fingerprint-profiles-modal__table-header">
                {{ t('admin.tlsFingerprintProfiles.columns.description') }}
              </th>
              <th class="tls-fingerprint-profiles-modal__table-header">
                {{ t('admin.tlsFingerprintProfiles.columns.grease') }}
              </th>
              <th class="tls-fingerprint-profiles-modal__table-header">
                {{ t('admin.tlsFingerprintProfiles.columns.alpn') }}
              </th>
              <th class="tls-fingerprint-profiles-modal__table-header">
                {{ t('admin.tlsFingerprintProfiles.columns.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="tls-fingerprint-profiles-modal__table-body">
            <tr v-for="profile in profiles" :key="profile.id" class="tls-fingerprint-profiles-modal__table-row">
              <td class="tls-fingerprint-profiles-modal__table-cell">
                <div class="tls-fingerprint-profiles-modal__text-strong text-sm font-medium">{{ profile.name }}</div>
              </td>
              <td class="tls-fingerprint-profiles-modal__table-cell">
                <div v-if="profile.description" class="tls-fingerprint-profiles-modal__subtitle max-w-xs truncate text-sm">
                  {{ profile.description }}
                </div>
                <div v-else class="tls-fingerprint-profiles-modal__text-soft text-xs">—</div>
              </td>
              <td class="tls-fingerprint-profiles-modal__table-cell">
                <Icon
                  :name="profile.enable_grease ? 'check' : 'lock'"
                  size="sm"
                  :class="profile.enable_grease ? 'tls-fingerprint-profiles-modal__icon--enabled' : 'tls-fingerprint-profiles-modal__text-soft'"
                />
              </td>
              <td class="tls-fingerprint-profiles-modal__table-cell">
                <div v-if="profile.alpn_protocols?.length" class="flex flex-wrap gap-1">
                  <span
                    v-for="proto in profile.alpn_protocols.slice(0, 3)"
                    :key="proto"
                    class="theme-chip theme-chip--compact theme-chip--accent text-xs"
                  >
                    {{ proto }}
                  </span>
                  <span v-if="profile.alpn_protocols.length > 3" class="tls-fingerprint-profiles-modal__subtitle text-xs">
                    +{{ profile.alpn_protocols.length - 3 }}
                  </span>
                </div>
                <div v-else class="tls-fingerprint-profiles-modal__text-soft text-xs">—</div>
              </td>
              <td class="tls-fingerprint-profiles-modal__table-cell">
                <div class="flex items-center gap-1">
                  <button
                    @click="handleEdit(profile)"
                    :class="getActionButtonClasses('info')"
                    :title="t('common.edit')"
                  >
                    <Icon name="edit" size="sm" />
                  </button>
                  <button
                    @click="handleDelete(profile)"
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
      :title="showEditModal ? t('admin.tlsFingerprintProfiles.editProfile') : t('admin.tlsFingerprintProfiles.createProfile')"
      width="wide"
      :z-index="60"
      @close="closeFormModal"
    >
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <!-- Paste YAML -->
        <div>
          <label class="input-label">{{ t('admin.tlsFingerprintProfiles.form.pasteYaml') }}</label>
          <textarea
            v-model="yamlInput"
            rows="4"
            class="input font-mono text-xs"
            :placeholder="t('admin.tlsFingerprintProfiles.form.pasteYamlPlaceholder')"
            @paste="handleYamlPaste"
          />
          <div class="mt-1 flex items-center gap-2">
            <button type="button" @click="parseYamlInput" class="btn btn-secondary btn-sm">
              {{ t('admin.tlsFingerprintProfiles.form.parseYaml') }}
            </button>
            <p class="tls-fingerprint-profiles-modal__subtitle text-xs">
              {{ t('admin.tlsFingerprintProfiles.form.pasteYamlHint') }}
              <a href="https://tls.sub2api.org" target="_blank" rel="noopener noreferrer" class="tls-fingerprint-profiles-modal__link underline">{{ t('admin.tlsFingerprintProfiles.form.openCollector') }}</a>
            </p>
          </div>
        </div>

        <hr class="tls-fingerprint-profiles-modal__divider" />

        <!-- Basic Info -->
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
          <div>
            <label class="input-label">{{ t('admin.tlsFingerprintProfiles.form.name') }}</label>
            <input
              v-model="form.name"
              type="text"
              required
              class="input"
              :placeholder="t('admin.tlsFingerprintProfiles.form.namePlaceholder')"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.tlsFingerprintProfiles.form.description') }}</label>
            <input
              v-model="form.description"
              type="text"
              class="input"
              :placeholder="t('admin.tlsFingerprintProfiles.form.descriptionPlaceholder')"
            />
          </div>
        </div>

        <!-- GREASE Toggle -->
        <div class="flex items-center gap-3">
          <button
            type="button"
            @click="form.enable_grease = !form.enable_grease"
            :class="getGreaseToggleTrackClasses(form.enable_grease)"
          >
            <span
              :class="[
                'tls-fingerprint-profiles-modal__toggle-thumb pointer-events-none inline-block h-4 w-4 transform rounded-full ring-0 transition duration-200 ease-in-out',
                form.enable_grease ? 'translate-x-4' : 'translate-x-0'
              ]"
            />
          </button>
          <div>
            <span class="tls-fingerprint-profiles-modal__text-body text-sm font-medium">
              {{ t('admin.tlsFingerprintProfiles.form.enableGrease') }}
            </span>
            <p class="tls-fingerprint-profiles-modal__subtitle text-xs">
              {{ t('admin.tlsFingerprintProfiles.form.enableGreaseHint') }}
            </p>
          </div>
        </div>

        <!-- TLS Array Fields - 2 column grid -->
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
          <div>
            <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.cipherSuites') }}</label>
            <textarea
              v-model="fieldInputs.cipher_suites"
              rows="2"
              class="input font-mono text-xs"
              :placeholder="'0x1301, 0x1302, 0xc02c'"
            />
            <p class="input-hint text-xs">{{ t('admin.tlsFingerprintProfiles.form.cipherSuitesHint') }}</p>
          </div>

          <div>
            <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.curves') }}</label>
            <textarea
              v-model="fieldInputs.curves"
              rows="2"
              class="input font-mono text-xs"
              :placeholder="'29, 23, 24'"
            />
            <p class="input-hint text-xs">{{ t('admin.tlsFingerprintProfiles.form.curvesHint') }}</p>
          </div>

          <div>
            <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.signatureAlgorithms') }}</label>
            <textarea
              v-model="fieldInputs.signature_algorithms"
              rows="2"
              class="input font-mono text-xs"
              :placeholder="'0x0403, 0x0804, 0x0401'"
            />
          </div>

          <div>
            <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.supportedVersions') }}</label>
            <textarea
              v-model="fieldInputs.supported_versions"
              rows="2"
              class="input font-mono text-xs"
              :placeholder="'0x0304, 0x0303'"
            />
          </div>

          <div>
            <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.keyShareGroups') }}</label>
            <textarea
              v-model="fieldInputs.key_share_groups"
              rows="2"
              class="input font-mono text-xs"
              :placeholder="'29, 23'"
            />
          </div>

          <div>
            <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.extensions') }}</label>
            <textarea
              v-model="fieldInputs.extensions"
              rows="2"
              class="input font-mono text-xs"
              :placeholder="'0x0000, 0x0005, 0x000a'"
            />
          </div>

          <div>
            <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.pointFormats') }}</label>
            <textarea
              v-model="fieldInputs.point_formats"
              rows="2"
              class="input font-mono text-xs"
              :placeholder="'0'"
            />
          </div>

          <div>
            <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.pskModes') }}</label>
            <textarea
              v-model="fieldInputs.psk_modes"
              rows="2"
              class="input font-mono text-xs"
              :placeholder="'1'"
            />
          </div>
        </div>

        <!-- ALPN Protocols - full width -->
        <div>
          <label class="input-label text-xs">{{ t('admin.tlsFingerprintProfiles.form.alpnProtocols') }}</label>
          <textarea
            v-model="fieldInputs.alpn_protocols"
            rows="2"
            class="input font-mono text-xs"
            :placeholder="'h2, http/1.1'"
          />
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
      :title="t('admin.tlsFingerprintProfiles.deleteProfile')"
      :message="t('admin.tlsFingerprintProfiles.deleteConfirmMessage', { name: deletingProfile?.name })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { TLSFingerprintProfile } from '@/api/admin/tlsFingerprintProfile'
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

const profiles = ref<TLSFingerprintProfile[]>([])
const loading = ref(false)
const submitting = ref(false)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const editingProfile = ref<TLSFingerprintProfile | null>(null)
const deletingProfile = ref<TLSFingerprintProfile | null>(null)
const yamlInput = ref('')

// Raw string inputs for array fields
const fieldInputs = reactive({
  cipher_suites: '',
  curves: '',
  point_formats: '',
  signature_algorithms: '',
  alpn_protocols: '',
  supported_versions: '',
  key_share_groups: '',
  psk_modes: '',
  extensions: ''
})

const form = reactive({
  name: '',
  description: null as string | null,
  enable_grease: false
})

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getActionButtonClasses = (tone: 'info' | 'danger') => {
  return joinClassNames(
    'tls-fingerprint-profiles-modal__action-button',
    tone === 'info'
      ? 'tls-fingerprint-profiles-modal__action-button--info'
      : 'tls-fingerprint-profiles-modal__action-button--danger'
  )
}

const getGreaseToggleTrackClasses = (enabled: boolean) => {
  return joinClassNames(
    'tls-fingerprint-profiles-modal__toggle-track',
    enabled
      ? 'tls-fingerprint-profiles-modal__toggle-track--enabled'
      : 'tls-fingerprint-profiles-modal__toggle-track--disabled'
  )
}

// Load profiles when dialog opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    loadProfiles()
  }
})

const loadProfiles = async () => {
  loading.value = true
  try {
    profiles.value = await adminAPI.tlsFingerprintProfiles.list()
  } catch (error) {
    appStore.showError(t('admin.tlsFingerprintProfiles.loadFailed'))
    console.error('Error loading TLS fingerprint profiles:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.description = null
  form.enable_grease = false
  fieldInputs.cipher_suites = ''
  fieldInputs.curves = ''
  fieldInputs.point_formats = ''
  fieldInputs.signature_algorithms = ''
  fieldInputs.alpn_protocols = ''
  fieldInputs.supported_versions = ''
  fieldInputs.key_share_groups = ''
  fieldInputs.psk_modes = ''
  fieldInputs.extensions = ''
  yamlInput.value = ''
}

/**
 * Parse YAML output from tls-fingerprint-web and fill form fields.
 * Expected format:
 *   # comment lines
 *   profile_key:
 *     name: "Profile Name"
 *     enable_grease: false
 *     cipher_suites: [4866, 4867, ...]
 *     alpn_protocols: ["h2", "http/1.1"]
 *     ...
 */
const parseYamlInput = () => {
  const text = yamlInput.value.trim()
  if (!text) return

  // Simple YAML parser for flat key-value structure
  // Extracts "key: value" lines, handling arrays like [1, 2, 3] and ["h2", "http/1.1"]
  const lines = text.split('\n')

  let foundName = false

  for (const line of lines) {
    const trimmed = line.trim()
    // Skip comments and empty lines
    if (!trimmed || trimmed.startsWith('#')) continue

    // Match "key: value" pattern (must have at least 2 leading spaces to be a property)
    const match = trimmed.match(/^(\w+):\s*(.+)$/)
    if (!match) continue

    const [, key, rawValue] = match
    const value = rawValue.trim()

    switch (key) {
      case 'name': {
        // Remove surrounding quotes
        const unquoted = value.replace(/^["']|["']$/g, '')
        if (unquoted) {
          form.name = unquoted
          foundName = true
        }
        break
      }
      case 'enable_grease':
        form.enable_grease = value === 'true'
        break
      case 'cipher_suites':
      case 'curves':
      case 'point_formats':
      case 'signature_algorithms':
      case 'supported_versions':
      case 'key_share_groups':
      case 'psk_modes':
      case 'extensions': {
        // Parse YAML array: [1, 2, 3] — values are decimal integers from tls-fingerprint-web
        const arrMatch = value.match(/^\[(.*)?\]$/)
        if (arrMatch) {
          const inner = arrMatch[1] || ''
          fieldInputs[key as keyof typeof fieldInputs] = inner
            .split(',')
            .map(s => s.trim())
            .filter(s => s.length > 0)
            .join(', ')
        }
        break
      }
      case 'alpn_protocols': {
        // Parse string array: ["h2", "http/1.1"]
        const arrMatch = value.match(/^\[(.*)?\]$/)
        if (arrMatch) {
          const inner = arrMatch[1] || ''
          fieldInputs.alpn_protocols = inner
            .split(',')
            .map(s => s.trim().replace(/^["']|["']$/g, ''))
            .filter(s => s.length > 0)
            .join(', ')
        }
        break
      }
    }
  }

  if (foundName) {
    appStore.showSuccess(t('admin.tlsFingerprintProfiles.form.yamlParsed'))
  } else {
    appStore.showError(t('admin.tlsFingerprintProfiles.form.yamlParseFailed'))
  }
}

// Auto-parse on paste event
const handleYamlPaste = () => {
  // Use nextTick to ensure v-model has updated
  setTimeout(() => parseYamlInput(), 50)
}

const closeFormModal = () => {
  showCreateModal.value = false
  showEditModal.value = false
  editingProfile.value = null
  resetForm()
}

// Parse a comma-separated string of numbers supporting both hex (0x...) and decimal
const parseNumericArray = (input: string): number[] => {
  if (!input.trim()) return []
  return input
    .split(',')
    .map(s => s.trim())
    .filter(s => s.length > 0)
    .map(s => s.startsWith('0x') || s.startsWith('0X') ? parseInt(s, 16) : parseInt(s, 10))
    .filter(n => !isNaN(n))
}

// Parse a comma-separated string of string values
const parseStringArray = (input: string): string[] => {
  if (!input.trim()) return []
  return input
    .split(',')
    .map(s => s.trim())
    .filter(s => s.length > 0)
}

// Format a number as hex with 0x prefix and 4-digit padding
const formatHex = (n: number): string => '0x' + n.toString(16).padStart(4, '0')

// Format numeric arrays for display in textarea (null-safe)
const formatNumericArray = (arr: number[] | null | undefined): string => (arr ?? []).map(formatHex).join(', ')

// For point_formats and psk_modes (uint8), show as plain numbers (null-safe)
const formatPlainNumericArray = (arr: number[] | null | undefined): string => (arr ?? []).join(', ')

const handleEdit = (profile: TLSFingerprintProfile) => {
  editingProfile.value = profile
  form.name = profile.name
  form.description = profile.description
  form.enable_grease = profile.enable_grease
  fieldInputs.cipher_suites = formatNumericArray(profile.cipher_suites)
  fieldInputs.curves = formatPlainNumericArray(profile.curves)
  fieldInputs.point_formats = formatPlainNumericArray(profile.point_formats)
  fieldInputs.signature_algorithms = formatNumericArray(profile.signature_algorithms)
  fieldInputs.alpn_protocols = (profile.alpn_protocols ?? []).join(', ')
  fieldInputs.supported_versions = formatNumericArray(profile.supported_versions)
  fieldInputs.key_share_groups = formatPlainNumericArray(profile.key_share_groups)
  fieldInputs.psk_modes = formatPlainNumericArray(profile.psk_modes)
  fieldInputs.extensions = formatNumericArray(profile.extensions)
  showEditModal.value = true
}

const handleDelete = (profile: TLSFingerprintProfile) => {
  deletingProfile.value = profile
  showDeleteDialog.value = true
}

const handleSubmit = async () => {
  if (!form.name.trim()) {
    appStore.showError(t('admin.tlsFingerprintProfiles.form.name') + ' ' + t('common.required'))
    return
  }

  submitting.value = true
  try {
    const data = {
      name: form.name.trim(),
      description: form.description?.trim() || null,
      enable_grease: form.enable_grease,
      cipher_suites: parseNumericArray(fieldInputs.cipher_suites),
      curves: parseNumericArray(fieldInputs.curves),
      point_formats: parseNumericArray(fieldInputs.point_formats),
      signature_algorithms: parseNumericArray(fieldInputs.signature_algorithms),
      alpn_protocols: parseStringArray(fieldInputs.alpn_protocols),
      supported_versions: parseNumericArray(fieldInputs.supported_versions),
      key_share_groups: parseNumericArray(fieldInputs.key_share_groups),
      psk_modes: parseNumericArray(fieldInputs.psk_modes),
      extensions: parseNumericArray(fieldInputs.extensions)
    }

    if (showEditModal.value && editingProfile.value) {
      await adminAPI.tlsFingerprintProfiles.update(editingProfile.value.id, data)
      appStore.showSuccess(t('admin.tlsFingerprintProfiles.updateSuccess'))
    } else {
      await adminAPI.tlsFingerprintProfiles.create(data)
      appStore.showSuccess(t('admin.tlsFingerprintProfiles.createSuccess'))
    }

    closeFormModal()
    loadProfiles()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.tlsFingerprintProfiles.saveFailed'))
    console.error('Error saving TLS fingerprint profile:', error)
  } finally {
    submitting.value = false
  }
}

const confirmDelete = async () => {
  if (!deletingProfile.value) return

  try {
    await adminAPI.tlsFingerprintProfiles.delete(deletingProfile.value.id)
    appStore.showSuccess(t('admin.tlsFingerprintProfiles.deleteSuccess'))
    showDeleteDialog.value = false
    deletingProfile.value = null
    loadProfiles()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.tlsFingerprintProfiles.deleteFailed'))
    console.error('Error deleting TLS fingerprint profile:', error)
  }
}
</script>

<style scoped>
.tls-fingerprint-profiles-modal__subtitle,
.tls-fingerprint-profiles-modal__status-state {
  color: var(--theme-page-muted);
}

.tls-fingerprint-profiles-modal__content {
  display: flex;
  flex-direction: column;
  gap: var(--theme-table-layout-gap);
}

.tls-fingerprint-profiles-modal__status-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) 0;
}

.tls-fingerprint-profiles-modal__text-strong,
.tls-fingerprint-profiles-modal__text-body {
  color: var(--theme-page-text);
}

.tls-fingerprint-profiles-modal__text-soft,
.tls-fingerprint-profiles-modal__status-icon,
.tls-fingerprint-profiles-modal__empty-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.tls-fingerprint-profiles-modal__icon--enabled {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.tls-fingerprint-profiles-modal__empty-state {
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) var(--theme-table-mobile-card-padding);
  text-align: center;
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 78%, transparent);
  border-radius: calc(var(--theme-button-radius) + 4px);
  background: color-mix(in srgb, var(--theme-surface-soft) 76%, var(--theme-surface));
}

.tls-fingerprint-profiles-modal__empty-icon-wrap {
  margin: 0 auto calc(var(--theme-table-mobile-card-padding) * 0.75);
  display: flex;
  height: var(--theme-empty-icon-size);
  width: var(--theme-empty-icon-size);
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, var(--theme-surface));
}

.tls-fingerprint-profiles-modal__table-shell {
  max-height: var(--theme-proxy-quality-table-max-height);
  overflow: auto;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.tls-fingerprint-profiles-modal__table-head {
  background: var(--theme-table-head-bg);
}

.tls-fingerprint-profiles-modal__table-header {
  padding: calc(var(--theme-button-padding-y) * 0.8) calc(var(--theme-button-padding-x) * 0.6);
  text-align: left;
  font-size: var(--theme-table-head-font-size);
  font-weight: 500;
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
  color: var(--theme-table-head-text);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.tls-fingerprint-profiles-modal__table-cell {
  padding: calc(var(--theme-button-padding-y) * 0.8) calc(var(--theme-button-padding-x) * 0.6);
}

.tls-fingerprint-profiles-modal__table-body tr + tr td {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.tls-fingerprint-profiles-modal__table-row:hover {
  background: var(--theme-table-row-hover);
}

.tls-fingerprint-profiles-modal__action-button {
  padding: 0.25rem;
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
  border-radius: calc(var(--theme-button-radius) - 4px);
  transition: color 0.2s ease, background-color 0.2s ease;
}

.tls-fingerprint-profiles-modal__action-button--info:hover {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.tls-fingerprint-profiles-modal__action-button--danger:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.tls-fingerprint-profiles-modal__link {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.tls-fingerprint-profiles-modal__link:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%);
}

.tls-fingerprint-profiles-modal__divider {
  border-color: color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.tls-fingerprint-profiles-modal__toggle-track {
  position: relative;
  display: inline-flex;
  height: 1.25rem;
  width: 2.25rem;
  flex-shrink: 0;
  cursor: pointer;
  border-radius: 999px;
  border: 2px solid transparent;
  transition: background-color 0.2s ease;
  outline: none;
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.tls-fingerprint-profiles-modal__toggle-track--enabled {
  background: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-accent-strong));
}

.tls-fingerprint-profiles-modal__toggle-track--disabled {
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.tls-fingerprint-profiles-modal__toggle-thumb {
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}
</style>
