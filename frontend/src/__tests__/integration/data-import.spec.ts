import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import ImportDataModal from '@/components/admin/account/ImportDataModal.vue'

const { showError, showSuccess, importData } = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
  importData: vi.fn()
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      importData
    }
  }
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('ImportDataModal', () => {
  const BaseDialogStub = defineComponent({
    name: 'BaseDialogStub',
    props: {
      show: { type: Boolean, default: false }
    },
    template: '<div v-if="show"><slot /><slot name="footer" /></div>'
  })

  function createDeferred<T>() {
    let resolve!: (value: T) => void
    const promise = new Promise<T>((res) => {
      resolve = res
    })

    return { promise, resolve }
  }

  function mountModal() {
    return mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub
        }
      }
    })
  }

  function attachFile(input: ReturnType<typeof mount>['element'], file: File) {
    Object.defineProperty(input, 'files', {
      value: [file],
      configurable: true
    })
  }

  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    importData.mockReset()
  })

  it('未选择文件时提示错误', async () => {
    const wrapper = mountModal()

    await wrapper.find('form').trigger('submit')
    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportSelectFile')
  })

  it('无效 JSON 时提示解析失败', async () => {
    const wrapper = mountModal()

    const input = wrapper.find('input[type="file"]')
    const file = new File(['invalid json'], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve('invalid json')
    })
    attachFile(input.element, file)

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await Promise.resolve()

    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportParseFailed')
  })

  it('close-reopen 后忽略旧导入结果与 imported 事件', async () => {
    const importRequest = createDeferred({
      account_created: 1,
      account_failed: 0,
      proxy_created: 0,
      proxy_reused: 0,
      proxy_failed: 0,
      errors: []
    })
    importData.mockReturnValueOnce(importRequest.promise)

    const wrapper = mountModal()
    const input = wrapper.find('input[type="file"]')
    const file = new File(['{}'], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve('{}')
    })
    attachFile(input.element, file)

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({ show: true })
    await flushPromises()

    importRequest.resolve({
      account_created: 1,
      account_failed: 0,
      proxy_created: 0,
      proxy_reused: 0,
      proxy_failed: 0,
      errors: []
    })
    await flushPromises()

    expect(showSuccess).not.toHaveBeenCalled()
    expect(showError).not.toHaveBeenCalled()
    expect(wrapper.emitted('imported')).toBeFalsy()
    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('重新选文件后忽略旧导入结果', async () => {
    const importRequest = createDeferred({
      account_created: 1,
      account_failed: 0,
      proxy_created: 0,
      proxy_reused: 0,
      proxy_failed: 0,
      errors: []
    })
    importData.mockReturnValueOnce(importRequest.promise)

    const wrapper = mountModal()
    const input = wrapper.find('input[type="file"]')
    const oldFile = new File(['{}'], 'old.json', { type: 'application/json' })
    Object.defineProperty(oldFile, 'text', {
      value: () => Promise.resolve('{}')
    })
    attachFile(input.element, oldFile)

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    const newFile = new File(['{"fresh":true}'], 'new.json', { type: 'application/json' })
    Object.defineProperty(newFile, 'text', {
      value: () => Promise.resolve('{"fresh":true}')
    })
    attachFile(input.element, newFile)
    await input.trigger('change')
    await flushPromises()

    importRequest.resolve({
      account_created: 1,
      account_failed: 0,
      proxy_created: 0,
      proxy_reused: 0,
      proxy_failed: 0,
      errors: []
    })
    await flushPromises()

    expect(showSuccess).not.toHaveBeenCalled()
    expect(showError).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('new.json')
    expect(wrapper.emitted('imported')).toBeFalsy()
  })
})
