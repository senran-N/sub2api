<template>
  <div class="flex flex-col items-start gap-1">
    <span class="theme-text-muted text-sm">{{ formattedExpiresAt }}</span>
    <div v-if="isExpired || (account.auto_pause_on_expired && value)" class="flex items-center gap-1">
      <span
        v-if="isExpired"
        class="theme-chip theme-chip--compact theme-chip--warning"
      >
        {{ t('admin.accounts.expired') }}
      </span>
      <span
        v-if="account.auto_pause_on_expired && value"
        class="theme-chip theme-chip--compact theme-chip--success"
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
