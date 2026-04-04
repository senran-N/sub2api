<template>
  <button
    v-if="accountCount > 0"
    type="button"
    class="inline-flex items-center rounded bg-gray-100 px-2 py-0.5 text-xs font-medium text-primary-700 hover:bg-gray-200 dark:bg-dark-600 dark:text-primary-300 dark:hover:bg-dark-500"
    @click="emit('accounts', proxy)"
  >
    {{ t('admin.groups.accountsCount', { count: accountCount }) }}
  </button>
  <span
    v-else
    class="inline-flex items-center rounded bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-800 dark:bg-dark-600 dark:text-gray-300"
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
