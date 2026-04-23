import { describe, expect, it, vi } from 'vitest'

vi.mock('@/api/admin/accounts', () => ({
  getAntigravityDefaultModelMapping: vi.fn()
}))

vi.mock('@/api/admin/modelCatalog', () => ({
  getModelCatalog: vi.fn()
}))

import { getModelCatalog } from '@/api/admin/modelCatalog'
import {
  buildModelMappingObject,
  ensureModelCatalogLoaded,
  getModelOptionsByPlatform,
  getModelsByPlatform,
  getPresetMappingsByPlatform
} from '../useModelWhitelist'

describe('useModelWhitelist', () => {
  it('openai 模型列表包含 GPT-5.5 与 GPT-5.4 官方快照', () => {
    const models = getModelsByPlatform('openai')

    expect(models).toContain('gpt-5.5')
    expect(models).toContain('gpt-5.4')
    expect(models).toContain('gpt-5.4-mini')
    expect(models).toContain('gpt-5.4-2026-03-05')
  })

  it('openai 预设映射包含 GPT-5.5 透传项', () => {
    expect(getPresetMappingsByPlatform('openai')).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ from: 'gpt-5.5', to: 'gpt-5.5' })
      ])
    )
  })

  it('antigravity 模型列表包含图片模型兼容项', () => {
    const models = getModelsByPlatform('antigravity')

    expect(models).toContain('claude-opus-4-7')
    expect(models).toContain('gemini-2.5-flash-image')
    expect(models).toContain('gemini-3.1-flash-image')
    expect(models).toContain('gemini-3-pro-image')
  })

  it('Claude 与 Bedrock 预设包含 Opus 4.7', () => {
    expect(getPresetMappingsByPlatform('anthropic')).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ from: 'claude-opus-4-7', to: 'claude-opus-4-7' })
      ])
    )
    expect(getPresetMappingsByPlatform('antigravity')).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ from: 'claude-opus-4-7', to: 'claude-opus-4-7' })
      ])
    )
    expect(getPresetMappingsByPlatform('bedrock')).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ from: 'claude-opus-4-7', to: 'us.anthropic.claude-opus-4-7-v1' })
      ])
    )
  })

  it('gemini 模型列表包含原生生图模型', () => {
    const models = getModelsByPlatform('gemini')

    expect(models).toContain('gemini-2.5-flash-image')
    expect(models).toContain('gemini-3.1-flash-image')
    expect(models.indexOf('gemini-3.1-flash-image')).toBeLessThan(models.indexOf('gemini-2.0-flash'))
    expect(models.indexOf('gemini-2.5-flash-image')).toBeLessThan(models.indexOf('gemini-2.5-flash'))
  })

  it('antigravity 模型列表会把新的 Gemini 图片模型排在前面', () => {
    const models = getModelsByPlatform('antigravity')

    expect(models.indexOf('gemini-3.1-flash-image')).toBeLessThan(models.indexOf('gemini-2.5-flash'))
    expect(models.indexOf('gemini-2.5-flash-image')).toBeLessThan(models.indexOf('gemini-2.5-flash-lite'))
  })

  it('whitelist 模式会忽略通配符条目', () => {
    const mapping = buildModelMappingObject('whitelist', ['claude-*', 'gemini-3.1-flash-image'], [])
    expect(mapping).toEqual({
      'gemini-3.1-flash-image': 'gemini-3.1-flash-image'
    })
  })

  it('whitelist 模式会保留 GPT-5.4 官方快照的精确映射', () => {
    const mapping = buildModelMappingObject('whitelist', ['gpt-5.4-2026-03-05'], [])

    expect(mapping).toEqual({
      'gpt-5.4-2026-03-05': 'gpt-5.4-2026-03-05'
    })
  })

  it('whitelist keeps GPT-5.4 mini exact mapping', () => {
    const mapping = buildModelMappingObject('whitelist', ['gpt-5.4-mini'], [])

    expect(mapping).toEqual({
      'gpt-5.4-mini': 'gpt-5.4-mini'
    })
  })

  it('grok 模型列表和预设映射来自后端 catalog', async () => {
    vi.mocked(getModelCatalog).mockResolvedValueOnce({
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
        },
        {
          id: 'grok-imagine-video',
          display_name: 'Grok Imagine Video',
          capability: 'video',
          protocol_family: 'media_job',
          required_tier: 'super',
          aliases: [],
          supports_stream: false,
          supports_tools: false
        }
      ]
    })

    await ensureModelCatalogLoaded('grok')

    expect(getModelsByPlatform('grok')).toEqual(['grok-3', 'grok-imagine-video'])
    expect(getModelOptionsByPlatform('grok')).toEqual([
      { value: 'grok-3', label: 'Grok 3 (grok-3)' },
      { value: 'grok-imagine-video', label: 'Grok Imagine Video (grok-imagine-video)' }
    ])
    expect(getPresetMappingsByPlatform('grok')).toEqual([
      { label: 'Grok 3', from: 'grok-3', to: 'grok-3', tone: 'info' },
      { label: 'Grok Imagine Video', from: 'grok-imagine-video', to: 'grok-imagine-video', tone: 'brand-rose' }
    ])
  })
})
