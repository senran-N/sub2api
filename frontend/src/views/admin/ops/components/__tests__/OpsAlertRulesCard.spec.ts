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
    mockGetAllGroups.mockResolvedValue([])
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
})
