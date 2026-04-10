import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, reactive } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsDashboard from '../OpsDashboard.vue'

const mockGetAdvancedSettings = vi.fn()
const mockListAlertRules = vi.fn()
const mockGetMetricThresholds = vi.fn()
const mockGetDashboardSnapshotV2 = vi.fn()
const mockGetThroughputTrend = vi.fn()
const mockGetLatencyHistogram = vi.fn()
const mockGetErrorDistribution = vi.fn()
const mockGetErrorTrend = vi.fn()
const mockGetDashboardOverview = vi.fn()
const adminSettingsState = reactive({
  opsMonitoringEnabled: true,
  opsQueryModeDefault: 'auto',
  fetch: vi.fn().mockResolvedValue(undefined),
})
const showError = vi.fn()
const routerReplace = vi.fn()
const route = reactive({
  query: {} as Record<string, any>,
})

vi.mock('@vueuse/core', () => ({
  useDebounceFn: (fn: (...args: any[]) => any) => fn,
  useIntervalFn: () => ({
    pause: vi.fn(),
    resume: vi.fn(),
  }),
}))

vi.mock('@/api/admin/ops', () => {
  const opsAPI = {
    getAdvancedSettings: (...args: any[]) => mockGetAdvancedSettings(...args),
    listAlertRules: (...args: any[]) => mockListAlertRules(...args),
    getMetricThresholds: (...args: any[]) => mockGetMetricThresholds(...args),
    getDashboardSnapshotV2: (...args: any[]) => mockGetDashboardSnapshotV2(...args),
    getThroughputTrend: (...args: any[]) => mockGetThroughputTrend(...args),
    getLatencyHistogram: (...args: any[]) => mockGetLatencyHistogram(...args),
    getErrorDistribution: (...args: any[]) => mockGetErrorDistribution(...args),
    getErrorTrend: (...args: any[]) => mockGetErrorTrend(...args),
    getDashboardOverview: (...args: any[]) => mockGetDashboardOverview(...args),
  }

  return {
    opsAPI,
    default: opsAPI,
  }
})

vi.mock('@/stores', () => ({
  useAdminSettingsStore: () => adminSettingsState,
  useAppStore: () => ({
    showError,
  }),
}))

vi.mock('vue-router', () => ({
  useRoute: () => route,
  useRouter: () => ({
    replace: (...args: any[]) => routerReplace(...args),
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const AppLayoutStub = defineComponent({
  name: 'AppLayoutStub',
  template: '<div class="app-layout-stub"><slot /></div>',
})

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: '' },
  },
  emits: ['close'],
  template: '<div v-if="show" class="base-dialog-stub"><slot /></div>',
})

const GenericStub = defineComponent({
  name: 'GenericStub',
  template: '<div class="generic-stub" />',
})

const HeaderStub = defineComponent({
  name: 'OpsDashboardHeader',
  emits: ['open-alert-rules'],
  template: '<div class="ops-dashboard-header-stub" />',
})

function createSnapshotResponse() {
  return {
    generated_at: '2026-04-11T00:00:00Z',
    overview: {
      start_time: '2026-04-11T00:00:00Z',
      end_time: '2026-04-11T01:00:00Z',
      platform: '',
      success_count: 0,
      error_count_total: 0,
      business_limited_count: 0,
      error_count_sla: 0,
      request_count_total: 0,
      request_count_sla: 0,
      token_consumed: 0,
      sla: 1,
      error_rate: 0,
      upstream_error_rate: 0,
      upstream_error_count_excl_429_529: 0,
      upstream_429_count: 0,
      upstream_529_count: 0,
      qps: { current: 0, peak: 0, avg: 0 },
      tps: { current: 0, peak: 0, avg: 0 },
      duration: {},
      ttft: {},
    },
    throughput_trend: {
      bucket: '1m',
      points: [],
      by_platform: [],
      top_groups: [],
    },
    error_trend: {
      bucket: '1m',
      points: [],
    },
  }
}

function mountDashboard() {
  return mount(OpsDashboard, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        BaseDialog: BaseDialogStub,
        OpsDashboardHeader: HeaderStub,
        OpsDashboardSkeleton: GenericStub,
        OpsConcurrencyCard: GenericStub,
        OpsErrorDistributionChart: GenericStub,
        OpsErrorTrendChart: GenericStub,
        OpsLatencyChart: GenericStub,
        OpsThroughputTrendChart: GenericStub,
        OpsSwitchRateTrendChart: GenericStub,
        OpsAlertEventsCard: GenericStub,
        OpsOpenAITokenStatsCard: GenericStub,
        OpsSystemLogTable: GenericStub,
        OpsSettingsDialog: GenericStub,
        OpsErrorDetailsModal: GenericStub,
        OpsErrorDetailModal: GenericStub,
        OpsRequestDetailsModal: GenericStub,
        OpsAlertRulesCard: GenericStub,
      },
    },
  })
}

describe('OpsDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    route.query = {}
    adminSettingsState.opsMonitoringEnabled = true
    adminSettingsState.opsQueryModeDefault = 'auto'
    adminSettingsState.fetch.mockResolvedValue(undefined)
    routerReplace.mockImplementation(async (payload: any) => {
      route.query = payload?.query ?? {}
    })

    mockGetAdvancedSettings.mockResolvedValue({
      display_alert_events: true,
      display_openai_token_stats: false,
      auto_refresh_enabled: false,
      auto_refresh_interval_seconds: 30,
    })
    mockGetMetricThresholds.mockResolvedValue(null)
    mockGetDashboardSnapshotV2.mockResolvedValue(createSnapshotResponse())
    mockGetThroughputTrend.mockResolvedValue({
      bucket: '1m',
      points: [],
      by_platform: [],
      top_groups: [],
    })
    mockGetLatencyHistogram.mockResolvedValue({
      start_time: '2026-04-11T00:00:00Z',
      end_time: '2026-04-11T01:00:00Z',
      platform: '',
      total_requests: 0,
      buckets: [],
    })
    mockGetErrorDistribution.mockResolvedValue({
      total: 0,
      items: [],
    })
    mockGetErrorTrend.mockResolvedValue({
      bucket: '1m',
      points: [],
    })
    mockGetDashboardOverview.mockResolvedValue(createSnapshotResponse().overview)
  })

  it('无告警规则时在首屏展示基线横幅，并可直接打开规则面板', async () => {
    mockListAlertRules.mockResolvedValue([])

    const wrapper = mountDashboard()
    await flushPromises()

    expect(wrapper.text()).toContain('admin.ops.alertRules.dashboardBaseline.title')

    const button = wrapper.findAll('button').find((item) =>
      item.text().includes('admin.ops.alertRules.dashboardBaseline.action')
    )
    expect(button).toBeDefined()

    await button!.trigger('click')
    await flushPromises()

    expect(wrapper.find('.base-dialog-stub').exists()).toBe(true)
  })

  it('已有告警规则时不显示基线横幅', async () => {
    mockListAlertRules.mockResolvedValue([
      {
        id: 1,
        name: 'baseline-ready',
        enabled: true,
      },
    ])

    const wrapper = mountDashboard()
    await flushPromises()

    expect(wrapper.text()).not.toContain('admin.ops.alertRules.dashboardBaseline.title')
  })
})
