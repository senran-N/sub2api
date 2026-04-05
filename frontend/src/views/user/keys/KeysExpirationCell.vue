<template>
  <span :class="getApiKeyExpirationTextClass(value)">
    {{ value ? formatDateTime(value) : t('keys.noExpiration') }}
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatDateTime } from '@/utils/format'
import { getApiKeyExpirationTextClass } from './keysView'

defineProps<{
  value: string | null | undefined
}>()

const { t } = useI18n()
</script>
