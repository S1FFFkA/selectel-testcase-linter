package analyzer_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/S1FFFkA/selectel-testcase-linter/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	// analysistest internally invokes "go". In some IDE setups on Windows,
	// PATH may point to a different Go patch version than the one used to
	// build the test binary, which causes version mismatch failures.
	goroot := runtime.GOROOT()
	t.Setenv("GOROOT", goroot)
	t.Setenv("GOTOOLCHAIN", "local")
	currentPath := os.Getenv("PATH")
	goBin := filepath.Join(goroot, "bin")
	t.Setenv("PATH", goBin+string(os.PathListSeparator)+currentPath)

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "a")
}
