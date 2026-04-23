<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="email-verify-view__title">
          {{ t('auth.verifyYourEmail') }}
        </h2>
        <p class="email-verify-view__subtitle">
          {{ t('auth.sendCodeDesc') }}
          <span class="email-verify-view__email">{{ session.email }}</span>
        </p>
      </div>

      <div
        v-if="!session.hasRegisterData"
        class="email-verify-view__notice email-verify-view__notice--warning"
      >
        <div class="flex items-start gap-3">
          <div class="email-verify-view__notice-icon-shell flex-shrink-0">
            <Icon name="exclamationCircle" size="md" class="email-verify-view__notice-icon email-verify-view__notice-icon--warning" />
          </div>
          <div class="text-sm">
            <p class="email-verify-view__notice-title email-verify-view__notice-title--warning font-medium">{{ t('auth.sessionExpired') }}</p>
            <p class="email-verify-view__notice-text email-verify-view__notice-text--warning mt-1">{{ t('auth.sessionExpiredDesc') }}</p>
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
            class="email-verify-view__code-input input text-center font-mono text-xl"
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
          class="email-verify-view__notice email-verify-view__notice--success"
        >
          <div class="flex items-start gap-3">
            <div class="email-verify-view__notice-icon-shell flex-shrink-0">
              <Icon name="checkCircle" size="md" class="email-verify-view__notice-icon email-verify-view__notice-icon--success" />
            </div>
            <p class="email-verify-view__notice-text email-verify-view__notice-text--success text-sm">
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
            class="email-verify-view__notice email-verify-view__notice--danger"
          >
            <div class="flex items-start gap-3">
              <div class="email-verify-view__notice-icon-shell flex-shrink-0">
                <Icon name="exclamationCircle" size="md" class="email-verify-view__notice-icon email-verify-view__notice-icon--danger" />
              </div>
              <p class="email-verify-view__notice-text email-verify-view__notice-text--danger text-sm">
                {{ errorMessage }}
              </p>
            </div>
          </div>
        </transition>

        <button type="submit" :disabled="isLoading || !verifyCode" class="btn btn-primary w-full">
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
          {{ isLoading ? t('auth.verifying') : t('auth.verifyAndCreate') }}
        </button>

        <div class="text-center">
          <button
            v-if="countdown > 0"
            type="button"
            disabled
            class="email-verify-view__countdown cursor-not-allowed text-sm"
          >
            {{ t('auth.resendCountdown', { countdown }) }}
          </button>
          <button
            v-else
            type="button"
            :disabled="isResendDisabled"
            class="email-verify-view__inline-link text-sm disabled:cursor-not-allowed disabled:opacity-50"
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
        class="email-verify-view__footer-link flex items-center gap-2"
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
import {
  completeLinuxDoOAuthRegistration,
  getPublicSettings,
  sendPendingOAuthVerifyCode,
  sendVerifyCode
} from '@/api/auth'
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

const isPendingOAuthEmailVerify = computed(
  () => authStore.hasPendingAuthSession && session.pendingProvider === 'linuxdo'
)

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
      !isPendingOAuthEmailVerify.value &&
      !isRegistrationEmailSuffixAllowed(session.email, settings.registrationEmailSuffixWhitelist)
    ) {
      errorMessage.value = buildEmailSuffixNotAllowedMessage()
      appStore.showError(errorMessage.value)
      return
    }

    const payload = buildSendVerifyCodePayload(
      session.email,
      resendTurnstileToken.value,
      session.initialTurnstileToken
    )
    const response = isPendingOAuthEmailVerify.value
      ? await sendPendingOAuthVerifyCode(payload)
      : await sendVerifyCode(payload)

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
      !isPendingOAuthEmailVerify.value &&
      !isRegistrationEmailSuffixAllowed(session.email, settings.registrationEmailSuffixWhitelist)
    ) {
      errorMessage.value = buildEmailSuffixNotAllowedMessage()
      appStore.showError(errorMessage.value)
      return
    }

    if (isPendingOAuthEmailVerify.value) {
      const tokenData = await completeLinuxDoOAuthRegistration({
        email: session.email,
        password: session.password,
        verify_code: verifyCode.value.trim(),
        invitation_code: session.invitationCode || undefined,
        adoptDisplayName: session.adoptDisplayName,
      })

      if (tokenData.access_token) {
        if (tokenData.refresh_token) {
          localStorage.setItem('refresh_token', tokenData.refresh_token)
        }
        if (tokenData.expires_in) {
          localStorage.setItem('token_expires_at', String(Date.now() + tokenData.expires_in * 1000))
        }
        await authStore.setToken(tokenData.access_token)
        authStore.clearPendingAuthSession()
        sessionStorage.removeItem('register_data')
        appStore.showSuccess(t('auth.accountCreatedSuccess', { siteName: settings.siteName }))
        await router.push('/dashboard')
        return
      }

      if (tokenData.step === 'choose_account_action_required') {
        authStore.setPendingAuthSession({
          token: authStore.pendingAuthSession?.token || '',
          token_field: authStore.pendingAuthSession?.token_field || 'pending_auth_token',
          provider: authStore.pendingAuthSession?.provider || session.pendingProvider || 'linuxdo',
          redirect: authStore.pendingAuthSession?.redirect,
        })
        sessionStorage.removeItem('register_data')
        appStore.showError(t('auth.oauthFlow.accountExistsSwitchToBind'))
        await router.push('/auth/linuxdo/callback')
        return
      }

      throw new Error(t('auth.verifyFailed'))
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
  void router.push(isPendingOAuthEmailVerify.value ? '/auth/linuxdo/callback' : '/register')
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

.email-verify-view__title,
.email-verify-view__notice-title {
  color: var(--theme-page-text);
}

.email-verify-view__title {
  font-size: 1.5rem;
  font-weight: 700;
}

.email-verify-view__subtitle,
.email-verify-view__countdown {
  color: var(--theme-page-muted);
  font-size: 0.875rem;
}

.email-verify-view__email {
  color: var(--theme-page-text);
  font-weight: 500;
}

.email-verify-view__notice {
  border: 1px solid var(--theme-card-border);
  border-radius: var(--theme-auth-feedback-radius);
  padding: var(--theme-auth-feedback-padding);
}

.email-verify-view__code-input {
  padding-block: var(--theme-auth-verify-code-padding-y);
  letter-spacing: var(--theme-auth-verify-code-letter-spacing);
}

.email-verify-view__notice-content,
.email-verify-view__notice-icon-shell {
  display: flex;
}

.email-verify-view__notice-icon-shell {
  align-items: center;
  justify-content: center;
}

.email-verify-view__notice--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  border-color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 24%, var(--theme-card-border));
}

.email-verify-view__notice-icon--warning,
.email-verify-view__notice-text--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.email-verify-view__notice--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 24%, var(--theme-card-border));
}

.email-verify-view__notice-icon--success,
.email-verify-view__notice-text--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.email-verify-view__notice--danger {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 24%, var(--theme-card-border));
}

.email-verify-view__notice-icon--danger,
.email-verify-view__notice-text--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.email-verify-view__inline-link,
.email-verify-view__footer-link {
  color: var(--theme-accent);
  transition: color 0.2s ease;
}

.email-verify-view__inline-link:hover,
.email-verify-view__footer-link:hover {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}
</style>
