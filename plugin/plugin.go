package plugin

import (
	"github.com/S1FFFkA/selectel-testcase-linter/analyzer"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("logslinter", New)
}

type loglintPlugin struct {
	analyzer *analysis.Analyzer
}

func New(settings any) (register.LinterPlugin, error) {
	a, err := analyzer.NewAnalyzerWithSettings(settings)
	if err != nil {
		return nil, err
	}
	return &loglintPlugin{analyzer: a}, nil
}

func (p *loglintPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{p.analyzer}, nil
}

func (p *loglintPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
