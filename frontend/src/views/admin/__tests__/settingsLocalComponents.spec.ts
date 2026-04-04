import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import SettingsLoadingState from '../settings/SettingsLoadingState.vue'
import SettingsSaveBar from '../settings/SettingsSaveBar.vue'
import SettingsTabsNav from '../settings/SettingsTabsNav.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('settings local components', () => {
  it('renders loading spinner state', () => {
    const wrapper = mount(SettingsLoadingState)
    expect(wrapper.find('.animate-spin').exists()).toBe(true)
  })

  it('renders tabs and emits active tab updates', async () => {
    const wrapper = mount(SettingsTabsNav, {
      props: {
        activeTab: 'general',
        tabs: [
          { key: 'general', icon: 'home' },
          { key: 'security', icon: 'shield' }
        ]
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    expect(buttons[0].classes()).toContain('settings-tab-active')
    await buttons[1].trigger('click')
    expect(wrapper.emitted('update:activeTab')?.[0]).toEqual(['security'])
  })

  it('renders save bar states', () => {
    const wrapper = mount(SettingsSaveBar, {
      props: {
        saving: true,
        disabled: false
      }
    })

    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('admin.settings.saving')
    expect(wrapper.find('.animate-spin').exists()).toBe(true)
  })
})
