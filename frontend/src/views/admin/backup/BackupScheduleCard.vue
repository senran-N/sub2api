<template>
  <div class="backup-schedule-card card">
    <div class="mb-4">
      <h3 class="backup-schedule-card__title text-base font-semibold">
        {{ t('admin.backup.schedule.title') }}
      </h3>
      <p class="backup-schedule-card__description mt-1 text-sm">
        {{ t('admin.backup.schedule.description') }}
      </p>
    </div>

    <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
      <label class="backup-schedule-card__checkbox inline-flex items-center gap-2 text-sm md:col-span-2">
        <input v-model="form.enabled" type="checkbox" />
        <span>{{ t('admin.backup.schedule.enabled') }}</span>
      </label>
      <div>
        <label class="backup-schedule-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.schedule.cronExpr') }}
        </label>
        <input v-model="form.cron_expr" class="input w-full" placeholder="0 2 * * *" />
        <p class="backup-schedule-card__description mt-1 text-xs">
          {{ t('admin.backup.schedule.cronHint') }}
        </p>
      </div>
      <div>
        <label class="backup-schedule-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.schedule.retainDays') }}
        </label>
        <input v-model.number="form.retain_days" type="number" min="0" class="input w-full" />
        <p class="backup-schedule-card__description mt-1 text-xs">
          {{ t('admin.backup.schedule.retainDaysHint') }}
        </p>
      </div>
      <div>
        <label class="backup-schedule-card__field-label mb-1 block text-xs font-medium">
          {{ t('admin.backup.schedule.retainCount') }}
        </label>
        <input v-model.number="form.retain_count" type="number" min="0" class="input w-full" />
        <p class="backup-schedule-card__description mt-1 text-xs">
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

<style scoped>
.backup-schedule-card__title,
.backup-schedule-card__field-label,
.backup-schedule-card__checkbox {
  color: var(--theme-page-text);
}

.backup-schedule-card__description {
  color: var(--theme-page-muted);
}

.backup-schedule-card {
  padding: var(--theme-settings-card-panel-padding);
}
</style>
