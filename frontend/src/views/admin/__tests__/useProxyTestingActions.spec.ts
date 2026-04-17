import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import type { Proxy, ProxyQualityCheckResult } from '@/types'
import { useProxyTestingActions } from '../proxies/useProxyTestingActions'

const { listProxies, testProxy, checkProxyQuality } = vi.hoisted(() => ({
  listProxies: vi.fn(),
  testProxy: vi.fn(),
  checkProxyQuality: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    proxies: {
      list: listProxies,
      testProxy,
      checkProxyQuality
    }
  }
}))

function createProxy(overrides: Partial<Proxy> = {}): Proxy {
  return {
    id: 1,
    name: 'Proxy',
    protocol: 'http',
    host: 'proxy.local',
    port: 8080,
    username: null,
    password: null,
    status: 'active',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

function createQualityResult(
  proxyId: number,
  overrides: Partial<ProxyQualityCheckResult> = {}
): ProxyQualityCheckResult {
  return {
    proxy_id: proxyId,
    score: 90,
    grade: 'A',
    summary: 'Healthy',
    exit_ip: '1.1.1.1',
    country: 'United States',
    country_code: 'US',
    base_latency_ms: 120,
    passed_count: 2,
    warn_count: 0,
    failed_count: 0,
    challenge_count: 0,
    checked_at: 1234567890,
    items: [{ target: 'base_connectivity', status: 'pass' }],
    ...overrides
  }
}

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })

  return {
    promise,
    resolve,
    reject
  }
}

function createComposable(options: {
  proxies?: Proxy[]
  selectedIds?: number[]
} = {}) {
  const proxies = ref(options.proxies ?? [createProxy({ id: 1 }), createProxy({ id: 2 })])
  const selectedProxyIds = ref(new Set(options.selectedIds ?? []))
  const selectedCount = ref(selectedProxyIds.value.size)
  const loadProxies = vi.fn(async () => {})
  const showSuccess = vi.fn()
  const showError = vi.fn()
  const showInfo = vi.fn()

  const composable = useProxyTestingActions({
    proxies,
    selectedProxyIds,
    selectedCount,
    loadProxies,
    getBatchFilters: () => ({
      protocol: 'http',
      status: 'active',
      search: 'edge'
    }),
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key,
    showSuccess,
    showError,
    showInfo
  })

  return {
    composable,
    proxies,
    loadProxies,
    selectedCount,
    selectedProxyIds,
    showSuccess,
    showError,
    showInfo
  }
}

