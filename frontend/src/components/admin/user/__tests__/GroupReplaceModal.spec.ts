import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import GroupReplaceModal from '../GroupReplaceModal.vue'

const mockReplaceGroup = vi.fn()
const showSuccessMock = vi.fn()
const showErrorMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      replaceGroup: (...args: any[]) => mockReplaceGroup(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
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

function createUser(id: number) {
  return {
    id,
    email: `user-${id}@example.com`,
    username: `user-${id}`,
  }
}

function createGroup(id: number, name: string, overrides: Record<string, unknown> = {}) {
  return {
    id,
    name,
    platform: 'openai',
    status: 'active',
    is_exclusive: true,
    subscription_type: 'standard',
    ...overrides,
  }
}

describe('GroupReplaceModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockReplaceGroup.mockResolvedValue({ migrated_keys: 2 })
  })

  it('submits the selected replacement group', async () => {
    const wrapper = mount(GroupReplaceModal, {
      props: {
        show: true,
        user: createUser(1),
        oldGroup: { id: 1, name: 'Old Group' },
        allGroups: [
          createGroup(1, 'Old Group'),
          createGroup(2, 'New Group'),
          createGroup(3, 'Inactive Group', { status: 'inactive' }),
        ],
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: IconStub,
        },
      },
    })

    await wrapper.find('input[type="radio"]').setValue()
    await wrapper.find('.btn-primary').trigger('click')
    await flushPromises()

    expect(mockReplaceGroup).toHaveBeenCalledWith(1, 1, 2)
    expect(showSuccessMock).toHaveBeenCalledWith('admin.users.replaceGroupSuccess')
    expect(wrapper.emitted('success')).toBeTruthy()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('ignores a stale replace result after close and reopen', async () => {
    const replaceRequest = createDeferred<{ migrated_keys: number }>()
    mockReplaceGroup.mockReturnValueOnce(replaceRequest.promise)

    const wrapper = mount(GroupReplaceModal, {
      props: {
        show: true,
        user: createUser(1),
        oldGroup: { id: 1, name: 'Old Group' },
        allGroups: [
          createGroup(1, 'Old Group'),
          createGroup(2, 'Alpha Replacement'),
        ],
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: IconStub,
        },
      },
    })

    await wrapper.find('input[type="radio"]').setValue()
    await wrapper.find('.btn-primary').trigger('click')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({
      show: true,
      user: createUser(2),
      oldGroup: { id: 3, name: 'Another Group' },
      allGroups: [
        createGroup(3, 'Another Group'),
        createGroup(4, 'Beta Replacement'),
      ],
    })
    await flushPromises()

    const nextRadio = wrapper.find('input[type="radio"]')
    await nextRadio.setValue()
    expect(wrapper.find('.btn-primary').attributes('disabled')).toBeUndefined()

    replaceRequest.resolve({ migrated_keys: 4 })
    await flushPromises()

    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('success')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
    expect(wrapper.text()).toContain('Beta Replacement')
  })
})
