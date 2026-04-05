<template>
  <Teleport to="body">
    <div v-if="show && position && row">
      <div class="fixed inset-0 z-[9998]" @click="emit('close')"></div>
      <div
        class="fixed z-[9999] w-48 overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5 dark:bg-dark-800 dark:ring-white/10"
        :style="{ top: `${position.top}px`, left: `${position.left}px` }"
      >
        <div class="py-1">
          <button
            v-if="!hideCcsImportButton"
            class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
            @click="handleImport"
          >
            <Icon name="upload" size="sm" class="text-blue-500" />
            {{ t('keys.importToCcSwitch') }}
          </button>
          <button
            class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
            @click="handleToggleStatus"
          >
            <Icon
              :name="row.status === 'active' ? 'ban' : 'checkCircle'"
              size="sm"
              :class="row.status === 'active' ? 'text-yellow-500' : 'text-green-500'"
            />
            {{ row.status === 'active' ? t('keys.disable') : t('keys.enable') }}
          </button>
          <div class="my-1 border-t border-gray-100 dark:border-dark-700"></div>
          <button
            class="flex w-full items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
            @click="handleDelete"
          >
            <Icon name="trash" size="sm" />
            {{ t('common.delete') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ApiKey } from '@/types'

defineProps<{
  show: boolean
  position: { top: number; left: number } | null
  row: ApiKey | null
  hideCcsImportButton: boolean
}>()

const emit = defineEmits<{
  close: []
  import: []
  'toggle-status': []
  delete: []
}>()

const { t } = useI18n()

function handleImport() {
  emit('import')
  emit('close')
}

function handleToggleStatus() {
  emit('toggle-status')
  emit('close')
}

function handleDelete() {
  emit('delete')
  emit('close')
}
</script>
