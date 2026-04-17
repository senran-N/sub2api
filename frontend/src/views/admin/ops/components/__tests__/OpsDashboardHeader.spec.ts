import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, reactive } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsDashboardHeader from '../OpsDashboardHeader.vue'

const mockGetRealtimeTrafficSummary = vi.fn()
const mockGetAllGroups = vi.fn()

const adminSettingsStore = reactive({
  opsRealtimeMonitoringEnabled: true,
  setOpsRealtimeMonitoringEnabledLocal(value: boolean) {
    this.opsRealtimeMonitoringEnabled = value
  }
})

vi.mock('@/api', () => ({
  adminAPI: {
    groups: {
      getAll: (...args: any[]) => mockGetAllGroups(...args)
    }
  }
}))

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getRealtimeTrafficSummary: (...args: any[]) => mockGetRealtimeTrafficSummary(...args)
  }
}))

vi.mock('@/stores', () => ({
  useAdminSettingsStore: () => adminSettingsStore
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: {
      type: [String, Number, Boolean, Object],
      default: ''
    },
    options: {
      type: Array,
      default: () => []
    }
  },
  emits: ['update:modelValue'],
  template: '<div class="select-stub" />'
})

const HelpTooltipStub = defineComponent({
  name: 'HelpTooltipStub',
  template: '<div class="help-tooltip-stub" />'
})

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  template: '<div class="base-dialog-stub"><slot /></div>'
})

const IconStub = defineComponent({
  name: 'IconStub',
  template: '<span class="icon-stub" />'
})

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function makeRealtimeSummary(platform: string, currentQps: number, currentTps: number) {
  return {
    window: '1min',
    start_time: '2026-04-17T00:00:00Z',
    end_time: '2026-04-17T00:01:00Z',
    platform,
    group_id: null,
    qps: {
      current: currentQps,
      peak: currentQps + 1,
      avg: currentQps + 2
    },
    tps: {
      current: currentTps,
      peak: currentTps + 1,
      avg: currentTps + 2
    }
  }
}

function makeRuntimeObservability() {
  return {
    summary: {
      scheduling_runtime_kernel: {
        avg_fetched_accounts_per_page: 2,
        acquire_success_rate: 0.5,
        wait_plan_success_rate: 0.25
      },
      idempotency: {
        avg_processing_duration_ms: 12
      }
    }
  }
}

const overview = {
  health_score: 98,
  request_count_total: 10,
  token_consumed: 20,
  qps: { avg: 1.2 },
  tps: { avg: 3.4 },
  sla: 0.99,
  error_rate: 0.01,
  upstream_error_rate: 0.02,
  duration: {},
  ttft: {},
  system_metrics: {
    db_ok: true,
    redis_ok: true
  },
  job_heartbeats: []
}

describe('OpsDashboardHeader', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    adminSettingsStore.opsRealtimeMonitoringEnabled = true
    mockGetAllGroups.mockResolvedValue([])
  })

  it('keeps the latest realtime summary when platform switches quickly', async () => {
    const initialRealtime = {
      summary: makeRealtimeSummary('openai', 1.1, 2.2),
      runtime_observability: makeRuntimeObservability()
    }
    const slowRealtime = deferred<any>()
    const fastRealtime = deferred<any>()

    mockGetRealtimeTrafficSummary
      .mockResolvedValueOnce(initialRealtime)
      .mockResolvedValueOnce(initialRealtime)
      .mockReturnValueOnce(slowRealtime.promise)
      .mockReturnValueOnce(fastRealtime.promise)

    const wrapper = mount(OpsDashboardHeader, {
      props: {
        overview,
        platform: 'openai',
        groupId: null,
        timeRange: '1h',
        queryMode: 'auto',
        loading: false,
        lastUpdated: new Date('2026-04-17T00:00:00Z'),
        thresholds: null,
        autoRefreshEnabled: false,
        autoRefreshCountdown: 0,
        fullscreen: false,
        customStartTime: null,
        customEndTime: null
      },
      global: {
        stubs: {
          Select: SelectStub,
          HelpTooltip: HelpTooltipStub,
          BaseDialog: BaseDialogStub,
          Icon: IconStub
        }
      }
    })

    await flushPromises()

    await wrapper.setProps({ platform: 'anthropic' })
    await wrapper.setProps({ platform: 'gemini' })

    fastRealtime.resolve({
      summary: makeRealtimeSummary('gemini', 9.9, 8.8),
      runtime_observability: makeRuntimeObservability()
    })
    await flushPromises()

    slowRealtime.resolve({
      summary: makeRealtimeSummary('anthropic', 3.3, 4.4),
      runtime_observability: makeRuntimeObservability()
    })
    await flushPromises()

    const realtimePanelText = wrapper.get('.ops-dashboard-header__realtime').text()
    expect(realtimePanelText).toContain('9.9')
    expect(realtimePanelText).toContain('8.8')
    expect(realtimePanelText).not.toContain('3.3')
    expect(realtimePanelText).not.toContain('4.4')
    expect(mockGetRealtimeTrafficSummary).toHaveBeenNthCalledWith(3, '1min', 'anthropic', null)
    expect(mockGetRealtimeTrafficSummary).toHaveBeenNthCalledWith(4, '1min', 'gemini', null)
  })
})
