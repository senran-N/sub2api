import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import ModelRestrictionSection from "../ModelRestrictionSection.vue";
import type { ModelMapping } from "../credentialsBuilder";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, values?: Record<string, unknown>) =>
        values ? `${key}:${JSON.stringify(values)}` : key,
    }),
  };
});

const mappings: ModelMapping[] = [{ from: "gpt-*", to: "gpt-4o" }];

function mountSection(
  props: Partial<{
    mode: "whitelist" | "mapping";
    disabled: boolean;
  }> = {},
) {
  return mount(ModelRestrictionSection, {
    props: {
      mode: "mapping",
      allowedModels: [],
      platform: "openai",
      mappings,
      presetMappings: [
        {
          label: "Preset",
          from: "from",
          to: "to",
          tone: "success",
        },
      ],
      mappingKey: (mapping: ModelMapping) => `${mapping.from}-${mapping.to}`,
      ...props,
    },
    global: {
      stubs: {
        ModelWhitelistSelector: {
          props: ["modelValue", "platform"],
          emits: ["update:modelValue"],
          template:
            '<button type="button" @click="$emit(\'update:modelValue\', [\'gpt-4o\'])">selector</button>',
        },
      },
    },
  });
}

describe("ModelRestrictionSection", () => {
  it("emits mode, mapping, and preset actions", async () => {
    const wrapper = mountSection();
    const inputs = wrapper.findAll("input");

    await wrapper
      .findAll("button")
      .find((button) => button.text().includes("admin.accounts.modelWhitelist"))!
      .trigger("click");
    await inputs[0].setValue("claude-*");
    await inputs[1].setValue("claude-sonnet");
    await wrapper
      .find(".model-restriction-section__remove-button")
      .trigger("click");
    await wrapper
      .findAll("button")
      .find((button) => button.text().includes("Preset"))!
      .trigger("click");
    await wrapper
      .findAll("button")
      .find((button) => button.text().includes("admin.accounts.addMapping"))!
      .trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.mapRequestModels");
    expect(wrapper.emitted("update:mode")).toEqual([["whitelist"]]);
    expect(wrapper.emitted("updateMapping")).toEqual([
      [0, "from", "claude-*"],
      [0, "to", "claude-sonnet"],
    ]);
    expect(wrapper.emitted("removeMapping")).toEqual([[0]]);
    expect(wrapper.emitted("addPreset")).toEqual([["from", "to"]]);
    expect(wrapper.emitted("addMapping")).toEqual([[]]);
  });

  it("emits whitelist model updates through the selector", async () => {
    const wrapper = mountSection({ mode: "whitelist" });

    await wrapper
      .findAll("button")
      .find((button) => button.text() === "selector")!
      .trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.selectedModels");
    expect(wrapper.emitted("update:allowedModels")).toEqual([[["gpt-4o"]]]);
  });

  it("renders disabled notice without controls", () => {
    const wrapper = mountSection({ disabled: true });

    expect(wrapper.text()).toContain(
      "admin.accounts.openai.modelRestrictionDisabledByPassthrough",
    );
    expect(wrapper.find(".model-restriction-section__mode-button").exists()).toBe(
      false,
    );
  });
});
