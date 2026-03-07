package main

import (
	"github.com/S1FFFkA/selectel-testcase-linter/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
