import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useUserUsageViewData } from '../useUserUsageViewData'

const { query, getStatsByDateRange, list } = vi.hoisted(() => ({
  query: vi.fn(),
  getStatsByDateRange: vi.fn(),
  list: vi.fn()
}))

vi.mock('@/api', () => ({
  usageAPI: {
    query,
    getStatsByDateRange
  },
  keysAPI: {
    list
  }
}))

vi.mock('@/utils/format', () => ({
  formatReasoningEffort: (value: string | null | undefined) => value ?? '-'
}))

function createUsageLog() {
  return {
    request_id: 'req-1',
    actual_cost: 0.1,
    total_cost: 0.1,
    rate_multiplier: 1,
    input_cost: 0.01,
    output_cost: 0.02,
    cache_creation_cost: 0,
    cache_read_cost: 0,
    input_tokens: 10,
    output_tokens: 20,
    cache_creation_tokens: 0,
    cache_read_tokens: 0,
    cache_creation_5m_tokens: 0,
    cache_creation_1h_tokens: 0,
    image_count: 0,
    image_size: null,
    first_token_ms: 12,
    duration_ms: 345,
    created_at: '2026-04-04T00:00:00Z',
    model: 'gpt-5.4',
    reasoning_effort: 'medium',
    inbound_endpoint: '/v1/chat/completions',
    request_type: 'stream',
    stream: true,
    api_key: { name: 'demo-key' }
  }
}

function createComposable() {
  const filters = ref({
    api_key_id: 3,
    start_date: '2026-03-29',
    end_date: '2026-04-04'
  })
  const startDate = ref('2026-03-29')
  const endDate = ref('2026-04-04')
  const pagination = {
    page: 2,
    page_size: 20,
    total: 1,
    pages: 1
  }
  const showError = vi.fn()
  const showWarning = vi.fn()
  const showSuccess = vi.fn()
  const showInfo = vi.fn()

  const composable = useUserUsageViewData({
    filters,
    startDate,
    endDate,
    pagination,
    showError,
    showWarning,
    showSuccess,
    showInfo,
    t: (key: string) => key
  })

  return {
    composable,
    filters,
    pagination,
    showError,
    showWarning,
    showSuccess,
    showInfo
  }
}

describe('useUserUsageViewData', () => {
  beforeEach(() => {
    query.mockReset()
    getStatsByDateRange.mockReset()
    list.mockReset()

    query.mockResolvedValue({
      items: [createUsageLog()],
      total: 1,
      pages: 1
    })
    getStatsByDateRange.mockResolvedValue({
      total_requests: 1,
      total_tokens: 30,
      total_cost: 0.1,
      total_actual_cost: 0.1,
      average_duration_ms: 123
    })
    list.mockResolvedValue({
      items: [{ id: 3, name: 'demo-key' }]
    })
  })

  it('loads logs, stats, and api keys and normalizes pagination changes', async () => {
    const setup = createComposable()

    await setup.composable.loadInitialData()

    expect(list).toHaveBeenCalledWith(1, 100)
    expect(query).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 2,
        page_size: 20,
        api_key_id: 3
      }),
      expect.any(Object)
    )
    expect(getStatsByDateRange).toHaveBeenCalledWith('2026-03-29', '2026-04-04', 3)
    expect(setup.composable.usageLogs.value).toHaveLength(1)

    setup.composable.applyFilters()
    expect(setup.pagination.page).toBe(1)

    setup.composable.handlePageSizeChange(50)
    expect(setup.pagination.page_size).toBe(50)
    expect(setup.pagination.page).toBe(1)
  })

  it('exports csv data and reports success', async () => {
    const setup = createComposable()
    let exportedBlob: Blob | null = null
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    window.URL.createObjectURL = vi.fn((blob: Blob | MediaSource) => {
      exportedBlob = blob as Blob
      return 'blob:usage-export'
    }) as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

    await setup.composable.exportToCSV()

    expect(exportedBlob).not.toBeNull()
    expect(clickSpy).toHaveBeenCalled()
    expect(setup.showInfo).toHaveBeenCalledWith('usage.preparingExport')
    expect(setup.showSuccess).toHaveBeenCalledWith('usage.exportSuccess')

    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
    clickSpy.mockRestore()
  })

  it('warns instead of exporting empty datasets', async () => {
    const setup = createComposable()
    setup.pagination.total = 0

    await setup.composable.exportToCSV()

    expect(setup.showWarning).toHaveBeenCalledWith('usage.noDataToExport')
  })
})
