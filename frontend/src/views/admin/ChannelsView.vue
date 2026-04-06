<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-64">
              <Icon
                name="search"
                size="md"
                class="channel-view__search-icon absolute left-3 top-1/2 -translate-y-1/2"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.channels.searchChannels')"
                class="input pl-10"
                @input="handleSearch"
              />
            </div>

            <Select
              v-model="filters.status"
              :options="statusFilterOptions"
              :placeholder="t('admin.channels.allStatus')"
              class="w-40"
              @change="loadChannels"
            />
          </div>

          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
              @click="loadChannels"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button class="btn btn-primary" @click="openCreateDialog">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('admin.channels.createChannel') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="channels" :loading="loading">
          <template #cell-name="{ value }">
            <span class="channel-view__text-strong font-medium">{{ value }}</span>
          </template>

          <template #cell-description="{ value }">
            <span class="channel-view__text-muted text-sm">{{ value || '-' }}</span>
          </template>

          <template #cell-status="{ row }">
            <Toggle :modelValue="row.status === 'active'" @update:modelValue="toggleChannelStatus(row)" />
          </template>

          <template #cell-group_count="{ row }">
            <span class="theme-chip theme-chip--regular theme-chip--neutral inline-flex items-center">
              {{ (row.group_ids || []).length }}
              {{ t('admin.channels.groupsUnit') }}
            </span>
          </template>

          <template #cell-pricing_count="{ row }">
            <span class="theme-chip theme-chip--regular theme-chip--neutral inline-flex items-center">
              {{ (row.model_pricing || []).length }}
              {{ t('admin.channels.pricingUnit') }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="channel-view__text-muted text-sm">
              {{ formatDate(value) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                :class="getActionButtonClasses('info')"
                @click="openEditDialog(row)"
              >
                <Icon name="edit" size="sm" />
                <span class="text-xs">{{ t('common.edit') }}</span>
              </button>
              <button
                :class="getActionButtonClasses('danger')"
                @click="handleDelete(row)"
              >
                <Icon name="trash" size="sm" />
                <span class="text-xs">{{ t('common.delete') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.channels.noChannelsYet')"
              :description="t('admin.channels.createFirstChannel')"
              :action-text="t('admin.channels.createChannel')"
              @action="openCreateDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="showDialog"
      :title="editingChannel ? t('admin.channels.editChannel') : t('admin.channels.createChannel')"
      width="extra-wide"
      @close="closeDialog"
    >
      <div class="channel-dialog-body">
        <div class="channel-view__tabs flex flex-shrink-0 items-center">
          <button
            type="button"
            class="channel-tab"
            :class="activeTab === 'basic' ? 'channel-tab-active' : 'channel-tab-inactive'"
            @click="activeTab = 'basic'"
          >
            {{ t('admin.channels.form.basicSettings') }}
          </button>
          <button
            v-for="section in form.platforms.filter(s => s.enabled)"
            :key="section.platform"
            type="button"
            class="channel-tab group"
            :class="activeTab === section.platform ? 'channel-tab-active' : 'channel-tab-inactive'"
            @click="activeTab = section.platform"
          >
            <PlatformIcon :platform="section.platform" size="xs" :class="getPlatformTextClass(section.platform)" />
            <span :class="getPlatformTextClass(section.platform)">{{ t(`admin.groups.platforms.${section.platform}`, section.platform) }}</span>
          </button>
        </div>

        <form id="channel-form" class="flex-1 overflow-y-auto pt-4" @submit.prevent="handleSubmit">
          <div v-show="activeTab === 'basic'" class="space-y-5">
            <div>
              <label class="input-label">{{ t('admin.channels.form.name') }} <span class="channel-view__required">*</span></label>
              <input
                v-model="form.name"
                type="text"
                required
                class="input"
                :placeholder="t('admin.channels.form.namePlaceholder')"
              />
            </div>

            <div>
              <label class="input-label">{{ t('admin.channels.form.description') }}</label>
              <textarea
                v-model="form.description"
                rows="2"
                class="input"
                :placeholder="t('admin.channels.form.descriptionPlaceholder')"
              />
            </div>

            <div v-if="editingChannel">
              <label class="input-label">{{ t('admin.channels.form.status') }}</label>
              <Select v-model="form.status" :options="statusEditOptions" />
            </div>

            <div>
              <label class="channel-view__checkbox-row flex cursor-pointer items-center gap-2">
                <input
                  v-model="form.restrict_models"
                  type="checkbox"
                  class="channel-view__checkbox h-4 w-4 rounded"
                />
                <span class="input-label mb-0">{{ t('admin.channels.form.restrictModels') }}</span>
              </label>
              <p class="channel-view__text-soft mt-1 ml-6 text-xs">
                {{ t('admin.channels.form.restrictModelsHint') }}
              </p>
            </div>

            <div>
              <label class="input-label">{{ t('admin.channels.form.billingModelSource') }}</label>
              <Select v-model="form.billing_model_source" :options="billingModelSourceOptions" />
              <p class="channel-view__text-soft mt-1 text-xs">
                {{ t('admin.channels.form.billingModelSourceHint') }}
              </p>
            </div>

            <div class="space-y-3">
              <label class="input-label mb-0">{{ t('admin.channels.form.platformConfig') }}</label>
              <div class="flex flex-wrap gap-2">
                <label
                  v-for="platform in platformOrder"
                  :key="platform"
                  :class="getPlatformToggleClasses(platform, activePlatforms.includes(platform))"
                >
                  <input
                    type="checkbox"
                    :checked="activePlatforms.includes(platform)"
                    class="channel-view__checkbox h-3.5 w-3.5 rounded"
                    @change="togglePlatform(platform)"
                  />
                  <PlatformIcon :platform="platform" size="xs" :class="getPlatformTextClass(platform)" />
                  <span :class="getPlatformTextClass(platform)">{{ t(`admin.groups.platforms.${platform}`, platform) }}</span>
                </label>
              </div>
            </div>
          </div>

          <div
            v-for="(section, sIdx) in form.platforms"
            :key="`tab-${section.platform}`"
            v-show="section.enabled && activeTab === section.platform"
            class="space-y-4"
          >
            <div>
              <label class="input-label text-xs">
                {{ t('admin.channels.form.groups') }} <span class="channel-view__required">*</span>
                <span v-if="section.group_ids.length > 0" class="channel-view__text-soft ml-1 font-normal">
                  ({{ t('admin.channels.form.selectedCount', { count: section.group_ids.length }) }})
                </span>
              </label>
              <div class="channel-view__group-list">
                <div v-if="groupsLoading" class="channel-view__group-state channel-view__text-muted text-center text-xs">
                  {{ t('common.loading') }}
                </div>
                <div v-else-if="getGroupsForPlatform(section.platform).length === 0" class="channel-view__group-state channel-view__text-muted text-center text-xs">
                  {{ t('admin.channels.form.noGroupsAvailable') }}
                </div>
                <div v-else class="flex flex-wrap gap-1">
                  <label
                    v-for="group in getGroupsForPlatform(section.platform)"
                    :key="group.id"
                    :class="getGroupChipClasses(section.platform, section.group_ids.includes(group.id), isGroupInOtherChannel(group.id, section.platform))"
                  >
                    <input
                      type="checkbox"
                      :checked="section.group_ids.includes(group.id)"
                      :disabled="isGroupInOtherChannel(group.id, section.platform)"
                      class="channel-view__checkbox h-3 w-3 rounded"
                      @change="toggleGroupInSection(sIdx, group.id)"
                    />
                    <span :class="['font-medium', getPlatformTextClass(group.platform)]">{{ group.name }}</span>
                    <span :class="['channel-view__rate-badge text-[10px]', getRateBadgeClass(group.platform)]">{{ group.rate_multiplier }}x</span>
                    <span class="channel-view__text-soft text-[10px]">{{ group.account_count || 0 }}</span>
                    <span v-if="isGroupInOtherChannel(group.id, section.platform)" class="channel-view__text-soft text-[10px]">
                      {{ getGroupInOtherChannelLabel(group.id) }}
                    </span>
                  </label>
                </div>
              </div>
            </div>

            <div>
              <div class="mb-1 flex items-center justify-between">
                <label class="input-label mb-0 text-xs">{{ t('admin.channels.form.modelMapping') }}</label>
                <button type="button" class="channel-view__text-button text-xs" @click="addMappingEntry(sIdx)">
                  + {{ t('common.add') }}
                </button>
              </div>
              <div
                v-if="Object.keys(section.model_mapping).length === 0"
                class="channel-view__empty-box text-center text-xs"
              >
                {{ t('admin.channels.form.noMappingRules') }}
              </div>
              <div v-else class="space-y-1">
                <div v-for="(_, srcModel) in section.model_mapping" :key="srcModel" class="flex items-center gap-2">
                  <input
                    :value="srcModel"
                    type="text"
                    class="input flex-1 text-xs"
                    :class="getPlatformTextClass(section.platform)"
                    :placeholder="t('admin.channels.form.mappingSource')"
                    @change="renameMappingKey(sIdx, srcModel, ($event.target as HTMLInputElement).value)"
                  />
                  <span class="channel-view__text-soft text-xs">→</span>
                  <input
                    :value="section.model_mapping[srcModel]"
                    type="text"
                    class="input flex-1 text-xs"
                    :class="getPlatformTextClass(section.platform)"
                    :placeholder="t('admin.channels.form.mappingTarget')"
                    @input="section.model_mapping[srcModel] = ($event.target as HTMLInputElement).value"
                  />
                  <button type="button" class="channel-view__icon-button channel-view__icon-button--danger" @click="removeMappingEntry(sIdx, srcModel)">
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </div>
            </div>

            <div>
              <div class="mb-1 flex items-center justify-between">
                <label class="input-label mb-0 text-xs">{{ t('admin.channels.form.modelPricing') }}</label>
                <button type="button" class="channel-view__text-button text-xs" @click="addPricingEntry(sIdx)">
                  + {{ t('common.add') }}
                </button>
              </div>
              <div
                v-if="section.model_pricing.length === 0"
                class="channel-view__empty-box text-center text-xs"
              >
                {{ t('admin.channels.form.noPricingRules') }}
              </div>
              <div v-else class="space-y-2">
                <PricingEntryCard
                  v-for="(entry, idx) in section.model_pricing"
                  :key="idx"
                  :entry="entry"
                  :platform="section.platform"
                  @update="updatePricingEntry(sIdx, idx, $event)"
                  @remove="removePricingEntry(sIdx, idx)"
                />
              </div>
            </div>
          </div>
        </form>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeDialog">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="channel-form" :disabled="submitting" class="btn btn-primary">
            {{ submitting
              ? t('common.submitting')
              : editingChannel
                ? t('common.update')
                : t('common.create') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.channels.deleteChannel')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { Channel, ChannelModelPricing, CreateChannelRequest, UpdateChannelRequest } from '@/api/admin/channels'
import type { AdminGroup, GroupPlatform } from '@/types'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import PricingEntryCard from '@/components/admin/channel/PricingEntryCard.vue'
import type { PricingFormEntry } from '@/components/admin/channel/types'
import {
  apiIntervalsToForm,
  findModelConflict,
  formIntervalsToAPI,
  mTokToPerToken,
  perTokenToMTok,
  validateIntervals
} from '@/components/admin/channel/types'

interface PlatformSection {
  platform: GroupPlatform
  enabled: boolean
  collapsed: boolean
  group_ids: number[]
  model_mapping: Record<string, string>
  model_pricing: PricingFormEntry[]
}

const { t } = useI18n()
const appStore = useAppStore()

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('admin.channels.columns.name'), sortable: true },
  { key: 'description', label: t('admin.channels.columns.description'), sortable: false },
  { key: 'status', label: t('admin.channels.columns.status'), sortable: true },
  { key: 'group_count', label: t('admin.channels.columns.groups'), sortable: false },
  { key: 'pricing_count', label: t('admin.channels.columns.pricing'), sortable: false },
  { key: 'created_at', label: t('admin.channels.columns.createdAt'), sortable: true },
  { key: 'actions', label: t('admin.channels.columns.actions'), sortable: false }
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('admin.channels.allStatus') },
  { value: 'active', label: t('admin.channels.statusActive') },
  { value: 'disabled', label: t('admin.channels.statusDisabled') }
])

const statusEditOptions = computed(() => [
  { value: 'active', label: t('admin.channels.statusActive') },
  { value: 'disabled', label: t('admin.channels.statusDisabled') }
])

const billingModelSourceOptions = computed(() => [
  { value: 'channel_mapped', label: t('admin.channels.form.billingModelSourceChannelMapped') },
  { value: 'requested', label: t('admin.channels.form.billingModelSourceRequested') },
  { value: 'upstream', label: t('admin.channels.form.billingModelSourceUpstream') }
])

const channels = ref<Channel[]>([])
const loading = ref(false)
const searchQuery = ref('')
const filters = reactive({ status: '' })
const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0
})

const showDialog = ref(false)
const editingChannel = ref<Channel | null>(null)
const submitting = ref(false)
const showDeleteDialog = ref(false)
const deletingChannel = ref<Channel | null>(null)
const activeTab = ref<string>('basic')

const allGroups = ref<AdminGroup[]>([])
const groupsLoading = ref(false)
const allChannelsForConflict = ref<Channel[]>([])

const form = reactive({
  name: '',
  description: '',
  status: 'active',
  restrict_models: false,
  billing_model_source: 'channel_mapped' as string,
  platforms: [] as PlatformSection[]
})

let abortController: AbortController | null = null
let searchTimeout: ReturnType<typeof setTimeout>

const platformOrder: GroupPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity']

function joinClassNames(...classNames: Array<string | false | null | undefined>): string {
  return classNames.filter(Boolean).join(' ')
}

function getPlatformTextClass(platform: string): string {
  switch (platform) {
    case 'anthropic': return 'channel-view__tone-text channel-view__tone-text--brand-orange'
    case 'openai': return 'channel-view__tone-text channel-view__tone-text--success'
    case 'gemini': return 'channel-view__tone-text channel-view__tone-text--info'
    case 'antigravity': return 'channel-view__tone-text channel-view__tone-text--brand-purple'
    case 'sora': return 'channel-view__tone-text channel-view__tone-text--brand-rose'
    default: return 'channel-view__text-muted'
  }
}

function getRateBadgeClass(platform: string): string {
  switch (platform) {
    case 'anthropic': return 'theme-chip theme-chip--compact theme-chip--brand-orange'
    case 'openai': return 'theme-chip theme-chip--compact theme-chip--success'
    case 'gemini': return 'theme-chip theme-chip--compact theme-chip--info'
    case 'antigravity': return 'theme-chip theme-chip--compact theme-chip--brand-purple'
    case 'sora': return 'theme-chip theme-chip--compact theme-chip--brand-rose'
    default: return 'theme-chip theme-chip--compact theme-chip--neutral'
  }
}

function getPlatformToggleClasses(platform: GroupPlatform, active: boolean): string {
  return joinClassNames(
    'channel-view__platform-toggle inline-flex cursor-pointer items-center gap-1.5 border text-sm transition-colors',
    active && 'channel-view__platform-toggle--active',
    getPlatformTextClass(platform)
  )
}

function getGroupChipClasses(platform: GroupPlatform, selected: boolean, disabled: boolean): string {
  return joinClassNames(
    'channel-view__group-chip inline-flex cursor-pointer items-center gap-1.5 border text-xs transition-colors',
    selected && 'channel-view__group-chip--selected',
    disabled && 'opacity-40',
    getPlatformTextClass(platform)
  )
}

function getActionButtonClasses(tone: 'info' | 'danger'): string {
  return joinClassNames(
    'channel-view__action-button flex flex-col items-center gap-0.5 transition-colors',
    tone === 'info' ? 'channel-view__action-button--info' : 'channel-view__action-button--danger'
  )
}

function formatDate(value: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleDateString()
}

const activePlatforms = computed(() => form.platforms.filter(s => s.enabled).map(s => s.platform))

function addPlatformSection(platform: GroupPlatform) {
  form.platforms.push({
    platform,
    enabled: true,
    collapsed: false,
    group_ids: [],
    model_mapping: {},
    model_pricing: []
  })
}

function togglePlatform(platform: GroupPlatform) {
  const section = form.platforms.find(s => s.platform === platform)
  if (section) {
    section.enabled = !section.enabled
    if (!section.enabled && activeTab.value === platform) {
      activeTab.value = 'basic'
    }
    return
  }
  addPlatformSection(platform)
}

function getGroupsForPlatform(platform: GroupPlatform): AdminGroup[] {
  return allGroups.value.filter(group => group.platform === platform)
}

const groupToChannelMap = computed(() => {
  const map = new Map<number, Channel>()
  for (const channel of allChannelsForConflict.value) {
    if (editingChannel.value && channel.id === editingChannel.value.id) continue
    for (const groupID of channel.group_ids || []) {
      map.set(groupID, channel)
    }
  }
  return map
})

function isGroupInOtherChannel(groupId: number, _platform: string): boolean {
  return groupToChannelMap.value.has(groupId)
}

function getGroupChannelName(groupId: number): string {
  return groupToChannelMap.value.get(groupId)?.name || ''
}

function getGroupInOtherChannelLabel(groupId: number): string {
  const name = getGroupChannelName(groupId)
  return t('admin.channels.form.inOtherChannel', { name })
}

const deleteConfirmMessage = computed(() => {
  const name = deletingChannel.value?.name || ''
  return t('admin.channels.deleteConfirm', { name })
})

function toggleGroupInSection(sectionIdx: number, groupId: number) {
  const section = form.platforms[sectionIdx]
  const idx = section.group_ids.indexOf(groupId)
  if (idx >= 0) {
    section.group_ids.splice(idx, 1)
  } else {
    section.group_ids.push(groupId)
  }
}

function addPricingEntry(sectionIdx: number) {
  form.platforms[sectionIdx].model_pricing.push({
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

function updatePricingEntry(sectionIdx: number, idx: number, updated: PricingFormEntry) {
  form.platforms[sectionIdx].model_pricing.splice(idx, 1, updated)
}

function removePricingEntry(sectionIdx: number, idx: number) {
  form.platforms[sectionIdx].model_pricing.splice(idx, 1)
}

function addMappingEntry(sectionIdx: number) {
  const mapping = form.platforms[sectionIdx].model_mapping
  let key = ''
  let i = 1
  while (key === '' || key in mapping) {
    key = `model-${i}`
    i++
  }
  mapping[key] = ''
}

function removeMappingEntry(sectionIdx: number, key: string) {
  delete form.platforms[sectionIdx].model_mapping[key]
}

function renameMappingKey(sectionIdx: number, oldKey: string, newKey: string) {
  const trimmed = newKey.trim()
  if (!trimmed || trimmed === oldKey) return
  const mapping = form.platforms[sectionIdx].model_mapping
  if (trimmed in mapping) return
  const value = mapping[oldKey]
  delete mapping[oldKey]
  mapping[trimmed] = value
}

function formToAPI(): {
  group_ids: number[]
  model_pricing: ChannelModelPricing[]
  model_mapping: Record<string, Record<string, string>>
} {
  const group_ids: number[] = []
  const model_pricing: ChannelModelPricing[] = []
  const model_mapping: Record<string, Record<string, string>> = {}

  for (const section of form.platforms) {
    if (!section.enabled) continue
    group_ids.push(...section.group_ids)

    if (Object.keys(section.model_mapping).length > 0) {
      model_mapping[section.platform] = { ...section.model_mapping }
    }

    for (const entry of section.model_pricing) {
      if (entry.models.length === 0) continue
      model_pricing.push({
        platform: section.platform,
        models: entry.models,
        billing_mode: entry.billing_mode,
        input_price: mTokToPerToken(entry.input_price),
        output_price: mTokToPerToken(entry.output_price),
        cache_write_price: mTokToPerToken(entry.cache_write_price),
        cache_read_price: mTokToPerToken(entry.cache_read_price),
        image_output_price: mTokToPerToken(entry.image_output_price),
        per_request_price: entry.per_request_price != null && entry.per_request_price !== '' ? Number(entry.per_request_price) : null,
        intervals: formIntervalsToAPI(entry.intervals || [])
      })
    }
  }

  return { group_ids, model_pricing, model_mapping }
}

function apiToForm(channel: Channel): PlatformSection[] {
  const groupPlatformMap = new Map<number, GroupPlatform>()
  for (const group of allGroups.value) {
    groupPlatformMap.set(group.id, group.platform)
  }

  const active = new Set<GroupPlatform>()
  for (const groupID of channel.group_ids || []) {
    const platform = groupPlatformMap.get(groupID)
    if (platform) active.add(platform)
  }
  for (const pricing of channel.model_pricing || []) {
    if (pricing.platform) active.add(pricing.platform as GroupPlatform)
  }
  for (const platform of Object.keys(channel.model_mapping || {})) {
    if (platformOrder.includes(platform as GroupPlatform)) {
      active.add(platform as GroupPlatform)
    }
  }

  const sections: PlatformSection[] = []
  for (const platform of platformOrder) {
    if (!active.has(platform)) continue

    const groupIds = (channel.group_ids || []).filter(groupID => groupPlatformMap.get(groupID) === platform)
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
      group_ids: groupIds,
      model_mapping: { ...mapping },
      model_pricing: pricing
    })
  }

  return sections
}

async function loadChannels() {
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  loading.value = true

  try {
    const response = await adminAPI.channels.list(
      pagination.page,
      pagination.page_size,
      {
        status: filters.status || undefined,
        search: searchQuery.value || undefined
      },
      { signal: ctrl.signal }
    )

    if (ctrl.signal.aborted || abortController !== ctrl) return
    channels.value = response.items || []
    pagination.total = response.total
  } catch (error: any) {
    if (error?.name === 'AbortError' || error?.code === 'ERR_CANCELED') return
    appStore.showError(t('admin.channels.loadError'))
    console.error('Error loading channels:', error)
  } finally {
    if (abortController === ctrl) {
      loading.value = false
      abortController = null
    }
  }
}

async function loadGroups() {
  groupsLoading.value = true
  try {
    allGroups.value = await adminAPI.groups.getAll()
  } catch (error) {
    console.error('Error loading groups:', error)
  } finally {
    groupsLoading.value = false
  }
}

async function loadAllChannelsForConflict() {
  try {
    const response = await adminAPI.channels.list(1, 1000)
    allChannelsForConflict.value = response.items || []
  } catch {
    allChannelsForConflict.value = channels.value
  }
}

function handleSearch() {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadChannels()
  }, 300)
}

