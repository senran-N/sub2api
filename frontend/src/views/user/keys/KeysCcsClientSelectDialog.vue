<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="$emit('close')">
    <div class="space-y-4">
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ description }}
      </p>

      <div class="grid grid-cols-2 gap-3">
        <button
          @click="$emit('select', 'claude')"
          class="rounded-xl border-2 border-gray-200 p-4 transition-all hover:border-primary-500 hover:bg-primary-50 dark:border-dark-600 dark:hover:border-primary-500 dark:hover:bg-primary-900/20"
        >
          <div class="flex flex-col items-center gap-2">
            <Icon name="terminal" size="xl" class="text-gray-600 dark:text-gray-400" />
            <span class="font-medium text-gray-900 dark:text-white">{{ claudeLabel }}</span>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ claudeDescription }}</span>
          </div>
        </button>

        <button
          @click="$emit('select', 'gemini')"
          class="rounded-xl border-2 border-gray-200 p-4 transition-all hover:border-primary-500 hover:bg-primary-50 dark:border-dark-600 dark:hover:border-primary-500 dark:hover:bg-primary-900/20"
        >
          <div class="flex flex-col items-center gap-2">
            <Icon name="sparkles" size="xl" class="text-gray-600 dark:text-gray-400" />
            <span class="font-medium text-gray-900 dark:text-white">{{ geminiLabel }}</span>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ geminiDescription }}</span>
          </div>
        </button>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button @click="$emit('close')" class="btn btn-secondary">
          {{ cancelLabel }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { CcsClientType } from './keysView'

defineProps<{
  show: boolean
  title: string
  description: string
  claudeLabel: string
  claudeDescription: string
  geminiLabel: string
  geminiDescription: string
  cancelLabel: string
}>()

defineEmits<{
  close: []
  select: [value: CcsClientType]
}>()
</script>
