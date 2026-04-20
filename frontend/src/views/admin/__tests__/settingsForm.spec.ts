import { describe, expect, it } from 'vitest'
import type { DefaultSubscriptionSetting, SystemSettings } from '@/api/admin/settings'
import type { CustomEndpoint, CustomMenuItem } from '@/types'
import {
  addCustomEndpoint,
  addCustomMenuItem,
  addNextDefaultSubscription,
  buildSendTestEmailRequest,
  buildSmtpTestConnectionRequest,
  buildSettingsUpdatePayload,
  createDefaultSettingsForm,
  hydrateSettingsForm,
  moveCustomMenuItem,
  removeCustomEndpoint,
  removeCustomMenuItem,
  removeDefaultSubscription,
  type SettingsForm
} from '../settings/settingsForm'

function createSettingsForm(overrides: Partial<SettingsForm> = {}): SettingsForm {
  return {
    registration_enabled: true,
    email_verify_enabled: false,
    registration_email_suffix_whitelist: [],
    promo_code_enabled: true,
    invitation_code_enabled: false,
    password_reset_enabled: false,
    totp_enabled: false,
    totp_encryption_key_configured: false,
    default_balance: 0,
    default_concurrency: 1,
    default_subscriptions: [],
    site_name: 'Sub2API',
    site_logo: '',
    site_subtitle: 'Subscription to API Conversion Platform',
    api_base_url: '',
    contact_info: '',
    doc_url: '',
    home_content: '',
    backend_mode_enabled: false,
    hide_ccs_import_button: false,
    purchase_subscription_enabled: false,
    purchase_subscription_url: '',
    custom_menu_items: [],
    custom_endpoints: [],
    frontend_url: '',
    smtp_host: '',
    smtp_port: 587,
    smtp_username: '',
    smtp_password: '',
    smtp_password_configured: false,
    smtp_from_email: '',
    smtp_from_name: '',
    smtp_use_tls: true,
    turnstile_enabled: false,
    turnstile_site_key: '',
    turnstile_secret_key: '',
    turnstile_secret_key_configured: false,
    linuxdo_connect_enabled: false,
    linuxdo_connect_client_id: '',
    linuxdo_connect_client_secret: '',
    linuxdo_connect_client_secret_configured: false,
    linuxdo_connect_redirect_url: '',
    enable_model_fallback: false,
    fallback_model_anthropic: 'claude-3-5-sonnet-20241022',
    fallback_model_openai: 'gpt-4o',
    fallback_model_grok: 'grok-3',
    fallback_model_gemini: 'gemini-2.5-pro',
    fallback_model_antigravity: 'gemini-2.5-pro',
    grok_image_output_format: 'local_url',
    grok_video_output_format: 'local_url',
    grok_media_proxy_enabled: true,
    grok_media_cache_retention_hours: 72,
    grok_quota_sync_interval_seconds: 900,
    grok_capability_probe_interval_seconds: 21600,
    grok_session_validity_check_interval: 1800,
    grok_video_timeout: 600,
    enable_identity_patch: true,
    identity_patch_prompt: '',
    ops_monitoring_enabled: true,
    ops_realtime_monitoring_enabled: true,
    ops_query_mode_default: 'auto',
    ops_metrics_interval_seconds: 60,
    min_claude_code_version: '',
    max_claude_code_version: '',
    allow_ungrouped_key_scheduling: false,
    enable_fingerprint_unification: true,
    enable_metadata_passthrough: false,
    enable_cch_signing: false,
    ...overrides
  }
}

function createSystemSettings(overrides: Partial<SystemSettings> = {}): SystemSettings {
  const form = createSettingsForm(overrides as Partial<SettingsForm>)
  const {
    smtp_password,
    turnstile_secret_key,
    linuxdo_connect_client_secret,
    ...settings
  } = form
  void smtp_password
  void turnstile_secret_key
  void linuxdo_connect_client_secret
  return settings
}

