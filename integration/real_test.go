package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRealRun(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	repoRoot := filepath.Dir(filepath.Dir(file))
	testModuleRoot := filepath.Join(repoRoot, "testdata")
	binPath := buildLoglintBinary(t, repoRoot)

	cmd := exec.Command(binPath, "./src/p")
	cmd.Dir = testModuleRoot
	cmd.Env = append(goToolchainEnv(), os.Environ()...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit code due to found issues, output:\n%s", string(out))
	}

	output := string(out)
	mustContain(t, output, "log message should start with a lowercase letter")
	mustContain(t, output, "log message should be in English only")
	mustContain(t, output, "log message should not contain special characters or emoji")
	mustContain(t, output, "log message may expose sensitive data")
}

func goToolchainEnv() []string {
	goroot := runtime.GOROOT()
	goBin := filepath.Join(goroot, "bin")
	pathValue := goBin + string(os.PathListSeparator) + os.Getenv("PATH")
	return []string{
		"GOROOT=" + goroot,
		"GOTOOLCHAIN=local",
		"PATH=" + pathValue,
	}
}

func mustContain(t *testing.T, output string, needle string) {
	t.Helper()
	if !strings.Contains(output, needle) {
		t.Fatalf("expected output to contain %q, got:\n%s", needle, output)
	}
}

func buildLoglintBinary(t *testing.T, repoRoot string) string {
	t.Helper()
	binName := "loglint-test-bin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(t.TempDir(), binName)
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/loglint")
	cmd.Dir = repoRoot
	cmd.Env = append(goToolchainEnv(), os.Environ()...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build loglint binary: %v\n%s", err, string(out))
	}
	return binPath
}
