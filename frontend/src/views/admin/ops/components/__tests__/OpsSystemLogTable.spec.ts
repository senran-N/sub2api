import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsSystemLogTable from '../OpsSystemLogTable.vue'

const mockListSystemLogs = vi.fn()
const mockGetSystemLogSinkHealth = vi.fn()
const mockGetRuntimeLogConfig = vi.fn()
const mockUpdateRuntimeLogConfig = vi.fn()
const mockResetRuntimeLogConfig = vi.fn()
const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listSystemLogs: (...args: any[]) => mockListSystemLogs(...args),
    getSystemLogSinkHealth: (...args: any[]) => mockGetSystemLogSinkHealth(...args),
    getRuntimeLogConfig: (...args: any[]) => mockGetRuntimeLogConfig(...args),
    updateRuntimeLogConfig: (...args: any[]) => mockUpdateRuntimeLogConfig(...args),
    resetRuntimeLogConfig: (...args: any[]) => mockResetRuntimeLogConfig(...args),
    cleanupSystemLogs: vi.fn(),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

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

const PaginationStub = defineComponent({
  name: 'PaginationStub',
  template: '<div class="pagination-stub" />',
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

function makeLog(id: number, message: string) {
  return {
    id,
    level: 'info',
    message,
    created_at: '2026-04-17T00:00:00Z',
    request_id: '',
    client_request_id: '',
    user_id: null,
    account_id: null,
    platform: '',
    model: '',
    extra: {},
  }
}

describe('OpsSystemLogTable', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListSystemLogs.mockResolvedValue({
      items: [],
      total: 0,
    })
    mockUpdateRuntimeLogConfig.mockResolvedValue({
      level: 'info',
      enable_sampling: false,
      sampling_initial: 100,
      sampling_thereafter: 100,
      caller: true,
      stacktrace_level: 'error',
      retention_days: 30,
    })
    mockResetRuntimeLogConfig.mockResolvedValue({
      level: 'info',
      enable_sampling: false,
      sampling_initial: 100,
      sampling_thereafter: 100,
      caller: true,
      stacktrace_level: 'error',
      retention_days: 30,
    })
    mockGetSystemLogSinkHealth.mockResolvedValue({
      queue_depth: 0,
      queue_capacity: 0,
      dropped_count: 0,
      write_failed_count: 0,
      written_count: 0,
      avg_write_delay_ms: 0,
      last_error: '',
    })
  })

  it('surfaces backend detail when runtime log config loading fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockGetRuntimeLogConfig.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'runtime log config detail error',
        },
      },
      message: 'generic runtime log config error',
    })

    mount(OpsSystemLogTable, {
      global: {
        stubs: {
          Select: SelectStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    expect(showError).toHaveBeenCalledWith('runtime log config detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })

  it('keeps the latest log query result when searches are triggered back to back', async () => {
    mockGetRuntimeLogConfig.mockResolvedValue({
      level: 'info',
      enable_sampling: false,
      sampling_initial: 100,
      sampling_thereafter: 100,
      caller: true,
      stacktrace_level: 'error',
      retention_days: 30,
    })

    const wrapper = mount(OpsSystemLogTable, {
      global: {
        stubs: {
          Select: SelectStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    const slowResponse = deferred<any>()
    const fastResponse = deferred<any>()
    mockListSystemLogs
      .mockReturnValueOnce(slowResponse.promise)
      .mockReturnValueOnce(fastResponse.promise)

    const queryInput = wrapper.get('input[placeholder="消息/request_id"]')
    const queryButton = wrapper
      .findAll('button')
      .find((button) => button.text() === '查询')

    expect(queryButton).toBeTruthy()
    await queryInput.setValue('first')
    await queryButton!.trigger('click')
    await queryInput.setValue('second')
    await queryButton!.trigger('click')

    fastResponse.resolve({
      items: [makeLog(2, 'second-hit')],
      total: 1,
    })
    await flushPromises()

    slowResponse.resolve({
      items: [makeLog(1, 'first-hit')],
      total: 1,
    })
    await flushPromises()

    expect(wrapper.text()).toContain('second-hit')
    expect(wrapper.text()).not.toContain('first-hit')
  })

  it('ignores a stale mount load after runtime config save applies', async () => {
    const staleLoad = deferred<any>()
    const saveResponse = deferred<any>()
    mockGetRuntimeLogConfig.mockImplementationOnce(() => staleLoad.promise)
    mockUpdateRuntimeLogConfig.mockImplementationOnce(() => saveResponse.promise)

    const wrapper = mount(OpsSystemLogTable, {
      global: {
        stubs: {
          Select: SelectStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    const runtimeNumberInputs = wrapper.findAll('input[type="number"]').slice(0, 3)
    await runtimeNumberInputs[0].setValue('5')
    await runtimeNumberInputs[1].setValue('10')
    await runtimeNumberInputs[2].setValue('14')
    await wrapper.find('input[type="checkbox"]').setValue(true)

    const saveButton = wrapper
      .findAll('button')
      .find((button) => button.text() === '保存并生效')

    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')

    saveResponse.resolve({
      level: 'warn',
      enable_sampling: true,
      sampling_initial: 5,
      sampling_thereafter: 10,
      caller: true,
      stacktrace_level: 'fatal',
      retention_days: 14,
    })
    await flushPromises()

    staleLoad.resolve({
      level: 'debug',
      enable_sampling: false,
      sampling_initial: 300,
      sampling_thereafter: 400,
      caller: false,
      stacktrace_level: 'error',
      retention_days: 90,
    })
    await flushPromises()

    const selectModels = wrapper.findAllComponents(SelectStub).map((component) => component.props('modelValue'))
    expect(selectModels.slice(0, 2)).toEqual(['warn', 'fatal'])
    expect(wrapper.findAll('input[type="number"]').slice(0, 3).map((input) => input.element.value)).toEqual([
      '5',
      '10',
      '14',
    ])
    expect(wrapper.findAll('input[type="checkbox"]')[0].element.checked).toBe(true)
    expect(wrapper.findAll('input[type="checkbox"]')[1].element.checked).toBe(true)
    expect(showSuccess).toHaveBeenCalledWith('日志运行时配置已生效')
    expect(wrapper.text()).not.toContain('加载中...')
  })

  it('ignores a stale mount load after runtime config reset applies', async () => {
    const staleLoad = deferred<any>()
    const resetResponse = deferred<any>()
    mockGetRuntimeLogConfig.mockImplementationOnce(() => staleLoad.promise)
    mockResetRuntimeLogConfig.mockImplementationOnce(() => resetResponse.promise)
    vi.spyOn(window, 'confirm').mockReturnValue(true)

    const wrapper = mount(OpsSystemLogTable, {
      global: {
        stubs: {
          Select: SelectStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    const resetButton = wrapper
      .findAll('button')
      .find((button) => button.text() === '回滚默认值')

    expect(resetButton).toBeTruthy()
    await resetButton!.trigger('click')

    resetResponse.resolve({
      level: 'error',
      enable_sampling: false,
      sampling_initial: 7,
      sampling_thereafter: 9,
      caller: false,
      stacktrace_level: 'fatal',
      retention_days: 21,
    })
    await flushPromises()

    staleLoad.resolve({
      level: 'debug',
      enable_sampling: true,
      sampling_initial: 1000,
      sampling_thereafter: 1000,
      caller: true,
      stacktrace_level: 'none',
      retention_days: 120,
    })
    await flushPromises()

    const selectModels = wrapper.findAllComponents(SelectStub).map((component) => component.props('modelValue'))
    expect(selectModels.slice(0, 2)).toEqual(['error', 'fatal'])
    expect(wrapper.findAll('input[type="number"]').slice(0, 3).map((input) => input.element.value)).toEqual([
      '7',
      '9',
      '21',
    ])
    expect(wrapper.findAll('input[type="checkbox"]')[0].element.checked).toBe(false)
    expect(wrapper.findAll('input[type="checkbox"]')[1].element.checked).toBe(false)
    expect(showSuccess).toHaveBeenCalledWith('已回滚到启动日志配置')
    expect(wrapper.text()).not.toContain('加载中...')
  })
})
