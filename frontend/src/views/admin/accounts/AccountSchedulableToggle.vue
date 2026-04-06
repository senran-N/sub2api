<template>
  <button
    class="account-schedulable-toggle relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent"
    :class="[
      account.schedulable
        ? 'account-schedulable-toggle--active'
        : 'account-schedulable-toggle--inactive'
    ]"
    :disabled="loading"
    :title="account.schedulable ? t('admin.accounts.schedulableEnabled') : t('admin.accounts.schedulableDisabled')"
    @click="emit('toggle', account)"
  >
    <span
      class="account-schedulable-toggle__thumb pointer-events-none inline-block h-4 w-4 transform rounded-full ring-0"
      :class="[account.schedulable ? 'translate-x-4' : 'translate-x-0']"
    />
  </button>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'

defineProps<{
  account: Account
  loading: boolean
}>()

const emit = defineEmits<{
  toggle: [account: Account]
}>()

const { t } = useI18n()
</script>

<style scoped>
.account-schedulable-toggle {
  transition: background-color 0.2s ease, box-shadow 0.2s ease;
}

.account-schedulable-toggle:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px color-mix(in srgb, var(--theme-surface) 100%, transparent),
    0 0 0 4px color-mix(in srgb, var(--theme-accent) 30%, transparent);
}

.account-schedulable-toggle:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.account-schedulable-toggle--active {
  background: var(--theme-accent);
}

.account-schedulable-toggle--active:hover {
  background: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}

.account-schedulable-toggle--inactive {
  background: color-mix(in srgb, var(--theme-page-muted) 28%, transparent);
}

.account-schedulable-toggle--inactive:hover {
  background: color-mix(in srgb, var(--theme-page-muted) 40%, transparent);
}

.account-schedulable-toggle__thumb {
  background: var(--theme-surface);
  box-shadow: 0 1px 2px color-mix(in srgb, var(--theme-page-text) 14%, transparent);
  transition: transform 0.2s ease;
}
</style>
