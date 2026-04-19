import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import PlatformIcon from '../PlatformIcon.vue'

describe('PlatformIcon', () => {
  it('renders the dedicated Grok icon instead of the generic fallback', () => {
    const wrapper = mount(PlatformIcon, {
      props: {
        platform: 'grok',
        size: 'xs'
      }
    })

    expect(wrapper.html()).toContain("7.978-5.897")
  })
})
