import type { PublicSettings, RegisterRequest } from '@/types'
import {
  normalizeRegistrationEmailSuffixWhitelist,
  isRegistrationEmailSuffixAllowed
} from '@/utils/registrationEmailPolicy'

export interface RegisterFormData {
  email: string
  password: string
  promo_code: string
  invitation_code: string
  aff_code: string
}

export interface RegisterFormErrors {
  email: string
  password: string
  turnstile: string
  invitation_code: string
}

export interface RegisterCodeValidationState {
  valid: boolean
  invalid: boolean
  message: string
}

export interface RegisterPromoValidationState extends RegisterCodeValidationState {
  bonusAmount: number | null
}

export interface RegisterSettingsState {
  registrationEnabled: boolean
  emailVerifyEnabled: boolean
  promoCodeEnabled: boolean
  invitationCodeEnabled: boolean
  turnstileEnabled: boolean
  turnstileSiteKey: string
  siteName: string
  linuxdoOAuthEnabled: boolean
  wechatOAuthEnabled: boolean
  oidcOAuthEnabled: boolean
  oidcOAuthProviderName: string
  registrationEmailSuffixWhitelist: string[]
}

type Translate = (key: string, params?: Record<string, unknown>) => string

export interface ValidateRegisterFormOptions {
  emailSuffixWhitelist: string[]
  formData: RegisterFormData
  invitationCodeEnabled: boolean
  locale: string
  t: Translate
  turnstileEnabled: boolean
  turnstileToken: string
}

const DEFAULT_SITE_NAME = 'Sub2API'

export function createRegisterFormData(): RegisterFormData {
  return {
    email: '',
    password: '',
    promo_code: '',
    invitation_code: '',
    aff_code: ''
  }
}

export function createRegisterFormErrors(): RegisterFormErrors {
  return {
    email: '',
    password: '',
    turnstile: '',
    invitation_code: ''
  }
}

export function createRegisterPromoValidationState(): RegisterPromoValidationState {
  return {
    valid: false,
    invalid: false,
    bonusAmount: null,
    message: ''
  }
}

export function createRegisterCodeValidationState(): RegisterCodeValidationState {
  return {
    valid: false,
    invalid: false,
    message: ''
  }
}

export function createRegisterSettingsState(): RegisterSettingsState {
  return {
    registrationEnabled: true,
    emailVerifyEnabled: false,
    promoCodeEnabled: true,
    invitationCodeEnabled: false,
    turnstileEnabled: false,
    turnstileSiteKey: '',
    siteName: DEFAULT_SITE_NAME,
    linuxdoOAuthEnabled: false,
    wechatOAuthEnabled: false,
    oidcOAuthEnabled: false,
    oidcOAuthProviderName: 'OIDC',
    registrationEmailSuffixWhitelist: []
  }
}

export function applyRegisterPublicSettings(
  state: RegisterSettingsState,
  settings: PublicSettings
): void {
  state.registrationEnabled = settings.registration_enabled
  state.emailVerifyEnabled = settings.email_verify_enabled
  state.promoCodeEnabled = settings.promo_code_enabled
  state.invitationCodeEnabled = settings.invitation_code_enabled
  state.turnstileEnabled = settings.turnstile_enabled
  state.turnstileSiteKey = settings.turnstile_site_key || ''
  state.siteName = settings.site_name || DEFAULT_SITE_NAME
  state.linuxdoOAuthEnabled = settings.linuxdo_oauth_enabled
  state.wechatOAuthEnabled = settings.wechat_oauth_enabled
  state.oidcOAuthEnabled = settings.oidc_oauth_enabled
  state.oidcOAuthProviderName = settings.oidc_oauth_provider_name || 'OIDC'
  state.registrationEmailSuffixWhitelist = normalizeRegistrationEmailSuffixWhitelist(
    settings.registration_email_suffix_whitelist || []
  )
}