function handlePageChange(page: number) {
  pagination.page = page
  loadChannels()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadChannels()
}

function resetForm() {
  form.name = ''
  form.description = ''
  form.status = 'active'
  form.restrict_models = false
  form.billing_model_source = 'channel_mapped'
  form.platforms = []
  activeTab.value = 'basic'
}

async function openCreateDialog() {
  editingChannel.value = null
  resetForm()
  await Promise.all([loadGroups(), loadAllChannelsForConflict()])
  showDialog.value = true
}

async function openEditDialog(channel: Channel) {
  editingChannel.value = channel
  form.name = channel.name
  form.description = channel.description || ''
  form.status = channel.status
  form.restrict_models = channel.restrict_models || false
  form.billing_model_source = channel.billing_model_source || 'channel_mapped'
  await Promise.all([loadGroups(), loadAllChannelsForConflict()])
  form.platforms = apiToForm(channel)
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  editingChannel.value = null
  resetForm()
}

async function handleSubmit() {
  if (submitting.value) return
  if (!form.name.trim()) {
    appStore.showError(t('admin.channels.nameRequired'))
    return
  }

  for (const section of form.platforms.filter(s => s.enabled)) {
    if (section.group_ids.length === 0) {
      const platformLabel = t(`admin.groups.platforms.${section.platform}`, section.platform)
      appStore.showError(t('admin.channels.noGroupsSelected', { platform: platformLabel }))
      activeTab.value = section.platform
      return
    }
    for (const entry of section.model_pricing) {
      if (entry.models.length === 0) {
        const platformLabel = t(`admin.groups.platforms.${section.platform}`, section.platform)
        appStore.showError(t('admin.channels.emptyModelsInPricing', { platform: platformLabel }))
        activeTab.value = section.platform
        return
      }
    }
  }

  for (const section of form.platforms.filter(s => s.enabled)) {
    const allModels: string[] = []
    for (const entry of section.model_pricing) {
      allModels.push(...entry.models)
    }
    const pricingConflict = findModelConflict(allModels)
    if (pricingConflict) {
      appStore.showError(
        t('admin.channels.modelConflict', { model1: pricingConflict[0], model2: pricingConflict[1] })
      )
      activeTab.value = section.platform
      return
    }

    const mappingKeys = Object.keys(section.model_mapping)
    if (mappingKeys.length === 0) continue
    const mappingConflict = findModelConflict(mappingKeys)
    if (mappingConflict) {
      appStore.showError(
        t('admin.channels.mappingConflict', { model1: mappingConflict[0], model2: mappingConflict[1] })
      )
      activeTab.value = section.platform
      return
    }
  }

  for (const section of form.platforms.filter(s => s.enabled)) {
    for (const entry of section.model_pricing) {
      if (entry.models.length === 0) continue
      if (
        (entry.billing_mode === 'per_request' || entry.billing_mode === 'image') &&
        (entry.per_request_price == null || entry.per_request_price === '') &&
        (!entry.intervals || entry.intervals.length === 0)
      ) {
        appStore.showError(t('admin.channels.form.perRequestPriceRequired'))
        return
      }
    }
  }

  for (const section of form.platforms.filter(s => s.enabled)) {
    for (const entry of section.model_pricing) {
      if (!entry.intervals || entry.intervals.length === 0) continue
      const intervalErr = validateIntervals(entry.intervals)
      if (!intervalErr) continue
      const platformLabel = t(`admin.groups.platforms.${section.platform}`, section.platform)
      const modelLabel = entry.models.join(', ') || '未命名'
      appStore.showError(`${platformLabel} - ${modelLabel}: ${intervalErr}`)
      activeTab.value = section.platform
      return
    }
  }

  const { group_ids, model_pricing, model_mapping } = formToAPI()

  submitting.value = true
  try {
    if (editingChannel.value) {
      const req: UpdateChannelRequest = {
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        status: form.status,
        group_ids,
        model_pricing,
        model_mapping: Object.keys(model_mapping).length > 0 ? model_mapping : {},
        billing_model_source: form.billing_model_source,
        restrict_models: form.restrict_models
      }
      await adminAPI.channels.update(editingChannel.value.id, req)
      appStore.showSuccess(t('admin.channels.updateSuccess'))
    } else {
      const req: CreateChannelRequest = {
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        group_ids,
        model_pricing,
        model_mapping: Object.keys(model_mapping).length > 0 ? model_mapping : {},
        billing_model_source: form.billing_model_source,
        restrict_models: form.restrict_models
      }
      await adminAPI.channels.create(req)
      appStore.showSuccess(t('admin.channels.createSuccess'))
    }
    closeDialog()
    loadChannels()
  } catch (error: any) {
    const msg = error.response?.data?.detail || (
      editingChannel.value
        ? t('admin.channels.updateError')
        : t('admin.channels.createError')
    )
    appStore.showError(msg)
    console.error('Error saving channel:', error)
  } finally {
    submitting.value = false
  }
}

