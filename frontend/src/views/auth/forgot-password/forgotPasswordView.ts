import type { ForgotPasswordRequest } from '@/api/auth'
import type { PublicSettings } from '@/types'

export interface ForgotPasswordFormData {
  email: string
}

export interface ForgotPasswordFormErrors {
  email: string
  turnstile: string
}

export interface ForgotPasswordSettingsState {
  turnstileEnabled: boolean
  turnstileSiteKey: string
}

type Translate = (key: string) => string

interface ForgotPasswordErrorLike {
  message?: string
  response?: {
    data?: {
      detail?: string
      message?: string
    }
  }
}

export interface ValidateForgotPasswordFormOptions {
  formData: ForgotPasswordFormData
  t: Translate
  turnstileEnabled: boolean
  turnstileToken: string
}

export function createForgotPasswordFormData(): ForgotPasswordFormData {
  return {
    email: ''
  }
}

export function createForgotPasswordFormErrors(): ForgotPasswordFormErrors {
  return {
    email: '',
    turnstile: ''
  }
}

export function createForgotPasswordSettingsState(): ForgotPasswordSettingsState {
  return {
    turnstileEnabled: false,
    turnstileSiteKey: ''
  }
}

export function applyForgotPasswordPublicSettings(
  state: ForgotPasswordSettingsState,
  settings: PublicSettings
): void {
  state.turnstileEnabled = settings.turnstile_enabled
  state.turnstileSiteKey = settings.turnstile_site_key || ''
}

export function isForgotPasswordEmailValid(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

export function validateForgotPasswordForm({
  formData,
  t,
  turnstileEnabled,
  turnstileToken
}: ValidateForgotPasswordFormOptions): ForgotPasswordFormErrors {
  const errors = createForgotPasswordFormErrors()

  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
  } else if (!isForgotPasswordEmailValid(formData.email)) {
    errors.email = t('auth.invalidEmail')
  }

  if (turnstileEnabled && !turnstileToken) {
    errors.turnstile = t('auth.completeVerification')
  }

  return errors
}

export function hasForgotPasswordFormErrors(errors: ForgotPasswordFormErrors): boolean {
  return Object.values(errors).some(Boolean)
}

export function buildForgotPasswordSubmitPayload(
  formData: ForgotPasswordFormData,
  turnstileEnabled: boolean,
  turnstileToken: string
): ForgotPasswordRequest {
  return {
    email: formData.email,
    turnstile_token: turnstileEnabled ? turnstileToken : undefined
  }
}

export function resolveForgotPasswordErrorMessage(
  error: unknown,
  t: Translate
): string {
  const forgotPasswordError = error as ForgotPasswordErrorLike | null
  return (
    forgotPasswordError?.response?.data?.detail ||
    forgotPasswordError?.response?.data?.message ||
    forgotPasswordError?.message ||
    t('auth.sendResetLinkFailed')
  )
}
