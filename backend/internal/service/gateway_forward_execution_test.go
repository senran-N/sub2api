package service

import "testing"

import "github.com/stretchr/testify/require"

func TestLooksLikeToolSignatureError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "tool use marker",
			message: "invalid signature for tool_use block",
			want:    true,
		},
		{
			name:    "tool result marker",
			message: "Tool_Result payload signature mismatch",
			want:    true,
		},
		{
			name:    "function call marker",
			message: "function_call signature failed",
			want:    true,
		},
		{
			name:    "function response marker",
			message: "FunctionResponse content does not match signature",
			want:    true,
		},
		{
			name:    "plain signature error",
			message: "thinking block signature invalid",
			want:    false,
		},
		{
			name:    "empty message",
			message: "",
			want:    false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, testCase.want, looksLikeToolSignatureError(testCase.message))
		})
	}
}
