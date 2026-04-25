import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import AnthropicAccountTypeSection from "../AnthropicAccountTypeSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection(
  accountCategory: "oauth-based" | "apikey" | "bedrock" = "oauth-based",
) {
  return mount(AnthropicAccountTypeSection, {
    props: {
      accountCategory,
    },
    global: {
      stubs: {
        CreateAccountChoiceCard: {
          props: ["selected", "tone", "icon", "title", "description"],
          emits: ["select"],
          template: `
            <button
              type="button"
              :data-selected="selected"
              :data-tone="tone"
              :data-icon="icon"
              @click="$emit('select')"
            >
              <span>{{ title }}</span>
              <span>{{ description }}</span>
            </button>
          `,
        },
      },
    },
  });
}

describe("AnthropicAccountTypeSection", () => {
  it("emits account category changes without owning parent state", async () => {
    const wrapper = mountSection();
    const buttons = wrapper.findAll("button");

    await buttons[1].trigger("click");
    await buttons[2].trigger("click");
    await buttons[0].trigger("click");

    expect(wrapper.emitted("update:accountCategory")).toEqual([
      ["apikey"],
      ["bedrock"],
      ["oauth-based"],
    ]);
  });

  it("renders Claude Code, Claude Console, and Bedrock options", () => {
    const wrapper = mountSection("bedrock");
    const buttons = wrapper.findAll("button");

    expect(wrapper.text()).toContain("admin.accounts.claudeCode");
    expect(wrapper.text()).toContain("admin.accounts.oauthSetupToken");
    expect(wrapper.text()).toContain("admin.accounts.claudeConsole");
    expect(wrapper.text()).toContain("admin.accounts.apiKey");
    expect(wrapper.text()).toContain("admin.accounts.bedrockLabel");
    expect(wrapper.text()).toContain("admin.accounts.bedrockDesc");
    expect(buttons[2].attributes("data-selected")).toBe("true");
  });
});
