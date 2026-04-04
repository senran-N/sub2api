import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import type { SoraS3Profile } from '@/api/admin/settings'
import { createDefaultSoraS3ProfileForm } from '../dataManagementView'
import SoraProfileDrawer from '../datamanagement/SoraProfileDrawer.vue'
import SoraProfilesCard from '../datamanagement/SoraProfilesCard.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

function createProfile(overrides: Partial<SoraS3Profile> = {}): SoraS3Profile {
  return {
    profile_id: 'main',
    name: 'Main',
    is_active: true,
    enabled: true,
    endpoint: 'https://example.com',
    region: 'auto',
    bucket: 'bucket',
    access_key_id: 'AK',
    secret_access_key_configured: true,
    prefix: 'sora/',
    force_path_style: false,
    cdn_url: '',
    default_storage_quota_bytes: 5 * 1024 * 1024 * 1024,
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('data management local components', () => {
  it('renders profile card and emits toolbar and row actions', async () => {
    const wrapper = mount(SoraProfilesCard, {
      props: {
        profiles: [
          createProfile(),
          createProfile({ profile_id: 'secondary', is_active: false })
        ],
        loading: false,
        activating: false,
        deleting: false
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

    expect(wrapper.emitted('create')?.length).toBe(1)
    expect(wrapper.emitted('reload')?.length).toBe(1)
    expect(wrapper.emitted('edit')?.[0]).toEqual(['main'])
    expect(wrapper.emitted('remove')?.[0]).toEqual(['main'])
    expect(wrapper.emitted('edit')?.[1]).toEqual(['secondary'])
    expect(wrapper.emitted('activate')?.[0]).toEqual(['secondary'])
    expect(wrapper.emitted('remove')?.[1]).toEqual(['secondary'])
  })

  it('renders empty profile state', () => {
    const wrapper = mount(SoraProfilesCard, {
      props: {
        profiles: [],
        loading: false,
        activating: false,
        deleting: false
      }
    })

    expect(wrapper.text()).toContain('admin.settings.soraS3.empty')
  })

  it('renders drawer and emits close, test, and save', async () => {
    const form = createDefaultSoraS3ProfileForm()
    form.enabled = true

    const wrapper = mount(SoraProfileDrawer, {
      props: {
        open: true,
        creating: true,
        saving: false,
        testing: false,
        form
      },
      global: {
        stubs: {
          Teleport: true,
          Transition: false
        }
      }
    })

    expect(wrapper.text()).toContain('admin.settings.soraS3.createTitle')

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')
    await buttons[3].trigger('click')

    expect(wrapper.emitted('close')?.length).toBe(2)
    expect(wrapper.emitted('test')?.length).toBe(1)
    expect(wrapper.emitted('save')?.length).toBe(1)
  })
})
