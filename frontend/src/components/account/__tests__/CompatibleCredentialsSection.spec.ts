import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import CompatibleCredentialsSection from "../CompatibleCredentialsSection.vue";
import type { ModelMapping } from "../credentialsBuilder";

type CompatibleCredentialsSectionProps = InstanceType<
  typeof CompatibleCredentialsSection
>["$props"];

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

vi.mock("@/composables/useModelWhitelist", async () => {
  const actual = await vi.importActual<
    typeof import("@/composables/useModelWhitelist")
  >("@/composables/useModelWhitelist");
  return {
    ...actual,
    getPresetMappingChipClasses: () => ["preset-chip"],
  };
});

function mountSection(
  platform: "openai" | "gemini" = "gemini",
  propOverrides: Partial<CompatibleCredentialsSectionProps> = {},
) {
  return mount(CompatibleCredentialsSection, {
    props: {
      allowedModels: ["gpt-4.1"],
      apiKeyHint: "API key hint",
      apiKeyPlaceholder: "sk-...",
      apiKeyValue: "sk-old",
      baseUrl: "https://old.example.test/v1",
      baseUrlHint: "Base URL hint",
      baseUrlPlaceholder: "https://api.example.test/v1",
      baseUrlPresets: [
        { label: "Default", value: "https://default.example.test/v1" },
      ],
      customErrorCodeInput: null,
      customErrorCodesEnabled: false,
      mappingKey: (mapping: ModelMapping) => `${mapping.from}:${mapping.to}`,
      mappings: [{ from: "requested-model", to: "actual-model" }],
      modelRestrictionDisabled: false,
      modelRestrictionMode: "mapping",
      platform,
      poolModeEnabled: false,
      poolModeRetryCount: 2,
      presetMappings: [
        {
          label: "Preset",
          from: "preset-from",
          to: "preset-to",
          tone: "success",
        },
      ],
      selectedErrorCodes: [429],
      tierAiStudio: "aistudio_free",
      ...propOverrides,
    },
    global: {
      stubs: {
        CustomErrorCodesSection: {
          props: ["enabled", "inputValue", "selectedCodes"],
          emits: [
            "update:enabled",
            "update:inputValue",
            "toggleCode",
            "addCode",
            "removeCode",
          ],
          template: `
            <section data-testid="custom-errors">
              <button type="button" data-testid="toggle-code" @click="$emit('toggleCode', 500)">toggle</button>
              <button type="button" data-testid="add-code" @click="$emit('addCode')">add</button>
              <button type="button" data-testid="remove-code" @click="$emit('removeCode', 429)">remove</button>
              <button type="button" data-testid="enable-errors" @click="$emit('update:enabled', true)">enable</button>
              <button type="button" data-testid="input-error" @click="$emit('update:inputValue', 503)">input</button>
            </section>
          `,
        },
        GeminiApiKeyTierSection: {
          props: ["tierAiStudio"],
          emits: ["update:tierAiStudio"],
          template: `
            <button type="button" data-testid="gemini-tier" @click="$emit('update:tierAiStudio', 'aistudio_paid')">
              {{ tierAiStudio }}
            </button>
          `,
        },
        ModelRestrictionSection: {
          props: [
            "mode",
            "allowedModels",
            "platform",
            "mappings",
            "presetMappings",
            "mappingKey",
            "disabled",
          ],
          emits: [
            "update:mode",
            "update:allowedModels",
            "addMapping",
            "removeMapping",
            "addPreset",
            "updateMapping",
          ],
          template: `
            <section data-testid="model-restriction">
              <button type="button" data-testid="mode" @click="$emit('update:mode', 'whitelist')">mode</button>
              <button type="button" data-testid="models" @click="$emit('update:allowedModels', ['gpt-4.1-mini'])">models</button>
              <button type="button" data-testid="add-mapping" @click="$emit('addMapping')">add</button>
              <button type="button" data-testid="remove-mapping" @click="$emit('removeMapping', 0)">remove</button>
              <button type="button" data-testid="add-preset" @click="$emit('addPreset', 'preset-from', 'preset-to')">preset</button>
              <button type="button" data-testid="update-mapping" @click="$emit('updateMapping', 0, 'to', 'new-model')">update</button>
            </section>
          `,
        },
        PoolModeSection: {
          props: ["enabled", "retryCount"],
          emits: ["update:enabled", "update:retryCount"],
          template: `
            <section data-testid="pool-mode">
              <button type="button" data-testid="pool-enabled" @click="$emit('update:enabled', true)">enable</button>
              <button type="button" data-testid="pool-retry" @click="$emit('update:retryCount', 4)">retry</button>
            </section>
          `,
        },
      },
    },
  });
}

