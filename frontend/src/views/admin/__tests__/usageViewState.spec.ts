import { describe, expect, it } from 'vitest'
import {
  applyUsageDateRangeState,
  applyUsagePageChange,
  applyUsagePageSizeChange,
  applyUsageRouteQueryState,
  buildDefaultUsageFilters,
  buildResetUsageState,
  formatUsageLocalDate,
  getLast24HoursUsageRange,
  getUsageGranularityForRange,
  getUsageQueryNumberValue,
  getUsageQueryStringValue,
  resetUsagePaginationPage
} from '../usageViewState'

describe('usageViewState helpers', () => {
  it('formats local dates and derives default 24 hour ranges', () => {
    const now = new Date('2026-04-04T12:00:00Z')

    expect(formatUsageLocalDate(now)).toBe('2026-04-04')
    expect(getLast24HoursUsageRange(now)).toEqual({
      startDate: '2026-04-03',
      endDate: '2026-04-04'
    })
  })

  it('derives query values and range granularity', () => {
    expect(getUsageQueryStringValue(['', '2026-04-02'])).toBe('2026-04-02')
    expect(getUsageQueryStringValue(null)).toBeUndefined()
    expect(getUsageQueryNumberValue('42')).toBe(42)
    expect(getUsageQueryNumberValue('x')).toBeUndefined()
    expect(getUsageGranularityForRange('2026-04-03', '2026-04-04')).toBe('hour')
    expect(getUsageGranularityForRange('2026-04-01', '2026-04-04')).toBe('day')
  })

  it('builds default filters and applies route query overrides', () => {
    const defaults = buildDefaultUsageFilters({
      startDate: '2026-04-03',
      endDate: '2026-04-04'
    })

    expect(defaults).toEqual({
      user_id: undefined,
      model: undefined,
      group_id: undefined,
      request_type: undefined,
      billing_type: null,
      start_date: '2026-04-03',
      end_date: '2026-04-04'
    })

    expect(
      applyUsageRouteQueryState(
        {
          start_date: '2026-04-01',
          end_date: '2026-04-05',
          user_id: '7'
        },
        defaults,
        {
          startDate: '2026-04-03',
          endDate: '2026-04-04'
        }
      )
    ).toEqual({
      filters: {
        ...defaults,
        user_id: 7,
        start_date: '2026-04-01',
        end_date: '2026-04-05'
      },
      range: {
        startDate: '2026-04-01',
        endDate: '2026-04-05'
      },
      granularity: 'day'
    })
  })

  it('applies date changes, reset state, and pagination mutations', () => {
    const filters = buildDefaultUsageFilters({
      startDate: '2026-04-03',
      endDate: '2026-04-04'
    })

    expect(
      applyUsageDateRangeState(
        {
          startDate: '2026-04-04',
          endDate: '2026-04-04'
        },
        filters
      )
    ).toEqual({
      filters: {
        ...filters,
        start_date: '2026-04-04',
        end_date: '2026-04-04'
      },
      range: {
        startDate: '2026-04-04',
        endDate: '2026-04-04'
      },
      granularity: 'hour'
    })

    expect(buildResetUsageState(new Date('2026-04-04T12:00:00Z'))).toEqual({
      filters: {
        start_date: '2026-04-03',
        end_date: '2026-04-04',
        request_type: undefined,
        billing_type: null
      },
      range: {
        startDate: '2026-04-03',
        endDate: '2026-04-04'
      },
      granularity: 'hour'
    })

    const pagination = {
      page: 5,
      page_size: 20
    }
    applyUsagePageChange(pagination, 7)
    expect(pagination.page).toBe(7)
    applyUsagePageSizeChange(pagination, 50)
    expect(pagination).toEqual({
      page: 1,
      page_size: 50
    })
    pagination.page = 4
    resetUsagePaginationPage(pagination)
    expect(pagination.page).toBe(1)
  })
})
