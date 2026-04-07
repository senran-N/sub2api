<template>
  <div v-if="form.platform === 'antigravity' || form.platform === 'gemini'" class="border-t pt-4">
    <label class="theme-text-default mb-2 block font-medium">
      {{ t('admin.groups.imagePricing.title') }}
    </label>
    <p class="theme-text-muted mb-3 text-xs">
      {{ t('admin.groups.imagePricing.description') }}
    </p>
    <div class="grid grid-cols-3 gap-3">
      <div>
        <label class="input-label">1K ($)</label>
        <input
          v-model.number="form.image_price_1k"
          type="number"
          step="0.001"
          min="0"
          class="input"
          placeholder="0.134"
        />
      </div>
      <div>
        <label class="input-label">2K ($)</label>
        <input
          v-model.number="form.image_price_2k"
          type="number"
          step="0.001"
          min="0"
          class="input"
          placeholder="0.201"
        />
      </div>
      <div>
        <label class="input-label">4K ($)</label>
        <input
          v-model.number="form.image_price_4k"
          type="number"
          step="0.001"
          min="0"
          class="input"
          placeholder="0.268"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CreateGroupForm, EditGroupForm } from './groupsForm'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const { t } = useI18n()
</script>
