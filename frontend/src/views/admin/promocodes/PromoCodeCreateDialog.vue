<template>
  <BaseDialog
    :show="show"
    :title="t('admin.promo.createCode')"
    width="normal"
    @close="emit('close')"
  >
    <form id="create-promo-form" class="space-y-4" @submit.prevent="emit('submit')">
      <div>
        <label class="input-label">
          {{ t('admin.promo.code') }}
          <span class="ml-1 text-xs font-normal text-gray-400">({{ t('admin.promo.autoGenerate') }})</span>
        </label>
        <input
          v-model="form.code"
          type="text"
          class="input font-mono uppercase"
          :placeholder="t('admin.promo.codePlaceholder')"
        />
      </div>
      <div>
        <label class="input-label">{{ t('admin.promo.bonusAmount') }}</label>
        <input
          v-model.number="form.bonus_amount"
          type="number"
          step="0.01"
          min="0"
          required
          class="input"
        />
      </div>
      <div>
        <label class="input-label">
          {{ t('admin.promo.maxUses') }}
          <span class="ml-1 text-xs font-normal text-gray-400">({{ t('admin.promo.zeroUnlimited') }})</span>
        </label>
        <input
          v-model.number="form.max_uses"
          type="number"
          min="0"
          class="input"
        />
      </div>
      <div>
        <label class="input-label">
          {{ t('admin.promo.expiresAt') }}
          <span class="ml-1 text-xs font-normal text-gray-400">({{ t('common.optional') }})</span>
        </label>
        <input
          v-model="form.expires_at_str"
          type="datetime-local"
          class="input"
        />
      </div>
      <div>
        <label class="input-label">
          {{ t('admin.promo.notes') }}
          <span class="ml-1 text-xs font-normal text-gray-400">({{ t('common.optional') }})</span>
        </label>
        <textarea
          v-model="form.notes"
          rows="2"
          class="input"
          :placeholder="t('admin.promo.notesPlaceholder')"
        ></textarea>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button type="submit" form="create-promo-form" :disabled="submitting" class="btn btn-primary">
          {{ submitting ? t('common.creating') : t('common.create') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PromoCodeCreateForm } from '../promoCodeForm'

defineProps<{
  show: boolean
  form: PromoCodeCreateForm
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
  submit: []
}>()

const { t } = useI18n()
</script>
