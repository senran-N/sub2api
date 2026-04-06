<template>
  <div class="card backup-operations-card__root">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="backup-operations-card__title text-base font-semibold">
          {{ t('admin.backup.operations.title') }}
        </h3>
        <p class="backup-operations-card__description mt-1 text-sm">
          {{ t('admin.backup.operations.description') }}
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <div class="flex items-center gap-1">
          <label class="backup-operations-card__label text-xs">
            {{ t('admin.backup.operations.expireDays') }}
          </label>
          <input
            :value="manualExpireDays"
            type="number"
            min="0"
            class="input w-20 text-xs"
            @input="handleExpireDaysInput"
          />
        </div>
        <button type="button" class="btn btn-primary btn-sm" :disabled="creating" @click="emit('create')">
          {{ creating ? t('admin.backup.operations.backing') : t('admin.backup.operations.createBackup') }}
        </button>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="emit('refresh')">
          {{ loading ? t('common.loading') : t('common.refresh') }}
        </button>
      </div>
    </div>

      <div class="table-container table-wrapper overflow-x-auto">
        <table class="table w-full backup-operations-card__table text-sm">
        <thead>
          <tr>
            <th class="backup-operations-card__head-cell">ID</th>
            <th class="backup-operations-card__head-cell">{{ t('admin.backup.columns.status') }}</th>
            <th class="backup-operations-card__head-cell">{{ t('admin.backup.columns.fileName') }}</th>
            <th class="backup-operations-card__head-cell">{{ t('admin.backup.columns.size') }}</th>
            <th class="backup-operations-card__head-cell">{{ t('admin.backup.columns.expiresAt') }}</th>
            <th class="backup-operations-card__head-cell">{{ t('admin.backup.columns.triggeredBy') }}</th>
            <th class="backup-operations-card__head-cell">{{ t('admin.backup.columns.startedAt') }}</th>
            <th class="backup-operations-card__head-cell">{{ t('admin.backup.columns.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="record in backups" :key="record.id" class="align-top">
            <td class="backup-operations-card__mono-cell">{{ record.id }}</td>
            <td>
              <span class="theme-chip theme-chip--compact" :class="getBackupStatusClass(record.status)">
                {{ getStatusLabel(record) }}
              </span>
            </td>
            <td class="backup-operations-card__mono-cell">{{ record.file_name }}</td>
            <td class="backup-operations-card__meta-cell">{{ formatBackupSize(record.size_bytes) }}</td>
            <td class="backup-operations-card__meta-cell">
              {{ record.expires_at ? formatBackupDate(record.expires_at) : t('admin.backup.neverExpire') }}
            </td>
            <td class="backup-operations-card__meta-cell">
              {{
                record.triggered_by === 'scheduled'
                  ? t('admin.backup.trigger.scheduled')
                  : t('admin.backup.trigger.manual')
              }}
            </td>
            <td class="backup-operations-card__meta-cell">{{ formatBackupDate(record.started_at) }}</td>
            <td class="backup-operations-card__actions-cell">
              <div class="flex flex-wrap gap-1">
                <button
                  v-if="record.status === 'completed'"
                  type="button"
                  class="btn btn-secondary btn-xs"
                  @click="emit('download', record.id)"
                >
                  {{ t('admin.backup.actions.download') }}
                </button>
                <button
                  v-if="record.status === 'completed'"
                  type="button"
                  class="btn btn-secondary btn-xs"
                  :disabled="restoringId === record.id"
                  @click="emit('restore', record.id)"
                >
                  {{
                    restoringId === record.id
                      ? t('common.loading')
                      : t('admin.backup.actions.restore')
                  }}
                </button>
                <button
                  type="button"
                  class="btn btn-danger btn-xs"
                  @click="emit('remove', record.id)"
                >
                  {{ t('common.delete') }}
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="backups.length === 0">
            <td colspan="8" class="backup-operations-card__empty text-center text-sm">
              {{ t('admin.backup.empty') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { BackupRecord } from '@/api/admin/backup'
import {
  formatBackupDate,
  formatBackupSize,
  getBackupStatusClass
} from '../backupView'

defineProps<{
  backups: BackupRecord[]
  loading: boolean
  creating: boolean
  restoringId: string
  manualExpireDays: number
}>()

const emit = defineEmits<{
  'update:manualExpireDays': [value: number]
  create: []
  refresh: []
  download: [id: string]
  restore: [id: string]
  remove: [id: string]
}>()

const { t } = useI18n()

const getStatusLabel = (record: BackupRecord) => {
  if (record.status === 'running' && record.progress) {
    return t(`admin.backup.progress.${record.progress}`)
  }
  return t(`admin.backup.status.${record.status}`)
}

const handleExpireDaysInput = (event: Event) => {
  const { value } = event.target as HTMLInputElement
  emit('update:manualExpireDays', value === '' ? 0 : Number(value))
}
</script>

<style scoped>
.backup-operations-card__title {
  color: var(--theme-page-text);
}

.backup-operations-card__description,
.backup-operations-card__label,
.backup-operations-card__empty {
  color: var(--theme-page-muted);
}

.backup-operations-card__head-cell {
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.backup-operations-card__mono-cell {
  font-family: var(--theme-font-mono);
  font-size: 0.75rem;
  color: var(--theme-page-text);
}

.backup-operations-card__meta-cell,
.backup-operations-card__actions-cell {
  font-size: 0.75rem;
  color: color-mix(in srgb, var(--theme-page-text) 78%, transparent);
}

.backup-operations-card__root {
  padding: var(--theme-backup-operations-card-padding);
}

.backup-operations-card__table {
  min-width: var(--theme-backup-operations-table-min-width);
}

.backup-operations-card__empty {
  padding: var(--theme-backup-operations-empty-padding-y) 0;
}
</style>
