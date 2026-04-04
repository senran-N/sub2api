import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import SubscriptionFiltersBar from '../subscriptions/SubscriptionFiltersBar.vue'
import SubscriptionFilterUserSearch from '../subscriptions/SubscriptionFilterUserSearch.vue'
import SubscriptionToolbarActions from '../subscriptions/SubscriptionToolbarActions.vue'
import SubscriptionColumnSettingsMenu from '../subscriptions/SubscriptionColumnSettingsMenu.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('subscription toolbar components', () => {
  it('renders filters bar and wires filter and action events', async () => {
    const SelectStub = {
      props: ['modelValue', 'options'],
      emits: ['update:modelValue', 'change'],
      template: `
        <button
          class="select-stub"
          @click="
            $emit('update:modelValue', options[1]?.value ?? modelValue);
            $emit('change', options[1]?.value ?? modelValue)
          "
        >
          {{ modelValue }}
        </button>
      `
    }

    const FilterStub = {
      props: ['keyword'],
      emits: ['update:keyword', 'search', 'focus', 'select-user', 'clear-user'],
      template: `
        <div>
          <button class="filter-keyword" @click="$emit('update:keyword', 'bob')">keyword</button>
          <button class="filter-search" @click="$emit('search')">search</button>
          <button class="filter-focus" @click="$emit('focus')">focus</button>
          <button class="filter-select" @click="$emit('select-user', { id: 3, email: 'bob@example.com' })">select</button>
          <button class="filter-clear" @click="$emit('clear-user')">clear</button>
        </div>
      `
    }

    const ActionsStub = {
      emits: ['refresh', 'set-user-mode', 'toggle-column', 'guide', 'assign'],
      template: `
        <div>
          <button class="actions-refresh" @click="$emit('refresh')">refresh</button>
          <button class="actions-mode" @click="$emit('set-user-mode', 'username')">mode</button>
          <button class="actions-toggle" @click="$emit('toggle-column', 'status')">toggle</button>
          <button class="actions-guide" @click="$emit('guide')">guide</button>
          <button class="actions-assign" @click="$emit('assign')">assign</button>
        </div>
      `
    }

    const wrapper = mount(SubscriptionFiltersBar, {
      props: {
        filterUserKeyword: 'ali',
        filterUserResults: [{ id: 2, email: 'alice@example.com' }],
        filterUserLoading: false,
        showFilterUserDropdown: true,
        selectedFilterUser: { id: 2, email: 'alice@example.com' },
        status: '',
        groupId: '',
        platform: '',
        statusOptions: [
          { value: '', label: 'All' },
          { value: 'active', label: 'Active' }
        ],
        groupOptions: [
          { value: '', label: 'All' },
          { value: '1', label: 'Group 1' }
        ],
        platformFilterOptions: [
          { value: '', label: 'All' },
          { value: 'openai', label: 'OpenAI' }
        ],
        loading: false,
        userColumnMode: 'email',
        toggleableColumns: [
          { key: 'status', label: 'Status', sortable: false }
        ],
        isColumnVisible: () => true
      },
      global: {
        stubs: {
          Select: SelectStub,
          SubscriptionFilterUserSearch: FilterStub,
          SubscriptionToolbarActions: ActionsStub
        }
      }
    })

    await wrapper.find('.filter-keyword').trigger('click')
    await wrapper.find('.filter-search').trigger('click')
    await wrapper.find('.filter-focus').trigger('click')
    await wrapper.find('.filter-select').trigger('click')
    await wrapper.find('.filter-clear').trigger('click')

    const selects = wrapper.findAll('.select-stub')
    await selects[0].trigger('click')
    await selects[1].trigger('click')
    await selects[2].trigger('click')

    await wrapper.find('.actions-refresh').trigger('click')
    await wrapper.find('.actions-mode').trigger('click')
    await wrapper.find('.actions-toggle').trigger('click')
    await wrapper.find('.actions-guide').trigger('click')
    await wrapper.find('.actions-assign').trigger('click')

    expect(wrapper.emitted('update:filterUserKeyword')?.[0]).toEqual(['bob'])
    expect(wrapper.emitted('search-filter-users')?.length).toBe(1)
    expect(wrapper.emitted('show-filter-user-dropdown')?.length).toBe(1)
    expect(wrapper.emitted('select-filter-user')?.[0]).toEqual([{ id: 3, email: 'bob@example.com' }])
    expect(wrapper.emitted('clear-filter-user')?.length).toBe(1)
    expect(wrapper.emitted('update:status')?.[0]).toEqual(['active'])
    expect(wrapper.emitted('update:groupId')?.[0]).toEqual(['1'])
    expect(wrapper.emitted('update:platform')?.[0]).toEqual(['openai'])
    expect(wrapper.emitted('apply-filters')?.length).toBe(3)
    expect(wrapper.emitted('refresh')?.length).toBe(1)
    expect(wrapper.emitted('set-user-mode')?.[0]).toEqual(['username'])
    expect(wrapper.emitted('toggle-column')?.[0]).toEqual(['status'])
    expect(wrapper.emitted('guide')?.length).toBe(1)
    expect(wrapper.emitted('assign')?.length).toBe(1)
  })

  it('renders filter user search and emits interactions', async () => {
    const wrapper = mount(SubscriptionFilterUserSearch, {
      props: {
        keyword: 'ali',
        results: [{ id: 2, email: 'alice@example.com' }],
        loading: false,
        showDropdown: true,
        selectedUser: { id: 2, email: 'alice@example.com' }
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    await wrapper.find('input').setValue('bob')
    await wrapper.find('button[title="common.clear"]').trigger('click')
    const resultButtons = wrapper.findAll('button')
    await resultButtons[resultButtons.length - 1].trigger('click')

    expect(wrapper.emitted('update:keyword')?.[0]).toEqual(['bob'])
    expect(wrapper.emitted('search')?.length).toBe(1)
    expect(wrapper.emitted('clear-user')?.length).toBe(1)
    expect(wrapper.emitted('select-user')?.[0]).toEqual([{ id: 2, email: 'alice@example.com' }])
  })

  it('renders toolbar actions and re-emits column menu events', async () => {
    const wrapper = mount(SubscriptionToolbarActions, {
      props: {
        loading: false,
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
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    expect(wrapper.findComponent(SubscriptionColumnSettingsMenu).exists()).toBe(true)

    const menuButtons = wrapper.findComponent(SubscriptionColumnSettingsMenu).findAll('button')
    await menuButtons[1].trigger('click')
    await menuButtons[2].trigger('click')
    await buttons[2].trigger('click')
    await buttons[3].trigger('click')

    expect(wrapper.emitted('refresh')?.length).toBe(1)
    expect(wrapper.emitted('set-user-mode')?.[0]).toEqual(['username'])
    expect(wrapper.emitted('toggle-column')?.[0]).toEqual(['usage'])
    expect(wrapper.emitted('guide')?.length).toBe(1)
    expect(wrapper.emitted('assign')?.length).toBe(1)
  })
})
