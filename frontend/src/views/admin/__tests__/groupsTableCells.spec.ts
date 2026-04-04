import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupAccountCountCell from '../groups/GroupAccountCountCell.vue'
import GroupActionsCell from '../groups/GroupActionsCell.vue'
import GroupBillingTypeCell from '../groups/GroupBillingTypeCell.vue'
import GroupCapacityCell from '../groups/GroupCapacityCell.vue'
import GroupExclusivityBadge from '../groups/GroupExclusivityBadge.vue'
import GroupRateMultiplierCell from '../groups/GroupRateMultiplierCell.vue'
import GroupStatusBadge from '../groups/GroupStatusBadge.vue'
import GroupUsageCell from '../groups/GroupUsageCell.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('group table cells', () => {
  it('renders subscription billing details', () => {
    const wrapper = mount(GroupBillingTypeCell, {
      props: {
        group: {
          subscription_type: 'subscription',
          daily_limit_usd: 5,
          weekly_limit_usd: 10,
          monthly_limit_usd: 20
        } as any
      }
    })

    expect(wrapper.text()).toContain('admin.groups.subscription.subscription')
    expect(wrapper.text()).toContain('$5/admin.groups.limitDay')
    expect(wrapper.text()).toContain('$10/admin.groups.limitWeek')
    expect(wrapper.text()).toContain('$20/admin.groups.limitMonth')
  })

  it('renders account count summary', () => {
    const wrapper = mount(GroupAccountCountCell, {
      props: {
        group: {
          active_account_count: 8,
          rate_limited_account_count: 3,
          account_count: 12
        } as any
      }
    })

    expect(wrapper.text()).toContain('admin.groups.accountsAvailable')
    expect(wrapper.text()).toContain('5')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).toContain('12')
  })

  it('renders rate multiplier and exclusivity badges', () => {
    const rateWrapper = mount(GroupRateMultiplierCell, {
      props: {
        rateMultiplier: 1.75
      }
    })
    expect(rateWrapper.text()).toContain('1.75x')

    const exclusiveWrapper = mount(GroupExclusivityBadge, {
      props: {
        exclusive: true
      }
    })
    expect(exclusiveWrapper.text()).toContain('admin.groups.exclusive')

    const publicWrapper = mount(GroupExclusivityBadge, {
      props: {
        exclusive: false
      }
    })
    expect(publicWrapper.text()).toContain('admin.groups.public')
  })

  it('renders capacity summary and empty state', () => {
    const emptyWrapper = mount(GroupCapacityCell)
    expect(emptyWrapper.text()).toContain('—')

    const wrapper = mount(GroupCapacityCell, {
      props: {
        capacity: {
          concurrencyUsed: 2,
          concurrencyMax: 5,
          sessionsUsed: 1,
          sessionsMax: 3,
          rpmUsed: 4,
          rpmMax: 10
        }
      }
    })

    expect(wrapper.text()).toContain('2')
    expect(wrapper.text()).toContain('5')
    expect(wrapper.text()).toContain('1')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).toContain('4')
    expect(wrapper.text()).toContain('10')
  })

  it('renders usage summary and loading state', () => {
    const loadingWrapper = mount(GroupUsageCell, {
      props: {
        loading: true
      }
    })
    expect(loadingWrapper.text()).toContain('—')

    const wrapper = mount(GroupUsageCell, {
      props: {
        loading: false,
        summary: {
          today_cost: 12.345,
          total_cost: 123.456
        }
      }
    })

    expect(wrapper.text()).toContain('admin.groups.usageToday')
    expect(wrapper.text()).toContain('$12.35')
    expect(wrapper.text()).toContain('$123.5')
  })

  it('renders status badge text', () => {
    const wrapper = mount(GroupStatusBadge, {
      props: {
        status: 'inactive'
      }
    })

    expect(wrapper.text()).toContain('admin.accounts.status.inactive')
  })

  it('emits action events with the row group', async () => {
    const group = { id: 9, name: 'Alpha' } as any
    const wrapper = mount(GroupActionsCell, {
      props: {
        group
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

    expect(wrapper.emitted('edit')?.[0]).toEqual([group])
    expect(wrapper.emitted('rate-multipliers')?.[0]).toEqual([group])
    expect(wrapper.emitted('delete')?.[0]).toEqual([group])
  })
})
