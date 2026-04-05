<template>
  <BaseDialog :show="show" :title="title" width="normal" @close="emit('close')">
    <form id="key-form" class="space-y-5" @submit.prevent="emit('submit')">
      <div>
        <label class="input-label">{{ t('keys.nameLabel') }}</label>
        <input
          v-model="formData.name"
          type="text"
          required
          class="input"
          :placeholder="t('keys.namePlaceholder')"
          data-tour="key-form-name"
        />
      </div>

      <div>
        <label class="input-label">{{ t('keys.groupLabel') }}</label>
        <Select
          v-model="formData.group_id"
          :options="groupOptions"
          :placeholder="t('keys.selectGroup')"
          :searchable="true"
          :search-placeholder="t('keys.searchGroup')"
          data-tour="key-form-group"
        >
          <template #selected="{ option }">
            <GroupBadge
              v-if="option"
              :name="(option as unknown as UserKeyGroupOption).label"
              :platform="(option as unknown as UserKeyGroupOption).platform"
              :subscription-type="(option as unknown as UserKeyGroupOption).subscriptionType"
              :rate-multiplier="(option as unknown as UserKeyGroupOption).rate"
              :user-rate-multiplier="(option as unknown as UserKeyGroupOption).userRate"
            />
            <span v-else class="text-gray-400">{{ t('keys.selectGroup') }}</span>
          </template>
          <template #option="{ option, selected }">
            <GroupOptionItem
              :name="(option as unknown as UserKeyGroupOption).label"
              :platform="(option as unknown as UserKeyGroupOption).platform"
              :subscription-type="(option as unknown as UserKeyGroupOption).subscriptionType"
              :rate-multiplier="(option as unknown as UserKeyGroupOption).rate"
              :user-rate-multiplier="(option as unknown as UserKeyGroupOption).userRate"
              :description="(option as unknown as UserKeyGroupOption).description"
              :selected="selected"
            />
          </template>
        </Select>
      </div>

      <div v-if="!isEditMode" class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="input-label mb-0">{{ t('keys.customKeyLabel') }}</label>
          <button
            type="button"
            :class="toggleClass(formData.use_custom_key)"
            @click="formData.use_custom_key = !formData.use_custom_key"
          >
            <span :class="toggleThumbClass(formData.use_custom_key)" />
          </button>
        </div>
        <div v-if="formData.use_custom_key">
          <input
            v-model="formData.custom_key"
            type="text"
            class="input font-mono"
            :placeholder="t('keys.customKeyPlaceholder')"
            :class="{ 'border-red-500 dark:border-red-500': customKeyError }"
          />
          <p v-if="customKeyError" class="mt-1 text-sm text-red-500">{{ customKeyError }}</p>
          <p v-else class="input-hint">{{ t('keys.customKeyHint') }}</p>
        </div>
      </div>

      <div v-if="isEditMode">
        <label class="input-label">{{ t('keys.statusLabel') }}</label>
        <Select
          v-model="formData.status"
          :options="statusOptions"
          :placeholder="t('keys.selectStatus')"
        />
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="input-label mb-0">{{ t('keys.ipRestriction') }}</label>
          <button
            type="button"
            :class="toggleClass(formData.enable_ip_restriction)"
            @click="formData.enable_ip_restriction = !formData.enable_ip_restriction"
          >
            <span :class="toggleThumbClass(formData.enable_ip_restriction)" />
          </button>
        </div>

        <div v-if="formData.enable_ip_restriction" class="space-y-4 pt-2">
          <div>
            <label class="input-label">{{ t('keys.ipWhitelist') }}</label>
            <textarea
              v-model="formData.ip_whitelist"
              rows="3"
              class="input font-mono text-sm"
              :placeholder="t('keys.ipWhitelistPlaceholder')"
            />
            <p class="input-hint">{{ t('keys.ipWhitelistHint') }}</p>
          </div>

          <div>
            <label class="input-label">{{ t('keys.ipBlacklist') }}</label>
            <textarea
              v-model="formData.ip_blacklist"
              rows="3"
              class="input font-mono text-sm"
              :placeholder="t('keys.ipBlacklistPlaceholder')"
            />
            <p class="input-hint">{{ t('keys.ipBlacklistHint') }}</p>
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <label class="input-label">{{ t('keys.quotaLimit') }}</label>
        <div class="space-y-4">
          <div>
            <div class="relative">
              <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">$</span>
              <input
                v-model.number="formData.quota"
                type="number"
                step="0.01"
                min="0"
                class="input pl-7"
                :placeholder="t('keys.quotaAmountPlaceholder')"
              />
            </div>
            <p class="input-hint">{{ t('keys.quotaAmountHint') }}</p>
          </div>

          <div v-if="isEditMode && selectedKey && selectedKey.quota > 0">
            <label class="input-label">{{ t('keys.quotaUsed') }}</label>
            <div class="flex items-center gap-2">
              <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700">
                <span class="font-medium text-gray-900 dark:text-white">
                  ${{ selectedKey.quota_used?.toFixed(4) || '0.0000' }}
                </span>
                <span class="mx-2 text-gray-400">/</span>
                <span class="text-gray-500 dark:text-gray-400">
                  ${{ selectedKey.quota?.toFixed(2) || '0.00' }}
                </span>
              </div>
              <button
                type="button"
                class="btn btn-secondary text-sm"
                :title="t('keys.resetQuotaUsed')"
                @click="emit('reset-quota')"
              >
                {{ t('keys.reset') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="input-label mb-0">{{ t('keys.rateLimitSection') }}</label>
          <button
            type="button"
            :class="toggleClass(formData.enable_rate_limit)"
            @click="formData.enable_rate_limit = !formData.enable_rate_limit"
          >
            <span :class="toggleThumbClass(formData.enable_rate_limit)" />
          </button>
        </div>

        <div v-if="formData.enable_rate_limit" class="space-y-4 pt-2">
          <p class="input-hint -mt-2">{{ t('keys.rateLimitHint') }}</p>

          <div v-for="window in rateLimitWindows" :key="window.key">
            <label class="input-label">{{ window.label }}</label>
            <div class="relative">
              <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">$</span>
              <input
                v-model.number="formData[window.modelKey]"
                type="number"
                step="0.01"
                min="0"
                class="input pl-7"
                placeholder="0"
              />
            </div>

            <div v-if="isEditMode && selectedKey && selectedKey[window.limitKey] > 0" class="mt-2">
              <div class="flex items-center gap-2">
                <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 text-sm dark:bg-dark-700">
                  <span :class="['font-medium', getUsageTone(selectedKey[window.usageKey], selectedKey[window.limitKey])]">
                    ${{ selectedKey[window.usageKey]?.toFixed(4) || '0.0000' }}
                  </span>
                  <span class="mx-2 text-gray-400">/</span>
                  <span class="text-gray-500 dark:text-gray-400">
                    ${{ selectedKey[window.limitKey]?.toFixed(2) || '0.00' }}
                  </span>
                </div>
              </div>
              <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  :class="['h-full rounded-full transition-all', getUsageBarTone(selectedKey[window.usageKey], selectedKey[window.limitKey])]"
                  :style="{ width: getUsageWidth(selectedKey[window.usageKey], selectedKey[window.limitKey]) }"
                />
              </div>
            </div>
          </div>

          <div
            v-if="
              isEditMode &&
              selectedKey &&
              (selectedKey.rate_limit_5h > 0 ||
                selectedKey.rate_limit_1d > 0 ||
                selectedKey.rate_limit_7d > 0)
            "
          >
            <button type="button" class="btn btn-secondary text-sm" @click="emit('reset-rate-limit')">
              {{ t('keys.resetRateLimitUsage') }}
            </button>
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="input-label mb-0">{{ t('keys.expiration') }}</label>
          <button
            type="button"
            :class="toggleClass(formData.enable_expiration)"
            @click="formData.enable_expiration = !formData.enable_expiration"
          >
            <span :class="toggleThumbClass(formData.enable_expiration)" />
          </button>
        </div>

        <div v-if="formData.enable_expiration" class="space-y-4 pt-2">
          <div class="flex flex-wrap gap-2">
            <button
              v-for="days in ['7', '30', '90']"
              :key="days"
              type="button"
              :class="[
                'rounded-lg px-3 py-1.5 text-sm transition-colors',
                formData.expiration_preset === days
                  ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600'
              ]"
              @click="emit('set-expiration-days', Number(days))"
            >
              {{ isEditMode ? t('keys.extendDays', { days }) : t('keys.expiresInDays', { days }) }}
            </button>
            <button
              type="button"
              :class="[
                'rounded-lg px-3 py-1.5 text-sm transition-colors',
                formData.expiration_preset === 'custom'
                  ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600'
              ]"
              @click="formData.expiration_preset = 'custom'"
            >
              {{ t('keys.customDate') }}
            </button>
          </div>

          <div>
            <label class="input-label">{{ t('keys.expirationDate') }}</label>
            <input v-model="formData.expiration_date" type="datetime-local" class="input" />
            <p class="input-hint">{{ t('keys.expirationDateHint') }}</p>
          </div>

          <div v-if="isEditMode && selectedKey?.expires_at" class="text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('keys.currentExpiration') }}: </span>
            <span class="font-medium text-gray-900 dark:text-white">
              {{ formatDateTime(selectedKey.expires_at) }}
            </span>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          form="key-form"
          type="submit"
          :disabled="submitting"
          class="btn btn-primary"
          data-tour="key-form-submit"
        >
          <svg
            v-if="submitting"
            class="-ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{ submitting ? t('keys.saving') : isEditMode ? t('common.update') : t('common.create') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import type { ApiKey } from '@/types'
