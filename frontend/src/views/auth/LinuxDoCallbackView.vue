<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="linuxdo-callback-view__title">
          {{ t('auth.linuxdo.callbackTitle') }}
        </h2>
        <p class="linuxdo-callback-view__description mt-2 text-sm">
          {{ isProcessing ? t('auth.linuxdo.callbackProcessing') : t('auth.linuxdo.callbackHint') }}
        </p>
      </div>

      <div v-if="pendingMode === 'choice'" class="space-y-4 rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-800/60">
        <div class="space-y-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">
            {{ t('auth.oauthFlow.chooseHowToContinue') }}
          </p>
          <p class="text-xs text-gray-500 dark:text-dark-400">
            {{ suggestedEmailText }}
          </p>
        </div>
        <div class="grid gap-3 sm:grid-cols-2">
          <button class="btn btn-secondary w-full" :disabled="isSubmitting" @click="switchToBindMode">
            {{ t('auth.oauthFlow.bindExistingAccount') }}
          </button>
          <button class="btn btn-primary w-full" :disabled="isSubmitting" @click="switchToCreateMode">
            {{ t('auth.oauthFlow.createNewAccount') }}
          </button>
        </div>
      </div>

      <div v-else-if="pendingMode === 'bind'" class="space-y-4">
        <p class="linuxdo-callback-view__body text-sm">
          {{ t('auth.oauthFlow.bindLoginHint', { providerName }) }}
        </p>
        <div class="space-y-3">
          <input
            v-model="bindEmail"
            type="email"
            class="input w-full"
            :placeholder="t('auth.emailPlaceholder')"
            :disabled="isSubmitting"
            @keyup.enter="handleBindLogin"
          />
          <input
            v-model="bindPassword"
            type="password"
            class="input w-full"
            :placeholder="t('auth.passwordPlaceholder')"
            :disabled="isSubmitting"
            @keyup.enter="handleBindLogin"
          />
          <label v-if="hasSuggestedDisplayName" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <input v-model="adoptDisplayName" type="checkbox" class="h-4 w-4" :disabled="isSubmitting" />
            <span>{{ t('auth.oauthFlow.useProviderDisplayName') }}</span>
          </label>
          <p v-if="accountActionError" class="linuxdo-callback-view__error-text text-sm">
            {{ accountActionError }}
          </p>
          <button class="btn btn-primary w-full" :disabled="isSubmitting || !bindEmail.trim() || !bindPassword" @click="handleBindLogin">
            {{ isSubmitting ? t('common.processing') : t('auth.oauthFlow.logInAndBind') }}
          </button>
          <button class="btn btn-secondary w-full" :disabled="isSubmitting" @click="switchToCreateMode">
            {{ t('auth.oauthFlow.useDifferentEmail') }}
          </button>
        </div>
      </div>

      <div v-else-if="pendingMode === 'create'" class="space-y-4">
        <p class="linuxdo-callback-view__body text-sm">
          {{ t('auth.oauthFlow.createAccountHint') }}
        </p>
        <div class="space-y-3">
          <input
            v-model="createEmail"
            type="email"
            class="input w-full"
            :placeholder="t('auth.emailPlaceholder')"
            :disabled="isSubmitting"
          />
          <div class="flex gap-2">
            <input
              v-model="verifyCode"
              type="text"
              inputmode="numeric"
              maxlength="6"
              class="input flex-1"
              :placeholder="t('auth.verificationCode')"
              :disabled="isSubmitting"
            />
            <button class="btn btn-secondary whitespace-nowrap" :disabled="isSendCodeDisabled" @click="handleSendCode">
              {{ sendCodeButtonText }}
            </button>
          </div>
          <input
            v-model="createPassword"
            type="password"
            class="input w-full"
            :placeholder="t('auth.passwordPlaceholder')"
            :disabled="isSubmitting"
          />
          <input
            v-model="invitationCode"
            type="text"
            class="input w-full"
            :placeholder="t('auth.invitationCodePlaceholder')"
            :disabled="isSubmitting"
          />
          <label v-if="hasSuggestedDisplayName" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <input v-model="adoptDisplayName" type="checkbox" class="h-4 w-4" :disabled="isSubmitting" />
            <span>{{ t('auth.oauthFlow.useProviderDisplayName') }}</span>
          </label>
          <p v-if="codeSentNotice" class="text-sm text-green-600 dark:text-green-400">
            {{ codeSentNotice }}
          </p>
          <p v-if="accountActionError" class="linuxdo-callback-view__error-text text-sm">
            {{ accountActionError }}
          </p>
          <button
            class="btn btn-primary w-full"
            :disabled="isSubmitting || !createEmail.trim() || !createPassword || !invitationCode.trim()"
            @click="handleCreateAccount"
          >
            {{ isSubmitting ? t('common.processing') : t('auth.linuxdo.completeRegistration') }}
          </button>
          <button class="btn btn-secondary w-full" :disabled="isSubmitting" @click="switchToBindMode">
            {{ t('auth.oauthFlow.bindExistingAccount') }}
          </button>
        </div>
      </div>

      <div v-else-if="pendingMode === 'totp'" class="space-y-4">
        <p class="linuxdo-callback-view__body text-sm">
          {{ t('auth.oauthFlow.totpHint', { providerName, account: totpUserEmailMasked || providerName }) }}
        </p>
        <div class="space-y-3">
          <input
            v-model="totpCode"
            type="text"
            inputmode="numeric"
            maxlength="6"
            class="input w-full"
            placeholder="123456"
            :disabled="isSubmitting"
            @keyup.enter="handleSubmitTotpChallenge"
          />
          <p v-if="accountActionError" class="linuxdo-callback-view__error-text text-sm">
            {{ accountActionError }}
          </p>
          <button class="btn btn-primary w-full" :disabled="isSubmitting || totpCode.trim().length !== 6" @click="handleSubmitTotpChallenge">
            {{ isSubmitting ? t('common.processing') : t('auth.oauthFlow.verifyAndContinue') }}
          </button>
        </div>
      </div>

      <transition name="fade">
        <div v-if="errorMessage" class="linuxdo-callback-view__error-card">
          <div class="flex items-start gap-3">
            <div class="flex-shrink-0">
              <Icon name="exclamationCircle" size="md" class="linuxdo-callback-view__error-icon" />
            </div>
            <div class="space-y-2">
              <p class="linuxdo-callback-view__error-text text-sm">
                {{ errorMessage }}
              </p>
              <router-link to="/login" class="btn btn-primary">
                {{ t('auth.linuxdo.backToLogin') }}
              </router-link>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore, useAppStore } from '@/stores'
