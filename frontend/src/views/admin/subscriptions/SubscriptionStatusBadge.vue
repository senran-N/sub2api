<template>
  <span :class="badgeClass">
    {{ t(`admin.subscriptions.status.${status}`) }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  status: 'active' | 'expired' | 'revoked'
}>()

const { t } = useI18n()

const badgeClass = computed(() => [
  'badge',
  props.status === 'active'
    ? 'badge-success'
    : props.status === 'expired'
      ? 'badge-warning'
      : 'badge-danger'
])
</script>
