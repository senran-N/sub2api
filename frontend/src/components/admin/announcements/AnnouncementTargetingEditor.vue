<template>
  <div class="announcement-targeting-editor">
    <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <div class="announcement-targeting-editor__title text-sm font-medium">
          {{ t('admin.announcements.form.targetingMode') }}
        </div>
        <div class="announcement-targeting-editor__description mt-1 text-xs">
          {{ mode === 'all' ? t('admin.announcements.form.targetingAll') : t('admin.announcements.form.targetingCustom') }}
        </div>
      </div>

      <div class="flex items-center gap-3">
        <label class="announcement-targeting-editor__radio-label flex items-center gap-2 text-sm">
          <input
            type="radio"
            name="announcement-targeting-mode"
            value="all"
            :checked="mode === 'all'"
            @change="setMode('all')"
            class="announcement-targeting-editor__radio h-4 w-4"
          />
          {{ t('admin.announcements.form.targetingAll') }}
        </label>
        <label class="announcement-targeting-editor__radio-label flex items-center gap-2 text-sm">
          <input
            type="radio"
            name="announcement-targeting-mode"
            value="custom"
            :checked="mode === 'custom'"
            @change="setMode('custom')"
            class="announcement-targeting-editor__radio h-4 w-4"
          />
          {{ t('admin.announcements.form.targetingCustom') }}
        </label>
      </div>
    </div>

    <div v-if="mode === 'custom'" class="mt-4 space-y-4">
      <div class="flex items-center justify-between">
        <div class="announcement-targeting-editor__title text-sm font-medium">
          OR
          <span class="announcement-targeting-editor__description ml-1 text-xs font-normal">
            ({{ anyOf.length }}/50)
          </span>
        </div>
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="anyOf.length >= 50"
          @click="addOrGroup"
        >
          <Icon name="plus" size="sm" class="mr-1" />
          {{ t('admin.announcements.form.addOrGroup') }}
        </button>
      </div>

      <div v-if="anyOf.length === 0" class="announcement-targeting-editor__empty text-sm">
        {{ t('admin.announcements.form.targetingCustom') }}: {{ t('admin.announcements.form.addOrGroup') }}
      </div>

      <div
        v-for="(group, groupIndex) in anyOf"
        :key="groupIndex"
        class="announcement-targeting-editor__group"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="announcement-targeting-editor__title text-sm font-medium">
              {{ t('admin.announcements.form.targetingCustom') }} #{{ groupIndex + 1 }}
              <span class="announcement-targeting-editor__description ml-2 text-xs font-normal">AND ({{ (group.all_of?.length || 0) }}/50)</span>
            </div>
            <div class="announcement-targeting-editor__description mt-1 text-xs">
              {{ t('admin.announcements.form.addAndCondition') }}
            </div>
          </div>

          <button
            type="button"
            class="btn btn-secondary"
            @click="removeOrGroup(groupIndex)"
          >
            <Icon name="trash" size="sm" class="mr-1" />
            {{ t('common.delete') }}
          </button>
        </div>

        <div class="mt-4 space-y-3">
          <div
            v-for="(cond, condIndex) in (group.all_of || [])"
            :key="condIndex"
            class="announcement-targeting-editor__condition"
          >
            <div class="flex flex-col gap-3 md:flex-row md:items-end">
              <div class="w-full md:w-52">
                <label class="input-label">{{ t('admin.announcements.form.conditionType') }}</label>
                <Select
                  :model-value="cond.type"
                  :options="conditionTypeOptions"
                  @update:model-value="(v) => setConditionType(groupIndex, condIndex, v as any)"
                />
              </div>

              <div v-if="cond.type === 'subscription'" class="flex-1">
                <label class="input-label">{{ t('admin.announcements.form.selectPackages') }}</label>
                <GroupSelector
                  v-model="subscriptionSelections[groupIndex][condIndex]"
                  :groups="groups"
                />
              </div>

              <div v-else class="flex flex-1 flex-col gap-3 sm:flex-row">
                <div class="w-full sm:w-44">
                  <label class="input-label">{{ t('admin.announcements.form.operator') }}</label>
                  <Select
                    :model-value="cond.operator"
                    :options="balanceOperatorOptions"
                    @update:model-value="(v) => setOperator(groupIndex, condIndex, v as any)"
                  />
                </div>
                <div class="w-full sm:flex-1">
                  <label class="input-label">{{ t('admin.announcements.form.balanceValue') }}</label>
                  <input
                    :value="String(cond.value ?? '')"
                    type="number"
                    step="any"
                    class="input"
                    @input="(e) => setBalanceValue(groupIndex, condIndex, (e.target as HTMLInputElement).value)"
                  />
                </div>
              </div>

              <div class="flex justify-end">
                <button
                  type="button"
                  class="btn btn-secondary"
                  @click="removeAndCondition(groupIndex, condIndex)"
                >
                  <Icon name="trash" size="sm" class="mr-1" />
                  {{ t('common.delete') }}
                </button>
              </div>
            </div>
          </div>

          <div class="flex justify-end">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="(group.all_of?.length || 0) >= 50"
              @click="addAndCondition(groupIndex)"
            >
              <Icon name="plus" size="sm" class="mr-1" />
              {{ t('admin.announcements.form.addAndCondition') }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="validationError" class="announcement-targeting-editor__validation text-sm">
        {{ validationError }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type {
  AdminGroup,
  AnnouncementTargeting,
  AnnouncementCondition,
  AnnouncementConditionGroup,
  AnnouncementConditionType,
  AnnouncementOperator
} from '@/types'

import Select from '@/components/common/Select.vue'
import GroupSelector from '@/components/common/GroupSelector.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const props = defineProps<{
  modelValue: AnnouncementTargeting
  groups: AdminGroup[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: AnnouncementTargeting): void
}>()

const anyOf = computed(() => props.modelValue?.any_of ?? [])

type Mode = 'all' | 'custom'
const mode = computed<Mode>(() => (anyOf.value.length === 0 ? 'all' : 'custom'))

const conditionTypeOptions = computed(() => [
  { value: 'subscription', label: t('admin.announcements.form.conditionSubscription') },
  { value: 'balance', label: t('admin.announcements.form.conditionBalance') }
])

const balanceOperatorOptions = computed(() => [
  { value: 'gt', label: t('admin.announcements.operators.gt') },
  { value: 'gte', label: t('admin.announcements.operators.gte') },
  { value: 'lt', label: t('admin.announcements.operators.lt') },
  { value: 'lte', label: t('admin.announcements.operators.lte') },
  { value: 'eq', label: t('admin.announcements.operators.eq') }
])

function setMode(next: Mode) {
  if (next === 'all') {
    emit('update:modelValue', { any_of: [] })
    return
  }
  if (anyOf.value.length === 0) {
    emit('update:modelValue', { any_of: [{ all_of: [defaultSubscriptionCondition()] }] })
  }
}

function defaultSubscriptionCondition(): AnnouncementCondition {
  return {
    type: 'subscription' as AnnouncementConditionType,
    operator: 'in' as AnnouncementOperator,
    group_ids: []
  }
}

function defaultBalanceCondition(): AnnouncementCondition {
  return {
    type: 'balance' as AnnouncementConditionType,
    operator: 'gte' as AnnouncementOperator,
    value: 0
  }
}

type TargetingDraft = {
  any_of: AnnouncementConditionGroup[]
}

function updateTargeting(mutator: (draft: TargetingDraft) => void) {
  const draft: TargetingDraft = JSON.parse(JSON.stringify(props.modelValue ?? { any_of: [] }))
  if (!draft.any_of) draft.any_of = []
  mutator(draft)
  emit('update:modelValue', draft)
}

function addOrGroup() {
  updateTargeting((draft) => {
    if (draft.any_of.length >= 50) return
    draft.any_of.push({ all_of: [defaultSubscriptionCondition()] })
  })
}

function removeOrGroup(groupIndex: number) {
  updateTargeting((draft) => {
    draft.any_of.splice(groupIndex, 1)
  })
}

function addAndCondition(groupIndex: number) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group.all_of) group.all_of = []
    if (group.all_of.length >= 50) return
    group.all_of.push(defaultSubscriptionCondition())
  })
}

