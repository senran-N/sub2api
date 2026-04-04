<template>
  <div class="card p-6">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">
          {{ t('admin.settings.soraS3.title') }}
        </h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t('admin.settings.soraS3.description') }}
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <button type="button" class="btn btn-secondary btn-sm" @click="emit('create')">
          {{ t('admin.settings.soraS3.newProfile') }}
        </button>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="emit('reload')">
          {{ loading ? t('common.loading') : t('admin.settings.soraS3.reloadProfiles') }}
        </button>
      </div>
    </div>

    <div class="overflow-x-auto">
      <table class="w-full min-w-[1000px] text-sm">
        <thead>
          <tr class="border-b border-gray-200 text-left text-xs uppercase tracking-wide text-gray-500 dark:border-dark-700 dark:text-gray-400">
            <th class="py-2 pr-4">{{ t('admin.settings.soraS3.columns.profile') }}</th>
            <th class="py-2 pr-4">{{ t('admin.settings.soraS3.columns.active') }}</th>
            <th class="py-2 pr-4">{{ t('admin.settings.soraS3.columns.endpoint') }}</th>
            <th class="py-2 pr-4">{{ t('admin.settings.soraS3.columns.bucket') }}</th>
            <th class="py-2 pr-4">{{ t('admin.settings.soraS3.columns.quota') }}</th>
            <th class="py-2 pr-4">{{ t('admin.settings.soraS3.columns.updatedAt') }}</th>
            <th class="py-2">{{ t('admin.settings.soraS3.columns.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="profile in profiles"
            :key="profile.profile_id"
            class="border-b border-gray-100 align-top dark:border-dark-800"
          >
            <td class="py-3 pr-4">
              <div class="font-mono text-xs">{{ profile.profile_id }}</div>
              <div class="mt-1 text-xs text-gray-600 dark:text-gray-400">{{ profile.name }}</div>
            </td>
            <td class="py-3 pr-4">
              <span
                class="rounded px-2 py-0.5 text-xs"
                :class="profile.is_active ? activeBadgeClass : inactiveBadgeClass"
              >
                {{ profile.is_active ? t('common.enabled') : t('common.disabled') }}
              </span>
            </td>
            <td class="py-3 pr-4 text-xs">
              <div>{{ profile.endpoint || '-' }}</div>
              <div class="mt-1 text-gray-500 dark:text-gray-400">{{ profile.region || '-' }}</div>
            </td>
            <td class="py-3 pr-4 text-xs">{{ profile.bucket || '-' }}</td>
            <td class="py-3 pr-4 text-xs">{{ formatStorageQuotaGB(profile.default_storage_quota_bytes) }}</td>
            <td class="py-3 pr-4 text-xs">{{ formatDataManagementDate(profile.updated_at) }}</td>
            <td class="py-3 text-xs">
              <div class="flex flex-wrap gap-2">
                <button type="button" class="btn btn-secondary btn-xs" @click="emit('edit', profile.profile_id)">
                  {{ t('common.edit') }}
                </button>
                <button
                  v-if="!profile.is_active"
                  type="button"
                  class="btn btn-secondary btn-xs"
                  :disabled="activating"
                  @click="emit('activate', profile.profile_id)"
                >
                  {{ t('admin.settings.soraS3.activateProfile') }}
                </button>
                <button
                  type="button"
                  class="btn btn-danger btn-xs"
                  :disabled="deleting"
                  @click="emit('remove', profile.profile_id)"
                >
                  {{ t('common.delete') }}
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="profiles.length === 0">
            <td colspan="7" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.soraS3.empty') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SoraS3Profile } from '@/api/admin/settings'
import {
  formatDataManagementDate,
  formatStorageQuotaGB
} from '../dataManagementView'

defineProps<{
  profiles: SoraS3Profile[]
  loading: boolean
  activating: boolean
  deleting: boolean
}>()

const emit = defineEmits<{
  create: []
  reload: []
  edit: [profileID: string]
  activate: [profileID: string]
  remove: [profileID: string]
}>()

const { t } = useI18n()

const activeBadgeClass = 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
const inactiveBadgeClass = 'bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-gray-300'
</script>
