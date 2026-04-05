<template>
  <span :class="['badge', getApiKeyStatusBadgeClass(status)]">
    {{ t(`keys.status.${status}`) }}
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { ApiKey } from '@/types'
import { getApiKeyStatusBadgeClass } from './keysView'

defineProps<{
  status: ApiKey['status']
}>()

const { t } = useI18n()
</script>
