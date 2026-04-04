<template>
  <div class="flex flex-col items-start gap-1">
    <span class="text-sm text-gray-500 dark:text-dark-400">{{ formattedExpiresAt }}</span>
    <div v-if="isExpired || (account.auto_pause_on_expired && value)" class="flex items-center gap-1">
      <span
        v-if="isExpired"
        class="inline-flex items-center rounded-md bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
      >
        {{ t('admin.accounts.expired') }}
      </span>
      <span
        v-if="account.auto_pause_on_expired && value"
        class="inline-flex items-center rounded-md bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300"
      >
        {{ t('admin.accounts.autoPauseOnExpired') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'
import { formatAccountExpiresAt, isAccountExpired } from '../accountsView'

const props = defineProps<{
  account: Account
  value: number | null
}>()

const { t } = useI18n()

const formattedExpiresAt = computed(() => formatAccountExpiresAt(props.value))
const isExpired = computed(() => isAccountExpired(props.value))
</script>
