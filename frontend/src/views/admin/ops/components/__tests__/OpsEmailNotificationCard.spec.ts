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

describe('OpsEmailNotificationCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetEmailNotificationConfig.mockResolvedValue({
      alert: {
        enabled: true,
        recipients: ['ops@example.com'],
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
    })
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
})
