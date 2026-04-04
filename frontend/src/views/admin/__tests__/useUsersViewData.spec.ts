import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import type { AdminGroup, AdminUser, UserAttributeDefinition } from '@/types'
import { useUsersViewData } from '../useUsersViewData'

const { listUsers, getAllGroups, listEnabledDefinitions, getBatchUserAttributes, getBatchUsersUsage } =
  vi.hoisted(() => ({
    listUsers: vi.fn(),
    getAllGroups: vi.fn(),
    listEnabledDefinitions: vi.fn(),
    getBatchUserAttributes: vi.fn(),
    getBatchUsersUsage: vi.fn()
  }))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      list: listUsers
    },
    groups: {
      getAll: getAllGroups
    },
    userAttributes: {
      listEnabledDefinitions,
      getBatchUserAttributes
    },
    dashboard: {
      getBatchUsersUsage
    }
  }
}))

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

function createGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
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

describe('useUsersViewData', () => {
  const showError = vi.fn()
  const setCurrentUserIds = vi.fn()
  const resetSecondaryDataState = vi.fn()
  const scheduleUsersSecondaryDataLoad = vi.fn()
  const isSecondaryDataRequestCurrent = vi.fn(() => true)

  beforeEach(() => {
    showError.mockReset()
    setCurrentUserIds.mockReset()
    resetSecondaryDataState.mockReset()
    scheduleUsersSecondaryDataLoad.mockReset()
    isSecondaryDataRequestCurrent.mockReset()
    isSecondaryDataRequestCurrent.mockReturnValue(true)
    listUsers.mockReset()
    getAllGroups.mockReset()
    listEnabledDefinitions.mockReset()
    getBatchUserAttributes.mockReset()
    getBatchUsersUsage.mockReset()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  function createComposable() {
    return useUsersViewData({
      t: (key: string) => key,
      showError,
      filters: {
        role: 'admin',
        status: '',
        group: ''
      },
      activeAttributeFilters: { 9: 'north' },
      searchQuery: ref('  alice  '),
      pagination: {
        page: 2,
        page_size: 20,
        total: 0,
        pages: 0
      },
      hasVisibleUsageColumn: computed(() => true),
      hasVisibleSubscriptionsColumn: computed(() => true),
      hasVisibleAttributeColumns: computed(() => true),
      isSecondaryDataRequestCurrent,
      setCurrentUserIds,
      resetSecondaryDataState,
      scheduleUsersSecondaryDataLoad
    })
  }

  it('loads groups once and loads attribute definitions', async () => {
    const data = createComposable()
    getAllGroups.mockResolvedValue([createGroup({ id: 1 })])
    listEnabledDefinitions.mockResolvedValue([createAttributeDefinition({ id: 7 })])

    await data.loadAllGroups()
    await data.loadAllGroups()
    await data.loadAttributeDefinitions()

    expect(getAllGroups).toHaveBeenCalledTimes(1)
    expect(data.allGroups.value).toHaveLength(1)
    expect(listEnabledDefinitions).toHaveBeenCalledTimes(1)
    expect(data.attributeDefinitions.value[0].id).toBe(7)
  })

  it('loads users, normalizes filters, and schedules secondary data fetches', async () => {
    const data = createComposable()
    listUsers.mockResolvedValue({
      items: [createAdminUser({ id: 7 }), createAdminUser({ id: 8 })],
      total: 2,
      pages: 1
    })

    await data.loadUsers()

    expect(listUsers).toHaveBeenCalledWith(
      2,
      20,
      {
        role: 'admin',
        status: undefined,
        search: 'alice',
        group_name: undefined,
        attributes: { 9: 'north' },
        include_subscriptions: true
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal)
      })
    )
    expect(data.users.value.map((user) => user.id)).toEqual([7, 8])
    expect(setCurrentUserIds).toHaveBeenCalledWith([7, 8])
    expect(resetSecondaryDataState).toHaveBeenCalledTimes(1)
    expect(scheduleUsersSecondaryDataLoad).toHaveBeenCalledTimes(1)
  })

  it('loads secondary usage and attribute data only for current requests', async () => {
    const data = createComposable()
    data.attributeDefinitions.value = [createAttributeDefinition({ id: 3 })]
    getBatchUsersUsage.mockResolvedValue({
      stats: {
        3: {
          today_actual_cost: 1,
          total_actual_cost: 2
        }
      }
    })
    getBatchUserAttributes.mockResolvedValue({
      attributes: {
        3: {
          3: 'vip'
        }
      }
    })

    await data.loadUsersSecondaryData([3], undefined, 4)

    expect(getBatchUsersUsage).toHaveBeenCalledWith([3])
    expect(getBatchUserAttributes).toHaveBeenCalledWith([3])
    expect(data.usageStats.value['3']).toEqual({
      today_actual_cost: 1,
      total_actual_cost: 2
    })
    expect(data.userAttributeValues.value[3]).toEqual({ 3: 'vip' })
  })

  it('surfaces list errors and ignores abort-like failures', async () => {
    const data = createComposable()
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    listUsers.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'load failed'
        }
      }
    })
    await data.loadUsers()
    expect(showError).toHaveBeenCalledWith('load failed')

    listUsers.mockRejectedValueOnce({
      name: 'CanceledError'
    })
    await data.loadUsers()
    expect(showError).toHaveBeenCalledTimes(1)
    expect(consoleSpy).toHaveBeenCalledTimes(1)
  })
})
