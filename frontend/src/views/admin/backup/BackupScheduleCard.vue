<template>
  <div class="card p-6">
    <div class="mb-4">
      <h3 class="text-base font-semibold text-gray-900 dark:text-white">
        {{ t('admin.backup.schedule.title') }}
      </h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.backup.schedule.description') }}
      </p>
    </div>

    <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
      <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 md:col-span-2">
        <input v-model="form.enabled" type="checkbox" />
        <span>{{ t('admin.backup.schedule.enabled') }}</span>
      </label>
      <div>
        <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
          {{ t('admin.backup.schedule.cronExpr') }}
        </label>
        <input v-model="form.cron_expr" class="input w-full" placeholder="0 2 * * *" />
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.backup.schedule.cronHint') }}
        </p>
      </div>
      <div>
        <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
          {{ t('admin.backup.schedule.retainDays') }}
        </label>
        <input v-model.number="form.retain_days" type="number" min="0" class="input w-full" />
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.backup.schedule.retainDaysHint') }}
        </p>
      </div>
      <div>
        <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
          {{ t('admin.backup.schedule.retainCount') }}
        </label>
        <input v-model.number="form.retain_count" type="number" min="0" class="input w-full" />
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.backup.schedule.retainCountHint') }}
        </p>
      </div>
    </div>

    <div class="mt-4">
      <button type="button" class="btn btn-primary btn-sm" :disabled="saving" @click="emit('save')">
        {{ saving ? t('common.loading') : t('common.save') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { BackupScheduleConfig } from '@/api/admin/backup'

defineProps<{
  form: BackupScheduleConfig
  saving: boolean
}>()

const emit = defineEmits<{
  save: []
}>()

const { t } = useI18n()
</script>
