<template>
  <span :class="badgeClass">
    {{ getAnnouncementStatusLabel(status, t) }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Announcement } from '@/types'
import { getAnnouncementStatusLabel } from './announcementsForm'

const props = defineProps<{
  status: Announcement['status']
}>()

const { t } = useI18n()

const badgeClass = computed(() => [
  'badge',
  props.status === 'active'
    ? 'badge-success'
    : props.status === 'draft'
      ? 'badge-gray'
      : 'badge-warning'
])
</script>
