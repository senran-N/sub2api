<template>
  <div class="flex flex-col gap-3">
    <div class="flex flex-wrap items-center gap-3">
      <SearchInput
        :model-value="search"
        :placeholder="t('keys.searchPlaceholder')"
        class="w-full sm:w-64"
        @update:model-value="emit('update:search', $event)"
        @search="emit('apply')"
      />
      <Select
        :model-value="groupId"
        class="w-40"
        :options="groupOptions"
        @update:model-value="emit('update:groupId', $event)"
      />
      <Select
        :model-value="status"
        class="w-40"
        :options="statusOptions"
        @update:model-value="emit('update:status', $event)"
      />
    </div>
    <EndpointPopover
      v-if="publicSettings?.api_base_url || (publicSettings?.custom_endpoints?.length ?? 0) > 0"
      :api-base-url="publicSettings?.api_base_url || ''"
      :custom-endpoints="publicSettings?.custom_endpoints || []"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import SearchInput from '@/components/common/SearchInput.vue'
import Select from '@/components/common/Select.vue'
import EndpointPopover from '@/components/keys/EndpointPopover.vue'
import type { PublicSettings } from '@/types'

defineProps<{
  search: string
  groupId: string | number
  status: string
  groupOptions: Array<{ value: string | number; label: string }>
  statusOptions: Array<{ value: string; label: string }>
  publicSettings: PublicSettings | null
}>()

const emit = defineEmits<{
  'update:search': [value: string]
  'update:groupId': [value: string | number | boolean | null]
  'update:status': [value: string | number | boolean | null]
  apply: []
}>()

const { t } = useI18n()
</script>
