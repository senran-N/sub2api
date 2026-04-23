import { describe, expect, it } from 'vitest'
import type { PublicSettings, TotpLoginResponse } from '@/types'
import {
  applyLoginPublicSettings,
  applyTotpLoginState,
  buildLoginSubmitPayload,
  createLoginSettingsState,
  createLoginTotpState,
  hasLoginFormErrors,
  resetTotpLoginState,
  resolveLoginErrorMessage,
  resolveLoginRedirectTarget,
  resolveTotpLoginErrorMessage,
  shouldShowLoginOAuthDivider,
  validateLoginForm
} from '../login/loginView'

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
  linuxdo_oauth_enabled: true,
  backend_mode_enabled: true,
  version: '1.0.0'
})

describe('loginView', () => {
  it('maps public settings into local login state', () => {
    const state = createLoginSettingsState()
    applyLoginPublicSettings(state, createPublicSettings())

    expect(state).toEqual({
      turnstileEnabled: true,
      turnstileSiteKey: 'turnstile-site-key',
      linuxdoOAuthEnabled: true,
      wechatOAuthEnabled: false,
      oidcOAuthEnabled: false,
      oidcOAuthProviderName: 'OIDC',
      backendModeEnabled: true,
      passwordResetEnabled: true
    })
  })

  it('shows a single oauth divider only when oauth login is available outside backend mode', () => {
    const state = createLoginSettingsState()

    expect(shouldShowLoginOAuthDivider(state)).toBe(false)

    state.wechatOAuthEnabled = true
    expect(shouldShowLoginOAuthDivider(state)).toBe(true)

    state.oidcOAuthEnabled = true
    expect(shouldShowLoginOAuthDivider(state)).toBe(true)

    state.backendModeEnabled = true
    expect(shouldShowLoginOAuthDivider(state)).toBe(false)
  })

  it('validates login form fields and turnstile state', () => {
    const errors = validateLoginForm({
      formData: { email: '', password: '' },
      t,
      turnstileEnabled: true,
      turnstileToken: ''
    })

    expect(errors).toEqual({
      email: 'auth.emailRequired',
      password: 'auth.passwordRequired',
      turnstile: 'auth.completeVerification'
    })
    expect(hasLoginFormErrors(errors)).toBe(true)
  })

  it('builds the login submit payload without fake turnstile fallback', () => {
    expect(
      buildLoginSubmitPayload(
        { email: 'user@example.com', password: 'password123' },
        false,
        'turnstile-token'
      )
    ).toEqual({
      email: 'user@example.com',
      password: 'password123',
      turnstile_token: undefined
    })
  })

  it('tracks totp modal state explicitly', () => {
    const state = createLoginTotpState()
    const response: TotpLoginResponse = {
      requires_2fa: true,
      temp_token: 'temp-token',
      user_email_masked: 'u***@example.com'
    }

    applyTotpLoginState(state, response)
    expect(state).toEqual({
      showModal: true,
      tempToken: 'temp-token',
      userEmailMasked: 'u***@example.com'
    })

    resetTotpLoginState(state)
    expect(state).toEqual({
      showModal: false,
      tempToken: '',
      userEmailMasked: ''
    })
  })

  it('resolves redirect targets and error messages predictably', () => {
    expect(resolveLoginRedirectTarget('/admin', false)).toBe('/admin')
    expect(resolveLoginRedirectTarget('https://evil.example.com', true)).toBe('/admin/dashboard')
    expect(resolveLoginRedirectTarget('//evil.example.com', false)).toBe('/dashboard')
    expect(resolveLoginRedirectTarget(undefined, false)).toBe('/dashboard')
    expect(resolveLoginRedirectTarget(undefined, true)).toBe('/admin/dashboard')

    expect(
      resolveLoginErrorMessage({ response: { data: { detail: 'detail' } } }, t)
    ).toBe('detail')
    expect(
      resolveTotpLoginErrorMessage({ response: { data: { message: 'totp' } } }, t)
    ).toBe('totp')
  })
})
