<template>
  <div class="mt-4 border-t pt-4">
    <div>
      <label class="input-label">{{ t('admin.groups.subscription.type') }}</label>
      <Select
        v-model="form.subscription_type"
        :options="subscriptionTypeOptions"
        :disabled="subscriptionTypeDisabled"
      />
      <p class="input-hint">{{ subscriptionTypeHint }}</p>
    </div>

    <div
      v-if="form.subscription_type === 'subscription'"
      class="group-subscription-section__limits space-y-4 border-l-2 pl-4"
    >
      <div>
        <label class="input-label">{{ t('admin.groups.subscription.dailyLimit') }}</label>
        <input
          v-model.number="form.daily_limit_usd"
          type="number"
          step="0.01"
          min="0"
          class="input"
          :placeholder="t('admin.groups.subscription.noLimit')"
        />
      </div>
      <div>
        <label class="input-label">{{ t('admin.groups.subscription.weeklyLimit') }}</label>
        <input
          v-model.number="form.weekly_limit_usd"
          type="number"
          step="0.01"
          min="0"
          class="input"
          :placeholder="t('admin.groups.subscription.noLimit')"
        />
      </div>
      <div>
        <label class="input-label">{{ t('admin.groups.subscription.monthlyLimit') }}</label>
        <input
          v-model.number="form.monthly_limit_usd"
          type="number"
          step="0.01"
          min="0"
          class="input"
          :placeholder="t('admin.groups.subscription.noLimit')"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import type { CreateGroupForm, EditGroupForm } from '../groupsForm'

defineProps<{
  form: CreateGroupForm | EditGroupForm
  subscriptionTypeOptions: SelectOption[]
  subscriptionTypeHint: string
  subscriptionTypeDisabled?: boolean
}>()

const { t } = useI18n()
</script>

<style scoped>
.group-subscription-section__limits {
  border-left-color: color-mix(in srgb, var(--theme-accent) 28%, var(--theme-card-border));
}
</style>