describe('custom item helpers', () => {
  it('adds, removes, and moves menu items while keeping sort_order contiguous', () => {
    const items: CustomMenuItem[] = []
    addCustomMenuItem(items)
    addCustomMenuItem(items)
    addCustomMenuItem(items)

    items[0].id = 'first'
    items[1].id = 'second'
    items[2].id = 'third'

    moveCustomMenuItem(items, 2, -1)
    expect(items.map((item) => item.id)).toEqual(['first', 'third', 'second'])
    expect(items.map((item) => item.sort_order)).toEqual([0, 1, 2])

    removeCustomMenuItem(items, 0)
    expect(items.map((item) => item.id)).toEqual(['third', 'second'])
    expect(items.map((item) => item.sort_order)).toEqual([0, 1])
  })

  it('adds and removes endpoints in place', () => {
    const endpoints: CustomEndpoint[] = []
    addCustomEndpoint(endpoints)
    addCustomEndpoint(endpoints)
    endpoints[0].name = 'alpha'
    endpoints[1].name = 'beta'

    removeCustomEndpoint(endpoints, 0)
    expect(endpoints).toEqual([
      {
        name: 'beta',
        endpoint: '',
        description: ''
      }
    ])
  })
})

describe('default subscription helpers', () => {
  it('adds the first missing group and removes by index', () => {
    const subscriptions: DefaultSubscriptionSetting[] = [
      {
        group_id: 2,
        validity_days: 30
      }
    ]

    addNextDefaultSubscription(subscriptions, [{ id: 2 }, { id: 5 }, { id: 9 }])
    expect(subscriptions).toEqual([
      {
        group_id: 2,
        validity_days: 30
      },
      {
        group_id: 5,
        validity_days: 30
      }
    ])

    removeDefaultSubscription(subscriptions, 0)
    expect(subscriptions).toEqual([
      {
        group_id: 5,
        validity_days: 30
      }
    ])
  })
})

describe('hydrateSettingsForm', () => {
  it('does not hardcode a Grok fallback model before backend settings load', () => {
    expect(createDefaultSettingsForm().fallback_model_grok).toBe('')
  })

  it('normalizes loaded settings and clears transient secret inputs', () => {
    const form = createSettingsForm({
      smtp_password: 'manual-secret',
      turnstile_secret_key: 'turnstile-secret',
      linuxdo_connect_client_secret: 'linuxdo-secret'
    })

    const tags = hydrateSettingsForm(
      form,
      createSystemSettings({
        registration_email_suffix_whitelist: ['@Example.com', '@example.com', '@foo.bar'],
        default_subscriptions: [
          { group_id: 10, validity_days: 3.9 },
          { group_id: 0, validity_days: 20 },
          { group_id: 12, validity_days: -1 }
        ]
      })
    )

    expect(form.default_subscriptions).toEqual([
      {
        group_id: 10,
        validity_days: 3
      }
    ])
    expect(tags).toEqual(['example.com', 'foo.bar'])
    expect(form.smtp_password).toBe('')
    expect(form.turnstile_secret_key).toBe('')
    expect(form.linuxdo_connect_client_secret).toBe('')
  })
})

