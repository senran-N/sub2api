<template>
  <div class="flex items-center gap-1">
    <button
      type="button"
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20 dark:hover:text-green-400"
      @click="emit('use', row)"
    >
      <Icon name="terminal" size="sm" />
      <span class="text-xs">{{ t('keys.useKey') }}</span>
    </button>
    <button
      type="button"
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
      @click="emit('edit', row)"
    >
      <Icon name="edit" size="sm" />
      <span class="text-xs">{{ t('common.edit') }}</span>
    </button>
    <div class="relative" :ref="buttonRef">
      <button
        type="button"
        class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-300"
        @click.stop="emit('toggle-more', row.id)"
      >
        <Icon name="moreVertical" size="sm" />
        <span class="text-xs">{{ t('common.more') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ApiKey } from '@/types'

defineProps<{
  row: ApiKey
  buttonRef: (el: Element | import('vue').ComponentPublicInstance | null) => void
}>()

const emit = defineEmits<{
  use: [row: ApiKey]
  edit: [row: ApiKey]
  'toggle-more': [keyId: number]
}>()

const { t } = useI18n()
</script>
