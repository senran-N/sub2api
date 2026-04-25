<template>
  <div class="space-y-5">
    <div>
      <label class="input-label">{{ nameLabel }}</label>
      <input
        :value="name"
        type="text"
        :required="nameRequired"
        class="input"
        :placeholder="namePlaceholder"
        data-tour="account-form-name"
        @input="emit('update:name', ($event.target as HTMLInputElement).value)"
      />
      <p v-if="nameHint" class="input-hint">{{ nameHint }}</p>
    </div>

    <div>
      <label class="input-label">{{ t("admin.accounts.notes") }}</label>
      <textarea
        :value="notes"
        rows="3"
        class="input"
        :placeholder="t('admin.accounts.notesPlaceholder')"
        @input="emit('update:notes', ($event.target as HTMLTextAreaElement).value)"
      ></textarea>
      <p class="input-hint">{{ t("admin.accounts.notesHint") }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

defineProps<{
  name: string;
  notes: string;
  nameLabel: string;
  namePlaceholder: string;
  nameHint: string;
  nameRequired: boolean;
}>();

const emit = defineEmits<{
  "update:name": [value: string];
  "update:notes": [value: string];
}>();

const { t } = useI18n();
</script>
