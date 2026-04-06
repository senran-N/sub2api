<template>
  <div class="text-sm">
    <div class="flex items-center gap-1.5">
      <span class="theme-text-muted">{{ t('admin.users.today') }}:</span>
      <span class="font-medium theme-text-default">
        ${{ todayCost.toFixed(4) }}
      </span>
    </div>
    <div class="mt-0.5 flex items-center gap-1.5">
      <span class="theme-text-muted">{{ t('admin.users.total') }}:</span>
      <span class="font-medium theme-text-default">
        ${{ totalCost.toFixed(4) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  usage?: {
    today_actual_cost?: number
    total_actual_cost?: number
  }
}>()

const { t } = useI18n()

const todayCost = computed(() => props.usage?.today_actual_cost ?? 0)
const totalCost = computed(() => props.usage?.total_actual_cost ?? 0)
</script>
