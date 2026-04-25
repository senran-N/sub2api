import { describe, expect, it, vi } from "vitest";
import { useEditCustomErrorCodes } from "../useEditCustomErrorCodes";

function createCodes(confirmSelection = vi.fn(() => true)) {
  return {
    confirmSelection,
    showDuplicate: vi.fn(),
    showInvalid: vi.fn(),
    codes: useEditCustomErrorCodes({
      confirmSelection,
      showDuplicate: vi.fn(),
      showInvalid: vi.fn(),
    }),
  };
}

describe("useEditCustomErrorCodes", () => {
  it("hydrates custom error code state from credentials", () => {
    const { codes } = createCodes();

    codes.hydrateCustomErrorCodesFromCredentials({
      custom_error_codes_enabled: true,
      custom_error_codes: [429, 529],
    });

    expect(codes.customErrorCodesEnabled.value).toBe(true);
    expect(codes.selectedErrorCodes.value).toEqual([429, 529]);
    expect(codes.customErrorCodeInput.value).toBeNull();
  });

  it("toggles codes and respects confirmation", () => {
    const confirmSelection = vi
      .fn()
      .mockReturnValueOnce(false)
      .mockReturnValueOnce(true);
    const codes = useEditCustomErrorCodes({
      confirmSelection,
      showDuplicate: vi.fn(),
      showInvalid: vi.fn(),
    });

    codes.toggleErrorCode(429);
    expect(codes.selectedErrorCodes.value).toEqual([]);

    codes.toggleErrorCode(429);
    expect(codes.selectedErrorCodes.value).toEqual([429]);

    codes.toggleErrorCode(429);
    expect(codes.selectedErrorCodes.value).toEqual([]);
  });

  it("validates custom input and reports duplicates", () => {
    const showDuplicate = vi.fn();
    const showInvalid = vi.fn();
    const codes = useEditCustomErrorCodes({
      confirmSelection: vi.fn(() => true),
      showDuplicate,
      showInvalid,
    });

    codes.customErrorCodeInput.value = 99;
    codes.addCustomErrorCode();
    expect(showInvalid).toHaveBeenCalledTimes(1);

    codes.customErrorCodeInput.value = 500;
    codes.addCustomErrorCode();
    expect(codes.selectedErrorCodes.value).toEqual([500]);
    expect(codes.customErrorCodeInput.value).toBeNull();

    codes.customErrorCodeInput.value = 500;
    codes.addCustomErrorCode();
    expect(showDuplicate).toHaveBeenCalledTimes(1);

    codes.removeErrorCode(500);
    expect(codes.selectedErrorCodes.value).toEqual([]);
  });
});
