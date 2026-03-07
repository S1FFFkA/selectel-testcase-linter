package sensitive

import "testing"

func TestMatcherRegexAndKeywords(t *testing.T) {
	m, err := NewMatcher([]string{"token"}, []string{`(?i)secret[_-]?key`})
	if err != nil {
		t.Fatalf("NewMatcher() error = %v", err)
	}

	if !containsKeyWithSeparator("token: abc", m) {
		t.Fatalf("expected token with separator to be sensitive")
	}
	if !containsKeyWithSeparator("secret_key=abc", m) {
		t.Fatalf("expected regex-based key to be sensitive")
	}
	if containsKeyWithSeparator("token validated", m) {
		t.Fatalf("did not expect neutral text to be sensitive")
	}
	if !isWord("mySecretKeyValue", m) {
		t.Fatalf("expected regex to match identifier")
	}
}

func TestNewMatcherInvalidRegex(t *testing.T) {
	_, err := NewMatcher([]string{"token"}, []string{"("})
	if err == nil {
		t.Fatalf("expected error for invalid regex")
	}
}
