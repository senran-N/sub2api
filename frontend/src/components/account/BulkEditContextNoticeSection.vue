<template>
  <div :class="getNoticeClasses('blue')">
    <p class="text-sm">
      <Icon name="infoCircle" size="md" class="mr-1.5 inline" :stroke-width="2" />
      {{ t("admin.accounts.bulkEdit.selectionInfo", { count }) }}
    </p>
  </div>

  <div v-if="isMixedPlatform" :class="getNoticeClasses('amber')">
    <p class="text-sm">
      <Icon
        name="exclamationTriangle"
        size="md"
        class="mr-1.5 inline"
        :stroke-width="2"
      />
      {{
        t("admin.accounts.bulkEdit.mixedPlatformWarning", {
          platforms: platforms.join(", "),
        })
      }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";

const props = defineProps<{
  count: number;
  platforms: string[];
}>();

const { t } = useI18n();

const isMixedPlatform = computed(() => props.platforms.length > 1);

const getNoticeClasses = (tone: "amber" | "blue") => [
  "bulk-edit-context-notice-section__notice",
  "bulk-edit-context-notice-section__notice-card",
  "border",
  `bulk-edit-context-notice-section__notice--${tone}`,
];
</script>

<style scoped>
.bulk-edit-context-notice-section__notice {
  border-radius: var(--theme-auth-feedback-radius);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.bulk-edit-context-notice-section__notice-card {
  padding: var(--theme-auth-callback-feedback-padding);
}

.bulk-edit-context-notice-section__notice--blue {
  --bulk-edit-notice-tone-rgb: var(--theme-info-rgb);
}

.bulk-edit-context-notice-section__notice--amber {
  --bulk-edit-notice-tone-rgb: var(--theme-warning-rgb);
}

.bulk-edit-context-notice-section__notice--blue,
.bulk-edit-context-notice-section__notice--amber {
  background: color-mix(
    in srgb,
    rgb(var(--bulk-edit-notice-tone-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--bulk-edit-notice-tone-rgb)) 84%,
    var(--theme-page-text)
  );
}
</style>
