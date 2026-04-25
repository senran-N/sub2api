import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import GeminiOAuthOptionsSection from "../GeminiOAuthOptionsSection.vue";
import type { GeminiOAuthType } from "../createAccountModalHelpers";

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
  oauthType: GeminiOAuthType = "google_one",
  aiStudioOauthEnabled = false,
) {
  return mount(GeminiOAuthOptionsSection, {
    props: {
      oauthType,
      aiStudioOauthEnabled,
      gcpProjectHelpLink: "https://example.test/gcp-project",
      tierGoogleOne: "google_one_free",
      tierGcp: "gcp_standard",
      tierAiStudio: "aistudio_free",
    },
    global: {
      stubs: {
        CreateAccountChoiceCard: {
          props: ["selected", "tone", "icon", "title", "description", "disabled"],
          emits: ["select"],
          template: `
            <button
              type="button"
              :data-selected="selected"
              :data-tone="tone"
              :data-icon="icon"
              :data-title="title"
              :disabled="disabled"
              @click="$emit('select')"
            >
              <span>{{ title }}</span>
              <span>{{ description }}</span>
              <slot name="meta" />
            </button>
          `,
        },
        Icon: true,
      },
    },
  });
}

describe("GeminiOAuthOptionsSection", () => {
  it("emits OAuth type changes without owning parent state", async () => {
    const wrapper = mountSection();

    await wrapper
      .find('[data-title="admin.accounts.gemini.oauthType.codeAssistTitle"]')
      .trigger("click");
    await wrapper
      .find('[data-title="admin.accounts.gemini.oauthType.googleOneTitle"]')
      .trigger("click");

    expect(wrapper.emitted("update:oauthType")).toEqual([
      ["code_assist"],
      ["google_one"],
    ]);
  });

  it("keeps advanced OAuth expansion local and respects AI Studio availability", async () => {
    const wrapper = mountSection("google_one", false);

    await wrapper
      .find(".gemini-oauth-options-section__inline-toggle")
      .trigger("click");

    const aiStudioButton = wrapper.find(
      '[data-title="admin.accounts.gemini.oauthType.customTitle"]',
    );
    expect(aiStudioButton.exists()).toBe(true);
    expect(aiStudioButton.attributes("disabled")).toBeDefined();
    await aiStudioButton.trigger("click");
    expect(wrapper.emitted("update:oauthType")).toBeUndefined();
  });

  it("emits AI Studio selection when the provider capability allows it", async () => {
    const wrapper = mountSection("google_one", true);

    await wrapper
      .find(".gemini-oauth-options-section__inline-toggle")
      .trigger("click");
    await wrapper
      .find('[data-title="admin.accounts.gemini.oauthType.customTitle"]')
      .trigger("click");

    expect(wrapper.emitted("update:oauthType")).toEqual([["ai_studio"]]);
  });

  it("emits tier changes for the active OAuth type", async () => {
    const googleOneWrapper = mountSection("google_one");
    await googleOneWrapper.find("select").setValue("google_ai_ultra");
    expect(googleOneWrapper.emitted("update:tierGoogleOne")).toEqual([
      ["google_ai_ultra"],
    ]);

    const codeAssistWrapper = mountSection("code_assist");
    await codeAssistWrapper.find("select").setValue("gcp_enterprise");
    expect(codeAssistWrapper.emitted("update:tierGcp")).toEqual([
      ["gcp_enterprise"],
    ]);

    const aiStudioWrapper = mountSection("ai_studio", true);
    await aiStudioWrapper.find("select").setValue("aistudio_paid");
    expect(aiStudioWrapper.emitted("update:tierAiStudio")).toEqual([
      ["aistudio_paid"],
    ]);
  });

  it("renders the Code Assist project guidance link", () => {
    const wrapper = mountSection("code_assist");

    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.oauthType.codeAssistRequirement",
    );
    expect(wrapper.find("a").attributes("href")).toBe(
      "https://example.test/gcp-project",
    );
  });
});