describe("CompatibleCredentialsSection", () => {
  it("emits base URL and API key updates without owning parent state", async () => {
    const wrapper = mountSection();
    const inputs = wrapper.findAll("input");

    await wrapper.find(".preset-chip").trigger("click");
    await inputs[0].setValue("https://custom.example.test/v1");
    await inputs[1].setValue("sk-new");

    expect(wrapper.emitted("update:baseUrl")).toEqual([
      ["https://default.example.test/v1"],
      ["https://custom.example.test/v1"],
    ]);
    expect(wrapper.emitted("update:apiKeyValue")).toEqual([["sk-new"]]);
    expect(wrapper.text()).toContain("Base URL hint");
    expect(wrapper.text()).toContain("API key hint");
  });

  it("forwards nested section events to the modal orchestration layer", async () => {
    const wrapper = mountSection();

    await wrapper.find('[data-testid="gemini-tier"]').trigger("click");
    await wrapper.find('[data-testid="mode"]').trigger("click");
    await wrapper.find('[data-testid="models"]').trigger("click");
    await wrapper.find('[data-testid="add-mapping"]').trigger("click");
    await wrapper.find('[data-testid="remove-mapping"]').trigger("click");
    await wrapper.find('[data-testid="add-preset"]').trigger("click");
    await wrapper.find('[data-testid="update-mapping"]').trigger("click");
    await wrapper.find('[data-testid="pool-enabled"]').trigger("click");
    await wrapper.find('[data-testid="pool-retry"]').trigger("click");
    await wrapper.find('[data-testid="toggle-code"]').trigger("click");
    await wrapper.find('[data-testid="add-code"]').trigger("click");
    await wrapper.find('[data-testid="remove-code"]').trigger("click");
    await wrapper.find('[data-testid="enable-errors"]').trigger("click");
    await wrapper.find('[data-testid="input-error"]').trigger("click");

    expect(wrapper.emitted("update:tierAiStudio")).toEqual([
      ["aistudio_paid"],
    ]);
    expect(wrapper.emitted("update:modelRestrictionMode")).toEqual([
      ["whitelist"],
    ]);
    expect(wrapper.emitted("update:allowedModels")).toEqual([
      [["gpt-4.1-mini"]],
    ]);
    expect(wrapper.emitted("addMapping")).toEqual([[]]);
    expect(wrapper.emitted("removeMapping")).toEqual([[0]]);
    expect(wrapper.emitted("addPreset")).toEqual([["preset-from", "preset-to"]]);
    expect(wrapper.emitted("updateMapping")).toEqual([[0, "to", "new-model"]]);
    expect(wrapper.emitted("update:poolModeEnabled")).toEqual([[true]]);
    expect(wrapper.emitted("update:poolModeRetryCount")).toEqual([[4]]);
    expect(wrapper.emitted("toggleCode")).toEqual([[500]]);
    expect(wrapper.emitted("addCode")).toEqual([[]]);
    expect(wrapper.emitted("removeCode")).toEqual([[429]]);
    expect(wrapper.emitted("update:customErrorCodesEnabled")).toEqual([[true]]);
    expect(wrapper.emitted("update:customErrorCodeInput")).toEqual([[503]]);
  });

  it("only renders the Gemini tier selector for Gemini-compatible accounts", () => {
    expect(mountSection("gemini").find('[data-testid="gemini-tier"]').exists()).toBe(
      true,
    );
    expect(mountSection("openai").find('[data-testid="gemini-tier"]').exists()).toBe(
      false,
    );
  });

  it("supports edit-mode API key behavior and optional nested sections", () => {
    const wrapper = mountSection("gemini", {
      apiKeyAutocomplete: "new-password",
      apiKeyLabel: "Edit API key",
      ignorePasswordManagers: true,
      showGeminiApiKeyTier: false,
      showModelRestriction: false,
    });
    const apiKeyInput = wrapper.find('input[type="password"]');

    expect(wrapper.text()).toContain("Edit API key");
    expect(apiKeyInput.attributes("required")).toBeUndefined();
    expect(apiKeyInput.attributes("autocomplete")).toBe("new-password");
    expect(apiKeyInput.attributes("data-lpignore")).toBe("true");
    expect(apiKeyInput.attributes("data-bwignore")).toBe("true");
    expect(wrapper.find('[data-testid="gemini-tier"]').exists()).toBe(false);
    expect(wrapper.find('[data-testid="model-restriction"]').exists()).toBe(
      false,
    );
  });
});
