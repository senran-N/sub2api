import type {
  BetaPolicyRule,
  DefaultSubscriptionSetting,
  OverloadCooldownSettings,
  RectifierSettings,
  SendTestEmailRequest,
  SystemSettings,
  StreamTimeoutSettings,
  TestSmtpRequest,
  UpdateSettingsRequest
} from '@/api/admin/settings'
import type { AdminGroup, CustomEndpoint, CustomMenuItem } from '@/types'
import {
  normalizeRegistrationEmailSuffixDomains,
  normalizeRegistrationEmailSuffixWhitelist
} from '@/utils/registrationEmailPolicy'

export type SettingsForm = SystemSettings & {
  smtp_password: string
  turnstile_secret_key: string
  linuxdo_connect_client_secret: string
}

export type SettingsRegistrationFields = Pick<
  SettingsForm,
  | 'registration_enabled'
  | 'email_verify_enabled'
  | 'promo_code_enabled'
  | 'invitation_code_enabled'
  | 'password_reset_enabled'
  | 'frontend_url'
  | 'totp_enabled'
  | 'totp_encryption_key_configured'
>

export type SettingsSmtpFields = Pick<
  SettingsForm,
  | 'smtp_host'
  | 'smtp_port'
  | 'smtp_username'
  | 'smtp_password'
  | 'smtp_password_configured'
  | 'smtp_from_email'
  | 'smtp_from_name'
  | 'smtp_use_tls'
>

export type SettingsTurnstileFields = Pick<
  SettingsForm,
  | 'turnstile_enabled'
  | 'turnstile_site_key'
  | 'turnstile_secret_key'
  | 'turnstile_secret_key_configured'
>

export type SettingsDefaultsFields = Pick<
  SettingsForm,
  'default_balance' | 'default_concurrency' | 'default_subscriptions'
>

export type SettingsPurchaseFields = Pick<
  SettingsForm,
  'purchase_subscription_enabled' | 'purchase_subscription_url'
>

export type SettingsLinuxdoFields = Pick<
  SettingsForm,
  | 'linuxdo_connect_enabled'
  | 'linuxdo_connect_client_id'
  | 'linuxdo_connect_client_secret'
  | 'linuxdo_connect_client_secret_configured'
  | 'linuxdo_connect_redirect_url'
>

export type SettingsClaudeCodeFields = Pick<
  SettingsForm,
  'min_claude_code_version' | 'max_claude_code_version'
>

export type SettingsSchedulingFields = Pick<
  SettingsForm,
  'allow_ungrouped_key_scheduling'
>

export type SettingsCustomMenuFields = Pick<SettingsForm, 'custom_menu_items'>

export type SettingsGatewayForwardingFields = Pick<
  SettingsForm,
  'enable_fingerprint_unification' | 'enable_metadata_passthrough' | 'enable_cch_signing'
>

export type SettingsSiteFields = Pick<
  SettingsForm,
  | 'backend_mode_enabled'
  | 'site_name'
  | 'site_subtitle'
  | 'frontend_theme'
  | 'api_base_url'
  | 'custom_endpoints'
  | 'contact_info'
  | 'doc_url'
  | 'site_logo'
  | 'home_content'
  | 'hide_ccs_import_button'
>

export type SettingsPayloadError =
  | {
      code: 'duplicate_default_subscription'
      groupId: number
    }
  | {
      code: 'purchase_url_required'
    }
  | {
      code: 'purchase_url_invalid'
    }

export type SettingsPayloadResult =
  | {
      ok: true
      payload: UpdateSettingsRequest
    }
  | {
      ok: false
      error: SettingsPayloadError
    }

type WritableSecretFields = Pick<
  SettingsForm,
  'smtp_password' | 'turnstile_secret_key' | 'linuxdo_connect_client_secret'
>

interface NormalizedSettingsUrls {
  frontendUrl: string
  docUrl: string
  purchaseSubscriptionUrl: string
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
    frontend_theme: 'factory',
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
    enable_metadata_passthrough: false,
    enable_cch_signing: false
  }
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

export function addCustomMenuItem(items: CustomMenuItem[]): void {
  items.push({
    id: '',
    label: '',
    icon_svg: '',
    url: '',
    visibility: 'user',
    sort_order: items.length
  })
}

export function removeCustomMenuItem(items: CustomMenuItem[], index: number): void {
  items.splice(index, 1)
  reindexCustomMenuItems(items)
}

export function moveCustomMenuItem(
  items: CustomMenuItem[],
  index: number,
  direction: -1 | 1
): void {
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= items.length) {
    return
  }

  const current = items[index]
  items[index] = items[targetIndex]
  items[targetIndex] = current
  reindexCustomMenuItems(items)
}

