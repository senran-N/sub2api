<template>
  <div v-if="row.image_count > 0" class="flex items-center gap-1.5">
    <svg class="h-4 w-4 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
      />
    </svg>
    <span class="font-medium text-gray-900 dark:text-white">{{ row.image_count }}{{ t('usage.imageUnit') }}</span>
    <span class="text-gray-400">({{ row.image_size || '2K' }})</span>
  </div>

  <div v-else class="flex items-center gap-1.5">
    <div class="space-y-1.5 text-sm">
      <div class="flex items-center gap-2">
        <div class="inline-flex items-center gap-1">
          <Icon name="arrowDown" size="sm" class="text-emerald-500" />
          <span class="font-medium text-gray-900 dark:text-white">
            {{ row.input_tokens.toLocaleString() }}
          </span>
        </div>
        <div class="inline-flex items-center gap-1">
          <Icon name="arrowUp" size="sm" class="text-violet-500" />
          <span class="font-medium text-gray-900 dark:text-white">
            {{ row.output_tokens.toLocaleString() }}
          </span>
        </div>
      </div>

      <div
        v-if="row.cache_read_tokens > 0 || row.cache_creation_tokens > 0"
        class="flex items-center gap-2"
      >
        <div v-if="row.cache_read_tokens > 0" class="inline-flex items-center gap-1">
          <Icon name="inbox" size="sm" class="text-sky-500" />
          <span class="font-medium text-sky-600 dark:text-sky-400">
            {{ formatUserUsageCacheTokens(row.cache_read_tokens) }}
          </span>
        </div>

        <div v-if="row.cache_creation_tokens > 0" class="inline-flex items-center gap-1">
          <Icon name="edit" size="sm" class="text-amber-500" />
          <span class="font-medium text-amber-600 dark:text-amber-400">
            {{ formatUserUsageCacheTokens(row.cache_creation_tokens) }}
          </span>
          <span
            v-if="row.cache_creation_1h_tokens > 0"
            class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-orange-100 text-orange-600 ring-1 ring-inset ring-orange-200 dark:bg-orange-500/20 dark:text-orange-400 dark:ring-orange-500/30"
          >
            1h
          </span>
          <span
            v-if="row.cache_ttl_overridden"
            :title="t('usage.cacheTtlOverriddenHint')"
            class="inline-flex cursor-help items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-rose-100 text-rose-600 ring-1 ring-inset ring-rose-200 dark:bg-rose-500/20 dark:text-rose-400 dark:ring-rose-500/30"
          >
            R
          </span>
        </div>
      </div>
    </div>

    <button
      type="button"
      class="group relative"
      :aria-label="t('usage.tokenDetails')"
      @mouseenter="emit('show-details', $event, row)"
      @mouseleave="emit('hide-details')"
    >
      <div
        class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-blue-100 dark:bg-gray-700 dark:group-hover:bg-blue-900/50"
      >
        <Icon
          name="infoCircle"
          size="xs"
          class="text-gray-400 group-hover:text-blue-500 dark:text-gray-500 dark:group-hover:text-blue-400"
        />
      </div>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UsageLog } from '@/types'
import { formatUserUsageCacheTokens } from '../userUsageView'

defineProps<{
  row: UsageLog
}>()

const emit = defineEmits<{
  'show-details': [event: MouseEvent, row: UsageLog]
  'hide-details': []
}>()

const { t } = useI18n()
</script>
