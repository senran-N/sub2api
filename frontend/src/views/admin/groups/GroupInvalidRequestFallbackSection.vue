<template>
  <div v-if="showSection" class="border-t pt-4">
    <label class="input-label">{{ t('admin.groups.invalidRequestFallback.title') }}</label>
    <Select
      v-model="form.fallback_group_id_on_invalid_request"
      :options="options"
      :placeholder="t('admin.groups.invalidRequestFallback.noFallback')"
    />
    <p class="input-hint">{{ t('admin.groups.invalidRequestFallback.hint') }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import type { CreateGroupForm, EditGroupForm, NullableNumberSelectOption } from '../groupsForm'

const props = defineProps<{
  form: CreateGroupForm | EditGroupForm
  options: NullableNumberSelectOption[]
}>()

const { t } = useI18n()

const showSection = computed(() => {
  return (
    ['anthropic', 'antigravity'].includes(props.form.platform) &&
    props.form.subscription_type !== 'subscription'
  )
})
</script>