export function addCustomEndpoint(items: CustomEndpoint[]): void {
  items.push({
    name: '',
    endpoint: '',
    description: ''
  })
}

export function removeCustomEndpoint(items: CustomEndpoint[], index: number): void {
  items.splice(index, 1)
}

export function addNextDefaultSubscription(
  items: DefaultSubscriptionSetting[],
  groups: Pick<AdminGroup, 'id'>[]
): void {
  if (groups.length === 0) {
    return
  }

  const existing = new Set(items.map((item) => item.group_id))
  const candidate = groups.find((group) => !existing.has(group.id))
  if (!candidate) {
    return
  }

  items.push({
    group_id: candidate.id,
    validity_days: 30
  })
}

export function removeDefaultSubscription(
  items: DefaultSubscriptionSetting[],
  index: number
): void {
  items.splice(index, 1)
}

export function hydrateSettingsForm(form: SettingsForm, settings: SystemSettings): string[] {
  Object.assign(form, settings)
  form.default_subscriptions = normalizeDefaultSubscriptions(settings.default_subscriptions)
  resetTransientSecrets(form)
  return normalizeRegistrationEmailSuffixDomains(settings.registration_email_suffix_whitelist)
}

export function buildSettingsUpdatePayload(
  form: SettingsForm,
  registrationEmailSuffixWhitelistTags: string[]
): SettingsPayloadResult {
  const defaultSubscriptions = normalizeDefaultSubscriptions(form.default_subscriptions)
  const duplicateDefaultSubscription = findDuplicateDefaultSubscription(defaultSubscriptions)
  if (duplicateDefaultSubscription) {
    return {
      ok: false,
      error: {
        code: 'duplicate_default_subscription',
        groupId: duplicateDefaultSubscription.group_id
      }
    }
  }

  const normalizedUrls = normalizeSettingsUrlsForPayload(form)
  const purchaseValidationError = validatePurchaseUrlForPayload(form, normalizedUrls)
  if (purchaseValidationError) {
    return {
      ok: false,
      error: purchaseValidationError
    }
  }

  return {
    ok: true,
    payload: {
      registration_enabled: form.registration_enabled,
      email_verify_enabled: form.email_verify_enabled,
      registration_email_suffix_whitelist: normalizeRegistrationEmailSuffixWhitelist(
        registrationEmailSuffixWhitelistTags
      ),
      promo_code_enabled: form.promo_code_enabled,
      invitation_code_enabled: form.invitation_code_enabled,
      password_reset_enabled: form.password_reset_enabled,
      totp_enabled: form.totp_enabled,
      default_balance: form.default_balance,
      default_concurrency: form.default_concurrency,
      default_subscriptions: defaultSubscriptions,
      site_name: form.site_name,
      site_logo: form.site_logo,
      site_subtitle: form.site_subtitle,
      frontend_theme: form.frontend_theme,
      api_base_url: form.api_base_url,
      contact_info: form.contact_info,
      doc_url: normalizedUrls.docUrl,
      home_content: form.home_content,
      backend_mode_enabled: form.backend_mode_enabled,
      hide_ccs_import_button: form.hide_ccs_import_button,
      purchase_subscription_enabled: form.purchase_subscription_enabled,
      purchase_subscription_url: normalizedUrls.purchaseSubscriptionUrl,
      custom_menu_items: form.custom_menu_items,
      custom_endpoints: form.custom_endpoints,
      frontend_url: normalizedUrls.frontendUrl,
      smtp_host: form.smtp_host,
      smtp_port: form.smtp_port,
      smtp_username: form.smtp_username,
      smtp_password: form.smtp_password || undefined,
      smtp_from_email: form.smtp_from_email,
      smtp_from_name: form.smtp_from_name,
      smtp_use_tls: form.smtp_use_tls,
      turnstile_enabled: form.turnstile_enabled,
      turnstile_site_key: form.turnstile_site_key,
      turnstile_secret_key: form.turnstile_secret_key || undefined,
      linuxdo_connect_enabled: form.linuxdo_connect_enabled,
      linuxdo_connect_client_id: form.linuxdo_connect_client_id,
      linuxdo_connect_client_secret: form.linuxdo_connect_client_secret || undefined,
      linuxdo_connect_redirect_url: form.linuxdo_connect_redirect_url,
      enable_model_fallback: form.enable_model_fallback,
      fallback_model_anthropic: form.fallback_model_anthropic,
      fallback_model_openai: form.fallback_model_openai,
      fallback_model_gemini: form.fallback_model_gemini,
      fallback_model_antigravity: form.fallback_model_antigravity,
      enable_identity_patch: form.enable_identity_patch,
      identity_patch_prompt: form.identity_patch_prompt,
      min_claude_code_version: form.min_claude_code_version,
      max_claude_code_version: form.max_claude_code_version,
      allow_ungrouped_key_scheduling: form.allow_ungrouped_key_scheduling,
      enable_fingerprint_unification: form.enable_fingerprint_unification,
      enable_metadata_passthrough: form.enable_metadata_passthrough,
      enable_cch_signing: form.enable_cch_signing
    }
  }
}

