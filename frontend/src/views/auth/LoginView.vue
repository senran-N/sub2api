<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('auth.welcomeBack') }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('auth.signInToAccount') }}
        </p>
      </div>

      <LinuxDoOAuthSection
        v-if="settings.linuxdoOAuthEnabled && !settings.backendModeEnabled"
        :disabled="isLoading"
      />

      <form class="space-y-5" @submit.prevent="handleLogin">
        <div>
          <label for="email" class="input-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="mail" size="md" class="text-gray-400 dark:text-dark-500" />
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

        <LoginPasswordField
          v-model="formData.password"
          :disabled="isLoading"
          :error="errors.password"
          :show-forgot-password="showForgotPassword"
        />

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
            class="rounded-xl border border-red-200 bg-red-50 p-4 dark:border-red-800/50 dark:bg-red-900/20"
          >
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon name="exclamationCircle" size="md" class="text-red-500" />
              </div>
              <p class="text-sm text-red-700 dark:text-red-400">
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
            class="-ml-1 mr-2 h-4 w-4 animate-spin text-white"
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
          <Icon v-else name="login" size="md" class="mr-2" />
          {{ submitLabel }}
        </button>
      </form>
    </div>

    <template v-if="!settings.backendModeEnabled" #footer>
      <p class="text-gray-500 dark:text-dark-400">
        {{ t('auth.dontHaveAccount') }}
        <router-link
          to="/register"
          class="font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
        >
          {{ t('auth.signUp') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>

  <TotpLoginModal
    v-if="totpState.showModal"
    ref="totpModalRef"
    :temp-token="totpState.tempToken"
    :user-email-masked="totpState.userEmailMasked"
    @verify="handle2FAVerify"
    @cancel="handle2FACancel"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getPublicSettings, isTotp2FARequired } from '@/api/auth'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import Icon from '@/components/icons/Icon.vue'
import { AuthLayout } from '@/components/layout'
import { useAppStore, useAuthStore } from '@/stores'
import type { TotpLoginResponse } from '@/types'
import LoginPasswordField from './login/LoginPasswordField.vue'
import {
  applyLoginPublicSettings,
  applyTotpLoginState,
  buildLoginSubmitPayload,
  createLoginFormData,
  createLoginFormErrors,
  createLoginSettingsState,
  createLoginTotpState,
  hasLoginFormErrors,
  resetTotpLoginState,
  resolveLoginErrorMessage,
  resolveLoginRedirectTarget,
  resolveTotpLoginErrorMessage,
  validateLoginForm
} from './login/loginView'

const { t } = useI18n()

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

const isLoading = ref(false)
const errorMessage = ref('')
const settings = reactive(createLoginSettingsState())
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref('')
const totpModalRef = ref<InstanceType<typeof TotpLoginModal> | null>(null)

const formData = reactive(createLoginFormData())
const errors = reactive(createLoginFormErrors())
const totpState = reactive(createLoginTotpState())

const isSubmitDisabled = computed(
  () => isLoading.value || (settings.turnstileEnabled && !turnstileToken.value)
)

const showForgotPassword = computed(
  () => settings.passwordResetEnabled && !settings.backendModeEnabled
)

const submitLabel = computed(() =>
  isLoading.value ? t('auth.signingIn') : t('auth.signIn')
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
    validateLoginForm({
      formData,
      t,
      turnstileEnabled: settings.turnstileEnabled,
      turnstileToken: turnstileToken.value
    })
  )

  return !hasLoginFormErrors(errors)
}

async function handleLogin(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    const response = await authStore.login(
      buildLoginSubmitPayload(formData, settings.turnstileEnabled, turnstileToken.value)
    )

    if (isTotp2FARequired(response)) {
      applyTotpLoginState(totpState, response as TotpLoginResponse)
      isLoading.value = false
      return
    }

    appStore.showSuccess(t('auth.loginSuccess'))
    await router.push(resolveLoginRedirectTarget(router.currentRoute.value.query.redirect))
  } catch (error: unknown) {
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    errorMessage.value = resolveLoginErrorMessage(error, t)
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

async function handle2FAVerify(code: string): Promise<void> {
  if (totpModalRef.value) {
    totpModalRef.value.setVerifying(true)
  }

  try {
    await authStore.login2FA(totpState.tempToken, code)
    resetTotpLoginState(totpState)
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.push(resolveLoginRedirectTarget(router.currentRoute.value.query.redirect))
  } catch (error: unknown) {
    if (totpModalRef.value) {
      totpModalRef.value.setError(resolveTotpLoginErrorMessage(error, t))
      totpModalRef.value.setVerifying(false)
    }
  }
}

function handle2FACancel(): void {
  resetTotpLoginState(totpState)
}

onMounted(async () => {
  const expiredFlag = sessionStorage.getItem('auth_expired')
  if (expiredFlag) {
    sessionStorage.removeItem('auth_expired')
    const message = t('auth.reloginRequired')
    errorMessage.value = message
    appStore.showWarning(message)
  }

  try {
    applyLoginPublicSettings(settings, await getPublicSettings())
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
</style>
