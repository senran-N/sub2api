<template>
  <div class="flex items-center gap-1">
    <button
      type="button"
      class="theme-action-button theme-action-button--success keys-row-actions__button flex flex-col items-center gap-0.5"
      @click="emit('use', row)"
    >
      <Icon name="terminal" size="sm" />
      <span class="text-xs">{{ t('keys.useKey') }}</span>
    </button>
    <button
      type="button"
      class="theme-action-button theme-action-button--accent keys-row-actions__button flex flex-col items-center gap-0.5"
      @click="emit('edit', row)"
    >
      <Icon name="edit" size="sm" />
      <span class="text-xs">{{ t('common.edit') }}</span>
    </button>
    <div class="relative" :ref="buttonRef">
      <button
        type="button"
        class="theme-action-button theme-action-button--neutral keys-row-actions__button flex flex-col items-center gap-0.5"
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

<style scoped>
.keys-row-actions__button {
  border-radius: var(--theme-key-row-action-radius);
  padding: var(--theme-key-row-action-padding);
}
</style>
