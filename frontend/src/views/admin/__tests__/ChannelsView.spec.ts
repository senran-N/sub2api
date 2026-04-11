import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ChannelsView from '../ChannelsView.vue'

const { showError, showSuccess } = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

const { listChannels, getAllGroups, updateChannel, createChannel, removeChannel } = vi.hoisted(() => ({
  listChannels: vi.fn(),
  getAllGroups: vi.fn(),
  updateChannel: vi.fn(),
  createChannel: vi.fn(),
  removeChannel: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    channels: {
      list: listChannels,
      update: updateChannel,
      create: createChannel,
      remove: removeChannel
    },
    groups: {
      getAll: getAllGroups
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('admin ChannelsView', () => {
  beforeEach(() => {
    listChannels.mockReset()
    getAllGroups.mockReset()
    updateChannel.mockReset()
    createChannel.mockReset()
    removeChannel.mockReset()
    showError.mockReset()
    showSuccess.mockReset()

    listChannels.mockResolvedValue({
      items: [{
        id: 1,
        name: 'Main channel',
        description: 'Primary',
        status: 'active',
        billing_model_source: 'channel_mapped',
        restrict_models: false,
        group_ids: [11],
        model_pricing: [],
        model_mapping: {},
        created_at: '2026-04-01T00:00:00Z',
        updated_at: '2026-04-01T00:00:00Z'
      }],
      total: 1
    })
    getAllGroups.mockResolvedValue([
      {
        id: 11,
        name: 'OpenAI Pro',
        platform: 'openai',
        rate_multiplier: 1,
        account_count: 2
      }
    ])
  })

  it('loads channels on mount and opens the create dialog with feature data prepared', async () => {
    const wrapper = mount(ChannelsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: { template: '<div><slot name="empty" /></div>' },
          Pagination: true,
          BaseDialog: {
            props: ['show', 'title'],
            template: '<div class="dialog" :data-show="show"><slot /></div>'
          },
          ConfirmDialog: true,
          EmptyState: { template: '<div />' },
          Select: {
            props: ['modelValue', 'options', 'placeholder'],
            emits: ['update:modelValue', 'change'],
            template: '<div class="select-stub" />'
          },
          PlatformIcon: true,
          Toggle: true,
          Icon: true,
          PricingEntryCard: true
        }
      }
    })

    await flushPromises()

    expect(listChannels).toHaveBeenCalledWith(1, expect.any(Number), {
      status: undefined,
      search: undefined
    }, { signal: expect.any(AbortSignal) })
    expect(getAllGroups).toHaveBeenCalledTimes(1)

    await wrapper.find('button.btn.btn-primary').trigger('click')
    await flushPromises()

    expect(getAllGroups).toHaveBeenCalledTimes(2)
    expect(listChannels).toHaveBeenCalledWith(1, 1000)
    expect(wrapper.find('.dialog').attributes('data-show')).toBe('true')
  })

  it('surfaces resolved request messages for create, toggle, and delete failures', async () => {
    const wrapper = mount(ChannelsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: { template: '<div><slot name="empty" /></div>' },
          Pagination: true,
          BaseDialog: {
            props: ['show', 'title'],
            template: '<div class="dialog" :data-show="show"><slot /></div>'
          },
          ConfirmDialog: true,
          EmptyState: { template: '<div />' },
          Select: {
            props: ['modelValue', 'options', 'placeholder'],
            emits: ['update:modelValue', 'change'],
            template: '<div class="select-stub" />'
          },
          PlatformIcon: true,
          Toggle: true,
          Icon: true,
          PricingEntryCard: true
        }
      }
    })

    await flushPromises()

    const vm = wrapper.vm as any
    vm.form.name = 'Main channel'
    vm.form.platforms = [{
      platform: 'openai',
      enabled: true,
      collapsed: false,
      group_ids: [11],
      model_mapping: {},
      model_pricing: []
    }]

    createChannel.mockRejectedValueOnce(new Error('create unavailable'))
    await vm.handleSubmit()

    updateChannel.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'toggle blocked'
        }
      }
    })
    await vm.toggleChannelStatus({
      id: 1,
      status: 'active'
    })

    vm.deletingChannel = { id: 1, name: 'Main channel' }
    removeChannel.mockRejectedValueOnce(new Error('delete unavailable'))
    await vm.confirmDelete()

    expect(showError).toHaveBeenNthCalledWith(1, 'create unavailable')
    expect(showError).toHaveBeenNthCalledWith(2, 'toggle blocked')
    expect(showError).toHaveBeenNthCalledWith(3, 'delete unavailable')
  })
})
