<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="forgot-password-view__title">
          {{ t('auth.forgotPasswordTitle') }}
        </h2>
        <p class="forgot-password-view__subtitle">
          {{ t('auth.forgotPasswordHint') }}
        </p>
      </div>

      <div v-if="isSubmitted" class="space-y-6">
        <div class="forgot-password-view__notice forgot-password-view__notice--success">
          <div class="flex flex-col items-center gap-4 text-center">
            <div class="forgot-password-view__notice-icon-shell flex h-12 w-12 items-center justify-center rounded-full">
              <Icon name="checkCircle" size="lg" class="forgot-password-view__notice-icon forgot-password-view__notice-icon--success" />
            </div>
            <div>
              <h3 class="forgot-password-view__notice-title forgot-password-view__notice-title--success text-lg font-semibold">
                {{ t('auth.resetEmailSent') }}
              </h3>
              <p class="forgot-password-view__notice-text forgot-password-view__notice-text--success mt-2 text-sm">
                {{ t('auth.resetEmailSentHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="text-center">
          <router-link
            to="/login"
            class="forgot-password-view__inline-link inline-flex items-center gap-2 font-medium"
          >
            <Icon name="arrowLeft" size="sm" />
            {{ t('auth.backToLogin') }}
          </router-link>
        </div>
      </div>

      <form v-else class="space-y-5" @submit.prevent="handleSubmit">
        <div>
          <label for="email" class="input-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="relative">
            <div class="forgot-password-view__email-icon pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="mail" size="md" class="forgot-password-view__email-icon-symbol" />
            </div>
            <input
              id="email"
              v-model="formData.email"
              type="email"
              required
              autofocus
              autocomplete="email"
              :disabled="isLoading"
              class="input pl-11"
              :class="{ 'input-error': errors.email }"
              :placeholder="t('auth.emailPlaceholder')"
            />
          </div>
          <p v-if="errors.email" class="input-error-text">
            {{ errors.email }}
          </p>
        </div>

        <div v-if="settings.turnstileEnabled && settings.turnstileSiteKey">
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="settings.turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
          <p v-if="errors.turnstile" class="input-error-text mt-2 text-center">
            {{ errors.turnstile }}
          </p>
        </div>

        <transition name="fade">
          <div
            v-if="errorMessage"
            class="forgot-password-view__notice forgot-password-view__notice--danger"
          >
            <div class="flex items-start gap-3">
              <div class="forgot-password-view__notice-icon-shell flex-shrink-0">
                <Icon name="exclamationCircle" size="md" class="forgot-password-view__notice-icon forgot-password-view__notice-icon--danger" />
              </div>
              <p class="forgot-password-view__notice-text forgot-password-view__notice-text--danger text-sm">
                {{ errorMessage }}
              </p>
            </div>
          </div>
        </transition>

        <button
          type="submit"
          :disabled="isSubmitDisabled"
          class="btn btn-primary w-full"
        >
          <svg
            v-if="isLoading"
            class="theme-filled-spinner -ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <Icon v-else name="mail" size="md" class="mr-2" />
          {{ submitLabel }}
        </button>
      </form>
    </div>

    <template #footer>
      <p class="forgot-password-view__footer-text">
        {{ t('auth.rememberedPassword') }}
        <router-link
          to="/login"
          class="forgot-password-view__inline-link font-medium"
        >
          {{ t('auth.signIn') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { forgotPassword, getPublicSettings } from '@/api/auth'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import Icon from '@/components/icons/Icon.vue'
import { AuthLayout } from '@/components/layout'
import { useAppStore } from '@/stores'
import {
  applyForgotPasswordPublicSettings,
  buildForgotPasswordSubmitPayload,
  createForgotPasswordFormData,
  createForgotPasswordFormErrors,
  createForgotPasswordSettingsState,
  hasForgotPasswordFormErrors,
  resolveForgotPasswordErrorMessage,
  validateForgotPasswordForm
} from './forgot-password/forgotPasswordView'

const { t } = useI18n()
const appStore = useAppStore()

const isLoading = ref(false)
const isSubmitted = ref(false)
const errorMessage = ref('')
const settings = reactive(createForgotPasswordSettingsState())
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref('')

const formData = reactive(createForgotPasswordFormData())
const errors = reactive(createForgotPasswordFormErrors())

const isSubmitDisabled = computed(
  () => isLoading.value || (settings.turnstileEnabled && !turnstileToken.value)
)

const submitLabel = computed(() =>
  isLoading.value ? t('auth.sendingResetLink') : t('auth.sendResetLink')
)

function onTurnstileVerify(token: string): void {
  turnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

function validateForm(): boolean {
  Object.assign(
    errors,
    validateForgotPasswordForm({
      formData,
      t,
      turnstileEnabled: settings.turnstileEnabled,
      turnstileToken: turnstileToken.value
    })
  )

  return !hasForgotPasswordFormErrors(errors)
}

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    await forgotPassword(
      buildForgotPasswordSubmitPayload(formData, settings.turnstileEnabled, turnstileToken.value)
    )

    isSubmitted.value = true
    appStore.showSuccess(t('auth.resetEmailSent'))
  } catch (error: unknown) {
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    errorMessage.value = resolveForgotPasswordErrorMessage(error, t)
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

onMounted(async () => {
  try {
    applyForgotPasswordPublicSettings(settings, await getPublicSettings())
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.forgot-password-view__title,
.forgot-password-view__notice-title {
  color: var(--theme-page-text);
}

.forgot-password-view__title {
  font-size: 1.5rem;
  font-weight: 700;
}

.forgot-password-view__subtitle,
.forgot-password-view__footer-text {
  color: var(--theme-page-muted);
  font-size: 0.875rem;
}

.forgot-password-view__email-icon-symbol {
  color: var(--theme-page-muted);
}

.forgot-password-view__notice {
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 2px);
}

.forgot-password-view__notice--success {
  padding: var(--theme-auth-callback-card-padding);
}

.forgot-password-view__notice-icon-shell {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.forgot-password-view__notice--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 24%, var(--theme-card-border));
}

.forgot-password-view__notice-icon--success,
.forgot-password-view__notice-text--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.forgot-password-view__notice--danger {
  padding: var(--theme-markdown-block-padding);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 24%, var(--theme-card-border));
}

.forgot-password-view__notice-icon--danger,
.forgot-password-view__notice-text--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.forgot-password-view__inline-link {
  color: var(--theme-accent);
  transition: color 0.2s ease;
}

.forgot-password-view__inline-link:hover {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}
</style>
