<template>
  <div v-if="row.image_count > 0" class="flex items-center gap-1.5">
    <svg class="user-usage-token-cell__image-icon h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
      />
    </svg>
    <span class="user-usage-token-cell__value font-medium">{{ row.image_count }}{{ t('usage.imageUnit') }}</span>
    <span class="user-usage-token-cell__subtle">({{ row.image_size || '2K' }})</span>
  </div>

  <div v-else class="flex items-center gap-1.5">
    <div class="space-y-1.5 text-sm">
      <div class="flex items-center gap-2">
        <div class="inline-flex items-center gap-1">
          <Icon name="arrowDown" size="sm" class="user-usage-token-cell__icon user-usage-token-cell__icon--success" />
          <span class="user-usage-token-cell__value font-medium">
            {{ row.input_tokens.toLocaleString() }}
          </span>
        </div>
        <div class="inline-flex items-center gap-1">
          <Icon name="arrowUp" size="sm" class="user-usage-token-cell__icon user-usage-token-cell__icon--brand" />
          <span class="user-usage-token-cell__value font-medium">
            {{ row.output_tokens.toLocaleString() }}
          </span>
        </div>
      </div>

      <div
        v-if="row.cache_read_tokens > 0 || row.cache_creation_tokens > 0"
        class="flex items-center gap-2"
      >
        <div v-if="row.cache_read_tokens > 0" class="inline-flex items-center gap-1">
          <Icon name="inbox" size="sm" class="user-usage-token-cell__icon user-usage-token-cell__icon--info" />
          <span class="user-usage-token-cell__value user-usage-token-cell__value--info font-medium">
            {{ formatUserUsageCacheTokens(row.cache_read_tokens) }}
          </span>
        </div>

        <div v-if="row.cache_creation_tokens > 0" class="inline-flex items-center gap-1">
          <Icon name="edit" size="sm" class="user-usage-token-cell__icon user-usage-token-cell__icon--warning" />
          <span class="user-usage-token-cell__value user-usage-token-cell__value--warning font-medium">
            {{ formatUserUsageCacheTokens(row.cache_creation_tokens) }}
          </span>
          <span
            v-if="row.cache_creation_1h_tokens > 0"
            class="user-usage-token-cell__badge user-usage-token-cell__badge--brand inline-flex items-center text-[10px] font-medium leading-tight"
          >
            1h
          </span>
          <span
            v-if="row.cache_ttl_overridden"
            :title="t('usage.cacheTtlOverriddenHint')"
            class="user-usage-token-cell__badge user-usage-token-cell__badge--danger inline-flex cursor-help items-center text-[10px] font-medium leading-tight"
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
        class="user-usage-token-cell__info-shell flex h-4 w-4 cursor-help items-center justify-center rounded-full transition-colors"
      >
        <Icon
          name="infoCircle"
          size="xs"
          class="user-usage-token-cell__info-icon"
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

<style scoped>
.user-usage-token-cell__value {
  color: var(--theme-page-text);
}

.user-usage-token-cell__subtle {
  color: var(--theme-page-muted);
}

.user-usage-token-cell__image-icon {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.user-usage-token-cell__icon--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.user-usage-token-cell__icon--brand {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.user-usage-token-cell__icon--info,
.user-usage-token-cell__value--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.user-usage-token-cell__icon--warning,
.user-usage-token-cell__value--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.user-usage-token-cell__badge {
  border-radius: var(--theme-button-radius);
  padding: 0 var(--theme-usage-progress-chip-padding-x);
  border: 1px solid transparent;
}

.user-usage-token-cell__badge--brand {
  background: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 84%, var(--theme-page-text));
  border-color: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 20%, var(--theme-card-border));
}

.user-usage-token-cell__badge--danger {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 20%, var(--theme-card-border));
}

.user-usage-token-cell__info-shell {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.group:hover .user-usage-token-cell__info-shell {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.user-usage-token-cell__info-icon {
  color: var(--theme-page-muted);
}

.group:hover .user-usage-token-cell__info-icon {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}
</style>
