import { beforeEach, describe, expect, it, vi } from 'vitest'

const { getModelCatalog } = vi.hoisted(() => ({
  getModelCatalog: vi.fn()
}))

vi.mock('@/api/admin/accounts', () => ({
  getAntigravityDefaultModelMapping: vi.fn()
}))

vi.mock('@/api/admin/modelCatalog', () => ({
  getModelCatalog
}))

async function loadModelWhitelistModule() {
  return import('../useModelWhitelist')
}

describe('useModelWhitelist Grok catalog contract', () => {
  beforeEach(() => {
    vi.resetModules()
    getModelCatalog.mockReset()
  })

  it('does not fetch the Grok catalog from getters alone', async () => {
    const modelWhitelist = await loadModelWhitelistModule()

    expect(modelWhitelist.getModelsByPlatform('grok')).toEqual([])
    expect(modelWhitelist.getModelOptionsByPlatform('grok')).toEqual([])
    expect(modelWhitelist.getPresetMappingsByPlatform('grok')).toEqual([])
    expect(getModelCatalog).not.toHaveBeenCalled()
  })

  it('retries Grok catalog loading after a failed request', async () => {
    const modelWhitelist = await loadModelWhitelistModule()

    getModelCatalog
      .mockRejectedValueOnce(new Error('catalog unavailable'))
      .mockResolvedValueOnce({
        platform: 'grok',
        models: [
          {
            id: 'grok-3',
            display_name: 'Grok 3',
            capability: 'chat',
            protocol_family: 'responses',
            required_tier: 'basic',
            aliases: ['grok-4.20-auto'],
            supports_stream: true,
            supports_tools: true
          }
        ]
      })

    await expect(modelWhitelist.ensureModelCatalogLoaded('grok')).resolves.toEqual([])
    await expect(modelWhitelist.ensureModelCatalogLoaded('grok')).resolves.toEqual([
      expect.objectContaining({
        id: 'grok-3',
        display_name: 'Grok 3'
      })
    ])

    expect(modelWhitelist.getModelsByPlatform('grok')).toEqual(['grok-3'])
    expect(getModelCatalog).toHaveBeenCalledTimes(2)
  })
})
