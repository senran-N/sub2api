import { describe, expect, it } from 'vitest'
import type { Channel } from '@/api/admin/channels'
import type { AdminGroup } from '@/types'
import {
  addChannelMappingEntry,
  buildChannelGroupConflictMap,
  buildChannelSectionsFromAPI,
  createDefaultChannelFormState,
  renameChannelMappingKey,
  serializeChannelForm,
  validateChannelForm
} from '../channelForm'
import { platformOrder } from '../viewHelpers'

const t = (key: string, params?: Record<string, unknown>) =>
  params ? `${key}:${JSON.stringify(params)}` : key

function createChannel(overrides: Partial<Channel> = {}): Channel {
  return {
    id: 1,
    name: 'Main channel',
    description: '',
    status: 'active',
    billing_model_source: 'channel_mapped',
    restrict_models: false,
    group_ids: [11],
    model_pricing: [],
    model_mapping: {},
    created_at: '',
    updated_at: '',
    ...overrides
  }
}

function createGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 11,
    name: 'OpenAI Pro',
    description: '',
    status: 'active',
    platform: 'openai',
    channel_id: 0,
    sort_order: 0,
    account_count: 1,
    rate_multiplier: 1,
    max_tokens: 0,
    user_id: 0,
    models: [],
    created_at: '',
    updated_at: '',
    ...overrides
  }
}

describe('channelForm helpers', () => {
  it('serializes enabled platform sections only', () => {
    const form = createDefaultChannelFormState()
    form.name = 'Main'
    form.platforms = [
      {
        platform: 'openai',
        enabled: true,
        collapsed: false,
        group_ids: [11],
        model_mapping: { 'gpt-4.1': 'gpt-4.1-mini' },
        model_pricing: [{
          models: ['gpt-4.1'],
          billing_mode: 'token',
          input_price: 1500000,
          output_price: 3000000,
          cache_write_price: null,
          cache_read_price: null,
          image_output_price: null,
          per_request_price: null,
          intervals: []
        }]
      },
      {
        platform: 'gemini',
        enabled: false,
        collapsed: false,
        group_ids: [22],
        model_mapping: {},
        model_pricing: []
      }
    ]

    expect(serializeChannelForm(form)).toEqual({
      group_ids: [11],
      model_mapping: {
        openai: { 'gpt-4.1': 'gpt-4.1-mini' }
      },
      model_pricing: [{
        platform: 'openai',
        models: ['gpt-4.1'],
        billing_mode: 'token',
        input_price: 1.5,
        output_price: 3,
        cache_write_price: null,
        cache_read_price: null,
        image_output_price: null,
        per_request_price: null,
        intervals: []
      }]
    })
  })

  it('hydrates channel sections from grouped API data', () => {
    const sections = buildChannelSectionsFromAPI(
      createChannel({
        group_ids: [11, 12],
        model_mapping: {
          openai: { 'gpt-4.1': 'gpt-4.1-mini' }
        },
        model_pricing: [{
          platform: 'openai',
          models: ['gpt-4.1'],
          billing_mode: 'token',
          input_price: 1.2,
          output_price: 2.4,
          cache_write_price: null,
          cache_read_price: null,
          image_output_price: null,
          per_request_price: null,
          intervals: []
        }]
      }),
      [
        createGroup({ id: 11, platform: 'openai' }),
        createGroup({ id: 12, platform: 'gemini', name: 'Gemini Pro' })
      ],
      platformOrder
    )

    expect(sections.map(section => section.platform)).toEqual(['openai', 'gemini'])
    expect(sections[0].group_ids).toEqual([11])
    expect(sections[0].model_mapping).toEqual({ 'gpt-4.1': 'gpt-4.1-mini' })
    expect(sections[0].model_pricing[0].input_price).toBe(1200000)
    expect(sections[1].group_ids).toEqual([12])
  })

  it('validates channel conflicts and mapping key helpers', () => {
    const form = createDefaultChannelFormState()
    form.name = 'Conflict channel'
    form.platforms = [{
      platform: 'openai',
      enabled: true,
      collapsed: false,
      group_ids: [11],
      model_mapping: {},
      model_pricing: []
    }]

    addChannelMappingEntry(form.platforms, 0)
    renameChannelMappingKey(form.platforms, 0, 'model-1', 'gpt-*')
    form.platforms[0].model_mapping['gpt-4.1'] = 'gpt-4.1-mini'

    expect(validateChannelForm(form, t)).toEqual({
      message: 'admin.channels.mappingConflict:{"model1":"gpt-*","model2":"gpt-4.1"}',
      activeTab: 'openai'
    })
  })

  it('builds group conflict map excluding current edit target', () => {
    const conflictMap = buildChannelGroupConflictMap(
      [
        createChannel({ id: 1, name: 'Current', group_ids: [11] }),
        createChannel({ id: 2, name: 'Other', group_ids: [22] })
      ],
      1
    )

    expect(conflictMap.has(11)).toBe(false)
    expect(conflictMap.get(22)?.name).toBe('Other')
  })
})
