package ruleset

import (
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/config"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/model"
	"golang.org/x/tools/go/analysis"
)

type Context struct {
	Config config.Config
}

type Rule interface {
	Check(pass *analysis.Pass, msg model.LogMessage, ctx Context)
}
