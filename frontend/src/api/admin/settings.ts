/**
 * Admin Settings API endpoints
 * Handles system settings management for administrators
 */

import { apiClient } from '../client'
import type { CustomMenuItem, CustomEndpoint, NotifyEmailEntry } from '@/types'

export interface DefaultSubscriptionSetting {
  group_id: number
  validity_days: number
}

export type AuthSourceType = 'email' | 'linuxdo' | 'oidc' | 'wechat'

export interface AuthSourceDefaultsValue {
  balance: number
  concurrency: number
  subscriptions: DefaultSubscriptionSetting[]
  grant_on_signup: boolean
  grant_on_first_bind: boolean
}

/**
 * System settings interface
 */
export interface SystemSettings {
  // Registration settings
  registration_enabled: boolean
  email_verify_enabled: boolean
  force_email_on_third_party_signup: boolean
  registration_email_suffix_whitelist: string[]
  promo_code_enabled: boolean
  password_reset_enabled: boolean
  frontend_url: string
  invitation_code_enabled: boolean
  totp_enabled: boolean // TOTP 双因素认证
  totp_encryption_key_configured: boolean // TOTP 加密密钥是否已配置
  // Default settings
  default_balance: number
  default_concurrency: number
  default_subscriptions: DefaultSubscriptionSetting[]
  auth_source_default_email_balance: number
  auth_source_default_email_concurrency: number
  auth_source_default_email_subscriptions: DefaultSubscriptionSetting[]
  auth_source_default_email_grant_on_signup: boolean
  auth_source_default_email_grant_on_first_bind: boolean
  auth_source_default_linuxdo_balance: number
  auth_source_default_linuxdo_concurrency: number
  auth_source_default_linuxdo_subscriptions: DefaultSubscriptionSetting[]
  auth_source_default_linuxdo_grant_on_signup: boolean
  auth_source_default_linuxdo_grant_on_first_bind: boolean
  auth_source_default_oidc_balance: number
  auth_source_default_oidc_concurrency: number
  auth_source_default_oidc_subscriptions: DefaultSubscriptionSetting[]
  auth_source_default_oidc_grant_on_signup: boolean
  auth_source_default_oidc_grant_on_first_bind: boolean
  auth_source_default_wechat_balance: number
  auth_source_default_wechat_concurrency: number
  auth_source_default_wechat_subscriptions: DefaultSubscriptionSetting[]
  auth_source_default_wechat_grant_on_signup: boolean
  auth_source_default_wechat_grant_on_first_bind: boolean
  // OEM settings
  site_name: string
  site_logo: string
  site_subtitle: string
  frontend_theme?: string
  api_base_url: string
  contact_info: string
  doc_url: string
  home_content: string
  hide_ccs_import_button: boolean
  purchase_subscription_enabled: boolean
  purchase_subscription_url: string
  payment_enabled?: boolean
  backend_mode_enabled: boolean
  custom_menu_items: CustomMenuItem[]
  custom_endpoints: CustomEndpoint[]
  // SMTP settings
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password_configured: boolean
  smtp_from_email: string
  smtp_from_name: string
  smtp_use_tls: boolean
  balance_low_notify_enabled: boolean
  balance_low_notify_threshold: number
  balance_low_notify_recharge_url: string
  account_quota_notify_enabled: boolean
  account_quota_notify_emails: NotifyEmailEntry[]
  // Cloudflare Turnstile settings
  turnstile_enabled: boolean
  turnstile_site_key: string
  turnstile_secret_key_configured: boolean

  // LinuxDo Connect OAuth settings
  linuxdo_connect_enabled: boolean
  linuxdo_connect_client_id: string
  linuxdo_connect_client_secret_configured: boolean
  linuxdo_connect_redirect_url: string

  // WeChat Connect OAuth settings
  wechat_connect_enabled: boolean
  wechat_connect_open_app_id: string
  wechat_connect_open_app_secret_configured: boolean
  wechat_connect_mp_app_id: string
  wechat_connect_mp_app_secret_configured: boolean
  wechat_connect_mobile_app_id: string
  wechat_connect_mobile_app_secret_configured: boolean
  wechat_connect_open_enabled: boolean
  wechat_connect_mp_enabled: boolean
  wechat_connect_mobile_enabled: boolean
  wechat_connect_mode: string
  wechat_connect_scopes: string
  wechat_connect_redirect_url: string
  wechat_connect_frontend_redirect_url: string

