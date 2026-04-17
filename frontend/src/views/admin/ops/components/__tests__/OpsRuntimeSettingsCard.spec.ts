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
    mockGetAlertRuntimeSettings.mockResolvedValue(makeRuntimeSettings('ops:default'))
    mockUpdateAlertRuntimeSettings.mockResolvedValue(makeRuntimeSettings('ops:default'))
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

  it('keeps save ownership when a stale refresh resolves before runtime settings save finishes', async () => {
    const staleRefresh = deferred<any>()
    const saveResponse = deferred<any>()

    const wrapper = mount(OpsRuntimeSettingsCard, {
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    await flushPromises()

    mockGetAlertRuntimeSettings.mockReset()
    mockGetAlertRuntimeSettings.mockReturnValueOnce(staleRefresh.promise)
    mockUpdateAlertRuntimeSettings.mockReset()
    mockUpdateAlertRuntimeSettings.mockReturnValueOnce(saveResponse.promise)

    const refreshButton = wrapper.get('.ops-runtime-settings-card__refresh')
    await refreshButton.trigger('click')

    const editButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.edit')
    expect(editButton).toBeTruthy()
    await editButton!.trigger('click')

    const saveButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.save')
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')

    staleRefresh.resolve(makeRuntimeSettings('ops:stale'))
    await flushPromises()

    const pendingSaveButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.saving')
    expect(pendingSaveButton).toBeTruthy()
    expect(pendingSaveButton!.attributes('disabled')).toBeDefined()
    expect(refreshButton.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('ops:default')
    expect(wrapper.text()).not.toContain('ops:stale')

    saveResponse.resolve(makeRuntimeSettings('ops:saved'))
    await flushPromises()

    expect(wrapper.text()).toContain('ops:saved')
    expect(wrapper.text()).not.toContain('ops:stale')
  })

  it('ignores a stale refresh after runtime settings save applies', async () => {
    const staleRefresh = deferred<any>()
    const saveResponse = deferred<any>()

    const wrapper = mount(OpsRuntimeSettingsCard, {
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    await flushPromises()

    mockGetAlertRuntimeSettings.mockReset()
    mockGetAlertRuntimeSettings.mockReturnValueOnce(staleRefresh.promise)
    mockUpdateAlertRuntimeSettings.mockReset()
    mockUpdateAlertRuntimeSettings.mockReturnValueOnce(saveResponse.promise)

    await wrapper.get('.ops-runtime-settings-card__refresh').trigger('click')

    const editButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.edit')
    expect(editButton).toBeTruthy()
    await editButton!.trigger('click')

    const saveButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.save')
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')

    saveResponse.resolve(makeRuntimeSettings('ops:saved'))
    await flushPromises()

    staleRefresh.resolve(makeRuntimeSettings('ops:stale'))
    await flushPromises()

    expect(wrapper.text()).toContain('ops:saved')
    expect(wrapper.text()).not.toContain('ops:stale')
    expect(showSuccess).toHaveBeenCalledWith('admin.ops.runtime.saveSuccess')
  })
})
