<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="reset-password-view__title">
          {{ t('auth.resetPasswordTitle') }}
        </h2>
        <p class="reset-password-view__subtitle">
          {{ t('auth.resetPasswordHint') }}
        </p>
      </div>

      <div v-if="isInvalidLink" class="space-y-6">
        <div class="reset-password-view__notice reset-password-view__notice--danger">
          <div class="reset-password-view__notice-content">
            <div class="reset-password-view__notice-icon-shell">
              <Icon name="exclamationCircle" size="lg" class="reset-password-view__notice-icon reset-password-view__notice-icon--danger" />
            </div>
            <div>
              <h3 class="reset-password-view__notice-title reset-password-view__notice-title--danger">
                {{ t('auth.invalidResetLink') }}
              </h3>
              <p class="reset-password-view__notice-text reset-password-view__notice-text--danger">
                {{ t('auth.invalidResetLinkHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="text-center">
          <router-link
            to="/forgot-password"
            class="reset-password-view__inline-link"
          >
            {{ t('auth.requestNewResetLink') }}
          </router-link>
        </div>
      </div>

      <div v-else-if="isSuccess" class="space-y-6">
        <div class="reset-password-view__notice reset-password-view__notice--success">
          <div class="reset-password-view__notice-content">
            <div class="reset-password-view__notice-icon-shell">
              <Icon name="checkCircle" size="lg" class="reset-password-view__notice-icon reset-password-view__notice-icon--success" />
            </div>
            <div>
              <h3 class="reset-password-view__notice-title reset-password-view__notice-title--success">
                {{ t('auth.passwordResetSuccess') }}
              </h3>
              <p class="reset-password-view__notice-text reset-password-view__notice-text--success">
                {{ t('auth.passwordResetSuccessHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="text-center">
          <router-link to="/login" class="btn btn-primary inline-flex items-center gap-2">
            <Icon name="login" size="md" />
            {{ t('auth.signIn') }}
          </router-link>
        </div>
      </div>

      <form v-else class="space-y-5" @submit.prevent="handleSubmit">
        <div>
          <label for="email" class="input-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="relative">
            <div class="reset-password-view__email-icon">
              <Icon name="mail" size="md" class="reset-password-view__email-icon-symbol" />
            </div>
            <input
              id="email"
              :value="routeState.email"
              type="email"
              readonly
              disabled
              class="input reset-password-view__email-input"
            />
          </div>
        </div>

        <ResetPasswordField
          id="password"
          v-model="formData.password"
          :disabled="isLoading"
          :error="errors.password"
          :label="t('auth.newPassword')"
          :placeholder="t('auth.newPasswordPlaceholder')"
        />

        <ResetPasswordField
          id="confirmPassword"
          v-model="formData.confirmPassword"
          :disabled="isLoading"
          :error="errors.confirmPassword"
          :label="t('auth.confirmPassword')"
          :placeholder="t('auth.confirmPasswordPlaceholder')"
        />

        <transition name="fade">
          <div
            v-if="errorMessage"
            class="reset-password-view__error-banner"
          >
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon name="exclamationCircle" size="md" class="reset-password-view__notice-icon reset-password-view__notice-icon--danger" />
              </div>
              <p class="reset-password-view__notice-text reset-password-view__notice-text--danger">
                {{ errorMessage }}
              </p>
            </div>
          </div>
        </transition>

        <button type="submit" :disabled="isLoading" class="btn btn-primary w-full">
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
          <Icon v-else name="checkCircle" size="md" class="mr-2" />
          {{ isLoading ? t('auth.resettingPassword') : t('auth.resetPassword') }}
        </button>
      </form>
    </div>

    <template #footer>
      <p class="reset-password-view__footer-text">
        {{ t('auth.rememberedPassword') }}
        <router-link
          to="/login"
          class="reset-password-view__inline-link"
        >
          {{ t('auth.signIn') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { resetPassword } from '@/api/auth'
import Icon from '@/components/icons/Icon.vue'
import { AuthLayout } from '@/components/layout'
import { useAppStore } from '@/stores'
import ResetPasswordField from './reset-password/ResetPasswordField.vue'
import {
  buildResetPasswordSubmitPayload,
  createResetPasswordFormData,
  createResetPasswordFormErrors,
  hasResetPasswordFormErrors,
  isResetPasswordLinkInvalid,
  resolveResetPasswordErrorMessage,
  resolveResetPasswordRouteState,
  validateResetPasswordForm
} from './reset-password/resetPasswordView'

const { t } = useI18n()

const route = useRoute()
const appStore = useAppStore()

const isLoading = ref(false)
const isSuccess = ref(false)
const errorMessage = ref('')

const routeState = reactive(resolveResetPasswordRouteState({}))
const formData = reactive(createResetPasswordFormData())
const errors = reactive(createResetPasswordFormErrors())

const isInvalidLink = computed(() => isResetPasswordLinkInvalid(routeState))

function validateForm(): boolean {
  Object.assign(errors, validateResetPasswordForm(formData, t))
  return !hasResetPasswordFormErrors(errors)
}

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    await resetPassword(buildResetPasswordSubmitPayload(routeState, formData))
    isSuccess.value = true
    appStore.showSuccess(t('auth.passwordResetSuccess'))
  } catch (error: unknown) {
    errorMessage.value = resolveResetPasswordErrorMessage(error, t)
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  Object.assign(routeState, resolveResetPasswordRouteState(route.query))
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

.reset-password-view__title,
.reset-password-view__notice-title {
  color: var(--theme-page-text);
}

.reset-password-view__title {
  font-size: 1.5rem;
  font-weight: 700;
}

.reset-password-view__subtitle,
.reset-password-view__footer-text {
  color: var(--theme-page-muted);
  font-size: 0.875rem;
}

.reset-password-view__notice,
.reset-password-view__error-banner {
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  padding: 1.5rem;
}

.reset-password-view__notice--danger,
.reset-password-view__error-banner {
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 32%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 8%, var(--theme-surface));
}

.reset-password-view__notice--success {
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 32%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 8%, var(--theme-surface));
}

.reset-password-view__notice-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  text-align: center;
}

.reset-password-view__notice-icon-shell {
  display: flex;
  height: 3rem;
  width: 3rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--theme-surface-soft) 90%, var(--theme-surface));
}

.reset-password-view__notice-title {
  font-size: 1.125rem;
  font-weight: 600;
}

.reset-password-view__notice-text {
  margin-top: 0.5rem;
  font-size: 0.875rem;
}

.reset-password-view__notice-icon--danger,
.reset-password-view__notice-title--danger,
.reset-password-view__notice-text--danger {
  color: rgb(var(--theme-danger-rgb));
}

.reset-password-view__notice-icon--success,
.reset-password-view__notice-title--success,
.reset-password-view__notice-text--success {
  color: rgb(var(--theme-success-rgb));
}

.reset-password-view__inline-link {
  color: var(--theme-accent);
  font-weight: 500;
  transition: color 0.18s ease;
}

.reset-password-view__inline-link:hover,
.reset-password-view__inline-link:focus-visible {
  color: color-mix(in srgb, var(--theme-accent) 76%, var(--theme-accent-strong));
  outline: none;
}

.reset-password-view__email-icon {
  pointer-events: none;
  position: absolute;
  inset-block: 0;
  left: 0;
  display: flex;
  align-items: center;
  padding-left: 0.875rem;
}

.reset-password-view__email-icon-symbol {
  color: var(--theme-page-muted);
}

.reset-password-view__email-input {
  background: color-mix(in srgb, var(--theme-surface-soft) 90%, var(--theme-input-bg));
  padding-left: 2.75rem;
}
</style>
