<template>
  <teleport to="body">
    <transition name="modal">
      <div
        v-if="show"
        class="backup-r2-guide-modal fixed inset-0 z-50 flex items-center justify-center"
        @mousedown.self="emit('close')"
      >
        <div class="backup-r2-guide-modal__backdrop fixed inset-0" @click="emit('close')"></div>
        <div class="backup-r2-guide-modal__panel">
          <button
            type="button"
            class="backup-r2-guide-modal__close-button"
            @click="emit('close')"
          >
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>

          <h2 class="backup-r2-guide-modal__title">
            {{ t('admin.backup.r2Guide.title') }}
          </h2>
          <p class="backup-r2-guide-modal__intro">
            {{ t('admin.backup.r2Guide.intro') }}
          </p>

          <div class="backup-r2-guide-modal__section">
            <h3 class="backup-r2-guide-modal__section-title">
              <span class="backup-r2-guide-modal__step-badge">1</span>
              {{ t('admin.backup.r2Guide.step1.title') }}
            </h3>
            <ol class="backup-r2-guide-modal__list">
              <li>{{ t('admin.backup.r2Guide.step1.line1') }}</li>
              <li>{{ t('admin.backup.r2Guide.step1.line2') }}</li>
              <li>{{ t('admin.backup.r2Guide.step1.line3') }}</li>
            </ol>
          </div>

          <div class="backup-r2-guide-modal__section">
            <h3 class="backup-r2-guide-modal__section-title">
              <span class="backup-r2-guide-modal__step-badge">2</span>
              {{ t('admin.backup.r2Guide.step2.title') }}
            </h3>
            <ol class="backup-r2-guide-modal__list">
              <li>{{ t('admin.backup.r2Guide.step2.line1') }}</li>
              <li>{{ t('admin.backup.r2Guide.step2.line2') }}</li>
              <li>{{ t('admin.backup.r2Guide.step2.line3') }}</li>
              <li>{{ t('admin.backup.r2Guide.step2.line4') }}</li>
            </ol>
            <div class="backup-r2-guide-modal__notice backup-r2-guide-modal__notice--warning">
              {{ t('admin.backup.r2Guide.step2.warning') }}
            </div>
          </div>

          <div class="backup-r2-guide-modal__section">
            <h3 class="backup-r2-guide-modal__section-title">
              <span class="backup-r2-guide-modal__step-badge">3</span>
              {{ t('admin.backup.r2Guide.step3.title') }}
            </h3>
            <p class="backup-r2-guide-modal__section-text">
              {{ t('admin.backup.r2Guide.step3.desc') }}
            </p>
            <code class="backup-r2-guide-modal__code-block">
              https://&lt;{{ t('admin.backup.r2Guide.step3.accountId') }}&gt;.r2.cloudflarestorage.com
            </code>
          </div>

          <div class="backup-r2-guide-modal__section">
            <h3 class="backup-r2-guide-modal__section-title">
              <span class="backup-r2-guide-modal__step-badge">4</span>
              {{ t('admin.backup.r2Guide.step4.title') }}
            </h3>
            <div class="backup-r2-guide-modal__table-shell">
              <table class="w-full text-sm">
                <tbody>
                  <tr
                    v-for="(row, index) in r2ConfigRows"
                    :key="index"
                    class="backup-r2-guide-modal__table-row"
                  >
                    <td class="backup-r2-guide-modal__table-key">
                      {{ row.field }}
                    </td>
                    <td class="backup-r2-guide-modal__table-value">
                      <code class="text-xs">{{ row.value }}</code>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="backup-r2-guide-modal__notice backup-r2-guide-modal__notice--success">
            {{ t('admin.backup.r2Guide.freeTier') }}
          </div>

          <div class="backup-r2-guide-modal__footer">
            <button type="button" class="btn btn-primary btn-sm" @click="emit('close')">
              {{ t('common.close') }}
            </button>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { buildBackupR2ConfigRows } from './backupView'

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const r2ConfigRows = computed(() => buildBackupR2ConfigRows(t))
</script>

<style scoped>
.backup-r2-guide-modal__backdrop {
  background: var(--theme-overlay-strong);
}

.backup-r2-guide-modal__panel {
  position: relative;
  max-height: 85vh;
  width: 100%;
  max-width: 42rem;
  overflow-y: auto;
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
  padding: 1.5rem;
}

.backup-r2-guide-modal__close-button {
  position: absolute;
  right: 1rem;
  top: 1rem;
  color: var(--theme-page-muted);
  transition: color 0.18s ease, background-color 0.18s ease;
}

.backup-r2-guide-modal__close-button:hover,
.backup-r2-guide-modal__close-button:focus-visible {
  color: var(--theme-page-text);
  outline: none;
}

.backup-r2-guide-modal__title {
  color: var(--theme-page-text);
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.backup-r2-guide-modal__intro,
.backup-r2-guide-modal__section-text,
.backup-r2-guide-modal__table-value {
  color: var(--theme-page-muted);
}

.backup-r2-guide-modal__intro {
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.backup-r2-guide-modal__section {
  margin-bottom: 1.25rem;
}

.backup-r2-guide-modal__section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--theme-page-text);
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.backup-r2-guide-modal__step-badge {
  display: inline-flex;
  height: 1.5rem;
  width: 1.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--theme-accent) 14%, var(--theme-surface));
  color: var(--theme-accent);
  font-size: 0.75rem;
  font-weight: 700;
}

.backup-r2-guide-modal__list,
.backup-r2-guide-modal__section-text,
.backup-r2-guide-modal__code-block,
.backup-r2-guide-modal__table-shell {
  margin-left: 2rem;
}

.backup-r2-guide-modal__list {
  list-style: decimal;
  color: var(--theme-page-muted);
  font-size: 0.875rem;
  padding-left: 1rem;
}

.backup-r2-guide-modal__notice {
  border-radius: calc(var(--theme-button-radius) + 2px);
  font-size: 0.75rem;
  margin-top: 0.5rem;
  padding: 0.75rem;
}

.backup-r2-guide-modal__notice--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 74%, var(--theme-page-text));
}

.backup-r2-guide-modal__notice--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 74%, var(--theme-page-text));
}

.backup-r2-guide-modal__code-block {
  display: block;
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
  font-size: 0.75rem;
  margin-top: 0.25rem;
  padding: 0.5rem 0.75rem;
}

.backup-r2-guide-modal__table-shell {
  overflow: hidden;
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-button-radius) + 2px);
}

.backup-r2-guide-modal__table-row {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
}

.backup-r2-guide-modal__table-row:last-child {
  border-bottom: 0;
}

.backup-r2-guide-modal__table-key {
  white-space: nowrap;
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
  font-weight: 600;
  padding: 0.5rem 0.75rem;
}

.backup-r2-guide-modal__table-value {
  padding: 0.5rem 0.75rem;
}

.backup-r2-guide-modal__footer {
  margin-top: 1rem;
  text-align: right;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.backup-r2-guide-modal {
  padding: var(--theme-settings-card-panel-padding);
}
</style>
