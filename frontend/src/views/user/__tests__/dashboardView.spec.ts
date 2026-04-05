import { describe, expect, it } from 'vitest'
import { createDashboardDateRange, formatDashboardDateValue } from '../dashboard/dashboardView'

describe('dashboardView helpers', () => {
  it('formats dashboard dates as yyyy-mm-dd', () => {
    expect(formatDashboardDateValue(new Date('2026-04-05T18:30:00Z'))).toBe('2026-04-05')
  })

  it('builds the default seven day dashboard range', () => {
    expect(createDashboardDateRange(new Date('2026-04-05T00:00:00Z'))).toEqual({
      startDate: '2026-03-30',
      endDate: '2026-04-05'
    })
  })
})