function removeAndCondition(groupIndex: number, condIndex: number) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return
    group.all_of.splice(condIndex, 1)
  })
}

function setConditionType(groupIndex: number, condIndex: number, nextType: AnnouncementConditionType) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    if (nextType === 'subscription') {
      group.all_of[condIndex] = defaultSubscriptionCondition()
    } else {
      group.all_of[condIndex] = defaultBalanceCondition()
    }
  })
}

function setOperator(groupIndex: number, condIndex: number, op: AnnouncementOperator) {
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    const cond = group.all_of[condIndex]
    if (!cond) return

    cond.operator = op
  })
}

function setBalanceValue(groupIndex: number, condIndex: number, raw: string) {
  const n = raw === '' ? 0 : Number(raw)
  updateTargeting((draft) => {
    const group = draft.any_of[groupIndex]
    if (!group?.all_of) return

    const cond = group.all_of[condIndex]
    if (!cond) return

    cond.value = Number.isFinite(n) ? n : 0
  })
}

// We keep group_ids selection in a parallel reactive map because GroupSelector is numeric list.
// Then we mirror it back to targeting.group_ids via a watcher.
const subscriptionSelections = reactive<Record<number, Record<number, number[]>>>({})

function ensureSelectionPath(groupIndex: number, condIndex: number) {
  if (!subscriptionSelections[groupIndex]) subscriptionSelections[groupIndex] = {}
  if (!subscriptionSelections[groupIndex][condIndex]) subscriptionSelections[groupIndex][condIndex] = []
}

