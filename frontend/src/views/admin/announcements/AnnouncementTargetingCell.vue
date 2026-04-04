<template>
  <span class="text-sm text-gray-600 dark:text-gray-300">
    {{ getAnnouncementTargetingSummary(targeting, t) }}
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Announcement } from '@/types'
import { getAnnouncementTargetingSummary } from '../announcementsForm'

defineProps<{
  targeting: Announcement['targeting']
}>()

const { t } = useI18n()
</script>
