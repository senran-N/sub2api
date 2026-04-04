import { describe, expect, it } from 'vitest'
import type { AdminGroup, AdminUser, UserAttributeDefinition } from '@/types'
import {
  DEFAULT_USER_HIDDEN_COLUMNS,
  applyUsersPageChange,
  applyUsersPageSizeChange,
  buildUserAttributeColumns,
  buildUserGroupFilterOptions,
  buildUserListFilters,
  buildUserTableColumns,
  createDefaultUsersFilters,
  filterVisibleUserColumns,
  formatUserAttributeValue,
  getAttributeDefinitionName,
  getUserColumnToggleEffects,
  getUserGroupsSummary,
  getUserSubscriptionDaysRemaining,
  toggleBuiltInUserFilter,
  toggleUserAttributeFilter
} from '../usersTable'

function createAdminGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 1,
    name: 'group',
    description: null,
    platform: 'anthropic',
    rate_multiplier: 1,
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
    supported_model_scopes: [],
    account_count: 0,
    active_account_count: 0,
    rate_limited_account_count: 0,
    default_mapped_model: '',
    sort_order: 0,
    ...overrides
  }
}

function createAdminUser(overrides: Partial<AdminUser> = {}): AdminUser {
  return {
    id: 1,
    email: 'user@example.com',
    username: 'user',
    role: 'user',
    balance: 0,
    status: 'active',
    allowed_groups: [],
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    notes: '',
    group_rates: {},
    current_concurrency: 0,
    sora_storage_quota_bytes: 0,
    sora_storage_used_bytes: 0,
    concurrency: 1,
    ...overrides
  } as AdminUser
}

function createAttributeDefinition(
  overrides: Partial<UserAttributeDefinition> = {}
): UserAttributeDefinition {
  return {
    id: 1,
    key: 'department',
    name: 'Department',
    description: '',
    type: 'select',
    options: [],
    required: false,
    validation: {},
    placeholder: '',
    display_order: 0,
    enabled: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

describe('usersTable helpers', () => {
  it('builds normalized list filters and pagination mutations', () => {
    const filters = createDefaultUsersFilters()
    filters.role = 'admin'
    filters.status = 'disabled'
    filters.group = 'VIP'

    expect(
      buildUserListFilters(filters, '  alice  ', { 1: ' north ', 2: '' }, true)
    ).toEqual({
      role: 'admin',
      status: 'disabled',
      search: 'alice',
      group_name: 'VIP',
      attributes: { 1: 'north' },
      include_subscriptions: true
    })

    const pagination = { page: 5, page_size: 20, pages: 3 }
    applyUsersPageChange(pagination, 9)
    expect(pagination.page).toBe(3)
    applyUsersPageSizeChange(pagination, 50)
    expect(pagination).toEqual({ page: 1, page_size: 50, pages: 3 })
  })

  it('formats attribute values and names across raw, select, and multi-select fields', () => {
    const definitions = [
      createAttributeDefinition({
        id: 1,
        name: 'Department',
        type: 'select',
        options: [{ value: 'rd', label: 'R&D' }]
      }),
      createAttributeDefinition({
        id: 2,
        name: 'Tags',
        type: 'multi_select',
        options: [
          { value: 'vip', label: 'VIP' },
          { value: 'cn', label: 'China' }
        ]
      })
    ]

    const values = {
      7: {
        1: 'rd',
        2: '["vip","cn"]',
        3: 'raw'
      }
    }

    expect(getAttributeDefinitionName(definitions, 1)).toBe('Department')
    expect(getAttributeDefinitionName(definitions, 99)).toBe('99')
    expect(formatUserAttributeValue(values, definitions, 7, 1)).toBe('R&D')
    expect(formatUserAttributeValue(values, definitions, 7, 2)).toBe('VIP, China')
    expect(formatUserAttributeValue(values, definitions, 7, 3)).toBe('raw')
    expect(formatUserAttributeValue(values, definitions, 8, 1)).toBe('-')
  })

  it('builds columns, visibility, group options, and user group summaries', () => {
    const attributeColumns = buildUserAttributeColumns([
      createAttributeDefinition({ id: 1, name: 'Department', enabled: true }),
      createAttributeDefinition({ id: 2, name: 'Hidden', enabled: false })
    ])
    expect(attributeColumns).toEqual([
      { key: 'attr_1', label: 'Department', sortable: false }
    ])

    const allColumns = buildUserTableColumns(attributeColumns, {
      user: 'User',
      id: 'ID',
      username: 'Username',
      notes: 'Notes',
      role: 'Role',
      groups: 'Groups',
      subscriptions: 'Subscriptions',
      balance: 'Balance',
      usage: 'Usage',
      concurrency: 'Concurrency',
      status: 'Status',
      created: 'Created',
      actions: 'Actions'
    })
    expect(filterVisibleUserColumns(allColumns, new Set(DEFAULT_USER_HIDDEN_COLUMNS))).not.toEqual(allColumns)

    const groups = [
      createAdminGroup({ id: 1, name: 'Exclusive', is_exclusive: true }),
      createAdminGroup({ id: 2, name: 'Public', is_exclusive: false }),
      createAdminGroup({ id: 3, name: 'Subscription', is_exclusive: true, subscription_type: 'subscription' })
    ]

    expect(buildUserGroupFilterOptions(groups, 'All Groups')).toEqual([
      { value: '', label: 'All Groups' },
      { value: 'Exclusive', label: 'Exclusive' }
    ])

    expect(
      getUserGroupsSummary(groups, createAdminUser({ allowed_groups: [1] }))
    ).toEqual({
      exclusive: [groups[0]],
      publicGroups: [groups[1]]
    })
  })

  it('toggles built-in and attribute filters, derives side effects, and computes subscription days', () => {
    const visibleFilters = new Set<string>(['role'])
    const filters = createDefaultUsersFilters()
    filters.role = 'user'

    expect(toggleBuiltInUserFilter(visibleFilters, filters, 'role')).toEqual({
      shouldLoadGroups: false
    })
    expect(visibleFilters.has('role')).toBe(false)
    expect(filters.role).toBe('')

    expect(toggleBuiltInUserFilter(visibleFilters, filters, 'group')).toEqual({
      shouldLoadGroups: true
    })
    expect(visibleFilters.has('group')).toBe(true)

    const activeAttributeFilters: Record<number, string> = {}
    toggleUserAttributeFilter(visibleFilters, activeAttributeFilters, 5)
    expect(activeAttributeFilters[5]).toBe('')
    toggleUserAttributeFilter(visibleFilters, activeAttributeFilters, 5)
    expect(activeAttributeFilters[5]).toBeUndefined()

    expect(getUserColumnToggleEffects('usage', true)).toEqual({
      refreshSecondaryData: true,
      reloadUsers: false,
      loadGroups: false
    })
    expect(getUserColumnToggleEffects('subscriptions', false)).toEqual({
      refreshSecondaryData: false,
      reloadUsers: true,
      loadGroups: false
    })
    expect(getUserColumnToggleEffects('groups', true)).toEqual({
      refreshSecondaryData: false,
      reloadUsers: false,
      loadGroups: true
    })

    expect(
      getUserSubscriptionDaysRemaining(
        '2026-04-10T00:00:00Z',
        new Date('2026-04-04T00:00:00Z')
      )
    ).toBe(6)
  })
})
