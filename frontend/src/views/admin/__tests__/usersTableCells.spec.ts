import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { AdminUser } from '@/types'
import UserActionsCell from '../users/UserActionsCell.vue'
import UserActionMenu from '../users/UserActionMenu.vue'
import UserAttributeValueCell from '../users/UserAttributeValueCell.vue'
import UserBalanceCell from '../users/UserBalanceCell.vue'
import UserCreatedAtCell from '../users/UserCreatedAtCell.vue'
import UserEmailCell from '../users/UserEmailCell.vue'
import UserGroupsCell from '../users/UserGroupsCell.vue'
import UserNotesCell from '../users/UserNotesCell.vue'
import UserRoleCell from '../users/UserRoleCell.vue'
import UserStatusCell from '../users/UserStatusCell.vue'
import UserSubscriptionsCell from '../users/UserSubscriptionsCell.vue'
import UserUsernameCell from '../users/UserUsernameCell.vue'
import UserUsageCell from '../users/UserUsageCell.vue'

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
  formatDateTime: (value: string) => `formatted:${value}`
}))

function createUser(overrides: Partial<AdminUser> = {}): AdminUser {
  return {
    id: 1,
    email: 'user@example.com',
    username: 'alice',
    role: 'user',
    balance: 12.5,
    concurrency: 2,
    status: 'active',
    allowed_groups: [],
    subscriptions: [],
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    notes: 'Example notes for this user entry',
    group_rates: {},
    current_concurrency: 1,
    sora_storage_quota_bytes: 0,
    sora_storage_used_bytes: 0,
    ...overrides
  }
}

describe('user table cells', () => {
  it('renders email, notes, usage, and status cells', () => {
    const emailWrapper = mount(UserEmailCell, {
      props: {
        email: 'user@example.com'
      }
    })
    expect(emailWrapper.text()).toContain('user@example.com')
    expect(emailWrapper.text()).toContain('U')

    const notesWrapper = mount(UserNotesCell, {
      props: {
        notes: '1234567890123456789012345678901'
      }
    })
    expect(notesWrapper.text()).toContain('...')

    const usageWrapper = mount(UserUsageCell, {
      props: {
        usage: {
          today_actual_cost: 1.23456,
          total_actual_cost: 9.87654
        }
      }
    })
    expect(usageWrapper.text()).toContain('$1.2346')
    expect(usageWrapper.text()).toContain('$9.8765')

    const statusWrapper = mount(UserStatusCell, {
      props: {
        status: 'disabled'
      }
    })
    expect(statusWrapper.text()).toContain('admin.users.disabled')

    const usernameWrapper = mount(UserUsernameCell, {
      props: {
        value: ''
      }
    })
    expect(usernameWrapper.text()).toContain('-')

    const roleWrapper = mount(UserRoleCell, {
      props: {
        value: 'admin'
      }
    })
    expect(roleWrapper.text()).toContain('admin.users.roles.admin')

    const attributeWrapper = mount(UserAttributeValueCell, {
      props: {
        value: 'Operations'
      }
    })
    expect(attributeWrapper.text()).toContain('Operations')

    const createdAtWrapper = mount(UserCreatedAtCell, {
      props: {
        value: '2026-04-01T00:00:00Z'
      }
    })
    expect(createdAtWrapper.text()).toContain('formatted:2026-04-01T00:00:00Z')
  })

  it('renders subscriptions and empty state', () => {
    const wrapper = mount(UserSubscriptionsCell, {
      props: {
        user: createUser({
          subscriptions: [
            {
              id: 10,
              expires_at: '2026-04-10T00:00:00Z',
              group: {
                name: 'Premium',
                platform: 'openai',
                subscription_type: 'subscription',
                rate_multiplier: 1
              }
            }
          ] as any
        })
      },
      global: {
        stubs: {
          GroupBadge: {
            props: ['name', 'daysRemaining', 'title'],
            template: '<div>{{ name }} {{ daysRemaining }} {{ title }}</div>'
          },
          Icon: true
        }
      }
    })
    expect(wrapper.text()).toContain('Premium')
    expect(wrapper.text()).toContain('formatted:2026-04-10T00:00:00Z')

    const emptyWrapper = mount(UserSubscriptionsCell, {
      props: {
        user: createUser()
      },
      global: {
        stubs: {
          GroupBadge: true,
          Icon: true
        }
      }
    })
    expect(emptyWrapper.text()).toContain('admin.users.noSubscription')
  })

  it('emits balance and action events', async () => {
    const user = createUser()

    const balanceWrapper = mount(UserBalanceCell, {
      props: {
        user
      }
    })
    const balanceButtons = balanceWrapper.findAll('button')
    await balanceButtons[0].trigger('click')
    await balanceButtons[1].trigger('click')
    expect(balanceWrapper.emitted('history')?.[0]).toEqual([user])
    expect(balanceWrapper.emitted('deposit')?.[0]).toEqual([user])

    const actionsWrapper = mount(UserActionsCell, {
      props: {
        user,
        menuOpen: true
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })
    const actionButtons = actionsWrapper.findAll('button')
    await actionButtons[0].trigger('click')
    await actionButtons[1].trigger('click')
    await actionButtons[2].trigger('click')

    expect(actionsWrapper.emitted('edit')?.[0]).toEqual([user])
    expect(actionsWrapper.emitted('toggle-status')?.[0]).toEqual([user])
    expect(actionsWrapper.emitted('open-menu')?.[0]?.[0]).toEqual(user)
  })

  it('renders groups cell and emits group actions', async () => {
    const user = createUser()
    const wrapper = mount(UserGroupsCell, {
      props: {
        user,
        hasGroupsData: true,
        expanded: true,
        summary: {
          exclusive: [{ id: 9, name: 'Exclusive' }],
          publicGroups: [{ id: 10, name: 'Public' }]
        }
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('admin.users.exclusiveLabel')
    expect(wrapper.text()).toContain('admin.users.publicLabel')

    const clickable = wrapper.findAll('span, div').find((node) => node.text().includes('Exclusive'))
    if (clickable) {
      await clickable.trigger('click')
    }

    const entries = wrapper.findAll('div').filter((node) => node.text().includes('Exclusive'))
    const replaceEntry = entries[entries.length - 1]
    await replaceEntry.trigger('click')
    expect(wrapper.emitted('replace-group')?.[0]).toEqual([user, { id: 9, name: 'Exclusive' }])
  })

  it('renders action menu and emits menu actions', async () => {
    const user = createUser()
    const wrapper = mount(UserActionMenu, {
      props: {
        user,
        position: { top: 10, left: 20 }
      },
      global: {
        stubs: {
          teleport: true,
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')
    await buttons[3].trigger('click')
    await buttons[4].trigger('click')
    await buttons[5].trigger('click')

    expect(wrapper.emitted('api-keys')?.[0]).toEqual([user])
    expect(wrapper.emitted('groups')?.[0]).toEqual([user])
    expect(wrapper.emitted('deposit')?.[0]).toEqual([user])
    expect(wrapper.emitted('withdraw')?.[0]).toEqual([user])
    expect(wrapper.emitted('history')?.[0]).toEqual([user])
    expect(wrapper.emitted('delete')?.[0]).toEqual([user])
    expect(wrapper.emitted('close')?.length).toBe(6)
  })
})
