import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import UserAttributeForm from '../UserAttributeForm.vue'

const mockListEnabledDefinitions = vi.fn()
const mockGetUserAttributeValues = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    userAttributes: {
      listEnabledDefinitions: (...args: any[]) => mockListEnabledDefinitions(...args),
      getUserAttributeValues: (...args: any[]) => mockGetUserAttributeValues(...args)
    }
  }
}))

vi.mock('@/components/common/Select.vue', () => ({
  default: {
    name: 'SelectStub',
    props: ['modelValue', 'options'],
    emits: ['update:modelValue', 'change'],
    template: '<div class="select-stub" />'
  }
}))

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createDefinition() {
  return {
    id: 1,
    key: 'department',
    name: 'Department',
    description: '',
    type: 'text',
    options: [],
    required: false,
    validation: {},
    placeholder: '',
    display_order: 0,
    enabled: true,
    created_at: '2026-04-17T00:00:00Z',
    updated_at: '2026-04-17T00:00:00Z'
  }
}

describe('UserAttributeForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListEnabledDefinitions.mockResolvedValue([createDefinition()])
    mockGetUserAttributeValues.mockResolvedValue([])
  })

  it('clears local values immediately when switching users before new values load', async () => {
    const secondUserValues = createDeferred<Array<{ attribute_id: number; value: string }>>()
    mockGetUserAttributeValues
      .mockResolvedValueOnce([{ attribute_id: 1, value: 'alpha' }])
      .mockImplementationOnce(() => secondUserValues.promise)

    const wrapper = mount(UserAttributeForm, {
      props: {
        userId: 1,
        modelValue: {}
      }
    })

    await flushPromises()
    expect((wrapper.find('input').element as HTMLInputElement).value).toBe('alpha')

    await wrapper.setProps({ userId: 2 })
    await flushPromises()
    expect((wrapper.find('input').element as HTMLInputElement).value).toBe('')
    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toEqual({})

    secondUserValues.resolve([{ attribute_id: 1, value: 'beta' }])
    await flushPromises()
    expect((wrapper.find('input').element as HTMLInputElement).value).toBe('beta')
  })

  it('ignores stale user attribute responses that finish after a newer user load', async () => {
    const firstUserValues = createDeferred<Array<{ attribute_id: number; value: string }>>()
    const secondUserValues = createDeferred<Array<{ attribute_id: number; value: string }>>()
    mockGetUserAttributeValues
      .mockImplementationOnce(() => firstUserValues.promise)
      .mockImplementationOnce(() => secondUserValues.promise)

    const wrapper = mount(UserAttributeForm, {
      props: {
        userId: 1,
        modelValue: {}
      }
    })

    await flushPromises()
    await wrapper.setProps({ userId: 2 })

    secondUserValues.resolve([{ attribute_id: 1, value: 'beta' }])
    await flushPromises()
    expect((wrapper.find('input').element as HTMLInputElement).value).toBe('beta')

    firstUserValues.resolve([{ attribute_id: 1, value: 'alpha' }])
    await flushPromises()
    expect((wrapper.find('input').element as HTMLInputElement).value).toBe('beta')
    expect((wrapper.find('input').element as HTMLInputElement).value).not.toBe('alpha')
  })
})
