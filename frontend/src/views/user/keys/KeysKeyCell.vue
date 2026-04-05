<template>
  <div class="flex items-center gap-2">
    <code class="code text-xs">{{ maskUserApiKey(value) }}</code>
    <button
      type="button"
      class="rounded-lg p-1 transition-colors hover:bg-gray-100 dark:hover:bg-dark-700"
      :class="
        copiedKeyId === rowId
          ? 'text-green-500'
          : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
      "
      :title="copiedKeyId === rowId ? t('keys.copied') : t('keys.copyToClipboard')"
      @click="emit('copy', value, rowId)"
    >
      <Icon v-if="copiedKeyId === rowId" name="check" size="sm" :stroke-width="2" />
      <Icon v-else name="clipboard" size="sm" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { maskUserApiKey } from './keysView'

defineProps<{
  value: string
  rowId: number
  copiedKeyId: number | null
}>()

const emit = defineEmits<{
  copy: [value: string, rowId: number]
}>()

const { t } = useI18n()
</script>