async function toggleChannelStatus(channel: Channel) {
  const newStatus = channel.status === 'active' ? 'disabled' : 'active'
  try {
    await adminAPI.channels.update(channel.id, { status: newStatus })
    if (filters.status && filters.status !== newStatus) {
      await loadChannels()
    } else {
      channel.status = newStatus
    }
  } catch (error) {
    appStore.showError(t('admin.channels.updateError'))
    console.error('Error toggling channel status:', error)
  }
}

function handleDelete(channel: Channel) {
  deletingChannel.value = channel
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!deletingChannel.value) return

  try {
    await adminAPI.channels.remove(deletingChannel.value.id)
    appStore.showSuccess(t('admin.channels.deleteSuccess'))
    showDeleteDialog.value = false
    deletingChannel.value = null
    loadChannels()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.channels.deleteError'))
    console.error('Error deleting channel:', error)
  }
}

onMounted(() => {
  loadChannels()
  loadGroups()
})

onUnmounted(() => {
  clearTimeout(searchTimeout)
  abortController?.abort()
})
</script>

<style scoped>
.channel-dialog-body {
  display: flex;
  flex-direction: column;
  height: 70vh;
  min-height: 400px;
}

.channel-view__search-icon,
.channel-view__text-soft,
.channel-view__text-muted {
  color: var(--theme-page-muted);
}

