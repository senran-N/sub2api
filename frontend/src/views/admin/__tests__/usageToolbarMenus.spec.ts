import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UsageColumnSettingsControl from '../usage/UsageColumnSettingsControl.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('usage toolbar menus', () => {
  it('renders column settings and emits toggles', async () => {
    const wrapper = mount(UsageColumnSettingsControl, {
      props: {
        toggleableColumns: [
          { key: 'model', label: 'Model', sortable: false },
          { key: 'endpoint', label: 'Endpoint', sortable: false }
        ],
        isColumnVisible: (key: string) => key === 'model'
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    await wrapper.find('button[title="admin.users.columnSettings"]').trigger('click')
    const buttons = wrapper.findAll('.menu-item')
    await buttons[1].trigger('click')

    expect(wrapper.emitted('toggle-column')?.[0]).toEqual(['endpoint'])
  })
})
