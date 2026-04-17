import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsAlertRulesCard from '../OpsAlertRulesCard.vue'

const mockListAlertRules = vi.fn()
const mockCreateAlertRule = vi.fn()
const mockUpdateAlertRule = vi.fn()
const mockDeleteAlertRule = vi.fn()
const mockGetAllGroups = vi.fn()
const showSuccess = vi.fn()
const showError = vi.fn()
const showInfo = vi.fn()
const showWarning = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listAlertRules: (...args: any[]) => mockListAlertRules(...args),
    createAlertRule: (...args: any[]) => mockCreateAlertRule(...args),
    updateAlertRule: (...args: any[]) => mockUpdateAlertRule(...args),
    deleteAlertRule: (...args: any[]) => mockDeleteAlertRule(...args),
  },
}))

vi.mock('@/api', () => ({
  adminAPI: {
    groups: {
      getAll: (...args: any[]) => mockGetAllGroups(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess,
    showError,
    showInfo,
    showWarning,
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, any>) => {
        if (!params) return key
        return `${key}:${JSON.stringify(params)}`
      },
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
  template: '<div v-if="show" class="base-dialog-stub"><slot /><template v-if="$slots.footer"><slot name="footer" /></template></div>',
})

const ConfirmDialogStub = defineComponent({
  name: 'ConfirmDialogStub',
  props: {
    show: { type: Boolean, default: false },
  },
  emits: ['confirm', 'cancel'],
  template: '<div v-if="show" class="confirm-dialog-stub" />',
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

function makeRule(id: number, name: string) {
  return {
    id,
    name,
    description: `${name} description`,
    enabled: true,
    metric_type: 'error_rate',
    operator: '>',
    threshold: 1,
    window_minutes: 1,
    sustained_minutes: 2,
    severity: 'P1',
    cooldown_minutes: 10,
    notify_email: true,
  }
}

function mountComponent() {
  return mount(OpsAlertRulesCard, {
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        ConfirmDialog: ConfirmDialogStub,
        Select: SelectStub,
      },
    },
  })
}

describe('OpsAlertRulesCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListAlertRules.mockResolvedValue([])
    mockGetAllGroups.mockResolvedValue([])
    mockCreateAlertRule.mockResolvedValue(undefined)
    mockUpdateAlertRule.mockResolvedValue(undefined)
    mockDeleteAlertRule.mockResolvedValue(undefined)
  })

  it('在空规则时显示初始化引导，并可一键创建四条推荐规则', async () => {
    mockListAlertRules
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([
        { id: 1, name: 'r1' },
        { id: 2, name: 'r2' },
        { id: 3, name: 'r3' },
        { id: 4, name: 'r4' },
      ])
    mockCreateAlertRule.mockResolvedValue(undefined)

    const wrapper = mountComponent()
    await flushPromises()

    expect(wrapper.text()).toContain('admin.ops.alertRules.emptyState.title')

    const createAllButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('admin.ops.alertRules.presets.createAll'))
    expect(createAllButton).toBeDefined()

    await createAllButton!.trigger('click')
    await flushPromises()

    expect(mockCreateAlertRule).toHaveBeenCalledTimes(4)
    expect(showSuccess).toHaveBeenCalledWith('admin.ops.alertRules.presets.createSuccess:{"count":4}')
  })

  it('按规则语义去重，而不是按名称去重', async () => {
    mockListAlertRules
      .mockResolvedValueOnce([
        {
          id: 11,
          name: '自定义 Acquire 告警',
          description: 'same semantics, different name',
          enabled: true,
          metric_type: 'scheduler_acquire_success_rate',
          operator: '<',
          threshold: 75,
          window_minutes: 5,
          sustained_minutes: 3,
          severity: 'P1',
          cooldown_minutes: 15,
          notify_email: true,
        },
      ])
      .mockResolvedValueOnce([])
    mockCreateAlertRule.mockResolvedValue(undefined)

    const wrapper = mountComponent()
    await flushPromises()

    expect(wrapper.text()).toContain('admin.ops.alertRules.presets.created')

    const createAllButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('admin.ops.alertRules.presets.createAll'))
    expect(createAllButton).toBeDefined()

    await createAllButton!.trigger('click')
    await flushPromises()

    expect(mockCreateAlertRule).toHaveBeenCalledTimes(3)
    expect(mockCreateAlertRule).not.toHaveBeenCalledWith(
      expect.objectContaining({
        metric_type: 'scheduler_acquire_success_rate',
        operator: '<',
        threshold: 75,
        window_minutes: 5,
        sustained_minutes: 3,
      })
    )
  })

  it('推荐规则部分创建失败时给出部分成功提示', async () => {
    mockListAlertRules
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([])

    mockCreateAlertRule
      .mockResolvedValueOnce(undefined)
      .mockResolvedValueOnce(undefined)
      .mockRejectedValueOnce(new Error('boom'))
      .mockResolvedValueOnce(undefined)

    const wrapper = mountComponent()
    await flushPromises()

    const createAllButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('admin.ops.alertRules.presets.createAll'))
    expect(createAllButton).toBeDefined()

    await createAllButton!.trigger('click')
    await flushPromises()

    expect(mockCreateAlertRule).toHaveBeenCalledTimes(4)
    expect(showWarning).toHaveBeenCalledWith('admin.ops.alertRules.presets.createPartial:{"success":3,"failed":1}')
  })

  it('加载失败时优先展示后端 detail', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockListAlertRules.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'alert rules detail error'
        }
      },
      message: 'generic alert rules error'
    })

    mountComponent()
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('alert rules detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })

  it('忽略保存前发起且晚返回的旧 refresh 结果，并保持 saving ownership', async () => {
    const saveResponse = deferred<void>()
    const staleRefresh = deferred<any[]>()
    mockListAlertRules.mockResolvedValueOnce([makeRule(1, 'current rule')])

    const wrapper = mountComponent()
    await flushPromises()

    mockListAlertRules.mockReset()
    mockListAlertRules
      .mockReturnValueOnce(staleRefresh.promise)
      .mockResolvedValueOnce([makeRule(2, 'saved rule')])
    mockUpdateAlertRule.mockReset()
    mockUpdateAlertRule.mockReturnValueOnce(saveResponse.promise)

    const editButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.edit')
    expect(editButton).toBeTruthy()
    await editButton!.trigger('click')

    const refreshButton = wrapper.get('.ops-alert-rules-card__refresh')
    await refreshButton.trigger('click')

    const saveButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.save')
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')
    await flushPromises()

    staleRefresh.resolve([makeRule(99, 'stale rule')])
    await flushPromises()

    const pendingSaveButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.saving')
    expect(pendingSaveButton).toBeTruthy()
    expect(pendingSaveButton!.attributes('disabled')).toBeDefined()
    expect(refreshButton.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('current rule')
    expect(wrapper.text()).not.toContain('stale rule')

    saveResponse.resolve()
    await flushPromises()

    expect(wrapper.text()).toContain('saved rule')
    expect(wrapper.text()).not.toContain('stale rule')
  })

  it('批量创建推荐规则后忽略晚返回的旧 refresh 结果', async () => {
    const staleRefresh = deferred<any[]>()
    mockListAlertRules.mockResolvedValueOnce([])

    const wrapper = mountComponent()
    await flushPromises()

    mockListAlertRules.mockReset()
    mockListAlertRules
      .mockReturnValueOnce(staleRefresh.promise)
      .mockResolvedValueOnce([makeRule(4, 'preset rule')])

    const refreshButton = wrapper.get('.ops-alert-rules-card__refresh')
    await refreshButton.trigger('click')

    const createAllButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('admin.ops.alertRules.presets.createAll'))
    expect(createAllButton).toBeTruthy()
    await createAllButton!.trigger('click')
    await flushPromises()

    staleRefresh.resolve([makeRule(98, 'stale rule')])
    await flushPromises()

    expect(mockCreateAlertRule).toHaveBeenCalledTimes(4)
    expect(wrapper.text()).toContain('preset rule')
    expect(wrapper.text()).not.toContain('stale rule')
    expect(refreshButton.attributes('disabled')).toBeUndefined()
  })

  it('删除规则后忽略晚返回的旧 refresh 结果', async () => {
    const deleteResponse = deferred<void>()
    const staleRefresh = deferred<any[]>()
    mockListAlertRules.mockResolvedValueOnce([makeRule(1, 'current rule')])

    const wrapper = mountComponent()
    await flushPromises()

    mockListAlertRules.mockReset()
    mockListAlertRules
      .mockReturnValueOnce(staleRefresh.promise)
      .mockResolvedValueOnce([])
    mockDeleteAlertRule.mockReset()
    mockDeleteAlertRule.mockReturnValueOnce(deleteResponse.promise)

    const deleteButton = wrapper
      .findAll('button')
      .find((button) => button.text() === 'common.delete')
    expect(deleteButton).toBeTruthy()
    await deleteButton!.trigger('click')

    const refreshButton = wrapper.get('.ops-alert-rules-card__refresh')
    await refreshButton.trigger('click')

    wrapper.getComponent(ConfirmDialogStub).vm.$emit('confirm')

    staleRefresh.resolve([makeRule(97, 'stale rule')])
    await flushPromises()

    expect(refreshButton.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('current rule')
    expect(wrapper.text()).not.toContain('stale rule')

    deleteResponse.resolve()
    await flushPromises()

    expect(wrapper.text()).toContain('admin.ops.alertRules.emptyState.title')
    expect(wrapper.text()).not.toContain('stale rule')
  })
})
