<template>
  <span class="theme-text-strong text-sm font-medium">
    <template v-if="code.type === 'balance'">${{ code.value.toFixed(2) }}</template>
    <template v-else-if="code.type === 'subscription'">
      {{ code.validity_days || 30 }} {{ t('admin.redeem.days') }}
      <span v-if="code.group" class="theme-text-muted ml-1 text-xs">
        ({{ code.group.name }})
      </span>
    </template>
    <template v-else>{{ code.value }}</template>
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { RedeemCode } from '@/types'

defineProps<{
  code: RedeemCode
}>()

const { t } = useI18n()
</script>
