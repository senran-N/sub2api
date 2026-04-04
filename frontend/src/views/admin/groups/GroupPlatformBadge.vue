<template>
  <span
    :class="[
      'inline-flex items-center rounded-full text-xs font-medium',
      badgeClass,
      getGroupPlatformBadgeClass(platform)
    ]"
  >
    <PlatformIcon v-if="showIcon" :platform="platform" size="xs" />
    {{ t(`admin.groups.platforms.${platform}`) }}
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { GroupPlatform } from '@/types'
import { getGroupPlatformBadgeClass } from '../groupsTable'

withDefaults(
  defineProps<{
    platform: GroupPlatform
    showIcon?: boolean
    badgeClass?: string
  }>(),
  {
    showIcon: true,
    badgeClass: 'gap-1.5 px-2.5 py-0.5'
  }
)

const { t } = useI18n()
</script>
