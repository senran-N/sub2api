<template>
  <div
    v-if="supportedPlatforms.includes(form.platform)"
    class="mt-4 space-y-4 border-t border-gray-200 pt-4 dark:border-dark-400"
  >
    <h4 class="mb-3 text-sm font-medium text-gray-700 dark:text-gray-300">
      账号过滤控制
    </h4>

    <div class="flex items-center justify-between">
      <div>
        <label class="text-sm text-gray-600 dark:text-gray-400">仅允许 OAuth 账号</label>
        <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
          {{ form.require_oauth_only ? '已启用 — 排除 API Key 类型账号' : '未启用' }}
        </p>
      </div>
      <Toggle v-model="form.require_oauth_only" />
    </div>

    <div class="flex items-center justify-between">
      <div>
        <label class="text-sm text-gray-600 dark:text-gray-400">
          仅允许隐私保护已设置的账号
        </label>
        <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
          {{ form.require_privacy_set ? '已启用 — Privacy 未设置的账号将被排除' : '未启用' }}
        </p>
      </div>
      <Toggle v-model="form.require_privacy_set" />
    </div>
  </div>
</template>

<script setup lang="ts">
import Toggle from '@/components/common/Toggle.vue'
import type { CreateGroupForm, EditGroupForm } from '../groupsForm'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const supportedPlatforms = ['openai', 'antigravity', 'anthropic', 'gemini']
</script>