export function resetRegisterPromoValidation(state: RegisterPromoValidationState): void {
  state.valid = false
  state.invalid = false
  state.bonusAmount = null
  state.message = ''
}

export function resetRegisterCodeValidation(state: RegisterCodeValidationState): void {
  state.valid = false
  state.invalid = false
  state.message = ''
}

export function buildRegisterPromoErrorMessage(
  errorCode: string | undefined,
  t: Translate
): string {
  switch (errorCode) {
    case 'PROMO_CODE_NOT_FOUND':
      return t('auth.promoCodeNotFound')
    case 'PROMO_CODE_EXPIRED':
      return t('auth.promoCodeExpired')
    case 'PROMO_CODE_DISABLED':
      return t('auth.promoCodeDisabled')
    case 'PROMO_CODE_MAX_USED':
      return t('auth.promoCodeMaxUsed')
    case 'PROMO_CODE_ALREADY_USED':
      return t('auth.promoCodeAlreadyUsed')
    default:
      return t('auth.promoCodeInvalid')
  }
}

export function buildRegisterInvitationErrorMessage(
  errorCode: string | undefined,
  t: Translate
): string {
  switch (errorCode) {
    case 'INVITATION_CODE_NOT_FOUND':
    case 'INVITATION_CODE_INVALID':
    case 'INVITATION_CODE_USED':
    case 'INVITATION_CODE_DISABLED':
    default:
      return t('auth.invitationCodeInvalid')
  }
}

export function buildRegisterEmailSuffixNotAllowedMessage(
  locale: string,
  whitelist: string[],
  t: Translate
): string {
  const normalizedWhitelist = normalizeRegistrationEmailSuffixWhitelist(whitelist)
  if (normalizedWhitelist.length === 0) {
    return t('auth.emailSuffixNotAllowed')
  }

  const separator = locale.toLowerCase().startsWith('zh') ? '、' : ', '
  return t('auth.emailSuffixNotAllowedWithAllowed', {
    suffixes: normalizedWhitelist.join(separator)
  })
}

export function isRegisterEmailValid(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

export function validateRegisterForm({
  emailSuffixWhitelist,
  formData,
  invitationCodeEnabled,
  locale,
  t,
  turnstileEnabled,
  turnstileToken
}: ValidateRegisterFormOptions): RegisterFormErrors {
  const errors = createRegisterFormErrors()

  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
  } else if (!isRegisterEmailValid(formData.email)) {
    errors.email = t('auth.invalidEmail')
  } else if (!isRegistrationEmailSuffixAllowed(formData.email, emailSuffixWhitelist)) {
    errors.email = buildRegisterEmailSuffixNotAllowedMessage(locale, emailSuffixWhitelist, t)
  }

  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
  }

  if (invitationCodeEnabled && !formData.invitation_code.trim()) {
    errors.invitation_code = t('auth.invitationCodeRequired')
  }

  if (turnstileEnabled && !turnstileToken) {
    errors.turnstile = t('auth.completeVerification')
  }

  return errors
}

export function hasRegisterFormErrors(errors: RegisterFormErrors): boolean {
  return Object.values(errors).some(Boolean)
}

export function buildRegisterSessionPayload(
  formData: RegisterFormData,
  turnstileToken: string
): RegisterRequest {
  return {
    email: formData.email,
    password: formData.password,
    turnstile_token: turnstileToken,
    promo_code: formData.promo_code || undefined,
    invitation_code: formData.invitation_code || undefined,
    aff_code: formData.aff_code || undefined
  }
}

export function buildRegisterSubmitPayload(
  formData: RegisterFormData,
  turnstileEnabled: boolean,
  turnstileToken: string
): RegisterRequest {
  return {
    email: formData.email,
    password: formData.password,
    turnstile_token: turnstileEnabled ? turnstileToken : undefined,
    promo_code: formData.promo_code || undefined,
    invitation_code: formData.invitation_code || undefined,
    aff_code: formData.aff_code || undefined
  }
}
