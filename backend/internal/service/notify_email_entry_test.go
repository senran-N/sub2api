//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseNotifyEmails_EmptyInputs(t *testing.T) {
	require.Equal(t, []NotifyEmailEntry{}, ParseNotifyEmails(""))
	require.Equal(t, []NotifyEmailEntry{}, ParseNotifyEmails("   "))
	require.Equal(t, []NotifyEmailEntry{}, ParseNotifyEmails("[]"))
}

func TestParseNotifyEmails_OldFormat(t *testing.T) {
	result := ParseNotifyEmails(`["alice@example.com", "", " bob@example.com "]`)
	require.Equal(t, []NotifyEmailEntry{
		{Email: "alice@example.com", Verified: false, Disabled: false},
		{Email: "bob@example.com", Verified: false, Disabled: false},
	}, result)
}

func TestParseNotifyEmails_NewFormat(t *testing.T) {
	result := ParseNotifyEmails(`[
		{"email":"alice@example.com","verified":true,"disabled":false},
		{"email":"bob@example.com","verified":false,"disabled":true}
	]`)
	require.Equal(t, []NotifyEmailEntry{
		{Email: "alice@example.com", Verified: true, Disabled: false},
		{Email: "bob@example.com", Verified: false, Disabled: true},
	}, result)
}

func TestParseNotifyEmails_InvalidJSON(t *testing.T) {
	require.Equal(t, []NotifyEmailEntry{}, ParseNotifyEmails(`{not valid json`))
	require.Equal(t, []NotifyEmailEntry{}, ParseNotifyEmails(`{"email":"a@b.com"}`))
}

func TestMarshalNotifyEmails_RoundTrip(t *testing.T) {
	entries := []NotifyEmailEntry{
		{Email: "User@Example.com", Verified: true, Disabled: false},
		{Email: "disabled@example.com", Verified: true, Disabled: true},
	}
	encoded := MarshalNotifyEmails(entries)
	require.NotEmpty(t, encoded)
	require.Equal(t, entries, ParseNotifyEmails(encoded))
}

func TestMarshalNotifyEmails_EmptySlice(t *testing.T) {
	require.Equal(t, "[]", MarshalNotifyEmails(nil))
	require.Equal(t, "[]", MarshalNotifyEmails([]NotifyEmailEntry{}))
}
