import { ref, type Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Proxy, ProxyQualityCheckResult } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  applyProxyConnectivityFromQualityResult,
  applyProxyLatencyResult,
  applyProxyQualityResult,
  createProxyBatchQualitySummary,
  recordProxyBatchQualityResult,
  type ProxyBatchQualitySummary
} from './proxyQuality'

interface ProxyBatchFilters {
  protocol?: string
  status?: 'active' | 'inactive'
  search?: string
}

interface ProxyTestingActionsOptions {
  proxies: Ref<Proxy[]>
  selectedProxyIds: Ref<Set<number>>
  selectedCount: Readonly<Ref<number>>
  loadProxies: () => void | Promise<void>
  getBatchFilters: () => ProxyBatchFilters
  t: (key: string, params?: Record<string, unknown>) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
  showInfo: (message: string) => void
}

function addPendingId(target: Ref<Set<number>>, proxyId: number) {
  const next = new Set(target.value)
  next.add(proxyId)
  target.value = next
}

function removePendingId(target: Ref<Set<number>>, proxyId: number) {
  const next = new Set(target.value)
  next.delete(proxyId)
  target.value = next
}

export function useProxyTestingActions(options: ProxyTestingActionsOptions) {
  const testingProxyIds = ref<Set<number>>(new Set())
  const qualityCheckingProxyIds = ref<Set<number>>(new Set())
  const batchTesting = ref(false)
  const batchQualityChecking = ref(false)
  const showQualityReportDialog = ref(false)
  const qualityReportProxy = ref<Proxy | null>(null)
  const qualityReport = ref<ProxyQualityCheckResult | null>(null)

  const withProxy = (proxyId: number, callback: (proxy: Proxy) => void) => {
    const target = options.proxies.value.find((proxy) => proxy.id === proxyId)
    if (!target) {
      return
    }

    callback(target)
  }

  const runProxyTest = async (proxyId: number, notify: boolean) => {
    addPendingId(testingProxyIds, proxyId)

    try {
      const result = await adminAPI.proxies.testProxy(proxyId)
      withProxy(proxyId, (proxy) => {
        applyProxyLatencyResult(proxy, result)
      })

      if (notify) {
        if (result.success) {
          const message = result.latency_ms
            ? options.t('admin.proxies.proxyWorkingWithLatency', { latency: result.latency_ms })
            : options.t('admin.proxies.proxyWorking')
          options.showSuccess(message)
        } else {
          options.showError(result.message || options.t('admin.proxies.proxyTestFailed'))
        }
      }

      return result
    } catch (error: unknown) {
      const message = resolveRequestErrorMessage(error, options.t('admin.proxies.failedToTest'))

      withProxy(proxyId, (proxy) => {
        applyProxyLatencyResult(proxy, { success: false, message })
      })

      if (notify) {
        options.showError(message)
      }

      console.error('Error testing proxy:', error)
      return null
    } finally {
      removePendingId(testingProxyIds, proxyId)
    }
  }

  const handleTestConnection = async (proxy: Proxy) => {
    await runProxyTest(proxy.id, true)
  }

  const handleQualityCheck = async (proxy: Proxy) => {
    addPendingId(qualityCheckingProxyIds, proxy.id)

    try {
      const result = await adminAPI.proxies.checkProxyQuality(proxy.id)

      withProxy(proxy.id, (target) => {
        applyProxyConnectivityFromQualityResult(target, result)
        applyProxyQualityResult(target, result)
      })

      qualityReportProxy.value = proxy
      qualityReport.value = result
      showQualityReportDialog.value = true

      options.showSuccess(
        options.t('admin.proxies.qualityCheckDone', {
          score: result.score,
          grade: result.grade
        })
      )
    } catch (error: unknown) {
      const message = resolveRequestErrorMessage(
        error,
        options.t('admin.proxies.qualityCheckFailed')
      )
      options.showError(message)
      console.error('Error checking proxy quality:', error)
    } finally {
      removePendingId(qualityCheckingProxyIds, proxy.id)
    }
  }

  const runBatchProxyQualityChecks = async (
    ids: number[]
  ): Promise<ProxyBatchQualitySummary> => {
    if (ids.length === 0) {
      return createProxyBatchQualitySummary(0)
    }

    const concurrency = 3
    const summary = createProxyBatchQualitySummary(ids.length)
    let index = 0

    const worker = async () => {
      while (index < ids.length) {
        const current = ids[index]
        index += 1
        addPendingId(qualityCheckingProxyIds, current)

        try {
          const result = await adminAPI.proxies.checkProxyQuality(current)

          withProxy(current, (proxy) => {
            applyProxyConnectivityFromQualityResult(proxy, result)
            applyProxyQualityResult(proxy, result)
          })

          recordProxyBatchQualityResult(summary, result)
        } catch {
          summary.failed += 1
        } finally {
          removePendingId(qualityCheckingProxyIds, current)
        }
      }
    }

    const workers = Array.from({ length: Math.min(concurrency, ids.length) }, () => worker())
    await Promise.all(workers)
    return summary
  }

  const closeQualityReportDialog = () => {
    showQualityReportDialog.value = false
    qualityReportProxy.value = null
    qualityReport.value = null
  }

  const fetchAllProxiesForBatch = async (): Promise<Proxy[]> => {
    const pageSize = 200
    const result: Proxy[] = []
    let page = 1
    let totalPages = 1

    while (page <= totalPages) {
      const response = await adminAPI.proxies.list(page, pageSize, options.getBatchFilters())
      result.push(...response.items)
      totalPages = response.pages || 1
      page += 1
    }

    return result
  }

  const resolveBatchTargetIds = async () => {
    if (options.selectedCount.value > 0) {
      return Array.from(options.selectedProxyIds.value)
    }

    const allProxies = await fetchAllProxiesForBatch()
    return allProxies.map((proxy) => proxy.id)
  }

  const runBatchProxyTests = async (ids: number[]) => {
    if (ids.length === 0) {
      return
    }

    const concurrency = 5
    let index = 0

    const worker = async () => {
      while (index < ids.length) {
        const current = ids[index]
        index += 1
        await runProxyTest(current, false)
      }
    }

    const workers = Array.from({ length: Math.min(concurrency, ids.length) }, () => worker())
    await Promise.all(workers)
  }

  const handleBatchTest = async () => {
    if (batchTesting.value) {
      return
    }

    batchTesting.value = true
    try {
      const ids = await resolveBatchTargetIds()
      if (ids.length === 0) {
        options.showInfo(options.t('admin.proxies.batchTestEmpty'))
        return
      }

      await runBatchProxyTests(ids)
      options.showSuccess(options.t('admin.proxies.batchTestDone', { count: ids.length }))
      await options.loadProxies()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.batchTestFailed'))
      )
      console.error('Error batch testing proxies:', error)
    } finally {
      batchTesting.value = false
    }
  }

  const handleBatchQualityCheck = async () => {
    if (batchQualityChecking.value) {
      return
    }

    batchQualityChecking.value = true
    try {
      const ids = await resolveBatchTargetIds()
      if (ids.length === 0) {
        options.showInfo(options.t('admin.proxies.batchQualityEmpty'))
        return
      }

      const summary = await runBatchProxyQualityChecks(ids)
      options.showSuccess(
        options.t('admin.proxies.batchQualityDone', {
          count: summary.total,
          healthy: summary.healthy,
          warn: summary.warn,
          challenge: summary.challenge,
          failed: summary.failed
        })
      )
      await options.loadProxies()
    } catch (error: unknown) {
      options.showError(
        resolveRequestErrorMessage(error, options.t('admin.proxies.batchQualityFailed'))
      )
      console.error('Error batch checking quality:', error)
    } finally {
      batchQualityChecking.value = false
    }
  }

  return {
    batchQualityChecking,
    batchTesting,
    closeQualityReportDialog,
    handleBatchQualityCheck,
    handleBatchTest,
    handleQualityCheck,
    handleTestConnection,
    qualityCheckingProxyIds,
    qualityReport,
    qualityReportProxy,
    showQualityReportDialog,
    testingProxyIds
  }
}
