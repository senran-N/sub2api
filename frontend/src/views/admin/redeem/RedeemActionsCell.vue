<template>
  <div class="flex items-center space-x-2">
    <button
      v-if="showDelete"
      class="theme-action-button theme-action-button--danger"
      @click="emit('delete')"
    >
      <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
        />
      </svg>
      <span class="text-xs">{{ t('common.delete') }}</span>
    </button>
    <span v-else class="theme-text-subtle">-</span>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  showDelete: boolean
}>()

const emit = defineEmits<{
  delete: []
}>()

const { t } = useI18n()
</script>
