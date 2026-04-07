import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import DataManagementView from '../DataManagementView.vue'

const { useDataManagementSoraProfiles } = vi.hoisted(() => ({
  useDataManagementSoraProfiles: vi.fn()
}))

vi.mock('../dataManagement/useSoraProfiles', () => ({
  useDataManagementSoraProfiles
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

function createComposable() {
  return {
    loadingSoraProfiles: ref(false),
    savingSoraProfile: ref(false),
    testingSoraProfile: ref(false),
    activatingSoraProfile: ref(false),
    deletingSoraProfile: ref(false),
    creatingSoraProfile: ref(false),
    soraProfileDrawerOpen: ref(false),
    soraS3Profiles: ref([]),
    soraProfileForm: ref({
      profile_id: 'main',
      name: 'Main',
      set_active: false,
      enabled: true,
      endpoint: '',
      region: '',
      bucket: '',
      access_key_id: '',
      secret_access_key: '',
      secret_access_key_configured: false,
      prefix: '',
      force_path_style: false,
      cdn_url: '',
      default_storage_quota_gb: 0
    }),
    loadSoraS3Profiles: vi.fn().mockResolvedValue(undefined),
    startCreateSoraProfile: vi.fn(),
    editSoraProfile: vi.fn(),
    closeSoraProfileDrawer: vi.fn(),
    saveSoraProfile: vi.fn(),
    testSoraProfileConnection: vi.fn(),
    activateSoraProfile: vi.fn(),
    removeSoraProfile: vi.fn()
  }
}

describe('admin DataManagementView', () => {
  it('loads data on mount and wires child events to the feature composable', async () => {
    const composable = createComposable()
    useDataManagementSoraProfiles.mockReturnValue(composable)

    const wrapper = mount(DataManagementView, {
      global: {
        stubs: {
          SoraProfilesCard: {
            template: `
              <div>
                <button class="create" @click="$emit('create')" />
                <button class="reload" @click="$emit('reload')" />
                <button class="edit" @click="$emit('edit', 'secondary')" />
                <button class="activate" @click="$emit('activate', 'secondary')" />
                <button class="remove" @click="$emit('remove', 'secondary')" />
              </div>
            `
          },
          SoraProfileDrawer: {
            template: `
              <div>
                <button class="close" @click="$emit('close')" />
                <button class="test" @click="$emit('test')" />
                <button class="save" @click="$emit('save')" />
              </div>
            `
          }
        }
      }
    })

    await flushPromises()

    expect(composable.loadSoraS3Profiles).toHaveBeenCalledTimes(1)

    await wrapper.find('button.create').trigger('click')
    await wrapper.find('button.reload').trigger('click')
    await wrapper.find('button.edit').trigger('click')
    await wrapper.find('button.activate').trigger('click')
    await wrapper.find('button.remove').trigger('click')
    await wrapper.find('button.close').trigger('click')
    await wrapper.find('button.test').trigger('click')
    await wrapper.find('button.save').trigger('click')

    expect(composable.startCreateSoraProfile).toHaveBeenCalledTimes(1)
    expect(composable.loadSoraS3Profiles).toHaveBeenCalledTimes(2)
    expect(composable.editSoraProfile).toHaveBeenCalledWith('secondary')
    expect(composable.activateSoraProfile).toHaveBeenCalledWith('secondary')
    expect(composable.removeSoraProfile).toHaveBeenCalledWith('secondary')
    expect(composable.closeSoraProfileDrawer).toHaveBeenCalledTimes(1)
    expect(composable.testSoraProfileConnection).toHaveBeenCalledTimes(1)
    expect(composable.saveSoraProfile).toHaveBeenCalledTimes(1)
  })
})
