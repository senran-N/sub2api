<template>
  <span class="theme-text-default text-sm">
    {{ getAnnouncementTargetingSummary(targeting, t) }}
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Announcement } from '@/types'
import { getAnnouncementTargetingSummary } from './announcementsForm'

defineProps<{
  targeting: Announcement['targeting']
}>()

const { t } = useI18n()
</script>
