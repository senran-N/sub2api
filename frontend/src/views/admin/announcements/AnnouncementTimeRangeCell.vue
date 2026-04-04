<template>
  <div class="text-sm text-gray-600 dark:text-gray-300">
    <div>
      <span class="font-medium">{{ t('admin.announcements.form.startsAt') }}:</span>
      <span class="ml-1">
        {{ startsAt ? formatDateTime(startsAt) : t('admin.announcements.timeImmediate') }}
      </span>
    </div>
    <div class="mt-0.5">
      <span class="font-medium">{{ t('admin.announcements.form.endsAt') }}:</span>
      <span class="ml-1">
        {{ endsAt ? formatDateTime(endsAt) : t('admin.announcements.timeNever') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatDateTime } from '@/utils/format'

defineProps<{
  startsAt?: string
  endsAt?: string
}>()

const { t } = useI18n()
</script>
