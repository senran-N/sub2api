<template>
  <div class="flex flex-wrap items-start justify-between gap-4">
    <div class="flex flex-1 flex-wrap items-center gap-3">
      <SubscriptionFilterUserSearch
        :keyword="filterUserKeyword"
        :results="filterUserResults"
        :loading="filterUserLoading"
        :show-dropdown="showFilterUserDropdown"
        :selected-user="selectedFilterUser"
        @update:keyword="emit('update:filterUserKeyword', $event)"
        @search="emit('search-filter-users')"
        @focus="emit('show-filter-user-dropdown')"
        @select-user="emit('select-filter-user', $event)"
        @clear-user="emit('clear-filter-user')"
      />
      <div class="w-full sm:w-40">
        <Select
          :model-value="status"
          :options="statusOptions"
          :placeholder="t('admin.subscriptions.allStatus')"
          @update:model-value="handleStatusUpdate"
          @change="emit('apply-filters')"
        />
      </div>
      <div class="w-full sm:w-48">
        <Select
          :model-value="groupId"
          :options="groupOptions"
          :placeholder="t('admin.subscriptions.allGroups')"
          @update:model-value="emit('update:groupId', normalizeSelectValue($event))"
          @change="emit('apply-filters')"
        />
      </div>
      <div class="w-full sm:w-40">
        <Select
          :model-value="platform"
          :options="platformFilterOptions"
          :placeholder="t('admin.subscriptions.allPlatforms')"
          @update:model-value="emit('update:platform', normalizeSelectValue($event))"
          @change="emit('apply-filters')"
        />
      </div>
    </div>

    <SubscriptionToolbarActions
      :loading="loading"
      :user-column-mode="userColumnMode"
      :toggleable-columns="toggleableColumns"
      :is-column-visible="isColumnVisible"
      @refresh="emit('refresh')"
      @set-user-mode="emit('set-user-mode', $event)"
      @toggle-column="emit('toggle-column', $event)"
      @guide="emit('guide')"
      @assign="emit('assign')"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SimpleUser } from '@/api/admin/usage'
import Select from '@/components/common/Select.vue'
import type { Column } from '@/components/common/types'
import type { SubscriptionStatusFilter } from './subscriptionForm'
import SubscriptionFilterUserSearch from './SubscriptionFilterUserSearch.vue'
import SubscriptionToolbarActions from './SubscriptionToolbarActions.vue'

defineProps<{
  filterUserKeyword: string
  filterUserResults: SimpleUser[]
  filterUserLoading: boolean
  showFilterUserDropdown: boolean
  selectedFilterUser: SimpleUser | null
  status: SubscriptionStatusFilter
  groupId: string
  platform: string
  statusOptions: Array<{ value: string; label: string }>
  groupOptions: Array<{ value: string; label: string }>
  platformFilterOptions: Array<{ value: string; label: string }>
  loading: boolean
  userColumnMode: 'email' | 'username'
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'update:filterUserKeyword': [value: string]
  'search-filter-users': []
  'show-filter-user-dropdown': []
  'select-filter-user': [user: SimpleUser]
  'clear-filter-user': []
  'update:status': [value: SubscriptionStatusFilter]
  'update:groupId': [value: string]
  'update:platform': [value: string]
  'apply-filters': []
  refresh: []
  'set-user-mode': [mode: 'email' | 'username']
  'toggle-column': [key: string]
  guide: []
  assign: []
}>()

const { t } = useI18n()

const normalizeSelectValue = (value: string | number | boolean | null) =>
  typeof value === 'string' ? value : ''

const handleStatusUpdate = (value: string | number | boolean | null) => {
  if (value === '' || value === 'active' || value === 'expired' || value === 'revoked') {
    emit('update:status', value)
  }
}
</script>
