<template>
  <BaseDialog
    :show="show"
    :title="t('keys.useKeyModal.title')"
    width="wide"
    @close="emit('close')"
  >
    <div class="space-y-4">
      <!-- No Group Assigned Warning -->
      <div v-if="!platform" class="use-key-modal__notice use-key-modal__notice--warning">
        <svg class="use-key-modal__notice-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
        </svg>
        <div>
          <p class="use-key-modal__notice-title">
            {{ t('keys.useKeyModal.noGroupTitle') }}
          </p>
          <p class="use-key-modal__notice-description">
            {{ t('keys.useKeyModal.noGroupDescription') }}
          </p>
        </div>
      </div>

      <!-- Platform-specific content -->
      <template v-else>
        <!-- Description -->
        <p class="use-key-modal__description">
          {{ platformDescription }}
        </p>

        <!-- Client Tabs -->
        <div v-if="clientTabs.length" class="use-key-modal__tabs-shell">
          <nav class="-mb-px flex space-x-6" aria-label="Client">
            <button
              v-for="tab in clientTabs"
              :key="tab.id"
              @click="activeClientTab = tab.id"
              :class="getTabClasses(activeClientTab === tab.id)"
            >
              <span class="flex items-center gap-2">
                <component :is="tab.icon" class="w-4 h-4" />
                {{ tab.label }}
              </span>
            </button>
          </nav>
        </div>

        <!-- OS/Shell Tabs -->
        <div v-if="showShellTabs" class="use-key-modal__tabs-shell">
          <nav class="-mb-px flex space-x-4" aria-label="Tabs">
            <button
              v-for="tab in currentTabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              :class="getTabClasses(activeTab === tab.id)"
            >
              <span class="flex items-center gap-2">
                <component :is="tab.icon" class="w-4 h-4" />
                {{ tab.label }}
              </span>
            </button>
          </nav>
        </div>

        <!-- Code Blocks (Stacked for multi-file platforms) -->
        <div class="space-y-4">
          <div
            v-for="(file, index) in currentFiles"
            :key="index"
            class="relative"
          >
            <!-- File Hint (if exists) -->
            <p v-if="file.hint" class="use-key-modal__file-hint">
              <Icon name="exclamationCircle" size="sm" class="flex-shrink-0" />
              {{ file.hint }}
            </p>
            <div class="use-key-modal__code-shell">
              <!-- Code Header -->
              <div class="use-key-modal__code-header">
                <span class="use-key-modal__code-path">{{ file.path }}</span>
                <button
                  @click="copyContent(file.content, index)"
                  :class="getCopyButtonClasses(copiedIndex === index)"
                >
                  <svg v-if="copiedIndex === index" class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  <svg v-else class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
                  </svg>
                  {{ copiedIndex === index ? t('keys.useKeyModal.copied') : t('keys.useKeyModal.copy') }}
                </button>
              </div>
              <!-- Code Content -->
              <pre class="use-key-modal__code-content"><code v-if="file.highlighted" v-html="file.highlighted"></code><code v-else v-text="file.content"></code></pre>
            </div>
          </div>
        </div>

        <!-- Usage Note -->
        <div v-if="showPlatformNote" class="use-key-modal__notice use-key-modal__notice--info">
          <Icon name="infoCircle" size="md" class="use-key-modal__notice-icon" />
          <p class="use-key-modal__notice-description">
            {{ platformNote }}
          </p>
        </div>
      </template>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button
          @click="emit('close')"
          class="btn btn-secondary"
        >
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { GroupPlatform } from '@/types'
import { useUseKeyModal } from './useUseKeyModal'

interface Props {
  show: boolean
  apiKey: string
  baseUrl: string
  platform: GroupPlatform | null
  allowMessagesDispatch?: boolean
}

interface Emits {
  (e: 'close'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const {
  activeClientTab,
  activeTab,
  clientTabs,
  copiedIndex,
  copyContent,
  currentFiles,
  currentTabs,
  platformDescription,
  platformNote,
  showPlatformNote,
  showShellTabs
} = useUseKeyModal(props, t)

function joinClassNames(...classNames: Array<string | false | null | undefined>) {
  return classNames.filter(Boolean).join(' ')
}

function getTabClasses(isActive: boolean) {
  return joinClassNames(
    'use-key-modal__tab',
    isActive ? 'use-key-modal__tab--active' : 'use-key-modal__tab--idle'
  )
}

function getCopyButtonClasses(isCopied: boolean) {
  return joinClassNames(
    'use-key-modal__copy-button',
    isCopied ? 'use-key-modal__copy-button--success' : 'use-key-modal__copy-button--idle'
  )
}
</script>

<style scoped>
.use-key-modal__description {
  color: var(--theme-page-muted);
  font-size: 0.875rem;
}

.use-key-modal__notice {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 1rem;
}

.use-key-modal__notice--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 9%, var(--theme-surface));
  border-color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 32%, var(--theme-card-border));
}

