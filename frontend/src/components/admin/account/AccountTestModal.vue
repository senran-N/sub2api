<template>
  <AccountTestModalBase
    :show="show"
    :account="account"
    :allow-custom-prompt="true"
    @close="emit('close')"
  />
</template>

<script setup lang="ts">
import AccountTestModalBase from '@/components/account/AccountTestModalBase.vue'
import type { Account } from '@/types'

defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()
</script>