describe('useProxyTestingActions', () => {
  beforeEach(() => {
    listProxies.mockReset()
    testProxy.mockReset()
    checkProxyQuality.mockReset()
  })

  it('tests a single proxy and patches latency state', async () => {
    const setup = createComposable()
    testProxy.mockResolvedValue({
      success: true,
      message: 'ok',
      latency_ms: 88,
      ip_address: '8.8.8.8',
      country: 'United States',
      country_code: 'US'
    })

    await setup.composable.handleTestConnection(setup.proxies.value[0])

    expect(testProxy).toHaveBeenCalledWith(1)
    expect(setup.proxies.value[0].latency_status).toBe('success')
    expect(setup.proxies.value[0].latency_ms).toBe(88)
    expect(setup.showSuccess).toHaveBeenCalledWith(
      'admin.proxies.proxyWorkingWithLatency:{"latency":88}'
    )
    expect(setup.composable.testingProxyIds.value.size).toBe(0)
  })

  it('opens the quality report and patches quality data', async () => {
    const setup = createComposable()
    checkProxyQuality.mockResolvedValue(createQualityResult(1, { score: 73, grade: 'B' }))

    await setup.composable.handleQualityCheck(setup.proxies.value[0])

    expect(checkProxyQuality).toHaveBeenCalledWith(1)
    expect(setup.composable.showQualityReportDialog.value).toBe(true)
    expect(setup.composable.qualityReportProxy.value?.id).toBe(1)
    expect(setup.proxies.value[0].quality_status).toBe('healthy')
    expect(setup.proxies.value[0].quality_score).toBe(73)
    expect(setup.proxies.value[0].latency_ms).toBe(120)
    expect(setup.showSuccess).toHaveBeenCalledWith(
      'admin.proxies.qualityCheckDone:{"score":73,"grade":"B"}'
    )
  })

  it('batch tests selected proxies and reloads the list', async () => {
    const setup = createComposable({ selectedIds: [1, 2] })
    testProxy.mockResolvedValue({
      success: true,
      message: 'ok',
      latency_ms: 50
    })

    await setup.composable.handleBatchTest()

    expect(testProxy).toHaveBeenCalledTimes(2)
    expect(testProxy).toHaveBeenNthCalledWith(1, 1)
    expect(testProxy).toHaveBeenNthCalledWith(2, 2)
    expect(setup.showSuccess).toHaveBeenCalledWith(
      'admin.proxies.batchTestDone:{"count":2}'
    )
    expect(setup.loadProxies).toHaveBeenCalledTimes(1)
  })

  it('loads all matching proxies for batch quality checks and summarizes results', async () => {
    const setup = createComposable({ selectedIds: [] })
    listProxies.mockResolvedValue({
      items: [createProxy({ id: 1 }), createProxy({ id: 2 })],
      total: 2,
      page: 1,
      page_size: 200,
      pages: 1
    })
    checkProxyQuality
      .mockResolvedValueOnce(createQualityResult(1, { warn_count: 1, score: 65, grade: 'C' }))
      .mockResolvedValueOnce(
        createQualityResult(2, { challenge_count: 1, score: 40, grade: 'D' })
      )

    await setup.composable.handleBatchQualityCheck()

    expect(listProxies).toHaveBeenCalledWith(1, 200, {
      protocol: 'http',
      status: 'active',
      search: 'edge'
    })
    expect(setup.showSuccess).toHaveBeenCalledWith(
      'admin.proxies.batchQualityDone:{"count":2,"healthy":0,"warn":1,"challenge":1,"failed":0}'
    )
    expect(setup.loadProxies).toHaveBeenCalledTimes(1)
  })

  it('uses resolved request messages for single and batch failures', async () => {
    const setup = createComposable({ selectedIds: [1] })
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    testProxy.mockRejectedValueOnce(new Error('test unavailable'))
    await setup.composable.handleTestConnection(setup.proxies.value[0])

    checkProxyQuality.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'quality blocked'
        }
      }
    })
    await setup.composable.handleQualityCheck(setup.proxies.value[0])

    const batchTest = createComposable({ selectedIds: [] })
    const batchQuality = createComposable({ selectedIds: [] })

    listProxies.mockRejectedValueOnce(new Error('batch list unavailable'))
    batchTest.selectedProxyIds.value = new Set()
    batchTest.selectedCount.value = 0
    await batchTest.composable.handleBatchTest()

    listProxies.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'batch quality blocked'
        }
      }
    })
    batchQuality.selectedProxyIds.value = new Set()
    batchQuality.selectedCount.value = 0
    await batchQuality.composable.handleBatchQualityCheck()

    expect(setup.showError).toHaveBeenNthCalledWith(1, 'test unavailable')
    expect(setup.showError).toHaveBeenNthCalledWith(2, 'quality blocked')
    expect(batchTest.showError).toHaveBeenCalledWith('batch list unavailable')
    expect(batchQuality.showError).toHaveBeenCalledWith('batch quality blocked')
    expect(consoleSpy).toHaveBeenCalledTimes(4)
  })

  it('keeps per-proxy testing pending until the newest request finishes and applies only the latest result', async () => {
    const setup = createComposable()
    const firstRequest = createDeferred<{
      success: boolean
      message: string
      latency_ms: number
    }>()
    const secondRequest = createDeferred<{
      success: boolean
      message: string
      latency_ms: number
    }>()

    testProxy
      .mockImplementationOnce(() => firstRequest.promise)
      .mockImplementationOnce(() => secondRequest.promise)

    const firstRun = setup.composable.handleTestConnection(setup.proxies.value[0])
    const secondRun = setup.composable.handleTestConnection(setup.proxies.value[0])

    expect(setup.composable.testingProxyIds.value.has(1)).toBe(true)

    firstRequest.resolve({
      success: true,
      message: 'old',
      latency_ms: 45
    })
    await firstRun

    expect(setup.composable.testingProxyIds.value.has(1)).toBe(true)
    expect(setup.proxies.value[0].latency_ms).toBeUndefined()
    expect(setup.showSuccess).not.toHaveBeenCalled()

    secondRequest.resolve({
      success: true,
      message: 'new',
      latency_ms: 67
    })
    await secondRun

    expect(setup.composable.testingProxyIds.value.has(1)).toBe(false)
    expect(setup.proxies.value[0].latency_ms).toBe(67)
    expect(setup.showSuccess).toHaveBeenCalledTimes(1)
    expect(setup.showSuccess).toHaveBeenCalledWith(
      'admin.proxies.proxyWorkingWithLatency:{"latency":67}'
    )
  })

  it('keeps per-proxy quality checking pending until the newest request finishes and applies only the latest report', async () => {
    const setup = createComposable()
    const firstRequest = createDeferred<ProxyQualityCheckResult>()
    const secondRequest = createDeferred<ProxyQualityCheckResult>()

    checkProxyQuality
      .mockImplementationOnce(() => firstRequest.promise)
      .mockImplementationOnce(() => secondRequest.promise)

    const firstRun = setup.composable.handleQualityCheck(setup.proxies.value[0])
    const secondRun = setup.composable.handleQualityCheck(setup.proxies.value[0])

    expect(setup.composable.qualityCheckingProxyIds.value.has(1)).toBe(true)

    firstRequest.resolve(createQualityResult(1, { score: 52, grade: 'D' }))
    await firstRun

    expect(setup.composable.qualityCheckingProxyIds.value.has(1)).toBe(true)
    expect(setup.composable.showQualityReportDialog.value).toBe(false)
    expect(setup.proxies.value[0].quality_score).toBeUndefined()
    expect(setup.showSuccess).not.toHaveBeenCalled()

    secondRequest.resolve(createQualityResult(1, { score: 96, grade: 'A' }))
    await secondRun

    expect(setup.composable.qualityCheckingProxyIds.value.has(1)).toBe(false)
    expect(setup.composable.showQualityReportDialog.value).toBe(true)
    expect(setup.composable.qualityReport?.value?.score).toBe(96)
    expect(setup.proxies.value[0].quality_score).toBe(96)
    expect(setup.showSuccess).toHaveBeenCalledTimes(1)
    expect(setup.showSuccess).toHaveBeenCalledWith(
      'admin.proxies.qualityCheckDone:{"score":96,"grade":"A"}'
    )
  })
})
