import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsRuntimeSettingsCard from '../OpsRuntimeSettingsCard.vue'

const mockGetAlertRuntimeSettings = vi.fn()
const mockUpdateAlertRuntimeSettings = vi.fn()
const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getAlertRuntimeSettings: (...args: any[]) => mockGetAlertRuntimeSettings(...args),
    updateAlertRuntimeSettings: (...args: any[]) => mockUpdateAlertRuntimeSettings(...args),
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

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function makeRuntimeSettings(lockKey: string) {
  return {
    evaluation_interval_seconds: 60,
    distributed_lock: {
      enabled: true,
      key: lockKey,
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

describe('OpsRuntimeSettingsCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUpdateAlertRuntimeSettings.mockResolvedValue(undefined)
  })

  it('keeps the latest runtime settings when initial load overlaps with refresh', async () => {
    const slowSettings = deferred<any>()
    const fastSettings = deferred<any>()
    mockGetAlertRuntimeSettings
      .mockReturnValueOnce(slowSettings.promise)
      .mockReturnValueOnce(fastSettings.promise)

    const wrapper = mount(OpsRuntimeSettingsCard, {
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    const refreshButton = wrapper.find('.ops-runtime-settings-card__refresh')
    await refreshButton.trigger('click')
    await flushPromises()

    fastSettings.resolve(makeRuntimeSettings('ops:latest'))
    await flushPromises()

    slowSettings.resolve(makeRuntimeSettings('ops:stale'))
    await flushPromises()

    expect(wrapper.text()).toContain('ops:latest')
    expect(wrapper.text()).not.toContain('ops:stale')
  })
})
