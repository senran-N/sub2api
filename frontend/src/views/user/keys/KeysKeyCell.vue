<template>
  <div class="flex items-center gap-2">
    <code class="code text-xs">{{ maskUserApiKey(value) }}</code>
    <button
      type="button"
      class="keys-key-cell__copy-button theme-icon-button"
      :class="
        copiedKeyId === rowId
          ? 'theme-icon-button--success'
          : 'theme-icon-button--neutral'
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

<style scoped>
.keys-key-cell__copy-button {
  border-radius: var(--theme-key-row-action-radius);
  padding: var(--theme-key-row-action-padding);
}
</style>