import {
  bindLinuxDoOAuthLogin,
  completeLinuxDoOAuthRegistration,
  exchangePendingOAuthCompletion,
  sendPendingOAuthVerifyCode,
  type PendingOAuthBindLoginResponse,
} from '@/api/auth'
import { resolveErrorMessage } from '@/utils/errorMessage'
import { sanitizeRedirectPath } from '@/utils/url'
import {
  clearOAuthAffiliateCode,
  loadOAuthAffiliateCode,
  resolveAffiliateReferralCode,
  storeOAuthAffiliateCode
} from '@/utils/oauthAffiliate'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const providerName = 'Linux.do'
const isProcessing = ref(true)
const isSubmitting = ref(false)
const isSendingCode = ref(false)
const errorMessage = ref('')
const accountActionError = ref('')
const codeSentNotice = ref('')
const pendingMode = ref<'idle' | 'choice' | 'bind' | 'create' | 'totp'>('idle')
const redirectTo = ref('/dashboard')
const pendingEmail = ref('')
const suggestedDisplayName = ref('')

const bindEmail = ref('')
const bindPassword = ref('')

const createEmail = ref('')
const createPassword = ref('')
const verifyCode = ref('')
const invitationCode = ref('')

const adoptDisplayName = ref(true)
const totpTempToken = ref('')
const totpCode = ref('')
const totpUserEmailMasked = ref('')
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const hasSuggestedDisplayName = computed(() => suggestedDisplayName.value.trim().length > 0)
const suggestedEmailText = computed(() => {
  return pendingEmail.value
    ? t('auth.oauthFlow.suggestedEmail', { email: pendingEmail.value })
    : t('auth.oauthFlow.chooseAccountActionHint')
})
const isSendCodeDisabled = computed(() => isSendingCode.value || isSubmitting.value || !createEmail.value.trim() || countdown.value > 0)
const sendCodeButtonText = computed(() => {
  if (isSendingCode.value) return t('common.processing')
  if (countdown.value > 0) return t('auth.resendCountdown', { countdown: countdown.value })
  return t('auth.oauthFlow.sendCodeAction')
})

