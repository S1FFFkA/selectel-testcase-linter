package config

import (
	"fmt"
	"strings"
)

type Config struct {
	Rules     RulesConfig
	Sensitive SensitiveConfig
	Autofix   AutofixConfig
}

type RulesConfig struct {
	LowercaseStart bool
	EnglishOnly    bool
	NoSpecials     bool
	SensitiveData  bool
}

type SensitiveConfig struct {
	Keywords []string
	Regex    []string
}

type AutofixConfig struct {
	LowercaseStart bool
	EnglishOnly    bool
	NoSpecials     bool
	SensitiveData  bool
}

func Default() Config {
	return Config{
		Rules: RulesConfig{
			LowercaseStart: true,
			EnglishOnly:    true,
			NoSpecials:     true,
			SensitiveData:  true,
		},
		Sensitive: SensitiveConfig{
			Keywords: defaultSensitiveKeywords(),
			Regex:    nil,
		},
		Autofix: AutofixConfig{
			LowercaseStart: true,
			EnglishOnly:    true,
			NoSpecials:     true,
			SensitiveData:  true,
		},
	}
}

func defaultSensitiveKeywords() []string {
	return []string{
		"password",
		"passwd",
		"pwd",
		"secret",
		"token",
		"api_key",
		"apikey",
		"access_key",
		"private_key",
		"credential",
		"bearer",
		"jwt",
		"session",
		"cookie",
	}
}

func ApplySettings(cfg *Config, settings any) error {
	parsed, err := Parse(*cfg, settings)
	if err != nil {
		return err
	}
	*cfg = parsed
	return nil
}

func Parse(base Config, raw any) (Config, error) {
	if raw == nil {
		return base, nil
	}
	asMap, ok := anyToMap(raw)
	if !ok {
		return Config{}, fmt.Errorf("config must be map")
	}
	cfg := base
	if err := applyMap(&cfg, asMap); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ApplyMap(cfg *Config, raw map[string]any) error {
	parsed, err := Parse(*cfg, raw)
	if err != nil {
		return err
	}
	*cfg = parsed
	return nil
}

func applyMap(cfg *Config, raw map[string]any) error {
	if cfg == nil {
		return fmt.Errorf("nil config")
	}

	if rulesRaw, ok := childMap(raw, "rules"); ok {
		if v, ok := firstBool(rulesRaw, "lowercase_start", "starts_with_lower", "starts_with_lowercase"); ok {
			cfg.Rules.LowercaseStart = v
		}
		if v, ok := boolValue(rulesRaw, "english_only"); ok {
			cfg.Rules.EnglishOnly = v
		}
		if v, ok := firstBool(rulesRaw, "no_specials_or_emoji", "no_emoji_or_special", "no_special_chars"); ok {
			cfg.Rules.NoSpecials = v
		}
		if v, ok := boolValue(rulesRaw, "sensitive_data"); ok {
			cfg.Rules.SensitiveData = v
		}
		if sensitiveDataRaw, ok := childMap(rulesRaw, "sensitive_data"); ok {
			if v, ok := firstBool(sensitiveDataRaw, "state", "enabled"); ok {
				cfg.Rules.SensitiveData = v
			}
			if words, ok := firstStringSlice(sensitiveDataRaw, "words", "keywords"); ok {
				cfg.Sensitive.Keywords = normalizeSlice(words)
			}
			if regex, ok := firstStringSlice(sensitiveDataRaw, "regex", "patterns"); ok {
				cfg.Sensitive.Regex = normalizeSlice(regex)
			}
		}
	}

	if sensitiveRaw, ok := childMap(raw, "sensitive"); ok {
		if v, ok := firstStringSlice(sensitiveRaw, "keywords", "words"); ok {
			cfg.Sensitive.Keywords = normalizeSlice(v)
		}
		if v, ok := firstStringSlice(sensitiveRaw, "regex", "patterns"); ok {
			cfg.Sensitive.Regex = normalizeSlice(v)
		}
	}

	if autofixRaw, ok := childMap(raw, "autofix"); ok {
		if v, ok := firstBool(autofixRaw, "lowercase_start", "starts_with_lower", "starts_with_lowercase"); ok {
			cfg.Autofix.LowercaseStart = v
		}
		if v, ok := boolValue(autofixRaw, "english_only"); ok {
			cfg.Autofix.EnglishOnly = v
		}
		if v, ok := firstBool(autofixRaw, "no_specials_or_emoji", "no_emoji_or_special", "no_special_chars"); ok {
			cfg.Autofix.NoSpecials = v
		}
		if v, ok := firstBool(autofixRaw, "sensitive_data", "sensitive"); ok {
			cfg.Autofix.SensitiveData = v
		}
	}

	return nil
}

func normalizeSlice(v []string) []string {
	out := make([]string, 0, len(v))
	for _, s := range v {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			out = append(out, strings.ToLower(trimmed))
		}
	}
	return out
}

func childMap(raw map[string]any, key string) (map[string]any, bool) {
	v, ok := raw[key]
	if !ok {
		return nil, false
	}
	return anyToMap(v)
}

func anyToMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, v := range m {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = v
		}
		return out, true
	default:
		return nil, false
	}
}

func boolValue(raw map[string]any, key string) (bool, bool) {
	v, ok := raw[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func stringSliceValue(raw map[string]any, key string) ([]string, bool) {
	v, ok := raw[key]
	if !ok {
		return nil, false
	}
	switch arr := v.(type) {
	case []string:
		return arr, true
	case []any:
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			s, ok := item.(string)
			if !ok {
				continue
			}
			out = append(out, s)
		}
		return out, true
	default:
		return nil, false
	}
}

func firstBool(raw map[string]any, keys ...string) (bool, bool) {
	for _, key := range keys {
		if v, ok := boolValue(raw, key); ok {
			return v, true
		}
	}
	return false, false
}

func firstStringSlice(raw map[string]any, keys ...string) ([]string, bool) {
	for _, key := range keys {
		if v, ok := stringSliceValue(raw, key); ok {
			return v, true
		}
	}
	return nil, false
}
