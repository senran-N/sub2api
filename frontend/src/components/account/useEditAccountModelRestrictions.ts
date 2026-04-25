import { ref } from "vue";
import {
  createEmptyModelRestrictionState,
  deriveAntigravityModelMappings,
  deriveModelRestrictionStateFromMapping,
} from "@/components/account/editAccountModalHelpers";
import {
  appendEmptyModelMapping,
  appendPresetModelMapping,
  removeModelMappingAt,
} from "@/components/account/accountModalInteractions";
import type { ModelMapping } from "@/components/account/credentialsBuilder";
import { createStableObjectKeyResolver } from "@/utils/stableObjectKey";

type ModelRestrictionMode = "whitelist" | "mapping";

interface EditAccountModelRestrictionOptions {
  onMappingExists?: (model: string) => void;
}

export function useEditAccountModelRestrictions(
  options: EditAccountModelRestrictionOptions = {},
) {
  const modelMappings = ref<ModelMapping[]>([]);
  const modelRestrictionMode = ref<ModelRestrictionMode>("whitelist");
  const allowedModels = ref<string[]>([]);
  const antigravityModelMappings = ref<ModelMapping[]>([]);
  const getModelMappingKey =
    createStableObjectKeyResolver<ModelMapping>("edit-model-mapping");
  const getAntigravityModelMappingKey =
    createStableObjectKeyResolver<ModelMapping>(
      "edit-antigravity-model-mapping",
    );

  const applyModelRestrictionState = (rawMapping: unknown) => {
    const nextState = deriveModelRestrictionStateFromMapping(rawMapping);
    modelRestrictionMode.value = nextState.mode;
    allowedModels.value = nextState.allowedModels;
    modelMappings.value = nextState.modelMappings;
  };

  const resetModelRestrictionState = () => {
    const nextState = createEmptyModelRestrictionState();
    modelRestrictionMode.value = nextState.mode;
    allowedModels.value = nextState.allowedModels;
    modelMappings.value = nextState.modelMappings;
  };

  const syncAntigravityModelRestrictionState = (
    credentials: Record<string, unknown> | undefined,
  ) => {
    antigravityModelMappings.value =
      deriveAntigravityModelMappings(credentials);
  };

  const resetAntigravityModelRestrictionState = () => {
    antigravityModelMappings.value = [];
  };

  const addModelMapping = () => {
    appendEmptyModelMapping(modelMappings.value);
  };

  const removeModelMapping = (index: number) => {
    removeModelMappingAt(modelMappings.value, index);
  };

  const updateModelMapping = (
    index: number,
    field: keyof ModelMapping,
    value: string,
  ) => {
    const mapping = modelMappings.value[index];
    if (!mapping) {
      return;
    }
    modelMappings.value[index] = {
      ...mapping,
      [field]: value,
    };
  };

  const addPresetMapping = (from: string, to: string) => {
    appendPresetModelMapping(modelMappings.value, from, to, (model) => {
      options.onMappingExists?.(model);
    });
  };

  const addAntigravityModelMapping = () => {
    appendEmptyModelMapping(antigravityModelMappings.value);
  };

  const removeAntigravityModelMapping = (index: number) => {
    removeModelMappingAt(antigravityModelMappings.value, index);
  };

  const updateAntigravityModelMapping = (
    index: number,
    field: keyof ModelMapping,
    value: string,
  ) => {
    const mapping = antigravityModelMappings.value[index];
    if (!mapping) {
      return;
    }
    antigravityModelMappings.value[index] = {
      ...mapping,
      [field]: value,
    };
  };

  const addAntigravityPresetMapping = (from: string, to: string) => {
    appendPresetModelMapping(
      antigravityModelMappings.value,
      from,
      to,
      (model) => {
        options.onMappingExists?.(model);
      },
    );
  };

  return {
    addAntigravityModelMapping,
    addAntigravityPresetMapping,
    addModelMapping,
    addPresetMapping,
    allowedModels,
    antigravityModelMappings,
    applyModelRestrictionState,
    getAntigravityModelMappingKey,
    getModelMappingKey,
    modelMappings,
    modelRestrictionMode,
    removeAntigravityModelMapping,
    removeModelMapping,
    resetAntigravityModelRestrictionState,
    resetModelRestrictionState,
    syncAntigravityModelRestrictionState,
    updateAntigravityModelMapping,
    updateModelMapping,
  };
}
