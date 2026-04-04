import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupSortOrderDialog from '../groups/GroupSortOrderDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: `
    <div v-if="show" class="dialog">
      <div class="title">{{ title }}</div>
      <slot />
      <slot name="footer" />
      <button class="dialog-close" @click="$emit('close')">close</button>
    </div>
  `
}

const VueDraggableStub = {
  props: ['modelValue'],
  emits: ['update:modelValue'],
  template: `
    <div class="draggable">
      <slot />
      <button
        class="reorder"
        @click="$emit('update:modelValue', [...modelValue].reverse())"
      >
        reorder
      </button>
    </div>
  `
}

describe('group sort order dialog', () => {
  it('renders groups and emits close/save/update actions', async () => {
    const groups = [
      { id: 1, name: 'Alpha', platform: 'anthropic' },
      { id: 2, name: 'Beta', platform: 'openai' }
    ] as any

    const wrapper = mount(GroupSortOrderDialog, {
      props: {
        show: true,
        groups,
        submitting: false
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          VueDraggable: VueDraggableStub,
          GroupPlatformBadge: {
            props: ['platform'],
            template: '<span class="platform">{{ platform }}</span>'
          },
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('Alpha')
    expect(wrapper.text()).toContain('Beta')

    await wrapper.get('.reorder').trigger('click')
    await wrapper.get('.btn-secondary').trigger('click')
    await wrapper.get('.btn-primary').trigger('click')
    await wrapper.get('.dialog-close').trigger('click')

    expect(wrapper.emitted('update:groups')?.[0]?.[0]).toEqual([...groups].reverse())
    expect(wrapper.emitted('close')).toHaveLength(2)
    expect(wrapper.emitted('save')).toHaveLength(1)
  })
})
