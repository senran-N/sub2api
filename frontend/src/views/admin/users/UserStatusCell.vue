<template>
  <div class="flex items-center gap-1.5">
    <span
      :class="[
        'inline-block h-2 w-2 rounded-full',
        status === 'active' ? 'bg-green-500' : 'bg-red-500'
      ]"
    ></span>
    <span class="text-sm text-gray-700 dark:text-gray-300">
      {{ status === 'active' ? t('common.active') : t('admin.users.disabled') }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  status: 'active' | 'disabled'
}>()

const { t } = useI18n()
</script>
