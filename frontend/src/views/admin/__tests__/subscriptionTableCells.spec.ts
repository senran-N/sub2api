import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { UserSubscription } from '@/types'
import SubscriptionActionsCell from '../subscriptions/SubscriptionActionsCell.vue'
import SubscriptionExpirationCell from '../subscriptions/SubscriptionExpirationCell.vue'
import SubscriptionGuideModal from '../subscriptions/SubscriptionGuideModal.vue'
import SubscriptionStatusBadge from '../subscriptions/SubscriptionStatusBadge.vue'
import SubscriptionUsageCell from '../subscriptions/SubscriptionUsageCell.vue'
import SubscriptionUserCell from '../subscriptions/SubscriptionUserCell.vue'

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

function createSubscription(overrides: Partial<UserSubscription> = {}): UserSubscription {
  return {
    id: 1,
    user_id: 2,
    group_id: 3,
    status: 'active',
    daily_usage_usd: 12.34,
    weekly_usage_usd: 23.45,
    monthly_usage_usd: 34.56,
    daily_window_start: '2026-04-04T00:00:00Z',
    weekly_window_start: '2026-04-01T00:00:00Z',
    monthly_window_start: '2026-04-01T00:00:00Z',
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    expires_at: '2026-04-10T00:00:00Z',
    user: {
      id: 2,
      email: 'user@example.com',
      username: 'alice'
    } as any,
    group: {
      id: 3,
      name: 'Premium',
      platform: 'openai',
      subscription_type: 'subscription',
      rate_multiplier: 1,
      daily_limit_usd: 20,
      weekly_limit_usd: 50,
      monthly_limit_usd: 100
    } as any,
    ...overrides
  }
}

describe('subscription table cells', () => {
  it('renders user cell in both modes', () => {
    const emailWrapper = mount(SubscriptionUserCell, {
      props: {
        subscription: createSubscription(),
        mode: 'email'
      }
    })
    expect(emailWrapper.text()).toContain('user@example.com')
    expect(emailWrapper.text()).toContain('U')

    const usernameWrapper = mount(SubscriptionUserCell, {
      props: {
        subscription: createSubscription(),
        mode: 'username'
      }
    })
    expect(usernameWrapper.text()).toContain('alice')
    expect(usernameWrapper.text()).toContain('A')
  })

  it('renders usage summary and unlimited state', () => {
    const wrapper = mount(SubscriptionUsageCell, {
      props: {
        subscription: createSubscription()
      }
    })

    expect(wrapper.text()).toContain('admin.subscriptions.daily')
    expect(wrapper.text()).toContain('$12.34')
    expect(wrapper.text()).toContain('admin.subscriptions.resetIn')

    const unlimitedWrapper = mount(SubscriptionUsageCell, {
      props: {
        subscription: createSubscription({
          group: {
            id: 3
          } as any
        })
      }
    })
    expect(unlimitedWrapper.text()).toContain('admin.subscriptions.unlimited')
  })

  it('renders expiration and status badges', () => {
    const expirationWrapper = mount(SubscriptionExpirationCell, {
      props: {
        expiresAt: '2026-04-10T00:00:00Z'
      }
    })
    expect(expirationWrapper.text()).toContain('2026')
    expect(expirationWrapper.text()).toContain('admin.subscriptions.daysRemaining')

    const noExpirationWrapper = mount(SubscriptionExpirationCell, {
      props: {
        expiresAt: null
      }
    })
    expect(noExpirationWrapper.text()).toContain('admin.subscriptions.noExpiration')

    const statusWrapper = mount(SubscriptionStatusBadge, {
      props: {
        status: 'revoked'
      }
    })
    expect(statusWrapper.text()).toContain('admin.subscriptions.status.revoked')
  })

  it('emits action events', async () => {
    const subscription = createSubscription()
    const wrapper = mount(SubscriptionActionsCell, {
      props: {
        subscription,
        resetting: false
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('adjust')?.[0]).toEqual([subscription])
    expect(wrapper.emitted('reset-quota')?.[0]).toEqual([subscription])
    expect(wrapper.emitted('revoke')?.[0]).toEqual([subscription])
  })

  it('renders guide modal content', () => {
    const wrapper = mount(SubscriptionGuideModal, {
      props: {
        show: true
      },
      global: {
        stubs: {
          teleport: true,
          transition: false,
          Icon: true,
          RouterLink: {
            props: ['to'],
            template: '<a><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('admin.subscriptions.guide.title')
    expect(wrapper.text()).toContain('admin.subscriptions.guide.step1.title')
    expect(wrapper.text()).toContain('admin.subscriptions.guide.actions.adjust')
  })
})