  // OIDC Connect OAuth settings
  oidc_connect_enabled: boolean
  oidc_connect_provider_name: string
  oidc_connect_client_id: string
  oidc_connect_client_secret_configured: boolean
  oidc_connect_issuer_url: string
  oidc_connect_discovery_url: string
  oidc_connect_authorize_url: string
  oidc_connect_token_url: string
  oidc_connect_userinfo_url: string
  oidc_connect_jwks_url: string
  oidc_connect_scopes: string
  oidc_connect_redirect_url: string
  oidc_connect_frontend_redirect_url: string
  oidc_connect_token_auth_method: string
  oidc_connect_use_pkce: boolean
  oidc_connect_validate_id_token: boolean
  oidc_connect_allowed_signing_algs: string
  oidc_connect_clock_skew_seconds: number
  oidc_connect_require_email_verified: boolean
  oidc_connect_userinfo_email_path: string
  oidc_connect_userinfo_id_path: string
  oidc_connect_userinfo_username_path: string

  // Model fallback configuration
  enable_model_fallback: boolean
  fallback_model_anthropic: string
  fallback_model_openai: string
  fallback_model_grok: string
  fallback_model_gemini: string
  fallback_model_antigravity: string
  grok_official_base_url: string
  grok_session_base_url: string
  grok_thinking_summary: boolean
  grok_show_search_sources: boolean
  grok_image_output_format: string
  grok_video_output_format: string
  grok_media_proxy_enabled: boolean
  grok_media_cache_retention_hours: number
  grok_quota_sync_interval_seconds: number
  grok_usage_sync_concurrency: number
  grok_capability_probe_interval_seconds: number
  grok_capability_probe_concurrency: number
  grok_session_validity_check_interval: number
  grok_video_timeout: number

  // Identity patch configuration (Claude -> Gemini)
  enable_identity_patch: boolean
  identity_patch_prompt: string

  // Ops Monitoring (vNext)
  ops_monitoring_enabled: boolean
  ops_realtime_monitoring_enabled: boolean
  ops_query_mode_default: 'auto' | 'raw' | 'preagg' | string
  ops_metrics_interval_seconds: number

  // Claude Code version check
  min_claude_code_version: string
  max_claude_code_version: string

  // 分组隔离
  allow_ungrouped_key_scheduling: boolean

  // Gateway forwarding behavior
  enable_fingerprint_unification: boolean
  enable_metadata_passthrough: boolean
  enable_cch_signing: boolean
}

