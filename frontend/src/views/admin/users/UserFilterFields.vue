<template>
  <div class="flex flex-1 flex-wrap items-center gap-3">
    <div class="relative w-full md:w-64">
      <Icon
        name="search"
        size="md"
        class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
      />
      <input
        :value="searchQuery"
        type="text"
        :placeholder="t('admin.users.searchUsers')"
        class="input pl-10"
        @input="handleSearchInput"
      />
    </div>

    <div v-if="visibleFilters.has('role')" class="w-full sm:w-32">
      <Select
        v-model="filters.role"
        :options="roleOptions"
        @change="applyFilter"
      />
    </div>

    <div v-if="visibleFilters.has('status')" class="w-full sm:w-32">
      <Select
        v-model="filters.status"
        :options="statusOptions"
        @change="applyFilter"
      />
    </div>

    <div v-if="visibleFilters.has('group')" class="w-full sm:w-44">
      <Select
        v-model="filters.group"
        :options="groupFilterOptions"
        searchable
        creatable
        :creatable-prefix="t('admin.users.fuzzySearch')"
        :search-placeholder="t('admin.users.searchGroups')"
        @change="applyFilter"
      />
    </div>

    <template v-for="(value, attrId) in activeAttributeFilters" :key="attrId">
      <div
        v-if="visibleFilters.has(`attr_${attrId}`)"
        class="relative w-full sm:w-36"
      >
        <input
          v-if="isTextAttribute(attrId)"
          :value="value"
          :placeholder="getAttributeDefinitionName(Number(attrId))"
          class="input w-full"
          @input="updateTextAttributeFilter(Number(attrId), $event)"
          @keyup.enter="applyFilter"
        />
        <input
          v-else-if="getAttributeDefinition(Number(attrId))?.type === 'number'"
          :value="value"
          type="number"
          :placeholder="getAttributeDefinitionName(Number(attrId))"
          class="input w-full"
          @input="updateTextAttributeFilter(Number(attrId), $event)"
          @keyup.enter="applyFilter"
        />
        <div v-else-if="isSelectAttribute(attrId)" class="w-full">
          <Select
            :model-value="value"
            :options="selectAttributeOptions(attrId)"
            @update:model-value="updateSelectAttributeFilter(Number(attrId), $event)"
          />
        </div>
        <input
          v-else
          :value="value"
          :placeholder="getAttributeDefinitionName(Number(attrId))"
          class="input w-full"
          @input="updateTextAttributeFilter(Number(attrId), $event)"
          @keyup.enter="applyFilter"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SelectOption } from '@/components/common/Select.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { UserAttributeDefinition } from '@/types'
import type { UsersFilterState } from '../usersTable'

const props = defineProps<{
  searchQuery: string
  filters: UsersFilterState
  visibleFilters: Set<string>
  groupFilterOptions: SelectOption[]
  activeAttributeFilters: Record<number, string>
  getAttributeDefinition: (attrId: number) => UserAttributeDefinition | undefined
  getAttributeDefinitionName: (attrId: number) => string
  updateAttributeFilter: (attrId: number, value: string) => void
  applyFilter: () => void
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'search-input': []
}>()

const { t } = useI18n()

const roleOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.users.allRoles') },
  { value: 'admin', label: t('admin.users.admin') },
  { value: 'user', label: t('admin.users.user') }
])

const statusOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.users.allStatus') },
  { value: 'active', label: t('common.active') },
  { value: 'disabled', label: t('admin.users.disabled') }
])

function handleSearchInput(event: Event) {
  emit('update:searchQuery', (event.target as HTMLInputElement).value)
  emit('search-input')
}

function isTextAttribute(attributeId: number): boolean {
  return ['text', 'textarea', 'email', 'url', 'date'].includes(
    props.getAttributeDefinition(Number(attributeId))?.type || 'text'
  )
}

function isSelectAttribute(attributeId: number): boolean {
  return ['select', 'multi_select'].includes(
    props.getAttributeDefinition(Number(attributeId))?.type || ''
  )
}

function selectAttributeOptions(attributeId: number): SelectOption[] {
  return [
    { value: '', label: props.getAttributeDefinitionName(attributeId) },
    ...((props.getAttributeDefinition(attributeId)?.options || []) as SelectOption[])
  ]
}

function updateTextAttributeFilter(attributeId: number, event: Event) {
  props.updateAttributeFilter(attributeId, (event.target as HTMLInputElement).value)
}

function updateSelectAttributeFilter(
  attributeId: number,
  value: string | number | boolean | null
) {
  props.updateAttributeFilter(attributeId, String(value ?? ''))
  props.applyFilter()
}
</script>
