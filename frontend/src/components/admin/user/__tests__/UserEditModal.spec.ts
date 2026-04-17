import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, watch } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UserEditModal from '../UserEditModal.vue'

const mockUpdateUser = vi.fn()
const mockUpdateUserAttributeValues = vi.fn()
const showSuccessMock = vi.fn()
const showErrorMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      update: (...args: any[]) => mockUpdateUser(...args),
    },
    userAttributes: {
      updateUserAttributeValues: (...args: any[]) => mockUpdateUserAttributeValues(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true),
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

const UserAttributeFormStub = defineComponent({
  name: 'UserAttributeFormStub',
  props: {
    modelValue: { type: Object, default: () => ({}) },
    userId: { type: Number, default: 0 },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    watch(
      () => props.userId,
      (userId) => {
        emit('update:modelValue', userId ? { profile: `user-${userId}` } : {})
      },
      { immediate: true }
    )

    return {}
  },
  template: '<div class="user-attribute-form-stub" />',
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createUser(id: number) {
  return {
    id,
    email: `user-${id}@example.com`,
    username: `user-${id}`,
    notes: `notes-${id}`,
    concurrency: 2,
  }
}

describe('UserEditModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUpdateUser.mockResolvedValue({})
    mockUpdateUserAttributeValues.mockResolvedValue({})
  })

  it('updates the user and attribute values for the active modal context', async () => {
    const wrapper = mount(UserEditModal, {
      props: {
        show: true,
        user: createUser(1),
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          UserAttributeForm: UserAttributeFormStub,
          Icon: IconStub,
        },
      },
    })

    await flushPromises()
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockUpdateUser).toHaveBeenCalledWith(1, {
      email: 'user-1@example.com',
      username: 'user-1',
      notes: 'notes-1',
      concurrency: 2,
    })
    expect(mockUpdateUserAttributeValues).toHaveBeenCalledWith(1, { profile: 'user-1' })
    expect(showSuccessMock).toHaveBeenCalledWith('admin.users.userUpdated')
    expect(wrapper.emitted('success')).toBeTruthy()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('ignores a stale save after close and reopen, preventing attribute follow-up writes', async () => {
    const updateRequest = createDeferred<Record<string, never>>()
    mockUpdateUser.mockReturnValueOnce(updateRequest.promise)

    const wrapper = mount(UserEditModal, {
      props: {
        show: true,
        user: createUser(1),
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          UserAttributeForm: UserAttributeFormStub,
          Icon: IconStub,
        },
      },
    })

    await flushPromises()
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({
      show: true,
      user: createUser(2),
    })
    await flushPromises()

    updateRequest.resolve({})
    await flushPromises()

    expect(mockUpdateUserAttributeValues).not.toHaveBeenCalled()
    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('success')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
  })
})
