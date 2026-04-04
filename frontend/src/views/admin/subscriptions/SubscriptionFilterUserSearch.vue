<template>
  <div
    class="relative w-full sm:w-64"
    data-filter-user-search
  >
    <Icon
      name="search"
      size="md"
      class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
    />
    <input
      :value="keyword"
      type="text"
      :placeholder="t('admin.users.searchUsers')"
      class="input pl-10 pr-8"
      @input="handleInput"
      @focus="emit('focus')"
    />
    <button
      v-if="selectedUser"
      type="button"
      class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
      :title="t('common.clear')"
      @click="emit('clear-user')"
    >
      <Icon name="x" size="sm" :stroke-width="2" />
    </button>

    <div
      v-if="showDropdown && (results.length > 0 || keyword)"
      class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
    >
      <div
        v-if="loading"
        class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400"
      >
        {{ t('common.loading') }}
      </div>
      <div
        v-else-if="results.length === 0 && keyword"
        class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400"
      >
        {{ t('common.noOptionsFound') }}
      </div>
      <button
        v-for="user in results"
        :key="user.id"
        type="button"
        class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
        @click="emit('select-user', user)"
      >
        <span class="font-medium text-gray-900 dark:text-white">{{ user.email }}</span>
        <span class="ml-2 text-gray-500 dark:text-gray-400">#{{ user.id }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SimpleUser } from '@/api/admin/usage'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  keyword: string
  results: SimpleUser[]
  loading: boolean
  showDropdown: boolean
  selectedUser: SimpleUser | null
}>()

const emit = defineEmits<{
  'update:keyword': [value: string]
  search: []
  focus: []
  'select-user': [user: SimpleUser]
  'clear-user': []
}>()

const { t } = useI18n()

const handleInput = (event: Event) => {
  emit('update:keyword', (event.target as HTMLInputElement).value)
  emit('search')
}
</script>
