<template>
  <div
    v-if="form.subscription_type !== 'subscription'"
    :data-tour="tourTarget"
  >
    <div class="mb-1.5 flex items-center gap-1">
      <label class="group-exclusive-section__label text-sm font-medium">
        {{ t('admin.groups.form.exclusive') }}
      </label>
      <div class="group relative inline-flex">
        <Icon
          name="questionCircle"
          size="sm"
          :stroke-width="2"
          class="group-exclusive-section__help-icon cursor-help transition-colors"
        />
        <div class="pointer-events-none absolute bottom-full left-0 z-50 mb-2 w-72 opacity-0 transition-all duration-200 group-hover:pointer-events-auto group-hover:opacity-100">
          <div class="group-exclusive-section__tooltip shadow-lg">
            <p class="mb-2 text-xs font-medium">{{ t('admin.groups.exclusiveTooltip.title') }}</p>
            <p class="group-exclusive-section__tooltip-muted mb-2 text-xs leading-relaxed">
              {{ t('admin.groups.exclusiveTooltip.description') }}
            </p>
            <div class="group-exclusive-section__tooltip-panel">
              <p class="group-exclusive-section__tooltip-muted text-xs leading-relaxed">
                <span class="group-exclusive-section__tooltip-accent inline-flex items-center gap-1">
                  <Icon name="lightbulb" size="xs" />
                  {{ t('admin.groups.exclusiveTooltip.example') }}
                </span>
                {{ t('admin.groups.exclusiveTooltip.exampleContent') }}
              </p>
            </div>
            <div class="group-exclusive-section__tooltip-arrow absolute -bottom-1.5 left-3 h-3 w-3 rotate-45"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="flex items-center gap-3">
      <Toggle v-model="form.is_exclusive" />
      <span class="group-exclusive-section__state text-sm">
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

<style scoped>
.group-exclusive-section__label {
  color: var(--theme-page-text);
}

.group-exclusive-section__help-icon,
.group-exclusive-section__state {
  color: var(--theme-page-muted);
}

.group-exclusive-section__help-icon:hover {
  color: var(--theme-accent);
}

.group-exclusive-section__tooltip {
  border-radius: var(--theme-tooltip-radius);
  padding: var(--theme-tooltip-padding);
  background: color-mix(in srgb, var(--theme-surface-contrast) 96%, var(--theme-surface));
  color: var(--theme-surface-contrast-text);
}

.group-exclusive-section__tooltip-muted {
  color: color-mix(in srgb, var(--theme-surface-contrast-text) 72%, transparent);
}

.group-exclusive-section__tooltip-panel {
  border-radius: var(--theme-button-radius);
  padding: var(--theme-settings-card-panel-padding);
  background: color-mix(in srgb, var(--theme-surface-contrast-text) 8%, transparent);
}

.group-exclusive-section__tooltip-accent {
  color: color-mix(in srgb, var(--theme-accent) 72%, var(--theme-surface-contrast-text));
}

.group-exclusive-section__tooltip-arrow {
  background: color-mix(in srgb, var(--theme-surface-contrast) 96%, var(--theme-surface));
}
</style>