.channel-view__required {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.channel-view__text-strong {
  color: var(--theme-page-text);
}

.channel-view__tabs {
  padding-left: var(--theme-table-cell-padding-x);
  padding-right: var(--theme-table-cell-padding-x);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.channel-view__checkbox-row {
  color: var(--theme-page-text);
}

.channel-view__checkbox {
  accent-color: var(--theme-accent);
}

.channel-view__platform-toggle,
.channel-view__group-chip,
.channel-view__group-list,
.channel-view__empty-box {
  border-color: color-mix(in srgb, var(--theme-card-border) 74%, transparent);
}

.channel-view__platform-toggle,
.channel-view__group-chip {
  border-radius: calc(var(--theme-button-radius) + 2px);
}

.channel-view__platform-toggle {
  padding: calc(var(--theme-button-padding-y) * 0.6) calc(var(--theme-button-padding-x) * 0.75);
}

.channel-view__group-chip {
  padding: calc(var(--theme-button-padding-y) * 0.4) calc(var(--theme-button-padding-x) * 0.5);
}

.channel-view__platform-toggle,
.channel-view__group-chip {
  background: color-mix(in srgb, var(--theme-surface-soft) 78%, var(--theme-surface));
}

.channel-view__platform-toggle:hover,
.channel-view__group-chip:hover {
  background: color-mix(in srgb, var(--theme-table-row-hover) 100%, var(--theme-surface));
}

.channel-view__platform-toggle--active,
.channel-view__group-chip--selected {
  border-color: color-mix(in srgb, var(--theme-accent) 34%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 72%, var(--theme-surface));
}

.channel-view__group-list {
  max-height: calc(var(--theme-search-dropdown-max-height) * 0.67);
  overflow: auto;
  border-radius: var(--theme-select-panel-radius);
  padding: calc(var(--theme-user-api-keys-dropdown-padding) + 0.125rem);
  background: color-mix(in srgb, var(--theme-surface-soft) 70%, var(--theme-surface));
}

.channel-view__group-state {
  padding: calc(var(--theme-button-padding-y) * 0.45) 0;
}

.channel-view__rate-badge {
  border-radius: 999px;
  padding: 0.125rem 0.25rem;
}

.channel-view__empty-box {
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: calc(var(--theme-button-padding-y) * 0.45) calc(var(--theme-button-padding-x) * 0.4);
  border-style: dashed;
  color: color-mix(in srgb, var(--theme-page-muted) 78%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 52%, var(--theme-surface));
}

.channel-view__text-button {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
  transition: color 0.2s ease;
}

.channel-view__text-button:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%);
}

