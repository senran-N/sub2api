import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import type { Group, RedeemCode } from '@/types'
import { useRedeemGeneration } from '../useRedeemGeneration'

const { getAllGroups, generateCodes } = vi.hoisted(() => ({
  getAllGroups: vi.fn(),
  generateCodes: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      getAll: getAllGroups
    },
    redeem: {
      generate: generateCodes
    }
  }
}))

function createGroup(overrides: Partial<Group> = {}): Group {
  return {
    id: 1,
    name: 'Pro',
    description: 'subscription plan',
    platform: 'openai',
    rate_multiplier: 1.5,
    status: 'active',
    subscription_type: 'subscription',
    ...overrides
  } as Group
}

function createCode(code: string): RedeemCode {
  return {
    id: code.length,
    code,
    type: 'subscription',
    value: 30,
    status: 'unused',
    used_by: null,
    used_at: null,
    created_at: '2026-04-04T00:00:00Z'
  }
}

function createComposable() {
  const showError = vi.fn()
  const copyToClipboard = vi.fn().mockResolvedValue(true)
  const reloadCodes = vi.fn().mockResolvedValue(undefined)
  const composable = useRedeemGeneration({
    t: (key: string) => key,
    showError,
    copyToClipboard,
    reloadCodes
  })

  return {
    composable,
    showError,
    copyToClipboard,
    reloadCodes
  }
}

describe('useRedeemGeneration', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-04-04T00:00:00Z'))

    getAllGroups.mockReset()
    generateCodes.mockReset()

    getAllGroups.mockResolvedValue([
      createGroup(),
      createGroup({ id: 2, name: 'Standard', subscription_type: 'standard' })
    ])
    generateCodes.mockResolvedValue([createCode('AAA'), createCode('BBB')])
  })

  it('loads subscription groups and keeps invitation values normalized', async () => {
    const setup = createComposable()

    await setup.composable.loadSubscriptionGroups()
    expect(setup.composable.subscriptionGroupOptions.value).toEqual([
      {
        value: 1,
        label: 'Pro',
        description: 'subscription plan',
        platform: 'openai',
        subscriptionType: 'subscription',
        rate: 1.5
      }
    ])

    setup.composable.generateForm.type = 'invitation'
    await nextTick()
    expect(setup.composable.generateForm.value).toBe(0)

    setup.composable.generateForm.type = 'balance'
    await nextTick()
    expect(setup.composable.generateForm.value).toBe(10)
  })

  it('validates subscription generation and resets result dialog state', async () => {
    const setup = createComposable()

    setup.composable.generateForm.type = 'subscription'
    setup.composable.generateForm.group_id = null
    await setup.composable.handleGenerateCodes()
    expect(setup.showError).toHaveBeenCalledWith('admin.redeem.groupRequired')
    expect(generateCodes).not.toHaveBeenCalled()

    setup.composable.showGenerateDialog.value = true
    setup.composable.generateForm.group_id = 7
    setup.composable.generateForm.validity_days = 90
    await setup.composable.handleGenerateCodes()

    expect(generateCodes).toHaveBeenCalledWith(1, 'subscription', 10, 7, 90)
    expect(setup.composable.showGenerateDialog.value).toBe(false)
    expect(setup.composable.showResultDialog.value).toBe(true)
    expect(setup.composable.generatedCodesText.value).toBe('AAA\nBBB')
    expect(setup.composable.textareaHeight.value).toBe('72px')
    expect(setup.composable.generateForm.group_id).toBeNull()
    expect(setup.composable.generateForm.validity_days).toBe(30)
    expect(setup.reloadCodes).toHaveBeenCalledTimes(1)

    setup.composable.closeResultDialog()
    expect(setup.composable.showResultDialog.value).toBe(false)
    expect(setup.composable.generatedCodes.value).toEqual([])
  })

  it('copies and downloads generated codes', async () => {
    const setup = createComposable()
    const originalCreateObjectURL = window.URL.createObjectURL
    const originalRevokeObjectURL = window.URL.revokeObjectURL
    const originalCreateElement = document.createElement.bind(document)
    const clickSpy = vi.fn()

    setup.composable.generatedCodes.value = [createCode('AAA'), createCode('BBB')]
    window.URL.createObjectURL = vi.fn(() => 'blob:redeem-generated') as typeof window.URL.createObjectURL
    window.URL.revokeObjectURL = vi.fn() as typeof window.URL.revokeObjectURL
    document.createElement = vi.fn((tagName: string) => {
      if (tagName === 'a') {
        const link = originalCreateElement('a')
        link.click = clickSpy
        return link
      }
      return originalCreateElement(tagName)
    }) as typeof document.createElement

    await setup.composable.copyGeneratedCodes()
    expect(setup.copyToClipboard).toHaveBeenCalledWith('AAA\nBBB', 'admin.redeem.copied')
    expect(setup.composable.copiedAll.value).toBe(true)
    await vi.advanceTimersByTimeAsync(2000)
    expect(setup.composable.copiedAll.value).toBe(false)

    setup.composable.downloadGeneratedCodes()
    expect(clickSpy).toHaveBeenCalledTimes(1)

    document.createElement = originalCreateElement
    window.URL.createObjectURL = originalCreateObjectURL
    window.URL.revokeObjectURL = originalRevokeObjectURL
  })
})
