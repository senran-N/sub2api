import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AccountAdminToolsButtons from '../accounts/AccountAdminToolsButtons.vue'
import AccountAutoRefreshMenu from '../accounts/AccountAutoRefreshMenu.vue'
import AccountColumnSettingsMenu from '../accounts/AccountColumnSettingsMenu.vue'
import AccountExportDialogOptions from '../accounts/AccountExportDialogOptions.vue'
import AccountPendingSyncBanner from '../accounts/AccountPendingSyncBanner.vue'
import AccountSecondaryActions from '../accounts/AccountSecondaryActions.vue'
import AccountSelectionCheckbox from '../accounts/AccountSelectionCheckbox.vue'
import AccountToolbarControls from '../accounts/AccountToolbarControls.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('account toolbar menus', () => {
  it('renders auto refresh menu and emits interactions', async () => {
    const wrapper = mount(AccountAutoRefreshMenu, {
      props: {
        enabled: true,
        intervals: [5, 10, 15],
        selectedIntervalSeconds: 10,
        labelForInterval: (seconds: number) => `${seconds}s`
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

    expect(wrapper.emitted('toggle-enabled')?.length).toBe(1)
    expect(wrapper.emitted('set-interval')?.[0]).toEqual([10])
  })

  it('renders column settings and emits toggles', async () => {
    const wrapper = mount(AccountColumnSettingsMenu, {
      props: {
        toggleableColumns: [
          { key: 'usage', label: 'Usage', sortable: false },
          { key: 'notes', label: 'Notes', sortable: false }
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

    expect(wrapper.emitted('toggle-column')?.[0]).toEqual(['notes'])
  })

  it('renders export dialog options and emits model updates', async () => {
    const wrapper = mount(AccountExportDialogOptions, {
      props: {
        modelValue: false
      }
    })

    const checkbox = wrapper.find('input')
    await checkbox.setValue(true)

    expect(wrapper.text()).toContain('admin.accounts.dataExportIncludeProxies')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([true])
  })

  it('renders secondary buttons and admin tools events', async () => {
    const secondaryWrapper = mount(AccountSecondaryActions, {
      props: {
        selectedCount: 2
      }
    })

    const secondaryButtons = secondaryWrapper.findAll('button')
    await secondaryButtons[0].trigger('click')
    await secondaryButtons[1].trigger('click')

    expect(secondaryWrapper.text()).toContain('admin.accounts.dataExportSelected')
    expect(secondaryWrapper.emitted('import')?.length).toBe(1)
    expect(secondaryWrapper.emitted('export')?.length).toBe(1)

    const toolsWrapper = mount(AccountAdminToolsButtons, {
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const toolButtons = toolsWrapper.findAll('button')
    await toolButtons[0].trigger('click')
    await toolButtons[1].trigger('click')

    expect(toolsWrapper.emitted('error-passthrough')?.length).toBe(1)
    expect(toolsWrapper.emitted('tls-profiles')?.length).toBe(1)
  })

  it('coordinates toolbar dropdowns and re-emits actions', async () => {
    const wrapper = mount(AccountToolbarControls, {
      props: {
        autoRefreshEnabled: true,
        autoRefreshCountdown: 9,
        autoRefreshIntervals: [10, 30],
        autoRefreshIntervalSeconds: 10,
        autoRefreshIntervalLabel: (seconds: number) => `${seconds}s`,
        toggleableColumns: [
          { key: 'usage', label: 'Usage', sortable: false },
          { key: 'notes', label: 'Notes', sortable: false }
        ],
        isColumnVisible: (key: string) => key === 'usage'
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    await wrapper.find('button[title="admin.accounts.autoRefresh"]').trigger('click')
    expect(wrapper.findComponent(AccountAutoRefreshMenu).exists()).toBe(true)

    const autoRefreshButtons = wrapper.findComponent(AccountAutoRefreshMenu).findAll('button')
    await autoRefreshButtons[0].trigger('click')
    await autoRefreshButtons[1].trigger('click')

    await wrapper.find('button[title="admin.errorPassthrough.title"]').trigger('click')
    await wrapper.find('button[title="admin.tlsFingerprintProfiles.title"]').trigger('click')
    await wrapper.find('button[title="admin.users.columnSettings"]').trigger('click')

    expect(wrapper.findComponent(AccountAutoRefreshMenu).exists()).toBe(false)
    expect(wrapper.findComponent(AccountColumnSettingsMenu).exists()).toBe(true)

    const columnButtons = wrapper.findComponent(AccountColumnSettingsMenu).findAll('button')
    await columnButtons[1].trigger('click')

    expect(wrapper.emitted('toggle-auto-refresh-enabled')?.length).toBe(1)
    expect(wrapper.emitted('set-auto-refresh-interval')?.[0]).toEqual([10])
    expect(wrapper.emitted('error-passthrough')?.length).toBe(1)
    expect(wrapper.emitted('tls-profiles')?.length).toBe(1)
    expect(wrapper.emitted('toggle-column')?.[0]).toEqual(['notes'])
  })

  it('renders pending sync banner and selection checkbox interactions', async () => {
    const bannerWrapper = mount(AccountPendingSyncBanner)
    await bannerWrapper.find('button').trigger('click')
    expect(bannerWrapper.text()).toContain('admin.accounts.listPendingSyncHint')
    expect(bannerWrapper.emitted('sync')?.length).toBe(1)

    const checkboxWrapper = mount(AccountSelectionCheckbox, {
      props: {
        checked: false
      }
    })

    await checkboxWrapper.find('input').setValue(true)
    expect(checkboxWrapper.emitted('change')?.length).toBe(1)
  })
})
