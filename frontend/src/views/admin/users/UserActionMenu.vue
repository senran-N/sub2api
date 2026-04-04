<template>
  <Teleport to="body">
    <div
      v-if="user && position"
      class="action-menu-content fixed z-[9999] w-48 overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5 dark:bg-dark-800 dark:ring-white/10"
      :style="{ top: `${position.top}px`, left: `${position.left}px` }"
    >
      <div class="py-1">
        <button
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="emitAndClose('api-keys', user)"
        >
          <Icon name="key" size="sm" class="text-gray-400" :stroke-width="2" />
          {{ t('admin.users.apiKeys') }}
        </button>

        <button
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="emitAndClose('groups', user)"
        >
          <Icon name="users" size="sm" class="text-gray-400" :stroke-width="2" />
          {{ t('admin.users.groups') }}
        </button>

        <div class="my-1 border-t border-gray-100 dark:border-dark-700"></div>

        <button
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="emitAndClose('deposit', user)"
        >
          <Icon name="plus" size="sm" class="text-emerald-500" :stroke-width="2" />
          {{ t('admin.users.deposit') }}
        </button>

        <button
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="emitAndClose('withdraw', user)"
        >
          <svg class="h-4 w-4 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
          </svg>
          {{ t('admin.users.withdraw') }}
        </button>

        <button
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="emitAndClose('history', user)"
        >
          <Icon name="dollar" size="sm" class="text-gray-400" :stroke-width="2" />
          {{ t('admin.users.balanceHistory') }}
        </button>

        <div class="my-1 border-t border-gray-100 dark:border-dark-700"></div>

        <button
          v-if="user.role !== 'admin'"
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
          @click="emitAndClose('delete', user)"
        >
          <Icon name="trash" size="sm" :stroke-width="2" />
          {{ t('common.delete') }}
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUser } from '@/types'

defineProps<{
  user: AdminUser | null
  position: { top: number; left: number } | null
}>()

const emit = defineEmits<{
  close: []
  'api-keys': [user: AdminUser]
  groups: [user: AdminUser]
  deposit: [user: AdminUser]
  withdraw: [user: AdminUser]
  history: [user: AdminUser]
  delete: [user: AdminUser]
}>()

const { t } = useI18n()

function emitAndClose(
  event: 'api-keys' | 'groups' | 'deposit' | 'withdraw' | 'history' | 'delete',
  user: AdminUser
) {
  if (event === 'api-keys') {
    emit('api-keys', user)
  } else if (event === 'groups') {
    emit('groups', user)
  } else if (event === 'deposit') {
    emit('deposit', user)
  } else if (event === 'withdraw') {
    emit('withdraw', user)
  } else if (event === 'history') {
    emit('history', user)
  } else {
    emit('delete', user)
  }

  emit('close')
}
</script>