export function buildSmtpTestConnectionRequest(
  form: SettingsForm,
  smtpPasswordManuallyEdited: boolean
): TestSmtpRequest {
  return {
    smtp_host: form.smtp_host,
    smtp_port: form.smtp_port,
    smtp_username: form.smtp_username,
    smtp_password: resolveSettingsSmtpPassword(form, smtpPasswordManuallyEdited),
    smtp_use_tls: form.smtp_use_tls
  }
}

export function buildSendTestEmailRequest(
  form: SettingsForm,
  email: string,
  smtpPasswordManuallyEdited: boolean
): SendTestEmailRequest {
  return {
    email,
    smtp_host: form.smtp_host,
    smtp_port: form.smtp_port,
    smtp_username: form.smtp_username,
    smtp_password: resolveSettingsSmtpPassword(form, smtpPasswordManuallyEdited),
    smtp_from_email: form.smtp_from_email,
    smtp_from_name: form.smtp_from_name,
    smtp_use_tls: form.smtp_use_tls
  }
}

export function normalizeDefaultSubscriptions(
  items: DefaultSubscriptionSetting[] | null | undefined
): DefaultSubscriptionSetting[] {
  if (!Array.isArray(items)) {
    return []
  }

  return items
    .filter((item) => item.group_id > 0 && item.validity_days > 0)
    .map((item) => ({
      group_id: item.group_id,
      validity_days: Math.min(36500, Math.max(1, Math.floor(item.validity_days)))
    }))
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

function reindexCustomMenuItems(items: CustomMenuItem[]): void {
  items.forEach((item, index) => {
    item.sort_order = index
  })
}

function resetTransientSecrets(form: WritableSecretFields): void {
  form.smtp_password = ''
  form.turnstile_secret_key = ''
  form.linuxdo_connect_client_secret = ''
}

function findDuplicateDefaultSubscription(
  items: DefaultSubscriptionSetting[]
): DefaultSubscriptionSetting | null {
  const seenGroupIDs = new Set<number>()
  for (const item of items) {
    if (seenGroupIDs.has(item.group_id)) {
      return item
    }
    seenGroupIDs.add(item.group_id)
  }
  return null
}

function normalizeSettingsUrlsForPayload(form: SettingsForm): NormalizedSettingsUrls {
  const frontendUrl = normalizeOptionalHttpUrl(form.frontend_url)
  const docUrl = normalizeOptionalHttpUrl(form.doc_url)

  if (form.purchase_subscription_enabled) {
    return {
      frontendUrl,
      docUrl,
      purchaseSubscriptionUrl: form.purchase_subscription_url
    }
  }

  return {
    frontendUrl,
    docUrl,
    purchaseSubscriptionUrl: normalizeOptionalHttpUrl(form.purchase_subscription_url)
  }
}

function validatePurchaseUrlForPayload(
  form: SettingsForm,
  urls: NormalizedSettingsUrls
): SettingsPayloadError | null {
  if (!form.purchase_subscription_enabled) {
    return null
  }

  if (!urls.purchaseSubscriptionUrl) {
    return {
      code: 'purchase_url_required'
    }
  }

  if (!isValidHttpUrl(urls.purchaseSubscriptionUrl)) {
    return {
      code: 'purchase_url_invalid'
    }
  }

  return null
}

function resolveSettingsSmtpPassword(
  form: Pick<SettingsForm, 'smtp_password'>,
  smtpPasswordManuallyEdited: boolean
): string {
  return smtpPasswordManuallyEdited ? form.smtp_password : ''
}

function normalizeOptionalHttpUrl(url: string): string {
  return isValidHttpUrl(url) ? url : ''
}

function isValidHttpUrl(url: string): boolean {
  if (!url) {
    return true
  }

  try {
    const target = new URL(url)
    return target.protocol === 'http:' || target.protocol === 'https:'
  } catch {
    return false
  }
}
