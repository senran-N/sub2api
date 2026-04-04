import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupImagePricingSection from '../groups/GroupImagePricingSection.vue'
import GroupSoraPricingSection from '../groups/GroupSoraPricingSection.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('group pricing sections', () => {
  it('binds image pricing fields for supported platforms', async () => {
    const form = {
      platform: 'antigravity',
      image_price_1k: null,
      image_price_2k: null,
      image_price_4k: null
    } as any

    const wrapper = mount(GroupImagePricingSection, {
      props: {
        form
      }
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('0.134')
    await inputs[1].setValue('0.201')
    await inputs[2].setValue('0.268')

    expect(form.image_price_1k).toBe(0.134)
    expect(form.image_price_2k).toBe(0.201)
    expect(form.image_price_4k).toBe(0.268)
  })

  it('binds sora pricing fields and storage quota through the shared section', async () => {
    const form = {
      platform: 'sora',
      sora_image_price_360: null,
      sora_image_price_540: null,
      sora_video_price_per_request: null,
      sora_video_price_per_request_hd: null,
      sora_storage_quota_gb: null
    } as any

    const wrapper = mount(GroupSoraPricingSection, {
      props: {
        form
      }
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('0.05')
    await inputs[1].setValue('0.08')
    await inputs[2].setValue('0.5')
    await inputs[3].setValue('0.8')
    await inputs[4].setValue('10')

    expect(form.sora_image_price_360).toBe(0.05)
    expect(form.sora_image_price_540).toBe(0.08)
    expect(form.sora_video_price_per_request).toBe(0.5)
    expect(form.sora_video_price_per_request_hd).toBe(0.8)
    expect(form.sora_storage_quota_gb).toBe(10)
  })
})
