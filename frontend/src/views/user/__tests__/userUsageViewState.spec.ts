import { describe, expect, it } from 'vitest'
import {
  applyUserUsageDateRange,
  applyUserUsagePageChange,
  applyUserUsagePageSizeChange,
  buildDefaultUserUsageFilters,
  buildResetUserUsageState,
  formatUserUsageLocalDate,
  getLast7DaysUserUsageRange,
  resetUserUsagePaginationPage
} from '../userUsageViewState'

describe('userUsageViewState', () => {
  it('formats local dates and builds default seven day range', () => {
    const now = new Date('2026-04-04T12:00:00Z')

    expect(formatUserUsageLocalDate(now)).toBe('2026-04-04')
    expect(getLast7DaysUserUsageRange(now)).toEqual({
      startDate: '2026-03-29',
      endDate: '2026-04-04'
    })
  })

  it('builds filter state and reset state', () => {
    const range = {
      startDate: '2026-03-29',
      endDate: '2026-04-04'
    }

    expect(buildDefaultUserUsageFilters(range)).toEqual({
      api_key_id: undefined,
      start_date: '2026-03-29',
      end_date: '2026-04-04'
    })

    expect(
      applyUserUsageDateRange(
        {
          api_key_id: 3,
          start_date: '2026-03-01',
          end_date: '2026-03-02'
        },
        range
      )
    ).toEqual({
      api_key_id: 3,
      start_date: '2026-03-29',
      end_date: '2026-04-04'
    })

    expect(buildResetUserUsageState(new Date('2026-04-04T12:00:00Z'))).toEqual({
      filters: {
        api_key_id: undefined,
        start_date: '2026-03-29',
        end_date: '2026-04-04'
      },
      range
    })
  })

  it('applies pagination mutations', () => {
    const pagination = {
      page: 4,
      page_size: 20
    }

    applyUserUsagePageChange(pagination, 7)
    expect(pagination.page).toBe(7)

    applyUserUsagePageSizeChange(pagination, 50)
    expect(pagination).toEqual({
      page: 1,
      page_size: 50
    })

    pagination.page = 3
    resetUserUsagePaginationPage(pagination)
    expect(pagination.page).toBe(1)
  })
})