export interface UpdateSettingsRequest {
  registration_enabled?: boolean
  email_verify_enabled?: boolean
  force_email_on_third_party_signup?: boolean
  registration_email_suffix_whitelist?: string[]
  promo_code_enabled?: boolean
  password_reset_enabled?: boolean
  frontend_url?: string
  invitation_code_enabled?: boolean
  totp_enabled?: boolean // TOTP 双因素认证
  default_balance?: number
  default_concurrency?: number
  default_subscriptions?: DefaultSubscriptionSetting[]
  auth_source_default_email_balance?: number
  auth_source_default_email_concurrency?: number
  auth_source_default_email_subscriptions?: DefaultSubscriptionSetting[]
  auth_source_default_email_grant_on_signup?: boolean
  auth_source_default_email_grant_on_first_bind?: boolean
  auth_source_default_linuxdo_balance?: number
  auth_source_default_linuxdo_concurrency?: number
  auth_source_default_linuxdo_subscriptions?: DefaultSubscriptionSetting[]
  auth_source_default_linuxdo_grant_on_signup?: boolean
  auth_source_default_linuxdo_grant_on_first_bind?: boolean
  auth_source_default_oidc_balance?: number
  auth_source_default_oidc_concurrency?: number
  auth_source_default_oidc_subscriptions?: DefaultSubscriptionSetting[]
  auth_source_default_oidc_grant_on_signup?: boolean
  auth_source_default_oidc_grant_on_first_bind?: boolean
  auth_source_default_wechat_balance?: number
  auth_source_default_wechat_concurrency?: number
  auth_source_default_wechat_subscriptions?: DefaultSubscriptionSetting[]
  auth_source_default_wechat_grant_on_signup?: boolean
  auth_source_default_wechat_grant_on_first_bind?: boolean
  site_name?: string
  site_logo?: string
  site_subtitle?: string
  frontend_theme?: string
  api_base_url?: string
  contact_info?: string
  doc_url?: string
  home_content?: string
  hide_ccs_import_button?: boolean
  purchase_subscription_enabled?: boolean
  purchase_subscription_url?: string
  backend_mode_enabled?: boolean
  custom_menu_items?: CustomMenuItem[]
  custom_endpoints?: CustomEndpoint[]
  smtp_host?: string
  smtp_port?: number
  smtp_username?: string
  smtp_password?: string
  smtp_from_email?: string
  smtp_from_name?: string
  smtp_use_tls?: boolean
  balance_low_notify_enabled?: boolean
  balance_low_notify_threshold?: number
  balance_low_notify_recharge_url?: string
  account_quota_notify_enabled?: boolean
  account_quota_notify_emails?: NotifyEmailEntry[]
  turnstile_enabled?: boolean
  turnstile_site_key?: string
  turnstile_secret_key?: string
  linuxdo_connect_enabled?: boolean
  linuxdo_connect_client_id?: string
  linuxdo_connect_client_secret?: string
  linuxdo_connect_redirect_url?: string
  wechat_connect_enabled?: boolean
  wechat_connect_open_app_id?: string
  wechat_connect_open_app_secret?: string
  wechat_connect_mp_app_id?: string
  wechat_connect_mp_app_secret?: string
  wechat_connect_mobile_app_id?: string
  wechat_connect_mobile_app_secret?: string
  wechat_connect_open_enabled?: boolean
  wechat_connect_mp_enabled?: boolean
  wechat_connect_mobile_enabled?: boolean
  wechat_connect_mode?: string
  wechat_connect_scopes?: string
  wechat_connect_redirect_url?: string
  wechat_connect_frontend_redirect_url?: string
  oidc_connect_enabled?: boolean
  oidc_connect_provider_name?: string
  oidc_connect_client_id?: string
  oidc_connect_client_secret?: string
  oidc_connect_issuer_url?: string
  oidc_connect_discovery_url?: string
  oidc_connect_authorize_url?: string
  oidc_connect_token_url?: string
  oidc_connect_userinfo_url?: string
  oidc_connect_jwks_url?: string
  oidc_connect_scopes?: string
  oidc_connect_redirect_url?: string
  oidc_connect_frontend_redirect_url?: string
  oidc_connect_token_auth_method?: string
  oidc_connect_use_pkce?: boolean
  oidc_connect_validate_id_token?: boolean
  oidc_connect_allowed_signing_algs?: string
  oidc_connect_clock_skew_seconds?: number
  oidc_connect_require_email_verified?: boolean
  oidc_connect_userinfo_email_path?: string
  oidc_connect_userinfo_id_path?: string
  oidc_connect_userinfo_username_path?: string
  enable_model_fallback?: boolean
  fallback_model_anthropic?: string
  fallback_model_openai?: string
  fallback_model_grok?: string
  fallback_model_gemini?: string
  fallback_model_antigravity?: string
  grok_official_base_url?: string
  grok_session_base_url?: string
  grok_thinking_summary?: boolean
  grok_show_search_sources?: boolean
  grok_image_output_format?: string
  grok_video_output_format?: string
  grok_media_proxy_enabled?: boolean
  grok_media_cache_retention_hours?: number
  grok_quota_sync_interval_seconds?: number
  grok_usage_sync_concurrency?: number
  grok_capability_probe_interval_seconds?: number
  grok_capability_probe_concurrency?: number
  grok_session_validity_check_interval?: number
  grok_video_timeout?: number
  enable_identity_patch?: boolean
  identity_patch_prompt?: string
  ops_monitoring_enabled?: boolean
  ops_realtime_monitoring_enabled?: boolean
  ops_query_mode_default?: 'auto' | 'raw' | 'preagg' | string
  ops_metrics_interval_seconds?: number
  min_claude_code_version?: string
  max_claude_code_version?: string
  allow_ungrouped_key_scheduling?: boolean
  enable_fingerprint_unification?: boolean
  enable_metadata_passthrough?: boolean
  enable_cch_signing?: boolean
}

/**
 * Get all system settings
 * @returns System settings
 */
export async function getSettings(): Promise<SystemSettings> {
  const { data } = await apiClient.get<SystemSettings>('/admin/settings')
  return data
}

/**
 * Update system settings
 * @param settings - Partial settings to update
 * @returns Updated settings
 */
export async function updateSettings(settings: UpdateSettingsRequest): Promise<SystemSettings> {
  const { data } = await apiClient.put<SystemSettings>('/admin/settings', settings)
  return data
}

/**
 * Test SMTP connection request
 */
export interface TestSmtpRequest {
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
  smtp_use_tls: boolean
}

/**
 * Test SMTP connection with provided config
 * @param config - SMTP configuration to test
 * @returns Test result message
 */
export async function testSmtpConnection(config: TestSmtpRequest): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/admin/settings/test-smtp', config)
  return data
}

/**
 * Send test email request
 */
export interface SendTestEmailRequest {
  email: string
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
  smtp_from_email: string
  smtp_from_name: string
  smtp_use_tls: boolean
}

/**
 * Send test email with provided SMTP config
 * @param request - Email address and SMTP config
 * @returns Test result message
 */
export async function sendTestEmail(request: SendTestEmailRequest): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(
    '/admin/settings/send-test-email',
    request
  )
  return data
}

/**
 * Admin API Key status response
 */
export interface AdminApiKeyStatus {
  exists: boolean
  masked_key: string
}

/**
 * Get admin API key status
 * @returns Status indicating if key exists and masked version
 */
export async function getAdminApiKey(): Promise<AdminApiKeyStatus> {
  const { data } = await apiClient.get<AdminApiKeyStatus>('/admin/settings/admin-api-key')
  return data
}

