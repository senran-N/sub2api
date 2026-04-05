<template>
  <KeysFormDialog
    :show="showCreateModal || showEditModal"
    :title="showEditModal ? editTitle : createTitle"
    :is-edit-mode="showEditModal"
    :form-data="formData"
    :group-options="groupOptions"
    :status-options="statusOptions"
    :custom-key-error="customKeyError"
    :selected-key="selectedKey"
    :submitting="submitting"
    @close="$emit('close-modals')"
    @submit="$emit('submit')"
    @reset-quota="$emit('confirm-reset-quota')"
    @reset-rate-limit="$emit('confirm-reset-rate-limit')"
    @set-expiration-days="$emit('set-expiration-days', $event)"
  />

  <ConfirmDialog
    :show="showDeleteDialog"
    :title="deleteTitle"
    :message="deleteMessage"
    :confirm-text="deleteConfirmText"
    :cancel-text="cancelText"
    :danger="true"
    @confirm="$emit('delete')"
    @cancel="$emit('update:showDeleteDialog', false)"
  />

  <ConfirmDialog
    :show="showResetQuotaDialog"
    :title="resetQuotaTitle"
    :message="resetQuotaMessage"
    :confirm-text="resetText"
    :cancel-text="cancelText"
    :danger="true"
    @confirm="$emit('reset-quota')"
    @cancel="$emit('update:showResetQuotaDialog', false)"
  />

  <ConfirmDialog
    :show="showResetRateLimitDialog"
    :title="resetRateLimitTitle"
    :message="resetRateLimitMessage"
    :confirm-text="resetText"
    :cancel-text="cancelText"
    :danger="true"
    @confirm="$emit('reset-rate-limit')"
    @cancel="$emit('update:showResetRateLimitDialog', false)"
  />

  <UseKeyModal
    :show="showUseKeyModal"
    :api-key="selectedKey?.key || ''"
    :base-url="publicSettings?.api_base_url || ''"
    :platform="selectedKey?.group?.platform || null"
    :allow-messages-dispatch="selectedKey?.group?.allow_messages_dispatch || false"
    @close="$emit('close-use-key-modal')"
  />

  <KeysCcsClientSelectDialog
    :show="showCcsClientSelect"
    :title="ccsClientSelectTitle"
    :description="ccsClientSelectDescription"
    :claude-label="claudeLabel"
    :claude-description="claudeDescription"
    :gemini-label="geminiLabel"
    :gemini-description="geminiDescription"
    :cancel-label="cancelText"
    @close="$emit('close-ccs-client-select')"
    @select="$emit('select-ccs-client', $event)"
  />
</template>

<script setup lang="ts">
import type { ApiKey, PublicSettings } from '@/types'
import type { UserKeyFormData, UserKeyGroupOption } from './keysForm'
import type { CcsClientType } from './keysView'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import UseKeyModal from '@/components/keys/UseKeyModal.vue'
import KeysCcsClientSelectDialog from './KeysCcsClientSelectDialog.vue'
import KeysFormDialog from './KeysFormDialog.vue'

defineProps<{
  showCreateModal: boolean
  showEditModal: boolean
  showDeleteDialog: boolean
  showResetQuotaDialog: boolean
  showResetRateLimitDialog: boolean
  showUseKeyModal: boolean
  showCcsClientSelect: boolean
  createTitle: string
  editTitle: string
  deleteTitle: string
  deleteMessage: string
  deleteConfirmText: string
  resetQuotaTitle: string
  resetQuotaMessage: string
  resetRateLimitTitle: string
  resetRateLimitMessage: string
  resetText: string
  cancelText: string
  ccsClientSelectTitle: string
  ccsClientSelectDescription: string
  claudeLabel: string
  claudeDescription: string
  geminiLabel: string
  geminiDescription: string
  formData: UserKeyFormData
  groupOptions: UserKeyGroupOption[]
  statusOptions: Array<{ value: string; label: string }>
  customKeyError: string
  selectedKey: ApiKey | null
  submitting: boolean
  publicSettings: PublicSettings | null | undefined
}>()

defineEmits<{
  'close-modals': []
  submit: []
  'confirm-reset-quota': []
  'confirm-reset-rate-limit': []
  'reset-quota': []
  'reset-rate-limit': []
  'set-expiration-days': [days: number]
  delete: []
  'update:showDeleteDialog': [value: boolean]
  'update:showResetQuotaDialog': [value: boolean]
  'update:showResetRateLimitDialog': [value: boolean]
  'close-use-key-modal': []
  'close-ccs-client-select': []
  'select-ccs-client': [value: CcsClientType]
}>()
</script>
