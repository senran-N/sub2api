import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import SyncFromCrsModal from '../SyncFromCrsModal.vue'

const {
  previewFromCrsMock,
  syncFromCrsMock,
  showSuccessMock,
  showErrorMock
} = vi.hoisted(() => ({
  previewFromCrsMock: vi.fn(),
  syncFromCrsMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      previewFromCrs: previewFromCrsMock,
      syncFromCrs: syncFromCrsMock
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: showSuccessMock,
    showError: showErrorMock
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

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: { type: Boolean, default: false }
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function makePreviewResult(name: string, id: string) {
  return {
    existing_accounts: [],
    new_accounts: [
      {
        crs_account_id: id,
        kind: 'oauth',
        name,
        platform: 'openai',
        type: 'oauth'
      }
    ]
  }
}

function makeSyncResult(overrides: Partial<Awaited<ReturnType<typeof syncFromCrsMock>>> = {}) {
  return {
    created: 1,
    updated: 0,
    skipped: 0,
    failed: 0,
    items: [],
    ...overrides
  }
}

function mountModal() {
  return mount(SyncFromCrsModal, {
    props: {
      show: true
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub
      }
    }
  })
}

async function fillCredentials(wrapper: ReturnType<typeof mountModal>, suffix = 'one') {
  await wrapper.get('#crs-base-url').setValue(`https://crs-${suffix}.example.com`)
  await wrapper.get('#crs-username').setValue(`user-${suffix}`)
  await wrapper.get('#crs-password').setValue(`password-${suffix}`)
}

describe('SyncFromCrsModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    previewFromCrsMock.mockResolvedValue(makePreviewResult('Imported Account', 'crs-1'))
    syncFromCrsMock.mockResolvedValue(makeSyncResult())
  })

  it('previews CRS accounts and enters preview step', async () => {
    const wrapper = mountModal()

    await fillCredentials(wrapper)
    await wrapper.get('#sync-from-crs-form').trigger('submit.prevent')
    await flushPromises()

    expect(previewFromCrsMock).toHaveBeenCalledWith({
      base_url: 'https://crs-one.example.com',
      username: 'user-one',
      password: 'password-one'
    })
    expect(wrapper.text()).toContain('Imported Account')
    expect(wrapper.text()).toContain('admin.accounts.syncNow')
  })

  it('ignores a stale preview result after close-reopen and newer preview', async () => {
    const firstPreview = createDeferred<ReturnType<typeof makePreviewResult>>()
    previewFromCrsMock.mockReturnValueOnce(firstPreview.promise)
    previewFromCrsMock.mockResolvedValueOnce(makePreviewResult('Fresh Preview Account', 'crs-2'))

    const wrapper = mountModal()

    await fillCredentials(wrapper, 'old')
    await wrapper.get('#sync-from-crs-form').trigger('submit.prevent')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({ show: true })
    await flushPromises()

    await fillCredentials(wrapper, 'new')
    await wrapper.get('#sync-from-crs-form').trigger('submit.prevent')
    await flushPromises()

    firstPreview.resolve(makePreviewResult('Stale Preview Account', 'crs-old'))
    await flushPromises()

    expect(wrapper.text()).toContain('Fresh Preview Account')
    expect(wrapper.text()).not.toContain('Stale Preview Account')
    expect(showErrorMock).not.toHaveBeenCalled()
  })

  it('ignores a stale sync success after close-reopen', async () => {
    const syncRequest = createDeferred<ReturnType<typeof makeSyncResult>>()
    syncFromCrsMock.mockReturnValueOnce(syncRequest.promise)

    const wrapper = mountModal()

    await fillCredentials(wrapper)
    await wrapper.get('#sync-from-crs-form').trigger('submit.prevent')
    await flushPromises()

    await wrapper.get('.btn-primary').trigger('click')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({ show: true })
    await flushPromises()

    syncRequest.resolve(makeSyncResult())
    await flushPromises()

    expect(showSuccessMock).not.toHaveBeenCalled()
    expect(showErrorMock).not.toHaveBeenCalled()
    expect(wrapper.emitted('synced')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
    expect(wrapper.find('#sync-from-crs-form').exists()).toBe(true)
  })
})
