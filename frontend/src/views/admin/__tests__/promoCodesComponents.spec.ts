import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import PromoCodeActionsCell from '../promocodes/PromoCodeActionsCell.vue'
import PromoCodeAmountCell from '../promocodes/PromoCodeAmountCell.vue'
import PromoCodeCodeCell from '../promocodes/PromoCodeCodeCell.vue'
import PromoCodeCreateDialog from '../promocodes/PromoCodeCreateDialog.vue'
import PromoCodeDateCell from '../promocodes/PromoCodeDateCell.vue'
import PromoCodeEditDialog from '../promocodes/PromoCodeEditDialog.vue'
import PromoCodeStatusBadge from '../promocodes/PromoCodeStatusBadge.vue'
import PromoCodeUsageCell from '../promocodes/PromoCodeUsageCell.vue'
import PromoCodeUsagesDialog from '../promocodes/PromoCodeUsagesDialog.vue'
import PromoCodesToolbar from '../promocodes/PromoCodesToolbar.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => `date:${value}`
}))

const BaseDialogStub = {
  props: ['show', 'title', 'width'],
  template: '<div><slot /><slot name="footer" /></div>'
}

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

const PaginationStub = {
  props: ['page', 'total', 'pageSize'],
  emits: ['update:page', 'update:page-size'],
  template: `
    <div>
      <button class="next-page" @click="$emit('update:page', page + 1)">next</button>
      <button class="resize-page" @click="$emit('update:page-size', 50)">resize</button>
    </div>
  `
}

describe('promo codes local components', () => {
  it('renders toolbar and emits search, filter, refresh, and create actions', async () => {
    const wrapper = mount(PromoCodesToolbar, {
      props: {
        searchQuery: 'welcome',
        status: '',
        statusOptions: [
          { value: '', label: 'All' },
          { value: 'active', label: 'Active' }
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

    await wrapper.find('input').setValue('bonus')
    await wrapper.find('.select-stub').trigger('click')
    const buttons = wrapper.findAll('button')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('update:searchQuery')?.[0]).toEqual(['bonus'])
    expect(wrapper.emitted('search')?.length).toBe(1)
    expect(wrapper.emitted('update:status')?.[0]).toEqual(['active'])
    expect(wrapper.emitted('refresh')?.length).toBe(2)
    expect(wrapper.emitted('create')?.length).toBe(1)
  })

  it('renders action cell and re-emits button clicks', async () => {
    const wrapper = mount(PromoCodeActionsCell, {
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

    expect(wrapper.emitted('copy-link')?.length).toBe(1)
    expect(wrapper.emitted('view-usages')?.length).toBe(1)
    expect(wrapper.emitted('edit')?.length).toBe(1)
    expect(wrapper.emitted('delete')?.length).toBe(1)
  })

  it('renders promo table cell components', async () => {
    const codeCell = mount(PromoCodeCodeCell, {
      props: {
        code: 'SPRING',
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

    expect(mount(PromoCodeAmountCell, {
      props: { amount: 8.5 }
    }).text()).toContain('$8.50')

    expect(mount(PromoCodeUsageCell, {
      props: { usedCount: 2, maxUses: 0 }
    }).text()).toContain('2 / ∞')

    expect(mount(PromoCodeStatusBadge, {
      props: {
        code: {
          id: 1,
          code: 'SPRING',
          bonus_amount: 8.5,
          max_uses: 10,
          used_count: 2,
          status: 'active',
          expires_at: null,
          notes: null,
          created_at: '2026-04-05T00:00:00Z',
          updated_at: '2026-04-05T00:00:00Z'
        }
      }
    }).text()).toContain('admin.promo.statusActive')

    expect(mount(PromoCodeDateCell, {
      props: {
        value: '2026-04-05T00:00:00Z'
      }
    }).text()).toContain('date:2026-04-05T00:00:00Z')
  })

  it('renders create and edit dialogs and submits form actions', async () => {
    const createWrapper = mount(PromoCodeCreateDialog, {
      props: {
        show: true,
        form: {
          code: '',
          bonus_amount: 1,
          max_uses: 0,
          expires_at_str: '',
          notes: ''
        },
        submitting: false
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub
        }
      }
    })

    await createWrapper.find('input').setValue('SPRING')
    await createWrapper.find('form').trigger('submit')
    expect(createWrapper.emitted('submit')?.length).toBe(1)

    const editWrapper = mount(PromoCodeEditDialog, {
      props: {
        show: true,
        form: {
          code: 'WELCOME',
          bonus_amount: 5,
          max_uses: 10,
          status: 'active',
          expires_at_str: '',
          notes: ''
        },
        statusOptions: [
          { value: 'active', label: 'Active' },
          { value: 'disabled', label: 'Disabled' }
        ],
        submitting: false
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub
        }
      }
    })

    await editWrapper.find('.select-stub').trigger('click')
    await editWrapper.find('form').trigger('submit')
    expect(editWrapper.emitted('submit')?.length).toBe(1)
  })

  it('renders usages dialog states and emits pagination events', async () => {
    const wrapper = mount(PromoCodeUsagesDialog, {
      props: {
        show: true,
        loading: false,
        usages: [
          {
            id: 1,
            promo_code_id: 2,
            user_id: 3,
            bonus_amount: 7.5,
            used_at: '2026-04-04T00:00:00Z',
            user: { email: 'alice@example.com' }
          }
        ],
        page: 1,
        pageSize: 10,
        total: 25
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Pagination: PaginationStub,
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('alice@example.com')
    expect(wrapper.text()).toContain('date:2026-04-04T00:00:00Z')
    await wrapper.find('.next-page').trigger('click')
    await wrapper.find('.resize-page').trigger('click')

    expect(wrapper.emitted('update:page')?.[0]).toEqual([2])
    expect(wrapper.emitted('update:page-size')?.[0]).toEqual([50])
  })
})
