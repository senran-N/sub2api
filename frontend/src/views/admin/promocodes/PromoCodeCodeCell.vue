<template>
  <div class="flex items-center space-x-2">
    <code class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ code }}</code>
    <button
      :class="[
        'flex items-center transition-colors',
        copied
          ? 'text-green-500'
          : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
      ]"
      :title="copied ? t('admin.promo.copied') : t('keys.copyToClipboard')"
      @click="emit('copy')"
    >
      <Icon v-if="!copied" name="copy" size="sm" :stroke-width="2" />
      <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M5 13l4 4L19 7"
        />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  code: string
  copied: boolean
}>()

const emit = defineEmits<{
  copy: []
}>()

const { t } = useI18n()
</script>
