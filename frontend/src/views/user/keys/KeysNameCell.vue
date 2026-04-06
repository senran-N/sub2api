<template>
  <div class="flex items-center gap-1.5">
    <span class="theme-text-strong font-medium">{{ value }}</span>
    <Icon
      v-if="hasApiKeyIpRestrictions(row)"
      name="shield"
      size="sm"
      class="keys-name-cell__shield"
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

<style scoped>
.keys-name-cell__shield {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}
</style>
