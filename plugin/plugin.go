package plugin

import (
	"github.com/S1FFFkA/selectel-testcase-linter/analyzer"
	"golang.org/x/tools/go/analysis"
)

func New(settings any) ([]*analysis.Analyzer, error) {
	a, err := analyzer.NewAnalyzerWithSettings(settings)
	if err != nil {
		return nil, err
	}
	return []*analysis.Analyzer{a}, nil
}
