<template>
  <div class="antigravity-model-mapping-section">
    <label class="input-label">{{ t("admin.accounts.modelRestriction") }}</label>

    <div>
      <div class="antigravity-model-mapping-section__notice mb-3">
        <p class="text-xs">
          {{ t("admin.accounts.mapRequestModels") }}
        </p>
      </div>

      <div v-if="mappings.length > 0" class="mb-3 space-y-2">
        <div
          v-for="(mapping, index) in mappings"
          :key="mappingKey(mapping)"
          class="space-y-1"
        >
          <div class="flex items-center gap-2">
            <input
              :value="mapping.from"
              type="text"
              :class="
                getCreateValidationInputClasses(
                  !isValidWildcardPattern(mapping.from),
                  'flex-1',
                )
              "
              :placeholder="t('admin.accounts.requestModel')"
              @input="updateMapping(index, 'from', $event)"
            />
            <Icon
              name="arrowRight"
              size="sm"
              class="antigravity-model-mapping-section__arrow flex-shrink-0"
              :stroke-width="2"
            />
            <input
              :value="mapping.to"
              type="text"
              :class="
                getCreateValidationInputClasses(
                  mapping.to.includes('*'),
                  'flex-1',
                )
              "
              :placeholder="t('admin.accounts.actualModel')"
              @input="updateMapping(index, 'to', $event)"
            />
            <button
              type="button"
              class="antigravity-model-mapping-section__remove"
              :aria-label="t('common.delete')"
              @click="emit('remove', index)"
            >
              <Icon name="trash" size="sm" :stroke-width="2" />
            </button>
          </div>
          <p
            v-if="!isValidWildcardPattern(mapping.from)"
            class="antigravity-model-mapping-section__error"
          >
            {{ t("admin.accounts.wildcardOnlyAtEnd") }}
          </p>
          <p
            v-if="mapping.to.includes('*')"
            class="antigravity-model-mapping-section__error"
          >
            {{ t("admin.accounts.targetNoWildcard") }}
          </p>
        </div>
      </div>

      <button
        type="button"
        class="btn btn-secondary mb-3 w-full border-2 border-dashed"
        @click="emit('add')"
      >
        <Icon name="plus" size="sm" class="mr-1 inline" :stroke-width="2" />
        {{ t("admin.accounts.addMapping") }}
      </button>

      <div class="flex flex-wrap gap-2">
        <button
          v-for="preset in presetMappings"
          :key="preset.label"
          type="button"
          :class="getPresetMappingChipClasses(preset.tone)"
          @click="emit('addPreset', preset.from, preset.to)"
        >
          + {{ preset.label }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import {
  getPresetMappingChipClasses,
  isValidWildcardPattern,
  type PresetMapping,
} from "@/composables/useModelWhitelist";
import { getCreateValidationInputClasses } from "@/components/account/accountModalClasses";
import type { ModelMapping } from "@/components/account/credentialsBuilder";

defineProps<{
  mappings: ModelMapping[];
  presetMappings: PresetMapping[];
  mappingKey: (mapping: ModelMapping) => string;
}>();

const emit = defineEmits<{
  add: [];
  remove: [index: number];
  addPreset: [from: string, to: string];
  updateMapping: [index: number, field: keyof ModelMapping, value: string];
}>();

const { t } = useI18n();

const updateMapping = (
  index: number,
  field: keyof ModelMapping,
  event: Event,
) => {
  const target = event.target;
  if (!(target instanceof HTMLInputElement)) {
    return;
  }
  emit("updateMapping", index, field, target.value);
};
</script>

<style scoped>
.antigravity-model-mapping-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.antigravity-model-mapping-section__notice {
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  border-radius: var(--theme-auth-feedback-radius);
  background: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 84%,
    var(--theme-page-text)
  );
  padding: var(--theme-auth-callback-feedback-padding);
}

.antigravity-model-mapping-section__arrow {
  color: var(--theme-page-muted);
}

.antigravity-model-mapping-section__remove {
  border-radius: var(--theme-button-radius);
  background: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 12%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 88%,
    var(--theme-page-text)
  );
  padding: 0.5rem;
}

.antigravity-model-mapping-section__error {
  color: rgb(var(--theme-danger-rgb));
  font-size: 0.75rem;
}
</style>
