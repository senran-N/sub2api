import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { Proxy } from '@/types'
import ProxyAccountCountCell from '../proxies/ProxyAccountCountCell.vue'
import ProxyActionsCell from '../proxies/ProxyActionsCell.vue'
import ProxyAddressCell from '../proxies/ProxyAddressCell.vue'
import ProxyAuthCell from '../proxies/ProxyAuthCell.vue'
import ProxyLatencyCell from '../proxies/ProxyLatencyCell.vue'
import ProxyLocationCell from '../proxies/ProxyLocationCell.vue'
import ProxyNameCell from '../proxies/ProxyNameCell.vue'
import ProxyProtocolBadge from '../proxies/ProxyProtocolBadge.vue'
import ProxySelectionCheckbox from '../proxies/ProxySelectionCheckbox.vue'
import ProxyStatusBadge from '../proxies/ProxyStatusBadge.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key
  })
}))

function createProxy(overrides: Partial<Proxy> = {}): Proxy {
  return {
    id: 1,
    name: 'Proxy A',
    protocol: 'socks5',
    host: 'proxy.local',
    port: 1080,
    username: 'alice',
    password: 'secret',
    status: 'active',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

describe('proxy table cells', () => {
  it('renders protocol and status badges', () => {
    const nameWrapper = mount(ProxyNameCell, {
      props: {
        value: 'Proxy A'
      }
    })
    expect(nameWrapper.text()).toContain('Proxy A')

    const protocolWrapper = mount(ProxyProtocolBadge, {
      props: {
        protocol: 'socks5h'
      }
    })
    expect(protocolWrapper.text()).toContain('SOCKS5H')

    const statusWrapper = mount(ProxyStatusBadge, {
      props: {
        status: 'inactive'
      }
    })
    expect(statusWrapper.text()).toContain('admin.accounts.status.inactive')
  })

  it('emits checkbox changes for selection controls', async () => {
    const wrapper = mount(ProxySelectionCheckbox, {
      props: {
        checked: false
      }
    })

    await wrapper.find('input').setValue(true)
    expect(wrapper.emitted('change')?.length).toBe(1)
  })

  it('renders address copy menu and emits address actions', async () => {
    const proxy = createProxy()
    const wrapper = mount(ProxyAddressCell, {
      props: {
        proxy,
        copyMenuOpen: true,
        copyFormats: [
          { label: 'fmt-1', value: 'value-1' },
          { label: 'fmt-2', value: 'value-2' }
        ]
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[0].trigger('contextmenu')
    await buttons[1].trigger('click')

    expect(wrapper.text()).toContain('proxy.local:1080')
    expect(wrapper.emitted('copy-url')?.[0]).toEqual([proxy])
    expect(wrapper.emitted('toggle-copy-menu')?.[0]).toEqual([proxy.id])
    expect(wrapper.emitted('copy-format')?.[0]).toEqual(['value-1'])
  })

  it('renders auth, location, and account count details', async () => {
    const proxy = createProxy({
      account_count: 3,
      country: 'US',
      city: 'Seattle',
      country_code: 'US'
    })

    const authWrapper = mount(ProxyAuthCell, {
      props: {
        proxy,
        passwordVisible: false
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })
    expect(authWrapper.text()).toContain('alice')
    expect(authWrapper.text()).toContain('••••••')
    await authWrapper.find('button').trigger('click')
    expect(authWrapper.emitted('toggle-password')?.[0]).toEqual([proxy.id])

    const locationWrapper = mount(ProxyLocationCell, {
      props: {
        proxy
      }
    })
    expect(locationWrapper.text()).toContain('US · Seattle')
    expect(locationWrapper.find('img').attributes('src')).toContain('/us.svg')

    const accountWrapper = mount(ProxyAccountCountCell, {
      props: {
        proxy
      }
    })
    expect(accountWrapper.text()).toContain('admin.groups.accountsCount')
    await accountWrapper.find('button').trigger('click')
    expect(accountWrapper.emitted('accounts')?.[0]).toEqual([proxy])
  })

  it('renders latency summary and quality badge', () => {
    const wrapper = mount(ProxyLatencyCell, {
      props: {
        proxy: createProxy({
          latency_ms: 120,
          quality_checked: 1,
          quality_grade: 'A',
          quality_score: 92,
          quality_status: 'healthy'
        })
      }
    })

    expect(wrapper.text()).toContain('120ms')
    expect(wrapper.text()).toContain('admin.proxies.qualityInline')
    expect(wrapper.text()).toContain('admin.proxies.qualityStatusHealthy')
  })

  it('emits action events from the action cell', async () => {
    const proxy = createProxy()
    const wrapper = mount(ProxyActionsCell, {
      props: {
        proxy,
        testing: false,
        qualityChecking: false
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
    await buttons[3].trigger('click')

    expect(wrapper.emitted('test')?.[0]).toEqual([proxy])
    expect(wrapper.emitted('quality-check')?.[0]).toEqual([proxy])
    expect(wrapper.emitted('edit')?.[0]).toEqual([proxy])
    expect(wrapper.emitted('delete')?.[0]).toEqual([proxy])
  })
})
