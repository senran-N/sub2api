import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { AdminGroup } from '@/types'
import type { SystemSettings } from '@/api/admin/settings'
import { useSettingsViewForm } from '../settings/useSettingsViewForm'

const { getSettings, updateSettings, testSmtpConnection, sendTestEmail, getAllGroups } = vi.hoisted(
  () => ({
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
    testSmtpConnection: vi.fn(),
    sendTestEmail: vi.fn(),
    getAllGroups: vi.fn()
  })
)

vi.mock('@/api', () => ({
  adminAPI: {
    settings: {
      getSettings,
      updateSettings,
      testSmtpConnection,
      sendTestEmail
    },
    groups: {
      getAll: getAllGroups
    }
  }
}))

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })

  return {
    promise,
    resolve,
    reject
  }
}

function createSystemSettings(overrides: Partial<SystemSettings> = {}): SystemSettings {
  return {
    registration_enabled: true,
    email_verify_enabled: false,
    registration_email_suffix_whitelist: ['@example.com'],
    promo_code_enabled: true,
    password_reset_enabled: true,
    frontend_url: '',
    invitation_code_enabled: false,
    totp_enabled: false,
    totp_encryption_key_configured: false,
    default_balance: 5,
    default_concurrency: 1,
    default_subscriptions: [{ group_id: 10, validity_days: 30 }],
    site_name: 'Sub2API',
    site_logo: '',
    site_subtitle: 'Subscription to API Conversion Platform',
    api_base_url: '',
    contact_info: '',
    doc_url: '',
    home_content: '',
    hide_ccs_import_button: false,
    purchase_subscription_enabled: false,
    purchase_subscription_url: '',
    backend_mode_enabled: false,
    custom_menu_items: [],
    custom_endpoints: [],
    smtp_host: 'smtp.example.com',
    smtp_port: 465,
    smtp_username: 'mailer',
    smtp_password_configured: true,
    smtp_from_email: 'noreply@example.com',
    smtp_from_name: 'Sub2API',
    smtp_use_tls: true,
    turnstile_enabled: false,
    turnstile_site_key: '',
    turnstile_secret_key_configured: false,
    linuxdo_connect_enabled: false,
    linuxdo_connect_client_id: '',
    linuxdo_connect_client_secret_configured: false,
    linuxdo_connect_redirect_url: '',
    enable_model_fallback: false,
    fallback_model_anthropic: 'claude-3-5-sonnet-20241022',
    fallback_model_openai: 'gpt-4o',
    fallback_model_grok: 'grok-3',
    fallback_model_gemini: 'gemini-2.5-pro',
    fallback_model_antigravity: 'gemini-2.5-pro',
    grok_thinking_summary: false,
    grok_show_search_sources: false,
    grok_image_output_format: 'upstream_url',
    grok_video_output_format: 'upstream_url',
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

function createGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 10,
    name: 'Starter',
    description: 'Starter plan',
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'subscription',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    sora_image_price_360: null,
    sora_image_price_540: null,
    sora_video_price_per_request: null,
    sora_video_price_per_request_hd: null,
    sora_storage_quota_bytes: 0,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    require_oauth_only: false,
    require_privacy_set: false,
    model_routing: null,
    model_routing_enabled: false,
    mcp_xml_inject: false,
    simulate_claude_max_enabled: false,
    created_at: '2026-04-04T00:00:00Z',
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('useSettingsViewForm', () => {
  beforeEach(() => {
    getSettings.mockReset()
    updateSettings.mockReset()
    testSmtpConnection.mockReset()
    sendTestEmail.mockReset()
    getAllGroups.mockReset()

    getSettings.mockResolvedValue(createSystemSettings())
    updateSettings.mockImplementation(async (payload) => createSystemSettings(payload))
    testSmtpConnection.mockResolvedValue({ message: 'smtp-ok' })
    sendTestEmail.mockResolvedValue({ message: 'mail-ok' })
    getAllGroups.mockResolvedValue([
      createGroup(),
      createGroup({ id: 20, name: 'Disabled', status: 'disabled' }),
      createGroup({ id: 30, name: 'OnDemand', subscription_type: 'package' })
    ])
  })

  it('loads settings, filters subscription groups, and saves updated settings', async () => {
    const showError = vi.fn()
    const showSuccess = vi.fn()
    const refreshPublicSettings = vi.fn().mockResolvedValue(undefined)
    const refreshAdminSettings = vi.fn().mockResolvedValue(undefined)
    const copyToClipboard = vi.fn().mockResolvedValue(true)
    const state = useSettingsViewForm({
      t: (key: string, params?: Record<string, unknown>) =>
        params ? `${key}:${JSON.stringify(params)}` : key,
      showError,
      showSuccess,
      refreshPublicSettings,
      refreshAdminSettings,
      copyToClipboard,
      location: {
        origin: 'https://sub2api.example.com',
        protocol: 'https:',
        host: 'sub2api.example.com'
      }
    })

    await state.loadSettings()
    await state.loadSubscriptionGroups()

    expect(state.form.site_name).toBe('Sub2API')
    expect(state.registrationEmailSuffixWhitelistTags.value).toEqual(['example.com'])
    expect(state.form.smtp_password).toBe('')
    expect(state.defaultSubscriptionGroupOptions.value).toEqual([
      expect.objectContaining({
        value: 10,
        label: 'Starter'
      })
    ])

    state.registrationEmailSuffixWhitelistDraft.value = '@foo.bar'
    state.commitRegistrationEmailSuffixWhitelistDraft()
    state.addDefaultSubscription()
    await state.setAndCopyLinuxdoRedirectUrl()
    expect(copyToClipboard).toHaveBeenCalledWith(
      'https://sub2api.example.com/api/v1/auth/oauth/linuxdo/callback',
      'admin.settings.linuxdo.redirectUrlSetAndCopied'
    )

    await state.saveSettings()
    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        registration_email_suffix_whitelist: ['@example.com', '@foo.bar'],
        default_subscriptions: [{ group_id: 10, validity_days: 30 }]
      })
    )
    expect(refreshPublicSettings).toHaveBeenCalledWith(true)
    expect(refreshAdminSettings).toHaveBeenCalledWith(true)
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.settingsSaved')
    expect(showError).not.toHaveBeenCalled()
  })

  it('tests smtp and sends test email with manual password edits only', async () => {
    const showError = vi.fn()
    const state = useSettingsViewForm({
      t: (key: string) => key,
      showError,
      showSuccess: vi.fn(),
      refreshPublicSettings: vi.fn().mockResolvedValue(undefined),
      refreshAdminSettings: vi.fn().mockResolvedValue(undefined),
      copyToClipboard: vi.fn().mockResolvedValue(true)
    })

    await state.loadSettings()

    await state.testSmtpConnection()
    expect(testSmtpConnection).toHaveBeenCalledWith(
      expect.objectContaining({
        smtp_password: ''
      })
    )

    state.smtpPasswordManuallyEdited.value = true
    state.form.smtp_password = 'secret'
    state.testEmailAddress.value = ''
    await state.sendTestEmail()
    expect(showError).toHaveBeenCalledWith('admin.settings.testEmail.enterRecipientHint')

    state.testEmailAddress.value = 'ops@example.com'
    await state.sendTestEmail()
    expect(sendTestEmail).toHaveBeenCalledWith(
      expect.objectContaining({
        email: 'ops@example.com',
        smtp_password: 'secret'
      })
    )
  })

  it('keeps the latest settings load and subscription groups results', async () => {
    const firstSettings = createDeferred<SystemSettings>()
    const secondSettings = createDeferred<SystemSettings>()
    const firstGroups = createDeferred<AdminGroup[]>()
    const secondGroups = createDeferred<AdminGroup[]>()

    getSettings
      .mockReset()
      .mockReturnValueOnce(firstSettings.promise)
      .mockReturnValueOnce(secondSettings.promise)
    getAllGroups
      .mockReset()
      .mockReturnValueOnce(firstGroups.promise)
      .mockReturnValueOnce(secondGroups.promise)

    const state = useSettingsViewForm({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess: vi.fn(),
      refreshPublicSettings: vi.fn().mockResolvedValue(undefined),
      refreshAdminSettings: vi.fn().mockResolvedValue(undefined),
      copyToClipboard: vi.fn().mockResolvedValue(true)
    })

    const firstLoad = state.loadSettings()
    const secondLoad = state.loadSettings()
    const firstGroupsLoad = state.loadSubscriptionGroups()
    const secondGroupsLoad = state.loadSubscriptionGroups()

    secondSettings.resolve(createSystemSettings({ site_name: 'Latest Settings' }))
    await secondLoad

    firstSettings.resolve(createSystemSettings({ site_name: 'Stale Settings' }))
    await firstLoad

    secondGroups.resolve([createGroup({ id: 11, name: 'Latest Group' })])
    await secondGroupsLoad

    firstGroups.resolve([createGroup({ id: 12, name: 'Stale Group' })])
    await firstGroupsLoad

    expect(state.form.site_name).toBe('Latest Settings')
    expect(state.defaultSubscriptionGroupOptions.value).toEqual([
      expect.objectContaining({
        value: 11,
        label: 'Latest Group'
      })
    ])
  })

  it('does not let a stale settings load overwrite a newer save', async () => {
    const pendingLoad = createDeferred<SystemSettings>()
    const saveResponse = createDeferred<SystemSettings>()

    getSettings.mockReset().mockReturnValueOnce(pendingLoad.promise)
    updateSettings.mockReset().mockReturnValueOnce(saveResponse.promise)

    const showSuccess = vi.fn()
    const refreshPublicSettings = vi.fn().mockResolvedValue(undefined)
    const refreshAdminSettings = vi.fn().mockResolvedValue(undefined)
    const state = useSettingsViewForm({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess,
      refreshPublicSettings,
      refreshAdminSettings,
      copyToClipboard: vi.fn().mockResolvedValue(true)
    })

    const loadPromise = state.loadSettings()
    state.form.site_name = 'Saved Name'
    const savePromise = state.saveSettings()

    saveResponse.resolve(createSystemSettings({ site_name: 'Saved Name' }))
    await savePromise

    pendingLoad.resolve(createSystemSettings({ site_name: 'Stale Name' }))
    await loadPromise

    expect(state.form.site_name).toBe('Saved Name')
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.settingsSaved')
    expect(refreshPublicSettings).toHaveBeenCalledWith(true)
    expect(refreshAdminSettings).toHaveBeenCalledWith(true)
    expect(state.loading.value).toBe(false)
    expect(state.saving.value).toBe(false)
  })
})
