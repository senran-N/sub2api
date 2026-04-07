<template>
  <span
    :class="[
      'group-platform-badge theme-chip theme-chip--regular inline-flex items-center text-xs',
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
import { getGroupPlatformBadgeClass } from './groupDisplay'

withDefaults(
  defineProps<{
    platform: GroupPlatform
    showIcon?: boolean
    badgeClass?: string
  }>(),
  {
    showIcon: true,
    badgeClass: 'gap-1.5'
  }
)

const { t } = useI18n()
</script>

<style scoped>
.group-platform-badge {
  gap: 0.375rem;
  padding: var(--theme-group-selector-padding);
  border-radius: var(--theme-button-radius);
}
</style>
