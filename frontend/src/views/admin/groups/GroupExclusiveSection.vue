<template>
  <div
    v-if="form.subscription_type !== 'subscription'"
    :data-tour="tourTarget"
  >
    <div class="mb-1.5 flex items-center gap-1">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('admin.groups.form.exclusive') }}
      </label>
      <div class="group relative inline-flex">
        <Icon
          name="questionCircle"
          size="sm"
          :stroke-width="2"
          class="cursor-help text-gray-400 transition-colors hover:text-primary-500 dark:text-gray-500 dark:hover:text-primary-400"
        />
        <div class="pointer-events-none absolute bottom-full left-0 z-50 mb-2 w-72 opacity-0 transition-all duration-200 group-hover:pointer-events-auto group-hover:opacity-100">
          <div class="rounded-lg bg-gray-900 p-3 text-white shadow-lg dark:bg-gray-800">
            <p class="mb-2 text-xs font-medium">{{ t('admin.groups.exclusiveTooltip.title') }}</p>
            <p class="mb-2 text-xs leading-relaxed text-gray-300">
              {{ t('admin.groups.exclusiveTooltip.description') }}
            </p>
            <div class="rounded bg-gray-800 p-2 dark:bg-gray-700">
              <p class="text-xs leading-relaxed text-gray-300">
                <span class="inline-flex items-center gap-1 text-primary-400">
                  <Icon name="lightbulb" size="xs" />
                  {{ t('admin.groups.exclusiveTooltip.example') }}
                </span>
                {{ t('admin.groups.exclusiveTooltip.exampleContent') }}
              </p>
            </div>
            <div class="absolute -bottom-1.5 left-3 h-3 w-3 rotate-45 bg-gray-900 dark:bg-gray-800"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="flex items-center gap-3">
      <Toggle v-model="form.is_exclusive" />
      <span class="text-sm text-gray-500 dark:text-gray-400">
        {{ form.is_exclusive ? t('admin.groups.exclusive') : t('admin.groups.public') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import type { CreateGroupForm, EditGroupForm } from '../groupsForm'

withDefaults(
  defineProps<{
    form: CreateGroupForm | EditGroupForm
    tourTarget?: string
  }>(),
  {
    tourTarget: undefined
  }
)

const { t } = useI18n()
</script>
