<template>
  <div v-if="hasGroupsData" class="flex flex-col gap-1">
    <span
      v-if="summary.exclusive.length > 0"
      class="group/ex relative inline-flex cursor-pointer items-center gap-1 whitespace-nowrap text-xs"
      @click.stop="emit('toggle-expanded', user.id)"
    >
      <Icon name="shield" size="xs" class="user-groups-cell__exclusive-icon h-3.5 w-3.5" />
      <span class="user-groups-cell__exclusive-count font-medium">{{ summary.exclusive.length }}</span>
      <span class="user-groups-cell__meta">{{ t('admin.users.exclusiveLabel') }}</span>
      <div
        v-if="!expanded"
        class="user-groups-cell__tooltip pointer-events-none absolute left-0 top-full z-50 mt-1.5 text-xs opacity-0 shadow-lg transition-opacity duration-75 group-hover/ex:opacity-100"
      >
        <div class="user-groups-cell__tooltip-arrow absolute left-4 bottom-full border-4 border-transparent"></div>
        <div class="flex flex-col gap-0.5 whitespace-nowrap">
          <span v-for="group in summary.exclusive" :key="group.id">{{ group.name }}</span>
        </div>
      </div>
      <div
        v-if="expanded"
        class="user-groups-cell__dropdown absolute left-0 top-full z-50 overflow-hidden text-xs shadow-xl"
      >
        <div class="user-groups-cell__dropdown-header text-[10px] font-medium uppercase tracking-wider">
          {{ t('admin.users.clickToReplace') }}
        </div>
        <div
          v-for="group in summary.exclusive"
          :key="group.id"
          class="user-groups-cell__dropdown-item flex cursor-pointer items-center gap-2 transition-colors"
          @click.stop="emit('replace-group', user, group)"
        >
          <Icon name="swap" size="xs" class="h-3.5 w-3.5 flex-shrink-0 opacity-50" />
          <span class="flex-1">{{ group.name }}</span>
        </div>
      </div>
    </span>

    <span
      v-if="summary.publicGroups.length > 0"
      class="group/pub relative inline-flex cursor-default items-center gap-1 whitespace-nowrap text-xs"
    >
      <Icon name="globe" size="xs" class="user-groups-cell__public-icon h-3.5 w-3.5" />
      <span class="user-groups-cell__public-count font-medium">{{ summary.publicGroups.length }}</span>
      <span class="user-groups-cell__subtle">{{ t('admin.users.publicLabel') }}</span>
      <div class="user-groups-cell__tooltip pointer-events-none absolute left-0 top-full z-50 mt-1.5 text-xs opacity-0 shadow-lg transition-opacity duration-75 group-hover/pub:opacity-100">
        <div class="user-groups-cell__tooltip-arrow absolute left-4 bottom-full border-4 border-transparent"></div>
        <div class="flex flex-col gap-0.5 whitespace-nowrap">
          <span v-for="group in summary.publicGroups" :key="group.id">{{ group.name }}</span>
        </div>
      </div>
    </span>

    <span
      v-if="summary.exclusive.length === 0 && summary.publicGroups.length === 0"
      class="user-groups-cell__subtle text-xs"
    >
      -
    </span>
  </div>
  <span v-else class="user-groups-cell__subtle text-xs">-</span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUser } from '@/types'

interface UserGroupSummaryItem {
  id: number
  name: string
}

defineProps<{
  user: AdminUser
  hasGroupsData: boolean
  expanded: boolean
  summary: {
    exclusive: UserGroupSummaryItem[]
    publicGroups: UserGroupSummaryItem[]
  }
}>()

const emit = defineEmits<{
  'toggle-expanded': [userId: number]
  'replace-group': [user: AdminUser, group: UserGroupSummaryItem]
}>()

const { t } = useI18n()
</script>

<style scoped>
.user-groups-cell__exclusive-icon,
.user-groups-cell__exclusive-count {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.user-groups-cell__public-icon,
.user-groups-cell__subtle,
.user-groups-cell__meta,
.user-groups-cell__dropdown-header {
  color: var(--theme-page-muted);
}

.user-groups-cell__public-count {
  color: var(--theme-page-text);
}

.user-groups-cell__tooltip {
  border-radius: var(--theme-tooltip-radius);
  background: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
  color: var(--theme-surface-contrast-text);
  padding: var(--theme-tooltip-padding);
}

.user-groups-cell__tooltip-arrow {
  border-bottom-color: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
}

.user-groups-cell__dropdown {
  margin-top: var(--theme-user-groups-dropdown-offset);
  min-width: var(--theme-user-groups-dropdown-min-width);
  border-radius: var(--theme-user-groups-dropdown-radius);
  padding-block: var(--theme-user-groups-dropdown-padding-y);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  background: var(--theme-surface);
}

.user-groups-cell__dropdown-header {
  padding: var(--theme-user-groups-dropdown-header-padding-y)
    var(--theme-user-groups-dropdown-header-padding-x);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.user-groups-cell__dropdown-item {
  padding: var(--theme-user-groups-dropdown-item-padding-y)
    var(--theme-user-groups-dropdown-item-padding-x);
  color: var(--theme-page-text);
}

.user-groups-cell__dropdown-item:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
  color: var(--theme-accent);
}
</style>
