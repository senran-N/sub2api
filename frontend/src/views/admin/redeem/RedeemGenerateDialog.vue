<template>
  <BaseDialog
    :show="show"
    :title="t('admin.redeem.generateCodesTitle')"
    width="narrow"
    close-on-click-outside
    @close="emit('close')"
  >
    <form id="redeem-generate-form" class="space-y-4" @submit.prevent="emit('submit')">
      <div>
        <label class="input-label">{{ t('admin.redeem.codeType') }}</label>
        <Select v-model="form.type" :options="typeOptions" />
      </div>

      <div v-if="form.type !== 'subscription' && form.type !== 'invitation'">
        <label class="input-label">
          {{
            form.type === 'balance'
              ? t('admin.redeem.amount')
              : t('admin.redeem.columns.value')
          }}
        </label>
        <input
          v-model.number="form.value"
          type="number"
          :step="form.type === 'balance' ? '0.01' : '1'"
          :min="form.type === 'balance' ? '0.01' : '1'"
          required
          class="input"
        />
      </div>

      <div v-if="form.type === 'invitation'" class="redeem-generate-dialog__hint">
        <p class="redeem-generate-dialog__hint-text text-sm">
          {{ t('admin.redeem.invitationHint') }}
        </p>
      </div>

      <template v-if="form.type === 'subscription'">
        <div>
          <label class="input-label">{{ t('admin.redeem.selectGroup') }}</label>
          <Select
            v-model="form.group_id"
            :options="subscriptionGroupOptions"
            :placeholder="t('admin.redeem.selectGroupPlaceholder')"
          >
            <template #selected="{ option }">
              <GroupBadge
                v-if="option"
                :name="(option as unknown as RedeemGroupOption).label"
                :platform="(option as unknown as RedeemGroupOption).platform"
                :subscription-type="(option as unknown as RedeemGroupOption).subscriptionType"
                :rate-multiplier="(option as unknown as RedeemGroupOption).rate"
              />
              <span v-else class="redeem-generate-dialog__placeholder">
                {{ t('admin.redeem.selectGroupPlaceholder') }}
              </span>
            </template>
            <template #option="{ option, selected }">
              <GroupOptionItem
                :name="(option as unknown as RedeemGroupOption).label"
                :platform="(option as unknown as RedeemGroupOption).platform"
                :subscription-type="(option as unknown as RedeemGroupOption).subscriptionType"
                :rate-multiplier="(option as unknown as RedeemGroupOption).rate"
                :description="(option as unknown as RedeemGroupOption).description"
                :selected="selected"
              />
            </template>
          </Select>
        </div>
        <div>
          <label class="input-label">{{ t('admin.redeem.validityDays') }}</label>
          <input
            v-model.number="form.validity_days"
            type="number"
            min="1"
            max="365"
            required
            class="input"
          />
        </div>
      </template>

      <div>
        <label class="input-label">{{ t('admin.redeem.count') }}</label>
        <input
          v-model.number="form.count"
          type="number"
          min="1"
          max="100"
          required
          class="input"
        />
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button type="submit" form="redeem-generate-form" :disabled="submitting" class="btn btn-primary">
          {{ submitting ? t('admin.redeem.generating') : t('admin.redeem.generate') }}
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
import type { RedeemGenerationForm, RedeemGroupOption } from '../redeemForm'

defineProps<{
  show: boolean
  form: RedeemGenerationForm
  typeOptions: Array<{ value: string; label: string }>
  subscriptionGroupOptions: RedeemGroupOption[]
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
  submit: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.redeem-generate-dialog__hint {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-info-rgb)) 26%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-settings-card-panel-padding);
}

.redeem-generate-dialog__hint-text {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.redeem-generate-dialog__placeholder {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}
</style>
