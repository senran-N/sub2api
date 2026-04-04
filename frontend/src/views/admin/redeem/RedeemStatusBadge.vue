<template>
  <span :class="badgeClass">
    {{ t(`admin.redeem.status.${status}`) }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { RedeemCode } from '@/types'

const props = defineProps<{
  status: RedeemCode['status']
}>()

const { t } = useI18n()

const badgeClass = computed(() => [
  'badge',
  props.status === 'unused'
    ? 'badge-success'
    : props.status === 'used'
      ? 'badge-gray'
      : 'badge-danger'
])
</script>
