<template>
  <Teleport to="body">
    <div v-if="show" class="redeem-generated-result-dialog fixed inset-0 z-50 flex items-center justify-center">
      <div class="redeem-generated-result-dialog__backdrop fixed inset-0" @click="emit('close')"></div>
      <div class="redeem-generated-result-dialog__panel relative z-10 w-full">
        <div
          class="redeem-generated-result-dialog__header flex items-center justify-between border-b"
        >
          <div class="flex items-center gap-3">
            <div
              class="redeem-generated-result-dialog__success-icon-shell flex h-10 w-10 items-center justify-center rounded-full"
            >
              <svg
                class="redeem-generated-result-dialog__success-icon h-5 w-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <div>
              <h2 class="redeem-generated-result-dialog__title text-base font-semibold">
                {{ t('admin.redeem.generatedSuccessfully') }}
              </h2>
              <p class="redeem-generated-result-dialog__description text-sm">
                {{ t('admin.redeem.codesCreated', { count }) }}
              </p>
            </div>
          </div>
          <button
            class="redeem-generated-result-dialog__close transition-colors"
            @click="emit('close')"
          >
            <Icon name="x" size="md" :stroke-width="2" />
          </button>
        </div>

        <div class="redeem-generated-result-dialog__body">
          <div class="relative">
            <textarea
              readonly
              :value="codesText"
              :style="{ height: textareaHeight }"
              class="redeem-generated-result-dialog__textarea w-full resize-none font-mono text-sm focus:outline-none"
            ></textarea>
          </div>
        </div>

        <div
          class="redeem-generated-result-dialog__footer flex justify-end gap-2 border-t"
        >
          <button
            :class="[
              'btn flex items-center gap-2 transition-all',
              copiedAll ? 'btn-success' : 'btn-secondary'
            ]"
            @click="emit('copy')"
          >
            <Icon v-if="!copiedAll" name="copy" size="sm" :stroke-width="2" />
            <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 13l4 4L19 7"
              />
            </svg>
            {{ copiedAll ? t('admin.redeem.copied') : t('admin.redeem.copyAll') }}
          </button>
          <button class="btn btn-primary flex items-center gap-2" @click="emit('download')">
            <Icon name="download" size="sm" :stroke-width="2" />
            {{ t('admin.redeem.download') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  show: boolean
  count: number
  codesText: string
  textareaHeight: string
  copiedAll: boolean
}>()

const emit = defineEmits<{
  close: []
  copy: []
  download: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.redeem-generated-result-dialog__backdrop {
  background: var(--theme-overlay-strong);
}

.redeem-generated-result-dialog {
  padding: var(--theme-settings-card-panel-padding);
}

.redeem-generated-result-dialog__panel {
  max-width: min(100%, var(--theme-redeem-result-dialog-width));
  border-radius: calc(var(--theme-surface-radius) + 8px);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
}

.redeem-generated-result-dialog__header,
.redeem-generated-result-dialog__footer {
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  padding: 1rem 1.25rem;
}

.redeem-generated-result-dialog__body {
  padding: 1.25rem;
}

.redeem-generated-result-dialog__success-icon-shell {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 12%, var(--theme-surface));
}

.redeem-generated-result-dialog__success-icon {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.redeem-generated-result-dialog__title {
  color: var(--theme-page-text);
}

.redeem-generated-result-dialog__description,
.redeem-generated-result-dialog__close {
  color: var(--theme-page-muted);
}

.redeem-generated-result-dialog__close {
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 0.375rem;
}

.redeem-generated-result-dialog__close:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}

.redeem-generated-result-dialog__textarea {
  border-radius: calc(var(--theme-surface-radius) + 2px);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-page-text) 88%, transparent);
  padding: 0.75rem;
}

.redeem-generated-result-dialog__footer {
  border-bottom-left-radius: calc(var(--theme-surface-radius) + 8px);
  border-bottom-right-radius: calc(var(--theme-surface-radius) + 8px);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
}
</style>
