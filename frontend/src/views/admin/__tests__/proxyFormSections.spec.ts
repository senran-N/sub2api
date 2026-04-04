import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ProxyBatchInputSection from '../proxies/ProxyBatchInputSection.vue'
import ProxyBatchParseSummary from '../proxies/ProxyBatchParseSummary.vue'
import ProxyFormFieldsSection from '../proxies/ProxyFormFieldsSection.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key
  })
}))

describe('proxy form sections', () => {
  it('renders proxy form fields and emits password events', async () => {
    const form = {
      name: 'Edge',
      protocol: 'http' as const,
      host: 'proxy.local',
      port: 8080,
      username: 'alice',
      password: 'secret',
      status: 'inactive' as const
    }

    const wrapper = mount(ProxyFormFieldsSection, {
      props: {
        form,
        protocolOptions: [
          { value: 'http', label: 'HTTP' },
          { value: 'socks5', label: 'SOCKS5' }
        ],
        passwordVisible: false,
        passwordPlaceholder: 'keep',
        showStatus: true,
        statusOptions: [
          { value: 'active', label: 'Active' },
          { value: 'inactive', label: 'Inactive' }
        ]
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
    expect(inputs).toHaveLength(5)
    expect((inputs[0].element as HTMLInputElement).value).toBe('Edge')
    expect((inputs[4].element as HTMLInputElement).placeholder).toBe('keep')

    await inputs[4].trigger('input')
    expect(wrapper.emitted('password-input')?.length).toBe(1)

    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('toggle-password-visibility')?.length).toBe(1)
    expect(wrapper.text()).toContain('inactive')
  })

  it('renders batch parse summary only when there is parsed data', () => {
    const emptyWrapper = mount(ProxyBatchParseSummary, {
      props: {
        summary: {
          total: 0,
          valid: 0,
          invalid: 0,
          duplicate: 0,
          proxies: []
        }
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })
    expect(emptyWrapper.text()).toBe('')

    const wrapper = mount(ProxyBatchParseSummary, {
      props: {
        summary: {
          total: 5,
          valid: 3,
          invalid: 1,
          duplicate: 1,
          proxies: []
        }
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('admin.proxies.parsedCount')
    expect(wrapper.text()).toContain('admin.proxies.invalidCount')
    expect(wrapper.text()).toContain('admin.proxies.duplicateCount')
  })

  it('renders batch input section and emits input updates', async () => {
    const wrapper = mount(ProxyBatchInputSection, {
      props: {
        modelValue: 'http://1.2.3.4:8080',
        summary: {
          total: 1,
          valid: 1,
          invalid: 0,
          duplicate: 0,
          proxies: []
        }
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const textarea = wrapper.find('textarea')
    await textarea.setValue('socks5://proxy.local:1080')

    expect(wrapper.text()).toContain('admin.proxies.batchInputHint')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['socks5://proxy.local:1080'])
    expect(wrapper.emitted('input')?.length).toBe(1)
  })
})
