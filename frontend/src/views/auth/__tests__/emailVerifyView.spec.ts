import { describe, expect, it } from 'vitest'
import type { PublicSettings } from '@/types'
import {
  applyEmailVerifyPublicSettings,
  buildEmailVerifyRegisterPayload,
  buildEmailVerifySuffixNotAllowedMessage,
  buildSendVerifyCodePayload,
  createEmailVerifySettingsState,
  parseRegisterSession,
  validateEmailVerifyCode
} from '../email-verify/emailVerifyView'

const t = (key: string, params?: Record<string, unknown>) =>
  params ? `${key}:${JSON.stringify(params)}` : key

const createPublicSettings = (): PublicSettings => ({
  registration_enabled: true,
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
  linuxdo_oauth_enabled: false,
  sora_client_enabled: false,
  backend_mode_enabled: false,
  version: '1.0.0'
})

describe('emailVerifyView', () => {
  it('parses register session payload from session storage', () => {
    expect(
      parseRegisterSession(
        JSON.stringify({
          email: 'user@example.com',
          password: 'password123',
          turnstile_token: 'turnstile-token',
          promo_code: 'PROMO',
          invitation_code: 'INVITE'
        })
      )
    ).toEqual({
      email: 'user@example.com',
      password: 'password123',
      initialTurnstileToken: 'turnstile-token',
      promoCode: 'PROMO',
      invitationCode: 'INVITE',
      hasRegisterData: true
    })

    expect(parseRegisterSession('{')).toEqual({
      email: '',
      password: '',
      initialTurnstileToken: '',
      promoCode: '',
      invitationCode: '',
      hasRegisterData: false
    })
  })

  it('maps public settings into local email verify state', () => {
    const state = createEmailVerifySettingsState()
    applyEmailVerifyPublicSettings(state, createPublicSettings())

    expect(state).toEqual({
      turnstileEnabled: true,
      turnstileSiteKey: 'turnstile-site-key',
      siteName: 'Example Site',
      registrationEmailSuffixWhitelist: ['@example.com', '@foo.bar']
    })
  })

  it('builds locale-aware suffix messages and code validation messages', () => {
    expect(
      buildEmailVerifySuffixNotAllowedMessage('zh-CN', ['@Example.com', '@foo.bar'], t)
    ).toContain('@example.com、@foo.bar')

    expect(validateEmailVerifyCode('', t)).toBe('auth.codeRequired')
    expect(validateEmailVerifyCode('12', t)).toBe('auth.invalidCode')
    expect(validateEmailVerifyCode('123456', t)).toBe('')
  })

  it('builds send-code and final register payloads from explicit state', () => {
    expect(
      buildSendVerifyCodePayload('user@example.com', 'resend-token', 'initial-token')
    ).toEqual({
      email: 'user@example.com',
      turnstile_token: 'resend-token'
    })

    expect(
      buildEmailVerifyRegisterPayload(
        {
          email: 'user@example.com',
          password: 'password123',
          initialTurnstileToken: 'initial-token',
          promoCode: 'PROMO',
          invitationCode: '',
          hasRegisterData: true
        },
        '123456'
      )
    ).toEqual({
      email: 'user@example.com',
      password: 'password123',
      verify_code: '123456',
      turnstile_token: 'initial-token',
      promo_code: 'PROMO',
      invitation_code: undefined
    })
  })
})
