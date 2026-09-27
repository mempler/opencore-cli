package builder

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/newcore-network/opencore-cli/internal/config"
)

// fakeTypecheckProject lays out a project whose "tsc" prints output and exits with code.
func fakeTypecheckProject(t *testing.T, output string, code int) string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	project := t.TempDir()
	writeTestFile(t, project, "tsconfig.json", "{}")
	writeTestFile(t, project, "node_modules/typescript/bin/tsc", "process.stdout.write("+
		quoteJS(output)+"); process.exit("+strconv.Itoa(code)+")\n")
	return project
}

func quoteJS(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, "'", `\'`, "\n", `\n`).Replace(s) + "'"
}

func TestRunTypecheck_Passes(t *testing.T) {
	project := fakeTypecheckProject(t, "", 0)

	result, err := RunTypecheck(context.Background(), project)
	if err != nil {
		t.Fatalf("RunTypecheck returned error: %v", err)
	}
	if !result.Passed() {
		t.Fatalf("expected a pass, got %+v", result)
	}
}

func TestRunTypecheck_CountsDiagnostics(t *testing.T) {
	output := "src/a.ts(1,1): error TS2345: Argument of type '\"no:existe\"' is not assignable.\n" +
		"src/b.ts(2,5): error TS2322: Type 'string' is not assignable to type 'number'.\n"
	project := fakeTypecheckProject(t, output, 2)

	result, err := RunTypecheck(context.Background(), project)
	if err != nil {
		t.Fatalf("type errors must be reported through the result, got error: %v", err)
	}
	if result.Passed() || result.ErrorCount != 2 {
		t.Fatalf("expected 2 errors, got %+v", result)
	}
	if !strings.Contains(result.Output, "no:existe") {
		t.Fatalf("expected tsc's output to be kept, got %q", result.Output)
	}
}

func TestRunTypecheck_FailsWhenTscCrashes(t *testing.T) {
	project := fakeTypecheckProject(t, "unexpected crash", 1)

	if _, err := RunTypecheck(context.Background(), project); err == nil {
		t.Fatal("expected an error when tsc exits non-zero without diagnostics")
	}
}

func TestRunTypecheck_RequiresTypescript(t *testing.T) {
	project := t.TempDir()
	writeTestFile(t, project, "tsconfig.json", "{}")

	_, err := RunTypecheck(context.Background(), project)
	if err == nil || !strings.Contains(err.Error(), "requires typescript") {
		t.Fatalf("expected a missing-typescript error, got %v", err)
	}
}

func typecheckBuilder(project string, enabled bool) *Builder {
	cfg := &config.Config{Build: config.BuildConfig{Typecheck: &enabled}}
	b := New(cfg)
	b.resourceBuilder.projectPath = project
	return b
}

func TestBuilderTypecheck_FailsTheBuildOnErrors(t *testing.T) {
	project := fakeTypecheckProject(t, "src/a.ts(1,1): error TS2345: bad\n", 2)

	err := typecheckBuilder(project, true).typecheck(context.Background(), nil, true)
	if err == nil || !strings.Contains(err.Error(), "1 error(s)") {
		t.Fatalf("expected the build to fail, got %v", err)
	}
}

func TestBuilderTypecheck_WarnOnlyKeepsGoing(t *testing.T) {
	project := fakeTypecheckProject(t, "src/a.ts(1,1): error TS2345: bad\n", 2)

	b := typecheckBuilder(project, true)
	b.SetTypecheckWarnOnly(true)
	if err := b.typecheck(context.Background(), nil, true); err != nil {
		t.Fatalf("warn-only mode must not fail, got %v", err)
	}
}

func TestBuilderTypecheck_DisabledSkipsTsc(t *testing.T) {
	// No tsconfig and no typescript: running tsc would fail, so passing proves it was skipped.
	if err := typecheckBuilder(t.TempDir(), false).typecheck(context.Background(), nil, true); err != nil {
		t.Fatalf("expected the check to be skipped, got %v", err)
	}
}