describe('buildSettingsUpdatePayload', () => {
  it('normalizes urls without mutating form, clamps subscriptions, prefixes whitelist tags, and omits empty secrets', () => {
    const form = createSettingsForm({
      frontend_url: 'not-a-url',
      doc_url: 'https://docs.example.com',
      purchase_subscription_enabled: false,
      purchase_subscription_url: 'still-not-a-url',
      smtp_password: '',
      turnstile_secret_key: '',
      linuxdo_connect_client_secret: '',
      default_subscriptions: [
        { group_id: 5, validity_days: 45.6 },
        { group_id: -1, validity_days: 30 }
      ]
    })

    const result = buildSettingsUpdatePayload(form, ['example.com', 'foo.bar'])
    expect(result.ok).toBe(true)
    if (!result.ok) {
      return
    }

    expect(form.frontend_url).toBe('not-a-url')
    expect(form.purchase_subscription_url).toBe('still-not-a-url')
    expect(result.payload.registration_email_suffix_whitelist).toEqual([
      '@example.com',
      '@foo.bar'
    ])
    expect(result.payload.frontend_url).toBe('')
    expect(result.payload.doc_url).toBe('https://docs.example.com')
    expect(result.payload.purchase_subscription_url).toBe('')
    expect(result.payload.default_subscriptions).toEqual([
      {
        group_id: 5,
        validity_days: 45
      }
    ])
    expect(result.payload.grok_session_validity_check_interval).toBe(1800)
    expect(result.payload.grok_video_timeout).toBe(600)
    expect(result.payload.smtp_password).toBeUndefined()
    expect(result.payload.turnstile_secret_key).toBeUndefined()
    expect(result.payload.linuxdo_connect_client_secret).toBeUndefined()
  })

  it('rejects duplicate default subscriptions before saving', () => {
    const form = createSettingsForm({
      default_subscriptions: [
        { group_id: 7, validity_days: 30 },
        { group_id: 7, validity_days: 15 }
      ]
    })

    expect(buildSettingsUpdatePayload(form, [])).toEqual({
      ok: false,
      error: {
        code: 'duplicate_default_subscription',
        groupId: 7
      }
    })
  })

  it('rejects missing or invalid purchase urls when purchase mode is enabled', () => {
    const requiredForm = createSettingsForm({
      purchase_subscription_enabled: true,
      purchase_subscription_url: ''
    })
    expect(buildSettingsUpdatePayload(requiredForm, [])).toEqual({
      ok: false,
      error: {
        code: 'purchase_url_required'
      }
    })

    const invalidForm = createSettingsForm({
      purchase_subscription_enabled: true,
      purchase_subscription_url: '/relative'
    })
    expect(buildSettingsUpdatePayload(invalidForm, [])).toEqual({
      ok: false,
      error: {
        code: 'purchase_url_invalid'
      }
    })
  })
})

describe('smtp request helpers', () => {
  it('builds smtp test request and test email request with manual password policy', () => {
    const form = createSettingsForm({
      smtp_host: 'smtp.example.com',
      smtp_port: 465,
      smtp_username: 'mailer',
      smtp_password: 'manual-secret',
      smtp_from_email: 'noreply@example.com',
      smtp_from_name: 'Sub2API',
      smtp_use_tls: true
    })

    expect(buildSmtpTestConnectionRequest(form, false)).toEqual({
      smtp_host: 'smtp.example.com',
      smtp_port: 465,
      smtp_username: 'mailer',
      smtp_password: '',
      smtp_use_tls: true
    })
    expect(buildSmtpTestConnectionRequest(form, true)).toEqual({
      smtp_host: 'smtp.example.com',
      smtp_port: 465,
      smtp_username: 'mailer',
      smtp_password: 'manual-secret',
      smtp_use_tls: true
    })

    expect(buildSendTestEmailRequest(form, 'ops@example.com', false)).toEqual({
      email: 'ops@example.com',
      smtp_host: 'smtp.example.com',
      smtp_port: 465,
      smtp_username: 'mailer',
      smtp_password: '',
      smtp_from_email: 'noreply@example.com',
      smtp_from_name: 'Sub2API',
      smtp_use_tls: true
    })
    expect(buildSendTestEmailRequest(form, 'ops@example.com', true)).toEqual({
      email: 'ops@example.com',
      smtp_host: 'smtp.example.com',
      smtp_port: 465,
      smtp_username: 'mailer',
      smtp_password: 'manual-secret',
      smtp_from_email: 'noreply@example.com',
      smtp_from_name: 'Sub2API',
      smtp_use_tls: true
    })
  })
})
