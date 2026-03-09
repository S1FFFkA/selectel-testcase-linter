package analyzer

import (
	"go/ast"

	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/config"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/extract"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/rules"
	go_translate "github.com/dinhcanh303/go_translate"
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

	an := &analysis.Analyzer{
		Name:     "logslinter",
		Doc:      "checks log messages in slog and zap for style and data-safety rules",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return runAnalysis(pass, baseCfg)
		},
	}
	return an, nil
}

func runAnalysis(pass *analysis.Pass, cfg config.Config) (interface{}, error) {
	matcher, err := rules.BuildSensitiveMatcher(cfg.Sensitive.Keywords, cfg.Sensitive.Regex)
	if err != nil {
		return nil, err
	}
	translator, _ := go_translate.NewTranslator(&go_translate.TranslateOptions{
		Provider: go_translate.ProviderGoogle,
	})

	checks := []rules.Rule{
		rules.NewLowercaseRule(),
		rules.NewEnglishRule(),
		rules.NewNoSpecialsRule(),
		rules.NewSensitiveRule(),
	}
	ctx := rules.Context{Config: cfg, Matcher: matcher, Translator: translator}

	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	ins.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		msg, ok := extract.GetMessage(pass, call)
		if !ok {
			return
		}
		applyRules(pass, msg, checks, ctx)
	})
	return nil, nil
}

func applyRules(pass *analysis.Pass, msg extract.Message, checks []rules.Rule, ctx rules.Context) {
	ctx.UnifiedFix = rules.BuildUnifiedLiteralFix(pass, msg, ctx)
	for _, rule := range checks {
		rule.Check(pass, msg, ctx)
	}
}
