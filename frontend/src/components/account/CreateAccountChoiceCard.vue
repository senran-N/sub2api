<template>
  <button
    type="button"
    :class="cardClasses"
    :disabled="disabled"
    @click="emit('select')"
  >
    <div :class="iconClasses">
      <Icon :name="icon" size="sm" />
    </div>
    <div>
      <span class="create-account-choice-card__title block text-sm font-medium">
        {{ title }}
      </span>
      <span class="create-account-choice-card__description text-xs">
        {{ description }}
      </span>
      <div v-if="$slots.meta" class="mt-2 flex flex-wrap gap-1.5">
        <slot name="meta" />
      </div>
    </div>
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Icon from "@/components/icons/Icon.vue";

type ChoiceTone =
  | "rose"
  | "orange"
  | "purple"
  | "amber"
  | "green"
  | "blue"
  | "emerald";

defineSlots<{
  meta?: () => unknown;
}>();

const props = withDefaults(
  defineProps<{
    selected: boolean;
    tone: ChoiceTone;
    icon: InstanceType<typeof Icon>["$props"]["name"];
    title: string;
    description: string;
    disabled?: boolean;
  }>(),
  {
    disabled: false,
  },
);

const emit = defineEmits<{
  select: [];
}>();

const selectedToneClass = computed(
  () => `create-account-choice-card--${props.tone}`,
);
const selectedIconToneClass = computed(
  () => `create-account-choice-card__icon--${props.tone}`,
);

const cardClasses = computed(() => [
  "create-account-choice-card",
  "flex items-center gap-3 border-2 text-left transition-all",
  props.selected
    ? selectedToneClass.value
    : "create-account-choice-card--idle",
  props.disabled && "create-account-choice-card--disabled",
]);

const iconClasses = computed(() => [
  "create-account-choice-card__icon",
  "flex h-8 w-8 shrink-0 items-center justify-center",
  props.selected
    ? selectedIconToneClass.value
    : "create-account-choice-card__icon--idle",
]);
</script>

<style scoped>
.create-account-choice-card {
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 0.75rem;
}

.create-account-choice-card--idle {
  border-color: color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 86%,
    var(--theme-surface)
  );
}

.create-account-choice-card--disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.create-account-choice-card--rose,
.create-account-choice-card__icon--rose {
  --create-account-choice-tone-rgb: var(--theme-brand-rose-rgb);
}

.create-account-choice-card--orange,
.create-account-choice-card__icon--orange {
  --create-account-choice-tone-rgb: var(--theme-brand-orange-rgb);
}

.create-account-choice-card--purple,
.create-account-choice-card__icon--purple {
  --create-account-choice-tone-rgb: var(--theme-brand-purple-rgb);
}

.create-account-choice-card--amber,
.create-account-choice-card__icon--amber {
  --create-account-choice-tone-rgb: var(--theme-warning-rgb);
}

.create-account-choice-card--green,
.create-account-choice-card__icon--green {
  --create-account-choice-tone-rgb: var(--theme-success-rgb);
}

.create-account-choice-card--blue,
.create-account-choice-card__icon--blue {
  --create-account-choice-tone-rgb: var(--theme-info-rgb);
}

.create-account-choice-card--emerald,
.create-account-choice-card__icon--emerald {
  --create-account-choice-tone-rgb: var(--theme-success-rgb);
}

.create-account-choice-card--rose,
.create-account-choice-card--orange,
.create-account-choice-card--purple,
.create-account-choice-card--amber,
.create-account-choice-card--green,
.create-account-choice-card--blue,
.create-account-choice-card--emerald {
  border-color: rgb(var(--create-account-choice-tone-rgb));
  background: color-mix(
    in srgb,
    rgb(var(--create-account-choice-tone-rgb)) 12%,
    var(--theme-surface)
  );
}

.create-account-choice-card__icon {
  border-radius: var(--theme-button-radius);
}

.create-account-choice-card__icon--idle {
  background: color-mix(in srgb, var(--theme-page-border) 58%, transparent);
  color: var(--theme-page-muted);
}

.create-account-choice-card__icon--rose,
.create-account-choice-card__icon--orange,
.create-account-choice-card__icon--purple,
.create-account-choice-card__icon--amber,
.create-account-choice-card__icon--green,
.create-account-choice-card__icon--blue,
.create-account-choice-card__icon--emerald {
  background: rgb(var(--create-account-choice-tone-rgb));
  color: var(--theme-filled-text);
}

.create-account-choice-card__title {
  color: var(--theme-page-text);
}

.create-account-choice-card__description {
  color: var(--theme-page-muted);
}
</style>
