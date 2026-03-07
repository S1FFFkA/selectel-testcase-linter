package analyzer

import (
	"go/ast"

	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/config"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/extract"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/model"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/rules/sensitive"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/rules/style"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/ruleset"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = NewDefaultAnalyzer()

func NewDefaultAnalyzer() *analysis.Analyzer {
	a, err := NewAnalyzerWithSettings(nil)
	if err != nil {
		panic(err)
	}
	return a
}

func NewAnalyzerWithSettings(settings any) (*analysis.Analyzer, error) {
	baseCfg := config.Default()
	if err := config.ApplySettings(&baseCfg, settings); err != nil {
		return nil, err
	}

	var configPath string
	an := &analysis.Analyzer{
		Name:     "loglint",
		Doc:      "checks log messages in slog and zap for style and data-safety rules",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			activeCfg, err := resolveConfig(baseCfg, configPath)
			if err != nil {
				return nil, err
			}
			return runAnalysis(pass, activeCfg)
		},
	}
	an.Flags.StringVar(&configPath, "config", "", "path to YAML config file")
	return an, nil
}

func resolveConfig(base config.Config, configPath string) (config.Config, error) {
	if configPath == "" {
		return base, nil
	}
	return config.LoadFromYAMLFile(configPath)
}

func runAnalysis(pass *analysis.Pass, cfg config.Config) (interface{}, error) {
	matcher, err := sensitive.NewMatcher(cfg.Sensitive.Keywords, cfg.Sensitive.Regex)
	if err != nil {
		return nil, err
	}

	rules := []ruleset.Rule{
		style.NewLowercaseRule(),
		style.NewEnglishRule(),
		style.NewNoSpecialsRule(),
		sensitive.NewDataRule(matcher),
	}
	ctx := ruleset.Context{Config: cfg}

	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	ins.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		msg, ok := extract.GetMessage(pass, call)
		if !ok {
			return
		}
		applyRules(pass, msg, rules, ctx)
	})
	return nil, nil
}

func applyRules(pass *analysis.Pass, msg model.LogMessage, rules []ruleset.Rule, ctx ruleset.Context) {
	for _, rule := range rules {
		rule.Check(pass, msg, ctx)
	}
}
