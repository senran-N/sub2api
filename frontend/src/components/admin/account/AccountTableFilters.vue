<template>
  <div class="flex flex-wrap items-center gap-3">
    <SearchInput
      :model-value="searchQuery"
      :placeholder="t('admin.accounts.searchAccounts')"
      class="w-full sm:w-64"
      @update:model-value="$emit('update:searchQuery', $event)"
      @search="$emit('change')"
    />
    <Select
      :model-value="filters.platform"
      class="w-40"
      :options="platformOptions"
      @update:model-value="updateFilter('platform', $event)"
      @change="$emit('change')"
    />
    <Select
      :model-value="filters.type"
      class="w-40"
      :options="typeOptions"
      @update:model-value="updateFilter('type', $event)"
      @change="$emit('change')"
    />
    <Select
      :model-value="filters.status"
      class="w-40"
      :options="statusOptions"
      @update:model-value="updateFilter('status', $event)"
      @change="$emit('change')"
    />
    <Select
      :model-value="filters.privacy_mode"
      class="w-40"
      :options="privacyOptions"
      @update:model-value="updateFilter('privacy_mode', $event)"
      @change="$emit('change')"
    />
    <Select
      :model-value="filters.group"
      class="w-40"
      :options="groupOptions"
      @update:model-value="updateFilter('group', $event)"
      @change="$emit('change')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import type { AdminGroup } from '@/types'
import type { AccountListQuery } from '@/views/admin/accounts/accountsList'

type AccountFilterKey = Exclude<keyof AccountListQuery, 'search'>

const props = withDefaults(
  defineProps<{
    searchQuery: string
    filters: AccountListQuery
    groups?: AdminGroup[]
  }>(),
  {
    groups: () => []
  }
)

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:filters': [value: AccountListQuery]
  change: []
}>()

const { t } = useI18n()

function updateFilter(key: AccountFilterKey, value: string | number | boolean | null) {
  emit('update:filters', {
    ...props.filters,
    [key]: typeof value === 'string' ? value : String(value ?? '')
  })
}

const platformOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allPlatforms') },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'antigravity', label: 'Antigravity' },
  { value: 'sora', label: 'Sora' }
])

const typeOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allTypes') },
  { value: 'oauth', label: t('admin.accounts.oauthType') },
  { value: 'setup-token', label: t('admin.accounts.setupToken') },
  { value: 'apikey', label: t('admin.accounts.apiKey') },
  { value: 'bedrock', label: 'AWS Bedrock' }
])

const statusOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allStatus') },
  { value: 'active', label: t('admin.accounts.status.active') },
  { value: 'inactive', label: t('admin.accounts.status.inactive') },
  { value: 'error', label: t('admin.accounts.status.error') },
  { value: 'rate_limited', label: t('admin.accounts.status.rateLimited') },
  { value: 'temp_unschedulable', label: t('admin.accounts.status.tempUnschedulable') }
])

const privacyOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allPrivacyModes') },
  { value: '__unset__', label: t('admin.accounts.privacyUnset') },
  { value: 'training_off', label: 'Privacy' },
  { value: 'training_set_cf_blocked', label: 'CF' },
  { value: 'training_set_failed', label: 'Fail' }
])

const groupOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allGroups') },
  { value: 'ungrouped', label: t('admin.accounts.ungroupedGroup') },
  ...props.groups.map((group) => ({
    value: String(group.id),
    label: group.name
  }))
])
</script>
