import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import OpenAIOptionsSection from "../OpenAIOptionsSection.vue";

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
  return mount(OpenAIOptionsSection, {
    props: {
      accountCategory,
      passthroughEnabled: false,
      wsMode: "off",
      wsModeOptions: [
        { value: "off", label: "Off" },
        { value: "ctx_pool", label: "Context pool" },
        { value: "passthrough", label: "Passthrough" },
      ],
      wsModeConcurrencyHintKey: "admin.accounts.openai.wsModeConcurrencyHint",
      codexCliOnlyEnabled: false,
    },
    global: {
      stubs: {
        Select: {
          props: ["modelValue", "options"],
          emits: ["update:modelValue"],
          template: `
            <button
              type="button"
              data-testid="ws-mode"
              :data-value="modelValue"
              @click="$emit('update:modelValue', 'ctx_pool')"
            />
          `,
        },
      },
    },
  });
}

describe("OpenAIOptionsSection", () => {
  it("emits passthrough, websocket mode, and Codex toggle updates", async () => {
    const wrapper = mountSection();
    const buttons = wrapper.findAll("button");

    await buttons[0].trigger("click");
    await wrapper.find('[data-testid="ws-mode"]').trigger("click");
    await buttons[2].trigger("click");

    expect(wrapper.emitted("update:passthroughEnabled")).toEqual([[true]]);
    expect(wrapper.emitted("update:wsMode")).toEqual([["ctx_pool"]]);
    expect(wrapper.emitted("update:codexCliOnlyEnabled")).toEqual([[true]]);
  });

  it("hides the Codex toggle for API key accounts", () => {
    const wrapper = mountSection("apikey");

    expect(wrapper.text()).toContain("admin.accounts.openai.oauthPassthrough");
    expect(wrapper.text()).toContain("admin.accounts.openai.wsMode");
    expect(wrapper.text()).not.toContain("admin.accounts.openai.codexCLIOnly");
  });

  it("renders the websocket concurrency hint from the parent", () => {
    const wrapper = mountSection();

    expect(wrapper.text()).toContain(
      "admin.accounts.openai.wsModeConcurrencyHint",
    );
  });
});
