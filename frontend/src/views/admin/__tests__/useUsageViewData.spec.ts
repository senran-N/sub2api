import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import { useUsageViewData, type ModelDistributionSource } from '../useUsageViewData'

const { usageList, usageStats, modelStats, snapshot, exportList } = vi.hoisted(() => ({
  usageList: vi.fn(),
  usageStats: vi.fn(),
  modelStats: vi.fn(),
  snapshot: vi.fn(),
  exportList: vi.fn()
}))

const { saveAs, aoaToSheet, sheetAddAoa, bookNew, bookAppendSheet, writeWorkbook } = vi.hoisted(
  () => ({
    saveAs: vi.fn(),
    aoaToSheet: vi.fn(() => ({ rows: [] })),
    sheetAddAoa: vi.fn(),
    bookNew: vi.fn(() => ({ sheets: [] })),
    bookAppendSheet: vi.fn(),
    writeWorkbook: vi.fn(() => new Uint8Array([1, 2, 3]))
  })
)

vi.mock('file-saver', () => ({
  saveAs
}))

vi.mock('xlsx', () => ({
  utils: {
    aoa_to_sheet: aoaToSheet,
    sheet_add_aoa: sheetAddAoa,
    book_new: bookNew,
    book_append_sheet: bookAppendSheet
  },
  write: writeWorkbook
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    usage: {
      list: usageList,
      getStats: usageStats
    },
    dashboard: {
      getModelStats: modelStats,
      getSnapshotV2: snapshot
    }
  }
}))

vi.mock('@/api/admin/usage', () => ({
  adminUsageAPI: {
    list: exportList
  }
}))

vi.mock('@/utils/format', () => ({
  formatReasoningEffort: (value: string | null | undefined) => value ?? '-'
}))

function createUsageLog() {
  return {
    id: 1,
    user_id: 7,
    api_key_id: 2,
    account_id: 3,
    group_id: 4,
    request_id: 'req_1',
    model: 'gpt-4.1',
    upstream_model: 'gpt-4.1-mini',
    reasoning_effort: 'medium',
    inbound_endpoint: '/v1/chat/completions',
    upstream_endpoint: '/chat',
    input_tokens: 10,
    output_tokens: 20,
    cache_creation_tokens: 0,
    cache_read_tokens: 0,
    cache_creation_5m_tokens: 0,
    cache_creation_1h_tokens: 0,
    input_cost: 0.1,
    output_cost: 0.2,
    cache_creation_cost: 0,
    cache_read_cost: 0,
    total_cost: 0.3,
    actual_cost: 0.3,
    rate_multiplier: 1,
    account_rate_multiplier: 1.5,
    billing_type: 1,
    request_type: 'stream',
    stream: true,
    duration_ms: 800,
    first_token_ms: 200,
    image_count: 0,
    image_size: null,
    user_agent: 'Mozilla',
    cache_ttl_overridden: false,
    created_at: '2026-04-04T00:00:00Z',
    ip_address: '127.0.0.1',
    user: { id: 7, email: 'user@example.com' },
    api_key: { id: 2, name: 'Key A' },
    group: { id: 4, name: 'Default' },
    account: { id: 3, name: 'Account A' }
  }
}

function createComposable(source: ModelDistributionSource = 'requested') {
  const filters = ref({
    user_id: 7,
    start_date: '2026-04-03',
    end_date: '2026-04-04',
    request_type: 'stream' as const,
    billing_type: null
  })
  const startDate = ref('2026-04-03')
  const endDate = ref('2026-04-04')
  const granularity = ref<'day' | 'hour'>('hour')
  const modelDistributionSource = ref<ModelDistributionSource>(source)
  const pagination = {
    page: 3,
    page_size: 20,
    total: 0
  }
  const showSuccess = vi.fn()
  const showError = vi.fn()

  const composable = useUsageViewData({
    filters,
    startDate,
    endDate,
    granularity,
    modelDistributionSource,
    pagination,
    t: (key: string) => key,
    showSuccess,
    showError
  })

  return {
    composable,
    filters,
    granularity,
    modelDistributionSource,
    pagination,
    showSuccess,
    showError
  }
}

