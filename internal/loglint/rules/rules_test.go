package rules

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"unicode"

	go_translate "github.com/dinhcanh303/go_translate"
)

func TestStartsWithLowercase(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{name: "lowercase", message: "server started", want: true},
		{name: "uppercase", message: "Server started", want: false},
		{name: "empty", message: "", want: true},
		{name: "digit first", message: "8080 started", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := startsWithLowercase(tt.message)
			if got != tt.want {
				t.Fatalf("startsWithLowercase(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestIsEnglishOnly(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{name: "english", message: "server started", want: true},
		{name: "english with digits", message: "api v2 started", want: true},
		{name: "cyrillic", message: "сервер запущен", want: false},
		{name: "mixed", message: "server запущен", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEnglishOnly(tt.message)
			if got != tt.want {
				t.Fatalf("isEnglishOnly(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

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
	tests := []struct {
		name            string
		message         string
		allowFormatVerb bool
		want            string
	}{
		{name: "remove punctuation", message: "warning: something went wrong...", want: "warning: something went wrong"},
		{name: "keep separators", message: "token: abc_value", want: "token: abc_value"},
		{name: "keep format verb", message: "failed %s", allowFormatVerb: true, want: "failed %s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeMessage(tt.message, tt.allowFormatVerb)
			if got != tt.want {
				t.Fatalf("sanitizeMessage(%q) = %q, want %q", tt.message, got, tt.want)
			}
		})
	}
}

func TestMatcherRegexAndKeywords(t *testing.T) {
	m, err := BuildSensitiveMatcher([]string{"token"}, []string{`(?i)secret[_-]?key`})
	if err != nil {
		t.Fatalf("BuildSensitiveMatcher() error = %v", err)
	}
	if !containsSensitiveKeyWithSeparator("token: abc", m) {
		t.Fatalf("expected token with separator to be sensitive")
	}
	if !containsSensitiveKeyWithSeparator("secret_key=abc", m) {
		t.Fatalf("expected regex-based key to be sensitive")
	}
	if containsSensitiveKeyWithSeparator("token validated", m) {
		t.Fatalf("did not expect neutral text to be sensitive")
	}
	if !isSensitiveWord("mySecretKeyValue", m) {
		t.Fatalf("expected regex to match identifier")
	}
}

func TestMatcherInvalidRegex(t *testing.T) {
	_, err := BuildSensitiveMatcher([]string{"token"}, []string{"("})
	if err == nil {
		t.Fatalf("expected error for invalid regex")
	}
}

func TestContainsSensitiveKeyWithSeparator(t *testing.T) {
	m, err := BuildSensitiveMatcher([]string{"password", "token"}, nil)
	if err != nil {
		t.Fatalf("BuildSensitiveMatcher() error = %v", err)
	}
	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "password colon", text: "password: abc", want: true},
		{name: "token equal", text: "token=abc", want: true},
		{name: "token validated", text: "token validated", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsSensitiveKeyWithSeparator(tt.text, m)
			if got != tt.want {
				t.Fatalf("containsSensitiveKeyWithSeparator(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

type fakeTranslator struct {
	out []string
	err error
}

func (f fakeTranslator) TranslateText(_ context.Context, _ []string, _ string, _ ...string) ([]string, error) {
	return f.out, f.err
}

func TestTranslateToEnglish(t *testing.T) {
	tests := []struct {
		name       string
		translator fakeTranslator
		want       string
	}{
		{
			name:       "success",
			translator: fakeTranslator{out: []string{"server error"}},
			want:       "server error",
		},
		{
			name:       "empty result",
			translator: fakeTranslator{out: []string{}},
			want:       "",
		},
		{
			name:       "error",
			translator: fakeTranslator{err: errors.New("network")},
			want:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateToEnglish("ошибка сервера", tt.translator)
			if got != tt.want {
				t.Fatalf("translateToEnglish() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTranslateQuality_Live(t *testing.T) {
	if os.Getenv("RUN_TRANSLATION_QUALITY") != "1" {
		t.Skip("set RUN_TRANSLATION_QUALITY=1 to run live translation quality test")
	}

	translator, err := go_translate.NewTranslator(&go_translate.TranslateOptions{
		Provider: go_translate.ProviderGoogle,
	})
	if err != nil {
		t.Fatalf("NewTranslator() error = %v", err)
	}

	tests := []struct {
		name            string
		input           string
		expectedAnyWord []string
	}{
		{
			name:            "database connection",
			input:           "ошибка подключения к базе данных",
			expectedAnyWord: []string{"error", "connection", "connect", "database", "failed"},
		},
		{
			name:            "server started",
			input:           "сервер успешно запущен",
			expectedAnyWord: []string{"server", "started", "successfully"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateToEnglish(tt.input, translator)
			if strings.TrimSpace(got) == "" {
				t.Fatalf("empty translation for %q", tt.input)
			}
			if !isEnglishOnly(got) || containsCyrillic(got) {
				t.Fatalf("translation is not English enough: %q", got)
			}
			lower := strings.ToLower(got)
			okWord := false
			for _, word := range tt.expectedAnyWord {
				if strings.Contains(lower, word) {
					okWord = true
					break
				}
			}
			if !okWord {
				t.Fatalf("translation quality check failed: got %q, expected one of %#v", got, tt.expectedAnyWord)
			}
		})
	}
}

func containsCyrillic(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}
