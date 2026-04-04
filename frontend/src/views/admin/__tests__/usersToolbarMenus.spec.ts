import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { UserAttributeDefinition } from '@/types'
import UserColumnSettingsMenu from '../users/UserColumnSettingsMenu.vue'
import UserFilterFields from '../users/UserFilterFields.vue'
import UserFilterSettingsMenu from '../users/UserFilterSettingsMenu.vue'
import UserToolbarActions from '../users/UserToolbarActions.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

function createAttribute(overrides: Partial<UserAttributeDefinition> = {}): UserAttributeDefinition {
  return {
    id: 1,
    key: 'department',
    name: 'Department',
    type: 'text',
    enabled: true,
    options: [],
    required: false,
    validation: {},
    placeholder: '',
    display_order: 0,
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    ...overrides
  }
}

describe('user toolbar menus', () => {
  it('renders filter fields and emits search updates', async () => {
    const updateAttributeFilter = vi.fn()
    const applyFilter = vi.fn()
    const wrapper = mount(UserFilterFields, {
      props: {
        searchQuery: 'alice',
        filters: {
          role: 'admin',
          status: '',
          group: ''
        },
        visibleFilters: new Set(['role', 'attr_1']),
        groupFilterOptions: [{ value: '', label: 'All' }],
        activeAttributeFilters: { 1: 'ops' },
        getAttributeDefinition: (attrId: number) =>
          attrId === 1 ? createAttribute() : undefined,
        getAttributeDefinitionName: () => 'Department',
        updateAttributeFilter,
        applyFilter
      },
      global: {
        stubs: {
          Select: {
            props: ['modelValue', 'options'],
            template: '<div class="select-stub">{{ modelValue }}</div>'
          },
          Icon: true
        }
      }
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('bob')
    expect(wrapper.emitted('update:searchQuery')?.[0]).toEqual(['bob'])
    expect(wrapper.emitted('search-input')?.length).toBe(1)

    await inputs[1].setValue('finance')
    expect(updateAttributeFilter).toHaveBeenCalledWith(1, 'finance')
  })

  it('renders filter settings menu and emits toggles', async () => {
    const attr = createAttribute({ id: 9, name: 'Team' })
    const wrapper = mount(UserFilterSettingsMenu, {
      props: {
        visibleFilters: new Set(['role', 'attr_9']),
        builtInFilters: [
          { key: 'role', name: 'Role' },
          { key: 'status', name: 'Status' }
        ],
        filterableAttributes: [attr]
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('toggle-built-in-filter')?.[0]).toEqual(['role'])
    expect(wrapper.emitted('toggle-attribute-filter')?.[0]).toEqual([attr])
  })

  it('renders column settings menu and emits column toggles', async () => {
    const wrapper = mount(UserColumnSettingsMenu, {
      props: {
        toggleableColumns: [
          { key: 'notes', label: 'Notes', sortable: false },
          { key: 'usage', label: 'Usage', sortable: false }
        ],
        isColumnVisible: (key: string) => key === 'notes'
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[1].trigger('click')

    expect(wrapper.emitted('toggle-column')?.[0]).toEqual(['usage'])
  })

  it('renders toolbar actions and emits toolbar events', async () => {
    const attr = createAttribute({ id: 4, name: 'Region' })
    const wrapper = mount(UserToolbarActions, {
      props: {
        loading: false,
        visibleFilters: new Set(['role', 'attr_4']),
        builtInFilters: [
          { key: 'role', name: 'Role' },
          { key: 'status', name: 'Status' }
        ],
        filterableAttributes: [attr],
        toggleableColumns: [
          { key: 'notes', label: 'Notes', sortable: false }
        ],
        isColumnVisible: () => true
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
    await buttons[3].trigger('click')
    await buttons[4].trigger('click')

    expect(wrapper.emitted('refresh')?.length).toBe(1)
    expect(wrapper.findComponent(UserFilterSettingsMenu).exists()).toBe(true)
    expect(wrapper.findComponent(UserColumnSettingsMenu).exists()).toBe(false)
    expect(wrapper.emitted('open-attributes')?.length).toBe(1)
    expect(wrapper.emitted('create')?.length).toBe(1)

    await wrapper.findComponent(UserFilterSettingsMenu).find('button').trigger('click')
    expect(wrapper.emitted('toggle-built-in-filter')?.[0]).toEqual(['role'])

    await buttons[2].trigger('click')
    expect(wrapper.findComponent(UserColumnSettingsMenu).exists()).toBe(true)

    const columnMenuButtons = wrapper.findComponent(UserColumnSettingsMenu).findAll('button')
    await columnMenuButtons[0].trigger('click')
    expect(wrapper.emitted('toggle-column')?.[0]).toEqual(['notes'])
  })
})
