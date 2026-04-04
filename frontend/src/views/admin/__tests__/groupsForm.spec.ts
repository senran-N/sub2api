import { describe, expect, it } from 'vitest'
import type { AdminGroup } from '@/types'
import {
  addCopyAccountsGroupSelection,
  applyCreateFormPlatformRules,
  applyCreateFormSubscriptionTypeRules,
  buildCopyAccountsGroupOptions,
  buildCreateGroupPayload,
  buildFallbackGroupOptions,
  buildInvalidRequestFallbackOptions,
  buildModelRoutingPayload,
  buildUpdateGroupPayload,
  createDefaultCreateGroupForm,
  createDefaultEditGroupForm,
  hydrateEditGroupForm,
  removeCopyAccountsGroupSelection,
  resetCreateGroupForm,
  resetEditGroupForm,
  toggleModelScope
} from '../groupsForm'

function createAdminGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 1,
    name: 'Anthropic Primary',
    description: 'main group',
    platform: 'anthropic',
    rate_multiplier: 1.2,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
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
    allow_messages_dispatch: false,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    model_routing: null,
    model_routing_enabled: false,
    mcp_xml_inject: true,
    supported_model_scopes: ['claude', 'gemini_text', 'gemini_image'],
    account_count: 2,
    active_account_count: 2,
    rate_limited_account_count: 0,
    default_mapped_model: '',
    sort_order: 10,
    ...overrides
  }
}

describe('group form defaults', () => {
  it('resets create and edit forms back to isolated defaults', () => {
    const createForm = createDefaultCreateGroupForm()
    createForm.name = 'dirty'
    createForm.supported_model_scopes.push('extra')
    createForm.copy_accounts_from_group_ids.push(7)
    resetCreateGroupForm(createForm)

    expect(createForm).toEqual(createDefaultCreateGroupForm())
    expect(createForm.supported_model_scopes).not.toBe(
      createDefaultCreateGroupForm().supported_model_scopes
    )

    const editForm = createDefaultEditGroupForm()
    editForm.status = 'inactive'
    editForm.copy_accounts_from_group_ids.push(9)
    resetEditGroupForm(editForm)

    expect(editForm).toEqual(createDefaultEditGroupForm())
  })
})

describe('hydrateEditGroupForm', () => {
  it('loads admin group fields and converts Sora quota bytes to GB', () => {
    const editForm = createDefaultEditGroupForm()
    editForm.copy_accounts_from_group_ids = [5]

    hydrateEditGroupForm(
      editForm,
      createAdminGroup({
        platform: 'openai',
        allow_messages_dispatch: true,
        default_mapped_model: 'gpt-5.4',
        model_routing_enabled: true,
        supported_model_scopes: ['claude'],
        sora_storage_quota_bytes: 1610612736
      })
    )

    expect(editForm.platform).toBe('openai')
    expect(editForm.allow_messages_dispatch).toBe(true)
    expect(editForm.default_mapped_model).toBe('gpt-5.4')
    expect(editForm.model_routing_enabled).toBe(true)
    expect(editForm.supported_model_scopes).toEqual(['claude'])
    expect(editForm.sora_storage_quota_gb).toBe(1.5)
    expect(editForm.copy_accounts_from_group_ids).toEqual([])
  })
})

