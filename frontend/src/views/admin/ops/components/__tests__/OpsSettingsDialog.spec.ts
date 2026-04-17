import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsSettingsDialog from '../OpsSettingsDialog.vue'

const mockGetAlertRuntimeSettings = vi.fn()
const mockGetEmailNotificationConfig = vi.fn()
const mockGetAdvancedSettings = vi.fn()
const mockGetMetricThresholds = vi.fn()
const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getAlertRuntimeSettings: (...args: any[]) => mockGetAlertRuntimeSettings(...args),
    getEmailNotificationConfig: (...args: any[]) => mockGetEmailNotificationConfig(...args),
    getAdvancedSettings: (...args: any[]) => mockGetAdvancedSettings(...args),
    getMetricThresholds: (...args: any[]) => mockGetMetricThresholds(...args),
    updateAlertRuntimeSettings: vi.fn(),
    updateEmailNotificationConfig: vi.fn(),
    updateAdvancedSettings: vi.fn(),
    updateMetricThresholds: vi.fn(),
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
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

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: '' },
    width: { type: String, default: '' },
  },
  emits: ['close'],
  template: '<div v-if="show" class="base-dialog-stub"><slot /><slot name="footer" /></div>',
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: {
      type: [String, Number, Boolean, Object],
      default: '',
    },
    options: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['update:modelValue', 'change'],
  template: '<div class="select-stub" />',
})

const ToggleStub = defineComponent({
  name: 'ToggleStub',
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue'],
  template: '<div class="toggle-stub" />',
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

function makeRuntimeSettings(evaluationIntervalSeconds: number) {
  return {
    evaluation_interval_seconds: evaluationIntervalSeconds,
    distributed_lock: {
      enabled: true,
      key: 'ops-lock',
      ttl_seconds: 30,
    },
    silencing: {
      enabled: false,
      global_until_rfc3339: '',
      global_reason: '',
      entries: [],
    },
    thresholds: {
      sla_percent_min: 99.5,
      ttft_p99_ms_max: 500,
      request_error_rate_percent_max: 5,
      upstream_error_rate_percent_max: 5,
    },
  }
}

function makeEmailConfig(recipient: string) {
  return {
    alert: {
      enabled: true,
      recipients: [recipient],
      min_severity: 'warning',
      rate_limit_per_hour: 10,
      batching_window_seconds: 60,
      include_resolved_alerts: false,
    },
    report: {
      enabled: false,
      recipients: [],
      daily_summary_enabled: false,
      daily_summary_schedule: '',
      weekly_summary_enabled: false,
      weekly_summary_schedule: '',
      error_digest_enabled: false,
      error_digest_schedule: '',
      error_digest_min_count: 1,
      account_health_enabled: false,
      account_health_schedule: '',
      account_health_error_rate_threshold: 5,
    },
  }
}

function makeAdvancedSettings(autoRefreshIntervalSeconds: number) {
  return {
    data_retention: {
      cleanup_enabled: true,
      cleanup_schedule: '0 0 * * *',
      error_log_retention_days: 7,
      minute_metrics_retention_days: 14,
      hourly_metrics_retention_days: 30,
    },
    aggregation: {
      aggregation_enabled: true,
    },
    ignore_count_tokens_errors: false,
    ignore_context_canceled: false,
    ignore_no_available_accounts: false,
    ignore_invalid_api_key_errors: false,
    ignore_insufficient_balance_errors: false,
    display_openai_token_stats: true,
    display_alert_events: true,
    auto_refresh_enabled: true,
    auto_refresh_interval_seconds: autoRefreshIntervalSeconds,
  }
}

describe('OpsSettingsDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('keeps the latest reopen result when settings requests overlap', async () => {
    const firstRuntime = deferred<any>()
    const firstEmail = deferred<any>()
    const firstAdvanced = deferred<any>()
    const firstThresholds = deferred<any>()
    const secondRuntime = deferred<any>()
    const secondEmail = deferred<any>()
    const secondAdvanced = deferred<any>()
    const secondThresholds = deferred<any>()

    mockGetAlertRuntimeSettings
      .mockReturnValueOnce(firstRuntime.promise)
      .mockReturnValueOnce(secondRuntime.promise)
    mockGetEmailNotificationConfig
      .mockReturnValueOnce(firstEmail.promise)
      .mockReturnValueOnce(secondEmail.promise)
    mockGetAdvancedSettings
      .mockReturnValueOnce(firstAdvanced.promise)
      .mockReturnValueOnce(secondAdvanced.promise)
    mockGetMetricThresholds
      .mockReturnValueOnce(firstThresholds.promise)
      .mockReturnValueOnce(secondThresholds.promise)

    const wrapper = mount(OpsSettingsDialog, {
      props: {
        show: false,
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          Toggle: ToggleStub,
        },
      },
    })

    await flushPromises()

    await wrapper.setProps({ show: true })
    await flushPromises()
    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({ show: true })
    await flushPromises()

    secondRuntime.resolve(makeRuntimeSettings(99))
    secondEmail.resolve(makeEmailConfig('latest-alert@example.com'))
    secondAdvanced.resolve(makeAdvancedSettings(45))
    secondThresholds.resolve({
      sla_percent_min: 97.5,
      ttft_p99_ms_max: 400,
      request_error_rate_percent_max: 3,
      upstream_error_rate_percent_max: 4,
    })
    await flushPromises()

    firstRuntime.resolve(makeRuntimeSettings(11))
    firstEmail.resolve(makeEmailConfig('stale-alert@example.com'))
    firstAdvanced.resolve(makeAdvancedSettings(15))
    firstThresholds.resolve({
      sla_percent_min: 91.5,
      ttft_p99_ms_max: 900,
      request_error_rate_percent_max: 9,
      upstream_error_rate_percent_max: 8,
    })
    await flushPromises()

    const inputs = wrapper.findAll('input')
    expect((inputs[0]?.element as HTMLInputElement).value).toBe('99')
    expect(wrapper.text()).toContain('latest-alert@example.com')
    expect(wrapper.text()).not.toContain('stale-alert@example.com')
  })
})
