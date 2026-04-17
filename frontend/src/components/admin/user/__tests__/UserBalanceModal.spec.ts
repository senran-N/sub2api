import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UserBalanceModal from '../UserBalanceModal.vue'

const mockUpdateBalance = vi.fn()
const showSuccessMock = vi.fn()
const showErrorMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      updateBalance: (...args: any[]) => mockUpdateBalance(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
  }),
}))

vi.mock('@/utils/errorMessage', () => ({
  resolveErrorMessage: (_error: unknown, fallbackMessage: string) => fallbackMessage,
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

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createUser(id: number, balance: number) {
  return {
    id,
    email: `user-${id}@example.com`,
    balance,
  }
}

describe('UserBalanceModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUpdateBalance.mockResolvedValue({})
  })

  it('updates the balance for the active modal context', async () => {
    const wrapper = mount(UserBalanceModal, {
      props: {
        show: true,
        user: createUser(1, 10),
        operation: 'add',
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    await wrapper.find('input[type="number"]').setValue('5')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockUpdateBalance).toHaveBeenCalledWith(1, 5, 'add', '')
    expect(showSuccessMock).toHaveBeenCalledWith('common.success')
    expect(wrapper.emitted('success')).toBeTruthy()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('ignores a stale balance update after close and reopen', async () => {
    const updateRequest = createDeferred<Record<string, never>>()
    mockUpdateBalance.mockReturnValueOnce(updateRequest.promise)

    const wrapper = mount(UserBalanceModal, {
      props: {
        show: true,
        user: createUser(1, 10),
        operation: 'add',
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    await wrapper.find('input[type="number"]').setValue('4')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({
      show: true,
      user: createUser(2, 20),
      operation: 'subtract',
    })
    await flushPromises()

    await wrapper.find('input[type="number"]').setValue('3')
    expect(wrapper.find('.btn-danger').attributes('disabled')).toBeUndefined()

    updateRequest.resolve({})
    await flushPromises()

    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('success')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
    expect(wrapper.text()).toContain('user-2@example.com')
  })
})
