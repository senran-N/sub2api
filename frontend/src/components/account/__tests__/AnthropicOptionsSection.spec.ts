import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import AnthropicOptionsSection from "../AnthropicOptionsSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("AnthropicOptionsSection", () => {
  it("emits API key passthrough toggle updates", async () => {
    const wrapper = mount(AnthropicOptionsSection, {
      props: {
        accountCategory: "apikey",
        apiKeyPassthroughEnabled: false,
      },
    });

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain(
      "admin.accounts.anthropic.apiKeyPassthrough",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.anthropic.apiKeyPassthroughDesc",
    );
    expect(wrapper.emitted("update:apiKeyPassthroughEnabled")).toEqual([
      [true],
    ]);
  });

  it("hides passthrough settings for non-API-key accounts", () => {
    const wrapper = mount(AnthropicOptionsSection, {
      props: {
        accountCategory: "oauth-based",
        apiKeyPassthroughEnabled: false,
      },
    });

    expect(wrapper.text()).toBe("");
  });
});