function parseFragmentParams(): URLSearchParams {
  const raw = typeof window !== 'undefined' ? window.location.hash : ''
  const hash = raw.startsWith('#') ? raw.slice(1) : raw
  return new URLSearchParams(hash)
}

function clearCountdown(): void {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

function startCountdown(seconds: number): void {
  clearCountdown()
  countdown.value = seconds
  countdownTimer = setInterval(() => {
    if (countdown.value <= 0) {
      clearCountdown()
      return
    }
    countdown.value -= 1
  }, 1000)
}

function applyPendingState(payload: PendingOAuthBindLoginResponse): void {
  redirectTo.value = sanitizeRedirectPath(
    payload.redirect || (route.query.redirect as string | undefined) || '/dashboard'
  )
  pendingEmail.value = (payload.resolved_email || payload.email || '').trim()
  suggestedDisplayName.value = (payload.suggested_display_name || '').trim()
  adoptDisplayName.value = hasSuggestedDisplayName.value
  createEmail.value = pendingEmail.value
  bindEmail.value = pendingEmail.value

  authStore.setPendingAuthSession({
    provider: payload.provider || 'linuxdo',
    redirect: redirectTo.value,
    adoption_required: payload.adoption_required,
    suggested_display_name: payload.suggested_display_name,
    suggested_avatar_url: payload.suggested_avatar_url,
  })

  if (payload.step === 'choose_account_action_required') {
    pendingMode.value = 'choice'
  } else if (payload.create_account_allowed) {
    pendingMode.value = 'create'
  } else if (payload.existing_account_bindable) {
    pendingMode.value = 'bind'
  } else {
    pendingMode.value = 'choice'
  }
}

async function completeLogin(tokens: { access_token?: string; refresh_token?: string; expires_in?: number }, redirect?: string) {
  if (!tokens.access_token) {
    throw new Error(t('auth.linuxdo.callbackMissingToken'))
  }
  if (tokens.refresh_token) {
    localStorage.setItem('refresh_token', tokens.refresh_token)
  }
  if (tokens.expires_in) {
    localStorage.setItem('token_expires_at', String(Date.now() + tokens.expires_in * 1000))
  }
  await authStore.setToken(tokens.access_token)
  clearOAuthAffiliateCode()
  appStore.showSuccess(t('auth.loginSuccess'))
  await router.replace(sanitizeRedirectPath(redirect || redirectTo.value || '/dashboard'))
}

function switchToBindMode(): void {
  accountActionError.value = ''
  pendingMode.value = 'bind'
}

function switchToCreateMode(): void {
  accountActionError.value = ''
  pendingMode.value = 'create'
}

async function handleSendCode() {
  if (!createEmail.value.trim()) return
  isSendingCode.value = true
  accountActionError.value = ''
  codeSentNotice.value = ''
  try {
    const response = await sendPendingOAuthVerifyCode({ email: createEmail.value.trim() })
    codeSentNotice.value = t('auth.oauthFlow.emailCodeSent')
    startCountdown(response.countdown)
  } catch (error: unknown) {
    accountActionError.value = resolveErrorMessage(error, t('auth.sendCodeFailed'))
  } finally {
    isSendingCode.value = false
  }
}

async function handleBindLogin() {
  isSubmitting.value = true
  accountActionError.value = ''
  try {
    const response = await bindLinuxDoOAuthLogin({
      email: bindEmail.value.trim(),
      password: bindPassword.value,
      adoptDisplayName: adoptDisplayName.value,
    })
    if (response.requires_2fa && response.temp_token) {
      totpTempToken.value = response.temp_token
      totpUserEmailMasked.value = response.user_email_masked || ''
      pendingMode.value = 'totp'
      return
    }
    await completeLogin(response, response.redirect)
  } catch (error: unknown) {
    accountActionError.value = resolveErrorMessage(error, t('auth.loginFailed'))
  } finally {
    isSubmitting.value = false
  }
}

async function handleCreateAccount() {
  if (appStore.cachedPublicSettings?.email_verify_enabled && !verifyCode.value.trim()) {
    sessionStorage.setItem('register_data', JSON.stringify({
      email: createEmail.value.trim(),
      password: createPassword.value,
      invitation_code: invitationCode.value.trim(),
      aff_code: loadOAuthAffiliateCode() || undefined,
      pending_provider: 'linuxdo',
      adopt_display_name: adoptDisplayName.value,
    }))
    await router.push('/email-verify')
    return
  }

  isSubmitting.value = true
  accountActionError.value = ''
  try {
    const response = await completeLinuxDoOAuthRegistration({
      email: createEmail.value.trim(),
      password: createPassword.value,
      verify_code: verifyCode.value.trim() || undefined,
      invitation_code: invitationCode.value.trim() || undefined,
      aff_code: loadOAuthAffiliateCode() || undefined,
      adoptDisplayName: adoptDisplayName.value,
    })

    if (response.access_token) {
      await completeLogin(response, redirectTo.value)
      return
    }

    if (response.step === 'choose_account_action_required') {
      applyPendingState(response)
      accountActionError.value = t('auth.oauthFlow.accountExistsSwitchToBind')
      return
    }

    throw new Error(t('auth.linuxdo.completeRegistrationFailed'))
  } catch (error: unknown) {
    accountActionError.value = resolveErrorMessage(error, t('auth.linuxdo.completeRegistrationFailed'))
  } finally {
    isSubmitting.value = false
  }
}

async function handleSubmitTotpChallenge() {
  if (!totpTempToken.value || totpCode.value.trim().length !== 6) return
  isSubmitting.value = true
  accountActionError.value = ''
  try {
    await authStore.login2FA(totpTempToken.value, totpCode.value.trim())
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.replace(redirectTo.value)
  } catch (error: unknown) {
    accountActionError.value = resolveErrorMessage(error, t('auth.loginFailed'))
  } finally {
    isSubmitting.value = false
  }
}

onMounted(async () => {
  const params = parseFragmentParams()
  const token = params.get('access_token') || ''
  const refreshToken = params.get('refresh_token') || ''
  const expiresInStr = params.get('expires_in') || ''
  const redirect = sanitizeRedirectPath(
    params.get('redirect') || (route.query.redirect as string | undefined) || '/dashboard'
  )
  const error = params.get('error')
  const errorDesc = params.get('error_description') || params.get('error_message') || ''
  storeOAuthAffiliateCode(
    resolveAffiliateReferralCode(
      params.get('aff'),
      params.get('aff_code'),
      route.query.aff,
      route.query.aff_code
    )
  )

  if (error) {
    errorMessage.value = errorDesc || error
    appStore.showError(errorMessage.value)
    isProcessing.value = false
    return
  }

  if (token) {
    try {
      if (refreshToken) {
        localStorage.setItem('refresh_token', refreshToken)
      }
      if (expiresInStr) {
        const expiresIn = parseInt(expiresInStr, 10)
        if (!Number.isNaN(expiresIn)) {
          localStorage.setItem('token_expires_at', String(Date.now() + expiresIn * 1000))
        }
      }
      await authStore.setToken(token)
      appStore.showSuccess(t('auth.loginSuccess'))
      await router.replace(redirect)
      return
    } catch (error: unknown) {
      errorMessage.value = resolveErrorMessage(error, t('auth.loginFailed'))
      appStore.showError(errorMessage.value)
      isProcessing.value = false
      return
    }
  }

  try {
    const response = await exchangePendingOAuthCompletion()

    if (response.auth_result === 'bind') {
      authStore.clearPendingAuthSession()
      await authStore.refreshUser()
      appStore.showSuccess(t('common.saved'))
      await router.replace(sanitizeRedirectPath(response.redirect || '/profile'))
      return
    }

    applyPendingState(response)
    isProcessing.value = false
  } catch (error: unknown) {
    errorMessage.value = resolveErrorMessage(error, t('auth.linuxdo.callbackMissingToken'))
    appStore.showError(errorMessage.value)
    isProcessing.value = false
  }
})

onUnmounted(() => {
  clearCountdown()
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

.linuxdo-callback-view__title {
  font-family: var(--theme-auth-section-title-font);
  font-size: var(--theme-auth-section-title-size);
  font-weight: 700;
  letter-spacing: var(--theme-auth-section-title-letter-spacing);
  color: var(--theme-page-text);
}

.linuxdo-callback-view__description {
  color: var(--theme-page-muted);
}

.linuxdo-callback-view__body {
  color: color-mix(in srgb, var(--theme-page-text) 82%, transparent);
}

.linuxdo-callback-view__error-card {
  padding: var(--theme-auth-callback-feedback-padding);
  border-radius: var(--theme-auth-feedback-radius);
  border: 1px solid color-mix(in srgb, rgb(var(--theme-danger-rgb)) 28%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.linuxdo-callback-view__error-icon,
.linuxdo-callback-view__error-text {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
