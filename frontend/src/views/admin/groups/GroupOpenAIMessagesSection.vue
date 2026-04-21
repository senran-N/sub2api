<template>
  <div
    v-if="form.platform === 'openai'"
    class="group-openai-messages-section mt-4 border-t pt-4"
  >
    <h3 class="group-openai-messages-section__title mb-3 text-sm font-medium">
      {{ t('admin.groups.openaiMessages.title') }}
    </h3>

    <div class="flex items-center justify-between">
      <label class="group-openai-messages-section__label text-sm">
        {{ t('admin.groups.openaiMessages.allowDispatch') }}
      </label>
      <Toggle
        v-model="form.allow_messages_dispatch"
        :aria-label="t('admin.groups.openaiMessages.allowDispatch')"
      />
    </div>
    <p class="group-openai-messages-section__hint mt-1 text-xs">
      {{ t('admin.groups.openaiMessages.allowDispatchHint') }}
    </p>

    <div v-if="form.allow_messages_dispatch" class="mt-3">
      <label class="input-label">{{ t('admin.groups.openaiMessages.defaultModel') }}</label>
      <input
        v-model="form.default_mapped_model"
        type="text"
        :placeholder="t('admin.groups.openaiMessages.defaultModelPlaceholder')"
        class="input"
      />
      <p class="input-hint">{{ t('admin.groups.openaiMessages.defaultModelHint') }}</p>

      <div class="group-openai-messages-section__grid mt-4">
        <div>
          <label class="input-label">{{ t('admin.groups.openaiMessages.opusModel') }}</label>
          <input
            v-model="form.opus_mapped_model"
            type="text"
            :placeholder="t('admin.groups.openaiMessages.opusModelPlaceholder')"
            class="input"
          />
        </div>
        <div>
          <label class="input-label">{{ t('admin.groups.openaiMessages.sonnetModel') }}</label>
          <input
            v-model="form.sonnet_mapped_model"
            type="text"
            :placeholder="t('admin.groups.openaiMessages.sonnetModelPlaceholder')"
            class="input"
          />
        </div>
        <div>
          <label class="input-label">{{ t('admin.groups.openaiMessages.haikuModel') }}</label>
          <input
            v-model="form.haiku_mapped_model"
            type="text"
            :placeholder="t('admin.groups.openaiMessages.haikuModelPlaceholder')"
            class="input"
          />
        </div>
      </div>
      <p class="input-hint">{{ t('admin.groups.openaiMessages.familyModelHint') }}</p>

      <div class="mt-4">
        <div class="group-openai-messages-section__row mb-2">
          <label class="input-label mb-0">{{ t('admin.groups.openaiMessages.exactMappings') }}</label>
          <button type="button" class="btn btn-secondary btn-sm" @click="addExactMapping">
            {{ t('admin.groups.openaiMessages.addExactMapping') }}
          </button>
        </div>
        <p class="input-hint">{{ t('admin.groups.openaiMessages.exactMappingsHint') }}</p>

        <div
          v-for="(row, index) in form.exact_model_mappings"
          :key="`mapping-${index}`"
          class="group-openai-messages-section__mapping-row"
        >
          <input
            v-model="row.claude_model"
            type="text"
            :placeholder="t('admin.groups.openaiMessages.claudeModelPlaceholder')"
            class="input"
          />
          <input
            v-model="row.target_model"
            type="text"
            :placeholder="t('admin.groups.openaiMessages.targetModelPlaceholder')"
            class="input"
          />
          <button type="button" class="btn btn-danger btn-sm" @click="removeExactMapping(index)">
            {{ t('common.actions.delete') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { CreateGroupForm, EditGroupForm } from './groupsForm'

const props = defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const { t } = useI18n()

function addExactMapping(): void {
  props.form.exact_model_mappings.push({
    claude_model: '',
    target_model: ''
  })
}

function removeExactMapping(index: number): void {
  props.form.exact_model_mappings.splice(index, 1)
}
</script>

<style scoped>
.group-openai-messages-section {
  border-color: var(--theme-page-border);
}

.group-openai-messages-section__grid {
  display: grid;
  gap: 0.75rem;
}

.group-openai-messages-section__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.group-openai-messages-section__mapping-row {
  display: grid;
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.group-openai-messages-section__title,
.group-openai-messages-section__label {
  color: var(--theme-page-text);
}

.group-openai-messages-section__hint {
  color: var(--theme-page-muted);
}

@media (min-width: 768px) {
  .group-openai-messages-section__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .group-openai-messages-section__mapping-row {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
  }
}
</style>
