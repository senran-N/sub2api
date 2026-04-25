import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import GrokSessionCredentialsSection from "../GrokSessionCredentialsSection.vue";

function mountSection() {
  return mount(GrokSessionCredentialsSection, {
    props: {
      mode: "single",
      singleToken: "token-a",
      batchInput: "name-a token-a",
      dryRun: false,
      testAfterCreate: true,
      result: null,
      submitting: false,
    },
    global: {
      stubs: {
        GrokSessionBatchImportPanel: {
          props: [
            "mode",
            "singleToken",
            "batchInput",
            "dryRun",
            "testAfterCreate",
            "result",
            "submitting",
          ],
          emits: [
            "update:mode",
            "update:singleToken",
            "update:batchInput",
            "update:dryRun",
            "update:testAfterCreate",
          ],
          template: `
            <div>
              <button data-testid="mode" @click="$emit('update:mode', 'batch')" />
              <button data-testid="single" @click="$emit('update:singleToken', 'token-b')" />
              <button data-testid="batch" @click="$emit('update:batchInput', 'name-b token-b')" />
              <button data-testid="dry-run" @click="$emit('update:dryRun', true)" />
              <button data-testid="test-after-create" @click="$emit('update:testAfterCreate', false)" />
              <span data-testid="single-token">{{ singleToken }}</span>
            </div>
          `,
        },
      },
    },
  });
}

describe("GrokSessionCredentialsSection", () => {
  it("passes parent state into the batch import panel", () => {
    const wrapper = mountSection();

    expect(wrapper.find('[data-testid="single-token"]').text()).toBe(
      "token-a",
    );
  });

  it("forwards Grok session credential updates to the modal owner", async () => {
    const wrapper = mountSection();

    await wrapper.find('[data-testid="mode"]').trigger("click");
    await wrapper.find('[data-testid="single"]').trigger("click");
    await wrapper.find('[data-testid="batch"]').trigger("click");
    await wrapper.find('[data-testid="dry-run"]').trigger("click");
    await wrapper.find('[data-testid="test-after-create"]').trigger("click");

    expect(wrapper.emitted("update:mode")).toEqual([["batch"]]);
    expect(wrapper.emitted("update:singleToken")).toEqual([["token-b"]]);
    expect(wrapper.emitted("update:batchInput")).toEqual([["name-b token-b"]]);
    expect(wrapper.emitted("update:dryRun")).toEqual([[true]]);
    expect(wrapper.emitted("update:testAfterCreate")).toEqual([[false]]);
  });
});
