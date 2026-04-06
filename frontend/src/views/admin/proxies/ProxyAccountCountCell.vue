<template>
  <button
    v-if="accountCount > 0"
    type="button"
    class="proxy-account-count-cell__button theme-chip theme-chip--compact theme-chip--accent inline-flex items-center"
    @click="emit('accounts', proxy)"
  >
    {{ t('admin.groups.accountsCount', { count: accountCount }) }}
  </button>
  <span
    v-else
    class="theme-chip theme-chip--compact theme-chip--neutral inline-flex items-center"
  >
    {{ t('admin.groups.accountsCount', { count: 0 }) }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Proxy } from '@/types'

const props = defineProps<{
  proxy: Proxy
}>()

const emit = defineEmits<{
  accounts: [proxy: Proxy]
}>()

const { t } = useI18n()

const accountCount = computed(() => props.proxy.account_count || 0)
</script>

<style scoped>
.proxy-account-count-cell__button {
  transition: background-color 0.2s ease, color 0.2s ease;
}

.proxy-account-count-cell__button:hover {
  --theme-chip-bg: color-mix(in srgb, var(--theme-accent-soft) 94%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, var(--theme-accent) 92%, var(--theme-page-text));
}
</style>
