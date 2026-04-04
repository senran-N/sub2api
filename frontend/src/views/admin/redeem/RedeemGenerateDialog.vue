<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="fixed inset-0 bg-black/50" @click="emit('close')"></div>
      <div
        class="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-xl dark:bg-dark-800"
      >
        <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('admin.redeem.generateCodesTitle') }}
        </h2>
        <form class="space-y-4" @submit.prevent="emit('submit')">
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

          <div v-if="form.type === 'invitation'" class="rounded-lg bg-blue-50 p-3 dark:bg-blue-900/20">
            <p class="text-sm text-blue-700 dark:text-blue-300">
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
                  <span v-else class="text-gray-400">
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

          <div class="flex justify-end gap-3 pt-2">
            <button type="button" class="btn btn-secondary" @click="emit('close')">
              {{ t('common.cancel') }}
            </button>
            <button type="submit" :disabled="submitting" class="btn btn-primary">
              {{ submitting ? t('admin.redeem.generating') : t('admin.redeem.generate') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
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
