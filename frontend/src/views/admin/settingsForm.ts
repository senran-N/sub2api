import type {
  DefaultSubscriptionSetting,
  SystemSettings,
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

  form.frontend_url = normalizeOptionalHttpUrl(form.frontend_url)
  form.doc_url = normalizeOptionalHttpUrl(form.doc_url)

  if (form.purchase_subscription_enabled) {
    if (!form.purchase_subscription_url) {
      return {
        ok: false,
        error: {
          code: 'purchase_url_required'
        }
      }
    }
    if (!isValidHttpUrl(form.purchase_subscription_url)) {
      return {
        ok: false,
        error: {
          code: 'purchase_url_invalid'
        }
      }
    }
  } else {
    form.purchase_subscription_url = normalizeOptionalHttpUrl(form.purchase_subscription_url)
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
      api_base_url: form.api_base_url,
      contact_info: form.contact_info,
      doc_url: form.doc_url,
      home_content: form.home_content,
      backend_mode_enabled: form.backend_mode_enabled,
      hide_ccs_import_button: form.hide_ccs_import_button,
      purchase_subscription_enabled: form.purchase_subscription_enabled,
      purchase_subscription_url: form.purchase_subscription_url,
      sora_client_enabled: form.sora_client_enabled,
      custom_menu_items: form.custom_menu_items,
      custom_endpoints: form.custom_endpoints,
      frontend_url: form.frontend_url,
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
      enable_metadata_passthrough: form.enable_metadata_passthrough
    }
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
