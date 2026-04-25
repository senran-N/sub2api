import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import GeminiApiKeyTierSection from "../GeminiApiKeyTierSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("GeminiApiKeyTierSection", () => {
  it("renders AI Studio tier options and emits changes", async () => {
    const wrapper = mount(GeminiApiKeyTierSection, {
      props: {
        tierAiStudio: "aistudio_free",
      },
    });

    expect(wrapper.text()).toContain("admin.accounts.gemini.tier.label");
    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.tier.aiStudio.free",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.tier.aiStudio.paid",
    );
    expect(wrapper.text()).toContain("admin.accounts.gemini.tier.aiStudioHint");

    await wrapper.find("select").setValue("aistudio_paid");

    expect(wrapper.emitted("update:tierAiStudio")).toEqual([
      ["aistudio_paid"],
    ]);
  });
});
