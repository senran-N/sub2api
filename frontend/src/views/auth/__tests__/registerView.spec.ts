import { describe, expect, it } from 'vitest'
import type { PublicSettings } from '@/types'
import {
  applyRegisterPublicSettings,
  buildRegisterEmailSuffixNotAllowedMessage,
  buildRegisterInvitationErrorMessage,
  buildRegisterPromoErrorMessage,
  buildRegisterSessionPayload,
  buildRegisterSubmitPayload,
  createRegisterFormData,
  createRegisterSettingsState,
  hasRegisterFormErrors,
  validateRegisterForm
} from '../register/registerView'

const t = (key: string, params?: Record<string, unknown>) =>
  params ? `${key}:${JSON.stringify(params)}` : key

const createPublicSettings = (): PublicSettings => ({
  registration_enabled: false,
  email_verify_enabled: true,
  registration_email_suffix_whitelist: ['@Example.com', '@foo.bar'],
  promo_code_enabled: true,
  password_reset_enabled: true,
  invitation_code_enabled: true,
  turnstile_enabled: true,
  turnstile_site_key: 'turnstile-site-key',
  site_name: 'Example Site',
  site_logo: '',
  site_subtitle: '',
  api_base_url: '',
  contact_info: '',
  doc_url: '',
  home_content: '',
  hide_ccs_import_button: false,
  purchase_subscription_enabled: false,
  purchase_subscription_url: '',
  custom_menu_items: [],
  custom_endpoints: [],
  linuxdo_oauth_enabled: true,
  sora_client_enabled: false,
  backend_mode_enabled: false,
  version: '1.0.0'
})

describe('registerView', () => {
  it('applies public settings into local register state', () => {
    const state = createRegisterSettingsState()

    applyRegisterPublicSettings(state, createPublicSettings())

    expect(state).toMatchObject({
      registrationEnabled: false,
      emailVerifyEnabled: true,
      promoCodeEnabled: true,
      invitationCodeEnabled: true,
      turnstileEnabled: true,
      turnstileSiteKey: 'turnstile-site-key',
      siteName: 'Example Site',
      linuxdoOAuthEnabled: true,
      registrationEmailSuffixWhitelist: ['@example.com', '@foo.bar']
    })
  })

  it('builds promo and invitation code messages from backend error codes', () => {
    expect(buildRegisterPromoErrorMessage('PROMO_CODE_EXPIRED', t)).toBe('auth.promoCodeExpired')
    expect(buildRegisterPromoErrorMessage('UNKNOWN', t)).toBe('auth.promoCodeInvalid')
    expect(buildRegisterInvitationErrorMessage('INVITATION_CODE_USED', t)).toBe(
      'auth.invitationCodeInvalid'
    )
  })

  it('builds email suffix validation messages with locale-specific separators', () => {
    expect(
      buildRegisterEmailSuffixNotAllowedMessage('zh-CN', ['@Example.com', '@foo.bar'], t)
    ).toContain('@example.com、@foo.bar')

    expect(
      buildRegisterEmailSuffixNotAllowedMessage('en-US', ['@Example.com', '@foo.bar'], t)
    ).toContain('@example.com, @foo.bar')
  })

  it('validates register form fields without fallback noise', () => {
    const formData = createRegisterFormData()

    expect(
      validateRegisterForm({
        emailSuffixWhitelist: ['@example.com'],
        formData,
        invitationCodeEnabled: true,
        locale: 'en-US',
        t,
        turnstileEnabled: true,
        turnstileToken: ''
      })
    ).toEqual({
      email: 'auth.emailRequired',
      password: 'auth.passwordRequired',
      turnstile: 'auth.completeVerification',
      invitation_code: 'auth.invitationCodeRequired'
    })

    formData.email = 'user@other.com'
    formData.password = '12345'
    formData.invitation_code = 'invite'

    const errors = validateRegisterForm({
      emailSuffixWhitelist: ['@example.com'],
      formData,
      invitationCodeEnabled: true,
      locale: 'en-US',
      t,
      turnstileEnabled: false,
      turnstileToken: ''
    })

    expect(errors.email).toContain('@example.com')
    expect(errors.password).toBe('auth.passwordMinLength')
    expect(errors.invitation_code).toBe('')
    expect(hasRegisterFormErrors(errors)).toBe(true)
  })

  it('builds register payloads for email verification and direct submit', () => {
    const formData = {
      email: 'user@example.com',
      password: 'password123',
      promo_code: 'PROMO',
      invitation_code: ''
    }

    expect(buildRegisterSessionPayload(formData, 'turnstile-token')).toEqual({
      email: 'user@example.com',
      password: 'password123',
      turnstile_token: 'turnstile-token',
      promo_code: 'PROMO',
      invitation_code: undefined
    })

    expect(buildRegisterSubmitPayload(formData, false, 'turnstile-token')).toEqual({
      email: 'user@example.com',
      password: 'password123',
      turnstile_token: undefined,
      promo_code: 'PROMO',
      invitation_code: undefined
    })
  })
})
