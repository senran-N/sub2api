import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UserCreateModal from '../UserCreateModal.vue'

const mockCreateUser = vi.fn()
const showSuccessMock = vi.fn()
const showErrorMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      create: (...args: any[]) => mockCreateUser(...args),
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

const IconStub = defineComponent({
  name: 'IconStub',
  template: '<span />',
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

describe('UserCreateModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockCreateUser.mockResolvedValue({})
  })

  it('creates a user for the active modal context', async () => {
    const wrapper = mount(UserCreateModal, {
      props: {
        show: true,
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: IconStub,
        },
      },
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('new@example.com')
    await inputs[1].setValue('secret123')
    await inputs[2].setValue('new-user')
    await inputs[3].setValue('5')
    await inputs[4].setValue('2')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockCreateUser).toHaveBeenCalledWith({
      email: 'new@example.com',
      password: 'secret123',
      username: 'new-user',
      notes: '',
      balance: 5,
      concurrency: 2,
    })
    expect(showSuccessMock).toHaveBeenCalledWith('admin.users.userCreated')
    expect(wrapper.emitted('success')).toBeTruthy()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('ignores a stale create result after close and reopen', async () => {
    const createRequest = createDeferred<Record<string, never>>()
    mockCreateUser.mockReturnValueOnce(createRequest.promise)

    const wrapper = mount(UserCreateModal, {
      props: {
        show: true,
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: IconStub,
        },
      },
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('alpha@example.com')
    await inputs[1].setValue('alpha-secret')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({ show: true })
    await flushPromises()

    const reopenedInputs = wrapper.findAll('input')
    await reopenedInputs[0].setValue('beta@example.com')
    expect(wrapper.find('.btn-primary').attributes('disabled')).toBeUndefined()

    createRequest.resolve({})
    await flushPromises()

    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('success')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
    expect(wrapper.findAll('input')[0].element).toHaveProperty('value', 'beta@example.com')
  })
})
