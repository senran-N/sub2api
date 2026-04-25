import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import OpenAIAccountTypeSection from "../OpenAIAccountTypeSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection(accountCategory: "oauth-based" | "apikey" = "oauth-based") {
  return mount(OpenAIAccountTypeSection, {
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

describe("OpenAIAccountTypeSection", () => {
  it("emits account category changes without owning parent state", async () => {
    const wrapper = mountSection();
    const buttons = wrapper.findAll("button");

    await buttons[1].trigger("click");
    await buttons[0].trigger("click");

    expect(wrapper.emitted("update:accountCategory")).toEqual([
      ["apikey"],
      ["oauth-based"],
    ]);
  });

  it("renders OpenAI OAuth and API key options", () => {
    const wrapper = mountSection("apikey");
    const buttons = wrapper.findAll("button");

    expect(wrapper.text()).toContain("OAuth");
    expect(wrapper.text()).toContain("API Key");
    expect(wrapper.text()).toContain("admin.accounts.types.chatgptOauth");
    expect(wrapper.text()).toContain("admin.accounts.types.responsesApi");
    expect(buttons[1].attributes("data-selected")).toBe("true");
  });
});
