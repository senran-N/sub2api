import { ref } from "vue";
import {
  createTempUnschedRule,
  loadTempUnschedRuleState,
  moveItemInPlace,
  type TempUnschedRuleForm,
} from "@/components/account/credentialsBuilder";
import { createStableObjectKeyResolver } from "@/utils/stableObjectKey";

export function useEditAccountTempUnschedRules() {
  const tempUnschedEnabled = ref(false);
  const tempUnschedRules = ref<TempUnschedRuleForm[]>([]);
  const getTempUnschedRuleKey =
    createStableObjectKeyResolver<TempUnschedRuleForm>(
      "edit-temp-unsched-rule",
    );

  const resetTempUnschedRules = () => {
    tempUnschedEnabled.value = false;
    tempUnschedRules.value = [];
  };

  const hydrateTempUnschedRulesFromCredentials = (
    credentials: Record<string, unknown> | undefined,
  ) => {
    const state = loadTempUnschedRuleState(credentials);
    tempUnschedEnabled.value = state.enabled;
    tempUnschedRules.value = state.rules;
  };

  const addTempUnschedRule = (preset?: TempUnschedRuleForm) => {
    tempUnschedRules.value.push(createTempUnschedRule(preset));
  };

  const removeTempUnschedRule = (index: number) => {
    tempUnschedRules.value.splice(index, 1);
  };

  const moveTempUnschedRule = (index: number, direction: number) => {
    moveItemInPlace(tempUnschedRules.value, index, direction);
  };

  const updateTempUnschedRule = (
    index: number,
    field: keyof TempUnschedRuleForm,
    value: TempUnschedRuleForm[keyof TempUnschedRuleForm],
  ) => {
    const rule = tempUnschedRules.value[index];
    if (!rule) {
      return;
    }

    const nextRule = { ...rule };
    if (field === "error_code" || field === "duration_minutes") {
      nextRule[field] = typeof value === "number" ? value : null;
    } else {
      nextRule[field] = typeof value === "string" ? value : "";
    }
    tempUnschedRules.value[index] = nextRule;
  };

  return {
    addTempUnschedRule,
    getTempUnschedRuleKey,
    hydrateTempUnschedRulesFromCredentials,
    moveTempUnschedRule,
    removeTempUnschedRule,
    resetTempUnschedRules,
    tempUnschedEnabled,
    tempUnschedRules,
    updateTempUnschedRule,
  };
}
