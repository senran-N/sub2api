<template>
  <div class="flex items-center gap-1.5">
    <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
    <Icon
      v-if="hasApiKeyIpRestrictions(row)"
      name="shield"
      size="sm"
      class="text-blue-500"
      :title="t('keys.ipRestrictionEnabled')"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { ApiKey } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import { hasApiKeyIpRestrictions } from './keysView'

defineProps<{
  value: string
  row: ApiKey
}>()

const { t } = useI18n()
</script>
