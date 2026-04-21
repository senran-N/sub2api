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
            <span v-else class="keys-form-dialog__placeholder">{{ t('keys.selectGroup') }}</span>
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
            :class="getCustomKeyInputClasses()"
            :placeholder="t('keys.customKeyPlaceholder')"
          />
          <p v-if="customKeyError" class="keys-form-dialog__error">{{ customKeyError }}</p>
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
              <span class="keys-form-dialog__prefix absolute left-3 top-1/2 -translate-y-1/2">$</span>
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
              <div class="keys-form-dialog__usage-box">
                <span class="keys-form-dialog__usage-current">
                  ${{ selectedKey.quota_used?.toFixed(4) || '0.0000' }}
                </span>
                <span class="keys-form-dialog__usage-separator">/</span>
                <span class="keys-form-dialog__usage-limit">
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
              <span class="keys-form-dialog__prefix absolute left-3 top-1/2 -translate-y-1/2">$</span>
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
                <div class="keys-form-dialog__usage-box keys-form-dialog__usage-box--compact">
                  <span :class="getUsageToneClasses(selectedKey[window.usageKey], selectedKey[window.limitKey])">
                    ${{ selectedKey[window.usageKey]?.toFixed(4) || '0.0000' }}
                  </span>
                  <span class="keys-form-dialog__usage-separator">/</span>
                  <span class="keys-form-dialog__usage-limit">
                    ${{ selectedKey[window.limitKey]?.toFixed(2) || '0.00' }}
                  </span>
                </div>
              </div>
              <div class="keys-form-dialog__progress-track">
                <div
                  :class="getUsageBarToneClasses(selectedKey[window.usageKey], selectedKey[window.limitKey])"
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
              :class="getExpirationPresetClasses(formData.expiration_preset === days)"
              @click="emit('set-expiration-days', Number(days))"
            >
              {{ isEditMode ? t('keys.extendDays', { days }) : t('keys.expiresInDays', { days }) }}
            </button>
            <button
              type="button"
              :class="getExpirationPresetClasses(formData.expiration_preset === 'custom')"
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
            <span class="keys-form-dialog__current-expiration-label">{{ t('keys.currentExpiration') }}: </span>
            <span class="keys-form-dialog__current-expiration-value">
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

