import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UserBalanceHistoryModal from '../UserBalanceHistoryModal.vue'

const mockGetUserBalanceHistory = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserBalanceHistory: (...args: any[]) => mockGetUserBalanceHistory(...args)
    }
  }
}))

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => value
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

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: '' }
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /></div>'
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: { type: String, default: '' },
    options: { type: Array, default: () => [] }
  },
  emits: ['update:modelValue', 'change'],
  template: '<div class="select-stub" />'
})

const IconStub = defineComponent({
  name: 'IconStub',
  template: '<span />'
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createHistoryResponse(label: string, totalRecharged: number) {
  return {
    items: [
      {
        id: totalRecharged,
        code: `${label.toLowerCase()}-code`,
        type: 'balance',
        value: 10,
        status: 'used',
        used_by: 1,
        used_at: '2026-04-17T00:00:00Z',
        created_at: '2026-04-17T00:00:00Z',
        group_id: null,
        validity_days: 30,
        notes: `${label} notes`
      }
    ],
    total: 1,
    page: 1,
    page_size: 15,
    pages: 1,
    total_recharged: totalRecharged
  }
}

describe('UserBalanceHistoryModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetUserBalanceHistory.mockResolvedValue(createHistoryResponse('Default', 5))
  })

  it('keeps the latest user history when requests resolve out of order', async () => {
    const firstLoad = createDeferred<ReturnType<typeof createHistoryResponse>>()
    const secondLoad = createDeferred<ReturnType<typeof createHistoryResponse>>()
    mockGetUserBalanceHistory
      .mockImplementationOnce(() => firstLoad.promise)
      .mockImplementationOnce(() => secondLoad.promise)

    const wrapper = mount(UserBalanceHistoryModal, {
      props: {
        show: true,
        user: {
          id: 1,
          email: 'first@example.com',
          username: 'first',
          balance: 10,
          notes: ''
        }
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          Icon: IconStub
        }
      }
    })

    await wrapper.setProps({
      user: {
        id: 2,
        email: 'second@example.com',
        username: 'second',
        balance: 20,
        notes: ''
      }
    })

    secondLoad.resolve(createHistoryResponse('Second', 22))
    await flushPromises()
    expect(wrapper.text()).toContain('Second notes')
    expect(wrapper.text()).toContain('$22.00')
    expect(wrapper.text()).not.toContain('First notes')

    firstLoad.resolve(createHistoryResponse('First', 11))
    await flushPromises()
    expect(wrapper.text()).toContain('Second notes')
    expect(wrapper.text()).toContain('$22.00')
    expect(wrapper.text()).not.toContain('First notes')
  })
})