.channel-view__icon-button,
.channel-view__action-button {
  color: color-mix(in srgb, var(--theme-page-muted) 78%, transparent);
}

.channel-view__icon-button {
  border-radius: calc(var(--theme-button-radius) - 1px);
  padding: calc(var(--theme-button-padding-y) * 0.25);
}

.channel-view__action-button {
  border-radius: var(--theme-button-radius);
  padding: calc(var(--theme-button-padding-y) * 0.5);
}

.channel-view__icon-button--danger:hover,
.channel-view__action-button--danger:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.channel-view__action-button--info:hover {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.channel-view__tone-text--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.channel-view__tone-text--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.channel-view__tone-text--brand-orange {
  color: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 84%, var(--theme-page-text));
}

.channel-view__tone-text--brand-purple {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.channel-view__tone-text--brand-rose {
  color: color-mix(in srgb, rgb(var(--theme-brand-rose-rgb)) 84%, var(--theme-page-text));
}

.channel-tab {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  white-space: nowrap;
  border-bottom: 2px solid transparent;
  padding: calc(var(--theme-button-padding-y) * 0.75) calc(var(--theme-button-padding-x) * 0.75);
  font-size: 0.875rem;
  font-weight: 500;
  transition: color 0.2s ease, border-color 0.2s ease;
}

.channel-tab-active {
  border-color: color-mix(in srgb, var(--theme-accent) 84%, transparent);
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.channel-tab-inactive {
  border-color: transparent;
  color: var(--theme-page-muted);
}

.channel-tab-inactive:hover {
  border-color: color-mix(in srgb, var(--theme-card-border) 82%, transparent);
  color: color-mix(in srgb, var(--theme-page-text) 88%, var(--theme-page-muted));
}

@media (min-width: 640px) {
  .channel-view__tabs {
    padding-left: calc(var(--theme-table-cell-padding-x) * 1.2);
    padding-right: calc(var(--theme-table-cell-padding-x) * 1.2);
  }
}
</style>
