import { describe, expect, it } from "vitest";
import { useEditCredentialFields } from "../useEditCredentialFields";

describe("useEditCredentialFields", () => {
  it("resets base credential inputs and pool mode", () => {
    const fields = useEditCredentialFields();

    fields.editBaseUrl.value = "https://custom.example.com";
    fields.editApiKey.value = "key";
    fields.editSessionToken.value = "session";
    fields.poolModeEnabled.value = true;
    fields.poolModeRetryCount.value = 7;
    fields.resetCredentialFields("https://api.example.com");

    expect(fields.editBaseUrl.value).toBe("https://api.example.com");
    expect(fields.editApiKey.value).toBe("");
    expect(fields.editSessionToken.value).toBe("");
    expect(fields.poolModeEnabled.value).toBe(false);
    expect(fields.poolModeRetryCount.value).toBe(3);
  });

  it("hydrates compatible credential base URL and clamps pool retry count", () => {
    const fields = useEditCredentialFields();

    fields.hydrateCompatibleCredentialFields(
      {
        base_url: "https://compatible.example.com",
        pool_mode: true,
        pool_mode_retry_count: 99,
      },
      "https://default.example.com",
    );

    expect(fields.editBaseUrl.value).toBe("https://compatible.example.com");
    expect(fields.poolModeEnabled.value).toBe(true);
    expect(fields.poolModeRetryCount.value).toBe(10);
  });

  it("hydrates Bedrock pool mode using the edit modal legacy retry semantics", () => {
    const fields = useEditCredentialFields();

    fields.hydrateBedrockPoolMode({
      pool_mode: true,
      pool_mode_retry_count: 99,
    });

    expect(fields.poolModeEnabled.value).toBe(true);
    expect(fields.poolModeRetryCount.value).toBe(99);
  });
});
