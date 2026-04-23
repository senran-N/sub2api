<template>
  <div class="card">
    <div class="settings-auth-source-defaults-card__header">
      <h2 class="settings-auth-source-defaults-card__title text-lg font-semibold">
        {{ t('admin.settings.defaults.authSourceDefaultsTitle') }}
      </h2>
      <p class="settings-auth-source-defaults-card__description mt-1 text-sm">
        {{ t('admin.settings.defaults.authSourceDefaultsDescription') }}
      </p>
    </div>

    <div class="settings-auth-source-defaults-card__content space-y-6">
      <div
        v-for="section in sections"
        :key="section.source"
        class="settings-auth-source-defaults-card__section"
      >
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h3 class="settings-auth-source-defaults-card__section-title text-base font-semibold">
              {{ t(section.titleKey) }}
            </h3>
            <p class="settings-auth-source-defaults-card__description text-sm">
              {{ t(section.descriptionKey) }}
            </p>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="settings-auth-source-defaults-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.defaults.defaultBalance') }}
            </label>
            <input
              :value="getBalance(section.source)"
              type="number"
              step="0.01"
              min="0"
              class="input"
              placeholder="0.00"
              @input="setBalance(section.source, ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="settings-auth-source-defaults-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.defaults.defaultConcurrency') }}
            </label>
            <input
              :value="getConcurrency(section.source)"
              type="number"
              min="1"
              class="input"
              placeholder="5"
              @input="setConcurrency(section.source, ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>

        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
          <div class="flex items-center justify-between rounded-lg border px-4 py-3">
            <div>
              <label class="settings-auth-source-defaults-card__field-label font-medium">
                {{ t('admin.settings.defaults.grantOnSignup') }}
              </label>
              <p class="settings-auth-source-defaults-card__description text-sm">
                {{ t('admin.settings.defaults.grantOnSignupHint') }}
              </p>
            </div>
            <Toggle
              :model-value="getGrantOnSignup(section.source)"
              :aria-label="t('admin.settings.defaults.grantOnSignup')"
              @update:model-value="setGrantOnSignup(section.source, $event)"
            />
          </div>

          <div class="flex items-center justify-between rounded-lg border px-4 py-3">
            <div>
              <label class="settings-auth-source-defaults-card__field-label font-medium">
                {{ t('admin.settings.defaults.grantOnFirstBind') }}
              </label>
              <p class="settings-auth-source-defaults-card__description text-sm">
                {{ t('admin.settings.defaults.grantOnFirstBindHint') }}
              </p>
            </div>
            <Toggle
              :model-value="getGrantOnFirstBind(section.source)"
              :aria-label="t('admin.settings.defaults.grantOnFirstBind')"
              @update:model-value="setGrantOnFirstBind(section.source, $event)"
            />
          </div>
        </div>

        <div class="mt-4">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="settings-auth-source-defaults-card__field-label font-medium">
                {{ t('admin.settings.defaults.defaultSubscriptions') }}
              </label>
              <p class="settings-auth-source-defaults-card__description text-sm">
                {{ t('admin.settings.defaults.authSourceSubscriptionsHint') }}
              </p>
            </div>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="defaultSubscriptionGroupOptions.length === 0"
              @click="$emit('add-auth-source-default-subscription', section.source)"
            >
              {{ t('admin.settings.defaults.addDefaultSubscription') }}
            </button>
          </div>

          <div
            v-if="getSubscriptions(section.source).length === 0"
            class="settings-auth-source-defaults-card__empty text-sm"
          >
            {{ t('admin.settings.defaults.defaultSubscriptionsEmpty') }}
          </div>

          <div v-else class="space-y-3">
            <div
              v-for="(item, index) in getSubscriptions(section.source)"
              :key="`${section.source}-${index}`"
              class="settings-auth-source-defaults-card__subscription-item grid grid-cols-1 gap-3 md:grid-cols-[1fr_var(--theme-settings-defaults-validity-column-width)_auto]"
            >
              <div>
                <label class="settings-auth-source-defaults-card__mini-label mb-1 block text-xs font-medium">
                  {{ t('admin.settings.defaults.subscriptionGroup') }}
                </label>
                <Select
                  v-model="item.group_id"
                  class="default-sub-group-select"
                  :options="defaultSubscriptionGroupOptions"
                  :placeholder="t('admin.settings.defaults.subscriptionGroup')"
                >
                  <template #selected="{ option }">
                    <GroupBadge
                      v-if="option"
                      :name="toDefaultSubscriptionGroupOption(option).label"
                      :platform="toDefaultSubscriptionGroupOption(option).platform"
                      :subscription-type="toDefaultSubscriptionGroupOption(option).subscriptionType"
                      :rate-multiplier="toDefaultSubscriptionGroupOption(option).rate || undefined"
                    />
                    <span v-else class="settings-auth-source-defaults-card__placeholder">
                      {{ t('admin.settings.defaults.subscriptionGroup') }}
                    </span>
                  </template>
                  <template #option="{ option, selected }">
                    <GroupOptionItem
                      :name="toDefaultSubscriptionGroupOption(option).label"
                      :platform="toDefaultSubscriptionGroupOption(option).platform"
                      :subscription-type="toDefaultSubscriptionGroupOption(option).subscriptionType"
                      :rate-multiplier="toDefaultSubscriptionGroupOption(option).rate || undefined"
                      :description="toDefaultSubscriptionGroupOption(option).description"
                      :selected="selected"
                    />
                  </template>
                </Select>
              </div>

              <div>
                <label class="settings-auth-source-defaults-card__mini-label mb-1 block text-xs font-medium">
                  {{ t('admin.settings.defaults.subscriptionValidityDays') }}
                </label>
                <input
                  v-model.number="item.validity_days"
                  type="number"
                  min="1"
                  max="36500"
                  class="input settings-auth-source-defaults-card__validity-input"
                />
              </div>

              <div class="flex items-end">
                <button
                  type="button"
                  class="btn btn-secondary w-full"
                  @click="$emit('remove-auth-source-default-subscription', section.source, index)"
                >
                  {{ t('common.delete') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import type { AuthSourceType, DefaultSubscriptionSetting } from '@/api/admin/settings'
import type { GroupPlatform, SubscriptionType } from '@/types'
import type { SettingsForm } from './settingsForm'

interface DefaultSubscriptionGroupOptionView {
  label: string
  description: string | null
  platform: GroupPlatform
  subscriptionType: SubscriptionType
  rate: number
}

const props = defineProps<{
  form: SettingsForm
  defaultSubscriptionGroupOptions: SelectOption[]
  toDefaultSubscriptionGroupOption: (option: unknown) => DefaultSubscriptionGroupOptionView
}>()

defineEmits<{
  'add-auth-source-default-subscription': [source: AuthSourceType]
  'remove-auth-source-default-subscription': [source: AuthSourceType, index: number]
}>()

const { t } = useI18n()

const sections: Array<{ source: AuthSourceType; titleKey: string; descriptionKey: string }> = [
  {
    source: 'email',
    titleKey: 'admin.settings.defaults.authSourceEmailTitle',
    descriptionKey: 'admin.settings.defaults.authSourceEmailDescription'
  },
  {
    source: 'linuxdo',
    titleKey: 'admin.settings.defaults.authSourceLinuxDoTitle',
    descriptionKey: 'admin.settings.defaults.authSourceLinuxDoDescription'
  },
  {
    source: 'oidc',
    titleKey: 'admin.settings.defaults.authSourceOidcTitle',
    descriptionKey: 'admin.settings.defaults.authSourceOidcDescription'
  },
  {
    source: 'wechat',
    titleKey: 'admin.settings.defaults.authSourceWeChatTitle',
    descriptionKey: 'admin.settings.defaults.authSourceWeChatDescription'
  }
]

function record() {
  return props.form as unknown as Record<string, unknown>
}

function numberField(source: AuthSourceType, suffix: 'balance' | 'concurrency'): number {
  return Number(record()[`auth_source_default_${source}_${suffix}`] ?? (suffix === 'balance' ? 0 : 5))
}

function setNumberField(source: AuthSourceType, suffix: 'balance' | 'concurrency', value: string) {
  const parsed = suffix === 'balance' ? Number(value || 0) : Math.max(1, Math.floor(Number(value || 1)))
  record()[`auth_source_default_${source}_${suffix}`] = Number.isFinite(parsed) ? parsed : (suffix === 'balance' ? 0 : 5)
}

function boolField(source: AuthSourceType, suffix: 'grant_on_signup' | 'grant_on_first_bind'): boolean {
  return record()[`auth_source_default_${source}_${suffix}`] === true
}

function setBoolField(source: AuthSourceType, suffix: 'grant_on_signup' | 'grant_on_first_bind', value: boolean) {
  record()[`auth_source_default_${source}_${suffix}`] = value
}

function getSubscriptions(source: AuthSourceType): DefaultSubscriptionSetting[] {
  return (record()[`auth_source_default_${source}_subscriptions`] as DefaultSubscriptionSetting[] | undefined) ?? []
}

function getBalance(source: AuthSourceType) {
  return numberField(source, 'balance')
}

function setBalance(source: AuthSourceType, value: string) {
  setNumberField(source, 'balance', value)
}

function getConcurrency(source: AuthSourceType) {
  return numberField(source, 'concurrency')
}

function setConcurrency(source: AuthSourceType, value: string) {
  setNumberField(source, 'concurrency', value)
}

function getGrantOnSignup(source: AuthSourceType) {
  return boolField(source, 'grant_on_signup')
}

function setGrantOnSignup(source: AuthSourceType, value: boolean) {
  setBoolField(source, 'grant_on_signup', value)
}

function getGrantOnFirstBind(source: AuthSourceType) {
  return boolField(source, 'grant_on_first_bind')
}

function setGrantOnFirstBind(source: AuthSourceType, value: boolean) {
  setBoolField(source, 'grant_on_first_bind', value)
}
</script>

<style scoped>
.settings-auth-source-defaults-card__header,
.settings-auth-source-defaults-card__section {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-auth-source-defaults-card__header {
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  padding: var(--theme-settings-card-header-padding-y) var(--theme-settings-card-header-padding-x);
}

.settings-auth-source-defaults-card__content {
  padding: var(--theme-settings-card-content-padding-y) var(--theme-settings-card-content-padding-x);
}

.settings-auth-source-defaults-card__section {
  padding-top: 1.25rem;
}

.settings-auth-source-defaults-card__title,
.settings-auth-source-defaults-card__section-title,
.settings-auth-source-defaults-card__field-label,
.settings-auth-source-defaults-card__mini-label {
  color: var(--theme-page-text);
}

.settings-auth-source-defaults-card__description,
.settings-auth-source-defaults-card__empty,
.settings-auth-source-defaults-card__placeholder {
  color: var(--theme-page-text-secondary);
}

.settings-auth-source-defaults-card__subscription-item {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  border-radius: 0.75rem;
  padding: 0.9rem;
}
</style>