import { formatDateTime } from '@/utils/format'
import type { UserKeyFormData, UserKeyGroupOption } from './keysForm'

defineProps<{
  show: boolean
  title: string
  isEditMode: boolean
  formData: UserKeyFormData
  groupOptions: UserKeyGroupOption[]
  statusOptions: Array<{ value: string; label: string }>
  customKeyError: string
  selectedKey: ApiKey | null
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
  submit: []
  'reset-quota': []
  'reset-rate-limit': []
  'set-expiration-days': [days: number]
}>()

const { t } = useI18n()

const rateLimitWindows = [
  {
    key: '5h',
    label: t('keys.rateLimit5h'),
    modelKey: 'rate_limit_5h',
    usageKey: 'usage_5h',
    limitKey: 'rate_limit_5h'
  },
  {
    key: '1d',
    label: t('keys.rateLimit1d'),
    modelKey: 'rate_limit_1d',
    usageKey: 'usage_1d',
    limitKey: 'rate_limit_1d'
  },
  {
    key: '7d',
    label: t('keys.rateLimit7d'),
    modelKey: 'rate_limit_7d',
    usageKey: 'usage_7d',
    limitKey: 'rate_limit_7d'
  }
] as const

function toggleClass(enabled: boolean): string[] {
  return [
    'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
    enabled ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
  ]
}

function toggleThumbClass(enabled: boolean): string[] {
  return [
    'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
    enabled ? 'translate-x-4' : 'translate-x-0'
  ]
}

function getUsageTone(usage: number, limit: number): string {
  if (usage >= limit) {
    return 'text-red-500'
  }
  if (usage >= limit * 0.8) {
    return 'text-yellow-500'
  }

  return 'text-gray-900 dark:text-white'
}

function getUsageBarTone(usage: number, limit: number): string {
  if (usage >= limit) {
    return 'bg-red-500'
  }
  if (usage >= limit * 0.8) {
    return 'bg-yellow-500'
  }

  return 'bg-green-500'
}

function getUsageWidth(usage: number, limit: number): string {
  if (limit <= 0) {
    return '0%'
  }

  return `${Math.min((usage / limit) * 100, 100)}%`
}
</script>