const props = defineProps<{
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

function joinClassNames(...classNames: Array<string | false | null | undefined>) {
  return classNames.filter(Boolean).join(' ')
}

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

function toggleClass(enabled: boolean): string {
  return joinClassNames(
    'keys-form-dialog__toggle',
    enabled ? 'keys-form-dialog__toggle--enabled' : 'keys-form-dialog__toggle--disabled'
  )
}

function toggleThumbClass(enabled: boolean): string {
  return joinClassNames(
    'keys-form-dialog__toggle-thumb',
    enabled ? 'translate-x-4' : 'translate-x-0'
  )
}

function getCustomKeyInputClasses(): string {
  return joinClassNames('input font-mono', props.customKeyError ? 'input-error' : '')
}

function getUsageToneClasses(usage: number, limit: number): string {
  if (usage >= limit) {
    return 'keys-form-dialog__usage-current keys-form-dialog__usage-current--danger'
  }
  if (usage >= limit * 0.8) {
    return 'keys-form-dialog__usage-current keys-form-dialog__usage-current--warning'
  }

  return 'keys-form-dialog__usage-current'
}

function getUsageBarToneClasses(usage: number, limit: number): string {
  if (usage >= limit) {
    return 'keys-form-dialog__progress-bar keys-form-dialog__progress-bar--danger'
  }
  if (usage >= limit * 0.8) {
    return 'keys-form-dialog__progress-bar keys-form-dialog__progress-bar--warning'
  }

  return 'keys-form-dialog__progress-bar keys-form-dialog__progress-bar--success'
}

function getExpirationPresetClasses(isSelected: boolean): string {
  return joinClassNames(
    'keys-form-dialog__expiration-chip',
    isSelected
      ? 'keys-form-dialog__expiration-chip--active'
      : 'keys-form-dialog__expiration-chip--idle'
  )
}

function getUsageWidth(usage: number, limit: number): string {
  if (limit <= 0) {
    return '0%'
  }

  return `${Math.min((usage / limit) * 100, 100)}%`
}
</script>

<style scoped>
.keys-form-dialog__placeholder,
.keys-form-dialog__prefix,
.keys-form-dialog__usage-limit,
.keys-form-dialog__current-expiration-label {
  color: var(--theme-page-muted);
}

.keys-form-dialog__placeholder {
  font-size: 0.875rem;
}

.keys-form-dialog__error {
  color: rgb(var(--theme-danger-rgb));
  font-size: 0.875rem;
  margin-top: 0.25rem;
}

.keys-form-dialog__toggle {
  /* Track: 36x20, positions the absolute-positioned thumb inside.
     Avoid `border: 2px solid transparent` tricks — use explicit thumb
     offsets instead so the thumb stays 2px away from all edges and the
     translate math ([0 .. (36-2-16-2)] = 0..16px) matches `translate-x-4`. */
  position: relative;
  display: inline-block;
  height: 1.25rem;
  width: 2.25rem;
  flex-shrink: 0;
  cursor: pointer;
  border-radius: 9999px;
  transition: background-color 0.2s ease;
}

.keys-form-dialog__toggle--enabled {
  background: var(--theme-accent);
}

.keys-form-dialog__toggle--disabled {
  background: color-mix(in srgb, var(--theme-page-border) 76%, var(--theme-surface));
}

.keys-form-dialog__toggle-thumb {
  pointer-events: none;
  position: absolute;
  top: 0.125rem;
  left: 0.125rem;
  height: 1rem;
  width: 1rem;
  border-radius: 9999px;
  /* Use surface (white in light themes) to keep the thumb crisp against
     the orange accent — `surface-contrast-text` is cream on Claude light
     (#F0EDE6) and looked muddy on the orange pill. Small drop shadow for
     lift instead of the page-level `--theme-card-shadow`. */
  background: var(--theme-surface);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.18);
  transition: transform 0.2s ease;
}

.keys-form-dialog__usage-box {
  flex: 1 1 0%;
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
  padding: 0.5rem 0.75rem;
}

.keys-form-dialog__usage-box--compact {
  font-size: 0.875rem;
}

.keys-form-dialog__usage-current,
.keys-form-dialog__current-expiration-value {
  color: var(--theme-page-text);
  font-weight: 600;
}

.keys-form-dialog__usage-current--danger {
  color: rgb(var(--theme-danger-rgb));
}

.keys-form-dialog__usage-current--warning {
  color: rgb(var(--theme-warning-rgb));
}

.keys-form-dialog__usage-separator {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
  margin: 0 0.5rem;
}

.keys-form-dialog__progress-track {
  height: 0.375rem;
  width: 100%;
  overflow: hidden;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--theme-page-border) 78%, var(--theme-surface));
  margin-top: 0.25rem;
}

.keys-form-dialog__progress-bar {
  height: 100%;
  border-radius: 9999px;
  transition: width 0.2s ease;
}

.keys-form-dialog__progress-bar--danger {
  background: rgb(var(--theme-danger-rgb));
}

.keys-form-dialog__progress-bar--warning {
  background: rgb(var(--theme-warning-rgb));
}

.keys-form-dialog__progress-bar--success {
  background: rgb(var(--theme-success-rgb));
}

.keys-form-dialog__expiration-chip {
  border-radius: calc(var(--theme-button-radius) + 2px);
  font-size: 0.875rem;
  padding: 0.375rem 0.75rem;
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    border-color 0.18s ease;
}

.keys-form-dialog__expiration-chip--active {
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-accent) 90%, var(--theme-page-text));
}

.keys-form-dialog__expiration-chip--idle {
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.keys-form-dialog__expiration-chip--idle:hover,
.keys-form-dialog__expiration-chip--idle:focus-visible {
  background: color-mix(in srgb, var(--theme-page-border) 68%, var(--theme-surface));
  color: var(--theme-page-text);
  outline: none;
}
</style>
