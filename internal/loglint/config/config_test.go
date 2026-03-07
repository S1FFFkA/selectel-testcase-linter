package config

import "testing"

func TestApplyMap(t *testing.T) {
	cfg := Default()
	raw := map[string]any{
		"rules": map[string]any{
			"english_only": false,
		},
		"sensitive": map[string]any{
			"keywords": []any{"token", "client_secret"},
			"regex":    []any{"(?i)auth[_-]?key"},
		},
		"autofix": map[string]any{
			"sensitive_data": false,
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
}
