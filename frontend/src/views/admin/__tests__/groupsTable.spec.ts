import { describe, expect, it } from 'vitest'
import type { AdminGroup } from '@/types'
import {
  applyGroupPageChange,
  applyGroupPageReset,
  applyGroupPageSizeChange,
  buildGroupListFilters,
  buildGroupSortOrderUpdates,
  formatGroupCost,
  getGroupPlatformBadgeClass,
  mapGroupCapacitySummary,
  mapGroupUsageSummary,
  sortGroupsBySortOrder
} from '../groupsTable'

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

describe('groupsTable helpers', () => {
  it('builds normalized list filters and page mutations', () => {
    expect(
      buildGroupListFilters(
        {
          platform: 'openai',
          status: 'inactive',
          is_exclusive: 'false'
        },
        '  hello  '
      )
    ).toEqual({
      platform: 'openai',
      status: 'inactive',
      is_exclusive: false,
      search: 'hello'
    })

    const pagination = {
      page: 5,
      page_size: 20
    }
    applyGroupPageReset(pagination)
    expect(pagination.page).toBe(1)
    applyGroupPageChange(pagination, 4)
    expect(pagination.page).toBe(4)
    applyGroupPageSizeChange(pagination, 50)
    expect(pagination).toEqual({
      page: 1,
      page_size: 50
    })
  })

  it('formats usage costs and maps summary arrays into keyed maps', () => {
    expect(formatGroupCost(1234.56)).toBe('1235')
    expect(formatGroupCost(123.45)).toBe('123.5')
    expect(formatGroupCost(12.345)).toBe('12.35')
    expect(getGroupPlatformBadgeClass('anthropic')).toBe('theme-chip--brand-orange')
    expect(getGroupPlatformBadgeClass('openai')).toBe('theme-chip--success')
    expect(getGroupPlatformBadgeClass('antigravity')).toBe('theme-chip--brand-purple')
    expect(getGroupPlatformBadgeClass('gemini')).toBe('theme-chip--info')

    expect(
      mapGroupUsageSummary([
        { group_id: 7, today_cost: 1.2, total_cost: 9.9 }
      ]).get(7)
    ).toEqual({
      today_cost: 1.2,
      total_cost: 9.9
    })

    expect(
      mapGroupCapacitySummary([
        {
          group_id: 4,
          concurrency_used: 1,
          concurrency_max: 10,
          sessions_used: 2,
          sessions_max: 20,
          rpm_used: 3,
          rpm_max: 30
        }
      ]).get(4)
    ).toEqual({
      concurrencyUsed: 1,
      concurrencyMax: 10,
      sessionsUsed: 2,
      sessionsMax: 20,
      rpmUsed: 3,
      rpmMax: 30
    })
  })

  it('sorts groups and builds spaced sort-order updates', () => {
    const groups = [
      createAdminGroup({ id: 1, sort_order: 20 }),
      createAdminGroup({ id: 2, sort_order: 0 }),
      createAdminGroup({ id: 3, sort_order: 10 })
    ]

    expect(sortGroupsBySortOrder(groups).map((group) => group.id)).toEqual([2, 3, 1])
    expect(buildGroupSortOrderUpdates(groups, 5)).toEqual([
      { id: 1, sort_order: 0 },
      { id: 2, sort_order: 5 },
      { id: 3, sort_order: 10 }
    ])
  })
})
