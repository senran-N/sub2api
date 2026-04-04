<template>
  <div class="card p-6">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">
          {{ t('admin.backup.operations.title') }}
        </h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t('admin.backup.operations.description') }}
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <div class="flex items-center gap-1">
          <label class="text-xs text-gray-600 dark:text-gray-400">
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

    <div class="overflow-x-auto">
      <table class="w-full min-w-[800px] text-sm">
        <thead>
          <tr class="border-b border-gray-200 text-left text-xs uppercase tracking-wide text-gray-500 dark:border-dark-700 dark:text-gray-400">
            <th class="py-2 pr-4">ID</th>
            <th class="py-2 pr-4">{{ t('admin.backup.columns.status') }}</th>
            <th class="py-2 pr-4">{{ t('admin.backup.columns.fileName') }}</th>
            <th class="py-2 pr-4">{{ t('admin.backup.columns.size') }}</th>
            <th class="py-2 pr-4">{{ t('admin.backup.columns.expiresAt') }}</th>
            <th class="py-2 pr-4">{{ t('admin.backup.columns.triggeredBy') }}</th>
            <th class="py-2 pr-4">{{ t('admin.backup.columns.startedAt') }}</th>
            <th class="py-2">{{ t('admin.backup.columns.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="record in backups"
            :key="record.id"
            class="border-b border-gray-100 align-top dark:border-dark-800"
          >
            <td class="py-3 pr-4 font-mono text-xs">{{ record.id }}</td>
            <td class="py-3 pr-4">
              <span class="rounded px-2 py-0.5 text-xs" :class="getBackupStatusClass(record.status)">
                {{ getStatusLabel(record) }}
              </span>
            </td>
            <td class="py-3 pr-4 text-xs">{{ record.file_name }}</td>
            <td class="py-3 pr-4 text-xs">{{ formatBackupSize(record.size_bytes) }}</td>
            <td class="py-3 pr-4 text-xs">
              {{ record.expires_at ? formatBackupDate(record.expires_at) : t('admin.backup.neverExpire') }}
            </td>
            <td class="py-3 pr-4 text-xs">
              {{
                record.triggered_by === 'scheduled'
                  ? t('admin.backup.trigger.scheduled')
                  : t('admin.backup.trigger.manual')
              }}
            </td>
            <td class="py-3 pr-4 text-xs">{{ formatBackupDate(record.started_at) }}</td>
            <td class="py-3 text-xs">
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
            <td colspan="8" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">
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
