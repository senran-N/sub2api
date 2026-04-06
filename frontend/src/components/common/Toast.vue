<template>
  <Teleport to="body">
    <div
      class="pointer-events-none fixed right-4 top-4 z-[9999] space-y-3"
      aria-live="polite"
      aria-atomic="true"
    >
      <TransitionGroup
        enter-active-class="transition ease-out duration-300"
        enter-from-class="opacity-0 translate-x-full"
        enter-to-class="opacity-100 translate-x-0"
        leave-active-class="transition ease-in duration-200"
        leave-from-class="opacity-100 translate-x-0"
        leave-to-class="opacity-0 translate-x-full"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="[
            'toast-surface pointer-events-auto overflow-hidden border-l-4',
            getToneClass(toast.type)
          ]"
        >
          <div class="toast-surface__body">
            <div class="flex items-start gap-3">
              <!-- Icon -->
              <div class="mt-0.5 flex-shrink-0">
                <Icon
                  :name="getToastIconName(toast.type)"
                  size="md"
                  :class="['toast-surface__icon', getToneClass(toast.type)]"
                  aria-hidden="true"
                />
              </div>

              <!-- Content -->
              <div class="min-w-0 flex-1">
                <p v-if="toast.title" class="toast-surface__title">
                  {{ toast.title }}
                </p>
                <p
                  :class="[
                    'toast-surface__message text-sm leading-relaxed',
                    toast.title
                      ? 'toast-surface__message--subtle mt-1'
                      : 'toast-surface__message--strong'
                  ]"
                >
                  {{ toast.message }}
                </p>
              </div>

              <!-- Close button -->
              <button
                @click="removeToast(toast.id)"
                class="toast-surface__close"
                aria-label="Close notification"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
          </div>

          <!-- Progress bar -->
          <div v-if="toast.duration" class="toast-surface__progress-track h-1">
            <div
              :class="['h-full toast-progress toast-surface__progress', getToneClass(toast.type)]"
              :style="{ animationDuration: `${toast.duration}ms` }"
            ></div>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const toasts = computed(() => appStore.toasts)

const getToastIconName = (type: string): 'checkCircle' | 'xCircle' | 'exclamationTriangle' | 'infoCircle' => {
  switch (type) {
    case 'success':
      return 'checkCircle'
    case 'error':
      return 'xCircle'
    case 'warning':
      return 'exclamationTriangle'
    case 'info':
    default:
      return 'infoCircle'
  }
}

const getToneClass = (type: string): string => {
  const tones: Record<string, string> = {
    success: 'toast-surface--success',
    error: 'toast-surface--error',
    warning: 'toast-surface--warning',
    info: 'toast-surface--info'
  }
  return tones[type] || tones.info
}

const removeToast = (id: string) => {
  appStore.hideToast(id)
}
</script>

<style scoped>
.toast-surface {
  --toast-tone: var(--theme-accent);
  min-width: min(calc(100vw - 1rem), var(--theme-toast-min-width));
  max-width: var(--theme-toast-max-width);
  border-radius: var(--theme-toast-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
  border-left-color: var(--toast-tone);
}

.toast-surface--success {
  --toast-tone: rgb(var(--theme-success-rgb));
}

.toast-surface--error {
  --toast-tone: rgb(var(--theme-danger-rgb));
}

.toast-surface--warning {
  --toast-tone: rgb(var(--theme-warning-rgb));
}

.toast-surface--info {
  --toast-tone: rgb(var(--theme-info-rgb));
}

.toast-surface__body {
  padding: var(--theme-markdown-block-padding);
}

.toast-surface__icon {
  color: var(--toast-tone);
}

.toast-surface__title,
.toast-surface__message--strong {
  color: var(--theme-page-text);
}

.toast-surface__message--subtle {
  color: var(--theme-page-muted);
}

.toast-surface__close {
  @apply -m-1 flex-shrink-0 transition-colors;
  padding: var(--theme-settings-inline-button-padding);
  border-radius: var(--theme-toast-radius);
  color: var(--theme-input-placeholder);
}

.toast-surface__close:hover {
  background: var(--theme-dropdown-item-hover-bg);
  color: var(--theme-page-text);
}

.toast-surface__progress-track {
  background: color-mix(in srgb, var(--theme-page-border) 75%, transparent);
}

.toast-surface__progress {
  background: var(--toast-tone);
}

.toast-progress {
  width: 100%;
  animation-name: toast-progress-shrink;
  animation-timing-function: linear;
  animation-fill-mode: forwards;
}

@keyframes toast-progress-shrink {
  from {
    width: 100%;
  }
  to {
    width: 0%;
  }
}
</style>
