import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import AntigravityModelMappingSection from "../AntigravityModelMappingSection.vue";
import type { PresetMapping } from "@/composables/useModelWhitelist";
import type { ModelMapping } from "@/components/account/credentialsBuilder";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

const mappings: ModelMapping[] = [
  { from: "claude-*", to: "claude-sonnet-4-5" },
];

const presetMappings: PresetMapping[] = [
  {
    label: "Claude->Sonnet",
    from: "claude-*",
    to: "claude-sonnet-4-5",
    tone: "info",
  },
];

function mountSection(extraProps: Record<string, unknown> = {}) {
  return mount(AntigravityModelMappingSection, {
    props: {
      mappings,
      presetMappings,
      mappingKey: (mapping: ModelMapping) => `${mapping.from}:${mapping.to}`,
      ...extraProps,
    },
    global: {
      stubs: {
        Icon: true,
      },
    },
  });
}

describe("AntigravityModelMappingSection", () => {
  it("emits granular mapping row updates", async () => {
    const wrapper = mountSection();
    const inputs = wrapper.findAll("input");

    await inputs[0].setValue("claude-sonnet-*");
    await inputs[1].setValue("claude-sonnet-4-6");

    expect(wrapper.emitted("updateMapping")).toEqual([
      [0, "from", "claude-sonnet-*"],
      [0, "to", "claude-sonnet-4-6"],
    ]);
  });

  it("emits row and preset actions without owning parent state", async () => {
    const wrapper = mountSection();
    const buttons = wrapper.findAll("button");

    await buttons[0].trigger("click");
    await buttons[1].trigger("click");
    await buttons[2].trigger("click");

    expect(wrapper.emitted("remove")).toEqual([[0]]);
    expect(wrapper.emitted("add")).toEqual([[]]);
    expect(wrapper.emitted("addPreset")).toEqual([
      ["claude-*", "claude-sonnet-4-5"],
    ]);
  });

  it("renders mapping validation errors inside the extracted section", () => {
    const wrapper = mountSection({
      mappings: [{ from: "claude-*bad", to: "target*" }],
    });

    expect(wrapper.text()).toContain("admin.accounts.wildcardOnlyAtEnd");
    expect(wrapper.text()).toContain("admin.accounts.targetNoWildcard");
  });
});
