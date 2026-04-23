<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="register-view__title">
          {{ t('auth.createAccount') }}
        </h2>
        <p class="register-view__description mt-2 text-sm">
          {{ t('auth.signUpToStart', { siteName: settings.siteName }) }}
        </p>
      </div>

      <LinuxDoOAuthSection v-if="settings.linuxdoOAuthEnabled" :disabled="isLoading" />
      <WechatOAuthSection v-if="settings.wechatOAuthEnabled" :disabled="isLoading" />
      <OidcOAuthSection
        v-if="settings.oidcOAuthEnabled"
        :disabled="isLoading"
        :provider-name="settings.oidcOAuthProviderName"
      />

      <div
        v-if="!settings.registrationEnabled && settingsLoaded"
        class="register-view__notice"
      >
        <div class="flex items-start gap-3">
          <div class="flex-shrink-0">
            <Icon name="exclamationCircle" size="md" class="register-view__notice-icon" />
          </div>
          <p class="register-view__notice-text text-sm">
            {{ t('auth.registrationDisabled') }}
          </p>
        </div>
      </div>

      <form v-else class="space-y-5" @submit.prevent="handleRegister">
        <div>
          <label for="email" class="input-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="mail" size="md" class="register-view__field-icon" />
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

        <RegisterPasswordField
          v-model="formData.password"
          :disabled="isLoading"
          :error="errors.password"
        />

        <RegisterCodeField
          v-if="settings.invitationCodeEnabled"
          id="invitation_code"
          v-model="formData.invitation_code"
          :disabled="isLoading"
          :error-text="invitationFieldError"
          icon-name="key"
          :invalid="invitationValidation.invalid"
          :label="t('auth.invitationCodeLabel')"
          :placeholder="t('auth.invitationCodePlaceholder')"
          :success-text="t('auth.invitationCodeValid')"
          :valid="invitationValidation.valid"
          :validating="invitationValidating"
          @input="handleInvitationCodeInput"
        />

        <RegisterCodeField
          v-if="settings.promoCodeEnabled"
          id="promo_code"
          v-model="formData.promo_code"
          :disabled="isLoading"
          :error-text="promoValidation.invalid ? promoValidation.message : ''"
          icon-name="gift"
          :invalid="promoValidation.invalid"
          :label="t('auth.promoCodeLabel')"
          :optional-label="t('common.optional')"
          :placeholder="t('auth.promoCodePlaceholder')"
          success-icon-name="gift"
          :success-text="promoSuccessText"
          :valid="promoValidation.valid"
          :validating="promoValidating"
          @input="handlePromoCodeInput"
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
            class="register-view__error-card"
          >
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon name="exclamationCircle" size="md" class="register-view__error-icon" />
              </div>
              <p class="register-view__error-text text-sm">
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
          <Icon v-else name="userPlus" size="md" class="mr-2" />
          {{ submitLabel }}
        </button>
      </form>
    </div>

    <template #footer>
      <p class="register-view__footer">
        {{ t('auth.alreadyHaveAccount') }}
        <router-link
          to="/login"
          class="register-view__footer-link font-medium transition-colors"
        >
          {{ t('auth.signIn') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getPublicSettings, validateInvitationCode, validatePromoCode } from '@/api/auth'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import WechatOAuthSection from '@/components/auth/WechatOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import Icon from '@/components/icons/Icon.vue'
import { AuthLayout } from '@/components/layout'
import { useAppStore, useAuthStore } from '@/stores'
import { buildAuthErrorMessage } from '@/utils/authError'
import RegisterCodeField from './register/RegisterCodeField.vue'
import RegisterPasswordField from './register/RegisterPasswordField.vue'
import {
  applyRegisterPublicSettings,
  buildRegisterInvitationErrorMessage,
  buildRegisterPromoErrorMessage,
  buildRegisterSessionPayload,
  buildRegisterSubmitPayload,
  createRegisterCodeValidationState,
  createRegisterFormData,
  createRegisterFormErrors,
  createRegisterPromoValidationState,
  createRegisterSettingsState,
  hasRegisterFormErrors,
  resetRegisterCodeValidation,
  resetRegisterPromoValidation,
  validateRegisterForm
} from './register/registerView'

const { t, locale } = useI18n()

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

const isLoading = ref(false)
const settingsLoaded = ref(false)
const errorMessage = ref('')
const settings = reactive(createRegisterSettingsState())
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref('')

const promoValidating = ref(false)
const promoValidation = reactive(createRegisterPromoValidationState())
let promoValidateTimeout: ReturnType<typeof setTimeout> | null = null

const invitationValidating = ref(false)
const invitationValidation = reactive(createRegisterCodeValidationState())
let invitationValidateTimeout: ReturnType<typeof setTimeout> | null = null

const formData = reactive(createRegisterFormData())
const errors = reactive(createRegisterFormErrors())

const invitationFieldError = computed(() =>
  invitationValidation.invalid ? invitationValidation.message : errors.invitation_code
)

const promoSuccessText = computed(() => {
  if (!promoValidation.valid || promoValidation.bonusAmount === null) {
    return ''
  }

  return t('auth.promoCodeValid', {
    amount: promoValidation.bonusAmount.toFixed(2)
  })
})

const isSubmitDisabled = computed(
  () => isLoading.value || (settings.turnstileEnabled && !turnstileToken.value)
)

const submitLabel = computed(() => {
  if (isLoading.value) {
    return t('auth.processing')
  }

  return settings.emailVerifyEnabled ? t('auth.continue') : t('auth.createAccount')
})

const clearValidationTimers = () => {
  if (promoValidateTimeout) {
    clearTimeout(promoValidateTimeout)
    promoValidateTimeout = null
  }

  if (invitationValidateTimeout) {
    clearTimeout(invitationValidateTimeout)
    invitationValidateTimeout = null
  }
}

const validatePromoCodeNow = async (code: string) => {
  if (!code.trim()) {
    return
  }

  promoValidating.value = true

  try {
    const result = await validatePromoCode(code)
    if (result.valid) {
      promoValidation.valid = true
      promoValidation.invalid = false
      promoValidation.bonusAmount = result.bonus_amount || 0
      promoValidation.message = ''
      return
    }

    promoValidation.valid = false
    promoValidation.invalid = true
    promoValidation.bonusAmount = null
    promoValidation.message = buildRegisterPromoErrorMessage(result.error_code, t)
  } catch {
    promoValidation.valid = false
    promoValidation.invalid = true
    promoValidation.message = t('auth.promoCodeInvalid')
  } finally {
    promoValidating.value = false
  }
}

const validateInvitationCodeNow = async (code: string) => {
  invitationValidating.value = true

  try {
    const result = await validateInvitationCode(code)
    if (result.valid) {
      invitationValidation.valid = true
      invitationValidation.invalid = false
      invitationValidation.message = ''
      return
    }

    invitationValidation.valid = false
    invitationValidation.invalid = true
    invitationValidation.message = buildRegisterInvitationErrorMessage(result.error_code, t)
  } catch {
    invitationValidation.valid = false
    invitationValidation.invalid = true
    invitationValidation.message = t('auth.invitationCodeInvalid')
  } finally {
    invitationValidating.value = false
  }
}

function handlePromoCodeInput(value: string): void {
  const code = value.trim()
  resetRegisterPromoValidation(promoValidation)

  if (!code) {
    promoValidating.value = false
    return
  }

  if (promoValidateTimeout) {
    clearTimeout(promoValidateTimeout)
  }

  promoValidateTimeout = setTimeout(() => {
    void validatePromoCodeNow(code)
  }, 500)
}

function handleInvitationCodeInput(value: string): void {
  const code = value.trim()
  resetRegisterCodeValidation(invitationValidation)
  errors.invitation_code = ''

  if (!code) {
    return
  }

  if (invitationValidateTimeout) {
    clearTimeout(invitationValidateTimeout)
  }

  invitationValidateTimeout = setTimeout(() => {
    void validateInvitationCodeNow(code)
  }, 500)
}

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
    validateRegisterForm({
      emailSuffixWhitelist: settings.registrationEmailSuffixWhitelist,
      formData,
      invitationCodeEnabled: settings.invitationCodeEnabled,
      locale: String(locale.value || ''),
      t,
      turnstileEnabled: settings.turnstileEnabled,
      turnstileToken: turnstileToken.value
    })
  )

  return !hasRegisterFormErrors(errors)
}

