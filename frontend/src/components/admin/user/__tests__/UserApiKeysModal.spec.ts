import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UserApiKeysModal from '../UserApiKeysModal.vue'

const mockGetUserApiKeys = vi.fn()
const mockGetAllGroups = vi.fn()
const mockUpdateApiKeyGroup = vi.fn()
const showSuccessMock = vi.fn()
const showErrorMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserApiKeys: (...args: any[]) => mockGetUserApiKeys(...args),
    },
    groups: {
      getAll: (...args: any[]) => mockGetAllGroups(...args),
    },
    apiKeys: {
      updateApiKeyGroup: (...args: any[]) => mockUpdateApiKeyGroup(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock,
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

function createGroup(id: number, name: string) {
  return {
    id,
    name,
    platform: 'openai',
    subscription_type: 'shared',
    rate_multiplier: 1,
    description: `${name} description`,
  }
}

describe('UserApiKeysModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetUserApiKeys.mockResolvedValue({ items: [] })
    mockGetAllGroups.mockResolvedValue([])
    mockUpdateApiKeyGroup.mockResolvedValue({
      api_key: createApiKey(1, 'Updated key'),
      auto_granted_group_access: false
    })
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
          GroupBadge: defineComponent({
            name: 'GroupBadgeStub',
            props: { name: { type: String, default: '' } },
            template: '<span>{{ name }}</span>',
          }),
          GroupOptionItem: defineComponent({
            name: 'GroupOptionItemStub',
            props: { name: { type: String, default: '' } },
            template: '<span>{{ name }}</span>',
          }),
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
          GroupBadge: defineComponent({
            name: 'GroupBadgeStub',
            props: { name: { type: String, default: '' } },
            template: '<span>{{ name }}</span>',
          }),
          GroupOptionItem: defineComponent({
            name: 'GroupOptionItemStub',
            props: { name: { type: String, default: '' } },
            template: '<span>{{ name }}</span>',
          }),
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

  it('ignores a stale group change after close and reopen, without clearing the new pending key state', async () => {
    const firstGroupChange = createDeferred<{
      api_key: ReturnType<typeof createApiKey>
      auto_granted_group_access: boolean
      granted_group_name?: string
    }>()
    const secondGroupChange = createDeferred<{
      api_key: ReturnType<typeof createApiKey>
      auto_granted_group_access: boolean
      granted_group_name?: string
    }>()

    mockGetUserApiKeys
      .mockResolvedValueOnce({ items: [createApiKey(1, 'Alpha key')] })
      .mockResolvedValueOnce({ items: [createApiKey(1, 'Beta key')] })
    mockGetAllGroups
      .mockResolvedValueOnce([createGroup(11, 'Alpha Group')])
      .mockResolvedValueOnce([createGroup(12, 'Beta Group')])
    mockUpdateApiKeyGroup
      .mockImplementationOnce(() => firstGroupChange.promise)
      .mockImplementationOnce(() => secondGroupChange.promise)

    const wrapper = mount(UserApiKeysModal, {
      props: {
        show: true,
        user: {
          id: 7,
          email: 'alpha@example.com',
          username: 'alpha-user',
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          GroupBadge: defineComponent({
            name: 'GroupBadgeStub',
            props: { name: { type: String, default: '' } },
            template: '<span>{{ name }}</span>',
          }),
          GroupOptionItem: defineComponent({
            name: 'GroupOptionItemStub',
            props: { name: { type: String, default: '' } },
            template: '<span>{{ name }}</span>',
          }),
          Teleport: true,
        },
      },
      attachTo: document.body,
    })

    await flushPromises()

    const alphaTrigger = wrapper.find('.user-api-keys-modal__group-trigger')
    expect(alphaTrigger.exists()).toBe(true)
    await alphaTrigger.trigger('click')
    await flushPromises()

    const alphaOption = wrapper
      .findAll('.user-api-keys-modal__dropdown-option')
      .find((option) => option.text().includes('Alpha Group'))
    expect(alphaOption).toBeTruthy()
    await alphaOption!.trigger('click')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({
      show: true,
      user: {
        id: 8,
        email: 'beta@example.com',
        username: 'beta-user',
      },
    })
    await flushPromises()

    const betaTrigger = wrapper.find('.user-api-keys-modal__group-trigger')
    expect(betaTrigger.exists()).toBe(true)
    await betaTrigger.trigger('click')
    await flushPromises()

    const betaOption = wrapper
      .findAll('.user-api-keys-modal__dropdown-option')
      .find((option) => option.text().includes('Beta Group'))
    expect(betaOption).toBeTruthy()
    await betaOption!.trigger('click')
    await flushPromises()

    expect(wrapper.find('.user-api-keys-modal__group-trigger').attributes('disabled')).toBeDefined()

    firstGroupChange.resolve({
      api_key: {
        ...createApiKey(1, 'Alpha key'),
        group_id: 11,
        group: createGroup(11, 'Alpha Group')
      },
      auto_granted_group_access: false
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Beta key')
    expect(wrapper.text()).not.toContain('Alpha key')
    expect(wrapper.find('.user-api-keys-modal__group-trigger').attributes('disabled')).toBeDefined()
    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()

    secondGroupChange.resolve({
      api_key: {
        ...createApiKey(1, 'Beta key'),
        group_id: 12,
        group: createGroup(12, 'Beta Group')
      },
      auto_granted_group_access: false
    })
    await flushPromises()

    expect(showSuccessMock).toHaveBeenCalledTimes(1)
    expect(showSuccessMock).toHaveBeenCalledWith('admin.users.groupChangedSuccess')
    expect(wrapper.find('.user-api-keys-modal__group-trigger').attributes('disabled')).toBeUndefined()
  })
})
