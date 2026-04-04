import { ref } from 'vue'
import type { SimpleUser } from '@/api/admin/usage'
import type { AssignSubscriptionForm, SubscriptionFiltersState } from './subscriptionForm'

interface SubscriptionsViewUserSearchesOptions {
  filters: SubscriptionFiltersState
  assignForm: AssignSubscriptionForm
  applyFilters: () => void
  searchUsers: (keyword: string) => Promise<SimpleUser[]>
}

export function useSubscriptionsViewUserSearches(
  options: SubscriptionsViewUserSearchesOptions
) {
  const filterUserKeyword = ref('')
  const filterUserResults = ref<SimpleUser[]>([])
  const filterUserLoading = ref(false)
  const showFilterUserDropdown = ref(false)
  const selectedFilterUser = ref<SimpleUser | null>(null)

  const userSearchKeyword = ref('')
  const userSearchResults = ref<SimpleUser[]>([])
  const userSearchLoading = ref(false)
  const showUserDropdown = ref(false)
  const selectedUser = ref<SimpleUser | null>(null)

  let filterUserSearchTimeout: ReturnType<typeof setTimeout> | null = null
  let userSearchTimeout: ReturnType<typeof setTimeout> | null = null

  const searchFilterUsers = async () => {
    const keyword = filterUserKeyword.value.trim()

    if (selectedFilterUser.value && keyword !== selectedFilterUser.value.email) {
      selectedFilterUser.value = null
      options.filters.user_id = null
      options.applyFilters()
    }

    if (!keyword) {
      filterUserResults.value = []
      return
    }

    filterUserLoading.value = true
    try {
      filterUserResults.value = await options.searchUsers(keyword)
    } catch (error) {
      console.error('Failed to search users:', error)
      filterUserResults.value = []
    } finally {
      filterUserLoading.value = false
    }
  }

  const debounceSearchFilterUsers = () => {
    if (filterUserSearchTimeout) {
      clearTimeout(filterUserSearchTimeout)
    }

    filterUserSearchTimeout = setTimeout(searchFilterUsers, 300)
  }

  const selectFilterUser = (user: SimpleUser) => {
    selectedFilterUser.value = user
    filterUserKeyword.value = user.email
    showFilterUserDropdown.value = false
    options.filters.user_id = user.id
    options.applyFilters()
  }

  const clearFilterUser = () => {
    selectedFilterUser.value = null
    filterUserKeyword.value = ''
    filterUserResults.value = []
    showFilterUserDropdown.value = false
    options.filters.user_id = null
    options.applyFilters()
  }

  const searchUsers = async () => {
    const keyword = userSearchKeyword.value.trim()

    if (selectedUser.value && keyword !== selectedUser.value.email) {
      selectedUser.value = null
      options.assignForm.user_id = null
    }

    if (!keyword) {
      userSearchResults.value = []
      return
    }

    userSearchLoading.value = true
    try {
      userSearchResults.value = await options.searchUsers(keyword)
    } catch (error) {
      console.error('Failed to search users:', error)
      userSearchResults.value = []
    } finally {
      userSearchLoading.value = false
    }
  }

  const debounceSearchUsers = () => {
    if (userSearchTimeout) {
      clearTimeout(userSearchTimeout)
    }

    userSearchTimeout = setTimeout(searchUsers, 300)
  }

  const selectUser = (user: SimpleUser) => {
    selectedUser.value = user
    userSearchKeyword.value = user.email
    showUserDropdown.value = false
    options.assignForm.user_id = user.id
  }

  const clearUserSelection = () => {
    selectedUser.value = null
    userSearchKeyword.value = ''
    userSearchResults.value = []
    options.assignForm.user_id = null
  }

  const resetAssignSearch = () => {
    selectedUser.value = null
    userSearchKeyword.value = ''
    userSearchResults.value = []
    showUserDropdown.value = false
    options.assignForm.user_id = null
  }

  const handleClickOutside = (event: MouseEvent) => {
    const target = event.target as HTMLElement
    if (!target.closest('[data-assign-user-search]')) {
      showUserDropdown.value = false
    }
    if (!target.closest('[data-filter-user-search]')) {
      showFilterUserDropdown.value = false
    }
  }

  const initialize = () => {
    document.addEventListener('click', handleClickOutside)
  }

  const dispose = () => {
    document.removeEventListener('click', handleClickOutside)

    if (filterUserSearchTimeout) {
      clearTimeout(filterUserSearchTimeout)
      filterUserSearchTimeout = null
    }
    if (userSearchTimeout) {
      clearTimeout(userSearchTimeout)
      userSearchTimeout = null
    }
  }

  return {
    filterUserKeyword,
    filterUserResults,
    filterUserLoading,
    showFilterUserDropdown,
    selectedFilterUser,
    userSearchKeyword,
    userSearchResults,
    userSearchLoading,
    showUserDropdown,
    selectedUser,
    debounceSearchFilterUsers,
    selectFilterUser,
    clearFilterUser,
    debounceSearchUsers,
    selectUser,
    clearUserSelection,
    resetAssignSearch,
    initialize,
    dispose
  }
}
