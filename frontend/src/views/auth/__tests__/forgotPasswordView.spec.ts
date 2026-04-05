import { describe, expect, it } from 'vitest'
import type { PublicSettings } from '@/types'
import {
  applyForgotPasswordPublicSettings,
  buildForgotPasswordSubmitPayload,
  createForgotPasswordSettingsState,
  hasForgotPasswordFormErrors,
  resolveForgotPasswordErrorMessage,
  validateForgotPasswordForm
} from '../forgot-password/forgotPasswordView'

const t = (key: string) => key

const createPublicSettings = (): PublicSettings => ({
  registration_enabled: true,
  email_verify_enabled: false,
  registration_email_suffix_whitelist: [],
  promo_code_enabled: true,
  password_reset_enabled: true,
  invitation_code_enabled: false,
  turnstile_enabled: true,
  turnstile_site_key: 'turnstile-site-key',
  site_name: 'Site',
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

describe('forgotPasswordView', () => {
  it('maps public settings into local forgot password state', () => {
    const state = createForgotPasswordSettingsState()
    applyForgotPasswordPublicSettings(state, createPublicSettings())

    expect(state).toEqual({
      turnstileEnabled: true,
      turnstileSiteKey: 'turnstile-site-key'
    })
  })

  it('validates email and turnstile requirements', () => {
    const errors = validateForgotPasswordForm({
      formData: { email: '' },
      t,
      turnstileEnabled: true,
      turnstileToken: ''
    })

    expect(errors).toEqual({
      email: 'auth.emailRequired',
      turnstile: 'auth.completeVerification'
    })
    expect(hasForgotPasswordFormErrors(errors)).toBe(true)
  })

  it('builds the forgot password payload without extra fallback data', () => {
    expect(
      buildForgotPasswordSubmitPayload({ email: 'user@example.com' }, false, 'turnstile-token')
    ).toEqual({
      email: 'user@example.com',
      turnstile_token: undefined
    })
  })

  it('prefers backend detail when resolving request failure messages', () => {
    expect(
      resolveForgotPasswordErrorMessage({ response: { data: { detail: 'detail' } } }, t)
    ).toBe('detail')
    expect(resolveForgotPasswordErrorMessage({ message: 'generic' }, t)).toBe('generic')
  })
})
