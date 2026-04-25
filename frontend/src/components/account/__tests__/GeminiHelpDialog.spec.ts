import { describe, expect, it, vi } from "vitest";
import { defineComponent } from "vue";
import { mount } from "@vue/test-utils";
import GeminiHelpDialog from "../GeminiHelpDialog.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

const BaseDialogStub = defineComponent({
  name: "BaseDialogStub",
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: "" },
  },
  emits: ["close"],
  template: `
    <section v-if="show" data-testid="gemini-help-shell">
      <h2>{{ title }}</h2>
      <slot />
      <slot name="footer" />
    </section>
  `,
});

const helpLinks = {
  apiKey: "https://example.test/api-key",
  aiStudioPricing: "https://example.test/pricing",
  countryChange: "https://example.test/country-change",
  countryCheck: "https://example.test/country-check",
  gcpProject: "https://example.test/gcp-project",
  geminiWebActivation: "https://example.test/gemini-web",
};

const quotaDocs = {
  aiStudio: "https://example.test/ai-studio-quota",
  codeAssist: "https://example.test/code-assist-quota",
  vertex: "https://example.test/vertex-quota",
};

function mountDialog(show = true) {
  return mount(GeminiHelpDialog, {
    props: {
      show,
      helpLinks,
      quotaDocs,
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
      },
    },
  });
}

describe("GeminiHelpDialog", () => {
  it("renders setup, quota, and API-key guidance from parent-provided links", () => {
    const wrapper = mountDialog();
    const hrefs = wrapper.findAll("a").map((link) => link.attributes("href"));

    expect(wrapper.text()).toContain("admin.accounts.gemini.helpDialog.title");
    expect(wrapper.text()).toContain("admin.accounts.gemini.setupGuide.title");
    expect(wrapper.text()).toContain("修改归属地");
    expect(wrapper.text()).toContain("admin.accounts.gemini.quotaPolicy.title");
    expect(wrapper.text()).toContain(
      "admin.accounts.gemini.helpDialog.apiKeySection",
    );
    expect(hrefs).toEqual(
      expect.arrayContaining([
        helpLinks.apiKey,
        helpLinks.aiStudioPricing,
        helpLinks.countryChange,
        helpLinks.countryCheck,
        helpLinks.gcpProject,
        helpLinks.geminiWebActivation,
        quotaDocs.aiStudio,
        quotaDocs.codeAssist,
        quotaDocs.vertex,
      ]),
    );
  });

  it("emits close without owning parent dialog state", async () => {
    const wrapper = mountDialog();

    await wrapper.find("button.btn-primary").trigger("click");

    expect(wrapper.emitted("close")).toEqual([[]]);
  });

  it("does not render dialog contents when hidden", () => {
    const wrapper = mountDialog(false);

    expect(wrapper.find('[data-testid="gemini-help-shell"]').exists()).toBe(
      false,
    );
  });
});
