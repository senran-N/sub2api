<template>
  <div class="flex items-center space-x-1">
    <button
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20 dark:hover:text-green-400"
      :title="t('admin.promo.copyRegisterLink')"
      @click="emit('copy-link')"
    >
      <Icon name="link" size="sm" />
    </button>
    <button
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400"
      :title="t('admin.promo.viewUsages')"
      @click="emit('view-usages')"
    >
      <Icon name="eye" size="sm" />
    </button>
    <button
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
      :title="t('common.edit')"
      @click="emit('edit')"
    >
      <Icon name="edit" size="sm" />
    </button>
    <button
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
      :title="t('common.delete')"
      @click="emit('delete')"
    >
      <Icon name="trash" size="sm" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const emit = defineEmits<{
  'copy-link': []
  'view-usages': []
  edit: []
  delete: []
}>()

const { t } = useI18n()
</script>
