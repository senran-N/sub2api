<template>
  <div v-if="hasGroupsData" class="flex flex-col gap-1">
    <span
      v-if="summary.exclusive.length > 0"
      class="group/ex relative inline-flex cursor-pointer items-center gap-1 whitespace-nowrap text-xs"
      @click.stop="emit('toggle-expanded', user.id)"
    >
      <Icon name="shield" size="xs" class="h-3.5 w-3.5 text-purple-500 dark:text-purple-400" />
      <span class="font-medium text-purple-600 dark:text-purple-400">{{ summary.exclusive.length }}</span>
      <span class="text-gray-500 dark:text-dark-400">{{ t('admin.users.exclusiveLabel') }}</span>
      <div
        v-if="!expanded"
        class="pointer-events-none absolute left-0 top-full z-50 mt-1.5 rounded bg-gray-900 px-2.5 py-1.5 text-xs text-white opacity-0 shadow-lg transition-opacity duration-75 group-hover/ex:opacity-100 dark:bg-dark-600"
      >
        <div class="absolute left-4 bottom-full border-4 border-transparent border-b-gray-900 dark:border-b-dark-600"></div>
        <div class="flex flex-col gap-0.5 whitespace-nowrap">
          <span v-for="group in summary.exclusive" :key="group.id">{{ group.name }}</span>
        </div>
      </div>
      <div
        v-if="expanded"
        class="absolute left-0 top-full z-50 mt-1.5 min-w-[160px] overflow-hidden rounded-lg border border-gray-200 bg-white py-1 text-xs shadow-xl dark:border-dark-600 dark:bg-dark-700"
      >
        <div class="border-b border-gray-100 px-3 py-1.5 text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:border-dark-600 dark:text-dark-400">
          {{ t('admin.users.clickToReplace') }}
        </div>
        <div
          v-for="group in summary.exclusive"
          :key="group.id"
          class="flex cursor-pointer items-center gap-2 px-3 py-2 text-gray-700 transition-colors hover:bg-primary-50 hover:text-primary-600 dark:text-dark-200 dark:hover:bg-primary-900/30 dark:hover:text-primary-400"
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
      <Icon name="globe" size="xs" class="h-3.5 w-3.5 text-gray-400 dark:text-dark-500" />
      <span class="font-medium text-gray-600 dark:text-dark-300">{{ summary.publicGroups.length }}</span>
      <span class="text-gray-400 dark:text-dark-500">{{ t('admin.users.publicLabel') }}</span>
      <div class="pointer-events-none absolute left-0 top-full z-50 mt-1.5 rounded bg-gray-900 px-2.5 py-1.5 text-xs text-white opacity-0 shadow-lg transition-opacity duration-75 group-hover/pub:opacity-100 dark:bg-dark-600">
        <div class="absolute left-4 bottom-full border-4 border-transparent border-b-gray-900 dark:border-b-dark-600"></div>
        <div class="flex flex-col gap-0.5 whitespace-nowrap">
          <span v-for="group in summary.publicGroups" :key="group.id">{{ group.name }}</span>
        </div>
      </div>
    </span>

    <span
      v-if="summary.exclusive.length === 0 && summary.publicGroups.length === 0"
      class="text-xs text-gray-400 dark:text-dark-500"
    >
      -
    </span>
  </div>
  <span v-else class="text-xs text-gray-400 dark:text-dark-500">-</span>
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
