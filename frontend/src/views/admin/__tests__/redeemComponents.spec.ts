import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import RedeemActionsCell from '../redeem/RedeemActionsCell.vue'
import RedeemCodeCell from '../redeem/RedeemCodeCell.vue'
import RedeemGeneratedResultDialog from '../redeem/RedeemGeneratedResultDialog.vue'
import RedeemGenerateDialog from '../redeem/RedeemGenerateDialog.vue'
import RedeemStatusBadge from '../redeem/RedeemStatusBadge.vue'
import RedeemToolbar from '../redeem/RedeemToolbar.vue'
import RedeemTypeBadge from '../redeem/RedeemTypeBadge.vue'
import RedeemValueCell from '../redeem/RedeemValueCell.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

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

describe('redeem local components', () => {
  it('renders toolbar and emits search/filter/action events', async () => {
    const wrapper = mount(RedeemToolbar, {
      props: {
        searchQuery: 'abc',
        type: '',
        status: '',
        typeOptions: [
          { value: '', label: 'All' },
          { value: 'subscription', label: 'Subscription' }
        ],
        statusOptions: [
          { value: '', label: 'All' },
          { value: 'unused', label: 'Unused' }
        ],
        loading: false
      },
      global: {
        stubs: {
          Icon: true,
          Select: SelectStub
        }
      }
    })

    await wrapper.find('input').setValue('vip')
    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')
    await buttons[3].trigger('click')
    await buttons[4].trigger('click')

    expect(wrapper.emitted('update:searchQuery')?.[0]).toEqual(['vip'])
    expect(wrapper.emitted('search')?.length).toBe(1)
    expect(wrapper.emitted('update:type')?.[0]).toEqual(['subscription'])
    expect(wrapper.emitted('update:status')?.[0]).toEqual(['unused'])
    expect(wrapper.emitted('refresh')?.length).toBe(3)
    expect(wrapper.emitted('export')?.length).toBe(1)
    expect(wrapper.emitted('generate')?.length).toBe(1)
  })

  it('renders delete action only when allowed', async () => {
    const deletable = mount(RedeemActionsCell, {
      props: { showDelete: true }
    })
    await deletable.find('button').trigger('click')
    expect(deletable.emitted('delete')?.length).toBe(1)

    const locked = mount(RedeemActionsCell, {
      props: { showDelete: false }
    })
    expect(locked.text()).toContain('-')
  })

  it('renders redeem table cell components', async () => {
    const codeCell = mount(RedeemCodeCell, {
      props: {
        code: 'VIP-123',
        copied: false
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })
    await codeCell.find('button').trigger('click')
    expect(codeCell.emitted('copy')?.length).toBe(1)

    const typeBadge = mount(RedeemTypeBadge, {
      props: {
        type: 'subscription'
      }
    })
    expect(typeBadge.text()).toContain('admin.redeem.types.subscription')

    const valueCell = mount(RedeemValueCell, {
      props: {
        code: {
          id: 1,
          code: 'VIP-123',
          type: 'subscription',
          value: 10,
          status: 'unused',
          used_by: null,
          used_at: null,
          created_at: '2026-04-05T00:00:00Z',
          validity_days: 30,
          group: {
            id: 1,
            name: 'Pro'
          }
        }
      }
    })
    expect(valueCell.text()).toContain('30')
    expect(valueCell.text()).toContain('Pro')

    const statusBadge = mount(RedeemStatusBadge, {
      props: {
        status: 'used'
      }
    })
    expect(statusBadge.text()).toContain('admin.redeem.status.used')
  })

  it('renders generate dialog modes and emits submit', async () => {
    const wrapper = mount(RedeemGenerateDialog, {
      props: {
        show: true,
        form: {
          type: 'subscription',
          value: 10,
          count: 1,
          group_id: 1,
          validity_days: 30
        },
        typeOptions: [
          { value: 'balance', label: 'Balance' },
          { value: 'subscription', label: 'Subscription' }
        ],
        subscriptionGroupOptions: [
          {
            value: 1,
            label: 'Pro',
            description: 'subscription',
            platform: 'openai',
            subscriptionType: 'subscription',
            rate: 1
          }
        ],
        submitting: false
      },
      global: {
        stubs: {
          teleport: true,
          Select: SelectStub,
          GroupBadge: true,
          GroupOptionItem: true
        }
      }
    })

    expect(wrapper.text()).toContain('admin.redeem.selectGroup')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')?.length).toBe(1)
  })

  it('renders generated result dialog and emits close/copy/download', async () => {
    const wrapper = mount(RedeemGeneratedResultDialog, {
      props: {
        show: true,
        count: 2,
        codesText: 'AAA\nBBB',
        textareaHeight: '72px',
        copiedAll: false
      },
      global: {
        stubs: {
          teleport: true,
          Icon: true
        }
      }
    })

    expect((wrapper.find('textarea').element as HTMLTextAreaElement).value).toBe('AAA\nBBB')
    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('close')?.length).toBe(1)
    expect(wrapper.emitted('copy')?.length).toBe(1)
    expect(wrapper.emitted('download')?.length).toBe(1)
  })
})
