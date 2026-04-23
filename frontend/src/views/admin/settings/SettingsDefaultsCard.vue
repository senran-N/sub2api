<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-defaults-card__title text-lg font-semibold">
        {{ t('admin.settings.defaults.title') }}
      </h2>
      <p class="settings-defaults-card__description mt-1 text-sm">
        {{ t('admin.settings.defaults.description') }}
      </p>
    </div>
    <div class="card-body settings-defaults-card__body">
      <div class="settings-defaults-card__panel settings-defaults-card__metrics-grid">
        <div>
          <label class="settings-defaults-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.defaults.defaultBalance') }}
          </label>
          <input
            v-model.number="form.default_balance"
            type="number"
            step="0.01"
            min="0"
            class="input"
            placeholder="0.00"
          />
          <p class="settings-defaults-card__description mt-1.5 text-xs">
            {{ t('admin.settings.defaults.defaultBalanceHint') }}
          </p>
        </div>
        <div>
          <label class="settings-defaults-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.defaults.defaultConcurrency') }}
          </label>
          <input
            v-model.number="form.default_concurrency"
            type="number"
            min="1"
            class="input"
            placeholder="1"
          />
          <p class="settings-defaults-card__description mt-1.5 text-xs">
            {{ t('admin.settings.defaults.defaultConcurrencyHint') }}
          </p>
        </div>
      </div>

      <div class="settings-defaults-card__section">
        <div class="settings-defaults-card__section-header">
          <div>
            <label class="settings-defaults-card__label font-medium">
              {{ t('admin.settings.defaults.defaultSubscriptions') }}
            </label>
            <p class="settings-defaults-card__description text-sm">
              {{ t('admin.settings.defaults.defaultSubscriptionsHint') }}
            </p>
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="defaultSubscriptionGroupOptions.length === 0"
            @click="$emit('add-default-subscription')"
          >
            {{ t('admin.settings.defaults.addDefaultSubscription') }}
          </button>
        </div>

        <div
          v-if="defaultSubscriptionGroupOptions.length === 0"
          class="settings-defaults-card__empty text-sm"
        >
          <p class="font-medium">
            {{ t('admin.settings.defaults.subscriptionGroupsRequired') }}
          </p>
          <p class="mt-1">
            {{ t('admin.settings.defaults.subscriptionGroupsRequiredHint') }}
          </p>
        </div>

        <div
          v-if="form.default_subscriptions.length === 0"
          class="settings-defaults-card__empty text-sm"
        >
          {{ t('admin.settings.defaults.defaultSubscriptionsEmpty') }}
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="(item, index) in form.default_subscriptions"
            :key="`default-sub-${index}`"
            class="settings-defaults-card__subscription-item"
          >
            <div class="settings-defaults-card__subscription-grid">
              <div>
                <label class="settings-defaults-card__mini-label mb-1 block text-xs font-medium">
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
                      :rate-multiplier="toDefaultSubscriptionGroupOption(option).rate"
                    />
                    <span v-else class="settings-defaults-card__placeholder">
                      {{ t('admin.settings.defaults.subscriptionGroup') }}
                    </span>
                  </template>
                  <template #option="{ option, selected }">
                    <GroupOptionItem
                      :name="toDefaultSubscriptionGroupOption(option).label"
                      :platform="toDefaultSubscriptionGroupOption(option).platform"
                      :subscription-type="toDefaultSubscriptionGroupOption(option).subscriptionType"
                      :rate-multiplier="toDefaultSubscriptionGroupOption(option).rate"
                      :description="toDefaultSubscriptionGroupOption(option).description"
                      :selected="selected"
                    />
                  </template>
                </Select>
              </div>
              <div>
                <label class="settings-defaults-card__mini-label mb-1 block text-xs font-medium">
                  {{ t('admin.settings.defaults.subscriptionValidityDays') }}
                </label>
                <input
                  v-model.number="item.validity_days"
                  type="number"
                  min="1"
                  max="36500"
                  class="input settings-defaults-card__validity-input"
                />
              </div>
            </div>
            <div class="settings-defaults-card__subscription-action">
              <button
                type="button"
                class="btn btn-secondary settings-defaults-card__delete-button w-full"
                @click="$emit('remove-default-subscription', index)"
              >
                {{ t('common.delete') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import type { SettingsDefaultsFields } from './settingsForm'
import type { GroupPlatform, SubscriptionType } from '@/types'

interface DefaultSubscriptionGroupOptionView {
  label: string
  description: string | null
  platform: GroupPlatform
  subscriptionType: SubscriptionType
  rate: number
}

defineProps<{
  form: SettingsDefaultsFields
  defaultSubscriptionGroupOptions: SelectOption[]
  toDefaultSubscriptionGroupOption: (option: unknown) => DefaultSubscriptionGroupOptionView
}>()

defineEmits<{
  'add-default-subscription': []
  'remove-default-subscription': [index: number]
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-defaults-card__title,
.settings-defaults-card__label,
.settings-defaults-card__field-label,
.settings-defaults-card__mini-label {
  color: var(--theme-page-text);
}

.settings-defaults-card__body {
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}

.settings-defaults-card__description,
.settings-defaults-card__empty,
.settings-defaults-card__placeholder {
  color: var(--theme-page-muted);
}

.settings-defaults-card__panel,
.settings-defaults-card__section {
  border-radius: var(--theme-settings-card-panel-radius);
  padding: var(--theme-settings-card-panel-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 76%, transparent);
}

.settings-defaults-card__metrics-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: var(--theme-settings-card-panel-padding);
}

.settings-defaults-card__section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.settings-defaults-card__empty {
  border-radius: var(--theme-settings-defaults-empty-radius);
  padding: var(--theme-settings-defaults-empty-padding-y)
    var(--theme-settings-defaults-empty-padding-x);
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 88%, transparent);
}

.settings-defaults-card__empty + .settings-defaults-card__empty {
  margin-top: 0.75rem;
}

.settings-defaults-card__subscription-item {
  border-radius: var(--theme-settings-defaults-subscription-item-radius);
  padding: var(--theme-settings-defaults-subscription-item-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 90%, transparent);
}

.settings-defaults-card__subscription-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 0.75rem;
}

.settings-defaults-card__subscription-action {
  display: flex;
  align-items: flex-end;
  margin-top: 0.75rem;
}

.settings-defaults-card__validity-input {
  height: var(--theme-settings-defaults-validity-input-height);
}

.settings-defaults-card__delete-button {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 78%, var(--theme-page-text));
}

.settings-defaults-card__delete-button:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 92%, var(--theme-page-text));
}

@media (min-width: 768px) {
  .settings-defaults-card__metrics-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .settings-defaults-card__subscription-grid {
    grid-template-columns: minmax(0, 1fr) var(--theme-settings-defaults-validity-column-width);
  }

  .settings-defaults-card__subscription-item {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0.75rem;
    align-items: end;
  }

  .settings-defaults-card__subscription-action {
    margin-top: 0;
    min-width: 6rem;
  }
}
</style>
