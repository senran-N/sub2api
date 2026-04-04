<template>
  <div v-if="form.platform === 'sora'" class="border-t pt-4">
    <label class="mb-2 block font-medium text-gray-700 dark:text-gray-300">
      {{ t('admin.groups.soraPricing.title') }}
    </label>
    <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
      {{ t('admin.groups.soraPricing.description') }}
    </p>
    <div class="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
      <div>
        <label class="input-label">{{ t('admin.groups.soraPricing.image360') }}</label>
        <input
          v-model.number="form.sora_image_price_360"
          type="number"
          step="0.001"
          min="0"
          class="input"
          placeholder="0.05"
        />
      </div>
      <div>
        <label class="input-label">{{ t('admin.groups.soraPricing.image540') }}</label>
        <input
          v-model.number="form.sora_image_price_540"
          type="number"
          step="0.001"
          min="0"
          class="input"
          placeholder="0.08"
        />
      </div>
    </div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <div>
        <label class="input-label">{{ t('admin.groups.soraPricing.video') }}</label>
        <input
          v-model.number="form.sora_video_price_per_request"
          type="number"
          step="0.001"
          min="0"
          class="input"
          placeholder="0.5"
        />
      </div>
      <div>
        <label class="input-label">{{ t('admin.groups.soraPricing.videoHd') }}</label>
        <input
          v-model.number="form.sora_video_price_per_request_hd"
          type="number"
          step="0.001"
          min="0"
          class="input"
          placeholder="0.8"
        />
      </div>
    </div>
    <div class="mt-3">
      <label class="input-label">{{ t('admin.groups.soraPricing.storageQuota') }}</label>
      <div class="flex items-center gap-2">
        <input
          v-model.number="form.sora_storage_quota_gb"
          type="number"
          step="0.1"
          min="0"
          class="input"
          placeholder="10"
        />
        <span class="shrink-0 text-sm text-gray-500">GB</span>
      </div>
      <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
        {{ t('admin.groups.soraPricing.storageQuotaHint') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CreateGroupForm, EditGroupForm } from '../groupsForm'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const { t } = useI18n()
</script>
