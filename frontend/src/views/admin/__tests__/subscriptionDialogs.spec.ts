import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { UserSubscription } from '@/types'
import SubscriptionAssignDialog from '../subscriptions/SubscriptionAssignDialog.vue'
import SubscriptionColumnSettingsMenu from '../subscriptions/SubscriptionColumnSettingsMenu.vue'
import SubscriptionExtendDialog from '../subscriptions/SubscriptionExtendDialog.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) =>
        params ? `${key}:${JSON.stringify(params)}` : key
    })
  }
})

vi.mock('@/utils/format', () => ({
  formatDateOnly: (value: string) => `date:${value}`
}))

function createSubscription(overrides: Partial<UserSubscription> = {}): UserSubscription {
  return {
    id: 1,
    user_id: 2,
    group_id: 3,
    status: 'active',
    daily_usage_usd: 0,
    weekly_usage_usd: 0,
    monthly_usage_usd: 0,
    daily_window_start: null,
    weekly_window_start: null,
    monthly_window_start: null,
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    expires_at: '2026-04-10T00:00:00Z',
    user: {
      email: 'user@example.com'
    } as any,
    ...overrides
  }
}

describe('subscription dialogs', () => {
  it('renders column settings menu and emits updates', async () => {
    const wrapper = mount(SubscriptionColumnSettingsMenu, {
      props: {
        userColumnMode: 'email',
        toggleableColumns: [
          { key: 'usage', label: 'Usage', sortable: false },
          { key: 'status', label: 'Status', sortable: false }
        ],
        isColumnVisible: (key: string) => key === 'usage'
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('set-user-mode')?.[0]).toEqual(['username'])
    expect(wrapper.emitted('toggle-column')?.[0]).toEqual(['usage'])
  })

  it('renders assign dialog and emits user search interactions', async () => {
    const form = {
      user_id: null,
      group_id: null,
      validity_days: 30
    }
    const wrapper = mount(SubscriptionAssignDialog, {
      props: {
        show: true,
        form,
        userKeyword: 'ali',
        userResults: [{ id: 2, email: 'alice@example.com' }],
        userLoading: false,
        showUserDropdown: true,
        selectedUser: null,
        groupOptions: [
          {
            value: 1,
            label: 'Premium',
            description: null,
            platform: 'openai',
            subscriptionType: 'subscription',
            rate: 1
          }
        ],
        submitting: false
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title', 'width'],
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Select: {
            props: ['modelValue', 'options'],
            template: '<div class="select-stub">{{ modelValue }}</div>'
          },
          GroupBadge: true,
          GroupOptionItem: true,
          Icon: true
        }
      }
    })

    const input = wrapper.find('input')
    await input.setValue('bob')
    expect(wrapper.emitted('update:userKeyword')?.[0]).toEqual(['bob'])
    expect(wrapper.emitted('search-users')?.length).toBe(1)

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    expect(wrapper.emitted('select-user')?.[0]).toEqual([{ id: 2, email: 'alice@example.com' }])
  })

  it('renders extend dialog and emits submit', async () => {
    const wrapper = mount(SubscriptionExtendDialog, {
      props: {
        show: true,
        subscription: createSubscription(),
        form: {
          days: 15
        },
        submitting: false
      },
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title', 'width'],
            template: '<div><slot /><slot name="footer" /></div>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('user@example.com')
    expect(wrapper.text()).toContain('date:2026-04-10T00:00:00Z')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')?.length).toBe(1)
  })
})
