<template>
  <div class="flex flex-col">
    <span class="theme-text-strong font-medium">{{ account.name }}</span>
    <span
      v-if="emailAddress"
      class="theme-text-muted account-name-cell__email truncate text-xs"
      :title="emailAddress"
    >
      {{ emailAddress }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Account } from '@/types'

const props = defineProps<{
  account: Account
}>()

const emailAddress = computed(() => {
  const value = props.account.extra?.email_address
  return typeof value === 'string' ? value : null
})
</script>

<style scoped>
.account-name-cell__email {
  max-width: var(--theme-account-name-secondary-max-width);
}
</style>
