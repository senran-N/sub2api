<template>
  <div class="space-y-5">
    <div>
      <label :id="nameLabelId" :for="nameInputId" class="input-label">
        {{ t('admin.groups.form.name') }}
      </label>
      <input
        :id="nameInputId"
        name="name"
        v-model="form.name"
        type="text"
        autocomplete="off"
        required
        class="input"
        :placeholder="namePlaceholder"
        :data-tour="nameTourTarget"
      />
    </div>
    <div>
      <label :id="descriptionLabelId" :for="descriptionInputId" class="input-label">
        {{ t('admin.groups.form.description') }}
      </label>
      <textarea
        :id="descriptionInputId"
        name="description"
        v-model="form.description"
        rows="3"
        autocomplete="off"
        class="input"
        :placeholder="descriptionPlaceholder"
      ></textarea>
    </div>
    <div>
      <label :id="platformLabelId" class="input-label">{{ t('admin.groups.form.platform') }}</label>
      <Select
        :id="platformInputId"
        name="platform"
        v-model="form.platform"
        :options="platformOptions"
        :disabled="platformDisabled"
        :aria-labelledby="platformLabelId"
        :data-tour="platformTourTarget"
        @change="handlePlatformChange"
      />
      <p class="input-hint">{{ platformHint }}</p>
    </div>
    <GroupCopyAccountsField
      :options="copyAccountsOptions"
      :selected-group-ids="form.copy_accounts_from_group_ids"
      :tooltip-text="copyAccountsTooltipText"
      :hint-text="copyAccountsHintText"
      @add-group="$emit('add-group', $event)"
      @remove-group="$emit('remove-group', $event)"
    />
    <div>
      <label :id="rateMultiplierLabelId" :for="rateMultiplierInputId" class="input-label">
        {{ t('admin.groups.form.rateMultiplier') }}
      </label>
      <input
        :id="rateMultiplierInputId"
        name="rate_multiplier"
        v-model.number="form.rate_multiplier"
        type="number"
        autocomplete="off"
        step="0.001"
        min="0.001"
        required
        class="input"
        :data-tour="rateMultiplierTourTarget"
      />
      <p v-if="rateMultiplierHint" class="input-hint">{{ rateMultiplierHint }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import type { CreateGroupForm, EditGroupForm, NumberSelectOption } from './groupsForm'
import GroupCopyAccountsField from './GroupCopyAccountsField.vue'

let groupBaseFieldsIdCounter = 0

const props = withDefaults(
  defineProps<{
    form: CreateGroupForm | EditGroupForm
    platformOptions: SelectOption[]
    copyAccountsOptions: NumberSelectOption[]
    copyAccountsTooltipText: string
    copyAccountsHintText: string
    platformHint: string
    platformDisabled?: boolean
    namePlaceholder?: string
    descriptionPlaceholder?: string
    rateMultiplierHint?: string
    resetCopyAccountsOnPlatformChange?: boolean
    nameTourTarget?: string
    platformTourTarget?: string
    rateMultiplierTourTarget?: string
  }>(),
  {
    platformDisabled: false,
    namePlaceholder: undefined,
    descriptionPlaceholder: undefined,
    rateMultiplierHint: undefined,
    resetCopyAccountsOnPlatformChange: false,
    nameTourTarget: undefined,
    platformTourTarget: undefined,
    rateMultiplierTourTarget: undefined
  }
)

defineEmits<{
  'add-group': [groupId: number]
  'remove-group': [groupId: number]
}>()

const { t } = useI18n()
const fieldIdPrefix = `group-base-fields-${++groupBaseFieldsIdCounter}`
const nameInputId = computed(() => `${fieldIdPrefix}-name`)
const nameLabelId = computed(() => `${fieldIdPrefix}-name-label`)
const descriptionInputId = computed(() => `${fieldIdPrefix}-description`)
const descriptionLabelId = computed(() => `${fieldIdPrefix}-description-label`)
const platformInputId = computed(() => `${fieldIdPrefix}-platform`)
const platformLabelId = computed(() => `${fieldIdPrefix}-platform-label`)
const rateMultiplierInputId = computed(() => `${fieldIdPrefix}-rate-multiplier`)
const rateMultiplierLabelId = computed(() => `${fieldIdPrefix}-rate-multiplier-label`)

function handlePlatformChange(): void {
  if (!props.resetCopyAccountsOnPlatformChange) {
    return
  }
  props.form.copy_accounts_from_group_ids = []
}
</script>
