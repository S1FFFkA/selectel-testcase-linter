package config

import "testing"

func TestApplyMap(t *testing.T) {
	cfg := Default()
	raw := map[string]any{
		"rules": map[string]any{
			"english_only": false,
		},
		"sensitive": map[string]any{
			"keywords": []any{"token", "client_secret="},
			"regex":    []any{"(?i)auth[_-]?key"},
		},
		"autofix": map[string]any{
			"sensitive_data": false,
			"english_only":   false,
		},
	}

	if err := ApplyMap(&cfg, raw); err != nil {
		t.Fatalf("ApplyMap() error = %v", err)
	}

	if cfg.Rules.EnglishOnly {
		t.Fatalf("expected english_only=false after apply")
	}
	if len(cfg.Sensitive.Keywords) != 2 || cfg.Sensitive.Keywords[1] != "client_secret" {
		t.Fatalf("unexpected keywords: %#v", cfg.Sensitive.Keywords)
	}
	if len(cfg.Sensitive.Regex) != 1 {
		t.Fatalf("unexpected regex: %#v", cfg.Sensitive.Regex)
	}
	if cfg.Autofix.SensitiveData {
		t.Fatalf("expected sensitive_data autofix=false after apply")
	}
	if cfg.Autofix.EnglishOnly {
		t.Fatalf("expected english_only autofix=false after apply")
	}
}

func TestApplyMapWithLinterLikeSettings(t *testing.T) {
	cfg := Default()
	raw := map[string]any{
		"rules": map[string]any{
			"starts_with_lower":   false,
			"english_only":        true,
			"no_emoji_or_special": false,
			"sensitive_data": map[string]any{
				"state": true,
				"words": []any{"password:", "token=", "merchant_pin"},
			},
		},
		"autofix": map[string]any{
			"starts_with_lower":   true,
			"english_only":        true,
			"no_emoji_or_special": false,
		},
	}

	if err := ApplyMap(&cfg, raw); err != nil {
		t.Fatalf("ApplyMap() error = %v", err)
	}

	if cfg.Rules.LowercaseStart {
		t.Fatalf("expected starts_with_lower=false after apply")
	}
	if cfg.Rules.NoSpecials {
		t.Fatalf("expected no_emoji_or_special=false after apply")
	}
	if !cfg.Rules.SensitiveData {
		t.Fatalf("expected sensitive_data.state=true after apply")
	}
	if len(cfg.Sensitive.Keywords) != 3 || cfg.Sensitive.Keywords[0] != "password" || cfg.Sensitive.Keywords[1] != "token" {
		t.Fatalf("unexpected sensitive words: %#v", cfg.Sensitive.Keywords)
	}
	if cfg.Autofix.NoSpecials {
		t.Fatalf("expected no_emoji_or_special autofix=false after apply")
	}
}
