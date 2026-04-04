<template>
  <div class="flex items-center gap-1">
    <button
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
      @click="emit('edit', user)"
    >
      <Icon name="edit" size="sm" />
      <span class="text-xs">{{ t('common.edit') }}</span>
    </button>

    <button
      v-if="user.role !== 'admin'"
      :class="[
        'flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors',
        user.status === 'active'
          ? 'hover:bg-orange-50 hover:text-orange-600 dark:hover:bg-orange-900/20 dark:hover:text-orange-400'
          : 'hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20 dark:hover:text-green-400'
      ]"
      @click="emit('toggle-status', user)"
    >
      <Icon v-if="user.status === 'active'" name="ban" size="sm" />
      <Icon v-else name="checkCircle" size="sm" />
      <span class="text-xs">
        {{ user.status === 'active' ? t('admin.users.disable') : t('admin.users.enable') }}
      </span>
    </button>

    <button
      class="action-menu-trigger flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:hover:bg-dark-700 dark:hover:text-white"
      :class="{ 'bg-gray-100 text-gray-900 dark:bg-dark-700 dark:text-white': menuOpen }"
      @click="emit('open-menu', user, $event)"
    >
      <Icon name="more" size="sm" />
      <span class="text-xs">{{ t('common.more') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUser } from '@/types'

defineProps<{
  user: AdminUser
  menuOpen: boolean
}>()

const emit = defineEmits<{
  edit: [user: AdminUser]
  'toggle-status': [user: AdminUser]
  'open-menu': [user: AdminUser, event: MouseEvent]
}>()

const { t } = useI18n()
</script>
