<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-affiliate-card__title text-lg font-semibold">
        {{ t('admin.settings.affiliate.title') }}
      </h2>
      <p class="settings-affiliate-card__description mt-1 text-sm">
        {{ t('admin.settings.affiliate.description') }}
      </p>
    </div>

    <div class="card-body space-y-5">
      <div class="settings-affiliate-card__section flex items-center justify-between gap-4">
        <div>
          <label class="settings-affiliate-card__label font-medium">
            {{ t('admin.settings.affiliate.enabled') }}
          </label>
          <p class="settings-affiliate-card__description text-sm">
            {{ t('admin.settings.affiliate.enabledHint') }}
          </p>
        </div>
        <Toggle
          v-model="form.affiliate_enabled"
          :aria-label="t('admin.settings.affiliate.enabled')"
        />
      </div>

      <div class="settings-affiliate-card__section grid gap-4 pt-4 md:grid-cols-2">
        <div>
          <label class="settings-affiliate-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.affiliate.rebateRate') }}
          </label>
          <input
            v-model.number="form.affiliate_rebate_rate"
            type="number"
            min="0"
            max="100"
            step="0.01"
            class="input"
          />
          <p class="settings-affiliate-card__description mt-1.5 text-xs">
            {{ t('admin.settings.affiliate.rebateRateHint') }}
          </p>
        </div>

        <div>
          <label class="settings-affiliate-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.affiliate.freezeHours') }}
          </label>
          <input
            v-model.number="form.affiliate_rebate_freeze_hours"
            type="number"
            min="0"
            step="1"
            class="input"
          />
          <p class="settings-affiliate-card__description mt-1.5 text-xs">
            {{ t('admin.settings.affiliate.freezeHoursHint') }}
          </p>
        </div>

        <div>
          <label class="settings-affiliate-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.affiliate.durationDays') }}
          </label>
          <input
            v-model.number="form.affiliate_rebate_duration_days"
            type="number"
            min="0"
            step="1"
            class="input"
          />
          <p class="settings-affiliate-card__description mt-1.5 text-xs">
            {{ t('admin.settings.affiliate.durationDaysHint') }}
          </p>
        </div>

        <div>
          <label class="settings-affiliate-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.affiliate.perInviteeCap') }}
          </label>
          <input
            v-model.number="form.affiliate_rebate_per_invitee_cap"
            type="number"
            min="0"
            step="0.01"
            class="input"
          />
          <p class="settings-affiliate-card__description mt-1.5 text-xs">
            {{ t('admin.settings.affiliate.perInviteeCapHint') }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { SettingsAffiliateFields } from './settingsForm'

defineProps<{
  form: SettingsAffiliateFields
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-affiliate-card__title,
.settings-affiliate-card__label,
.settings-affiliate-card__field-label {
  color: var(--theme-page-text);
}

.settings-affiliate-card__description {
  color: var(--theme-page-muted);
}

.settings-affiliate-card__section {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.settings-affiliate-card__section:first-child {
  border-top: 0;
}
</style>
