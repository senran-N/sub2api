import type { PublicSettings, RegisterRequest, SendVerifyCodeRequest } from '@/types'
import { normalizeRegistrationEmailSuffixWhitelist } from '@/utils/registrationEmailPolicy'

export interface EmailVerifyErrors {
  code: string
  turnstile: string
}

export interface EmailVerifySessionState {
  email: string
  password: string
  initialTurnstileToken: string
  promoCode: string
  invitationCode: string
  affiliateCode: string
  pendingProvider: string
  adoptDisplayName: boolean
  hasRegisterData: boolean
}

export interface EmailVerifySettingsState {
  turnstileEnabled: boolean
  turnstileSiteKey: string
  siteName: string
  registrationEmailSuffixWhitelist: string[]
}

type Translate = (key: string, params?: Record<string, unknown>) => string

const DEFAULT_SITE_NAME = 'Sub2API'

interface StoredRegisterSession {
  email?: string
  password?: string
  turnstile_token?: string
  promo_code?: string
  invitation_code?: string
  aff_code?: string
  pending_provider?: string
  adopt_display_name?: boolean
}

export function createEmailVerifyErrors(): EmailVerifyErrors {
  return {
    code: '',
    turnstile: ''
  }
}

export function createEmailVerifySessionState(): EmailVerifySessionState {
  return {
    email: '',
    password: '',
    initialTurnstileToken: '',
    promoCode: '',
    invitationCode: '',
    affiliateCode: '',
    pendingProvider: '',
    adoptDisplayName: false,
    hasRegisterData: false
  }
}

export function createEmailVerifySettingsState(): EmailVerifySettingsState {
  return {
    turnstileEnabled: false,
    turnstileSiteKey: '',
    siteName: DEFAULT_SITE_NAME,
    registrationEmailSuffixWhitelist: []
  }
}

export function parseRegisterSession(storageValue: string | null): EmailVerifySessionState {
  const session = createEmailVerifySessionState()
  if (!storageValue) {
    return session
  }

  try {
    const parsed = JSON.parse(storageValue) as StoredRegisterSession
    session.email = parsed.email || ''
    session.password = parsed.password || ''
    session.initialTurnstileToken = parsed.turnstile_token || ''
    session.promoCode = parsed.promo_code || ''
    session.invitationCode = parsed.invitation_code || ''
    session.affiliateCode = parsed.aff_code || ''
    session.pendingProvider = parsed.pending_provider || ''
    session.adoptDisplayName = parsed.adopt_display_name === true
    session.hasRegisterData = Boolean(session.email && session.password)
  } catch {
    return session
  }

  return session
}

export function applyEmailVerifyPublicSettings(
  state: EmailVerifySettingsState,
  settings: PublicSettings
): void {
  state.turnstileEnabled = settings.turnstile_enabled
  state.turnstileSiteKey = settings.turnstile_site_key || ''
  state.siteName = settings.site_name || DEFAULT_SITE_NAME
  state.registrationEmailSuffixWhitelist = normalizeRegistrationEmailSuffixWhitelist(
    settings.registration_email_suffix_whitelist || []
  )
}

export function buildEmailVerifySuffixNotAllowedMessage(
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

export function validateEmailVerifyCode(code: string, t: Translate): string {
  if (!code.trim()) {
    return t('auth.codeRequired')
  }

  if (!/^\d{6}$/.test(code.trim())) {
    return t('auth.invalidCode')
  }

  return ''
}

export function buildSendVerifyCodePayload(
  email: string,
  resendTurnstileToken: string,
  initialTurnstileToken: string
): SendVerifyCodeRequest {
  return {
    email,
    turnstile_token: resendTurnstileToken || initialTurnstileToken || undefined
  }
}

export function buildEmailVerifyRegisterPayload(
  session: EmailVerifySessionState,
  verifyCode: string
): RegisterRequest {
  return {
    email: session.email,
    password: session.password,
    verify_code: verifyCode.trim(),
    turnstile_token: session.initialTurnstileToken || undefined,
    promo_code: session.promoCode || undefined,
    invitation_code: session.invitationCode || undefined,
    aff_code: session.affiliateCode || undefined
  }
}
