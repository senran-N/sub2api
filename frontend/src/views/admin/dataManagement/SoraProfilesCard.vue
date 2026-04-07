<template>
  <div class="card sora-profiles-card">
    <div class="sora-profiles-card__header">
      <div>
        <h3 class="sora-profiles-card__title text-base font-semibold">
          {{ t('admin.settings.soraS3.title') }}
        </h3>
        <p class="sora-profiles-card__description mt-1 text-sm">
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
      <table class="sora-profiles-card__table">
        <thead>
          <tr class="sora-profiles-card__head-row">
            <th class="sora-profiles-card__head-cell">{{ t('admin.settings.soraS3.columns.profile') }}</th>
            <th class="sora-profiles-card__head-cell">{{ t('admin.settings.soraS3.columns.active') }}</th>
            <th class="sora-profiles-card__head-cell">{{ t('admin.settings.soraS3.columns.endpoint') }}</th>
            <th class="sora-profiles-card__head-cell">{{ t('admin.settings.soraS3.columns.bucket') }}</th>
            <th class="sora-profiles-card__head-cell">{{ t('admin.settings.soraS3.columns.quota') }}</th>
            <th class="sora-profiles-card__head-cell">{{ t('admin.settings.soraS3.columns.updatedAt') }}</th>
            <th class="sora-profiles-card__head-cell sora-profiles-card__head-cell--actions">{{ t('admin.settings.soraS3.columns.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="profile in profiles"
            :key="profile.profile_id"
            class="sora-profiles-card__row"
          >
            <td class="sora-profiles-card__cell">
              <div class="sora-profiles-card__id font-mono text-xs">{{ profile.profile_id }}</div>
              <div class="sora-profiles-card__description mt-1 text-xs">{{ profile.name }}</div>
            </td>
            <td class="sora-profiles-card__cell">
              <span
                class="sora-profiles-card__status-badge text-xs"
                :class="profile.is_active ? activeBadgeClass : inactiveBadgeClass"
              >
                {{ profile.is_active ? t('common.enabled') : t('common.disabled') }}
              </span>
            </td>
            <td class="sora-profiles-card__cell text-xs">
              <div>{{ profile.endpoint || '-' }}</div>
              <div class="sora-profiles-card__description mt-1">{{ profile.region || '-' }}</div>
            </td>
            <td class="sora-profiles-card__cell text-xs">{{ profile.bucket || '-' }}</td>
            <td class="sora-profiles-card__cell text-xs">{{ formatStorageQuotaGB(profile.default_storage_quota_bytes) }}</td>
            <td class="sora-profiles-card__cell text-xs">{{ formatDataManagementDate(profile.updated_at) }}</td>
            <td class="sora-profiles-card__cell sora-profiles-card__cell--actions text-xs">
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
            <td colspan="7" class="sora-profiles-card__empty sora-profiles-card__description text-center text-sm">
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
} from '../dataManagement/dataManagementHelpers'

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

const activeBadgeClass = 'badge badge-success'
const inactiveBadgeClass = 'badge badge-gray'
</script>

<style scoped>
.sora-profiles-card {
  padding: calc(var(--theme-table-mobile-card-padding) * 1.5);
}

.sora-profiles-card__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--theme-table-layout-gap);
  margin-bottom: var(--theme-table-layout-gap);
}

.sora-profiles-card__table {
  width: 100%;
  min-width: var(--theme-data-management-table-min-width);
  font-size: 0.875rem;
}

.sora-profiles-card__head-row {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  text-align: left;
  font-size: var(--theme-table-head-font-size);
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
}

.sora-profiles-card__head-cell,
.sora-profiles-card__cell {
  padding-top: var(--theme-table-cell-padding-y);
  padding-bottom: var(--theme-table-cell-padding-y);
  padding-right: var(--theme-table-cell-padding-x);
}

.sora-profiles-card__head-cell--actions,
.sora-profiles-card__cell--actions {
  padding-right: 0;
}

.sora-profiles-card__row {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  vertical-align: top;
}

.sora-profiles-card__status-badge {
  border-radius: var(--theme-button-radius);
  padding: 0.125rem 0.5rem;
}

.sora-profiles-card__empty {
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) 0;
}

.sora-profiles-card__title,
.sora-profiles-card__id,
.sora-profiles-card__row {
  color: var(--theme-page-text);
}

.sora-profiles-card__description,
.sora-profiles-card__head-row {
  color: var(--theme-page-muted);
}

</style>
