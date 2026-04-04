<template>
  <div class="space-y-5">
    <div>
      <label class="input-label">{{ t('admin.groups.form.name') }}</label>
      <input
        v-model="form.name"
        type="text"
        required
        class="input"
        :placeholder="namePlaceholder"
        :data-tour="nameTourTarget"
      />
    </div>
    <div>
      <label class="input-label">{{ t('admin.groups.form.description') }}</label>
      <textarea
        v-model="form.description"
        rows="3"
        class="input"
        :placeholder="descriptionPlaceholder"
      ></textarea>
    </div>
    <div>
      <label class="input-label">{{ t('admin.groups.form.platform') }}</label>
      <Select
        v-model="form.platform"
        :options="platformOptions"
        :disabled="platformDisabled"
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
      <label class="input-label">{{ t('admin.groups.form.rateMultiplier') }}</label>
      <input
        v-model.number="form.rate_multiplier"
        type="number"
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
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import type { CreateGroupForm, EditGroupForm, NumberSelectOption } from '../groupsForm'
import GroupCopyAccountsField from './GroupCopyAccountsField.vue'

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

function handlePlatformChange(): void {
  if (!props.resetCopyAccountsOnPlatformChange) {
    return
  }
  props.form.copy_accounts_from_group_ids = []
}
</script>
