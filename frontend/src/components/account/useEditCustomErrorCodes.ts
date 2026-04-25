import { ref } from "vue";

interface EditCustomErrorCodeOptions {
  confirmSelection: (code: number) => boolean;
  showDuplicate: () => void;
  showInvalid: () => void;
}

export function useEditCustomErrorCodes(options: EditCustomErrorCodeOptions) {
  const customErrorCodesEnabled = ref(false);
  const selectedErrorCodes = ref<number[]>([]);
  const customErrorCodeInput = ref<number | null>(null);

  const resetCustomErrorCodes = () => {
    customErrorCodesEnabled.value = false;
    selectedErrorCodes.value = [];
    customErrorCodeInput.value = null;
  };

  const hydrateCustomErrorCodesFromCredentials = (
    credentials: Record<string, unknown>,
  ) => {
    customErrorCodesEnabled.value =
      credentials.custom_error_codes_enabled === true;
    const existingErrorCodes = credentials.custom_error_codes as
      | number[]
      | undefined;
    selectedErrorCodes.value =
      existingErrorCodes && Array.isArray(existingErrorCodes)
        ? [...existingErrorCodes]
        : [];
    customErrorCodeInput.value = null;
  };

  const toggleErrorCode = (code: number) => {
    const index = selectedErrorCodes.value.indexOf(code);
    if (index === -1) {
      if (!options.confirmSelection(code)) {
        return;
      }
      selectedErrorCodes.value.push(code);
      return;
    }
    selectedErrorCodes.value.splice(index, 1);
  };

  const addCustomErrorCode = () => {
    const code = customErrorCodeInput.value;
    if (code === null || code < 100 || code > 599) {
      options.showInvalid();
      return;
    }
    if (selectedErrorCodes.value.includes(code)) {
      options.showDuplicate();
      return;
    }
    if (!options.confirmSelection(code)) {
      return;
    }
    selectedErrorCodes.value.push(code);
    customErrorCodeInput.value = null;
  };

  const removeErrorCode = (code: number) => {
    const index = selectedErrorCodes.value.indexOf(code);
    if (index !== -1) {
      selectedErrorCodes.value.splice(index, 1);
    }
  };

  return {
    addCustomErrorCode,
    customErrorCodeInput,
    customErrorCodesEnabled,
    hydrateCustomErrorCodesFromCredentials,
    removeErrorCode,
    resetCustomErrorCodes,
    selectedErrorCodes,
    toggleErrorCode,
  };
}
