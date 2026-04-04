<template>
  <div class="flex flex-col">
    <span class="font-medium text-gray-900 dark:text-white">{{ account.name }}</span>
    <span
      v-if="emailAddress"
      class="max-w-[200px] truncate text-xs text-gray-500 dark:text-gray-400"
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
