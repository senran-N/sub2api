<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="$emit('close')">
    <div class="space-y-4">
      <p class="keys-ccs-client-select-dialog__description text-sm">
        {{ description }}
      </p>

      <div class="grid grid-cols-2 gap-3">
        <button
          @click="$emit('select', 'claude')"
          class="keys-ccs-client-select-dialog__option border-2 transition-all"
        >
          <div class="flex flex-col items-center gap-2">
            <Icon name="terminal" size="xl" class="keys-ccs-client-select-dialog__icon" />
            <span class="keys-ccs-client-select-dialog__label font-medium">{{ claudeLabel }}</span>
            <span class="keys-ccs-client-select-dialog__description text-xs">{{ claudeDescription }}</span>
          </div>
        </button>

        <button
          @click="$emit('select', 'gemini')"
          class="keys-ccs-client-select-dialog__option border-2 transition-all"
        >
          <div class="flex flex-col items-center gap-2">
            <Icon name="sparkles" size="xl" class="keys-ccs-client-select-dialog__icon" />
            <span class="keys-ccs-client-select-dialog__label font-medium">{{ geminiLabel }}</span>
            <span class="keys-ccs-client-select-dialog__description text-xs">{{ geminiDescription }}</span>
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

<style scoped>
.keys-ccs-client-select-dialog__description,
.keys-ccs-client-select-dialog__icon {
  color: var(--theme-page-muted);
}

.keys-ccs-client-select-dialog__label {
  color: var(--theme-page-text);
}

.keys-ccs-client-select-dialog__option {
  border-radius: var(--theme-key-usage-card-radius);
  padding: var(--theme-markdown-block-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 84%, transparent);
}

.keys-ccs-client-select-dialog__option:hover {
  border-color: var(--theme-accent);
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
}
</style>
