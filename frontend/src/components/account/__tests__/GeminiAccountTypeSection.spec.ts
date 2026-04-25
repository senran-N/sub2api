import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import GeminiAccountTypeSection from "../GeminiAccountTypeSection.vue";

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
  return mount(GeminiAccountTypeSection, {
    props: {
      accountCategory,
      apiKeyHelpLink: "https://example.test/gemini-api-key",
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
        Icon: true,
      },
    },
  });
}

describe("GeminiAccountTypeSection", () => {
  it("emits account category changes without owning parent state", async () => {
    const wrapper = mountSection();
    const accountButtons = wrapper.findAll('[data-icon="key"]');

    await accountButtons[1].trigger("click");
    await accountButtons[0].trigger("click");

    expect(wrapper.emitted("update:accountCategory")).toEqual([
      ["apikey"],
      ["oauth-based"],
    ]);
  });

  it("emits a help event for parent-owned dialog state", async () => {
    const wrapper = mountSection();

    await wrapper
      .find(".gemini-account-type-section__help-button")
      .trigger("click");

    expect(wrapper.emitted("openHelp")).toEqual([[]]);
  });

  it("renders Gemini OAuth and API key options plus API key guidance", () => {
    const wrapper = mountSection("apikey");

    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.accountType.oauthTitle",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.accountType.oauthDesc",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.accountType.apiKeyTitle",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.accountType.apiKeyDesc",
    );
    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.accountType.apiKeyNote",
    );
    expect(wrapper.find("a").attributes("href")).toBe(
      "https://example.test/gemini-api-key",
    );
  });
});
