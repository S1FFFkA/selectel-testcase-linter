package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Rules     RulesConfig     `yaml:"rules"`
	Sensitive SensitiveConfig `yaml:"sensitive"`
	Autofix   AutofixConfig   `yaml:"autofix"`
}

type RulesConfig struct {
	LowercaseStart bool `yaml:"lowercase_start"`
	EnglishOnly    bool `yaml:"english_only"`
	NoSpecials     bool `yaml:"no_specials_or_emoji"`
	SensitiveData  bool `yaml:"sensitive_data"`
}

type SensitiveConfig struct {
	Keywords []string `yaml:"keywords"`
	Regex    []string `yaml:"regex"`
}

type AutofixConfig struct {
	LowercaseStart bool `yaml:"lowercase_start"`
	NoSpecials     bool `yaml:"no_specials_or_emoji"`
	SensitiveData  bool `yaml:"sensitive_data"`
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

func LoadFromYAMLFile(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var raw any
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return Config{}, fmt.Errorf("parse yaml config: %w", err)
	}
	return Parse(Default(), raw)
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
		if v, ok := boolValue(rulesRaw, "lowercase_start"); ok {
			cfg.Rules.LowercaseStart = v
		}
		if v, ok := boolValue(rulesRaw, "english_only"); ok {
			cfg.Rules.EnglishOnly = v
		}
		if v, ok := boolValue(rulesRaw, "no_specials_or_emoji"); ok {
			cfg.Rules.NoSpecials = v
		}
		if v, ok := boolValue(rulesRaw, "sensitive_data"); ok {
			cfg.Rules.SensitiveData = v
		}
	}

	if sensitiveRaw, ok := childMap(raw, "sensitive"); ok {
		if v, ok := stringSliceValue(sensitiveRaw, "keywords"); ok {
			cfg.Sensitive.Keywords = normalizeSlice(v)
		}
		if v, ok := stringSliceValue(sensitiveRaw, "regex"); ok {
			cfg.Sensitive.Regex = normalizeSlice(v)
		}
	}

	if autofixRaw, ok := childMap(raw, "autofix"); ok {
		if v, ok := boolValue(autofixRaw, "lowercase_start"); ok {
			cfg.Autofix.LowercaseStart = v
		}
		if v, ok := boolValue(autofixRaw, "no_specials_or_emoji"); ok {
			cfg.Autofix.NoSpecials = v
		}
		if v, ok := boolValue(autofixRaw, "sensitive_data"); ok {
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
