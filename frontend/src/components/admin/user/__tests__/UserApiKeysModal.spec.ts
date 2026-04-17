import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UserApiKeysModal from '../UserApiKeysModal.vue'

const mockGetUserApiKeys = vi.fn()
const mockGetAllGroups = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserApiKeys: (...args: any[]) => mockGetUserApiKeys(...args),
    },
    groups: {
      getAll: (...args: any[]) => mockGetAllGroups(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
  }),
}))

vi.mock('@/composables/useDocumentThemeVersion', () => ({
  useDocumentThemeVersion: () => ref(0),
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
  template: '<div v-if="show" class="base-dialog-stub"><slot /></div>',
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createApiKey(id: number, name: string) {
  return {
    id,
    name,
    status: 'active',
    key: 'sk-' + '1234567890abcdef1234567890',
    created_at: '2026-04-17T00:00:00Z',
    group_id: null,
    group: null
  }
}

describe('UserApiKeysModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetUserApiKeys.mockResolvedValue({ items: [] })
    mockGetAllGroups.mockResolvedValue([])
  })

  it('loads user api keys and groups immediately when mounted open', async () => {
    mount(UserApiKeysModal, {
      props: {
        show: true,
        user: {
          id: 7,
          email: 'ops@example.com',
          username: 'ops-user',
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          GroupBadge: true,
          GroupOptionItem: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()

    expect(mockGetUserApiKeys).toHaveBeenCalledWith(7)
    expect(mockGetAllGroups).toHaveBeenCalledTimes(1)
  })

  it('keeps the latest user api keys when requests resolve out of order', async () => {
    const firstKeys = createDeferred<{ items: ReturnType<typeof createApiKey>[] }>()
    const secondKeys = createDeferred<{ items: ReturnType<typeof createApiKey>[] }>()
    mockGetUserApiKeys
      .mockImplementationOnce(() => firstKeys.promise)
      .mockImplementationOnce(() => secondKeys.promise)

    const wrapper = mount(UserApiKeysModal, {
      props: {
        show: true,
        user: {
          id: 7,
          email: 'ops@example.com',
          username: 'ops-user'
        }
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          GroupBadge: true,
          GroupOptionItem: true,
          Teleport: true
        }
      }
    })

    await wrapper.setProps({
      user: {
        id: 8,
        email: 'beta@example.com',
        username: 'beta-user'
      }
    })

    secondKeys.resolve({ items: [createApiKey(2, 'Beta key')] })
    await flushPromises()
    expect(wrapper.text()).toContain('Beta key')
    expect(wrapper.text()).not.toContain('Alpha key')

    firstKeys.resolve({ items: [createApiKey(1, 'Alpha key')] })
    await flushPromises()
    expect(wrapper.text()).toContain('Beta key')
    expect(wrapper.text()).not.toContain('Alpha key')
  })
})
