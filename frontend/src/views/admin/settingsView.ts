import type {
  BetaPolicyRule,
  OverloadCooldownSettings,
  RectifierSettings,
  StreamTimeoutSettings
} from '@/api/admin/settings'
import type { SettingsForm } from './settingsForm'

const SETTINGS_BETA_DISPLAY_NAMES: Record<string, string> = {
  'fast-mode-2026-02-01': 'Fast Mode',
  'context-1m-2025-08-07': 'Context 1M'
}

export function createDefaultSettingsForm(): SettingsForm {
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
    sora_client_enabled: false,
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
    fallback_model_gemini: 'gemini-2.5-pro',
    fallback_model_antigravity: 'gemini-2.5-pro',
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
    enable_metadata_passthrough: false
  }
}

export function createDefaultOverloadCooldownSettings(): OverloadCooldownSettings {
  return {
    enabled: true,
    cooldown_minutes: 10
  }
}

export function createDefaultStreamTimeoutSettings(): StreamTimeoutSettings {
  return {
    enabled: true,
    action: 'temp_unsched',
    temp_unsched_minutes: 5,
    threshold_count: 3,
    threshold_window_minutes: 10
  }
}

export function createDefaultRectifierSettings(): RectifierSettings {
  return {
    enabled: true,
    thinking_signature_enabled: true,
    thinking_budget_enabled: true,
    apikey_signature_enabled: false,
    apikey_signature_patterns: []
  }
}

export function createDefaultBetaPolicyRules(): BetaPolicyRule[] {
  return []
}

export function sanitizeRectifierPatterns(
  patterns: string[] | null | undefined
): string[] {
  if (!Array.isArray(patterns)) {
    return []
  }

  return patterns
    .map((pattern) => pattern.trim())
    .filter((pattern) => pattern.length > 0)
}

export function maskSettingsApiKey(key: string): string {
  return `${key.substring(0, 10)}...${key.slice(-4)}`
}

export function getSettingsLinuxdoRedirectUrlSuggestion(
  location:
    | Pick<Location, 'origin' | 'protocol' | 'host'>
    | null
    | undefined
): string {
  if (!location) {
    return ''
  }

  const origin = location.origin || `${location.protocol}//${location.host}`
  return `${origin}/api/v1/auth/oauth/linuxdo/callback`
}

export function getSettingsBetaDisplayName(token: string): string {
  return SETTINGS_BETA_DISPLAY_NAMES[token] || token
}