async function handleRegister(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  if (formData.promo_code.trim()) {
    if (promoValidating.value) {
      errorMessage.value = t('auth.promoCodeValidating')
      return
    }

    if (promoValidation.invalid) {
      errorMessage.value = t('auth.promoCodeInvalidCannotRegister')
      return
    }
  }

  if (settings.invitationCodeEnabled) {
    if (invitationValidating.value) {
      errorMessage.value = t('auth.invitationCodeValidating')
      return
    }

    if (invitationValidation.invalid) {
      errorMessage.value = t('auth.invitationCodeInvalidCannotRegister')
      return
    }

    if (formData.invitation_code.trim() && !invitationValidation.valid) {
      errorMessage.value = t('auth.invitationCodeValidating')
      await validateInvitationCodeNow(formData.invitation_code.trim())

      if (!invitationValidation.valid) {
        errorMessage.value = t('auth.invitationCodeInvalidCannotRegister')
        return
      }
    }
  }

  isLoading.value = true

  try {
    if (settings.emailVerifyEnabled) {
      sessionStorage.setItem(
        'register_data',
        JSON.stringify(buildRegisterSessionPayload(formData, turnstileToken.value))
      )
      await router.push('/email-verify')
      return
    }

    await authStore.register(
      buildRegisterSubmitPayload(formData, settings.turnstileEnabled, turnstileToken.value)
    )

    appStore.showSuccess(t('auth.accountCreatedSuccess', { siteName: settings.siteName }))
    await router.push('/dashboard')
  } catch (error: unknown) {
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    errorMessage.value = buildAuthErrorMessage(error, {
      fallback: t('auth.registrationFailed')
    })
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

onMounted(async () => {
  try {
    const publicSettings = await getPublicSettings()
    applyRegisterPublicSettings(settings, publicSettings)

    if (settings.promoCodeEnabled && typeof route.query.promo === 'string' && route.query.promo) {
      formData.promo_code = route.query.promo
      await validatePromoCodeNow(route.query.promo)
    }
  } catch (error) {
    console.error('Failed to load public settings:', error)
  } finally {
    settingsLoaded.value = true
  }
})

onUnmounted(() => {
  clearValidationTimers()
})
</script>

<style scoped>
.register-view__title {
  font-family: var(--theme-auth-section-title-font);
  font-size: var(--theme-auth-section-title-size);
  font-weight: 700;
  letter-spacing: var(--theme-auth-section-title-letter-spacing);
  color: var(--theme-page-text);
}

.register-view__description,
.register-view__footer,
.register-view__field-icon {
  color: var(--theme-page-muted);
}

.register-view__notice {
  padding: var(--theme-markdown-block-padding);
  border-radius: var(--theme-auth-feedback-radius);
  border: 1px solid color-mix(in srgb, rgb(var(--theme-warning-rgb)) 28%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
}

.register-view__notice-icon,
.register-view__notice-text {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.register-view__error-card {
  padding: var(--theme-markdown-block-padding);
  border-radius: var(--theme-auth-feedback-radius);
  border: 1px solid color-mix(in srgb, rgb(var(--theme-danger-rgb)) 28%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.register-view__error-icon,
.register-view__error-text {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.register-view__footer-link {
  color: var(--theme-accent);
}

.register-view__footer-link:hover {
  color: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}
</style>

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
