<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('auth.verifyYourEmail') }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('auth.sendCodeDesc') }}
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ session.email }}</span>
        </p>
      </div>

      <div
        v-if="!session.hasRegisterData"
        class="rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-800/50 dark:bg-amber-900/20"
      >
        <div class="flex items-start gap-3">
          <div class="flex-shrink-0">
            <Icon name="exclamationCircle" size="md" class="text-amber-500" />
          </div>
          <div class="text-sm text-amber-700 dark:text-amber-400">
            <p class="font-medium">{{ t('auth.sessionExpired') }}</p>
            <p class="mt-1">{{ t('auth.sessionExpiredDesc') }}</p>
          </div>
        </div>
      </div>

      <form v-else class="space-y-5" @submit.prevent="handleVerify">
        <div>
          <label for="code" class="input-label text-center">
            {{ t('auth.verificationCode') }}
          </label>
          <input
            id="code"
            v-model="verifyCode"
            type="text"
            required
            autocomplete="one-time-code"
            inputmode="numeric"
            maxlength="6"
            :disabled="isLoading"
            class="input py-3 text-center font-mono text-xl tracking-[0.5em]"
            :class="{ 'input-error': errors.code }"
            placeholder="000000"
          />
          <p v-if="errors.code" class="input-error-text text-center">
            {{ errors.code }}
          </p>
          <p v-else class="input-hint text-center">{{ t('auth.verificationCodeHint') }}</p>
        </div>

        <div
          v-if="codeSent"
          class="rounded-xl border border-green-200 bg-green-50 p-4 dark:border-green-800/50 dark:bg-green-900/20"
        >
          <div class="flex items-start gap-3">
            <div class="flex-shrink-0">
              <Icon name="checkCircle" size="md" class="text-green-500" />
            </div>
            <p class="text-sm text-green-700 dark:text-green-400">
              {{ t('auth.codeSentSuccess') }}
            </p>
          </div>
        </div>

        <div
          v-if="settings.turnstileEnabled && settings.turnstileSiteKey && showResendTurnstile"
        >
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

        <button type="submit" :disabled="isLoading || !verifyCode" class="btn btn-primary w-full">
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
          <Icon v-else name="checkCircle" size="md" class="mr-2" />
          {{ isLoading ? t('auth.verifying') : t('auth.verifyAndCreate') }}
        </button>

        <div class="text-center">
          <button
            v-if="countdown > 0"
            type="button"
            disabled
            class="cursor-not-allowed text-sm text-gray-400 dark:text-dark-500"
          >
            {{ t('auth.resendCountdown', { countdown }) }}
          </button>
          <button
            v-else
            type="button"
            :disabled="isResendDisabled"
            class="text-sm text-primary-600 transition-colors hover:text-primary-500 disabled:cursor-not-allowed disabled:opacity-50 dark:text-primary-400 dark:hover:text-primary-300"
            @click="handleResendCode"
          >
            <span v-if="isSendingCode">{{ t('auth.sendingCode') }}</span>
            <span v-else-if="settings.turnstileEnabled && !showResendTurnstile">
              {{ t('auth.clickToResend') }}
            </span>
            <span v-else>{{ t('auth.resendCode') }}</span>
          </button>
        </div>
      </form>
    </div>

    <template #footer>
      <button
        class="flex items-center gap-2 text-gray-500 transition-colors hover:text-gray-700 dark:text-dark-400 dark:hover:text-gray-300"
        @click="handleBack"
      >
        <Icon name="arrowLeft" size="sm" />
        {{ t('auth.backToRegistration') }}
      </button>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getPublicSettings, sendVerifyCode } from '@/api/auth'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import Icon from '@/components/icons/Icon.vue'
import { AuthLayout } from '@/components/layout'
import { useAppStore, useAuthStore } from '@/stores'
import { buildAuthErrorMessage } from '@/utils/authError'
import { isRegistrationEmailSuffixAllowed } from '@/utils/registrationEmailPolicy'
import {
  applyEmailVerifyPublicSettings,
  buildEmailVerifyRegisterPayload,
  buildEmailVerifySuffixNotAllowedMessage,
  buildSendVerifyCodePayload,
  createEmailVerifyErrors,
  createEmailVerifySessionState,
  createEmailVerifySettingsState,
  parseRegisterSession,
  validateEmailVerifyCode
} from './email-verify/emailVerifyView'

