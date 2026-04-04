<template>
  <span :class="badgeClass">
    {{
      notifyMode === 'popup'
        ? t('admin.announcements.notifyModeLabels.popup')
        : t('admin.announcements.notifyModeLabels.silent')
    }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Announcement } from '@/types'

const props = defineProps<{
  notifyMode: Announcement['notify_mode']
}>()

const { t } = useI18n()

const badgeClass = computed(() => [
  'badge',
  props.notifyMode === 'popup' ? 'badge-warning' : 'badge-gray'
])
</script>
