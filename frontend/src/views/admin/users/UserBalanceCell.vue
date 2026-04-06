<template>
  <div class="flex items-center gap-2">
    <div class="group relative">
      <button
        class="user-balance-cell__history font-medium underline decoration-dashed underline-offset-4"
        @click="emit('history', user)"
      >
        ${{ user.balance.toFixed(2) }}
      </button>
      <div class="user-balance-cell__tooltip pointer-events-none absolute bottom-full left-1/2 z-50 mb-1.5 -translate-x-1/2 whitespace-nowrap text-xs opacity-0 shadow-lg transition-opacity duration-75 group-hover:opacity-100">
        {{ t('admin.users.balanceHistoryTip') }}
        <div class="user-balance-cell__tooltip-arrow absolute left-1/2 top-full -translate-x-1/2"></div>
      </div>
    </div>
    <button
      class="theme-action-button theme-action-button--success user-balance-cell__deposit text-xs font-medium"
      :title="t('admin.users.deposit')"
      @click.stop="emit('deposit', user)"
    >
      {{ t('admin.users.deposit') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminUser } from '@/types'

defineProps<{
  user: AdminUser
}>()

const emit = defineEmits<{
  deposit: [user: AdminUser]
  history: [user: AdminUser]
}>()

const { t } = useI18n()
</script>

<style scoped>
.user-balance-cell__history {
  color: var(--theme-page-text);
  text-decoration-color: color-mix(in srgb, var(--theme-page-muted) 42%, transparent);
  transition: color 0.2s ease, text-decoration-color 0.2s ease;
}

.user-balance-cell__history:hover {
  color: var(--theme-accent);
  text-decoration-color: color-mix(in srgb, var(--theme-accent) 48%, transparent);
}

.user-balance-cell__tooltip {
  background: var(--theme-dropdown-bg);
  color: var(--theme-page-text);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  padding: var(--theme-tooltip-padding);
  border-radius: var(--theme-tooltip-radius);
}

.user-balance-cell__tooltip-arrow {
  border-width: var(--theme-tooltip-arrow-size);
  border-style: solid;
  border-color: transparent;
  border-top-color: color-mix(in srgb, var(--theme-dropdown-bg) 100%, transparent);
}

.user-balance-cell__deposit {
  border-radius: var(--theme-button-radius);
  padding:
    var(--theme-account-usage-action-padding-y)
    var(--theme-account-usage-action-padding-x);
}
</style>
