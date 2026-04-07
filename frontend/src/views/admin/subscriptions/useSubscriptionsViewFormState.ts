import { reactive } from 'vue'
import {
  createDefaultAssignSubscriptionForm,
  createDefaultExtendSubscriptionForm,
  createDefaultSubscriptionFilters,
  resetAssignSubscriptionForm,
  resetExtendSubscriptionForm,
  type AssignSubscriptionForm,
  type ExtendSubscriptionForm,
  type SubscriptionFiltersState
} from './subscriptionForm'

export function useSubscriptionsViewFormState() {
  const filters = reactive<SubscriptionFiltersState>(createDefaultSubscriptionFilters())
  const assignForm = reactive<AssignSubscriptionForm>(createDefaultAssignSubscriptionForm())
  const extendForm = reactive<ExtendSubscriptionForm>(createDefaultExtendSubscriptionForm())

  const setFilterStatus = (status: SubscriptionFiltersState['status']) => {
    filters.status = status
  }

  const setFilterGroupId = (groupId: string) => {
    filters.group_id = groupId
  }

  const setFilterPlatform = (platform: string) => {
    filters.platform = platform
  }

  const selectFilterUser = (userId: number) => {
    filters.user_id = userId
  }

  const clearFilterUser = () => {
    filters.user_id = null
  }

  const selectAssignUser = (userId: number) => {
    assignForm.user_id = userId
  }

  const clearAssignUser = () => {
    assignForm.user_id = null
  }

  const resetAssignFormState = () => {
    resetAssignSubscriptionForm(assignForm)
  }

  const resetExtendFormState = () => {
    resetExtendSubscriptionForm(extendForm)
  }

  return {
    filters,
    assignForm,
    extendForm,
    setFilterStatus,
    setFilterGroupId,
    setFilterPlatform,
    selectFilterUser,
    clearFilterUser,
    selectAssignUser,
    clearAssignUser,
    resetAssignFormState,
    resetExtendFormState
  }
}
