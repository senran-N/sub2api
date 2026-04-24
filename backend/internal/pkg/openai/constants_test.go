package openai

import "testing"

func TestDefaultModels_ContainsGPT55(t *testing.T) {
	t.Parallel()

	byID := make(map[string]Model, len(DefaultModels))
	for _, model := range DefaultModels {
		byID[model.ID] = model
	}

	model, ok := byID["gpt-5.5"]
	if !ok {
		t.Fatal("expected default OpenAI models to expose gpt-5.5")
	}
	if model.DisplayName != "GPT-5.5" {
		t.Fatalf("expected gpt-5.5 display name %q, got %q", "GPT-5.5", model.DisplayName)
	}

	ids := DefaultModelIDs()
	found := false
	for _, id := range ids {
		if id == "gpt-5.5" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected DefaultModelIDs to include gpt-5.5")
	}
}
