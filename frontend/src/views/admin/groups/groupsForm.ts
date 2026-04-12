import type {
  AdminGroup,
  CreateGroupRequest,
  GroupPlatform,
  SelectOption,
  SubscriptionType,
  UpdateGroupRequest
} from '@/types'

export const DEFAULT_SUPPORTED_MODEL_SCOPES = ['claude', 'gemini_text', 'gemini_image'] as const
export const ACCOUNT_FILTER_PLATFORMS: readonly GroupPlatform[] = [
  'openai',
  'antigravity',
  'anthropic'
]

export interface SimpleAccount {
  id: number
  name: string
}

export interface ModelRoutingRule {
  pattern: string
  accounts: SimpleAccount[]
}

export interface NullableNumberSelectOption extends SelectOption {
  value: number | null
}

export interface NumberSelectOption extends SelectOption {
  value: number
}

interface GroupBaseForm {
  name: string
  description: string
  platform: GroupPlatform
  rate_multiplier: number
  is_exclusive: boolean
  subscription_type: SubscriptionType
  daily_limit_usd: number | null
  weekly_limit_usd: number | null
  monthly_limit_usd: number | null
  image_price_1k: number | null
  image_price_2k: number | null
  image_price_4k: number | null
  claude_code_only: boolean
  fallback_group_id: number | null
  fallback_group_id_on_invalid_request: number | null
  allow_messages_dispatch: boolean
  default_mapped_model: string
  require_oauth_only: boolean
  require_privacy_set: boolean
  model_routing_enabled: boolean
  supported_model_scopes: string[]
  mcp_xml_inject: boolean
  copy_accounts_from_group_ids: number[]
}

export interface CreateGroupForm extends GroupBaseForm {}

export interface EditGroupForm extends GroupBaseForm {
  status: AdminGroup['status']
}

export type GroupDialogForm = CreateGroupForm | EditGroupForm

export function createDefaultCreateGroupForm(): CreateGroupForm {
  return {
    name: '',
    description: '',
    platform: 'anthropic',
    rate_multiplier: 1.0,
    is_exclusive: false,
    subscription_type: 'standard',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    allow_messages_dispatch: false,
    default_mapped_model: 'gpt-5.4',
    require_oauth_only: false,
    require_privacy_set: false,
    model_routing_enabled: false,
    supported_model_scopes: [...DEFAULT_SUPPORTED_MODEL_SCOPES],
    mcp_xml_inject: true,
    copy_accounts_from_group_ids: []
  }
}

export function createDefaultEditGroupForm(): EditGroupForm {
  return {
    ...createDefaultCreateGroupForm(),
    default_mapped_model: '',
    status: 'active'
  }
}

export function resetCreateGroupForm(form: CreateGroupForm): void {
  Object.assign(form, createDefaultCreateGroupForm())
}

export function resetEditGroupForm(form: EditGroupForm): void {
  Object.assign(form, createDefaultEditGroupForm())
}

export function hydrateEditGroupForm(form: EditGroupForm, group: AdminGroup): void {
  Object.assign(form, createDefaultEditGroupForm(), {
    name: group.name,
    description: group.description || '',
    platform: group.platform,
    rate_multiplier: group.rate_multiplier,
    is_exclusive: group.is_exclusive,
    status: group.status,
    subscription_type: group.subscription_type || 'standard',
    daily_limit_usd: group.daily_limit_usd,
    weekly_limit_usd: group.weekly_limit_usd,
    monthly_limit_usd: group.monthly_limit_usd,
    image_price_1k: group.image_price_1k,
    image_price_2k: group.image_price_2k,
    image_price_4k: group.image_price_4k,
    claude_code_only: group.claude_code_only || false,
    fallback_group_id: group.fallback_group_id,
    fallback_group_id_on_invalid_request: group.fallback_group_id_on_invalid_request,
    allow_messages_dispatch: group.allow_messages_dispatch || false,
    default_mapped_model: group.default_mapped_model || '',
    require_oauth_only: group.require_oauth_only ?? false,
    require_privacy_set: group.require_privacy_set ?? false,
    model_routing_enabled: group.model_routing_enabled || false,
    supported_model_scopes:
      group.supported_model_scopes && group.supported_model_scopes.length > 0
        ? [...group.supported_model_scopes]
        : [...DEFAULT_SUPPORTED_MODEL_SCOPES],
    mcp_xml_inject: group.mcp_xml_inject ?? true,
    copy_accounts_from_group_ids: []
  })
}

export function toggleModelScope(scopes: string[], scope: string): void {
  const index = scopes.indexOf(scope)
  if (index === -1) {
    scopes.push(scope)
    return
  }
  scopes.splice(index, 1)
}

export function buildFallbackGroupOptions(
  groups: AdminGroup[],
  noFallbackLabel: string,
  currentId?: number | null
): NullableNumberSelectOption[] {
  return [
    { value: null, label: noFallbackLabel },
    ...groups
      .filter(
        (group) =>
          group.platform === 'anthropic' &&
          !group.claude_code_only &&
          group.status === 'active' &&
          group.id !== currentId
      )
      .map((group) => ({
        value: group.id,
        label: group.name
      }))
  ]
}

