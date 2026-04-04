import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ProxyActionToolbar from '../proxies/ProxyActionToolbar.vue'
import ProxyCreateDialogFooter from '../proxies/ProxyCreateDialogFooter.vue'
import ProxyCreateModeTabs from '../proxies/ProxyCreateModeTabs.vue'
import ProxyEditDialogFooter from '../proxies/ProxyEditDialogFooter.vue'
import ProxyFilterFields from '../proxies/ProxyFilterFields.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key
  })
}))

describe('proxy toolbar components', () => {
  it('renders filter fields and emits search and filter updates', async () => {
    const wrapper = mount(ProxyFilterFields, {
      props: {
        searchQuery: 'edge',
        protocol: '',
        status: '',
        protocolOptions: [
          { value: '', label: 'All' },
          { value: 'http', label: 'HTTP' }
        ],
        statusOptions: [
          { value: '', label: 'All' },
          { value: 'active', label: 'Active' }
        ]
      },
      global: {
        stubs: {
          Icon: true,
          Select: {
            props: ['modelValue', 'options'],
            template:
              '<button class="select-stub" @click="$emit(\'update:modelValue\', options[1].value); $emit(\'change\')">{{ modelValue }}</button>'
          }
        }
      }
    })

    const input = wrapper.find('input')
    await input.setValue('gateway')
    expect(wrapper.emitted('update:searchQuery')?.[0]).toEqual(['gateway'])
    expect(wrapper.emitted('search-input')?.length).toBe(1)

    const selects = wrapper.findAll('.select-stub')
    await selects[0].trigger('click')
    await selects[1].trigger('click')

    expect(wrapper.emitted('update:protocol')?.[0]).toEqual(['http'])
    expect(wrapper.emitted('protocol-change')?.length).toBe(1)
    expect(wrapper.emitted('update:status')?.[0]).toEqual(['active'])
    expect(wrapper.emitted('status-change')?.length).toBe(1)
  })

  it('renders action toolbar and emits button actions', async () => {
    const wrapper = mount(ProxyActionToolbar, {
      props: {
        loading: false,
        batchTesting: false,
        batchQualityChecking: false,
        selectedCount: 2
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
    await buttons[4].trigger('click')
    await buttons[5].trigger('click')
    await buttons[6].trigger('click')

    expect(wrapper.emitted('refresh')?.length).toBe(1)
    expect(wrapper.emitted('batch-test')?.length).toBe(1)
    expect(wrapper.emitted('batch-quality-check')?.length).toBe(1)
    expect(wrapper.emitted('batch-delete')?.length).toBe(1)
    expect(wrapper.emitted('import')?.length).toBe(1)
    expect(wrapper.emitted('export')?.length).toBe(1)
    expect(wrapper.emitted('create')?.length).toBe(1)
  })

  it('renders create mode tabs and dialog footers', async () => {
    const tabsWrapper = mount(ProxyCreateModeTabs, {
      props: {
        modelValue: 'standard'
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })
    await tabsWrapper.findAll('button')[1].trigger('click')
    expect(tabsWrapper.emitted('update:modelValue')?.[0]).toEqual(['batch'])

    const createFooter = mount(ProxyCreateDialogFooter, {
      props: {
        mode: 'batch',
        submitting: false,
        validCount: 3
      },
      global: {
        stubs: {
          ProxyLoadingSpinnerIcon: true
        }
      }
    })
    await createFooter.findAll('button')[1].trigger('click')
    expect(createFooter.text()).toContain('admin.proxies.importProxies:{"count":3}')
    expect(createFooter.emitted('batch-create')?.length).toBe(1)

    const editFooter = mount(ProxyEditDialogFooter, {
      props: {
        showSubmit: true,
        submitting: true
      },
      global: {
        stubs: {
          ProxyLoadingSpinnerIcon: true
        }
      }
    })
    expect(editFooter.text()).toContain('admin.proxies.updating')
    await editFooter.find('button').trigger('click')
    expect(editFooter.emitted('close')?.length).toBe(1)
  })
})
