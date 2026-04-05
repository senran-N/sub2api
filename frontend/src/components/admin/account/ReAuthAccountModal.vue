<template>
  <SharedReAuthAccountModal
    :show="show"
    :account="account"
    @close="emit('close')"
    @reauthorized="handleReauthorized"
  />
</template>

<script setup lang="ts">
import type { Account } from '@/types'
import SharedReAuthAccountModal from '@/components/account/ReAuthAccountModal.vue'

interface Props {
  show: boolean
  account: Account | null
}

const { show, account } = defineProps<Props>()
const emit = defineEmits<{
  close: []
  reauthorized: [account: Account]
}>()

const handleReauthorized = (updatedAccount?: Account) => {
  if (!updatedAccount) {
    return
  }
  emit('reauthorized', updatedAccount)
}
</script>
