import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsConcurrencyCard from '../OpsConcurrencyCard.vue'

const mockGetConcurrencyStats = vi.fn()
const mockGetAccountAvailabilityStats = vi.fn()
const mockGetUserConcurrencyStats = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getConcurrencyStats: (...args: any[]) => mockGetConcurrencyStats(...args),
    getAccountAvailabilityStats: (...args: any[]) => mockGetAccountAvailabilityStats(...args),
    getUserConcurrencyStats: (...args: any[]) => mockGetUserConcurrencyStats(...args),
  },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, any>) => {
        if (key === 'admin.ops.concurrency.totalRows' && params) {
          return `rows:${params.count}`
        }
        return key
      },
    }),
  }
})

const runtimeObservability = {
  summary: {
    scheduling_runtime_kernel: {
      avg_fetched_accounts_per_page: 4,
      acquire_success_rate: 0.9,
      wait_plan_success_rate: 0.8,
    },
    idempotency: {
      avg_processing_duration_ms: 20,
    },
  },
  scheduling_runtime_kernel: {
    runtime_wait_plan_attempts: 1,
    index_page_fetches: 1,
  },
}

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function makeConcurrencyStats(platform: string, currentInUse: number) {
  return {
    enabled: true,
    runtime_observability: runtimeObservability,
    platform: {
      [platform]: {
        max_capacity: 10,
        current_in_use: currentInUse,
        waiting_in_queue: 0,
      },
    },
    group: {},
    account: {},
  }
}

function makeAvailabilityStats(platform: string, availableCount: number) {
  return {
    enabled: true,
    runtime_observability: runtimeObservability,
    platform: {
      [platform]: {
        total_accounts: 5,
        available_count: availableCount,
        rate_limit_count: 0,
        error_count: 0,
      },
    },
    group: {},
    account: {},
  }
}

const HelpTooltipStub = defineComponent({
  name: 'HelpTooltipStub',
  template: '<div class="help-tooltip-stub" />',
})

describe('OpsConcurrencyCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetUserConcurrencyStats.mockResolvedValue({
      enabled: true,
      runtime_observability: runtimeObservability,
      user: {},
    })
  })

  it('keeps the latest refresh result when refreshes overlap', async () => {
    mockGetConcurrencyStats.mockResolvedValueOnce(makeConcurrencyStats('openai', 2))
    mockGetAccountAvailabilityStats.mockResolvedValueOnce(makeAvailabilityStats('openai', 4))

    const wrapper = mount(OpsConcurrencyCard, {
      props: {
        refreshToken: 0,
      },
      global: {
        stubs: {
          HelpTooltip: HelpTooltipStub,
        },
      },
    })

    await flushPromises()

    const slowConcurrency = deferred<any>()
    const slowAvailability = deferred<any>()
    const fastConcurrency = deferred<any>()
    const fastAvailability = deferred<any>()

    mockGetConcurrencyStats
      .mockReturnValueOnce(slowConcurrency.promise)
      .mockReturnValueOnce(fastConcurrency.promise)
    mockGetAccountAvailabilityStats
      .mockReturnValueOnce(slowAvailability.promise)
      .mockReturnValueOnce(fastAvailability.promise)

    await wrapper.setProps({ refreshToken: 1 })
    await wrapper.setProps({ refreshToken: 2 })

    fastConcurrency.resolve(makeConcurrencyStats('gemini', 7))
    fastAvailability.resolve(makeAvailabilityStats('gemini', 5))
    await flushPromises()

    slowConcurrency.resolve(makeConcurrencyStats('openai', 1))
    slowAvailability.resolve(makeAvailabilityStats('openai', 1))
    await flushPromises()

    const cardText = wrapper.text()
    expect(cardText).toContain('GEMINI')
    expect(cardText).not.toContain('OPENAI')
  })
})
