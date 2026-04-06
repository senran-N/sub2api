<template>
  <div
    class="subscription-filter-user-search relative w-full"
    data-filter-user-search
  >
    <Icon
      name="search"
      size="md"
      class="subscription-filter-user-search__icon absolute left-3 top-1/2 -translate-y-1/2"
    />
    <input
      :value="keyword"
      type="text"
      :placeholder="t('admin.users.searchUsers')"
      class="input subscription-filter-user-search__input"
      @input="handleInput"
      @focus="emit('focus')"
    />
    <button
      v-if="selectedUser"
      type="button"
      class="subscription-filter-user-search__clear absolute top-1/2 -translate-y-1/2"
      :title="t('common.clear')"
      @click="emit('clear-user')"
    >
      <Icon name="x" size="sm" :stroke-width="2" />
    </button>

    <div
      v-if="showDropdown && (results.length > 0 || keyword)"
      class="subscription-filter-user-search__dropdown absolute z-50 w-full overflow-auto"
    >
      <div
        v-if="loading"
        class="subscription-filter-user-search__muted subscription-filter-user-search__status text-sm"
      >
        {{ t('common.loading') }}
      </div>
      <div
        v-else-if="results.length === 0 && keyword"
        class="subscription-filter-user-search__muted subscription-filter-user-search__status text-sm"
      >
        {{ t('common.noOptionsFound') }}
      </div>
      <button
        v-for="user in results"
        :key="user.id"
        type="button"
        class="subscription-filter-user-search__option w-full text-left text-sm"
        @click="emit('select-user', user)"
      >
        <span class="subscription-filter-user-search__option-email font-medium">{{ user.email }}</span>
        <span class="subscription-filter-user-search__muted ml-2">#{{ user.id }}</span>
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

<style scoped>
.subscription-filter-user-search__icon,
.subscription-filter-user-search__clear,
.subscription-filter-user-search__muted {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.subscription-filter-user-search__input {
  padding-left: calc(var(--theme-button-padding-x) + 0.5rem);
  padding-right: calc(var(--theme-button-padding-x) * 0.8 + 0.25rem);
}

.subscription-filter-user-search {
  --subscription-filter-user-search-control-width: var(--theme-settings-menu-width-md);
}

@media (min-width: 640px) {
  .subscription-filter-user-search {
    width: var(--subscription-filter-user-search-control-width);
  }
}

.subscription-filter-user-search__clear {
  right: calc(var(--theme-floating-panel-gap) * 0.5 + 0.375rem);
}

.subscription-filter-user-search__clear:hover {
  color: var(--theme-page-text);
}

.subscription-filter-user-search__dropdown {
  margin-top: var(--theme-floating-panel-gap);
  max-height: var(--theme-search-dropdown-max-height);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.subscription-filter-user-search__status {
  padding: calc(var(--theme-button-padding-y) * 1.1) var(--theme-button-padding-x);
}

.subscription-filter-user-search__option {
  padding: calc(var(--theme-button-padding-y) * 0.8) var(--theme-button-padding-x);
  transition: background-color 0.2s ease;
}

.subscription-filter-user-search__option:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.subscription-filter-user-search__option-email {
  color: var(--theme-page-text);
}
</style>
