<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-auth-source-defaults-card__title text-lg font-semibold">
        {{ t('admin.settings.defaults.authSourceDefaultsTitle') }}
      </h2>
      <p class="settings-auth-source-defaults-card__description mt-1 text-sm">
        {{ t('admin.settings.defaults.authSourceDefaultsDescription') }}
      </p>
    </div>

    <div class="card-body settings-auth-source-defaults-card__body">
      <div
        v-if="defaultSubscriptionGroupOptions.length === 0"
        class="settings-auth-source-defaults-card__empty text-sm"
      >
        <p class="font-medium">
          {{ t('admin.settings.defaults.subscriptionGroupsRequired') }}
        </p>
        <p class="mt-1">
          {{ t('admin.settings.defaults.subscriptionGroupsRequiredHint') }}
        </p>
      </div>

      <div class="settings-auth-source-defaults-card__grid">
      <div
        v-for="section in sections"
        :key="section.source"
        class="settings-auth-source-defaults-card__source-panel"
      >
        <div class="settings-auth-source-defaults-card__source-header">
          <div>
            <h3 class="settings-auth-source-defaults-card__section-title text-base font-semibold">
              {{ t(section.titleKey) }}
            </h3>
            <p class="settings-auth-source-defaults-card__description text-sm">
              {{ t(section.descriptionKey) }}
            </p>
          </div>
        </div>

        <div class="settings-auth-source-defaults-card__source-body">
        <div class="settings-auth-source-defaults-card__metrics-grid">
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

        <div class="settings-auth-source-defaults-card__toggle-grid">
          <div class="settings-auth-source-defaults-card__toggle-panel">
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

          <div class="settings-auth-source-defaults-card__toggle-panel">
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

        <div class="settings-auth-source-defaults-card__subscriptions-panel">
          <div class="settings-auth-source-defaults-card__subscriptions-header">
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
              class="settings-auth-source-defaults-card__subscription-item"
            >
              <div class="settings-auth-source-defaults-card__subscription-grid">
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
              </div>

              <div class="settings-auth-source-defaults-card__subscription-action">
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

.settings-auth-source-defaults-card__body {
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}

.settings-auth-source-defaults-card__grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: var(--theme-settings-card-body-padding);
}

.settings-auth-source-defaults-card__source-panel {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-settings-card-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 72%, transparent);
}

.settings-auth-source-defaults-card__source-header {
  padding: var(--theme-settings-card-panel-padding);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-auth-source-defaults-card__source-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: var(--theme-settings-card-panel-padding);
}

.settings-auth-source-defaults-card__metrics-grid,
.settings-auth-source-defaults-card__toggle-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 0.75rem;
}

.settings-auth-source-defaults-card__toggle-panel,
.settings-auth-source-defaults-card__subscriptions-panel,
.settings-auth-source-defaults-card__subscription-item {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  border-radius: var(--theme-settings-defaults-subscription-item-radius);
  padding: var(--theme-settings-defaults-subscription-item-padding);
  background: color-mix(in srgb, var(--theme-surface) 90%, transparent);
}

.settings-auth-source-defaults-card__subscriptions-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.settings-auth-source-defaults-card__subscription-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 0.75rem;
}

.settings-auth-source-defaults-card__subscription-action {
  display: flex;
  align-items: flex-end;
  margin-top: 0.75rem;
}

.settings-auth-source-defaults-card__empty {
  border-radius: var(--theme-settings-defaults-empty-radius);
  padding: var(--theme-settings-defaults-empty-padding-y)
    var(--theme-settings-defaults-empty-padding-x);
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 88%, transparent);
}

@media (min-width: 768px) {
  .settings-auth-source-defaults-card__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .settings-auth-source-defaults-card__metrics-grid,
  .settings-auth-source-defaults-card__toggle-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .settings-auth-source-defaults-card__subscription-grid {
    grid-template-columns: minmax(0, 1fr) var(--theme-settings-defaults-validity-column-width);
  }

  .settings-auth-source-defaults-card__subscription-item {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0.75rem;
    align-items: end;
  }

  .settings-auth-source-defaults-card__subscription-action {
    margin-top: 0;
    min-width: 6rem;
  }
}
</style>
