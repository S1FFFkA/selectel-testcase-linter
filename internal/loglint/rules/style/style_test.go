package style

import "testing"

func TestHasDisallowedRune(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		allowFormatVerb bool
		want            bool
	}{
		{name: "plain", message: "server started", want: false},
		{name: "emoji", message: "server started 🚀", want: true},
		{name: "punctuation", message: "failed!!!", want: true},
		{name: "allowed separators", message: "token: abc", want: false},
		{name: "format without allow", message: "failed %s", allowFormatVerb: false, want: true},
		{name: "format with allow", message: "failed %s", allowFormatVerb: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasDisallowedRune(tt.message, tt.allowFormatVerb)
			if got != tt.want {
				t.Fatalf("hasDisallowedRune(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestSanitizeMessage(t *testing.T) {
	got := sanitizeMessage("warning: something went wrong...", false)
	if got != "warning: something went wrong" {
		t.Fatalf("unexpected sanitized message: %q", got)
	}
}
