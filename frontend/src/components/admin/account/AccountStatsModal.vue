<template>
  <SharedAccountStatsModal
    :show="show"
    :account="account"
    @close="emit('close')"
  />
</template>

<script setup lang="ts">
import type { Account } from '@/types'
import SharedAccountStatsModal from '@/components/account/AccountStatsModal.vue'

interface Props {
  show: boolean
  account: Account | null
}

const { show, account } = defineProps<Props>()
const emit = defineEmits<{
  close: []
}>()
</script>
