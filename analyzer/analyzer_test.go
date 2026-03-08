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

	goroot := runtime.GOROOT()
	t.Setenv("GOROOT", goroot)
	t.Setenv("GOTOOLCHAIN", "local")
	currentPath := os.Getenv("PATH")
	goBin := filepath.Join(goroot, "bin")
	t.Setenv("PATH", goBin+string(os.PathListSeparator)+currentPath)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}
	testdata := filepath.Join(filepath.Dir(wd), "testdata")
	analysistest.Run(t, testdata, analyzer.Analyzer, "loglinttest/src/p")
}
