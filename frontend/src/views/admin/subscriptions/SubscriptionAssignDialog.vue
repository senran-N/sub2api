<template>
  <BaseDialog
    :show="show"
    :title="t('admin.subscriptions.assignSubscription')"
    width="normal"
    @close="emit('close')"
  >
    <form
      id="assign-subscription-form"
      class="space-y-5"
      @submit.prevent="emit('submit')"
    >
      <div>
        <label class="input-label">{{ t('admin.subscriptions.form.user') }}</label>
        <div class="relative" data-assign-user-search>
          <input
            :value="userKeyword"
            type="text"
            class="input pr-8"
            :placeholder="t('admin.usage.searchUserPlaceholder')"
            @input="emit('update:userKeyword', ($event.target as HTMLInputElement).value)"
            @focus="emit('show-user-dropdown')"
            @input.capture="emit('search-users')"
          />
          <button
            v-if="selectedUser"
            type="button"
            class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            @click="emit('clear-user')"
          >
            <Icon name="x" size="sm" :stroke-width="2" />
          </button>
          <div
            v-if="showUserDropdown && (userResults.length > 0 || userKeyword)"
            class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
          >
            <div
              v-if="userLoading"
              class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400"
            >
              {{ t('common.loading') }}
            </div>
            <div
              v-else-if="userResults.length === 0 && userKeyword"
              class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400"
            >
              {{ t('common.noOptionsFound') }}
            </div>
            <button
              v-for="user in userResults"
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
      </div>

      <div>
        <label class="input-label">{{ t('admin.subscriptions.form.group') }}</label>
        <Select
          v-model="form.group_id"
          :options="groupOptions"
          :placeholder="t('admin.subscriptions.selectGroup')"
        >
          <template #selected="{ option }">
            <GroupBadge
              v-if="option"
              :name="(option as unknown as SubscriptionGroupOption).label"
              :platform="(option as unknown as SubscriptionGroupOption).platform"
              :subscription-type="(option as unknown as SubscriptionGroupOption).subscriptionType"
              :rate-multiplier="(option as unknown as SubscriptionGroupOption).rate"
            />
            <span v-else class="text-gray-400">{{ t('admin.subscriptions.selectGroup') }}</span>
          </template>
          <template #option="{ option, selected }">
            <GroupOptionItem
              :name="(option as unknown as SubscriptionGroupOption).label"
              :platform="(option as unknown as SubscriptionGroupOption).platform"
              :subscription-type="(option as unknown as SubscriptionGroupOption).subscriptionType"
              :rate-multiplier="(option as unknown as SubscriptionGroupOption).rate"
              :description="(option as unknown as SubscriptionGroupOption).description"
              :selected="selected"
            />
          </template>
        </Select>
        <p class="input-hint">{{ t('admin.subscriptions.groupHint') }}</p>
      </div>

      <div>
        <label class="input-label">{{ t('admin.subscriptions.form.validityDays') }}</label>
        <input v-model.number="form.validity_days" type="number" min="1" class="input" />
        <p class="input-hint">{{ t('admin.subscriptions.validityHint') }}</p>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="assign-subscription-form"
          :disabled="submitting"
          class="btn btn-primary"
        >
          <svg
            v-if="submitting"
            class="-ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ submitting ? t('admin.subscriptions.assigning') : t('admin.subscriptions.assign') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SimpleUser } from '@/api/admin/usage'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AssignSubscriptionForm, SubscriptionGroupOption } from '../subscriptionForm'

defineProps<{
  show: boolean
  form: AssignSubscriptionForm
  userKeyword: string
  userResults: SimpleUser[]
  userLoading: boolean
  showUserDropdown: boolean
  selectedUser: SimpleUser | null
  groupOptions: SubscriptionGroupOption[]
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
  submit: []
  'update:userKeyword': [value: string]
  'show-user-dropdown': []
  'search-users': []
  'select-user': [user: SimpleUser]
  'clear-user': []
}>()

const { t } = useI18n()
</script>
