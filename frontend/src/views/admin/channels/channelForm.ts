import type { Channel, ChannelModelPricing } from '@/api/admin/channels'
import type { AdminGroup, GroupPlatform } from '@/types'
import type { PricingFormEntry } from '@/components/admin/channel/types'
import {
  apiIntervalsToForm,
  findModelConflict,
  formIntervalsToAPI,
  mTokToPerToken,
  perTokenToMTok,
  validateIntervals
} from '@/components/admin/channel/types'

type TranslateFn = (key: string, params?: unknown) => string

export interface PlatformSection {
  platform: GroupPlatform
  enabled: boolean
  collapsed: boolean
  group_ids: number[]
  model_mapping: Record<string, string>
  model_pricing: PricingFormEntry[]
}

export interface ChannelFormState {
  name: string
  description: string
  status: string
  restrict_models: boolean
  billing_model_source: string
  platforms: PlatformSection[]
}

export interface ChannelValidationFailure {
  message: string
  activeTab?: string
}

export function createDefaultChannelFormState(): ChannelFormState {
  return {
    name: '',
    description: '',
    status: 'active',
    restrict_models: false,
    billing_model_source: 'channel_mapped',
    platforms: []
  }
}

export function resetChannelForm(form: ChannelFormState) {
  const next = createDefaultChannelFormState()
  form.name = next.name
  form.description = next.description
  form.status = next.status
  form.restrict_models = next.restrict_models
  form.billing_model_source = next.billing_model_source
  form.platforms = next.platforms
}

export function getActiveChannelPlatforms(sections: PlatformSection[]): GroupPlatform[] {
  return sections.filter(section => section.enabled).map(section => section.platform)
}

export function toggleChannelPlatform(
  sections: PlatformSection[],
  platform: GroupPlatform,
  activeTab: string
): string {
  const section = sections.find(item => item.platform === platform)
  if (section) {
    section.enabled = !section.enabled
    if (!section.enabled && activeTab === platform) {
      return 'basic'
    }
    return activeTab
  }

  sections.push({
    platform,
    enabled: true,
    collapsed: false,
    group_ids: [],
    model_mapping: {},
    model_pricing: []
  })
  return activeTab
}

export function buildChannelGroupConflictMap(channels: Channel[], editingChannelID?: number | null) {
  const map = new Map<number, Channel>()
  for (const channel of channels) {
    if (editingChannelID && channel.id === editingChannelID) {
      continue
    }
    for (const groupID of channel.group_ids || []) {
      map.set(groupID, channel)
    }
  }
  return map
}

export function toggleChannelGroupInSection(
  sections: PlatformSection[],
  sectionIdx: number,
  groupID: number
) {
  const section = sections[sectionIdx]
  const index = section.group_ids.indexOf(groupID)
  if (index >= 0) {
    section.group_ids.splice(index, 1)
    return
  }
  section.group_ids.push(groupID)
}

export function addChannelPricingEntry(sections: PlatformSection[], sectionIdx: number) {
  sections[sectionIdx].model_pricing.push({
    models: [],
    billing_mode: 'token',
    input_price: null,
    output_price: null,
    cache_write_price: null,
    cache_read_price: null,
    image_output_price: null,
    per_request_price: null,
    intervals: []
  })
}

export function updateChannelPricingEntry(
  sections: PlatformSection[],
  sectionIdx: number,
  entryIdx: number,
  updated: PricingFormEntry
) {
  sections[sectionIdx].model_pricing.splice(entryIdx, 1, updated)
}

export function removeChannelPricingEntry(
  sections: PlatformSection[],
  sectionIdx: number,
  entryIdx: number
) {
  sections[sectionIdx].model_pricing.splice(entryIdx, 1)
}

export function addChannelMappingEntry(sections: PlatformSection[], sectionIdx: number) {
  const mapping = sections[sectionIdx].model_mapping
  let key = ''
  let index = 1
  while (key === '' || key in mapping) {
    key = `model-${index}`
    index++
  }
  mapping[key] = ''
}

export function removeChannelMappingEntry(
  sections: PlatformSection[],
  sectionIdx: number,
  key: string
) {
  delete sections[sectionIdx].model_mapping[key]
}

export function renameChannelMappingKey(
  sections: PlatformSection[],
  sectionIdx: number,
  oldKey: string,
  newKey: string
) {
  const trimmed = newKey.trim()
  if (!trimmed || trimmed === oldKey) {
    return
  }

  const mapping = sections[sectionIdx].model_mapping
  if (trimmed in mapping) {
    return
  }

  const value = mapping[oldKey]
  delete mapping[oldKey]
  mapping[trimmed] = value
}

export function serializeChannelForm(form: ChannelFormState): {
  group_ids: number[]
  model_pricing: ChannelModelPricing[]
  model_mapping: Record<string, Record<string, string>>
} {
  const group_ids: number[] = []
  const model_pricing: ChannelModelPricing[] = []
  const model_mapping: Record<string, Record<string, string>> = {}

  for (const section of form.platforms) {
    if (!section.enabled) {
      continue
    }

    group_ids.push(...section.group_ids)

    if (Object.keys(section.model_mapping).length > 0) {
      model_mapping[section.platform] = { ...section.model_mapping }
    }

    for (const entry of section.model_pricing) {
      if (entry.models.length === 0) {
        continue
      }
      model_pricing.push({
        platform: section.platform,
        models: entry.models,
        billing_mode: entry.billing_mode,
        input_price: mTokToPerToken(entry.input_price),
        output_price: mTokToPerToken(entry.output_price),
        cache_write_price: mTokToPerToken(entry.cache_write_price),
        cache_read_price: mTokToPerToken(entry.cache_read_price),
        image_output_price: mTokToPerToken(entry.image_output_price),
        per_request_price:
          entry.per_request_price != null && entry.per_request_price !== ''
            ? Number(entry.per_request_price)
            : null,
        intervals: formIntervalsToAPI(entry.intervals || [])
      })
    }
  }

  return { group_ids, model_pricing, model_mapping }
}