.use-key-modal__notice--info {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 9%, var(--theme-surface));
  border-color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 32%, var(--theme-card-border));
}

.use-key-modal__notice-icon {
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.use-key-modal__notice-title {
  font-size: 0.875rem;
  font-weight: 600;
}

.use-key-modal__notice-description {
  margin-top: 0.25rem;
  font-size: 0.875rem;
}

.use-key-modal__notice--warning .use-key-modal__notice-icon,
.use-key-modal__notice--warning .use-key-modal__notice-title,
.use-key-modal__notice--warning .use-key-modal__notice-description,
.use-key-modal__file-hint {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 82%, var(--theme-page-text));
}

.use-key-modal__notice--info .use-key-modal__notice-icon,
.use-key-modal__notice--info .use-key-modal__notice-description {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 82%, var(--theme-page-text));
}

.use-key-modal__tabs-shell {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
}

.use-key-modal__tab {
  border-bottom: 2px solid transparent;
  padding: 0.625rem 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  transition:
    color 0.18s ease,
    border-color 0.18s ease;
}

.use-key-modal__tab--active {
  border-color: var(--theme-accent);
  color: var(--theme-accent);
}

.use-key-modal__tab--idle {
  color: var(--theme-page-muted);
}

.use-key-modal__tab--idle:hover,
.use-key-modal__tab--idle:focus-visible {
  border-color: color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  color: var(--theme-page-text);
  outline: none;
}

.use-key-modal__file-hint {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-bottom: 0.375rem;
  font-size: 0.75rem;
}

.use-key-modal__code-shell {
  overflow: hidden;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  --use-key-modal-terminal-fg: color-mix(in srgb, var(--theme-filled-text) 92%, transparent);
  --use-key-modal-terminal-muted: color-mix(in srgb, var(--theme-filled-text) 58%, transparent);
  --use-key-modal-terminal-subtle: color-mix(in srgb, var(--theme-filled-text) 48%, transparent);
  --use-key-modal-terminal-faint: color-mix(in srgb, var(--theme-filled-text) 34%, transparent);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-page-bg) 6%) 0%,
    color-mix(in srgb, var(--theme-surface-contrast) 90%, var(--theme-accent) 10%) 100%
  );
}

.use-key-modal__code-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-filled-text) 8%, transparent);
  background: color-mix(in srgb, var(--theme-filled-text) 4%, transparent);
  padding: 0.5rem 1rem;
}

.use-key-modal__code-path {
  color: var(--use-key-modal-terminal-muted);
  font-family: var(--theme-font-mono);
  font-size: 0.75rem;
}

.use-key-modal__copy-button {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 600;
  transition:
    background-color 0.18s ease,
    color 0.18s ease;
}

.use-key-modal__copy-button--idle {
  background: color-mix(in srgb, var(--theme-filled-text) 8%, transparent);
  color: color-mix(in srgb, var(--theme-filled-text) 74%, transparent);
}

.use-key-modal__copy-button--idle:hover,
.use-key-modal__copy-button--idle:focus-visible {
  background: color-mix(in srgb, var(--theme-filled-text) 16%, transparent);
  color: var(--theme-filled-text);
  outline: none;
}

.use-key-modal__copy-button--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 18%, transparent);
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 78%, var(--theme-filled-text));
}

.use-key-modal__code-content {
  overflow-x: auto;
  padding: 1rem;
  color: var(--use-key-modal-terminal-fg);
  font-family: var(--theme-font-mono);
  font-size: 0.875rem;
}

.use-key-modal__code-content :deep(.use-key-modal__syntax-keyword) {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 80%, var(--theme-filled-text));
}

.use-key-modal__code-content :deep(.use-key-modal__syntax-variable) {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 72%, var(--theme-filled-text));
}

.use-key-modal__code-content :deep(.use-key-modal__syntax-operator) {
  color: var(--use-key-modal-terminal-subtle);
}

.use-key-modal__code-content :deep(.use-key-modal__syntax-string) {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 78%, var(--theme-filled-text));
}

.use-key-modal__code-content :deep(.use-key-modal__syntax-comment) {
  color: var(--use-key-modal-terminal-faint);
}
</style>