export function buildInvalidRequestFallbackOptions(
  groups: AdminGroup[],
  noFallbackLabel: string,
  currentId?: number | null
): NullableNumberSelectOption[] {
  return [
    { value: null, label: noFallbackLabel },
    ...groups
      .filter(
        (group) =>
          group.platform === 'anthropic' &&
          group.status === 'active' &&
          group.subscription_type !== 'subscription' &&
          group.fallback_group_id_on_invalid_request === null &&
          group.id !== currentId
      )
      .map((group) => ({
        value: group.id,
        label: group.name
      }))
  ]
}

export function buildCopyAccountsGroupOptions(
  groups: AdminGroup[],
  platform: GroupPlatform,
  currentId?: number | null
): NumberSelectOption[] {
  return groups
    .filter(
      (group) =>
        group.platform === platform &&
        (group.account_count || 0) > 0 &&
        group.id !== currentId
    )
    .map((group) => ({
      value: group.id,
      label: `${group.name} (${group.account_count || 0} 个账号)`
    }))
}

export function addCopyAccountsGroupSelection(selectedGroupIds: number[], groupId: number): void {
  if (groupId <= 0 || selectedGroupIds.includes(groupId)) {
    return
  }

  selectedGroupIds.push(groupId)
}

export function removeCopyAccountsGroupSelection(
  selectedGroupIds: number[],
  groupId: number
): void {
  const index = selectedGroupIds.indexOf(groupId)
  if (index === -1) {
    return
  }

  selectedGroupIds.splice(index, 1)
}

export function applyCreateFormSubscriptionTypeRules(
  form: Pick<CreateGroupForm, 'subscription_type' | 'is_exclusive' | 'fallback_group_id_on_invalid_request'>
): void {
  if (form.subscription_type !== 'subscription') {
    return
  }

  form.is_exclusive = true
  form.fallback_group_id_on_invalid_request = null
}

export function applyCreateFormPlatformRules(
  form: Pick<
    CreateGroupForm,
    | 'platform'
    | 'fallback_group_id_on_invalid_request'
    | 'allow_messages_dispatch'
    | 'default_mapped_model'
    | 'require_oauth_only'
    | 'require_privacy_set'
  >
): void {
  if (!['anthropic', 'antigravity'].includes(form.platform)) {
    form.fallback_group_id_on_invalid_request = null
  }
  if (form.platform !== 'openai') {
    form.allow_messages_dispatch = false
    form.default_mapped_model = ''
  }
  if (!ACCOUNT_FILTER_PLATFORMS.includes(form.platform)) {
    form.require_oauth_only = false
    form.require_privacy_set = false
  }
}

export function buildModelRoutingPayload(
  rules: ModelRoutingRule[]
): Record<string, number[]> | null {
  const result: Record<string, number[]> = {}

  for (const rule of rules) {
    const pattern = rule.pattern.trim()
    if (!pattern) {
      continue
    }

    const accountIds = rule.accounts
      .map((account) => account.id)
      .filter((accountId) => accountId > 0)
    if (accountIds.length === 0) {
      continue
    }

    result[pattern] = accountIds
  }

  return Object.keys(result).length > 0 ? result : null
}

export function buildCreateGroupPayload(
  form: CreateGroupForm,
  routingRules: ModelRoutingRule[]
): CreateGroupRequest {
  return buildBaseGroupPayload(form, routingRules)
}

export function buildUpdateGroupPayload(
  form: EditGroupForm,
  routingRules: ModelRoutingRule[]
): UpdateGroupRequest {
  return {
    ...buildBaseGroupPayload(form, routingRules),
    status: form.status,
    fallback_group_id: form.fallback_group_id === null ? 0 : form.fallback_group_id,
    fallback_group_id_on_invalid_request:
      form.fallback_group_id_on_invalid_request === null
        ? 0
        : form.fallback_group_id_on_invalid_request
  }
}

export function normalizeOptionalLimit(value: number | string | null | undefined): number | null {
  if (value === null || value === undefined) {
    return null
  }

  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (!trimmed) {
      return null
    }

    const parsed = Number(trimmed)
    return Number.isFinite(parsed) && parsed > 0 ? parsed : null
  }

  return Number.isFinite(value) && value > 0 ? value : null
}

function buildBaseGroupPayload(
  form: CreateGroupForm | EditGroupForm,
  routingRules: ModelRoutingRule[]
): CreateGroupRequest {
  return {
    name: form.name,
    description: form.description,
    platform: form.platform,
    rate_multiplier: form.rate_multiplier,
    is_exclusive: form.is_exclusive,
    subscription_type: form.subscription_type,
    daily_limit_usd: normalizeOptionalLimit(form.daily_limit_usd),
    weekly_limit_usd: normalizeOptionalLimit(form.weekly_limit_usd),
    monthly_limit_usd: normalizeOptionalLimit(form.monthly_limit_usd),
    image_price_1k: form.image_price_1k,
    image_price_2k: form.image_price_2k,
    image_price_4k: form.image_price_4k,
    claude_code_only: form.claude_code_only,
    fallback_group_id: form.fallback_group_id,
    fallback_group_id_on_invalid_request: form.fallback_group_id_on_invalid_request,
    allow_messages_dispatch: form.allow_messages_dispatch,
    default_mapped_model: form.default_mapped_model,
    require_oauth_only: form.require_oauth_only,
    require_privacy_set: form.require_privacy_set,
    model_routing_enabled: form.model_routing_enabled,
    supported_model_scopes: [...form.supported_model_scopes],
    mcp_xml_inject: form.mcp_xml_inject,
    copy_accounts_from_group_ids: [...form.copy_accounts_from_group_ids],
    model_routing: buildModelRoutingPayload(routingRules)
  }
}
