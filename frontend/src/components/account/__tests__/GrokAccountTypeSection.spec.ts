import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import GrokAccountTypeSection from "../GrokAccountTypeSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection(accountCategory: "apikey" | "upstream" | "session" = "apikey") {
  return mount(GrokAccountTypeSection, {
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

describe("GrokAccountTypeSection", () => {
  it("emits account category changes without owning parent state", async () => {
    const wrapper = mountSection();
    const buttons = wrapper.findAll("button");

    await buttons[1].trigger("click");
    await buttons[2].trigger("click");
    await buttons[0].trigger("click");

    expect(wrapper.emitted("update:accountCategory")).toEqual([
      ["upstream"],
      ["session"],
      ["apikey"],
    ]);
  });

  it("renders Grok API key, upstream, and session options", () => {
    const wrapper = mountSection("session");
    const buttons = wrapper.findAll("button");

    expect(wrapper.text()).toContain("API Key");
    expect(wrapper.text()).toContain("Upstream");
    expect(wrapper.text()).toContain("Session");
    expect(wrapper.text()).toContain("admin.accounts.types.grokApiKey");
    expect(wrapper.text()).toContain("admin.accounts.types.grokUpstream");
    expect(wrapper.text()).toContain("admin.accounts.types.grokSession");
    expect(buttons[2].attributes("data-selected")).toBe("true");
  });
});