describe('routing helpers', () => {
  it('toggles model scopes and builds routing payloads from valid rules only', () => {
    const scopes = ['claude']
    toggleModelScope(scopes, 'gemini_text')
    toggleModelScope(scopes, 'claude')
    expect(scopes).toEqual(['gemini_text'])

    expect(
      buildModelRoutingPayload([
        {
          pattern: 'claude-*',
          accounts: [
            { id: 3, name: 'a' },
            { id: -1, name: 'bad' }
          ]
        },
        {
          pattern: '   ',
          accounts: [{ id: 5, name: 'ignored' }]
        },
        {
          pattern: 'gemini-*',
          accounts: []
        }
      ])
    ).toEqual({
      'claude-*': [3]
    })
  })

  it('builds fallback and copy-account options with the correct eligibility rules', () => {
    const groups = [
      createAdminGroup({ id: 1, name: 'alpha', account_count: 3 }),
      createAdminGroup({ id: 2, name: 'beta', claude_code_only: true }),
      createAdminGroup({ id: 3, name: 'gamma', fallback_group_id_on_invalid_request: 9 }),
      createAdminGroup({ id: 4, name: 'delta', platform: 'openai', account_count: 2 }),
      createAdminGroup({ id: 5, name: 'sub', subscription_type: 'subscription' })
    ]

    expect(buildFallbackGroupOptions(groups, 'none', 1)).toEqual([
      { value: null, label: 'none' },
      { value: 3, label: 'gamma' },
      { value: 5, label: 'sub' }
    ])

    expect(buildInvalidRequestFallbackOptions(groups, 'empty', 1)).toEqual([
      { value: null, label: 'empty' },
      { value: 2, label: 'beta' }
    ])

    expect(buildCopyAccountsGroupOptions(groups, 'openai')).toEqual([
      { value: 4, label: 'delta (2 个账号)' }
    ])
  })

  it('adds and removes copy-account selections without duplicate noise', () => {
    const selected = [2]

    addCopyAccountsGroupSelection(selected, 2)
    addCopyAccountsGroupSelection(selected, -1)
    addCopyAccountsGroupSelection(selected, 5)
    removeCopyAccountsGroupSelection(selected, 9)
    removeCopyAccountsGroupSelection(selected, 2)

    expect(selected).toEqual([5])
  })
})

describe('create form rules', () => {
  it('applies subscription and platform constraints without fallback-heavy branching', () => {
    const createForm = createDefaultCreateGroupForm()
    createForm.subscription_type = 'subscription'
    createForm.fallback_group_id_on_invalid_request = 4
    applyCreateFormSubscriptionTypeRules(createForm)
    expect(createForm.is_exclusive).toBe(true)
    expect(createForm.fallback_group_id_on_invalid_request).toBeNull()

    createForm.platform = 'sora'
    createForm.allow_messages_dispatch = true
    createForm.default_mapped_model = 'gpt-5.4'
    createForm.require_oauth_only = true
    createForm.require_privacy_set = true
    createForm.fallback_group_id_on_invalid_request = 8
    applyCreateFormPlatformRules(createForm)
    expect(createForm.fallback_group_id_on_invalid_request).toBeNull()
    expect(createForm.allow_messages_dispatch).toBe(false)
    expect(createForm.default_mapped_model).toBe('')
    expect(createForm.require_oauth_only).toBe(false)
    expect(createForm.require_privacy_set).toBe(false)
  })
})

describe('group payload builders', () => {
  it('builds create payload with normalized limits, quota bytes, and copied routing rules', () => {
    const createForm = createDefaultCreateGroupForm()
    createForm.name = 'primary'
    createForm.daily_limit_usd = '' as unknown as number
    createForm.weekly_limit_usd = 12
    createForm.monthly_limit_usd = -5
    createForm.sora_storage_quota_gb = 1.25
    createForm.allow_messages_dispatch = true
    createForm.copy_accounts_from_group_ids = [1, 2]

    const payload = buildCreateGroupPayload(createForm, [
      {
        pattern: 'claude-*',
        accounts: [{ id: 11, name: 'A' }]
      }
    ])

    expect(payload.daily_limit_usd).toBeNull()
    expect(payload.weekly_limit_usd).toBe(12)
    expect(payload.monthly_limit_usd).toBeNull()
    expect(payload.sora_storage_quota_bytes).toBe(1342177280)
    expect(payload.allow_messages_dispatch).toBe(true)
    expect(payload.copy_accounts_from_group_ids).toEqual([1, 2])
    expect(payload.model_routing).toEqual({
      'claude-*': [11]
    })
  })

  it('builds update payload and converts null fallback ids to zero', () => {
    const editForm = createDefaultEditGroupForm()
    editForm.name = 'secondary'
    editForm.status = 'inactive'
    editForm.fallback_group_id = null
    editForm.fallback_group_id_on_invalid_request = 9

    const payload = buildUpdateGroupPayload(editForm, [])
    expect(payload.status).toBe('inactive')
    expect(payload.fallback_group_id).toBe(0)
    expect(payload.fallback_group_id_on_invalid_request).toBe(9)
    expect(payload.model_routing).toBeNull()
  })
})