describe('useUsageViewData', () => {
  beforeEach(() => {
    usageList.mockReset()
    usageStats.mockReset()
    modelStats.mockReset()
    snapshot.mockReset()
    exportList.mockReset()
    saveAs.mockReset()
    aoaToSheet.mockClear()
    sheetAddAoa.mockClear()
    bookNew.mockClear()
    bookAppendSheet.mockClear()
    writeWorkbook.mockClear()

    usageList.mockResolvedValue({
      items: [createUsageLog()],
      total: 1
    })
    usageStats.mockResolvedValue({
      total_requests: 1,
      total_input_tokens: 10,
      total_output_tokens: 20,
      total_cache_tokens: 0,
      total_tokens: 30,
      total_cost: 0.3,
      total_actual_cost: 0.3,
      average_duration_ms: 800,
      endpoints: [],
      upstream_endpoints: [],
      endpoint_paths: []
    })
    modelStats.mockResolvedValue({
      models: [{ name: 'gpt-4.1', total_tokens: 30, request_count: 1, actual_cost: 0.3 }]
    })
    snapshot.mockResolvedValue({
      trend: [{ date: '2026-04-04', count: 1 }],
      groups: [{ id: 4, name: 'Default', request_count: 1, total_tokens: 30, actual_cost: 0.3 }]
    })
    exportList.mockResolvedValue({
      items: [createUsageLog()],
      total: 1
    })
  })

  it('reloads logs, stats, model data, and charts with normalized request filters', async () => {
    const setup = createComposable()

    setup.composable.applyFilters()
    await flushPromises()

    expect(setup.pagination.page).toBe(1)
    expect(usageList).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 20,
        request_type: 'stream',
        stream: true
      }),
      expect.any(Object)
    )
    expect(usageStats).toHaveBeenCalledWith(
      expect.objectContaining({
        request_type: 'stream',
        stream: true
      })
    )
    expect(modelStats).toHaveBeenCalledWith(
      expect.objectContaining({
        model_source: 'requested',
        request_type: 'stream',
        stream: true
      })
    )
    expect(snapshot).toHaveBeenCalledWith(
      expect.objectContaining({
        granularity: 'hour',
        include_group_stats: true,
        stream: true
      })
    )
    expect(setup.composable.usageLogs.value).toHaveLength(1)
    expect(setup.composable.groupStats.value).toHaveLength(1)
  })

  it('exports usage logs to excel and reports success', async () => {
    const setup = createComposable()

    await setup.composable.exportToExcel()

    expect(exportList).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 100,
        exact_total: true,
        request_type: 'stream',
        stream: true
      }),
      expect.any(Object)
    )
    expect(aoaToSheet).toHaveBeenCalledTimes(1)
    expect(sheetAddAoa).toHaveBeenCalledTimes(1)
    expect(saveAs).toHaveBeenCalledTimes(1)
    expect(setup.showSuccess).toHaveBeenCalledWith('usage.exportSuccess')
    expect(setup.showError).not.toHaveBeenCalled()
    expect(setup.composable.exporting.value).toBe(false)
  })

  it('treats manual export cancellation as a silent abort', async () => {
    const setup = createComposable()

    exportList.mockImplementation(
      (_params: unknown, options?: { signal?: AbortSignal }) =>
        new Promise((_, reject) => {
          if (options?.signal?.aborted) {
            reject(Object.assign(new Error('aborted'), { name: 'AbortError' }))
            return
          }

          options?.signal?.addEventListener('abort', () => {
            reject(Object.assign(new Error('aborted'), { name: 'AbortError' }))
          })
        })
    )

    const exportPromise = setup.composable.exportToExcel()
    setup.composable.cancelExport()
    await exportPromise

    expect(setup.showError).not.toHaveBeenCalled()
    expect(setup.showSuccess).not.toHaveBeenCalled()
    expect(setup.composable.exporting.value).toBe(false)
  })
})
