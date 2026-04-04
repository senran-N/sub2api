<template>
  <span :class="badgeClass">
    {{ t(`admin.redeem.types.${type}`) }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { RedeemCodeType } from '@/types'

const props = defineProps<{
  type: RedeemCodeType
}>()

const { t } = useI18n()

const badgeClass = computed(() => [
  'badge',
  props.type === 'balance'
    ? 'badge-success'
    : props.type === 'subscription'
      ? 'badge-warning'
      : 'badge-primary'
])
</script>
