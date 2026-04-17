import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsEmailNotificationCard from '../OpsEmailNotificationCard.vue'

const mockGetEmailNotificationConfig = vi.fn()
const mockUpdateEmailNotificationConfig = vi.fn()
const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getEmailNotificationConfig: (...args: any[]) => mockGetEmailNotificationConfig(...args),
    updateEmailNotificationConfig: (...args: any[]) => mockUpdateEmailNotificationConfig(...args),
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
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: {
      type: [String, Number, Boolean, Object],
      default: '',
    },
  },
  emits: ['update:modelValue'],
  template: '<div class="select-stub" />',
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

function makeConfig(minSeverity: 'warning' | 'critical' | 'info') {
  return {
    alert: {
      enabled: true,
      recipients: ['ops@example.com'],
      min_severity: minSeverity,
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

describe('OpsEmailNotificationCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetEmailNotificationConfig.mockResolvedValue(makeConfig('warning'))
    mockUpdateEmailNotificationConfig.mockResolvedValue(undefined)
  })

  it('prefers backend detail when loading email notification config fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockGetEmailNotificationConfig.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'email config detail error'
        }
      },
      message: 'generic email config error'
    })

    mount(OpsEmailNotificationCard, {
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
        },
      },
    })

    await flushPromises()

    expect(showError).toHaveBeenCalledWith('email config detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })

  it('keeps the latest refresh result when config reloads overlap', async () => {
    const slowConfig = deferred<any>()
    const fastConfig = deferred<any>()
    mockGetEmailNotificationConfig.mockReset()
    mockGetEmailNotificationConfig
      .mockReturnValueOnce(slowConfig.promise)
      .mockReturnValueOnce(fastConfig.promise)

    const wrapper = mount(OpsEmailNotificationCard, {
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
        },
      },
    })

    const refreshButton = wrapper.find('.ops-email-notification-card__refresh')
    await refreshButton.trigger('click')
    await flushPromises()

    fastConfig.resolve(makeConfig('info'))
    await flushPromises()

    slowConfig.resolve(makeConfig('critical'))
    await flushPromises()

    expect(wrapper.text()).toContain('info')
    expect(wrapper.text()).not.toContain('critical')
  })
})
