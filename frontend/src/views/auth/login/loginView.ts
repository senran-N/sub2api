import type { LoginRequest, PublicSettings, TotpLoginResponse } from '@/types'
import { sanitizeRedirectPath } from '@/utils/url'

export interface LoginFormData {
  email: string
  password: string
}

export interface LoginFormErrors {
  email: string
  password: string
  turnstile: string
}

export interface LoginSettingsState {
  turnstileEnabled: boolean
  turnstileSiteKey: string
  linuxdoOAuthEnabled: boolean
  wechatOAuthEnabled: boolean
  oidcOAuthEnabled: boolean
  oidcOAuthProviderName: string
  backendModeEnabled: boolean
  passwordResetEnabled: boolean
}

export interface LoginTotpState {
  showModal: boolean
  tempToken: string
  userEmailMasked: string
}

type Translate = (key: string) => string

interface LoginErrorLike {
  message?: string
  response?: {
    data?: {
      detail?: string
      message?: string
    }
  }
}

export interface ValidateLoginFormOptions {
  formData: LoginFormData
  t: Translate
  turnstileEnabled: boolean
  turnstileToken: string
}

export function createLoginFormData(): LoginFormData {
  return {
    email: '',
    password: ''
  }
}

export function createLoginFormErrors(): LoginFormErrors {
  return {
    email: '',
    password: '',
    turnstile: ''
  }
}

export function createLoginSettingsState(): LoginSettingsState {
  return {
    turnstileEnabled: false,
    turnstileSiteKey: '',
    linuxdoOAuthEnabled: false,
    wechatOAuthEnabled: false,
    oidcOAuthEnabled: false,
    oidcOAuthProviderName: 'OIDC',
    backendModeEnabled: false,
    passwordResetEnabled: false
  }
}

export function createLoginTotpState(): LoginTotpState {
  return {
    showModal: false,
    tempToken: '',
    userEmailMasked: ''
  }
}

export function applyLoginPublicSettings(
  state: LoginSettingsState,
  settings: PublicSettings
): void {
  state.turnstileEnabled = settings.turnstile_enabled === true
  state.turnstileSiteKey = settings.turnstile_site_key || ''
  state.linuxdoOAuthEnabled = settings.linuxdo_oauth_enabled === true
  state.wechatOAuthEnabled = settings.wechat_oauth_enabled === true
  state.oidcOAuthEnabled = settings.oidc_oauth_enabled === true
  state.oidcOAuthProviderName = settings.oidc_oauth_provider_name || 'OIDC'
  state.backendModeEnabled = settings.backend_mode_enabled === true
  state.passwordResetEnabled = settings.password_reset_enabled === true
}

export function isLoginEmailValid(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

export function validateLoginForm({
  formData,
  t,
  turnstileEnabled,
  turnstileToken
}: ValidateLoginFormOptions): LoginFormErrors {
  const errors = createLoginFormErrors()

  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
  } else if (!isLoginEmailValid(formData.email)) {
    errors.email = t('auth.invalidEmail')
  }

  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
  }

  if (turnstileEnabled && !turnstileToken) {
    errors.turnstile = t('auth.completeVerification')
  }

  return errors
}

export function hasLoginFormErrors(errors: LoginFormErrors): boolean {
  return Object.values(errors).some(Boolean)
}

export function buildLoginSubmitPayload(
  formData: LoginFormData,
  turnstileEnabled: boolean,
  turnstileToken: string
): LoginRequest {
  return {
    email: formData.email,
    password: formData.password,
    turnstile_token: turnstileEnabled ? turnstileToken : undefined
  }
}

export function applyTotpLoginState(
  state: LoginTotpState,
  response: TotpLoginResponse
): void {
  state.tempToken = response.temp_token || ''
  state.userEmailMasked = response.user_email_masked || ''
  state.showModal = true
}

export function resetTotpLoginState(state: LoginTotpState): void {
  state.showModal = false
  state.tempToken = ''
  state.userEmailMasked = ''
}

export function resolveLoginRedirectTarget(
  redirect: unknown,
  isAdmin = false
): string {
  return sanitizeRedirectPath(
    typeof redirect === 'string' ? redirect : undefined,
    isAdmin ? '/admin/dashboard' : '/dashboard'
  )
}

export function resolveLoginErrorMessage(
  error: unknown,
  t: Translate
): string {
  const loginError = error as LoginErrorLike | null
  return (
    loginError?.response?.data?.detail ||
    loginError?.response?.data?.message ||
    loginError?.message ||
    t('auth.loginFailed')
  )
}

export function resolveTotpLoginErrorMessage(
  error: unknown,
  t: Translate
): string {
  const loginError = error as LoginErrorLike | null
  return (
    loginError?.response?.data?.message ||
    loginError?.message ||
    t('profile.totp.loginFailed')
  )
}
