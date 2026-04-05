<template>
  <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="badgeClass">
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UsageLog } from '@/types'
import {
  getUsageRequestTypeBadgeClass,
  getUsageRequestTypeLabel
} from '@/utils/usagePresentation'

const props = defineProps<{
  log: UsageLog
}>()

const { t } = useI18n()

const label = computed(() => getUsageRequestTypeLabel(props.log, t))
const badgeClass = computed(() => getUsageRequestTypeBadgeClass(props.log))
</script>