// Sync from modelValue to subscriptionSelections (one-way: model -> local state)
watch(
  () => props.modelValue,
  (v) => {
    const groups = v?.any_of ?? []
    for (let gi = 0; gi < groups.length; gi++) {
      const allOf = groups[gi]?.all_of ?? []
      for (let ci = 0; ci < allOf.length; ci++) {
        const c = allOf[ci]
        if (c?.type === 'subscription') {
          ensureSelectionPath(gi, ci)
          // Only update if different to avoid triggering unnecessary updates
          const newIds = (c.group_ids ?? []).slice()
          const currentIds = subscriptionSelections[gi]?.[ci] ?? []
          if (JSON.stringify(newIds.sort()) !== JSON.stringify(currentIds.sort())) {
            subscriptionSelections[gi][ci] = newIds
          }
        }
      }
    }
  },
  { immediate: true }
)

// Sync from subscriptionSelections to modelValue (one-way: local state -> model)
// Use a debounced approach to avoid infinite loops
let syncTimeout: ReturnType<typeof setTimeout> | null = null
watch(
  () => subscriptionSelections,
  () => {
    // Debounce the sync to avoid rapid fire updates
    if (syncTimeout) clearTimeout(syncTimeout)

    syncTimeout = setTimeout(() => {
      // Build the new targeting state
      const newTargeting: TargetingDraft = JSON.parse(JSON.stringify(props.modelValue ?? { any_of: [] }))
      if (!newTargeting.any_of) newTargeting.any_of = []

      const groups = newTargeting.any_of ?? []
      for (let gi = 0; gi < groups.length; gi++) {
        const allOf = groups[gi]?.all_of ?? []
        for (let ci = 0; ci < allOf.length; ci++) {
          const c = allOf[ci]
          if (c?.type === 'subscription') {
            ensureSelectionPath(gi, ci)
            c.operator = 'in' as AnnouncementOperator
            c.group_ids = (subscriptionSelections[gi]?.[ci] ?? []).slice()
          }
        }
      }

      // Only emit if there's an actual change (deep comparison)
      if (JSON.stringify(props.modelValue) !== JSON.stringify(newTargeting)) {
        emit('update:modelValue', newTargeting)
      }
    }, 0)
  },
  { deep: true }
)

const validationError = computed(() => {
  if (mode.value !== 'custom') return ''

  const groups = anyOf.value
  if (groups.length === 0) return t('admin.announcements.form.addOrGroup')

  if (groups.length > 50) return 'any_of > 50'

  for (const g of groups) {
    const allOf = g?.all_of ?? []
    if (allOf.length === 0) return t('admin.announcements.form.addAndCondition')
    if (allOf.length > 50) return 'all_of > 50'

    for (const c of allOf) {
      if (c.type === 'subscription') {
        if (!c.group_ids || c.group_ids.length === 0) return t('admin.announcements.form.selectPackages')
      }
    }
  }

  return ''
})
</script>

<style scoped>
.announcement-targeting-editor {
  border-radius: var(--theme-announcement-targeting-editor-radius);
  padding: var(--theme-announcement-targeting-editor-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 92%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
}

.announcement-targeting-editor__title {
  color: var(--theme-page-text);
}

.announcement-targeting-editor__description,
.announcement-targeting-editor__radio-label,
.announcement-targeting-editor__empty {
  color: var(--theme-page-muted);
}

.announcement-targeting-editor__radio {
  accent-color: var(--theme-accent);
}

.announcement-targeting-editor__empty {
  border-radius: var(--theme-announcement-targeting-empty-radius);
  padding: var(--theme-announcement-targeting-empty-padding);
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 76%, transparent);
}

.announcement-targeting-editor__group {
  border-radius: var(--theme-announcement-targeting-group-radius);
  padding: var(--theme-announcement-targeting-group-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 92%, transparent);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.announcement-targeting-editor__condition {
  border-radius: var(--theme-announcement-targeting-condition-radius);
  padding: var(--theme-announcement-targeting-condition-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
}

.announcement-targeting-editor__validation {
  border-radius: var(--theme-announcement-targeting-validation-radius);
  padding: var(--theme-announcement-targeting-validation-padding);
  border: 1px solid color-mix(in srgb, rgb(var(--theme-danger-rgb)) 26%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