/**
 * Regenerate admin API key
 * @returns The new full API key (only shown once)
 */
export async function regenerateAdminApiKey(): Promise<{ key: string }> {
  const { data } = await apiClient.post<{ key: string }>('/admin/settings/admin-api-key/regenerate')
  return data
}

/**
 * Delete admin API key
 * @returns Success message
 */
export async function deleteAdminApiKey(): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>('/admin/settings/admin-api-key')
  return data
}

// ==================== Overload Cooldown Settings ====================

/**
 * Overload cooldown settings interface (529 handling)
 */
export interface OverloadCooldownSettings {
  enabled: boolean
  cooldown_minutes: number
}

export async function getOverloadCooldownSettings(): Promise<OverloadCooldownSettings> {
  const { data } = await apiClient.get<OverloadCooldownSettings>('/admin/settings/overload-cooldown')
  return data
}

export async function updateOverloadCooldownSettings(
  settings: OverloadCooldownSettings
): Promise<OverloadCooldownSettings> {
  const { data } = await apiClient.put<OverloadCooldownSettings>(
    '/admin/settings/overload-cooldown',
    settings
  )
  return data
}

// ==================== Stream Timeout Settings ====================

/**
 * Stream timeout settings interface
 */
export interface StreamTimeoutSettings {
  enabled: boolean
  action: 'temp_unsched' | 'error' | 'none'
  temp_unsched_minutes: number
  threshold_count: number
  threshold_window_minutes: number
}

/**
 * Get stream timeout settings
 * @returns Stream timeout settings
 */
export async function getStreamTimeoutSettings(): Promise<StreamTimeoutSettings> {
  const { data } = await apiClient.get<StreamTimeoutSettings>('/admin/settings/stream-timeout')
  return data
}

/**
 * Update stream timeout settings
 * @param settings - Stream timeout settings to update
 * @returns Updated settings
 */
export async function updateStreamTimeoutSettings(
  settings: StreamTimeoutSettings
): Promise<StreamTimeoutSettings> {
  const { data } = await apiClient.put<StreamTimeoutSettings>(
    '/admin/settings/stream-timeout',
    settings
  )
  return data
}

// ==================== Rectifier Settings ====================

/**
 * Rectifier settings interface
 */
export interface RectifierSettings {
  enabled: boolean
  thinking_signature_enabled: boolean
  thinking_budget_enabled: boolean
  apikey_signature_enabled: boolean
  apikey_signature_patterns: string[]
}

/**
 * Get rectifier settings
 * @returns Rectifier settings
 */
export async function getRectifierSettings(): Promise<RectifierSettings> {
  const { data } = await apiClient.get<RectifierSettings>('/admin/settings/rectifier')
  return data
}

/**
 * Update rectifier settings
 * @param settings - Rectifier settings to update
 * @returns Updated settings
 */
export async function updateRectifierSettings(
  settings: RectifierSettings
): Promise<RectifierSettings> {
  const { data } = await apiClient.put<RectifierSettings>(
    '/admin/settings/rectifier',
    settings
  )
  return data
}

// ==================== Beta Policy Settings ====================

/**
 * Beta policy rule interface
 */
export interface BetaPolicyRule {
  beta_token: string
  action: 'pass' | 'filter' | 'block'
  scope: 'all' | 'oauth' | 'apikey' | 'bedrock'
  error_message?: string
  model_whitelist?: string[]
  fallback_action?: 'pass' | 'filter' | 'block'
  fallback_error_message?: string
}

/**
 * Beta policy settings interface
 */
export interface BetaPolicySettings {
  rules: BetaPolicyRule[]
}

/**
 * Get beta policy settings
 * @returns Beta policy settings
 */
export async function getBetaPolicySettings(): Promise<BetaPolicySettings> {
  const { data } = await apiClient.get<BetaPolicySettings>('/admin/settings/beta-policy')
  return data
}

/**
 * Update beta policy settings
 * @param settings - Beta policy settings to update
 * @returns Updated settings
 */
export async function updateBetaPolicySettings(
  settings: BetaPolicySettings
): Promise<BetaPolicySettings> {
  const { data } = await apiClient.put<BetaPolicySettings>(
    '/admin/settings/beta-policy',
    settings
  )
  return data
}

export const settingsAPI = {
  getSettings,
  updateSettings,
  testSmtpConnection,
  sendTestEmail,
  getAdminApiKey,
  regenerateAdminApiKey,
  deleteAdminApiKey,
  getOverloadCooldownSettings,
  updateOverloadCooldownSettings,
  getStreamTimeoutSettings,
  updateStreamTimeoutSettings,
  getRectifierSettings,
  updateRectifierSettings,
  getBetaPolicySettings,
  updateBetaPolicySettings
}

export default settingsAPI
