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
            class="input subscription-assign-dialog__search-input"
            :placeholder="t('admin.usage.searchUserPlaceholder')"
            @input="emit('update:userKeyword', ($event.target as HTMLInputElement).value)"
            @focus="emit('show-user-dropdown')"
            @input.capture="emit('search-users')"
          />
          <button
            v-if="selectedUser"
            type="button"
            class="subscription-assign-dialog__clear absolute top-1/2 -translate-y-1/2"
            @click="emit('clear-user')"
          >
            <Icon name="x" size="sm" :stroke-width="2" />
          </button>
          <div
            v-if="showUserDropdown && (userResults.length > 0 || userKeyword)"
            class="subscription-assign-dialog__dropdown absolute z-50 w-full overflow-auto"
          >
            <div
              v-if="userLoading"
              class="subscription-assign-dialog__muted subscription-assign-dialog__status text-sm"
            >
              {{ t('common.loading') }}
            </div>
            <div
              v-else-if="userResults.length === 0 && userKeyword"
              class="subscription-assign-dialog__muted subscription-assign-dialog__status text-sm"
            >
              {{ t('common.noOptionsFound') }}
            </div>
            <button
              v-for="user in userResults"
              :key="user.id"
              type="button"
              class="subscription-assign-dialog__option w-full text-left text-sm"
              @click="emit('select-user', user)"
            >
              <span class="subscription-assign-dialog__option-email font-medium">{{ user.email }}</span>
              <span class="subscription-assign-dialog__muted ml-2">#{{ user.id }}</span>
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
            <span v-else class="subscription-assign-dialog__muted">{{ t('admin.subscriptions.selectGroup') }}</span>
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
import type { AssignSubscriptionForm, SubscriptionGroupOption } from './subscriptionForm'

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

<style scoped>
.subscription-assign-dialog__clear,
.subscription-assign-dialog__muted {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.subscription-assign-dialog__search-input {
  padding-right: calc(var(--theme-button-padding-x) * 0.8 + 0.25rem);
}

.subscription-assign-dialog__clear {
  right: calc(var(--theme-floating-panel-gap) * 0.5 + 0.375rem);
}

.subscription-assign-dialog__clear:hover {
  color: var(--theme-page-text);
}

.subscription-assign-dialog__dropdown {
  margin-top: var(--theme-floating-panel-gap);
  max-height: var(--theme-search-dropdown-max-height);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.subscription-assign-dialog__status {
  padding: calc(var(--theme-button-padding-y) * 1.1) var(--theme-button-padding-x);
}

.subscription-assign-dialog__option {
  padding: calc(var(--theme-button-padding-y) * 0.8) var(--theme-button-padding-x);
  transition: background-color 0.2s ease;
}

.subscription-assign-dialog__option:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.subscription-assign-dialog__option-email {
  color: var(--theme-page-text);
}
</style>
