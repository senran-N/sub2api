<template>
  <div class="flex items-center gap-1.5 text-sm">
    <span class="theme-text-success font-medium">
      ${{ row.actual_cost.toFixed(6) }}
    </span>
    <button
      type="button"
      class="group relative"
      :aria-label="t('usage.costDetails')"
      @mouseenter="emit('show-details', $event, row)"
      @mouseleave="emit('hide-details')"
    >
      <div
        class="theme-info-trigger"
      >
        <Icon
          name="infoCircle"
          size="xs"
          class="theme-info-trigger__icon"
        />
      </div>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UsageLog } from '@/types'

defineProps<{
  row: UsageLog
}>()

const emit = defineEmits<{
  'show-details': [event: MouseEvent, row: UsageLog]
  'hide-details': []
}>()

const { t } = useI18n()
</script>