export function buildChannelSectionsFromAPI(
  channel: Channel,
  allGroups: AdminGroup[],
  platformOrder: GroupPlatform[]
): PlatformSection[] {
  const groupPlatformMap = new Map<number, GroupPlatform>()
  for (const group of allGroups) {
    groupPlatformMap.set(group.id, group.platform)
  }

  const active = new Set<GroupPlatform>()
  for (const groupID of channel.group_ids || []) {
    const platform = groupPlatformMap.get(groupID)
    if (platform) {
      active.add(platform)
    }
  }
  for (const pricing of channel.model_pricing || []) {
    if (pricing.platform) {
      active.add(pricing.platform as GroupPlatform)
    }
  }
  for (const platform of Object.keys(channel.model_mapping || {})) {
    if (platformOrder.includes(platform as GroupPlatform)) {
      active.add(platform as GroupPlatform)
    }
  }

  const sections: PlatformSection[] = []
  for (const platform of platformOrder) {
    if (!active.has(platform)) {
      continue
    }

    const groupIDs = (channel.group_ids || []).filter(groupID => groupPlatformMap.get(groupID) === platform)
    const mapping = (channel.model_mapping || {})[platform] || {}
    const pricing = (channel.model_pricing || [])
      .filter(item => (item.platform || 'anthropic') === platform)
      .map(item => ({
        models: item.models || [],
        billing_mode: item.billing_mode,
        input_price: perTokenToMTok(item.input_price),
        output_price: perTokenToMTok(item.output_price),
        cache_write_price: perTokenToMTok(item.cache_write_price),
        cache_read_price: perTokenToMTok(item.cache_read_price),
        image_output_price: perTokenToMTok(item.image_output_price),
        per_request_price: item.per_request_price,
        intervals: apiIntervalsToForm(item.intervals || [])
      } as PricingFormEntry))

    sections.push({
      platform,
      enabled: true,
      collapsed: false,
      group_ids: groupIDs,
      model_mapping: { ...mapping },
      model_pricing: pricing
    })
  }

  return sections
}

export function validateChannelForm(
  form: ChannelFormState,
  t: TranslateFn
): ChannelValidationFailure | null {
  if (!form.name.trim()) {
    return { message: t('admin.channels.nameRequired'), activeTab: 'basic' }
  }

  const enabledSections = form.platforms.filter(section => section.enabled)

  for (const section of enabledSections) {
    if (section.group_ids.length === 0) {
      return {
        message: t('admin.channels.noGroupsSelected', {
          platform: t(`admin.groups.platforms.${section.platform}`, section.platform)
        }),
        activeTab: section.platform
      }
    }

    for (const entry of section.model_pricing) {
      if (entry.models.length === 0) {
        return {
          message: t('admin.channels.emptyModelsInPricing', {
            platform: t(`admin.groups.platforms.${section.platform}`, section.platform)
          }),
          activeTab: section.platform
        }
      }
    }
  }

  for (const section of enabledSections) {
    const allModels: string[] = []
    for (const entry of section.model_pricing) {
      allModels.push(...entry.models)
    }
    const pricingConflict = findModelConflict(allModels)
    if (pricingConflict) {
      return {
        message: t('admin.channels.modelConflict', {
          model1: pricingConflict[0],
          model2: pricingConflict[1]
        }),
        activeTab: section.platform
      }
    }

    const mappingKeys = Object.keys(section.model_mapping)
    if (mappingKeys.length === 0) {
      continue
    }
    const mappingConflict = findModelConflict(mappingKeys)
    if (mappingConflict) {
      return {
        message: t('admin.channels.mappingConflict', {
          model1: mappingConflict[0],
          model2: mappingConflict[1]
        }),
        activeTab: section.platform
      }
    }
  }

  for (const section of enabledSections) {
    for (const entry of section.model_pricing) {
      if (entry.models.length === 0) {
        continue
      }
      if (
        (entry.billing_mode === 'per_request' || entry.billing_mode === 'image') &&
        (entry.per_request_price == null || entry.per_request_price === '') &&
        (!entry.intervals || entry.intervals.length === 0)
      ) {
        return {
          message: t('admin.channels.form.perRequestPriceRequired'),
          activeTab: section.platform
        }
      }
    }
  }

  for (const section of enabledSections) {
    for (const entry of section.model_pricing) {
      if (!entry.intervals || entry.intervals.length === 0) {
        continue
      }
      const intervalError = validateIntervals(entry.intervals)
      if (!intervalError) {
        continue
      }
      return {
        message: `${t(`admin.groups.platforms.${section.platform}`, section.platform)} - ${entry.models.join(', ') || '未命名'}: ${intervalError}`,
        activeTab: section.platform
      }
    }
  }

  return null
}
