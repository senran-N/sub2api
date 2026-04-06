<template>
  <div class="oauth-callback-view min-h-screen">
    <div class="oauth-callback-view__container mx-auto">
      <div class="card oauth-callback-view__card">
        <h1 class="oauth-callback-view__title">
          {{ t('auth.oauth.callbackTitle') }}
        </h1>
        <p class="oauth-callback-view__description mt-2 text-sm">
          {{ t('auth.oauth.callbackDescription') }}
        </p>

        <div class="mt-6 space-y-4">
          <div>
            <label class="input-label">{{ t('auth.oauth.code') }}</label>
            <div class="flex gap-2">
              <input class="input flex-1 font-mono text-sm" :value="code" readonly />
              <button class="btn btn-secondary" type="button" :disabled="!code" @click="copy(code)">
                {{ t('common.copy') }}
              </button>
            </div>
          </div>

          <div>
            <label class="input-label">{{ t('auth.oauth.state') }}</label>
            <div class="flex gap-2">
              <input class="input flex-1 font-mono text-sm" :value="state" readonly />
              <button
                class="btn btn-secondary"
                type="button"
                :disabled="!state"
                @click="copy(state)"
              >
                {{ t('common.copy') }}
              </button>
            </div>
          </div>

          <div>
            <label class="input-label">{{ t('auth.oauth.fullUrl') }}</label>
            <div class="flex gap-2">
              <input class="input flex-1 font-mono text-xs" :value="fullUrl" readonly />
              <button
                class="btn btn-secondary"
                type="button"
                :disabled="!fullUrl"
                @click="copy(fullUrl)"
              >
                {{ t('common.copy') }}
              </button>
            </div>
          </div>

          <div v-if="error" class="oauth-callback-view__error">
            <p class="oauth-callback-view__error-text text-sm">{{ error }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useClipboard } from '@/composables/useClipboard'

const route = useRoute()
const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const code = computed(() => (route.query.code as string) || '')
const state = computed(() => (route.query.state as string) || '')
const error = computed(
  () => (route.query.error as string) || (route.query.error_description as string) || ''
)

const fullUrl = computed(() => {
  if (typeof window === 'undefined') return ''
  return window.location.href
})

const copy = (value: string) => {
  if (!value) return
  copyToClipboard(value, t('common.copiedToClipboard'))
}
</script>

<style scoped>
.oauth-callback-view {
  padding:
    var(--theme-markdown-block-padding)
    calc(var(--theme-markdown-block-padding) - 0.25rem)
    calc(var(--theme-auth-callback-card-padding) - 0.5rem);
  background:
    radial-gradient(circle at top center, color-mix(in srgb, var(--theme-accent-soft) 28%, transparent), transparent 42%),
    var(--theme-page-bg);
}

.oauth-callback-view__title {
  font-family: var(--theme-auth-callback-title-font);
  font-size: var(--theme-auth-callback-title-size);
  font-weight: 650;
  letter-spacing: var(--theme-auth-callback-title-letter-spacing);
  color: var(--theme-page-text);
}

.oauth-callback-view__description {
  color: var(--theme-page-muted);
}

.oauth-callback-view__container {
  max-width: var(--theme-auth-callback-max-width);
}

.oauth-callback-view__card {
  padding: var(--theme-auth-callback-card-padding);
}

.oauth-callback-view__error {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-danger-rgb)) 26%, var(--theme-card-border));
  padding: var(--theme-auth-callback-feedback-padding);
  border-radius: var(--theme-auth-callback-feedback-radius);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.oauth-callback-view__error-text {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