const { t, locale } = useI18n()

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

const isLoading = ref(false)
const isSendingCode = ref(false)
const errorMessage = ref('')
const codeSent = ref(false)
const verifyCode = ref('')
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const session = reactive(createEmailVerifySessionState())
const settings = reactive(createEmailVerifySettingsState())
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const resendTurnstileToken = ref('')
const showResendTurnstile = ref(false)
const errors = reactive(createEmailVerifyErrors())

const isResendDisabled = computed(
  () =>
    isSendingCode.value ||
    (settings.turnstileEnabled && showResendTurnstile.value && !resendTurnstileToken.value)
)

const buildEmailSuffixNotAllowedMessage = () =>
  buildEmailVerifySuffixNotAllowedMessage(
    String(locale.value || ''),
    settings.registrationEmailSuffixWhitelist,
    t
  )

const clearCountdownTimer = () => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

function startCountdown(seconds: number): void {
  countdown.value = seconds
  clearCountdownTimer()

  countdownTimer = setInterval(() => {
    if (countdown.value > 0) {
      countdown.value -= 1
      return
    }

    clearCountdownTimer()
  }, 1000)
}

function onTurnstileVerify(token: string): void {
  resendTurnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  resendTurnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  resendTurnstileToken.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

async function sendCode(): Promise<void> {
  isSendingCode.value = true
  errorMessage.value = ''

  try {
    if (
      !isRegistrationEmailSuffixAllowed(session.email, settings.registrationEmailSuffixWhitelist)
    ) {
      errorMessage.value = buildEmailSuffixNotAllowedMessage()
      appStore.showError(errorMessage.value)
      return
    }

    const response = await sendVerifyCode(
      buildSendVerifyCodePayload(
        session.email,
        resendTurnstileToken.value,
        session.initialTurnstileToken
      )
    )

    codeSent.value = true
    startCountdown(response.countdown)
    session.initialTurnstileToken = ''
    showResendTurnstile.value = false
    resendTurnstileToken.value = ''
  } catch (error: unknown) {
    errorMessage.value = buildAuthErrorMessage(error, {
      fallback: t('auth.sendCodeFailed')
    })
    appStore.showError(errorMessage.value)
  } finally {
    isSendingCode.value = false
  }
}

async function handleResendCode(): Promise<void> {
  if (settings.turnstileEnabled && !showResendTurnstile.value) {
    showResendTurnstile.value = true
    return
  }

  if (settings.turnstileEnabled && !resendTurnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    return
  }

  await sendCode()
}

function validateForm(): boolean {
  errors.code = validateEmailVerifyCode(verifyCode.value, t)
  return !errors.code
}

async function handleVerify(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    if (
      !isRegistrationEmailSuffixAllowed(session.email, settings.registrationEmailSuffixWhitelist)
    ) {
      errorMessage.value = buildEmailSuffixNotAllowedMessage()
      appStore.showError(errorMessage.value)
      return
    }

    await authStore.register(buildEmailVerifyRegisterPayload(session, verifyCode.value))
    sessionStorage.removeItem('register_data')
    appStore.showSuccess(t('auth.accountCreatedSuccess', { siteName: settings.siteName }))
    await router.push('/dashboard')
  } catch (error: unknown) {
    errorMessage.value = buildAuthErrorMessage(error, {
      fallback: t('auth.verifyFailed')
    })
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

function handleBack(): void {
  sessionStorage.removeItem('register_data')
  void router.push('/register')
}

onMounted(async () => {
  Object.assign(session, parseRegisterSession(sessionStorage.getItem('register_data')))

  try {
    applyEmailVerifyPublicSettings(settings, await getPublicSettings())
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }

  if (session.hasRegisterData) {
    await sendCode()
  }
})

onUnmounted(() => {
  clearCountdownTimer()
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
