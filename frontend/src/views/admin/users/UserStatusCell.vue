<template>
  <div class="flex items-center gap-1.5">
    <span
      :class="[
        'user-status-cell__dot inline-block h-2 w-2 rounded-full',
        status === 'active' ? 'user-status-cell__dot--active' : 'user-status-cell__dot--disabled'
      ]"
    ></span>
    <span class="theme-text-default text-sm">
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

<style scoped>
.user-status-cell__dot--active {
  background: rgb(var(--theme-success-rgb));
}

.user-status-cell__dot--disabled {
  background: rgb(var(--theme-danger-rgb));
}
</style>
